RabbitMQ cli consumer
---------------------

If you are a fellow PHP developer just like me you're probably aware of the following fact:
PHP really SUCKS in long running tasks.

When using RabbitMQ with pure PHP consumers you have to deal with stability issues. Probably you are killing your
consumers regularly just like me. And try to solve the problem with supervisord. Which also means on every deploy you
have to restart your consumers. A little bit dramatic if you ask me.

This library aims at PHP developers solving the above described problem with RabbitMQ. Why don't let the polling over to
a language as Go which is much better suited to run long running tasks.

# Installation

You have the choice to either compile yourself or by installing via package or binary.

## Binary

Binaries can be found at: https://github.com/ricbra/rabbitmq-cli-consumer/releases

## Compiling

This section assumes you're familiar with the Go language.

Use <code>go get</code> to get the source local:

```bash
$ go get github.com/ricbra/rabbitmq-cli-consumer
```

Change to the directory, e.g.:

```bash
$ cd $GOPATH/src/github.com/ricbra/rabbitmq-cli-consumer
```

Get the dependencies:

```bash
$ go get ./...
```

Then build and/or install:

```bash
$ go build
$ go install
```

# Usage

Run without arguments or with <code>--help</code> switch to show the helptext:

    $ rabbitmq-cli-consumer
    NAME:
       rabbitmq-cli-consumer - Consume RabbitMQ easily to any cli program

    USAGE:
       rabbitmq-cli-consumer [global options] command [command options] [arguments...]

    VERSION:
       0.0.1

    AUTHOR:
      Richard van den Brand - <richard@vandenbrand.org>

    COMMANDS:
       help, h	Shows a list of commands or help for one command

    GLOBAL OPTIONS:
       --executable, -e 	Location of executable
       --configuration, -c 	Location of configuration file
       --verbose, -V	Enable verbose mode (logs to stdout and stderr)
       --help, -h		show help
       --version, -v	print the version

A configuration file is required. Example:

```ini
[rabbitmq]
host = localhost
username = username-of-rabbitmq-user
password = secret
vhost=/your-vhost
port=5672
queue=name-of-queue

[logs]
error = /location/to/error.log
info = /location/to/info.log
```

When you've created the configuration you can start the consumer like this:

    $ rabbitmq-cli-consumer -e "/path/to/your/app argument --flag" -c /path/to/your/configuration.conf -V

Run without <code>-V</code> to get rid of the output:

    $ rabbitmq-cli-consumer -e "/path/to/your/app argument --flag" -c /path/to/your/configuration.conf

# Developing

Missing anything? Found a bug? I love to see your PR.


