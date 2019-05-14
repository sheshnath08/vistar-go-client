package vistar

import (
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

	defer func() {
		pop.Stop()
	}()

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
	threshold := 10 * time.Millisecond
	assert.True(t, since >= diff)
	assert.True(t, since <= (diff+threshold))
}
