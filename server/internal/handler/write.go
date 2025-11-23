package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"

	"github.com/google/uuid"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// readMessages starts a goroutine to read JSON messages and return channels
func (s *HandlerState) readMessages(ctx context.Context, c *websocket.Conn) (<-chan Message, <-chan error) {
	messageCh := make(chan Message, 1)
	errCh := make(chan error, 1)
	go func() {
		for {
			var m Message
			if err := c.ReadJSON(&m); err != nil {
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
	return messageCh, errCh
}

// sendJSON marshals a response and writes it to the websocket
func sendJSON(c *websocket.Conn, v MyResponse) error {
	payload, err := MarshalMyResponse(v)
	if err != nil {
		return err
	}
	return c.WriteJSON(payload)
}

func (s *HandlerState) Write(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(s.fileStore.FilePath); os.IsNotExist(err) {
		log.Printf("creating %s", s.fileStore.FilePath)
		if _, err := os.Create(s.fileStore.FilePath); err != nil {
			log.Fatalf("Can't open the file")
		}
	}

	ctx, cancel := context.WithCancel(r.Context())
	clientAddr := r.RemoteAddr
	log.Printf("websocket upgrade requested from %s", clientAddr)

	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade %s: %v", clientAddr, err)
		cancel()
		return
	}

	connectionId := uuid.New().String()
	s.addClient(connectionId)
	s.broker.Publish(Broadcast{sourceClientId: connectionId})

	messageCh, errCh := s.readMessages(ctx, c)
	broadcastCh := s.broker.Subscribe()

	requestNumber := 0
	first := true

	defer func() {
		logWithPrefix(connectionId, requestNumber, "closing the connection")
		s.removeClient(connectionId)
		s.broker.Publish(Broadcast{sourceClientId: connectionId})
		s.broker.Unsubscribe(broadcastCh)
		c.Close()
		cancel()
	}()

	for {
		select {
		case m := <-messageCh:
			if len(m.PatchObjs) == 0 && first {
				first = false
				isEditor := len(s.clientInfoList) == 1 && s.clientIndexMap[connectionId] == 0
				initial := s.fileStore.Read()
				if isEditor {
					sendJSON(c, &StateResponse{State: "EDITOR", InitialDoc: initial})
				} else {
					sendJSON(c, &StateResponse{State: "READER", InitialDoc: initial})
				}
				sendJSON(c, &ClientResponse{Count: len(s.clientInfoList)})
				continue
			}

			if len(m.PatchObjs) >= 1 && s.clientIndexMap[connectionId] == 0 {
				dmpPatches := make([]diffmatchpatch.Patch, 0, len(m.PatchObjs))
				for _, po := range m.PatchObjs {
					dmpPatches = append(dmpPatches, po.ToDMP(s.dmp))
				}
				result, _ := s.dmp.PatchApply(dmpPatches, s.fileStore.Read())
				s.fileStore.Write([]byte(result))
				sendJSON(c, &EditorResponse{Status: "OK"})
				s.broker.Publish(Broadcast{sourceClientId: connectionId})
			}

		case b := <-broadcastCh:
			if b.sourceClientId == connectionId {
				continue
			}
			sendJSON(c, &ClientResponse{Count: len(s.clientInfoList)})
			if s.clientIndexMap[connectionId] == 0 {
				sendJSON(c, &StateResponse{State: "EDITOR", InitialDoc: s.fileStore.Read()})
			} else {
				sendJSON(c, &ReaderResponse{Status: "OK", Doc: s.fileStore.Read()})
			}

		case err := <-errCh:
			logWithPrefix(connectionId, requestNumber, fmt.Sprintf("read %s: %v", clientAddr, err))
			return
		}

		requestNumber++
	}
}

func loggingPrefix(connectionId string, requestNumber int) string {
	return "[" + connectionId + " - " + fmt.Sprintf("%d", requestNumber) + "]"
}

func logWithPrefix(connectionId string, requestNumber int, message string) {
	prefix := loggingPrefix(connectionId, requestNumber)
	log.Printf("%s %s", prefix, message)
}
