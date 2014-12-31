package command

import (
	"os/exec"
)

type CommandFactory struct {
	Cmd  string
	Args []string
}

func (me CommandFactory) Create(body string) *exec.Cmd {
	return exec.Command(me.Cmd, append(me.Args, body)...)
}
