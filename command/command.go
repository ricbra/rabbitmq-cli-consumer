package command

import (
	"strings"
)

func Factory(baseCmd string) *CommandFactory {
	var pcs []string
	if split := strings.Split(baseCmd, " "); len(split) > 1 {
		baseCmd, pcs = split[0], split[1:]
	}
	return &CommandFactory{
		Cmd:  baseCmd,
		Args: pcs,
	}
}
