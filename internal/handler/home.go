package handler

import (
	"log"
	"net/http"
)

func (s *HandlerState) Home(w http.ResponseWriter, r *http.Request) {
	log.Printf("http %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	w.Write([]byte("This is a websocket server"))
}
