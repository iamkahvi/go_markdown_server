package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

const FILENAME = "output.md"

var addr = flag.String("addr", "0.0.0.0:8000", "http service address")

var isDev = os.Getenv("DEV")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		log.Printf("%v", r.Header.Get("Origin"))
		if isDev == "1" {
			return true
		}
		return r.Header.Get("Origin") == "https://write.kahvipatel.com"
	},
}

type Status string

const (
	Ok    Status = "OK"
	Error Status = "ERROR"
)

type PatchType int

const (
	Remove PatchType = -1
	None   PatchType = 0
	Add    PatchType = 1
)

type Patch struct {
	Type  PatchType `json:"type"`  // -1 | 0 | 1
	Value string    `json:"value"` // string
}

func (p *Patch) UnmarshalJSON(data []byte) error {
	// expect: [number, string]
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) != 2 {
		return fmt.Errorf("patch must have 2 elements")
	}

	if err := json.Unmarshal(raw[0], &p.Type); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[1], &p.Value); err != nil {
		return err
	}
	return nil
}

type Message struct {
	Patches []Patch `json:"patches"`
}

type MyResponse struct {
	Status string `json:"status"`        // "OK" | "ERROR"
	Doc    string `json:"doc,omitempty"` // optional
}

type HandlerState struct {
	num_clients int
}

func (s *HandlerState) write(w http.ResponseWriter, r *http.Request) {
	_, err := os.Open(FILENAME)
	if os.IsNotExist(err) {
		log.Printf("creating %s", FILENAME)
		_, err = os.Create(FILENAME)
		if err != nil {
			log.Fatalf("Can't open the file")
		}
	}

	clientAddr := r.RemoteAddr
	log.Printf("websocket upgrade requested from %s", clientAddr)

	c, err := upgrader.Upgrade(w, r, nil)
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
			log.Printf("read %s: %v", clientAddr, err)
			break
		}

		s.num_clients++

		log.Printf("message %d received from %s", len(m.Patches), clientAddr)
		log.Printf("total connected clients: %d", s.num_clients)

		if len(m.Patches) == 0 {
			resp := MyResponse{Status: string(Ok), Doc: readFile()}
			log.Printf("respond with full file %s", readFile())
			if err := c.WriteJSON(resp); err != nil {
				log.Printf("write %s: %v", clientAddr, err)
				return
			}
		}

		var doc_string string

		if len(m.Patches) > 1 {
			doc_string = constructDocString(m.Patches)
			log.Printf("writing %s to file from %s", doc_string, clientAddr)
			writeToFile([]byte(doc_string))
			resp := MyResponse{Status: string(Ok)}
			if err := c.WriteJSON(resp); err != nil {
				log.Printf("write %s: %v", clientAddr, err)
				return
			}
		}
	}
}

func constructDocString(patches []Patch) string {
	var doc_string string
	for _, patch := range patches {
		switch patch.Type {
		case Add:
			doc_string += patch.Value
			log.Printf("adding %s", patch.Value)
		case None:
			doc_string += patch.Value
			log.Printf("removing %s", patch.Value)
		case Remove:
			log.Printf("leaving %s", patch.Value)
		}
	}
	return doc_string
}

func writeToFile(value []byte) {
	err := os.WriteFile(FILENAME, value, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func readFile() string {
	data, err := os.ReadFile(FILENAME)
	if err != nil {
		log.Fatalf("error reading the file")
	}
	return string(data)
}

func home(w http.ResponseWriter, r *http.Request) {
	log.Printf("http %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	w.Write([]byte("This is a websocket server"))
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	s := &HandlerState{}

	http.HandleFunc("/write", s.write)
	http.HandleFunc("/", home)
	log.Printf("starting server on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
