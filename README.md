RabbitMQ cli consumer
---------------------

[![Build Status](https://travis-ci.org/ricbra/rabbitmq-cli-consumer.svg)](https://travis-ci.org/ricbra/rabbitmq-cli-consumer)

If you are a fellow PHP developer just like me you're probably aware of the following fact:
PHP really SUCKS in long running tasks.

When using RabbitMQ with pure PHP consumers you have to deal with stability issues. Probably you are killing your
consumers regularly just like me. And try to solve the problem with supervisord. Which also means on every deploy you
have to restart your consumers. A little bit dramatic if you ask me.

This library aims at PHP developers solving the above described problem with RabbitMQ. Why don't let the polling over to
a language as Go which is much better suited to run long running tasks.

# Installation

You have the choice to either compile yourself or by installing via package or binary.

## APT Package

As I'm a Debian user myself Debian-based peeps are lucky and can use my APT repository.

Add this line to your <code>/etc/apt/sources.list</code> file:

    deb http://apt.vandenbrand.org/debian testing main

Fetch and install GPG key:

    $ wget http://apt.vandenbrand.org/apt.vandenbrand.org.gpg.key
    $ sudo apt-key add apt.vandenbrand.org.gpg.key

Update and install:

    $ sudo apt-get update
    $ sudo apt-get install rabbitmq-cli-consumer

## Create .deb package for service install

    sudo apt-get install golang gccgo-go ruby -y
    # Ubuntu
    sudo apt-get install gccgo-go -y
    # Debian
    sudo apt-get install gccgo -y
    sudo gem install fpm
    ./build_service_deb.sh

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

## Fanout

Todo.

## Remote Procedure Call

No special configuration is required for enabling RPC mode. You should be aware
that any output on STDOUT will be returned as reply to the requesting client. To
demonstrate how RPC works, we'll implement [the example](https://www.rabbitmq.com/tutorials/tutorial-six-php.html) on the RabbitMQ site
using rabbitmq-cli-consumer.

We don't change the <code>rpc_client.php</code>, only the <code>rpc_server.php</code>:

```php
<?php

function fib($n) {
    if ($n == 0)
        return 0;
    if ($n == 1)
        return 1;
    return fib($n-1) + fib($n-2);
}

echo fib(base64_decode($argv[1]));
// Return exit(1); if something goes wrong, msg will be requeued
exit(0);
```

Configuration for the consumer:

```ini
[rabbitmq]
...

[prefetch]
count=1
global=Off

[queue]
name=rpc_queue
durable=On
autodelete=Off
exclusive=Off
nowait=Off

[exchange]
name=rpc_queue
autodelete=Off
type=direct
durable=On
```

# Configuration

A configuration file is required. Example:

```ini
[rabbitmq]
host = localhost
username = username-of-rabbitmq-user
password = secret
vhost=/your-vhost
port=5672
compression=Off

[queue]
nane=your-queue-name

[logs]
error = /location/to/error.log
info = /location/to/info.log
```

When you've created the configuration you can start the consumer like this:

    $ rabbitmq-cli-consumer -e "/path/to/your/app argument --flag" -c /path/to/your/configuration.conf -V

Run without <code>-V</code> to get rid of the output:

    $ rabbitmq-cli-consumer -e "/path/to/your/app argument --flag" -c /path/to/your/configuration.conf

## Configuration inheritance

By default rabbitmq-cli-consumer looks for configuration files in the following
order:

1. /etc/rabbitmq-cli-consumer/rabbitmq-cli-consumer.conf
2. ~/.rabbitmq-cli-consumer.conf
3. location passed in via <code>-c</code> option

## Override options on the command line:

Every option in the configuration file can be overwritten by passing them on the
cli. Use <code>On</code> and <code>Off</code> for <code>true</code> and <code>false</code>
respectively. Example:

    $ rabbitmq-cli-consumer -c /you/config.conf -e "/path/to/executble" --queue-key=custom-key

### Prefetch count

It's possible to configure the prefetch count and if you want set it as global. Add the following section to your
configuration to control these values:

```ini
[prefetch]
count=3
global=Off
```

### Configuring the exchange

It's also possible to configure the exchange and its options. When left out in the configuration file, the default
exchange will be used. To configure the exchange add the following to your configuration file:

```ini
[exchange]
name=mail
autodelete=Off
type=direct
durable=On
```

### Configuring the queue

All queue options are configurable. Full example:

```ini
[queue]
name=rpc_queue
durable=On
autodelete=Off
exclusive=Off
nowait=Off
```

If you want to define a TTL for your queue:

```ini
[queue]
name=rpc_queue
durable=On
autodelete=Off
exclusive=Off
nowait=Off
ttl=1200
```

When you want to configure the routing key:

```ini
[queue]
name=rpc_queue
durable=On
autodelete=Off
exclusive=Off
nowait=Off
key=your-routing-key
```

## The executable

Your executable receives the message as the last argument. So consider the following:

   $ rabbitmq-cli-consumer -e "/home/vagrant/current/app/command.php" -c example.conf -V

The <code>command.php</code> file should look like this:

```php
#!/usr/bin/env php
<?php
// This contains first argument
$message = $argv[1];

// Decode to get original value
$original = base64_decode($message);

// Start processing
if (do_heavy_lifting($original)) {
    // All well, then return 0
    exit(0);
}

// Let rabbitmq-cli-consumer know someting went wrong, message will be requeued.
exit(1);

```

Or a Symfony2 example:

    $ rabbitmq-cli-consumer -e "/path/to/symfony/app/console event:processing -e=prod" -c example.conf -V

Command looks like this:

```php
<?php

namespace Vendor\EventBundle\Command;

use Symfony\Bundle\FrameworkBundle\Command\ContainerAwareCommand;
use Symfony\Component\Console\Input\InputArgument;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Output\OutputInterface;

class TestCommand extends ContainerAwareCommand
{
    protected function configure()
    {
        $this
            ->addArgument('event', InputArgument::REQUIRED)
            ->setName('test:event')
        ;

    }

    protected function execute(InputInterface $input, OutputInterface $output)
    {
        $message = base64_decode($input->getArgument('event'));

        $this->getContainer()->get('mailer')->send($message);

        exit(0);
    }
}
```

## Compression

Depending on what you're passing around on the queue, it may be wise to enable compression support. If you don't you may
encouter the infamous "Argument list too long" error.

When compression is enabled, the message gets compressed with zlib maximum compression before it's base64 encoded. We
have to pay a performance penalty for this. If you are serializing large php objects I suggest to turn it on. Better
safe then sorry.

In your config:

```ini
[rabbitmq]
host = localhost
username = username-of-rabbitmq-user
password = secret
vhost=/your-vhost
port=5672
queue=name-of-queue
compression=On

[logs]
error = /location/to/error.log
info = /location/to/info.log
```

And in your php app:

```php
#!/usr/bin/env php
<?php
// This contains first argument
$message = $argv[1];

// Decode to get compressed value
$original = base64_decode($message);

// Uncompresss
if (! $original = gzuncompress($original)) {
    // Probably wanna throw some exception here
    exit(1);
}

// Start processing
if (do_heavy_lifting($original)) {
    // All well, then return 0
    exit(0);
}

// Let rabbitmq-cli-consumer know someting went wrong, message will be requeued.
exit(1);

```

## Log rotation

To close and reopen the logs send the USR1 signal:

    $  kill -s USR1 pid_of_process

# Developing

Missing anything? Found a bug? I love to see your PR.

## Setup development environment

It can be quite difficult to get an development environment up & running. I'll hope to
expand the instructions a bit in the future.

### Go and gvm

Todo.

### RabbitMQ

Start by installing docker.

Then:

    $ docker run -d -P -e RABBITMQ_NODENAME=rabbitmq-cli-consumer --name rabbitmq-cli-consumer rabbitmq:3-management

To see which ports are available:

    $ docker port rabbitmq-cli-consumer

You can login with <code>guest</code> / <code>guest</code>.
If you want stop the container:

    $ docker stop rabbitmq-cli-consumer

And to restart the container:

    # docker start rabbitmq-cli-consumer
