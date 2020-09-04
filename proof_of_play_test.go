package vistar

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testEvent struct {
	name    string
	message string
	source  string
	level   string
}

func TestConfirmSuccess(t *testing.T) {
	ad := Ad{
		"id":                "ad-id",
		"proof_of_play_url": "http://pop-url.com",
	}

	eventCalls := make([]*eventCall, 0, 0)
	eventFn := func(name string, message string, source string, level string) {
		eventCalls = append(eventCalls, &eventCall{
			name:    name,
			message: message,
			source:  source,
			level:   level})
	}
	mockPopFunc := func(method string, url string,
		data *ProofOfPlayRequest) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK}, nil
	}

	p := NewProofOfPlay(eventFn, mockPopFunc)

	err := p.Confirm(ad, int64(100))

	assert.Nil(t, err)
	assert.Len(t, eventCalls, 0)
}

func TestConfirmFail(t *testing.T) {
	ad := Ad{
		"id":                "ad-id",
		"proof_of_play_url": "http://pop-url.com",
	}

	eventCalls := make([]*eventCall, 0, 0)
	eventFn := func(name string, message string, source string, level string) {
		eventCalls = append(eventCalls, &eventCall{
			name:    name,
			message: message,
			source:  source,
			level:   level})
	}
	mockPopFunc := func(method string, url string,
		data *ProofOfPlayRequest) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(bytes.NewBufferString("Bad request")),
		}, errors.New("Request failed!!")
	}

	p := NewProofOfPlay(eventFn, mockPopFunc)

	err := p.Confirm(ad, int64(100))

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Request failed!!")
	assert.Len(t, eventCalls, 1)

	event := eventCalls[0]
	assert.Equal(t, event.name, "ad-pop-failed")
	assert.Equal(t, event.level, "warning")
	assert.Equal(t, event.source, "")
	assert.Equal(t, event.message, "adId: ad-id, error: Bad request")
}

func TestExpireSuccess(t *testing.T) {
	ad := Ad{
		"id":             "ad-id",
		"expiration_url": "http://expiration-url.com",
	}

	eventCalls := make([]*eventCall, 0, 0)
	eventFn := func(name string, message string, source string, level string) {
		eventCalls = append(eventCalls, &eventCall{
			name:    name,
			message: message,
			source:  source,
			level:   level})
	}
	mockPopFunc := func(method string, url string,
		data *ProofOfPlayRequest) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK}, nil
	}

	p := NewProofOfPlay(eventFn, mockPopFunc)

	err := p.Expire(ad)

	assert.Nil(t, err)
	assert.Len(t, eventCalls, 0)
}

func TestExpireFail(t *testing.T) {
	ad := Ad{
		"id":             "ad-id",
		"expiration_url": "http://expiration-url.com",
	}

	eventCalls := make([]*eventCall, 0, 0)
	eventFn := func(name string, message string, source string, level string) {
		eventCalls = append(eventCalls, &eventCall{
			name:    name,
			message: message,
			source:  source,
			level:   level})
	}
	mockPopFunc := func(method string, url string,
		data *ProofOfPlayRequest) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(bytes.NewBufferString("Bad request")),
		}, errors.New("Request failed!!")
	}

	p := NewProofOfPlay(eventFn, mockPopFunc)

	err := p.Expire(ad)

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Request failed!!")
	assert.Len(t, eventCalls, 1)

	event := eventCalls[0]
	assert.Equal(t, event.name, "ad-expire-failed")
	assert.Equal(t, event.level, "warning")
	assert.Equal(t, event.source, "")
	assert.Equal(t, event.message, "adId: ad-id, error: Bad request")
}
