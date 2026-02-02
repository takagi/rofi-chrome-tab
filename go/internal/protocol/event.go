package protocol

import (
	"encoding/json"
	"fmt"
)

type Event interface {
	isEvent()
}

type UpdatedEvent struct {
	Tabs []Tab `json:"tabs"`
}

func (UpdatedEvent) isEvent() {}

func unmarshalEvent[T Event](buf []byte) (Event, error) {
	var e T
	if err := json.Unmarshal(buf, &e); err != nil {
		return nil, err
	}
	return e, nil
}

func ParseEvent(buf []byte) (Event, error) {
	var header struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(buf, &header); err != nil {
		return nil, err
	}

	switch header.Type {
	case "updated":
		return unmarshalEvent[UpdatedEvent](buf)
	default:
		return nil, fmt.Errorf("unknown event type: %s", header.Type)
	}
}
