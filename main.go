package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

const FILENAME = "output.md"

var isDev = os.Getenv("DEV")

var addr = flag.String("addr", "0.0.0.0:8000", "http service address")

var leader = make(chan int, 100)

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
	First  Type = "first"
	Normal Type = "normal"
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

func write(w http.ResponseWriter, r *http.Request) {
	id := len(leader) + 1
	fmt.Printf("\nnew client %d\n", id)

	// First client in queue:
	isLeader := len(leader) == 0
	if isLeader {
		fmt.Println("found first leader")
	}

	leader <- id

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer c.Close()
	for {
		var m ClientMessage
		err := c.ReadJSON(&m)

		if err != nil {
			log.Println("read:", err)
			break
		}

		// check if client at front of queue
		if !isLeader && len(leader) == id-1 {
			isLeader = true
			fmt.Println("found leader")
			fmt.Printf("id: %d, len leader: %d\n", id, len(leader))
		} else {
			fmt.Printf("id: %d, len leader: %d\n", id, len(leader))
			continue
		}

		if m.Type == First {
			resp := ServerMessage{Status: Success, Type: First, Data: readFile()}

			err = c.WriteJSON(resp)
			if err != nil {
				log.Println("write:", err)
				break
			}
		} else if m.Type == Normal {
			writeToFile([]byte(m.Data))

			resp := ServerMessage{Status: Success, Type: Normal, Data: ""}

			err = c.WriteJSON(resp)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
	fmt.Println("leaving")
	<-leader
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
	fmt.Println("hello")

	// check if file exists
	_, err := os.Open(FILENAME)
	if os.IsNotExist(err) {
		_, err = os.Create(FILENAME)
		if err != nil {
			log.Fatalf("Can't open the file")
		}
	}

	http.HandleFunc("/write", write)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
