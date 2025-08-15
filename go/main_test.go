package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"
)

func TestSendCommand(t *testing.T) {
	var buf bytes.Buffer

	cmd := &SelectCommand{TabID: 42}

	err := SendCommand(&buf, cmd)
	if err != nil {
		t.Fatalf("SendCommand failed: %v", err)
	}

	var length uint32
	if err := binary.Read(&buf, binary.LittleEndian, &length); err != nil {
		t.Fatalf("failed to read length prefix: %v", err)
	}

	payload := make([]byte, length)
	if _, err := buf.Read(payload); err != nil {
		t.Fatalf("failed to read JSON payload: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if result["type"] != "select" {
		t.Errorf("unexpected type: got %v, want %v", result["type"], "select")
	}
	if result["tabId"] != float64(42) { // json.Unmarshal parses numbers as float64
		t.Errorf("unexpected tabId: got %v, want %v", result["tabId"], 42)
	}
}

func TestRecvEvent(t *testing.T) {
	tabs := []Tab{
		{ID: 1, Title: "Tab1", Host: "example.com"},
		{ID: 2, Title: "Tab2", Host: "example.org"},
	}

	event := struct {
		Type string `json:"type"`
		Tabs []Tab `json:"tabs"`
	}{
		Type: "updated",
		Tabs: tabs,
	}

	// Encode the test data as JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// Create a buffer simulating the input stream
	var buf bytes.Buffer

	// Write 4-byte length header in little-endian format
	length := uint32(len(jsonData))
	if err := binary.Write(&buf, binary.LittleEndian, length); err != nil {
		t.Fatalf("Failed to write length: %v", err)
	}

	// Write the JSON-encoded message body
	if _, err := buf.Write(jsonData); err != nil {
		t.Fatalf("Failed to write JSON data: %v", err)
	}

	// Call RecvEvent to parse the binary input
	got, err := RecvEvent(&buf)
	if err != nil {
		t.Fatalf("RecvEvent failed: %v", err)
	}

	// Assert returned value is *UpdatedEvent
	ev, ok := got.(*UpdatedEvent)
	if !ok {
		t.Fatalf("Expected *UpdatedEvent, got %T", got)
	}

	// Compare the contents
	if len(ev.Tabs) != len(tabs) {
		t.Fatalf("Expected %d tabs, got %d", len(tabs), len(ev.Tabs))
	}
	for i := range tabs {
		if ev.Tabs[i] != tabs[i] {
			t.Errorf("Tab mismatch at index %d: expected %+v, got %+v", i, tabs[i], ev.Tabs[i])
		}
	}
}
