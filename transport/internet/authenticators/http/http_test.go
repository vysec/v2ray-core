package http_test

import (
	"testing"

	"v2ray.com/core/common/alloc"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/testing/assert"
	. "v2ray.com/core/transport/internet/authenticators/http"
)

func TestReaderWriter(t *testing.T) {
	assert := assert.On(t)

	cache := alloc.NewBuffer()
	b := alloc.NewLocalBuffer(256)
	b.AppendFunc(serial.WriteString("abcd" + ENDING))
	writer := NewHeaderWriter(b)
	writer.Write(cache)
	cache.Write([]byte{'e', 'f', 'g'})

	reader := &HeaderReader{}
	buffer, err := reader.Read(cache)
	assert.Error(err).IsNil()
	assert.Bytes(buffer.Bytes()).Equals([]byte{'e', 'f', 'g'})
}

func TestRequestHeader(t *testing.T) {
	assert := assert.On(t)

	factory := HttpAuthenticatorFactory{}
	auth := factory.Create(&Config{
		Request: &RequestConfig{
			Uri: []string{"/"},
			Header: []*Header{
				{
					Name:  "Test",
					Value: []string{"Value"},
				},
			},
		},
	}).(HttpAuthenticator)

	cache := alloc.NewBuffer()
	err := auth.GetClientWriter().Write(cache)
	assert.Error(err).IsNil()

	assert.String(cache.String()).Equals("GET / HTTP/1.1\r\nTest: Value\r\n\r\n")
}
