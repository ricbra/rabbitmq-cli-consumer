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

		logger := log.New(os.Stderr, "", log.Ldate | log.Ltime)
		cfg, err := config.LoadAndParse(c.String("configuration"))

		if err != nil {
			logger.Fatalf("Failed parsing configuration: %s\n", err)
		}

		errLogger, err := createMultiLogger(cfg.Logs.Error, os.Stderr)
		infLogger, err := createMultiLogger(cfg.Logs.Info, os.Stdout)

		if err != nil {
			logger.Fatalf("Failed creating error log: %s\n", err)
		}

		factory := command.Factory(c.String("executable"))

		client, err := consumer.New(cfg, factory, errLogger, infLogger)
		if err != nil {
			errLogger.Fatalf( "Failed creating consumer: %s", err)
		}

		client.Consume()
	}

	app.Run(os.Args)
}

func createMultiLogger(filename string, out io.Writer) (*log.Logger, error) {
	file, err := os.Create(filename)

	if err != nil {
		return nil, err
	}

	return log.New(io.MultiWriter(out, file), "", log.Ldate | log.Ltime), nil
}
