package command

import (
	"fmt"
	"log"
)

type Executer interface {
	Execute(cmd Command) (result []byte, err error)
}

type CommandExecuter struct {
	errLogger *log.Logger
	infLogger *log.Logger
}

type Command interface {
	CombinedOutput() (out []byte, err error)
	Output() (out []byte, err error)
}

func New(errLogger, infLogger *log.Logger) *CommandExecuter {
	return &CommandExecuter{
		errLogger: errLogger,
		infLogger: infLogger,
	}
}

func (me CommandExecuter) Execute(cmd Command) (result []byte, err error) {
	me.infLogger.Println("Processing message...")

	out, err := cmd.Output()

	if err != nil {
		me.infLogger.Println("Failed. Check error log for details.")
		me.errLogger.Printf("Failed: %s\n", string(out[:]))
		me.errLogger.Printf("Error: %s\n", err)

		return out, fmt.Errorf("Error occured during execution of command: %s", err)
	}

	me.infLogger.Println("Processed!")

	return out, nil
}
