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
	numClients int
	dmp        *diffmatchpatch.DiffMatchPatch
	fileStore  storage.FileStore
	upgrader   websocket.Upgrader
	broker     broker.Broker[Broadcast]
	// the string in both these cases is a uuid
	mu             sync.Mutex
	clientIndexMap map[string]int
	clientInfoMap  []ClientState
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
	if (len(s.clientInfoMap)) == 0 {
		state = EDITOR
	} else {
		state = READER
	}

	s.clientIndexMap[id] = len(s.clientInfoMap)
	s.clientInfoMap = append(s.clientInfoMap, ClientState{
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

	copy(s.clientInfoMap[index:], s.clientInfoMap[index+1:])
	s.clientInfoMap = s.clientInfoMap[:len(s.clientInfoMap)-1]

	// update indices in the map
	for i := index; i < len(s.clientInfoMap); i++ {
		s.clientIndexMap[s.clientInfoMap[i].id] = i
	}
}

func NewHandlerState(
	dmp *diffmatchpatch.DiffMatchPatch,
	numClients int,
	fileStore storage.FileStore,
	upgrader websocket.Upgrader,
	broker broker.Broker[Broadcast],
) *HandlerState {
	return &HandlerState{numClients: numClients, dmp: dmp, fileStore: fileStore, upgrader: upgrader, broker: broker}
}
