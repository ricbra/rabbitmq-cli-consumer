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
       --include, -i	Include metadata. Passes message as JSON data including headers, properties and message body.
       --help, -h		show help
       --version, -v	print the version

## Configuration

A configuration file is required. Example:

```ini
[rabbitmq]
host = localhost
username = username-of-rabbitmq-user
password = secret
vhost=/your-vhost
port=5672
queue=name-of-queue
compression=Off

[logs]
error = /location/to/error.log
info = /location/to/info.log
```

When you've created the configuration you can start the consumer like this:

    $ rabbitmq-cli-consumer -e "/path/to/your/app argument --flag" -c /path/to/your/configuration.conf -V

Run without <code>-V</code> to get rid of the output:

    $ rabbitmq-cli-consumer -e "/path/to/your/app argument --flag" -c /path/to/your/configuration.conf

### Prefetch count

It's possible to configure the prefetch count and if you want set it as global. Add the following section to your
configuration to confol these values:

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
            ->setName('event:processing')
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

## Including properties and message headers


If you need to access message headers or properties, call the command with the
`--include, -i` option set.

    $ rabbitmq-cli-consumer -e "/home/vagrant/current/app/command.php" -c example.conf -i

The script then will receive a json encoded data structure which looks like
the following.

```json
{
  "properties": {
    "application_headers": {
      "name": "value"
    },
    "content_type": "",
    "content_encoding": "",
    "delivery_mode": 1,
    "priority": 0,
    "correlation_id": "",
    "reply_to": "",
    "expiration": "",
    "message_id": "",
    "timestamp": "0001-01-01T00:00:00Z",
    "type": "",
    "user_id": "",
    "app_id": ""
  },
  "delivery_info": {
    "message_count": 0,
    "consumer_tag": "ctag-./rabbitmq-cli-consumer-1",
    "delivery_tag": 2,
    "redelivered": true,
    "exchange": "example",
    "routing_key": ""
  },
  "body": ""
}

```

Change your script acording to the following example.

```php
#!/usr/bin/env php
<?php
// This contains first argument
$input = $argv[1];

// Decode to get original value also decrompress acording to your configuration.
$data = json_decode(base64_decode($input));

// Start processing
if (do_heavy_lifting($data->body, $data->properties)) {
    // All well, then return 0
    exit(0);
}

// Let rabbitmq-cli-consumer know someting went wrong, message will be requeued.
exit(1);
```

If you are using symfonies RabbitMQ bundle (`oldsound/rabbitmq-bundle`) you can
wrap the consumer with the following symfony command.

```php
<?php

namespace Vendor\EventBundle\Command;

use PhpAmqpLib\Message\AMQPMessage;
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
            ->setName('event:processing')
        ;

    }

    protected function execute(InputInterface $input, OutputInterface $output)
    {
        $data = json_decode(base64_decode($input->getArgument('event')));
        $message = new AMQPMessage($data->body, $data->properties);

        /** @var \PhpAmqpLib\Message\AMQPMessage\ConsumerInterface $consumer */
        $consumer = $this->getContainer()->get('consumer');

        if (false == $consumer->execute($message)) {
		exit(1);
	}
    }
}
```



# Developing

Missing anything? Found a bug? I love to see your PR.


