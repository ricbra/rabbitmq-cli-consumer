package consumer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpensChannel(t *testing.T) {
	t.Log("test")
}

func TestParseAndEscapesParamsInURI(t *testing.T) {
	uri := ParseURI("richard", "my@:secr%t", "localhost", "123", "/vhost")

	assert.Equal(t, "amqp://richard:my%40%3Asecr%25t@localhost:123/vhost", uri)
}

func TestAddsSlashWhenMissingInVhost(t *testing.T) {
	uri := ParseURI("richard", "secret", "localhost", "123", "vhost")

	assert.Equal(t, "amqp://richard:secret@localhost:123/vhost", uri)
}
