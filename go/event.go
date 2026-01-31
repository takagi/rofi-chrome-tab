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

func unmarshalEvent[T Event](buf []byte) (Event, error) {
	var e T
	if err := json.Unmarshal(buf, &e); err != nil {
		return nil, err
	}
	return e, nil
}

var eventUnmarshalerRegistry = map[string]func([]byte) (Event, error){
	"updated": unmarshalEvent[UpdatedEvent],
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

	unmarshaler, ok := eventUnmarshalerRegistry[header.Type]
	if !ok {
		return nil, fmt.Errorf("unknown event type: %s", header.Type)
	}
	return unmarshaler(buf)
}
