package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/ricbra/rabbitmq-cli-consumer/command"
	"github.com/ricbra/rabbitmq-cli-consumer/config"
	"github.com/ricbra/rabbitmq-cli-consumer/consumer"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "rabbitmq-cli-consumer"
	app.Usage = "Consume RabbitMQ easily to any cli program"
	app.Author = "Richard van den Brand"
	app.Email = "richard@vandenbrand.org"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "executable, e",
			Usage: "Location of executable",
		},
		cli.StringFlag{
			Name:  "configuration, c",
			Usage: "Location of configuration file",
		},
	}
	app.Action = func(c *cli.Context) {
		if c.String("configuration") == "" && c.String("executable") == "" {
			cli.ShowAppHelp(c)
			os.Exit(1)
		}

		cfg, err := config.LoadAndParse(c.String("configuration"))

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed parsing configuration: %s\n", err)
			os.Exit(1)
		}

		factory := command.Factory(c.String("executable"))

		client, err := consumer.New(cfg, factory)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed creating consumer: %s", err)
			os.Exit(1)
		}

		client.Consume()
	}

	app.Run(os.Args)
}
