package main

import (
	"github.com/codegangsta/cli"
	"github.com/ricbra/rabbitmq-cli-consumer/command"
	"github.com/ricbra/rabbitmq-cli-consumer/config"
	"github.com/ricbra/rabbitmq-cli-consumer/consumer"
	"os"
	"log"
	"io"
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

		errLogger := log.New(os.Stderr, "", log.Ldate | log.Ltime)
		cfg, err := config.LoadAndParse(c.String("configuration"))

		if err != nil {
			errLogger.Fatalf("Failed parsing configuration: %s\n", err)
		}

		file, err := os.Create(cfg.Logs.Error)

		if err != nil {
			errLogger.Fatalf("Failed creating error log: %s\n", err)
		}

		writer := io.MultiWriter(os.Stderr, file)
		errLogger = log.New(writer, "", log.Ldate | log.Ltime)
		factory := command.Factory(c.String("executable"))

		client, err := consumer.New(cfg, factory, errLogger)
		if err != nil {
			errLogger.Fatalf( "Failed creating consumer: %s", err)
		}

		client.Consume()
	}

	app.Run(os.Args)
}
