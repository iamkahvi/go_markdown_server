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

	messageCh := make(chan Message, 1) // drain ws reads
	errCh := make(chan error, 1)       // propagate read errors

	broadcastCh := s.broker.Subscribe()

	defer func() {
		log.Printf("closing the connection")
		s.numClients--
		s.removeClient(connectionId)
		s.broker.Publish(Broadcast{s.numClients})
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

	s.numClients++
	s.broker.Publish(Broadcast{s.numClients})

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
			payload, err := MarshalMyResponse(&ClientResponse{Count: s.numClients})
			if err != nil {
				log.Printf("marshal response: %v", err)
				return
			}
			if err := c.WriteJSON(payload); err != nil {
				return
			}

			// log.Printf("message %v", m)
			// log.Printf("file %v", s.fileStore.Read())

			// convert PatchObjs to library Patch type
			dmpPatches := make([]diffmatchpatch.Patch, 0, len(m.PatchObjs))
			for _, po := range m.PatchObjs {
				dmpPatches = append(dmpPatches, po.ToDMP(s.dmp))
			}

			result, _ := s.dmp.PatchApply(dmpPatches, s.fileStore.Read())
			// log.Printf("patch apply result: %v", result)

			if len(m.Patches) == 0 {
				payload, err := MarshalMyResponse(&EditorResponse{Status: "OK", Doc: s.fileStore.Read()})
				if err != nil {
					log.Printf("marshal response: %v", err)
					return
				}
				if err := c.WriteJSON(payload); err != nil {
					return
				}
			}

			if len(m.Patches) >= 1 {
				// doc_string := diff.ConstructDocString(m.Patches)
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
				s.broker.Publish(Broadcast{s.numClients})
			}
		case b := <-broadcastCh:
			log.Printf("got a broadcast: %v", b.NumClients)
			payload, err := MarshalMyResponse(&ClientResponse{Count: b.NumClients})
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
