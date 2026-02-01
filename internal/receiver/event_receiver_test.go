package receiver

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"rofi-chrome-tab/internal/event"
	"rofi-chrome-tab/internal/types"
)

func TestStartEventReceiver_ValidEvent(t *testing.T) {
	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Create local event channel
	evCh := make(chan event.Event, 1)

	// Prepare test event
	tabs := []types.Tab{
		{ID: 1, Title: "Test Tab", Host: "example.com"},
	}
	testEvent := struct {
		Type string      `json:"type"`
		Tabs []types.Tab `json:"tabs"`
	}{
		Type: "updated",
		Tabs: tabs,
	}

	// Marshal event to JSON
	jsonData, err := json.Marshal(testEvent)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// Start the event receiver
	StartEventReceiver(r, evCh)

	// Write length header and message to stdin
	length := uint32(len(jsonData))
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], length)

	if _, err := w.Write(lenBuf[:]); err != nil {
		t.Fatalf("Failed to write length header: %v", err)
	}
	if _, err := w.Write(jsonData); err != nil {
		t.Fatalf("Failed to write JSON data: %v", err)
	}

	// Wait for event to be received
	select {
	case ev := <-evCh:
		updatedEv, ok := ev.(event.UpdatedEvent)
		if !ok {
			t.Fatalf("Expected UpdatedEvent, got %T", ev)
		}
		if len(updatedEv.Tabs) != 1 {
			t.Fatalf("Expected 1 tab, got %d", len(updatedEv.Tabs))
		}
		if updatedEv.Tabs[0].ID != 1 || updatedEv.Tabs[0].Title != "Test Tab" {
			t.Errorf("Tab data mismatch: got %+v", updatedEv.Tabs[0])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

func TestStartEventReceiver_EOF(t *testing.T) {
	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()

	// Create local event channel
	evCh := make(chan event.Event, 1)

	// Start the event receiver
	StartEventReceiver(r, evCh)

	// Close the write end to simulate EOF
	w.Close()

	// Give the goroutine time to detect EOF and exit
	time.Sleep(100 * time.Millisecond)

	// If we got here without hanging, the test passes
	// The receiver should handle EOF gracefully and exit
}

func TestStartEventReceiver_MessageTooLarge(t *testing.T) {
	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Create local event channel
	evCh := make(chan event.Event, 1)

	// Start the event receiver
	StartEventReceiver(r, evCh)

	// Write a length header that exceeds the max message size (10MB)
	const maxMessageSize = 10 * 1024 * 1024
	length := uint32(maxMessageSize + 1)
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], length)

	if _, err := w.Write(lenBuf[:]); err != nil {
		t.Fatalf("Failed to write length header: %v", err)
	}

	// Give the goroutine time to process and reject the message
	time.Sleep(100 * time.Millisecond)

	// The receiver should have exited due to oversized message
	// Channel should be empty
	select {
	case ev := <-evCh:
		t.Fatalf("Expected no event due to oversized message, got %T", ev)
	default:
		// Expected: no event received
	}
}

func TestStartEventReceiver_InvalidJSON(t *testing.T) {
	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Create local event channel
	evCh := make(chan event.Event, 1)

	// Start the event receiver
	StartEventReceiver(r, evCh)

	// Write invalid JSON
	invalidJSON := []byte("not valid json")
	length := uint32(len(invalidJSON))
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], length)

	if _, err := w.Write(lenBuf[:]); err != nil {
		t.Fatalf("Failed to write length header: %v", err)
	}
	if _, err := w.Write(invalidJSON); err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	// Give the goroutine time to process the invalid message
	time.Sleep(100 * time.Millisecond)

	// The receiver should continue running and not send an event
	select {
	case ev := <-evCh:
		t.Fatalf("Expected no event due to invalid JSON, got %T", ev)
	default:
		// Expected: no event received
	}

	// Now send a valid event to verify the receiver is still running
	tabs := []types.Tab{{ID: 2, Title: "Valid Tab", Host: "example.org"}}
	testEvent := struct {
		Type string      `json:"type"`
		Tabs []types.Tab `json:"tabs"`
	}{
		Type: "updated",
		Tabs: tabs,
	}

	jsonData, err := json.Marshal(testEvent)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	length = uint32(len(jsonData))
	binary.LittleEndian.PutUint32(lenBuf[:], length)

	if _, err := w.Write(lenBuf[:]); err != nil {
		t.Fatalf("Failed to write length header: %v", err)
	}
	if _, err := w.Write(jsonData); err != nil {
		t.Fatalf("Failed to write JSON data: %v", err)
	}

	// Should receive the valid event
	select {
	case ev := <-evCh:
		if _, ok := ev.(event.UpdatedEvent); !ok {
			t.Fatalf("Expected UpdatedEvent, got %T", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for valid event after invalid JSON")
	}
}

func TestStartEventReceiver_PartialRead(t *testing.T) {
	// Create a buffer to simulate stdin with incomplete data
	var buf bytes.Buffer

	// Write only 2 bytes of the 4-byte length header
	buf.Write([]byte{0x01, 0x02})

	// Create a pipe
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Create local event channel
	evCh := make(chan event.Event, 1)

	// Start the event receiver
	StartEventReceiver(r, evCh)

	// Write partial length header
	if _, err := w.Write([]byte{0x01, 0x02}); err != nil {
		t.Fatalf("Failed to write partial data: %v", err)
	}

	// Close to trigger EOF mid-read
	w.Close()

	// Give the goroutine time to process and exit gracefully
	time.Sleep(100 * time.Millisecond)

	// Channel should be empty since we never sent a complete message
	select {
	case ev := <-evCh:
		t.Fatalf("Expected no event due to incomplete read, got %T", ev)
	default:
		// Expected: no event received
	}
}

func TestStartEventReceiver_MultipleEvents(t *testing.T) {
	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Create local event channel with larger buffer
	evCh := make(chan event.Event, 10)

	// Start the event receiver
	StartEventReceiver(r, evCh)

	// Send multiple events
	for i := 1; i <= 3; i++ {
		tabs := []types.Tab{{ID: i, Title: "Tab " + strconv.Itoa(i), Host: "example.com"}}
		testEvent := struct {
			Type string      `json:"type"`
			Tabs []types.Tab `json:"tabs"`
		}{
			Type: "updated",
			Tabs: tabs,
		}

		jsonData, err := json.Marshal(testEvent)
		if err != nil {
			t.Fatalf("Failed to marshal event %d: %v", i, err)
		}

		length := uint32(len(jsonData))
		var lenBuf [4]byte
		binary.LittleEndian.PutUint32(lenBuf[:], length)

		if _, err := w.Write(lenBuf[:]); err != nil {
			t.Fatalf("Failed to write length header for event %d: %v", i, err)
		}
		if _, err := w.Write(jsonData); err != nil {
			t.Fatalf("Failed to write JSON data for event %d: %v", i, err)
		}
	}

	// Receive all three events
	for i := 1; i <= 3; i++ {
		select {
		case ev := <-evCh:
			updatedEv, ok := ev.(event.UpdatedEvent)
			if !ok {
				t.Fatalf("Event %d: Expected UpdatedEvent, got %T", i, ev)
			}
			if len(updatedEv.Tabs) != 1 {
				t.Fatalf("Event %d: Expected 1 tab, got %d", i, len(updatedEv.Tabs))
			}
			if updatedEv.Tabs[0].ID != i {
				t.Errorf("Event %d: Expected tab ID %d, got %d", i, i, updatedEv.Tabs[0].ID)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("Timeout waiting for event %d", i)
		}
	}
}

func TestStartEventReceiver_EmptyMessage(t *testing.T) {
	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Create local event channel
	evCh := make(chan event.Event, 1)

	// Start the event receiver
	StartEventReceiver(r, evCh)

	// Write a zero-length message
	length := uint32(0)
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], length)

	if _, err := w.Write(lenBuf[:]); err != nil {
		t.Fatalf("Failed to write length header: %v", err)
	}

	// Give the goroutine time to process
	time.Sleep(100 * time.Millisecond)

	// Should not receive any event (empty message can't be parsed)
	select {
	case ev := <-evCh:
		t.Fatalf("Expected no event for empty message, got %T", ev)
	default:
		// Expected: no event received
	}
}
