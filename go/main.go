package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

// Main

type CommandWithConn struct {
	Cmd  Command
	Conn net.Conn
}

var (
	tabs  []Tab
	evCh  = make(chan Event, 1)
	cmdCh = make(chan CommandWithConn, 1)

	pid = os.Getpid()
)

func main() {
	// Set up log file
	logCloser, err := SetupLogging("/tmp/rofi-chrome-tab.log")
	if err != nil {
		os.Exit(1)
	}
	defer logCloser.Close()

	startEventReceiver()

	startCommandReceiver()

	for {
		select {
		case ev := <-evCh:
			handleEvent(ev)
		case cw := <-cmdCh:
			executeCommand(cw.Cmd, cw.Conn)
			cw.Conn.Close()
		}
	}
}

func handleEvent(ev Event) error {
	switch e := ev.(type) {
	case UpdatedEvent:
		tabs = e.Tabs
		return nil

	default:
		return fmt.Errorf("unknown event type: %T", ev)
	}
}

func executeCommand(cmd Command, conn net.Conn) error {
	switch c := cmd.(type) {
	case ListCommand:
		return listTabs(conn)

	case SelectCommand:
		SendAction(os.Stdout, SelectAction{TabID: c.TabID})
		return nil

	default:
		return fmt.Errorf("unknown command type: %T", cmd)
	}
}

func listTabs(w io.Writer) error {
	writer := bufio.NewWriter(w)
	defer writer.Flush()

	for _, tab := range tabs {
		line := fmt.Sprintf("%d,%d,%s,%s", pid, tab.ID, tab.Host, tab.Title)
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("write error: %v\n", err)
		}
	}
	return nil
}

func startEventReceiver() {
	// Receive events from stdin
	go func() {
		const maxMessageSize = 10 * 1024 * 1024 // 10MB limit
		for {
			// Read 4-byte length header
			lenBuf := make([]byte, 4)
			if _, err := io.ReadFull(os.Stdin, lenBuf); err != nil {
				if err == io.EOF {
					log.Println("stdin closed")
					return
				}
				log.Println("Error reading length header:", err)
				continue
			}
			length := binary.LittleEndian.Uint32(lenBuf)

			// Validate message length to prevent excessive memory allocation
			if length > maxMessageSize {
				log.Printf("Message too large: %d bytes (max %d bytes), closing stdin receiver", length, maxMessageSize)
				return
			}

			// Read message body
			buf := make([]byte, length)
			if _, err := io.ReadFull(os.Stdin, buf); err != nil {
				if err == io.EOF {
					log.Println("stdin closed")
					return
				}
				log.Println("Error reading message body:", err)
				continue
			}

			// Parse event from bytes
			ev, err := RecvEvent(buf)
			if err != nil {
				log.Println("Error parsing event:", err)
				continue
			}
			log.Printf("Received event: %T", ev)
			evCh <- ev
		}
	}()
}

func startCommandReceiver() {
	// Set up a socket file
	var socketPath string
	if !debug {
		socketPath = fmt.Sprintf("/tmp/native-app.%d.sock", pid)
	} else {
		socketPath = "/tmp/native-app.sock"
	}
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

		log.Printf("Listening on socket: %s", socketPath)

		for {
			conn, err := lis.Accept()
			if err != nil {
				log.Println("Accept error:", err)
				continue
			}

			go func(c net.Conn) {
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
				log.Printf("Received command: %T", cmd)

				cmdCh <- CommandWithConn{Cmd: cmd, Conn: c}
			}(conn)
		}
	}()
}
