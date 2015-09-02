package main

import (
	"io"
	"log"
	"os"
	"os/user"

	"code.google.com/p/gcfg"

	"github.com/codegangsta/cli"
	"github.com/ricbra/rabbitmq-cli-consumer/command"
	"github.com/ricbra/rabbitmq-cli-consumer/config"
	"github.com/ricbra/rabbitmq-cli-consumer/consumer"
	"github.com/spf13/afero"
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
			Name:  "quiet, Q",
			Usage: "Enable quite mode (disables loggging to stdout and stderr)",
		},
	}
	app.Action = func(c *cli.Context) {
		if c.String("executable") == "" {
			cli.ShowAppHelp(c)
			os.Exit(1)
		}

		verbose := !c.Bool("quiet")
		logger := log.New(os.Stderr, "", 0)

		// Config finding and parsing
		// Perhaps refactor into something more elegant
		user, _ := user.Current()
		locator := config.NewLocator([]string{c.String("configuration")}, &afero.OsFs{}, user)
		configs := []config.Config{}
		err, locations := locator.Locate()
		if err != nil {
			logger.Fatalf("Failed locating configuration: %s\n", err)
		}

		for _, path := range locations {
			logger.Printf("Found config: %s", path)
			cfg := config.Config{}
			if err := gcfg.ReadFileInto(&cfg, path); err == nil {
				configs = append(configs, cfg)
			} else {
				logger.Printf("Could not parse config: %s", err)
			}
		}
		merger := config.ConfigMerger{}
		cfg, _ := merger.Merge(configs)
		if !config.Validate(cfg, logger) {
			logger.Fatalf("Please fix configuration issues.")
		}

		// fmt.Println(config)
		// os.Exit(0)

		errLogger, err := createLogger(cfg.Logs.Error, verbose, os.Stderr)
		if err != nil {
			logger.Fatalf("Failed creating error log: %s", err)
		}

		infLogger, err := createLogger(cfg.Logs.Info, verbose, os.Stdout)
		if err != nil {
			logger.Fatalf("Failed creating info log: %s", err)
		}

		factory := command.Factory(c.String("executable"))

		client, err := consumer.New(&cfg, factory, errLogger, infLogger)
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
