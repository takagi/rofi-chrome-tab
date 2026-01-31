package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

type Action interface {
	Type() string
}

type SelectAction struct {
	TabID int `json:"tabId"`
}

func (a SelectAction) Type() string {
	return "select"
}

func SendAction(w io.Writer, a Action) error {
	payload, err := json.Marshal(a)
	if err != nil {
		return err
	}

	var body map[string]interface{}
	if err := json.Unmarshal(payload, &body); err != nil {
		return err
	}

	body["command"] = a.Type()

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	length := uint32(len(data))
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], length)

	if _, err := w.Write(lenBuf[:]); err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}
