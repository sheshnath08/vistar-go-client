package vistar

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRequestLength(t *testing.T) {
	data := []byte("this is test")

	req := httptest.NewRequest(
		"GET", "http://example.com/foo", bytes.NewBuffer(data))

	assert.Equal(t, getRequestLength(req), int64(len(data)))

	req.Header.Set("Content-Type", "application/json")

	// Header length is 28 and content length is 12 hence total is 40
	assert.Equal(t, getRequestLength(req), int64(40))

}

func TestGetResponseLength(t *testing.T) {
	assert.Equal(t, getResponseLength(nil), int64(0))

	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "<html><body>Hello World!</body></html>")
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	// Response length is 116 and it looks like following:
	// HTTP/1.1 200 OK
	// Connection: close
	// Content-Type: text/html; charset=utf-8
	// <html><body>Hello World!</body></html>
	assert.Equal(t, getResponseLength(resp), int64(116))
}

func TestUpdateStats(t *testing.T) {
	stats := Stats{}

	updateStats(&stats, 100, 1024)
	assert.Equal(t, stats.Count, int64(1))
	assert.Equal(t, stats.BytesSent, int64(100))
	assert.Equal(t, stats.BytesReceived, int64(1024))
	assert.Equal(t, stats.Total, int64(1124))
	assert.Equal(t, stats.Average, float64(1124))

	updateStats(&stats, 50, 2048)
	assert.Equal(t, stats.Count, int64(2))
	assert.Equal(t, stats.BytesSent, int64(150))
	assert.Equal(t, stats.BytesReceived, int64(3072))
	assert.Equal(t, stats.Total, int64(3222))
	assert.Equal(t, stats.Average, float64(1611))
}
