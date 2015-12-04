package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"os/user"
	"syscall"

	"code.google.com/p/gcfg"

	"github.com/codegangsta/cli"
	"github.com/ricbra/rabbitmq-cli-consumer/command"
	"github.com/ricbra/rabbitmq-cli-consumer/config"
	"github.com/ricbra/rabbitmq-cli-consumer/consumer"
	"github.com/spf13/afero"
)

var files []*os.File

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
		cli.StringFlag{
			Name:  "quiet, Q",
			Usage: "Enable quite mode (disables loggging to stdout and stderr)",
		},
		cli.StringFlag{
			Name:  "host, H",
			Usage: "IP or hostname of RabbitMQ server",
		},
		cli.StringFlag{
			Name:  "username, u",
			Usage: "RabbitMQ username",
		},
		cli.StringFlag{
			Name:  "password, p",
			Usage: "RabbitMQ password",
		},
		cli.StringFlag{
			Name:  "vhost, V",
			Usage: "RabbitMQ vhost",
		},
		cli.StringFlag{
			Name:  "port, P",
			Usage: "RabbitMQ port",
		},
		cli.StringFlag{
			Name:  "compression, o",
			Usage: "Enable compression of messages",
		},
		cli.StringFlag{
			Name:  "prefetch-count, C",
			Usage: "Prefetch count",
		},
		cli.StringFlag{
			Name:  "prefetch-global, G",
			Usage: "Set prefetch count as global",
		},
		cli.StringFlag{
			Name:  "queue-name,q",
			Usage: "Queue name",
		},
		cli.StringFlag{
			Name:  "queue-durable,D",
			Usage: "Mark queue as durable",
		},
		cli.StringFlag{
			Name:  "queue-autodelete,a",
			Usage: "Autodelete queue",
		},
		cli.StringFlag{
			Name:  "queue-exlusive,E",
			Usage: "Mark queue as exclusive",
		},
		cli.StringFlag{
			Name:  "queue-nowait, T",
			Usage: "Do not wait for the server to confirm the binding",
		},
		cli.StringFlag{
			Name:  "queue-key, k",
			Usage: "Routing key to bind the queue on",
		},
		cli.StringFlag{
			Name:  "exchange-name, X",
			Usage: "Exchange name",
		},
		cli.StringFlag{
			Name:  "exchange-autodelete, t",
			Usage: "Autodelete exchange",
		},
		cli.StringFlag{
			Name:  "exchange-type, y",
			Usage: "Exchange type (direct, fanout, topic or headers)",
		},
		cli.StringFlag{
			Name:  "exchange-durable, j",
			Usage: "Mark exchange as durable",
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

		// Read config settings passed in as option to the command
		configs = append(configs, config.CreateFromCliContext(c))
		merger := config.ConfigMerger{}
		cfg, _ := merger.Merge(configs)
		if !config.Validate(cfg, logger) {
			logger.Fatalf("Please fix configuration issues.")
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

		client, err := consumer.New(&cfg, factory, errLogger, infLogger)
		if err != nil {
			errLogger.Fatalf("Failed creating consumer: %s", err)
		}

		// Reopen logs on USR1
		sigs := make(chan os.Signal)
		signal.Notify(sigs, syscall.SIGUSR1)

		go func() {
			for _ = range sigs {
				for _, file := range files {
					filename := file.Name()
					file.Close()
					new, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)

					if err != nil {
						panic(fmt.Sprintf("Failed reopeing log file: %s", err))
					}
					file = new
				}
			}
		}()

		client.Consume()
	}

	app.Run(os.Args)
}

func createLogger(filename string, verbose bool, out io.Writer) (*log.Logger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)

	if err != nil {
		return nil, err
	}

	files = append(files, file)

	var writers = []io.Writer{
		file,
	}

	if verbose {
		writers = append(writers, out)
	}

	return log.New(io.MultiWriter(writers...), "", log.Ldate|log.Ltime), nil
}
