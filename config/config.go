package config

import (
	"path/filepath"

	"gopkg.in/gcfg.v1"
)

type Config struct {
	RabbitMq struct {
		Host        string
		Username    string
		Password    string
		Port        string
		Vhost       string
		Queue       string
		Compression bool
		Onfailure   int
	}
	Prefetch struct {
		Count  int
		Global bool
	}
	QueueSettings struct {
		Routingkey           string
		MessageTTL           int
		DeadLetterExchange   string
		DeadLetterRoutingKey string
	}
	Exchange struct {
		Name       string
		Autodelete bool
		Type       string
		Durable    bool
	}
	Logs struct {
		Error string
		Info  string
	}
}

func LoadAndParse(location string) (*Config, error) {
	if !filepath.IsAbs(location) {
		location, err := filepath.Abs(location)

		if err != nil {
			return nil, err
		}

		location = location
	}

	cfg := Config{}
	if err := gcfg.ReadFileInto(&cfg, location); err != nil {
		return nil, err
	}

	return &cfg, nil
}
