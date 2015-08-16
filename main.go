package main

import (
	"io"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/ricbra/rabbitmq-cli-consumer/command"
	"github.com/ricbra/rabbitmq-cli-consumer/config"
	"github.com/ricbra/rabbitmq-cli-consumer/consumer"
)

func main() {
	app := cli.NewApp()
	app.Name = "rabbitmq-cli-consumer"
	app.Usage = "Consume RabbitMQ easily to any cli program"
	app.Author = "Richard van den Brand"
	app.Email = "richard@vandenbrand.org"
	app.Version = "2.0.0-dev"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "executable, e",
			Usage: "Location of executable",
		},
		cli.StringFlag{
			Name:  "configuration, c",
			Usage: "Location of configuration file",
		},
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "Enable verbose mode (logs to stdout and stderr)",
		},
	}
	app.Action = func(c *cli.Context) {
		if c.String("configuration") == "" && c.String("executable") == "" {
			cli.ShowAppHelp(c)
			os.Exit(1)
		}

		verbose := c.Bool("verbose")

		logger := log.New(os.Stderr, "", log.Ldate|log.Ltime)
		cfg, err := config.LoadAndParse(c.String("configuration"))

		if err != nil {
			logger.Fatalf("Failed parsing configuration: %s\n", err)
		}

		errLogger, err := createLogger(cfg.Logs.Error, verbose, os.Stderr)
		if err != nil {
			logger.Fatalf("Failed creating error log: %s", err)
		}

		infLogger, err := createLogger(cfg.Logs.Info, verbose, os.Stdout)
		if err != nil {
			logger.Fatalf("Failed creating info log: %s", err)
		}

		factory := command.Factory(c.String("executable"))

		client, err := consumer.New(cfg, factory, errLogger, infLogger)
		if err != nil {
			errLogger.Fatalf("Failed creating consumer: %s", err)
		}

		client.Consume()
	}

	app.Run(os.Args)
}

func createLogger(filename string, verbose bool, out io.Writer) (*log.Logger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)

	if err != nil {
		return nil, err
	}

	var writers = []io.Writer{
		file,
	}

	if verbose {
		writers = append(writers, out)
	}

	return log.New(io.MultiWriter(writers...), "", log.Ldate|log.Ltime), nil
}
