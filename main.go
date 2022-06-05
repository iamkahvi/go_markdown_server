package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"hash/fnv"

	"github.com/gorilla/websocket"
)

const FILENAME = "output.md"
const FOLLOWER_DELAY = 2 * time.Second

var isDev = os.Getenv("DEV") == "1"

var addr = flag.String("addr", "0.0.0.0:8000", "http service address")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		log.Printf("%v", r.Header.Get("Origin"))
		if isDev {
			return true
		}
		return r.Header.Get("Origin") == "https://write.kahvipatel.com"
	},
}

type Type string

const (
	First Type = "first"
)

type Status string

const (
	Leader   Status = "leader"
	Follower Status = "follower"
)

type ClientMessage struct {
	Type Type   `json:"type"`
	Data string `json:"data"`
}

type ServerMessage struct {
	Status Status `json:"status"`
	Type   Type   `json:"type"`
	Data   string `json:"data"`
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

var huncho = MyQueue[uint32]{q: make([]uint32, 100)}

func write(w http.ResponseWriter, r *http.Request) {
	_, err := os.Open(FILENAME)
	if os.IsNotExist(err) {
		_, err = os.Create(FILENAME)
		if err != nil {
			log.Fatalf("Can't open the file")
		}
	}

	// This logs the connected client's domain
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("ERROR upgrading: ", err)
		return
	}

	var clientId = hash(c.RemoteAddr().String())
	var clientStatus Status

	c.SetCloseHandler(func(code int, text string) error {
		huncho.Remove(clientId)
		return nil
	})

	huncho.Push(clientId)

	log.Printf("the queue: %v", huncho.q)

	defer c.Close()
	for {
		var m ClientMessage
		err := c.ReadJSON(&m)

		if err != nil {
			log.Println("ERROR reading: ", err)
			return
		}

		if m == (ClientMessage{}) {
			log.Println("empty message")
		}

		head, empty := huncho.Head()

		if empty || head == clientId {
			log.Println("client is the leader")
			clientStatus = Leader
		} else {
			log.Println("client is follower")
			clientStatus = Follower
		}

		switch clientStatus {
		case Leader:
			if m.Type == First {
				resp := ServerMessage{Status: Leader, Type: First, Data: readFile()}

				err = c.WriteJSON(resp)
			} else {
				writeToFile([]byte(m.Data))
				resp := ServerMessage{Status: Leader}

				err = c.WriteJSON(resp)
			}

			if err != nil {
				log.Println("ERROR writing: ", err)
				os.Exit(1)
			}

		case Follower:
			resp := ServerMessage{Status: Follower, Data: readFile()}

			err = c.WriteJSON(resp)
			if err != nil {
				log.Println("ERROR writing: ", err)
				os.Exit(1)
			}

			time.Sleep(2 * time.Second)
		}
	}
}

func writeToFile(value []byte) {
	err := ioutil.WriteFile(FILENAME, value, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func readFile() string {
	data, err := ioutil.ReadFile(FILENAME)
	if err != nil {
		log.Fatalf("ERROR reading the file")
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
