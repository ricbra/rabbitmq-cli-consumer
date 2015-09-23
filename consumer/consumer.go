package consumer

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"

	"github.com/ricbra/rabbitmq-cli-consumer/command"
	"github.com/ricbra/rabbitmq-cli-consumer/config"
	"github.com/streadway/amqp"
)

// Consumer represents a consumer
type Consumer struct {
	Channel     *amqp.Channel
	Connection  *amqp.Connection
	Queue       string
	Factory     *command.CommandFactory
	ErrLogger   *log.Logger
	InfLogger   *log.Logger
	Executer    *command.CommandExecuter
	Compression bool
}

// Consume starts consuming messages from RabbitMQ
func (c *Consumer) Consume() {
	c.InfLogger.Println("Registering consumer... ")
	msgs, err := c.Channel.Consume(c.Queue, "", false, false, false, false, nil)
	if err != nil {
		c.ErrLogger.Fatalf("Failed to register a consumer: %s", err)
	}
	c.InfLogger.Println("Succeeded registering consumer.")

	defer c.Connection.Close()
	defer c.Channel.Close()

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			input := d.Body
			if c.Compression {
				var b bytes.Buffer
				w, err := zlib.NewWriterLevel(&b, zlib.BestCompression)
				if err != nil {
					c.ErrLogger.Println("Could not create zlib handler")
					d.Nack(true, true)
				}
				c.InfLogger.Println("Compressed message")
				w.Write(input)
				w.Close()

				input = b.Bytes()
			}

			cmd := c.Factory.Create(base64.StdEncoding.EncodeToString(input))
			if c.Executer.Execute(cmd) {
				d.Ack(true)
			} else {
				d.Nack(true, true)
			}
		}
	}()
	c.InfLogger.Println("Waiting for messages...")
	<-forever
}

// New returns a initialized consumer based on config
func New(cfg *config.Config, factory *command.CommandFactory, errLogger, infLogger *log.Logger) (*Consumer, error) {
	uri := ParseURI(cfg.RabbitMq.Username, cfg.RabbitMq.Password, cfg.RabbitMq.Host, cfg.RabbitMq.Port, cfg.RabbitMq.Vhost)

	infLogger.Println("Connecting RabbitMQ...")
	conn, err := Connect(uri)
	if nil != err {
		return nil, fmt.Errorf("Failed connecting RabbitMQ: %s", err.Error())
	}
	infLogger.Println("Connected.")

	infLogger.Println("Opening channel...")
	ch, err := conn.Channel()
	if nil != err {
		return nil, fmt.Errorf("Failed to open a channel: %s", err.Error())
	}
	infLogger.Println("Done.")

	if err := Initialize(cfg, ch, infLogger, errLogger); err != nil {
		return nil, err
	}

	return &Consumer{
		Channel:     ch,
		Connection:  conn,
		Queue:       cfg.RabbitMq.Queue,
		Factory:     factory,
		ErrLogger:   errLogger,
		InfLogger:   infLogger,
		Executer:    command.New(errLogger, infLogger),
		Compression: cfg.RabbitMq.Compression,
	}, nil
}

// Initialize channel according to config
func Initialize(cfg *config.Config, ch Channel, errLogger, infLogger *log.Logger) error {
	infLogger.Println("Setting QoS... ")

	if err := ch.Qos(cfg.Prefetch.Count, 0, cfg.Prefetch.Global); err != nil {
		return fmt.Errorf("Failed to set QoS: %s", err.Error())
	}

	infLogger.Println("Succeeded setting QoS.")

	infLogger.Printf("Declaring queue \"%s\"...", cfg.RabbitMq.Queue)

	table := amqp.Table{}
	_, err := ch.QueueDeclare(cfg.RabbitMq.Queue, true, false, false, false, table)

	if nil != err {
		return fmt.Errorf("Failed to declare queue: %s", err.Error())
	}

	// Empty Exchange name means default, no need to declare
	if "" != cfg.Exchange.Name {
		infLogger.Printf("Declaring exchange \"%s\"...", cfg.Exchange.Name)
		err = ch.ExchangeDeclare(cfg.Exchange.Name, cfg.Exchange.Type, cfg.Exchange.Durable, cfg.Exchange.Autodelete, false, false, amqp.Table{})

		if nil != err {
			return fmt.Errorf("Failed to declare exchange: %s", err.Error())
		}

		// Bind queue
		infLogger.Printf("Binding queue \"%s\" to exchange \"%s\"...", cfg.RabbitMq.Queue, cfg.Exchange.Name)
		err = ch.QueueBind(cfg.RabbitMq.Queue, "", cfg.Exchange.Name, false, amqp.Table{})

		if nil != err {
			return fmt.Errorf("Failed to bind queue to exchange: %s", err.Error())
		}
	}

	return nil
}

// Connect opens a connection to the given uri
func Connect(uri string) (*amqp.Connection, error) {
	return amqp.Dial(uri)
}

// ParseURI parses the URI based on config
func ParseURI(username, password, host, port, vhost string) string {
	if start := string(vhost[0]); start != "/" {
		vhost = fmt.Sprintf("/%s", vhost)
	}

	return fmt.Sprintf(
		"amqp://%s:%s@%s:%s%s",
		url.QueryEscape(username),
		url.QueryEscape(password),
		host,
		port,
		vhost,
	)
}

// Channel is the interface
type Channel interface {
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	Qos(prefetchCount, prefetchSize int, global bool) error
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
}
