package config

import (
	"errors"
	"fmt"
	"os/user"

	"github.com/spf13/afero"
)

type FileLocator struct {
	Paths      []string
	Filesystem afero.Fs
}

func (r FileLocator) Locate() (error, []string) {
	exists := []string{}
	for _, path := range r.Paths {
		if _, err := r.Filesystem.Stat(path); err == nil {
			exists = append(exists, path)
		}
	}
	if len(exists) == 0 {
		return errors.New("No configuration files found, exiting"), exists
	}

	return nil, exists
}

type Locator interface {
	Locate() (error, []string)
}

func NewLocator(paths []string, filesystem afero.Fs, user *user.User) Locator {
	if user != nil {
		paths = append([]string{fmt.Sprintf("%s/.rabbitmq-cli-consumer.conf", user.HomeDir)}, paths...)
	}
	paths = append([]string{"/etc/rabbitmq-cli-consumer/rabbitmq-cli-consumer.conf"}, paths...)

	return FileLocator{
		Paths:      paths,
		Filesystem: filesystem,
	}
}
