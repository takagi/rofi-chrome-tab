package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"rofi-chrome-tab/internal/action"
	"rofi-chrome-tab/internal/command"
	"rofi-chrome-tab/internal/event"
	"rofi-chrome-tab/internal/logger"
	"rofi-chrome-tab/internal/receiver"
	"rofi-chrome-tab/internal/types"
)

var (
	tabs  []types.Tab
	evCh  = make(chan event.Event, 1)
	cmdCh = make(chan receiver.CommandWithConn, 1)
)

func main() {
	// Set up log file
	logCloser, err := logger.SetupLogging("/tmp/rofi-chrome-tab.log")
	if err != nil {
		os.Exit(1)
	}
	defer logCloser.Close()

	pid := os.Getpid()

	receiver.StartEventReceiver(os.Stdin, evCh)

	_ = receiver.StartCommandReceiver(pid, logger.Debug, cmdCh)

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

func handleEvent(ev event.Event) error {
	switch e := ev.(type) {
	case event.UpdatedEvent:
		tabs = e.Tabs
		return nil

	default:
		return fmt.Errorf("unknown event type: %T", ev)
	}
}

func executeCommand(cmd command.Command, conn io.Writer, pid int) error {
	switch c := cmd.(type) {
	case command.ListCommand:
		return listTabs(conn, pid)

	case command.SelectCommand:
		action.SendAction(os.Stdout, action.SelectAction{TabID: c.TabID})
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
