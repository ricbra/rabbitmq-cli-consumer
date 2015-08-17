package config

import (
	"os/user"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestFindsFullPathFiles(t *testing.T) {
	fs := &afero.MemMapFs{}
	fs.Create("/etc/rabbitmq-cli-consmer/rabbitmq-cli-consumer.conf")

	u := NewLocator([]string{
		"/etc/rabbitmq-cli-consmer/rabbitmq-cli-consumer.conf",
		"/home/test/my-config.conf",
	}, fs, nil)

	assert.Equal(
		t,
		[]string{"/etc/rabbitmq-cli-consmer/rabbitmq-cli-consumer.conf"},
		u.Locate(),
	)
}

func TestFindsConfigInHomedir(t *testing.T) {
	fs := &afero.MemMapFs{}
	fs.Create("/home/fakeuser/.rabbitmq-cli-consumer.conf")
	user := createUser()

	u := NewLocator([]string{
		"/etc/rabbitmq-cli-consmer/rabbitmq-cli-consumer.conf",
		"/home/test/my-config.conf",
	}, fs, user)

	assert.Equal(
		t,
		[]string{"/home/fakeuser/.rabbitmq-cli-consumer.conf"},
		u.Locate(),
	)
}

func createUser() *user.User {
	return &user.User{
		Uid:      "1",
		Gid:      "1",
		Username: "fakeuser",
		Name:     "Foo Bar",
		HomeDir:  "/home/fakeuser",
	}
}
