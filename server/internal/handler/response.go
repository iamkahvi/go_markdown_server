package handler

import (
	"fmt"
)

type MyResponse interface {
	isMyResponse()
}

type EditorResponse struct {
	Status string `json:"status"`
	Doc    string `json:"doc,omitempty"`
}

func (EditorResponse) isMyResponse() {}

type ClientResponse struct {
	Count int `json:"count"`
}

func (ClientResponse) isMyResponse() {}

type responseEnvelope struct {
	Type   string `json:"type"`
	Status string `json:"status,omitempty"`
	Doc    string `json:"doc,omitempty"`
	Count  int    `json:"count,omitempty"`
}

func MarshalMyResponse(resp MyResponse) (responseEnvelope, error) {
	if resp == nil {
		return responseEnvelope{}, fmt.Errorf("unsupported MyResponse: <nil>")
	}

	switch r := resp.(type) {
	case EditorResponse:
		return responseEnvelope{
			Type:   "editor",
			Status: r.Status,
			Doc:    r.Doc,
		}, nil
	case *EditorResponse:
		return responseEnvelope{
			Type:   "editor",
			Status: r.Status,
			Doc:    r.Doc,
		}, nil
	case ClientResponse:
		return responseEnvelope{
			Type:  "client",
			Count: r.Count,
		}, nil
	case *ClientResponse:
		return responseEnvelope{
			Type:  "client",
			Count: r.Count,
		}, nil
	default:
		return responseEnvelope{}, fmt.Errorf("unsupported MyResponse: %T", resp)
	}
}
