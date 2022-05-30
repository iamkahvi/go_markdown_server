package main

import (
	"flag"
	"io/ioutil"
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
	Type *Type  `json:"type"`
	Data string `json:"data"`
}

type ServerMessage struct {
	Status Status `json:"status"`
	Type   Type   `json:"type"`
	Data   string `json:"data"`
}

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
	for {
		var m ClientMessage
		err := c.ReadJSON(&m)

		if err != nil {
			log.Println("read:", err)
			break
		}

		log.Printf("THE MESSAGE%v", m.Type)

		if m.Type == nil {
			writeToFile([]byte(m.Data))

			resp := ServerMessage{Status: Leader}

			err = c.WriteJSON(resp)
			if err != nil {
				log.Println("write:", err)
				break
			}
		} else if *m.Type == First {
			resp := ServerMessage{Status: Leader, Type: First, Data: readFile()}

			err = c.WriteJSON(resp)
			if err != nil {
				log.Println("write:", err)
				break
			}
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
