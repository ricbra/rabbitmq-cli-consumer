package command

import "log"

type CommandExecuter struct {
	errLogger *log.Logger
	infLogger *log.Logger
}

type Command interface {
	CombinedOutput() (out []byte, err error)
}

func New(errLogger, infLogger *log.Logger) *CommandExecuter {
	return &CommandExecuter{
		errLogger: errLogger,
		infLogger: infLogger,
	}
}

func (me CommandExecuter) Execute(cmd Command) bool {
	me.infLogger.Println("Processing message...")

	out, err := cmd.CombinedOutput()

	if err != nil {
		me.infLogger.Println("Failed. Check error log for details.")
		me.errLogger.Printf("Failed: %s\n", string(out[:]))
		me.errLogger.Printf("Error: %s\n", err)
		return false
	}

	me.infLogger.Println("Processed!")

	return true
}
