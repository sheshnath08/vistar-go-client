package vistar

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

var AdNotFound = errors.New("ad not found")

type CacheFunc func(string, time.Duration) (string, error)
type EventFunc func(string, string, string, string)

type Ad map[string]interface{}
type AdResponse struct {
	Advertisement []Ad `json:"advertisement,omitempty"`
}

type Client interface {
	GetAd(AdConfig, *AdRequest) (*AdResponse, error)
	Expire(string) error
	Confirm(string, int64) error
	GetInProgressAds() map[string]Ad
}

type client struct {
	httpClient    *http.Client
	pop           ProofOfPlay
	assetTTL      time.Duration
	cacheFn       CacheFunc
	eventFn       EventFunc
	lock          sync.RWMutex
	inProgressAds map[string]Ad
}

func NewClient(reqTimeout time.Duration, eventFn EventFunc, cacheFn CacheFunc,
	assetTTL time.Duration) *client {
	httpClient := &http.Client{Timeout: reqTimeout}
	pop := NewProofOfPlay(eventFn)
	return &client{
		pop:           pop,
		assetTTL:      assetTTL,
		httpClient:    httpClient,
		eventFn:       eventFn,
		cacheFn:       cacheFn,
		inProgressAds: make(map[string]Ad),
	}
}

func (c *client) GetInProgressAds() map[string]Ad {
	c.lock.RLock()
	defer c.lock.RUnlock()

	ret := map[string]Ad{}
	for k, v := range c.inProgressAds {
		ret[k] = v
	}
	return ret
}

func (c *client) Expire(adId string) error {
	ad, ok := c.removeFromInProgressList(adId)
	if !ok {
		return AdNotFound
	}

	c.pop.Expire(ad)
	return nil
}

func (c *client) Confirm(adId string, displayTime int64) error {
	ad, ok := c.removeFromInProgressList(adId)
	if !ok {
		return AdNotFound
	}

	c.pop.Confirm(ad, displayTime)
	return nil
}

func (c *client) GetAd(config AdConfig, req *AdRequest) (
	*AdResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	hreq, err := http.NewRequest("POST", config.ServerUrl(),
		bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	hreq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		c.publishEvent("ad-server-request-failed",
			fmt.Sprintf("code: %d, body: %s", resp.StatusCode, string(body)),
			"warning")
		return nil, fmt.Errorf("Ad server returned an error. code: %d, body: %s",
			resp.StatusCode, string(body))
	}

	adResponse := &AdResponse{}
	err = json.Unmarshal(body, adResponse)
	if err != nil {
		return nil, err
	}

	if len(adResponse.Advertisement) == 0 {
		c.publishEvent("ad-server-returned-no-ads", "", "warning")
		return adResponse, nil
	}

	if c.cacheFn == nil {
		return adResponse, nil
	}

	c.cacheAds(adResponse)
	cleanedResponse := c.tryToExpireAds(adResponse)
	return cleanedResponse, nil
}

func (c *client) cacheAds(resp *AdResponse) {
	if c.cacheFn == nil {
		return
	}
	var wg sync.WaitGroup
	for _, ad := range resp.Advertisement {
		originalUrl := ad["asset_url"].(string)
		wg.Add(1)
		go func(ad Ad) {
			defer wg.Done()
			local, err := c.cacheFn(originalUrl, c.assetTTL)
			if err != nil {
				log.Printf("Unable to cache asset %s, err: %v", originalUrl, err)
				c.publishEvent("app-cache-failed", originalUrl, "warning")
				ad["should_expire"] = true
				return
			}
			ad["original_asset_url"] = originalUrl
			ad["asset_url"] = local
			c.addToInProgressList(ad)
		}(ad)
	}
	wg.Wait()
}

func (c *client) tryToExpireAds(resp *AdResponse) *AdResponse {
	cleaned := &AdResponse{}
	for _, ad := range resp.Advertisement {
		shouldExpire, ok := ad["should_expire"].(bool)
		if ok && shouldExpire {
			c.pop.Expire(ad)
			continue
		}
		cleaned.Advertisement = append(cleaned.Advertisement, ad)
	}
	return cleaned
}

func (c *client) addToInProgressList(ad Ad) {
	c.lock.Lock()
	defer c.lock.Unlock()

	adId := ad["id"].(string)
	c.inProgressAds[adId] = ad
}

func (c *client) removeFromInProgressList(adId string) (Ad, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	ad, ok := c.inProgressAds[adId]
	delete(c.inProgressAds, adId)
	return ad, ok
}

func (c client) publishEvent(name string, message string, level string) {
	if c.eventFn == nil {
		return
	}
	c.eventFn(name, message, "", level)
}
