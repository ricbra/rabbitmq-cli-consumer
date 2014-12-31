package consumer

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/ricbra/rabbitmq-cli-consumer/command"
	"github.com/ricbra/rabbitmq-cli-consumer/config"
	"github.com/streadway/amqp"
	"os"
)

type Consumer struct {
	Channel    *amqp.Channel
	Connection *amqp.Connection
	Queue      string
	Factory    *command.CommandFactory
}

func (c *Consumer) Consume() {
	if err := c.Channel.Qos(3, 0, false); err != nil {
		panic(fmt.Sprintf("Failed to set QoS: %s", err))
	}

	msgs, err := c.Channel.Consume(c.Queue, "", false, false, false, false, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		panic(fmt.Sprintf("Failed to register a consumer: %s"))
	}

	defer c.Connection.Close()
	defer c.Channel.Close()

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			input := base64.StdEncoding.EncodeToString(d.Body)
			cmd := c.Factory.Create(input)
			if command.Execute(cmd) {
				d.Ack(true)
			} else {
				d.Nack(true, true)
			}
		}
	}()
	fmt.Println("  [*] Waiting for messages")
	<-forever
}

func New(cfg *config.Config, factory *command.CommandFactory) (*Consumer, error) {
	uri := fmt.Sprintf(
		"amqp://%s:%s@%s:%s%s",
		cfg.RabbitMq.Username,
		cfg.RabbitMq.Password,
		cfg.RabbitMq.Host,
		cfg.RabbitMq.Port,
		cfg.RabbitMq.Vhost,
	)
	conn, err := amqp.Dial(uri)
	if nil != err {
		return nil, errors.New(fmt.Sprintf("Failed connecting RabbitMQ: %s", err.Error()))
	}

	ch, err := conn.Channel()
	if nil != err {
		return nil, errors.New(fmt.Sprintf("Failed to open a channel: %s", err.Error()))
	}

	_, err = ch.QueueDeclare(cfg.RabbitMq.Queue, true, false, false, false, nil)

	if nil != err {
		return nil, errors.New(fmt.Sprintf("Failed to declare queue: %s", err.Error()))
	}

	return &Consumer{
		Channel:    ch,
		Connection: conn,
		Queue:      cfg.RabbitMq.Queue,
		Factory:    factory,
	}, nil
}
