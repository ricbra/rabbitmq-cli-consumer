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
	Channel     Channel
	Connection  *amqp.Connection
	Queue       string
	Factory     *command.CommandFactory
	ErrLogger   *log.Logger
	InfLogger   *log.Logger
	Executer    command.Executer
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
			delivery := RabbitMqDelivery{
				delivery: d,
				body:     d.Body,
			}
			c.ProcessMessage(delivery)
		}
	}()
	c.InfLogger.Println("Waiting for messages...")
	<-forever
}

// ProcessMessage content of one single message
func (c *Consumer) ProcessMessage(msg Delivery) {
	input := msg.Body()
	if c.Compression {
		var b bytes.Buffer
		w, err := zlib.NewWriterLevel(&b, zlib.BestCompression)
		if err != nil {
			c.ErrLogger.Println("Could not create zlib handler")
			return
		}
		c.InfLogger.Println("Decompressed message")
		w.Write(input)
		w.Close()

		input = b.Bytes()
	}

	cmd := c.Factory.Create(base64.StdEncoding.EncodeToString(input))
	out, err := c.Executer.Execute(cmd)

	if err != nil {
		msg.Nack(true, true)
		return
	}

	if msg.IsRpcMessage() {
		c.InfLogger.Println("Message is RPC message, trying to send reply...")

		if err := c.Reply(msg, out); err != nil {
			c.InfLogger.Println("Sending RPC reply failed. Check error log.")
			c.ErrLogger.Printf("Error occured during send RPC reply: %s", err)
			msg.Nack(true, true)

			return
		}
		c.InfLogger.Println("RPC reply send.")
	}

	// All went fine, ack message
	msg.Ack(true)
}

func (c *Consumer) Reply(msg Delivery, out []byte) error {
	return c.Channel.Publish(
		"",
		msg.ReplyTo(),
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: msg.CorrelationId(),
			Body:          out,
		},
	)
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
		Queue:       cfg.Queue.Name,
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

	infLogger.Printf("Declaring queue \"%s\"...", cfg.Queue.Name)

	table := amqp.Table{}
	if cfg.Queue.TTL > 0 {
		table["x-message-ttl"] = cfg.Queue.TTL
	}
	_, err := ch.QueueDeclare(cfg.Queue.Name, cfg.Queue.Durable, cfg.Queue.Autodelete, cfg.Queue.Exclusive, cfg.Queue.Nowait, table)

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
		infLogger.Printf("Binding queue \"%s\" to exchange \"%s\"...", cfg.Queue.Name, cfg.Exchange.Name)
		err = ch.QueueBind(cfg.Queue.Name, cfg.Queue.Key, cfg.Exchange.Name, false, amqp.Table{})

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
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Close() error
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
}

// Delivery interface describes interface for messages
type Delivery interface {
	Ack(multiple bool) error
	Nack(multiple, requeue bool) error
	Body() []byte
	IsRpcMessage() bool
	CorrelationId() string
	ReplyTo() string
}

type RabbitMqDelivery struct {
	body     []byte
	delivery amqp.Delivery
}

func (r RabbitMqDelivery) Ack(multiple bool) error {
	return r.delivery.Ack(multiple)
}

func (r RabbitMqDelivery) Nack(multiple, requeue bool) error {
	return r.delivery.Nack(multiple, requeue)
}

func (r RabbitMqDelivery) Body() []byte {
	return r.body
}

func (r RabbitMqDelivery) IsRpcMessage() bool {
	return r.delivery.ReplyTo != "" && r.delivery.CorrelationId != ""
}

func (r RabbitMqDelivery) CorrelationId() string {
	return r.delivery.CorrelationId
}

func (r RabbitMqDelivery) ReplyTo() string {
	return r.delivery.ReplyTo
}
