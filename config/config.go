package config

import (
	"gopkg.in/gcfg.v1"
	"path/filepath"
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
	}
	Prefetch struct {
		Count     int
		Global    bool
	}
	Exchange struct {
		Name		string
		Autodelete	bool
		Type		string
		Durable		bool
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
