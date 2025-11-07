package handler

import (
	"context"
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
	s.broker.Publish(Broadcast{})

	messageCh := make(chan Message, 1)
	errCh := make(chan error, 1)
	broadcastCh := s.broker.Subscribe()

	defer func() {
		log.Printf("closing the connection")
		s.removeClient(connectionId)
		s.broker.Publish(Broadcast{})
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

	for {
		select {
		case m := <-messageCh:
			// parse the message
			// TODO: the go routine could do this
			payload, err := MarshalMyResponse(&ClientResponse{Count: len(s.clientInfoList)})
			if err != nil {
				log.Printf("marshal response: %v", err)
				return
			}
			if err := c.WriteJSON(payload); err != nil {
				return
			}

			// if there are no patch objs, it's probably the first request
			if len(m.PatchObjs) == 0 {
				payload, err := MarshalMyResponse(&EditorResponse{Status: "OK", Doc: s.fileStore.Read()})
				if err != nil {
					log.Printf("marshal response: %v", err)
					return
				}
				if err := c.WriteJSON(payload); err != nil {
					return
				}
			}

			// if there are patch obj and this client is the editor, we'll apply them to the file
			if len(m.PatchObjs) >= 1 {
				// convert PatchObjs to library Patch type
				dmpPatches := make([]diffmatchpatch.Patch, 0, len(m.PatchObjs))
				for _, po := range m.PatchObjs {
					dmpPatches = append(dmpPatches, po.ToDMP(s.dmp))
				}
				result, _ := s.dmp.PatchApply(dmpPatches, s.fileStore.Read())
				s.fileStore.Write([]byte(result))
				payload, err := MarshalMyResponse(&EditorResponse{Status: "OK"})
				if err != nil {
					log.Printf("marshal response: %v", err)
					return
				}
				if err := c.WriteJSON(payload); err != nil {
					log.Printf("write %s: %v", clientAddr, err)
					return
				}
			}
		case <-broadcastCh:
			log.Printf("got a broadcast: %v", len(s.clientInfoList))
			payload, err := MarshalMyResponse(&ClientResponse{Count: len(s.clientInfoList)})
			if err != nil {
				log.Printf("marshal response: %v", err)
				return
			}
			if err := c.WriteJSON(payload); err != nil {
				return
			}
		case err := <-errCh:
			log.Printf("read %s: %v", clientAddr, err)
			return
		}
	}
}
