package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8000", "http service address")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
} // use default options

type Status string

const (
	First  Status = "first"
	Normal Status = "normal"
)

type Message struct {
	Status Status `json:"status"`
	Data   string `json:"data"`
}

func write(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("output.md")
	if err != nil {
		f, err = os.Create("output.md")
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
		mt, message, err := c.ReadMessage()

		if err != nil {
			log.Fatalf("Can't read the message")
		}

		var m Message
		err = json.Unmarshal(message, &m)

		if err != nil {
			log.Fatalf("Can't decode the JSON")
		}

		log.Printf("%v", m)

		if err != nil {
			log.Println("read:", err)
			break
		}
		writeToFile(message, f)
		// log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("output.md")
	if err != nil {
		f, err = os.Create("output.md")
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
		var m Message
		err := c.ReadJSON(&m)

		if err != nil {
			log.Println("read:", err)
			break
		}

		log.Printf("%v", m)
		if m.Status == First {
			file := make([]byte, 50)
			f.Read(file)
			c.WriteMessage(websocket.TextMessage, file)
		}

		// writeToFile([]byte(m.Data), f)
		// // log.Printf("recv: %s", message)
		// err = c.WriteMessage(websocket.TextMessage, []byte("goof"))
		// if err != nil {
		// 	log.Println("write:", err)
		// 	break
		// }
	}
}

func writeToFile(value []byte, f *os.File) int {
	n, err := f.WriteAt(value, 0)
	if err != nil {
		log.Fatal("Write fucked up")
	}
	return n
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))
