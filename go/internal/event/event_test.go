package event

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"

	"rofi-chrome-tab/internal/types"
)

func TestParseEvent(t *testing.T) {
	tabs := []types.Tab{
		{ID: 1, Title: "Tab1", Host: "example.com"},
		{ID: 2, Title: "Tab2", Host: "example.org"},
	}

	event := struct {
		Type string      `json:"type"`
		Tabs []types.Tab `json:"tabs"`
	}{
		Type: "updated",
		Tabs: tabs,
	}

	// Encode the test data as JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// Call ParseEvent to parse the JSON data
	got, err := ParseEvent(jsonData)
	if err != nil {
		t.Fatalf("ParseEvent failed: %v", err)
	}

	// Assert returned value is UpdatedEvent
	ev, ok := got.(UpdatedEvent)
	if !ok {
		t.Fatalf("Expected UpdatedEvent, got %T", got)
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

func TestParseEventUnknownType(t *testing.T) {
	payload := []byte(`{"type":"unknown"}`)
	if _, err := ParseEvent(payload); err == nil {
		t.Fatalf("expected error for unknown event type")
	}
}

func TestParseEventInvalidJSON(t *testing.T) {
	if _, err := ParseEvent([]byte("not json")); err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
}

func TestParseEventWithReceiverEncoding(t *testing.T) {
	tabs := []types.Tab{{ID: 1, Title: "Tab1", Host: "example.com"}}
	event := struct {
		Type string      `json:"type"`
		Tabs []types.Tab `json:"tabs"`
	}{
		Type: "updated",
		Tabs: tabs,
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	var buf bytes.Buffer
	length := uint32(len(jsonData))
	if err := binary.Write(&buf, binary.LittleEndian, length); err != nil {
		t.Fatalf("Failed to write length: %v", err)
	}
	if _, err := buf.Write(jsonData); err != nil {
		t.Fatalf("Failed to write JSON data: %v", err)
	}

	lenBuf := make([]byte, 4)
	if _, err := buf.Read(lenBuf); err != nil {
		t.Fatalf("Failed to read length: %v", err)
	}
	payload := make([]byte, length)
	if _, err := buf.Read(payload); err != nil {
		t.Fatalf("Failed to read JSON data: %v", err)
	}

	if _, err := ParseEvent(payload); err != nil {
		t.Fatalf("ParseEvent failed: %v", err)
	}
}
