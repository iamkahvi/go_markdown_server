package handler

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/iamkahvi/text_editor_server/internal/broker"
	"github.com/iamkahvi/text_editor_server/internal/storage"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type HandlerState struct {
	numClients int
	dmp        *diffmatchpatch.DiffMatchPatch
	fileStore  storage.FileStore
	upgrader   websocket.Upgrader
	broker     *broker.Broker[Broadcast]
}

func NewHandlerState(
	dmp *diffmatchpatch.DiffMatchPatch,
	numClients int,
	fileStore storage.FileStore,
	upgrader websocket.Upgrader,
) *HandlerState {
	return &HandlerState{numClients: numClients, dmp: dmp, fileStore: fileStore, upgrader: upgrader, broker: broker.NewBroker[Broadcast]()}
}

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

	connCh := make(chan Message, 1) // drain ws reads
	errCh := make(chan error, 1)    // propagate read errors

	defer func() {
		log.Printf("closing the connection")
		c.Close()
		cancel()
		close(connCh)
		close(errCh)
		s.numClients--
	}()

	if err != nil {
		log.Printf("upgrade %s: %v", clientAddr, err)
		return
	}
	s.numClients++

	go func() {
		for {
			var m Message
			err := c.ReadJSON(&m)
			if err != nil {
				errCh <- err
				return
			}
			select {
			case connCh <- m:
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	for {
		select {
		case m := <-connCh:
			payload, err := MarshalMyResponse(&ClientResponse{Count: s.numClients})
			if err != nil {
				log.Printf("marshal response: %v", err)
				return
			}
			if err := c.WriteJSON(payload); err != nil {
				return
			}

			log.Printf("message %v", m)
			log.Printf("file %v", s.fileStore.Read())

			// convert PatchObjs to library Patch type
			dmpPatches := make([]diffmatchpatch.Patch, 0, len(m.PatchObjs))
			for _, po := range m.PatchObjs {
				dmpPatches = append(dmpPatches, po.ToDMP(s.dmp))
			}

			result, _ := s.dmp.PatchApply(dmpPatches, s.fileStore.Read())
			log.Printf("patch apply result: %v", result)

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
			}
		case err := <-errCh:
			log.Printf("read %s: %v", clientAddr, err)
			return
		}
	}
}
