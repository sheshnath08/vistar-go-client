package vistar

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testEvent struct {
	name    string
	message string
	source  string
	level   string
}

func TestStop(t *testing.T) {
	pop := NewProofOfPlay(nil)

	ad := Ad{}
	pop.Expire(ad)

	recv := <-pop.requests
	assert.Equal(t, ad, recv.Ad)
	assert.False(t, recv.Status)

	pop.Stop()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ProofOfPlay.Stop() didn't cause a panic")
		}
	}()

	pop.Expire(ad)
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

func TestRetryPoP(t *testing.T) {
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
	pop.retryPoP(unexpReq)
	assert.Equal(t, len(pop.requests), 1)

	exp := float64(now.Add(-RetryInterval).Unix())
	expiredAd := Ad{"id": "expired", "lease_expiry": exp}
	expiredReq := &PoPRequest{Ad: expiredAd, Status: true, RequestTime: reqTime}
	pop.retryPoP(expiredReq)
	assert.Equal(t, len(pop.requests), 1)

	unexpReq = &PoPRequest{Ad: unexpAd, Status: false, RequestTime: reqTime}
	pop.retryPoP(unexpReq)
	assert.Equal(t, len(pop.requests), 2)

	// Retry requests should be delayed
	since := time.Since(now)
	assert.True(t, since >= diff)
}

func TestProcessResponse(t *testing.T) {
	pop := &proofOfPlay{
		requests:   make(chan *PoPRequest, 100),
		retryQueue: make(chan *PoPRequest, 100),
	}

	defer pop.Stop()

	okResp := &http.Response{StatusCode: http.StatusOK}
	err := pop.processResponse("pop", "ad1", okResp)
	assert.Nil(t, err)

	serverErrorResp := &http.Response{
		StatusCode: http.StatusInternalServerError,
	}
	err = pop.processResponse("pop", "ad1", serverErrorResp)
	assert.NotNil(t, err)

	errMsg := "A 400 error occurred"
	data, _ := json.Marshal(map[string]interface{}{"msg": errMsg})
	badRequestResponse := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       ioutil.NopCloser(bytes.NewReader(data)),
	}
	err = pop.processResponse("pop", "ad1", badRequestResponse)
	assert.Nil(t, err)
}
