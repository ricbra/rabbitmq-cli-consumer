package config

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFailsOnRequiredFields(t *testing.T) {
	config := createConfig(`[rabbitmq]
    host=`)

	var b bytes.Buffer
	logger := log.New(&b, "", 0)

	valid := Validate(config, logger)
	out := b.String()

	assert.Equal(t, false, valid)
	assert.Contains(t, out, "The option \"queue\" under section \"rabbitmq\" is invalid: This option is required")
	assert.Contains(t, out, "The option \"port\" under section \"rabbitmq\" is invalid: This option is required")
	assert.Contains(t, out, "The option \"username\" under section \"rabbitmq\" is invalid: This option is required")
	assert.Contains(t, out, "The option \"password\" under section \"rabbitmq\" is invalid: This option is required")
	assert.Contains(t, out, "The option \"vhost\" under section \"rabbitmq\" is invalid: This option is required")
	assert.Contains(t, out, "The option \"host\" under section \"rabbitmq\" is invalid: This option is required")
	assert.Contains(t, out, "The option \"count\" under section \"prefetch\" is invalid: This option is required")
	assert.Contains(t, out, "The option \"name\" under section \"exchange\" is invalid: This option is required")
	assert.Contains(t, out, "The option \"type\" under section \"exchange\" is invalid: This option is required")
	assert.Contains(t, out, "The option \"error\" under section \"logs\" is invalid: This option is required")
	assert.Contains(t, out, "The option \"info\" under section \"logs\" is invalid: This option is required")
}

func TestPassOnValidConfig(t *testing.T) {
	config := createConfig(
		`[rabbitmq]
    host=localhost
    username=test
    password=t3st
    vhost=test
    queue=test
    port=123

    [prefetch]
    count=3
    global=On

    [exchange]
    name=test
    autodelete=Off
    type=test
    durable=On

    [logs]
    error=a
    info=b
    `)

	var b bytes.Buffer
	logger := log.New(&b, "", 0)
	valid := Validate(config, logger)
	assert.Equal(t, true, valid)
}
