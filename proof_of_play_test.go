package vistar

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func NewTestHttpClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

type testEvent struct {
	name    string
	message string
	source  string
	level   string
}

func TestStop(t *testing.T) {
	pop := NewProofOfPlay(nil)

	ad := Ad{}
	req := &PoPRequest{
		Ad:          ad,
		Status:      false,
		RequestTime: time.Now(),
	}
	pop.processRequestFailure(req, nil)

	recv := <-pop.retryQueue
	assert.Equal(t, ad, recv.Ad)
	assert.False(t, recv.Status)

	pop.Stop()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ProofOfPlay.Stop() didn't cause a panic")
		}
	}()

	pop.processRequestFailure(req, nil)
}

func TestIsLeaseExpired(t *testing.T) {
	pop := NewProofOfPlay(nil)
	pop.Stop()

	expired := float64(time.Now().Add(-1 * time.Hour).Unix())
	ad := Ad{"lease_expiry": expired}
	isExpired := pop.isLeaseExpired(ad)
	assert.True(t, isExpired)

	unexpired := float64(time.Now().Add(1 * time.Hour).Unix())
	ad2 := Ad{"lease_expiry": unexpired}
	isExpired = pop.isLeaseExpired(ad2)
	assert.False(t, isExpired)
}

func TestProcessRetry(t *testing.T) {
	pop := &proofOfPlay{
		requests:   make(chan *PoPRequest, 100),
		retryQueue: make(chan *PoPRequest, 100),
	}

	defer pop.Stop()

	now := time.Now()
	diff := 100 * time.Millisecond
	reqTime := now.Add(-(RetryInterval - diff))
	unexp := float64(now.Add(1 * time.Hour).Unix())
	unexpAd := Ad{"id": "unexp", "lease_expiry": unexp}
	unexpReq := &PoPRequest{Ad: unexpAd, Status: true, RequestTime: reqTime}
	pop.processRetry(unexpReq)
	assert.Equal(t, len(pop.requests), 1)

	exp := float64(now.Add(-RetryInterval).Unix())
	expiredAd := Ad{"id": "expired", "lease_expiry": exp}
	expiredReq := &PoPRequest{Ad: expiredAd, Status: true, RequestTime: reqTime}
	pop.processRetry(expiredReq)
	assert.Equal(t, len(pop.requests), 1)

	unexpReq = &PoPRequest{Ad: unexpAd, Status: false, RequestTime: reqTime}
	pop.processRetry(unexpReq)
	assert.Equal(t, len(pop.requests), 2)

	// Retry requests should be delayed
	since := time.Since(now)
	assert.True(t, since >= diff)
}

func TestMakePoPRequestOKResponse(t *testing.T) {
	client := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		respData, _ := json.Marshal(map[string]interface{}{"msg": "It was OK"})
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(respData)),
		}, nil
	})

	pop := &proofOfPlay{
		requests:       make(chan *PoPRequest, 100),
		retryQueue:     make(chan *PoPRequest, 100),
		httpClient:     client,
		bandwidthStats: Stats{},
	}

	defer pop.Stop()

	adId := "1234"

	ad := Ad{"id": adId}
	popReq := &PoPRequest{Ad: ad, Status: true, RequestTime: time.Now()}
	reqData, _ := json.Marshal(popReq)
	req, _ := http.NewRequest("POST", "/adserver/pop", bytes.NewBuffer(reqData))

	err := pop.makePoPRequest(req, popReq, adId, "pop")
	assert.Nil(t, err)

	assert.Equal(t, pop.bandwidthStats.Count, int64(1))
}

func TestMakePoPRequest400Response(t *testing.T) {
	client := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		respData, _ := json.Marshal(map[string]interface{}{"msg": "Bad Request"})
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(bytes.NewReader(respData)),
		}, nil
	})

	pop := &proofOfPlay{
		requests:       make(chan *PoPRequest, 100),
		retryQueue:     make(chan *PoPRequest, 100),
		httpClient:     client,
		bandwidthStats: Stats{},
	}

	defer pop.Stop()

	adId := "1234"
	ad := Ad{"id": adId}
	popReq := &PoPRequest{Ad: ad, Status: true, RequestTime: time.Now()}
	reqData, _ := json.Marshal(popReq)
	req, _ := http.NewRequest("POST", "/adserver/pop", bytes.NewBuffer(reqData))

	err := pop.makePoPRequest(req, popReq, adId, "pop")
	popErr, ok := err.(*PoPError)
	assert.True(t, ok)
	assert.Equal(t, popErr.Status, http.StatusBadRequest)
	assert.Len(t, pop.retryQueue, 0)
	assert.Len(t, pop.requests, 0)

	assert.Equal(t, pop.bandwidthStats.Count, int64(1))
}

func TestMakePoPRequest500Response(t *testing.T) {
	client := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		respData, _ := json.Marshal(map[string]interface{}{"msg": "Conn error"})
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader(respData)),
		}, nil
	})

	pop := &proofOfPlay{
		requests:       make(chan *PoPRequest, 100),
		retryQueue:     make(chan *PoPRequest, 100),
		httpClient:     client,
		bandwidthStats: Stats{},
	}

	defer pop.Stop()

	adId := "1234"
	ad := Ad{"id": adId}
	popReq := &PoPRequest{Ad: ad, Status: true, RequestTime: time.Now()}
	reqData, _ := json.Marshal(popReq)
	req, _ := http.NewRequest("POST", "/adserver/pop", bytes.NewBuffer(reqData))

	err := pop.makePoPRequest(req, popReq, adId, "pop")
	popErr, ok := err.(*PoPError)
	assert.True(t, ok)
	assert.Equal(t, popErr.Status, http.StatusAccepted)

	assert.Len(t, pop.retryQueue, 1)

	assert.Equal(t, pop.bandwidthStats.Count, int64(1))
}

func TestMakePoPRequestErrors(t *testing.T) {
	client := NewTestHttpClient(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("HTTP Client error")
	})

	pop := &proofOfPlay{
		requests:       make(chan *PoPRequest, 100),
		retryQueue:     make(chan *PoPRequest, 100),
		httpClient:     client,
		bandwidthStats: Stats{},
	}

	defer pop.Stop()

	adId := "1234"
	ad := Ad{"id": adId}
	popReq := &PoPRequest{Ad: ad, Status: true, RequestTime: time.Now()}
	reqData, _ := json.Marshal(popReq)
	req, _ := http.NewRequest("POST", "/adserver/pop", bytes.NewBuffer(reqData))

	err := pop.makePoPRequest(req, popReq, adId, "pop")
	popErr, ok := err.(*PoPError)
	assert.True(t, ok)
	assert.Equal(t, popErr.Status, http.StatusAccepted)

	assert.Len(t, pop.retryQueue, 1)

	assert.Equal(t, pop.bandwidthStats.Count, int64(0))
}

func TestPopUpdateBandwidthStats(t *testing.T) {
	pop := &proofOfPlay{
		bandwidthStats: Stats{},
	}

	pop.updateBandwidthStats(int64(100), int64(1024))

	assert.Equal(t, pop.bandwidthStats.Count, int64(1))
	assert.Equal(t, pop.bandwidthStats.BytesSent, int64(100))
	assert.Equal(t, pop.bandwidthStats.BytesReceived, int64(1024))
	assert.Equal(t, pop.bandwidthStats.Total, int64(1124))
	assert.Equal(t, pop.bandwidthStats.Average, float64(1124))

	pop.updateBandwidthStats(int64(100), int64(1024))

	assert.Equal(t, pop.bandwidthStats.Count, int64(2))
	assert.Equal(t, pop.bandwidthStats.BytesSent, int64(200))
	assert.Equal(t, pop.bandwidthStats.BytesReceived, int64(2048))
	assert.Equal(t, pop.bandwidthStats.Total, int64(2248))
	assert.Equal(t, pop.bandwidthStats.Average, float64(1124))
}
