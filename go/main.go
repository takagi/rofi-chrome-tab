package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

// Actions

type Action interface {
	Type() string
}

type SelectAction struct {
	TabID int `json:"tabId"`
}

func (a SelectAction) Type() string {
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
	tabs = ev.Tabs // TODO: copy
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

type Command interface {
	Execute()
}

type ListCommand struct {}

func (c ListCommand) Execute() {
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for _, tab := range tabs {
		line := fmt.Sprintf("%d,%d,%s,%s", pid, tab.ID, tab.Host, tab.Title)
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "write error: %v\n", err)
			return
		}
	}
}

type SelectCommand struct {
	tabID int
}

func (c SelectCommand) Execute() {
	SendAction(os.Stdout, SelectAction{TabID: c.tabID})
}

func ParseCommand(line string) (Command, error) {
	fields := strings.Fields(line)
    if len(fields) == 0 {
        return nil, fmt.Errorf("empty command")
    }

	switch fields[0] {
	case "list":
		return ListCommand{}, nil
    case "select":
        if len(fields) < 2 {
            return nil, fmt.Errorf("select command requires a TabID")
        }

		tabID, err := strconv.Atoi(fields[1])
        if err != nil {
            return nil, fmt.Errorf("invalid TabID: %s", fields[1])
        }

        return SelectCommand{tabID: tabID}, nil
    default:
        return nil, fmt.Errorf("unknown command: %s", fields[0])
    }
}

// Main

var (
	tabs  []Tab
	evCh  = make(chan Event, 1)
	cmdCh = make(chan Command, 1)

	pid = os.Getpid()
)

func main() {
	// Set up log file
	logFile, err := os.OpenFile("/tmp/rofi_chrome_tab.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open log file:", err)
		os.Exit(1)
	} else {
		log.SetOutput(logFile)
	}
	defer logFile.Close()

	// Receive events from stdin
	go func() {
		for {
			ev, err := RecvEvent(os.Stdin)
			if err == io.EOF {
				log.Println("stdin closed")
				return
			}
			if err != nil {
				log.Println("Error receiving message:", err)
				continue
			}
			evCh <- ev
		}
	}()

	// Set up a socket file
	socketPath := fmt.Sprintf("/tmp/native-app.%d.sock", pid)
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatal(err)
	}

	// Receive commands from an Unix domain socket
	go func() {
		lis, err := net.Listen("unix", socketPath)
		if err != nil {
			log.Fatal("listen error:", err)
		}
		defer lis.Close()

		for {
			conn, err := lis.Accept()
			if err != nil {
				log.Println("Accept error:", err)
				continue
			}

			go func(c net.Conn) {
				defer c.Close()

				scanner := bufio.NewScanner(c)

				scanner.Scan()
				if err := scanner.Err(); err != nil {
					log.Println("Read error:", err)
				}

				line := strings.TrimSpace(scanner.Text())

				cmd, err := ParseCommand(line)
				if err != nil {
					log.Println("Parse error:", err, "line:", line)
				}

				cmdCh <- cmd
			}(conn)
		}
	}()

	for {
		select {
		case ev := <-evCh:
			ev.Handle()
		case cmd := <-cmdCh:
			cmd.Execute()
		}
	}
}
