package handler

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iamkahvi/text_editor_server/internal/broker"
	"github.com/iamkahvi/text_editor_server/internal/storage"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type HandlerState struct {
	dmp       *diffmatchpatch.DiffMatchPatch
	fileStore storage.FileStore
	upgrader  websocket.Upgrader
	broker    broker.Broker[Broadcast]
	// the string in both these cases is a uuid
	mu             *sync.Mutex
	clientIndexMap map[string]int
	clientInfoList []ClientState
}

type ClientStateState int

const (
	EDITOR ClientStateState = iota
	READER
)

type ClientState struct {
	id             string
	state          ClientStateState
	connectionTime time.Time
}

func (s *HandlerState) addClient(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var state ClientStateState
	if (len(s.clientInfoList)) == 0 {
		state = EDITOR
	} else {
		state = READER
	}

	s.clientIndexMap[id] = len(s.clientInfoList)
	s.clientInfoList = append(s.clientInfoList, ClientState{
		id:             id,
		state:          state,
		connectionTime: time.Now(),
	})
}

func (s *HandlerState) removeClient(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, ok := s.clientIndexMap[id]
	if !ok {
		return
	}

	delete(s.clientIndexMap, id)

	copy(s.clientInfoList[index:], s.clientInfoList[index+1:])
	s.clientInfoList = s.clientInfoList[:len(s.clientInfoList)-1]

	// update indices in the map
	for i := index; i < len(s.clientInfoList); i++ {
		s.clientIndexMap[s.clientInfoList[i].id] = i
	}
}

func NewHandlerState(
	dmp *diffmatchpatch.DiffMatchPatch,
	fileStore storage.FileStore,
	upgrader websocket.Upgrader,
	broker broker.Broker[Broadcast],
) *HandlerState {
	mu := &sync.Mutex{}
	clientIndexMap := make(map[string]int)
	clientInfoList := make([]ClientState, 0)

	return &HandlerState{
		dmp:            dmp,
		fileStore:      fileStore,
		upgrader:       upgrader,
		broker:         broker,
		mu:             mu,
		clientIndexMap: clientIndexMap,
		clientInfoList: clientInfoList,
	}
}
