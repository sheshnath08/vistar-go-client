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

func publishEventStub(events chan *testEvent) EventFunc {
	return func(name string, msg string, src string, level string) {
		event := testEvent{
			name:    name,
			message: msg,
			source:  src,
			level:   level,
		}
		events <- &event
	}
}

func TestRetryPoPs(t *testing.T) {
	pop := NewProofOfPlay(nil)

	exp := float64(time.Now().Add(-1 * time.Hour).Unix())
	unexp := float64(time.Now().Add(1 * time.Hour).Unix())
	expiredCount := 0
	unexpiredCount := 0

	for i := range [21]int{} {
		ad := Ad{"id": i}
		status := false
		if (i % 2) == 0 {
			ad["lease_expiry"] = exp
			status = true
			expiredCount = expiredCount + 1
		} else {
			ad["lease_expiry"] = unexp
			unexpiredCount = unexpiredCount + 1
		}
		req := &PoPRequest{Ad: ad, Status: status}
		pop.retryQueue <- req
	}

	pop.retryFailedPoPs()

	defer func() {
		counter := 0
		_, more := <-pop.retryQueue
		if more {
			counter = counter + 1
		} else {
			pop.Stop()
			assert.Equal(t, counter, unexpiredCount)
		}
	}()
}
