package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"
)

func TestSendAction(t *testing.T) {
	tests := []struct {
		name      string
		action    Action
		wantCmd   string
		wantTabID int
	}{
		{"select action", SelectAction{TabID: 42}, "select", 42},
		{"close action", CloseAction{TabID: 99}, "close", 99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := SendAction(&buf, tt.action)
			if err != nil {
				t.Fatalf("SendAction failed: %v", err)
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

			if result["command"] != tt.wantCmd {
				t.Errorf("unexpected command: got %v, want %v", result["command"], tt.wantCmd)
			}
			if result["tabId"] != float64(tt.wantTabID) { // json.Unmarshal parses numbers as float64
				t.Errorf("unexpected tabId: got %v, want %v", result["tabId"], tt.wantTabID)
			}
		})
	}
}
