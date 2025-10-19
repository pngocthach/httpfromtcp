package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHeaders(t *testing.T) {
	// Valid single header with lowercase key
	headers := NewHeaders()
	data := []byte("host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Valid single header with uppercase key (should normalize to lowercase)
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Valid single header with mixed-case key (should normalize to lowercase)
	headers = NewHeaders()
	data = []byte("Content-Type: application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "application/json", headers["content-type"])
	assert.Equal(t, 32, n)
	assert.False(t, done)

	// Valid header with extra whitespace (OWS should be trimmed)
	headers = NewHeaders()
	data = []byte("Host:    localhost:42069    \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 30, n)
	assert.False(t, done)

	// Empty line signals end of headers
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Invalid: space before colon
	headers = NewHeaders()
	data = []byte("Host : localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Invalid: non-ASCII character in header key
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Invalid: space in header key
	headers = NewHeaders()
	data = []byte("Invalid Key: some value\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestParseMultipleHeaders(t *testing.T) {
	t.Run("Valid multiple headers", func(t *testing.T) {
		headers := NewHeaders()
		// Parse is called multiple times to simulate incremental parsing

		// First call: parse first header line
		data := []byte("Set-Person: A\r\nset-person: B\r\n\r\n")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.False(t, done)
		assert.Equal(t, 15, n)
		assert.Equal(t, "A", headers["set-person"])

		// Second call: parse remaining data
		n2, done2, err2 := headers.Parse(data[n:])

		require.NoError(t, err2)
		assert.False(t, done2)
		assert.Equal(t, 15, n2)
		assert.Equal(t, "A,B", headers["set-person"])

		// Third call: parse final empty line
		n3, done3, err3 := headers.Parse(data[n+n2:])

		require.NoError(t, err3)
		assert.True(t, done3)
		assert.Equal(t, 2, n3)
	})
}
