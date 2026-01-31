package main

import (
	"bufio"
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
			ev.Handle()
		case cw := <-cmdCh:
			executeCommand(cw.Cmd, cw.Conn)
			cw.Conn.Close()
		}
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

func listTabs(conn net.Conn) error {
	writer := bufio.NewWriter(conn)
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
