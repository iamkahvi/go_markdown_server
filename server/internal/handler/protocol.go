package handler

import (
	"encoding/json"
	"fmt"

	"github.com/iamkahvi/text_editor_server/internal/diff"
)

type Message struct {
	Patches   []diff.Patch    `json:"patches"`
	PatchObjs []diff.PatchObj `json:"patchObjs"`
}

func (m Message) String() string {
	payload := struct {
		Patches   []diff.Patch    `json:"patches"`
		PatchObjs []diff.PatchObj `json:"patchObjs"`
	}{
		Patches:   m.Patches,
		PatchObjs: m.PatchObjs,
	}

	b, err := json.MarshalIndent(payload, "", "    ")
	if err != nil {
		return fmt.Sprintf("Message{error: %v}", err)
	}

	return string(b)
}

// the shape of the incoming message is:
// {
//     "patches": [],
//     "patchObjs": []
// }

func (m *Message) UnmarshalJSON(data []byte) error {
	type alias struct {
		Patches   []diff.Patch    `json:"patches"`
		PatchObjs []diff.PatchObj `json:"patchObjs"`
	}

	var payload alias
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("unmarshal message: %w", err)
	}

	if payload.Patches == nil {
		return fmt.Errorf("message must contain patches field")
	}

	if payload.PatchObjs == nil {
		return fmt.Errorf("message must contain patchObjs field")
	}

	m.Patches = payload.Patches
	m.PatchObjs = payload.PatchObjs

	return nil
}

type MyResponse struct {
	Status string `json:"status"`
	Doc    string `json:"doc,omitempty"`
}
