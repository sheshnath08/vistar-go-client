package vistar

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalAverage(t *testing.T) {
	assert.Equal(t, calcAverage(int64(100), int64(3)), "33.33")
	assert.Equal(t, calcAverage(int64(100), int64(10)), "10.00")
}

func TestGetRequestLength(t *testing.T) {
	data := []byte("this is test")

	req := httptest.NewRequest(
		"GET", "http://example.com/foo", bytes.NewBuffer(data))

	assert.Equal(t, getRequestLength(req), int64(14))

	req.Header.Set("Content-Type", "application/json")
	assert.Equal(t, getRequestLength(req), int64(49))

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

	assert.Equal(t, getResponseLength(resp), int64(116))
}
