package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

type Event interface {
	isEvent()
}

type UpdatedEvent struct {
	Tabs []Tab `json:"tabs"`
}

func (UpdatedEvent) isEvent() {}

var eventRegistry = map[string]func() Event{
	"updated": func() Event { return &UpdatedEvent{} },
}

func RecvEvent(r io.Reader) (Event, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint32(lenBuf)

	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	var header struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(buf, &header); err != nil {
		return nil, err
	}

	ctor, ok := eventRegistry[header.Type]
	if !ok {
		return nil, fmt.Errorf("unknown event type: %s", header.Type)
	}

	evPtr := ctor()
	if err := json.Unmarshal(buf, evPtr); err != nil {
		return nil, err
	}

	// Convert pointer to value
	switch e := evPtr.(type) {
	case *UpdatedEvent:
		return *e, nil
	default:
		return nil, fmt.Errorf("unexpected event type: %T", evPtr)
	}
}
