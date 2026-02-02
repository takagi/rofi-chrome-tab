package app

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"rofi-chrome-tab/internal/command_receiver"
	"rofi-chrome-tab/internal/debug"
	"rofi-chrome-tab/internal/event_receiver"
	"rofi-chrome-tab/internal/logging"
	"rofi-chrome-tab/internal/protocol"
)

// Main
func Run() error {
	// Set up log file
	logCloser, err := logging.SetupLogging("/tmp/rofi-chrome-tab.log")
	if err != nil {
		return err
	}
	defer logCloser.Close()

	pid := os.Getpid()
	evCh := make(chan protocol.Event, 1)
	cmdCh := make(chan command_receiver.CommandWithConn, 1)
	var tabs []protocol.Tab

	event_receiver.Start(os.Stdin, evCh)
	_ = command_receiver.Start(pid, debug.IsDebugMode(), cmdCh)

	for {
		select {
		case ev := <-evCh:
			if err := handleEvent(&tabs, ev); err != nil {
				log.Println("Error handling event:", err)
			}
		case cw := <-cmdCh:
			if err := executeCommand(tabs, cw.Cmd, cw.Conn, pid); err != nil {
				log.Println("Command error:", err)
			}
			cw.Conn.Close()
		}
	}
}

func handleEvent(tabs *[]protocol.Tab, ev protocol.Event) error {
	switch e := ev.(type) {
	case protocol.UpdatedEvent:
		*tabs = e.Tabs
		return nil
	default:
		return fmt.Errorf("unknown event type: %T", ev)
	}
}

func executeCommand(tabs []protocol.Tab, cmd protocol.Command, conn net.Conn, pid int) error {
	switch c := cmd.(type) {
	case protocol.ListCommand:
		return listTabs(conn, tabs, pid)
	case protocol.SelectCommand:
		return protocol.SendAction(os.Stdout, protocol.SelectAction(c))
	default:
		return fmt.Errorf("unknown command type: %T", cmd)
	}
}

func listTabs(w io.Writer, tabs []protocol.Tab, pid int) error {
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
