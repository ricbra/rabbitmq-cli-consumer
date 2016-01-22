package consumer

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/ricbra/rabbitmq-cli-consumer/command"
	"github.com/ricbra/rabbitmq-cli-consumer/config"
	"github.com/streadway/amqp"
	"log"
	"net/url"
	"encoding/json"
	"time"
)

type Consumer struct {
	Channel         *amqp.Channel
	Connection      *amqp.Connection
	Queue           string
	Factory         *command.CommandFactory
	ErrLogger       *log.Logger
	InfLogger       *log.Logger
	Executer        *command.CommandExecuter
	Compression     bool
	IncludeMetadata bool
}

type Properties struct{
	Headers         amqp.Table `json:"application_headers"`
	ContentType     string     `json:"content_type"`
	ContentEncoding string     `json:"content_encoding"`
	DeliveryMode    uint8      `json:"delivery_mode"`
	Priority        uint8      `json:"priority"`
	CorrelationId   string     `json:"correlation_id"`
	ReplyTo         string     `json:"reply_to"`
	Expiration      string     `json:"expiration"`
	MessageId       string     `json:"message_id"`
	Timestamp       time.Time  `json:"timestamp"`
	Type            string     `json:"type"`
	UserId          string     `json:"user_id"`
	AppId           string     `json:"app_id"`
}

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
			if c.IncludeMetadata {
				input, err = json.Marshal(&struct {
					Properties             `json:"properties"`
					Body        string     `json:"body"`
				}{

					Properties: Properties {
						Headers:         d.Headers,
						ContentType:     d.ContentType,
						ContentEncoding: d.ContentEncoding,
						DeliveryMode:    d.DeliveryMode,
						Priority:        d.Priority,
						CorrelationId:   d.CorrelationId,
						ReplyTo:         d.ReplyTo,
						Expiration:      d.Expiration,
						MessageId:       d.MessageId,
						Timestamp:       d.Timestamp,
						Type:            d.Type,
						UserId:          d.UserId,
					},

					Body:            string(d.Body),
				})
				if err != nil {
					c.ErrLogger.Fatalf("Failed to marshall: %s", err)
					d.Nack(true, true)
				}
			}
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

func New(cfg *config.Config, factory *command.CommandFactory, errLogger, infLogger *log.Logger) (*Consumer, error) {
	uri := fmt.Sprintf(
		"amqp://%s:%s@%s:%s%s",
		url.QueryEscape(cfg.RabbitMq.Username),
		url.QueryEscape(cfg.RabbitMq.Password),
		cfg.RabbitMq.Host,
		cfg.RabbitMq.Port,
		cfg.RabbitMq.Vhost,
	)

	infLogger.Println("Connecting RabbitMQ...")
	conn, err := amqp.Dial(uri)
	if nil != err {
		return nil, errors.New(fmt.Sprintf("Failed connecting RabbitMQ: %s", err.Error()))
	}
	infLogger.Println("Connected.")

	infLogger.Println("Opening channel...")
	ch, err := conn.Channel()
	if nil != err {
		return nil, errors.New(fmt.Sprintf("Failed to open a channel: %s", err.Error()))
	}
	infLogger.Println("Done.")

	infLogger.Println("Setting QoS... ")
	// Attempt to preserve BC here
	if cfg.Prefetch.Count == 0 {
		cfg.Prefetch.Count = 3
	}
	if err := ch.Qos(cfg.Prefetch.Count, 0, cfg.Prefetch.Global); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to set QoS: %s", err.Error()))
	}
	infLogger.Println("Succeeded setting QoS.")

	infLogger.Printf("Declaring queue \"%s\"...", cfg.RabbitMq.Queue)
	_, err = ch.QueueDeclare(cfg.RabbitMq.Queue, true, false, false, false, nil)

	if nil != err {
		return nil, errors.New(fmt.Sprintf("Failed to declare queue: %s", err.Error()))
	}

	// Check for missing exchange settings to preserve BC
	if "" == cfg.Exchange.Name && "" == cfg.Exchange.Type && !cfg.Exchange.Durable && !cfg.Exchange.Autodelete {
		cfg.Exchange.Type = "direct"
	}

	// Empty Exchange name means default, no need to declare
	if "" != cfg.Exchange.Name {
		infLogger.Printf("Declaring exchange \"%s\"...", cfg.Exchange.Name)
		err = ch.ExchangeDeclare(cfg.Exchange.Name, cfg.Exchange.Type, cfg.Exchange.Durable, cfg.Exchange.Autodelete, false, false, amqp.Table{})

		if nil != err {
			return nil, errors.New(fmt.Sprintf("Failed to declare exchange: %s", err.Error()))
		}

		// Bind queue
		infLogger.Printf("Binding queue \"%s\" to exchange \"%s\"...", cfg.RabbitMq.Queue, cfg.Exchange.Name)
		err = ch.QueueBind(cfg.RabbitMq.Queue, "", cfg.Exchange.Name, false, nil)

		if nil != err {
			return nil, errors.New(fmt.Sprintf("Failed to bind queue to exchange: %s", err.Error()))
		}
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
