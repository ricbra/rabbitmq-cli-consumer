package command

import (
	"os/exec"
	"log"
)

type CommandExecuter struct {
	errLogger *log.Logger
	infLogger *log.Logger
}

func New(errLogger, infLogger *log.Logger) *CommandExecuter {
	return &CommandExecuter{
		errLogger: errLogger,
		infLogger: infLogger,
	}
}

func (me CommandExecuter) Execute(cmd *exec.Cmd) bool {
	out, err := cmd.CombinedOutput()

	me.infLogger.Println("Processing message...")

	if err != nil {
		me.errLogger.Printf("Failed: %s\n", string(out[:]))
		return false
	}

	me.infLogger.Println(string(out[:]))

	return false
}
