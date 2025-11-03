package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/iamkahvi/text_editor_server/internal/diff"
	"github.com/iamkahvi/text_editor_server/internal/storage"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const FILENAME = "output.md"

type HandlerState struct {
	num_clients int
	dmp         *diffmatchpatch.DiffMatchPatch
	file_store  storage.FileStore
	upgrader    websocket.Upgrader
}

func NewHandlerState(dmp *diffmatchpatch.DiffMatchPatch, num_clients int, file_store storage.FileStore, upgrader websocket.Upgrader) *HandlerState {
	return &HandlerState{num_clients: num_clients, dmp: dmp, file_store: file_store, upgrader: upgrader}
}

func (s *HandlerState) Write(w http.ResponseWriter, r *http.Request) {
	_, err := os.Open(s.file_store.FilePath)
	if os.IsNotExist(err) {
		log.Printf("creating %s", FILENAME)
		_, err = os.Create(FILENAME)
		if err != nil {
			log.Fatalf("Can't open the file")
		}
	}

	clientAddr := r.RemoteAddr
	log.Printf("websocket upgrade requested from %s", clientAddr)

	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade %s: %v", clientAddr, err)
		return
	}

	defer func() {
		c.Close()
		s.num_clients--
	}()

	for {
		var m Message
		err := c.ReadJSON(&m)

		if err != nil {
			log.Printf("error marshalling json %v", err)
			break
		}

		s.num_clients++

		// log.Printf("message %d received from %s", len(m.Patches), clientAddr)
		// log.Printf("patch objects %s", m.PatchObjs)
		// convert PatchObjs to library Patch type

		// result, _ := s.dmp.PatchApply(libPatches, readFile())
		// log.Printf("patch apply result: %v", result)

		if len(m.Patches) == 0 {
			resp := MyResponse{Status: "OK", Doc: s.file_store.Read()}
			if err := c.WriteJSON(resp); err != nil {
				return
			}
		}

		var doc_string string

		if len(m.Patches) > 1 {
			doc_string = diff.ConstructDocString(m.Patches)
			s.file_store.Write([]byte(doc_string))
			resp := MyResponse{Status: "OK"}
			if err := c.WriteJSON(resp); err != nil {
				log.Printf("write %s: %v", clientAddr, err)
				return
			}
		}
	}
}
