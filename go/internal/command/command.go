package command

import (
	"fmt"
	"strconv"
	"strings"
)

type Command interface {
	isCommand()
}

type ListCommand struct{}

func (ListCommand) isCommand() {}

type SelectCommand struct {
	TabID int
}

func (SelectCommand) isCommand() {}

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

		return SelectCommand{TabID: tabID}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", fields[0])
	}
}
