package config

import (
	"testing"

	"code.google.com/p/gcfg"
	"github.com/stretchr/testify/assert"
)

func TestMergesConfigs(t *testing.T) {
	configs := []Config{
		createConfig(`[rabbitmq]
	host=rabbitmq.provider.com
	password=123pass
	vhost=test`),
		createConfig(`[rabbitmq]
	host=localhost
	queue=testqueue`),
	}

	merger := ConfigMerger{}
	config, _ := merger.Merge(configs)

	assert.Equal(t, config.RabbitMq.Host, "localhost")
	assert.Equal(t, config.RabbitMq.Password, "123pass")
	assert.Equal(t, config.RabbitMq.Vhost, "test")
}

func createConfig(config string) Config {
	cfg := Config{}
	gcfg.ReadStringInto(&cfg, config)

	return cfg
}
