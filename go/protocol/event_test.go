package protocol

import (
	"encoding/json"
	"testing"
)

func TestParseEvent(t *testing.T) {
	tabs := []Tab{
		{ID: 1, Title: "Tab1", Host: "example.com"},
		{ID: 2, Title: "Tab2", Host: "example.org"},
	}

	event := struct {
		Type string `json:"type"`
		Tabs []Tab  `json:"tabs"`
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
