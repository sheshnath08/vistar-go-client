package vistar

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

type Asset map[string]interface{}
type AssetResponse struct {
	Assets []Asset `json:"asset,omitempty"`
}

type Client interface {
	GetAd(AdConfig, *AdRequest) (*AdResponse, error)
	Expire(string) error
	Confirm(string, int64) (string, error)
	GetInProgressAds() map[string]Ad
	GetAssets(AdConfig, *AdRequest) (*AssetResponse, error)
	Close()
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

func (c *client) Close() {
	c.pop.Stop()
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

func (c *client) Confirm(adId string, displayTime int64) (string, error) {
	ad, ok := c.removeFromInProgressList(adId)
	if !ok {
		return "", AdNotFound
	}

	c.pop.Confirm(ad, displayTime)
	return ad["original_asset_url"].(string), nil
}

func (c *client) GetAd(config AdConfig, req *AdRequest) (
	*AdResponse, error) {
	body, err := c.post(config.ServerUrl(), config, req)
	if err != nil {
		return nil, err
	}

	resp := &AdResponse{}
	err = json.Unmarshal(body, resp)
	if err != nil {
		return nil, err
	}

	if len(resp.Advertisement) == 0 {
		c.publishEvent("ad-server-returned-no-ads", "", "warning")
		return resp, nil
	}

	if c.cacheFn == nil {
		return resp, nil
	}

	c.cacheAds(resp)
	cleanedResponse := c.tryToExpireAds(resp)
	return cleanedResponse, nil
}

func (c *client) GetAssets(config AdConfig, req *AdRequest) (
	*AssetResponse, error) {
	body, err := c.post(config.AssetEndpointUrl(), config, req)
	if err != nil {
		return nil, err
	}

	resp := &AssetResponse{}
	err = json.Unmarshal(body, resp)
	if err != nil {
		return nil, err
	}

	if len(resp.Assets) == 0 {
		c.publishEvent("ad-server-returned-no-assets", "", "warning")
	}

	return resp, nil
}

func (c *client) post(url string, config AdConfig, req *AdRequest) (
	[]byte, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	hreq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
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
		c.publishEvent(
			"ad-server-endpoint-failed",
			fmt.Sprintf("url: %s, code: %d, body: %s", url, resp.StatusCode,
				string(body)),
			"warning")
		return nil, fmt.Errorf("Ad server returned an error. url: %s, "+
			"code: %d, body: %s", url, resp.StatusCode, string(body))
	}
	return body, err
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
				c.publishEvent("app-cache-failed",
					fmt.Sprintf("url: %s, error: %s", originalUrl, err.Error()),
					"warning")
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
