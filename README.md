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

You have the choice to either compile yourself or by installing via package.

## Compiling

This section assumes you're familiar with the Go language.

Use <code>go get</code> to get the source local:

    $ go get github.com/ricbra/rabbitmq-cli-consumer

Change to the directory, e.g.:

    $ cd $GOPATH/src/github.com/ricbra/rabbitmq-cli-consumer

Get the dependencies:

    $ go get ./...

Then build and/or install:

    $ go build
    $ go install

# Usage

Soon to follow.

# Developing

Same here.


