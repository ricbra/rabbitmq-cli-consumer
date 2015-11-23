package command

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestExecutesCommand(t *testing.T) {
	var b bytes.Buffer
	errLogger := log.New(&b, "", 0)
	infLogger := log.New(&b, "", 0)

	cmd := new(TestCommand)

	cmd.On("Output").Return(make([]byte, 0), nil).Once()

	executer := New(errLogger, infLogger)
	executer.Execute(cmd)

	cmd.AssertExpectations(t)

}

type TestCommand struct {
	mock.Mock
}

func (t *TestCommand) CombinedOutput() (out []byte, err error) {
	argsT := t.Called()

	return argsT.Get(0).([]byte), argsT.Error(1)
}

func (t *TestCommand) Output() (out []byte, err error) {
	argsT := t.Called()

	return argsT.Get(0).([]byte), argsT.Error(1)
}
