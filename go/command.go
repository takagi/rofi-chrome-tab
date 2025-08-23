package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Command interface {
	Execute(net.Conn) error
}

type ListCommand struct{}

func (c ListCommand) Execute(conn net.Conn) error {
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

type SelectCommand struct {
	tabID int
}

func (c SelectCommand) Execute(net.Conn) error {
	SendAction(os.Stdout, SelectAction{TabID: c.tabID})
	return nil
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
