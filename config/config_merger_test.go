package config

import (
	"fmt"
	"testing"

	"code.google.com/p/gcfg"
)

func TestMergesConfigs(t *testing.T) {
	configs := []Config{
		createConfig("localhost", "queue1"),
		createConfig("test.host.com", "queue2"),
	}

	merger := ConfigMerger{}
	merger.Merge(configs)
}

func createConfig(host, queue string) Config {
	cfg := Config{}
	cfgStr := fmt.Sprintf(`[rabbitmq]
host=%s
queue=%s`, host, queue)
	gcfg.ReadStringInto(&cfg, cfgStr)

	return cfg
}
