package vistar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
