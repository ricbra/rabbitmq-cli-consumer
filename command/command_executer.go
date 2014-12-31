package command

import (
	"os/exec"
	"log"
)

type CommandExecuter struct {
	errLogger *log.Logger
}

func New(errLogger *log.Logger) *CommandExecuter {
	return &CommandExecuter{
		errLogger: errLogger,
	}
}

func (me CommandExecuter) Execute(cmd *exec.Cmd) bool {
	out, err := cmd.CombinedOutput()

	if err != nil {
		me.errLogger.Printf("Failed: %s\n", string(out[:]))
		return false
	}

	return false
}
