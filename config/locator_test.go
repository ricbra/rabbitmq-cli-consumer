package config

import (
	"os/user"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestFindsDefauktFullpathConfig(t *testing.T) {
	fs := &afero.MemMapFs{}
	fs.Create("/etc/rabbitmq-cli-consumer/rabbitmq-cli-consumer.conf")

	u := NewLocator([]string{
		"/home/test/my-config.conf",
	}, fs, nil)

	_, paths := u.Locate()
	assert.Equal(
		t,
		[]string{"/etc/rabbitmq-cli-consumer/rabbitmq-cli-consumer.conf"},
		paths,
	)
}

func TestFindsConfigInHomedir(t *testing.T) {
	fs := &afero.MemMapFs{}
	fs.Create("/home/fakeuser/.rabbitmq-cli-consumer.conf")
	user := createUser()

	u := NewLocator([]string{
		"/home/test/my-config.conf",
	}, fs, user)

	_, paths := u.Locate()
	assert.Equal(
		t,
		[]string{"/home/fakeuser/.rabbitmq-cli-consumer.conf"},
		paths,
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
