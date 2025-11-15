package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/iamkahvi/text_editor_server/config"
	"github.com/iamkahvi/text_editor_server/internal/broker"
	"github.com/iamkahvi/text_editor_server/internal/handler"
	"github.com/iamkahvi/text_editor_server/internal/storage"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var isDev = os.Getenv("DEV")

func main() {
	flag.Parse()
	log.SetFlags(0)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			log.Printf("%v", r.Header.Get("Origin"))
			if isDev == "1" {
				return true
			}
			return r.Header.Get("Origin") == cfg.Origin
		},
	}

	fileStore := storage.NewFileStore(cfg.DocumentPath)
	dmp := diffmatchpatch.New()
	broker := broker.NewBroker[handler.Broadcast]()
	go broker.Start()
	defer broker.Stop()
	s := handler.NewHandlerState(dmp, *fileStore, upgrader, *broker)

	http.HandleFunc("/write", s.Write)
	http.HandleFunc("/", s.Home)
	log.Printf("starting server on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, nil))
}
