package config

import (
	"log"
	"strings"

	"gopkg.in/validator.v2"
)

type Config struct {
	RabbitMq struct {
		Host        string `validate:"nonzero"`
		Username    string `validate:"nonzero"`
		Password    string `validate:"nonzero"`
		Port        string `validate:"nonzero"`
		Vhost       string `validate:"nonzero"`
		Queue       string `validate:"nonzero"`
		Compression bool
	}
	Prefetch struct {
		Count  int `validate:"nonzero"`
		Global bool
	}
	Exchange struct {
		Name       string `validate:"nonzero"`
		Autodelete bool
		Type       string `validate:"nonzero"`
		Durable    bool
	}
	Logs struct {
		Error string `validate:"nonzero"`
		Info  string `validate:"nonzero"`
	}
}

func Validate(config Config, logger *log.Logger) bool {
	if err := validator.Validate(config); err != nil {
		for f, e := range err.(validator.ErrorMap) {
			split := strings.Split(strings.ToLower(f), ".")
			msg := e.Error()
			switch msg {
			case "zero value":
				msg = "This option is required"
			}

			logger.Printf("The option \"%s\" under section \"%s\" is invalid: %s\n", split[1], split[0], msg)
		}
		return false
	}

	return true
}
