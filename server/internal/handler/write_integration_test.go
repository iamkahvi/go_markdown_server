package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/iamkahvi/notepad/server/internal/broker"
	"github.com/iamkahvi/notepad/server/internal/storage"
)

// setupTestServer creates a fully wired HandlerState and exposes /write.
func setupTestServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()

	mux := http.NewServeMux()

	dmp := diffmatchpatch.New()
	tmpFile := os.TempDir() + "/testdoc.txt"
	fileStore := storage.NewFileStore(tmpFile)
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	br := broker.NewBroker[Broadcast]()

	state := NewHandlerState(dmp, *fileStore, upgrader, *br)
	mux.HandleFunc("/write", state.Write)

	srv := httptest.NewServer(mux)
	wsURL := "ws" + srv.URL[len("http"):]
	return srv, wsURL + "/write"
}

// connectWS opens a websocket connection.
func connectWS(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", url, err)
	}
	return c
}

// readJSON reads a single JSON message from websocket
func readJSON(t *testing.T, c *websocket.Conn) map[string]any {
	t.Helper()
	var m map[string]any
	if err := c.ReadJSON(&m); err != nil {
		t.Fatalf("read json: %v", err)
	}
	return m
}

// basic placeholder test ensures scaffolding loads.
func TestFirstClientIsEditor(t *testing.T) {
	srv, wsURL := setupTestServer(t)
	defer srv.Close()
	conn := connectWS(t, wsURL)
	defer conn.Close()
	// trigger initial state
	t.Log("sending initial trigger message")
	if err := conn.WriteJSON(map[string]any{"patches": []any{}, "patchObjs": []any{}}); err != nil {
		t.Fatalf("write err: %v", err)
	}
	t.Log("waiting for initial response...")
	msg := readJSON(t, conn)
	if msg["type"] != "state" {
		t.Fatalf("expected type=state, got %v", msg)
	}
	if msg["state"] != "EDITOR" {
		t.Fatalf("expected EDITOR, got %v", msg)
	}
}

func TestIntegrationSetup(t *testing.T) {
	srv, wsURL := setupTestServer(t)
	defer srv.Close()

	conn := connectWS(t, wsURL)
	conn.Close()
}
