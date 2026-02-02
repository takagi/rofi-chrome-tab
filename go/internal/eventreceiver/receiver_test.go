package eventreceiver

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
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	evCh := make(chan event.Event, 1)

	tabs := []types.Tab{
		{ID: 1, Title: "Test Tab", Host: "example.com"},
	}
	ev := struct {
		Type string      `json:"type"`
		Tabs []types.Tab `json:"tabs"`
	}{
		Type: "updated",
		Tabs: tabs,
	}

	jsonData, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	Start(r, evCh)

	length := uint32(len(jsonData))
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], length)

	if _, err := w.Write(lenBuf[:]); err != nil {
		t.Fatalf("Failed to write length header: %v", err)
	}
	if _, err := w.Write(jsonData); err != nil {
		t.Fatalf("Failed to write JSON data: %v", err)
	}

	select {
	case got := <-evCh:
		updatedEv, ok := got.(event.UpdatedEvent)
		if !ok {
			t.Fatalf("Expected UpdatedEvent, got %T", got)
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
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()

	evCh := make(chan event.Event, 1)

	Start(r, evCh)

	w.Close()
	time.Sleep(100 * time.Millisecond)
}

func TestStartEventReceiver_MessageTooLarge(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	evCh := make(chan event.Event, 1)

	Start(r, evCh)

	const maxMessageSize = 10 * 1024 * 1024
	length := uint32(maxMessageSize + 1)
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], length)

	if _, err := w.Write(lenBuf[:]); err != nil {
		t.Fatalf("Failed to write length header: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	select {
	case got := <-evCh:
		t.Fatalf("Expected no event due to oversized message, got %T", got)
	default:
	}
}

func TestStartEventReceiver_InvalidJSON(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	evCh := make(chan event.Event, 1)

	Start(r, evCh)

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

	time.Sleep(100 * time.Millisecond)

	select {
	case got := <-evCh:
		t.Fatalf("Expected no event due to invalid JSON, got %T", got)
	default:
	}

	tabs := []types.Tab{{ID: 2, Title: "Valid Tab", Host: "example.org"}}
	ev := struct {
		Type string      `json:"type"`
		Tabs []types.Tab `json:"tabs"`
	}{
		Type: "updated",
		Tabs: tabs,
	}

	jsonData, err := json.Marshal(ev)
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

	select {
	case got := <-evCh:
		if _, ok := got.(event.UpdatedEvent); !ok {
			t.Fatalf("Expected UpdatedEvent, got %T", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for valid event after invalid JSON")
	}
}

func TestStartEventReceiver_PartialRead(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte{0x01, 0x02})

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	evCh := make(chan event.Event, 1)

	Start(r, evCh)

	if _, err := w.Write([]byte{0x01, 0x02}); err != nil {
		t.Fatalf("Failed to write partial data: %v", err)
	}

	w.Close()

	time.Sleep(100 * time.Millisecond)

	select {
	case got := <-evCh:
		t.Fatalf("Expected no event due to incomplete read, got %T", got)
	default:
	}
}

func TestStartEventReceiver_MultipleEvents(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	evCh := make(chan event.Event, 10)

	Start(r, evCh)

	for i := 1; i <= 3; i++ {
		tabs := []types.Tab{{ID: i, Title: "Tab " + strconv.Itoa(i), Host: "example.com"}}
		ev := struct {
			Type string      `json:"type"`
			Tabs []types.Tab `json:"tabs"`
		}{
			Type: "updated",
			Tabs: tabs,
		}

		jsonData, err := json.Marshal(ev)
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

	for i := 1; i <= 3; i++ {
		select {
		case got := <-evCh:
			updatedEv, ok := got.(event.UpdatedEvent)
			if !ok {
				t.Fatalf("Event %d: Expected UpdatedEvent, got %T", i, got)
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
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	evCh := make(chan event.Event, 1)

	Start(r, evCh)

	length := uint32(0)
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], length)

	if _, err := w.Write(lenBuf[:]); err != nil {
		t.Fatalf("Failed to write length header: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	select {
	case got := <-evCh:
		t.Fatalf("Expected no event for empty message, got %T", got)
	default:
	}
}
