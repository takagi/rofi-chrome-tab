package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

// Actions

type Action interface {
	Execute() error
	Type() string
}

type SelectAction struct {
	TabID int `json:"tabId"`
}

func (a *SelectAction) Execute() error {
	return nil
}

func (a *SelectAction) Type() string {
	return "select"
}

// SendAction

func SendAction(w io.Writer, a Action) error {
	payload, err := json.Marshal(a)
	if err != nil {
		return err
	}

	var body map[string]interface{}
	if err := json.Unmarshal(payload, &body); err != nil {
		return err
	}

	body["type"] = a.Type()

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

// Events

type Tab struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Host  string `json:"host"`
}

type Event interface {
	Handle() error
}

var eventRegistry = map[string]func() Event{
	"updated": func() Event { return &UpdatedEvent{} },
}

type UpdatedEvent struct {
	Tabs []Tab `json:"tabs"`
}

func (ev *UpdatedEvent) Handle() error {
	inCh <- ev.Tabs
	return nil
}

// RecvEvent

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

	ev := ctor()
	if err := json.Unmarshal(buf, ev); err != nil {
		return nil, err
	}
	return ev, nil
}

// Commands

func ParseCommand() {
}

// Tabs Store

var (
	inCh  = make(chan []Tab, 1)
	outCh = make(chan chan []Tab, 1)
)

func pull[T any](ch chan chan T) T {
	replyCh := make(chan T, 1)
	ch <- replyCh
	return <-replyCh
}

// Main

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Updated tabs: %v", pull(outCh))
}

func main() {
	// Set up log file
	logFile, err := os.OpenFile("/tmp/rofi_chrome_tab.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open log file:", err)
		os.Exit(1)
	} else {
		log.SetOutput(logFile)
	}

	socketPath := "/tmp/rofi-chrome-tab.sock"

	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatal(err)
	}

	// Receive events in a goroutine
	go func() {
		for {
			ev, err := RecvEvent(os.Stdin)
			if err != nil {
				log.Println("Error receiving message:", err)
				continue
			}
			ev.Handle()
		}
	}()

	// Tabs manager
	go func () {
		copyTabs := func (tabs []Tab) []Tab {
			cp := make([]Tab, len(tabs))
			copy(cp, tabs)
			return cp
		}

		var tabs []Tab
		for {
			select {
			case ts := <-inCh:
				tabs = copyTabs(ts)
				log.Printf("Updated tabs: %v", tabs)
			case replyCh := <-outCh:
				replyCh <- copyTabs(tabs)
			}
		}
	}()

	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}

		go handleConnection(conn)
	}
}
