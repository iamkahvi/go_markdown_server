package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

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

type Type string

const (
	First     Type = "first"
	Normal    Type = "normal"
	Increment Type = "increment"
	Counter   Type = "counter"
)

type Status string

const (
	Success Status = "success"
	Error   Status = "error"
)

type ClientMessage struct {
	Type   Type   `json:"type"`
	Status Status `json:"status"`
	Data   string `json:"data"`
}

type ServerMessage struct {
	Type   Type   `json:"type"`
	Status Status `json:"status"`
	Data   string `json:"data"`
}

type counterHub struct {
	mu      sync.Mutex
	counter int
	clients map[*websocket.Conn]bool
}

func newCounterHub() *counterHub {
	return &counterHub{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (h *counterHub) Register(conn *websocket.Conn) int {
	h.mu.Lock()
	h.clients[conn] = true
	count := h.counter
	h.mu.Unlock()
	return count
}

func (h *counterHub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
}

func (h *counterHub) Increment(delta int) {
	h.mu.Lock()
	h.counter += delta
	count := h.counter
	clients := make([]*websocket.Conn, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.Unlock()

	msg := ServerMessage{Status: Success, Type: Counter, Data: strconv.Itoa(count)}
	for _, client := range clients {
		if err := client.WriteJSON(msg); err != nil {
			log.Println("broadcast:", err)
		}
	}
}

var sharedCounter = newCounterHub()

func write(w http.ResponseWriter, r *http.Request) {
	_, err := os.Open(FILENAME)
	if os.IsNotExist(err) {
		_, err = os.Create(FILENAME)
		if err != nil {
			log.Fatalf("Can't open the file")
		}
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer c.Close()
	initialCount := sharedCounter.Register(c)
	defer sharedCounter.Unregister(c)

	if err := c.WriteJSON(ServerMessage{Status: Success, Type: Counter, Data: strconv.Itoa(initialCount)}); err != nil {
		log.Println("write:", err)
	}

	for {
		var m ClientMessage
		err := c.ReadJSON(&m)

		if err != nil {
			log.Println("read:", err)
			break
		}

		switch m.Type {
		case First:
			resp := ServerMessage{Status: Success, Type: First, Data: readFile()}
			if err := c.WriteJSON(resp); err != nil {
				log.Println("write:", err)
				return
			}
		case Normal:
			writeToFile([]byte(m.Data))
			resp := ServerMessage{Status: Success, Type: Normal, Data: ""}
			if err := c.WriteJSON(resp); err != nil {
				log.Println("write:", err)
				return
			}
		case Increment:
			delta := 1
			trimmed := strings.TrimSpace(m.Data)
			if trimmed != "" {
				value, parseErr := strconv.Atoi(trimmed)
				if parseErr != nil {
					errResp := ServerMessage{Status: Error, Type: Increment, Data: "invalid increment value"}
					if err := c.WriteJSON(errResp); err != nil {
						log.Println("write:", err)
					}
					continue
				}
				delta = value
			}
			sharedCounter.Increment(delta)
		default:
			errResp := ServerMessage{Status: Error, Type: m.Type, Data: "unsupported message type"}
			if err := c.WriteJSON(errResp); err != nil {
				log.Println("write:", err)
				return
			}
		}
	}
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
	w.Write([]byte("This is a websocket server"))
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/write", write)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
