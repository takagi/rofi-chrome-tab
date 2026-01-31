package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
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

	startEventReceiver(os.Stdin, evCh)

	socketPath := getSocketPath(pid, debug)
	startCommandReceiver(socketPath, cmdCh)

	for {
		select {
		case ev := <-evCh:
			handleEvent(ev)
		case cw := <-cmdCh:
			executeCommand(cw.Cmd, cw.Conn, pid)
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

func executeCommand(cmd Command, conn net.Conn, pid int) error {
	switch c := cmd.(type) {
	case ListCommand:
		return listTabs(conn, pid)

	case SelectCommand:
		SendAction(os.Stdout, SelectAction(c))
		return nil

	default:
		return fmt.Errorf("unknown command type: %T", cmd)
	}
}

func listTabs(w io.Writer, pid int) error {
	writer := bufio.NewWriter(w)
	defer writer.Flush()

	for _, tab := range tabs {
		line := fmt.Sprintf("%d,%d,%s,%s", pid, tab.ID, tab.Host, tab.Title)
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("write error: %v", err)
		}
	}
	return nil
}
