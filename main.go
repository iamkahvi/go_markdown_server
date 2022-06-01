package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"hash/fnv"

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

type Queue[T uint32 | bool] interface {
	Head() T
	Push(T)
	Pop() (T, bool)
	Remove(T)
}

type MyQueue struct {
	q    []uint32
	lock sync.RWMutex
}

type Value interface {
	// Somehow VScode won't let me define generics
	uint32 | bool
}

func (mq *MyQueue) Head() (uint32, bool) {
	temp, empty := mq.Pop()
	if empty {
		return 0, false
	}
	mq.Push(temp)
	return temp, true
}

func (mq *MyQueue) Push(val uint32) {
	mq.lock.RLock()
	mq.q = append(mq.q, val)
	mq.lock.Unlock()
}

func (mq *MyQueue) Pop() (uint32, bool) {
	mq.lock.RLock()
	if len(mq.q) == 0 {
		return 0, false
	}

	var temp = mq.q[len(mq.q)-1]
	mq.q = mq.q[:len(mq.q)-1]
	mq.lock.Unlock()
	return temp, true
}

func (mq *MyQueue) Remove(val uint32) bool {
	mq.lock.RLock()
	defer mq.lock.Unlock()

	for i, v := range mq.q {
		if v == val {
			mq.q = append(mq.q[:i], mq.q[i+1:]...)
			return true
		}
	}

	return false
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

var actualQueue = MyQueue{q: make([]uint32, 100), lock: sync.RWMutex{}}

func write(w http.ResponseWriter, r *http.Request) {
	_, err := os.Open(FILENAME)
	if os.IsNotExist(err) {
		_, err = os.Create(FILENAME)
		if err != nil {
			log.Fatalf("Can't open the file")
		}
	}

	c, err := upgrader.Upgrade(w, r, nil)

	var clientId = hash(c.RemoteAddr().String())

	c.SetCloseHandler(func(code int, text string) error {
		actualQueue.Remove(clientId)
		return nil
	})

	log.Printf("%v", c.RemoteAddr())
	actualQueue.Push(clientId)

	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	log.Printf("the queue: %v", actualQueue.q)

	defer c.Close()
	for {
		var m ClientMessage
		err := c.ReadJSON(&m)

		if err != nil {
			log.Println("read:", err)
			break
		}

		log.Printf("THE MESSAGE%v", m.Type)

		head, empty := actualQueue.Head()

		if m.Type == nil && head == clientId {
			writeToFile([]byte(m.Data))

			resp := ServerMessage{Status: Follower}

			err = c.WriteJSON(resp)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}

		if empty || *m.Type == First && head == clientId {
			resp := ServerMessage{Status: Leader, Type: First, Data: readFile()}

			err = c.WriteJSON(resp)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}

		if head != clientId {
			resp := ServerMessage{Status: Follower, Data: readFile()}

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
