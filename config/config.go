package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/codegangsta/cli"

	"code.google.com/p/gcfg"

	"gopkg.in/validator.v2"
)

// Config contains all config values
type Config struct {
	RabbitMq struct {
		Host        string `validate:"nonzero"`
		Username    string `validate:"nonzero"`
		Password    string `validate:"nonzero"`
		Port        string `validate:"nonzero"`
		Vhost       string `validate:"nonzero"`
		Compression bool
	}
	Prefetch struct {
		Count  int `validate:"nonzero"`
		Global bool
	}
	Queue struct {
		Name       string `validate:"nonzero"`
		Durable    bool
		Autodelete bool
		Exclusive  bool
		Nowait     bool
		Key        string
		TTL        int
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

// Validate validtes Config and prints errors.
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

// Default returns Config with default defined
func Default() Config {
	return CreateFromString(
		`[prefetch]
    count=3
    global=Off

    [exchange]
    autodelete=Off
    type=direct
    durable=On
    `)
}

// CreateFromString creates Config from string
func CreateFromString(config string) Config {
	cfg := Config{}
	gcfg.ReadStringInto(&cfg, config)

	return cfg
}

// CreateFromCliContext creates config from options passed to cli
func CreateFromCliContext(c *cli.Context) Config {
	str := fmt.Sprintf(
		`[rabbitmq]
		host=%s
		username=%s
		password=%s
		vhost=%s
		port=%s
		compression=%s

		[prefetch]
		count=%s
		global=%s

		[queue]
		name=%s
		durable=%s
		autodelete=%s
		exclusive=%s
		nowait=%s
		key=%s

		[exchange]
		name=%s
		autodelete=%s
		type=%s
		durable=%s`,
		c.String("host"),
		c.String("username"),
		c.String("password"),
		c.String("vhost"),
		c.String("port"),
		c.String("compression"),
		c.String("prefetch-count"),
		c.String("prefetch-global"),
		c.String("queue-name"),
		c.String("queue-durable"),
		c.String("queue-autodelete"),
		c.String("queue-exclusive"),
		c.String("queue-nowait"),
		c.String("queue-key"),
		c.String("exchange-name"),
		c.String("exchange-autodelete"),
		c.String("exchange-type"),
		c.String("exchange-durable"),
	)

	return CreateFromString(str)
}
