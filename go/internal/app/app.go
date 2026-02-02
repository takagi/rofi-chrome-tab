package app

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"rofi-chrome-tab/internal/action"
	"rofi-chrome-tab/internal/command"
	"rofi-chrome-tab/internal/commandreceiver"
	"rofi-chrome-tab/internal/debug"
	"rofi-chrome-tab/internal/event"
	"rofi-chrome-tab/internal/eventreceiver"
	"rofi-chrome-tab/internal/logging"
	"rofi-chrome-tab/internal/types"
)

func Run() error {
	logCloser, err := logging.SetupLogging("/tmp/rofi-chrome-tab.log")
	if err != nil {
		return err
	}
	defer logCloser.Close()

	pid := os.Getpid()
	evCh := make(chan event.Event, 1)
	cmdCh := make(chan commandreceiver.CommandWithConn, 1)
	var tabs []types.Tab

	eventreceiver.Start(os.Stdin, evCh)
	_ = commandreceiver.Start(pid, debug.IsDebugMode(), cmdCh)

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

func handleEvent(tabs *[]types.Tab, ev event.Event) error {
	switch e := ev.(type) {
	case event.UpdatedEvent:
		*tabs = e.Tabs
		return nil
	default:
		return fmt.Errorf("unknown event type: %T", ev)
	}
}

func executeCommand(tabs []types.Tab, cmd command.Command, conn net.Conn, pid int) error {
	switch c := cmd.(type) {
	case command.ListCommand:
		return listTabs(conn, tabs, pid)
	case command.SelectCommand:
		return action.SendAction(os.Stdout, action.SelectAction{TabID: c.TabID})
	default:
		return fmt.Errorf("unknown command type: %T", cmd)
	}
}

func listTabs(w io.Writer, tabs []types.Tab, pid int) error {
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
