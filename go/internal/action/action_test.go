package action

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"
)

func TestSendAction(t *testing.T) {
	var buf bytes.Buffer

	cmd := &SelectAction{TabID: 42}

	err := SendAction(&buf, cmd)
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

	if result["command"] != "select" {
		t.Errorf("unexpected command: got %v, want %v", result["command"], "select")
	}
	if result["tabId"] != float64(42) { // json.Unmarshal parses numbers as float64
		t.Errorf("unexpected tabId: got %v, want %v", result["tabId"], 42)
	}
}
