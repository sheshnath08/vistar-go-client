package vistar

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type cacheCall struct {
	url string
	ech chan error
	och chan string
}

type eventCall struct {
	name    string
	message string
	source  string
	level   string
}

func TestCacheAdsNoCacher(t *testing.T) {
	resp := &AdResponse{
		Advertisement: []Ad{
			map[string]interface{}{"asset_url": "url1"},
			map[string]interface{}{"asset_url": "url2"},
		},
	}

	client := NewClient(time.Second*1, nil, nil, time.Minute*1)

	client.cacheAds(resp)

	assert.Len(t, client.inProgressAds, 0)
	assert.Len(t, resp.Advertisement, 2)
	assert.Equal(t, resp.Advertisement[0]["asset_url"], "url1")
	assert.Equal(t, resp.Advertisement[1]["asset_url"], "url2")
}

func TestCacheAds(t *testing.T) {
	resp := &AdResponse{
		Advertisement: []Ad{
			map[string]interface{}{"id": "1", "asset_url": "url1"},
			map[string]interface{}{"id": "2", "asset_url": "url2"},
		},
	}

	cacheCalls := make(chan *cacheCall, 100)
	cacheFn := func(url string, ttl time.Duration) (string, error) {
		ech := make(chan error)
		och := make(chan string)

		cacheCalls <- &cacheCall{
			ech: ech,
			och: och,
			url: url}

		for {
			select {
			case err := <-ech:
				return "", err
			case p := <-och:
				return p, nil
			}
		}
	}

	eventCalls := make([]*eventCall, 0, 0)
	eventFn := func(name string, message string, source string, level string) {
		eventCalls = append(eventCalls, &eventCall{
			name:    name,
			message: message,
			source:  source,
			level:   level})
	}

	client := NewClient(time.Second*1, eventFn, cacheFn, time.Second*1)

	cacheError := errors.New("cache failed")
	cacheEntry := "/cached-url"
	done := make(chan bool)
	go func() {
		client.cacheAds(resp)
		assert.Len(t, resp.Advertisement, 2)
		assert.Equal(t, resp.Advertisement[0]["asset_url"], "url1")
		assert.Equal(t, resp.Advertisement[0]["should_expire"], true)

		assert.Equal(t, resp.Advertisement[1]["asset_url"], "/cached-url")
		assert.Equal(t, resp.Advertisement[1]["original_asset_url"], "url2")
		_, ok := resp.Advertisement[1]["should_expire"]
		assert.False(t, ok)

		assert.Len(t, client.inProgressAds, 1)
		assert.Equal(t, client.inProgressAds["2"]["asset_url"], "/cached-url")

		assert.Len(t, eventCalls, 1)
		assert.Equal(t, eventCalls[0].name, "app-cache-failed")
		done <- true
	}()

	go func() {
		call := <-cacheCalls
		// Fail to cache url1
		if call.url == "url1" {
			call.ech <- cacheError
		} else {
			call.och <- cacheEntry
		}
		// Cache url2 as /cached-url
		call = <-cacheCalls
		if call.url == "url1" {
			call.ech <- cacheError
		} else {
			call.och <- cacheEntry
		}
	}()

	<-done
}

func TestTryToExpireAdsAllValid(t *testing.T) {
	resp := &AdResponse{
		Advertisement: []Ad{
			map[string]interface{}{"id": "1", "asset_url": "url1"},
			map[string]interface{}{"id": "2", "asset_url": "url2"},
		},
	}

	pop := NewTestProofOfPlay()
	client := &client{pop: pop}

	nresp := client.tryToExpireAds(resp)
	assert.Len(t, nresp.Advertisement, 2)
	assert.Equal(t, nresp.Advertisement[0]["id"], "1")
	assert.Equal(t, nresp.Advertisement[1]["id"], "2")

	assert.Len(t, pop.requests, 0)
}

func TestTryToExpireAds(t *testing.T) {
	resp := &AdResponse{
		Advertisement: []Ad{
			map[string]interface{}{"id": "1", "asset_url": "url1",
				"should_expire": true},
			map[string]interface{}{"id": "2", "asset_url": "url2",
				"should_expire": false},
		},
	}

	pop := NewTestProofOfPlay()
	client := &client{pop: pop}

	nresp := client.tryToExpireAds(resp)
	assert.Len(t, nresp.Advertisement, 1)
	assert.Equal(t, nresp.Advertisement[0]["id"], "2")

	assert.Len(t, pop.requests, 1)
	assert.Equal(t, pop.requests[0].Ad["asset_url"], "url1")
}

func TestHidePopUrl(t *testing.T)  {
	resp := &AdResponse{
		Advertisement: []Ad{
			map[string]interface{}{"id": "1", "asset_url": "url1",
				"proof_of_play_url": "http://pop-url"},
			map[string]interface{}{"id": "2", "asset_url": "url2"},
		},
	}

	pop := NewTestProofOfPlay()
	client := &client{pop: pop}

	nresp := client.hidePopUrl(resp)
	assert.Len(t, nresp.Advertisement, 2)
	assert.Equal(t, nresp.Advertisement[0]["id"], "1")
	assert.NotContains(t, nresp.Advertisement[0], "proof_of_play_url")
	assert.Equal(t, nresp.Advertisement[1]["id"], "2")
	assert.NotContains(t, nresp.Advertisement[1], "proof_of_play_url")
}

func TestUpdateBandwidthStats(t *testing.T) {
	pop := NewTestProofOfPlay()
	client := &client{
		pop:            pop,
		bandwidthStats: make(map[string]Stats),
	}

	client.updateBandwidthStats("/test", int64(100), int64(1024))

	stats, ok := client.bandwidthStats["/test"]
	assert.True(t, ok)

	assert.Equal(t, stats.Count, int64(1))
	assert.Equal(t, stats.BytesSent, int64(100))
	assert.Equal(t, stats.BytesReceived, int64(1024))
	assert.Equal(t, stats.Total, int64(1124))
	assert.Equal(t, stats.Average, float64(1124))

	client.updateBandwidthStats("/test", int64(100), int64(1024))

	stats, ok = client.bandwidthStats["/test"]

	assert.True(t, ok)
	assert.Equal(t, stats.Count, int64(2))
	assert.Equal(t, stats.BytesSent, int64(200))
	assert.Equal(t, stats.BytesReceived, int64(2048))
	assert.Equal(t, stats.Total, int64(2248))
	assert.Equal(t, stats.Average, float64(1124))
}
