package command

import (
	"fmt"
	"os/exec"
	"strings"
)

func Execute(cmd *exec.Cmd) bool {
	out, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("Failed: %s\n", string(out[:]))
		//		log.Error())
		//		log.Error(err.Error())
		return false
	}
	fmt.Println(string(out[:]))
	//	log.Info(fmt.Sprintf("Processed message into command %s", args[0]))
	return true
}

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
