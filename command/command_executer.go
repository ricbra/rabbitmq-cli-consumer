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
	me.infLogger.Println("Processing message...")
	out, err := cmd.CombinedOutput()
	me.infLogger.Println("Processed!")

	if err != nil {
		me.infLogger.Println("Failed. Check error log for details.")
		me.errLogger.Printf("Failed: %s\n", string(out[:]))
		return false
	}

	return true
}
