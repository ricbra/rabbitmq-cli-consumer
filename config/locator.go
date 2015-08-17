package config

import (
	"fmt"
	"os/user"

	"github.com/spf13/afero"
)

type FileLocator struct {
	Paths      []string
	Filesystem afero.Fs
}

func (r FileLocator) Locate() []string {
	exists := []string{}
	for _, path := range r.Paths {
		if _, err := r.Filesystem.Stat(path); err == nil {
			exists = append(exists, path)
		}
	}

	return exists
}

type Locator interface {
	Locate() []string
}

func NewLocator(paths []string, filesystem afero.Fs, user *user.User) Locator {
	if user != nil {
		paths = append(paths, fmt.Sprintf("%s/.rabbitmq-cli-consumer.conf", user.HomeDir))
	}

	return FileLocator{
		Paths:      paths,
		Filesystem: filesystem,
	}
}
