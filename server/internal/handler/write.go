package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func (s *HandlerState) Write(w http.ResponseWriter, r *http.Request) {
	_, err := os.Open(s.fileStore.FilePath)
	if os.IsNotExist(err) {
		log.Printf("creating %s", s.fileStore.FilePath)
		_, err = os.Create(s.fileStore.FilePath)
		if err != nil {
			log.Fatalf("Can't open the file")
		}
	}

	ctx, cancel := context.WithCancel(r.Context())
	clientAddr := r.RemoteAddr
	log.Printf("websocket upgrade requested from %s", clientAddr)
	c, err := s.upgrader.Upgrade(w, r, nil)
	connectionId := uuid.New().String()
	s.addClient(connectionId)
	s.broker.Publish(Broadcast{sourceClientId: connectionId})

	messageCh := make(chan Message, 1)
	errCh := make(chan error, 1)
	broadcastCh := s.broker.Subscribe()

	requestNumber := 0

	defer func() {
		logWithPrefix(connectionId, requestNumber, "closing the connection")
		s.removeClient(connectionId)
		s.broker.Publish(Broadcast{sourceClientId: connectionId})
		s.broker.Unsubscribe(broadcastCh)
		close(messageCh)
		close(errCh)
		c.Close()
		cancel()
	}()

	if err != nil {
		log.Printf("upgrade %s: %v", clientAddr, err)
		return
	}

	// this gorountine reads from the websocket and pushes to a channel that we can select on
	go func() {
		for {
			var m Message
			err := c.ReadJSON(&m)
			if err != nil {
				errCh <- err
				return
			}
			select {
			case messageCh <- m:
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	var first = true

	for {
		select {
		case m := <-messageCh:
			// if there are no patch objs, it's probably the first request
			if len(m.PatchObjs) == 0 && first {
				first = false

				// send the initial state
				logWithPrefix(connectionId, requestNumber, "sending initial state")
				if len(s.clientInfoList) == 1 && s.clientIndexMap[connectionId] == 0 {
					response := &StateResponse{State: "EDITOR", InitialDoc: s.fileStore.Read()}
					payload, err := MarshalMyResponse(response)
					logWithPrefix(connectionId, requestNumber, fmt.Sprintf("sending initial state: %v", response))
					logWithPrefix(connectionId, requestNumber, fmt.Sprintf("sending initial state: %v", payload))
					if err != nil {
						log.Printf("marshal response: %v", err)
						return
					}
					if err := c.WriteJSON(payload); err != nil {
						return
					}
					payload, err = MarshalMyResponse(&ClientResponse{Count: len(s.clientInfoList)})
					if err != nil {
						logWithPrefix(connectionId, requestNumber, fmt.Sprintf("marshal response: %v", err))
						return
					}
					if err := c.WriteJSON(payload); err != nil {
						return
					}

				} else {
					payload, err := MarshalMyResponse(&StateResponse{State: "READER", InitialDoc: s.fileStore.Read()})
					if err != nil {
						logWithPrefix(connectionId, requestNumber, fmt.Sprintf("marshal response: %v", err))
						return
					}
					if err := c.WriteJSON(payload); err != nil {
						return
					}
					clientPayload, clientErr := MarshalMyResponse(&ClientResponse{Count: len(s.clientInfoList)})
					if clientErr != nil {
						logWithPrefix(connectionId, requestNumber, fmt.Sprintf("marshal response: %v", clientErr))
						return
					}
					if err := c.WriteJSON(clientPayload); err != nil {
						return
					}
				}
				continue
			}

			// if there are patch objs and this client is the editor, we'll apply them to the file
			if len(m.PatchObjs) >= 1 && s.clientIndexMap[connectionId] == 0 {
				// convert PatchObjs to library Patch type
				dmpPatches := make([]diffmatchpatch.Patch, 0, len(m.PatchObjs))
				for _, po := range m.PatchObjs {
					dmpPatches = append(dmpPatches, po.ToDMP(s.dmp))
				}
				result, _ := s.dmp.PatchApply(dmpPatches, s.fileStore.Read())
				logWithPrefix(connectionId, requestNumber, "applying patches")
				s.fileStore.Write([]byte(result))
				payload, err := MarshalMyResponse(&EditorResponse{Status: "OK"})
				if err != nil {
					logWithPrefix(connectionId, requestNumber, fmt.Sprintf("marshal response: %v", err))
					return
				}
				if err := c.WriteJSON(payload); err != nil {
					logWithPrefix(connectionId, requestNumber, fmt.Sprintf("write %s: %v", clientAddr, err))
					return
				}
				// publish to all other clients that there's an update
				s.broker.Publish(Broadcast{sourceClientId: connectionId})
			}
		case b := <-broadcastCh:
			// on broadcast, send the updated client count to all clients

			if b.sourceClientId == connectionId {
				continue
			}
			logWithPrefix(connectionId, requestNumber, fmt.Sprintf("got a broadcast: %v", len(s.clientInfoList)))
			payload, err := MarshalMyResponse(&ClientResponse{Count: len(s.clientInfoList)})
			if err != nil {
				logWithPrefix(connectionId, requestNumber, fmt.Sprintf("marshal response: %v", err))
				return
			}
			if err := c.WriteJSON(payload); err != nil {
				return
			}

			// if the client is first in line, send them a state response to become editor
			if s.clientIndexMap[connectionId] == 0 {
				readerPayload, readerErr := MarshalMyResponse(&StateResponse{State: "EDITOR", InitialDoc: s.fileStore.Read()})
				if readerErr != nil {
					logWithPrefix(connectionId, requestNumber, fmt.Sprintf("marshal response: %v", readerErr))
					return
				}
				if err := c.WriteJSON(readerPayload); err != nil {
					logWithPrefix(connectionId, requestNumber, fmt.Sprintf("write %s: %v", clientAddr, err))
					return
				}
			} else {
				// if the client is not first in line, send them a reader response
				readerPayload, readerErr := MarshalMyResponse(&ReaderResponse{Status: "OK", Doc: s.fileStore.Read()})
				if readerErr != nil {
					logWithPrefix(connectionId, requestNumber, fmt.Sprintf("marshal response: %v", readerErr))
					return
				}
				if err := c.WriteJSON(readerPayload); err != nil {
					logWithPrefix(connectionId, requestNumber, fmt.Sprintf("write %s: %v", clientAddr, err))
					return
				}
			}
		case err := <-errCh:
			logWithPrefix(connectionId, requestNumber, fmt.Sprintf("read %s: %v", clientAddr, err))
			return
		}
		requestNumber++
	}
}

// builds the logging prefix for a given connection and request number
// example output: [qew4u23e4 - 3]
func loggingPrefix(connectionId string, requestNumber int) string {
	return "[" + connectionId + " - " + fmt.Sprintf("%d", requestNumber) + "]"
}

func logWithPrefix(connectionId string, requestNumber int, message string) {
	prefix := loggingPrefix(connectionId, requestNumber)
	log.Printf("%s %s", prefix, message)
}
