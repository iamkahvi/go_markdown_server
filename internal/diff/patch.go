package diff

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type PatchType int

const (
	Remove PatchType = -1
	None   PatchType = 0
	Add    PatchType = 1
)

type Patch struct {
	Type  PatchType `json:"type"`  // -1 | 0 | 1
	Value string    `json:"value"` // string
}

func (p *Patch) UnmarshalJSON(data []byte) error {
	// expect: [number, string]
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) != 2 {
		return fmt.Errorf("patch must have 2 elements")
	}

	if err := json.Unmarshal(raw[0], &p.Type); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[1], &p.Value); err != nil {
		return err
	}
	return nil
}

type Operation int8

const (
	// DiffDelete item represents a delete diff.
	DiffDelete Operation = -1
	// DiffInsert item represents an insert diff.
	DiffInsert Operation = 1
	// DiffEqual item represents an equal diff.
	DiffEqual Operation = 0
)

// Diff represents one diff operation
type Diff struct {
	Type Operation
	Text string
}

func (d *Diff) UnmarshalJSON(data []byte) error {
	// expect: [number, string]
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) != 2 {
		return fmt.Errorf("patch must have 2 elements")
	}

	if err := json.Unmarshal(raw[0], &d.Type); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[1], &d.Text); err != nil {
		return err
	}
	return nil
}

type PatchObj struct {
	diffs   []Diff `json:"diffs"`
	Start1  *int   `json:"start1"`
	Start2  *int   `json:"start2"`
	Length1 int    `json:"length1"`
	Length2 int    `json:"length2"`
}

func (p PatchObj) String() string {
	var sb strings.Builder
	sb.WriteString("PatchObj {\n")

	sb.WriteString(fmt.Sprintf("  Start1: %v\n", nullableInt(p.Start1)))
	sb.WriteString(fmt.Sprintf("  Start2: %v\n", nullableInt(p.Start2)))
	sb.WriteString(fmt.Sprintf("  Length1: %d\n", p.Length1))
	sb.WriteString(fmt.Sprintf("  Length2: %d\n", p.Length2))
	sb.WriteString("  Diffs:\n")

	for _, d := range p.diffs {
		sb.WriteString(fmt.Sprintf("    [%d, %q]\n", d.Type, d.Text))
	}

	sb.WriteString("}")
	return sb.String()
}

func nullableInt(ptr *int) string {
	if ptr == nil {
		return "null"
	}
	return fmt.Sprintf("%d", *ptr)
}

func (pj *PatchObj) ToDMP(dmp *diffmatchpatch.DiffMatchPatch) diffmatchpatch.Patch {
	// Build a slice of diffs in the library's type.
	diffs := make([]diffmatchpatch.Diff, 0, len(pj.diffs))
	// Also reconstruct text1 as the concatenation of non-insert pieces (preimage)
	var text1Builder strings.Builder
	for _, d := range pj.diffs {
		var t diffmatchpatch.Operation
		switch d.Type {
		case DiffDelete:
			t = diffmatchpatch.DiffDelete
		case DiffInsert:
			t = diffmatchpatch.DiffInsert
		default:
			t = diffmatchpatch.DiffEqual
		}
		diffs = append(diffs, diffmatchpatch.Diff{Type: t, Text: d.Text})

		// text1 is the original text before the patch: include deletes and equals
		if t != diffmatchpatch.DiffInsert {
			text1Builder.WriteString(d.Text)
		}
	}

	text1 := text1Builder.String()
	// Use the library to create a proper Patch (it will compute Start/Length/context)
	patches := dmp.PatchMake(text1, diffs)
	if len(patches) == 0 {
		// return an empty Patch value
		return diffmatchpatch.Patch{}
	}
	// The PatchObj represents a single patch; return the first one.
	return patches[0]
}

// Converter transforms websocket payloads into diffmatchpatch structures.
type Converter struct {
	// TODO: add dependencies or configuration as needed.
}

// Apply converts incoming messages and returns resulting document state.
func (c *Converter) Apply(payload interface{}) (string, error) {
	// TODO: implement diff conversion logic.
	return "", nil
}
