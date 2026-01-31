package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"

	"rofi-chrome-tab/config"
	"rofi-chrome-tab/protocol"
	"rofi-chrome-tab/receiver"
)

// Main

var (
	tabs  []protocol.Tab
	evCh  = make(chan protocol.Event, 1)
	cmdCh = make(chan receiver.CommandWithConn, 1)

	pid = os.Getpid()
)

func main() {
	// Set up log file
	logCloser, err := config.SetupLogging("/tmp/rofi-chrome-tab.log")
	if err != nil {
		os.Exit(1)
	}
	defer logCloser.Close()

	receiver.StartEventReceiver(os.Stdin, evCh)

	socketPath := receiver.GetSocketPath(pid, config.Debug)
	receiver.StartCommandReceiver(socketPath, cmdCh)

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

func handleEvent(ev protocol.Event) error {
	switch e := ev.(type) {
	case protocol.UpdatedEvent:
		tabs = e.Tabs
		return nil

	default:
		return fmt.Errorf("unknown event type: %T", ev)
	}
}

func executeCommand(cmd protocol.Command, conn net.Conn) error {
	switch c := cmd.(type) {
	case protocol.ListCommand:
		return listTabs(conn)

	case protocol.SelectCommand:
		protocol.SendAction(os.Stdout, protocol.SelectAction(c))
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
			return fmt.Errorf("write error: %v", err)
		}
	}
	return nil
}
