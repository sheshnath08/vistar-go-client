package vistar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var PoPRequestTimeout = 30 * time.Second
var RetryInterval = 1 * time.Minute

type ProofOfPlayRequest struct {
	DisplayTime int64 `json:"display_time"`
}

type ProofOfPlay interface {
	Expire(Ad) error
	Confirm(Ad, int64) error
	Stop()
	GetStats() Stats
}

type PoPRequest struct {
	Ad          Ad
	Status      bool
	DisplayTime int64
	RequestTime time.Time
}

type PoPError struct {
	Status  int
	Message string
}

func (e *PoPError) Error() string {
	return fmt.Sprintf("%d: %s", e.Status, e.Message)
}

type testProofOfPlay struct {
	requests       []*PoPRequest
	retryQueue     []*PoPRequest
	bandwidthStats Stats
}

func NewTestProofOfPlay() *testProofOfPlay {
	return &testProofOfPlay{
		requests:       make([]*PoPRequest, 0, 0),
		bandwidthStats: Stats{},
	}
}

func (t *testProofOfPlay) Stop() {
}

func (t *testProofOfPlay) Confirm(ad Ad, displayTime int64) error {
	t.requests = append(
		t.requests,
		&PoPRequest{Ad: ad, Status: false, DisplayTime: displayTime})
	return nil
}

func (t *testProofOfPlay) Expire(ad Ad) error {
	t.requests = append(t.requests, &PoPRequest{Ad: ad, Status: true})
	return nil
}

func (t *testProofOfPlay) GetStats() Stats {
	return t.bandwidthStats
}

type proofOfPlay struct {
	eventFn        EventFunc
	httpClient     *http.Client
	requests       chan *PoPRequest
	retryQueue     chan *PoPRequest
	statsLock      sync.RWMutex
	bandwidthStats Stats
}

func NewProofOfPlay(eventFn EventFunc) *proofOfPlay {
	httpClient := &http.Client{Timeout: PoPRequestTimeout}
	pop := &proofOfPlay{
		eventFn:        eventFn,
		httpClient:     httpClient,
		requests:       make(chan *PoPRequest, 100),
		retryQueue:     make(chan *PoPRequest, 100),
		bandwidthStats: Stats{},
	}

	go pop.start()
	go pop.processRetries()
	return pop
}

func (p *proofOfPlay) Stop() {
	close(p.requests)
	close(p.retryQueue)
}

func (p *proofOfPlay) Expire(ad Ad) error {
	req := &PoPRequest{Ad: ad, Status: false}
	return p.expire(req)
}

func (p *proofOfPlay) Confirm(ad Ad, displayTime int64) error {
	req := &PoPRequest{Ad: ad, Status: true, DisplayTime: displayTime}
	return p.confirm(req)
}

func (p *proofOfPlay) GetStats() Stats {
	p.statsLock.Lock()
	defer p.statsLock.Unlock()

	return p.bandwidthStats
}

func (p *proofOfPlay) start() {
	for req := range p.requests {
		if req.Status {
			p.confirm(req)
		} else {
			p.expire(req)
		}
	}
}

func (p *proofOfPlay) processRequestFailure(req *PoPRequest, err error) {
	// Only raise the event on the first failed attempt
	if req.RequestTime.IsZero() {
		popType := p.getPoPType(req)
		p.publishEvent(
			fmt.Sprintf("ad-%s-failed", popType),
			fmt.Sprintf("adId: %s, error: %s", req.Ad["id"], err.Error()),
			"warning")
	}

	req.RequestTime = time.Now()
	p.retryQueue <- req
}

func (p proofOfPlay) publishEvent(name string, message string, level string) {
	if p.eventFn == nil {
		return
	}

	p.eventFn(name, message, "", level)
}

func (p *proofOfPlay) makePoPRequest(req *http.Request,
	popReq *PoPRequest, adId string, popType string) error {
	resp, err := p.httpClient.Do(req)
	if err != nil {
		// Connection error - retry the request
		p.processRequestFailure(popReq, err)
		return &PoPError{
			Status: http.StatusAccepted,
			Message: fmt.Sprintf(
				"Connection error, request will be retried: %s", err.Error()),
		}
	}

	defer resp.Body.Close()

	p.updateBandwidthStats(getRequestLength(req), getResponseLength(resp))

	code := resp.StatusCode
	// Response was OK: 1xx - 3xx
	if code >= http.StatusContinue && code < http.StatusBadRequest {
		return nil
	}

	// Bad request 4xx - We don't need to retry these
	if code >= http.StatusBadRequest && code < http.StatusInternalServerError {
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			p.publishEvent(
				fmt.Sprintf("ad-%s-failed", popType),
				fmt.Sprintf("adId: %s, error: %s", adId, body),
				"warning")
			err = &PoPError{
				Status:  code,
				Message: fmt.Sprintf("%s", body),
			}
		}

		return &PoPError{
			Status:  code,
			Message: err.Error(),
		}
	}

	// Ad server return 5xx - retry this request
	err = &PoPError{
		Status: http.StatusAccepted,
		Message: fmt.Sprintf(
			"Ad server responded %d: request will be retried.", code),
	}
	p.processRequestFailure(popReq, err)
	return err
}

func (p *proofOfPlay) confirm(popReq *PoPRequest) error {
	ad := popReq.Ad
	displayTime := popReq.DisplayTime
	confirmUrl, ok := ad["proof_of_play_url"].(string)
	if !ok {
		return &PoPError{
			Status:  http.StatusBadRequest,
			Message: "Invalid proof of play URL",
		}
	}

	adId, ok := ad["id"].(string)
	if !ok {
		return &PoPError{
			Status:  http.StatusBadRequest,
			Message: "Invalid ad id",
		}
	}

	data, err := json.Marshal(&ProofOfPlayRequest{DisplayTime: displayTime})
	if err != nil {
		return &PoPError{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	req, err := http.NewRequest("POST", confirmUrl, bytes.NewBuffer(data))
	if err != nil {
		return &PoPError{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Connection Error: %s", err.Error()),
		}
	}

	return p.makePoPRequest(req, popReq, adId, "pop")
}

func (p *proofOfPlay) expire(popReq *PoPRequest) error {
	ad := popReq.Ad
	expUrl, ok := ad["expiration_url"].(string)
	if !ok {
		return &PoPError{
			Status:  http.StatusBadRequest,
			Message: "Invalid expire url",
		}
	}

	adId, ok := ad["id"].(string)
	if !ok {
		return &PoPError{
			Status:  http.StatusBadRequest,
			Message: "Invalid ad id",
		}
	}

	req, err := http.NewRequest("GET", expUrl, nil)
	if err != nil {
		return &PoPError{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Connection Error: %s", err.Error()),
		}
	}

	return p.makePoPRequest(req, popReq, adId, "expire")
}

func (p *proofOfPlay) isLeaseExpired(ad Ad) bool {
	exp, ok := ad["lease_expiry"].(float64)
	if !ok {
		return true
	}

	expiry := time.Unix(int64(exp), 0)
	return time.Now().After(expiry)
}

func (p *proofOfPlay) processRetries() {
	for req := range p.retryQueue {
		p.processRetry(req)
	}
}

func (p *proofOfPlay) processRetry(req *PoPRequest) {
	ad := req.Ad

	if p.isLeaseExpired(ad) {
		popType := p.getPoPType(req)
		p.publishEvent(
			fmt.Sprintf("ad-%s-already-expired", popType),
			fmt.Sprintf("ad: %s, expiry: %d", ad["id"], ad["lease_expiry"]),
			"critical")
		return
	}

	sleepDuration := RetryInterval - time.Since(req.RequestTime)
	time.Sleep(sleepDuration)
	p.requests <- req
}

func (p proofOfPlay) getPoPType(req *PoPRequest) string {
	if req.Status {
		return "pop"
	}
	return "expire"
}

func (p *proofOfPlay) updateBandwidthStats(sentBytes int64,
	receivedBytes int64) {
	p.statsLock.Lock()
	defer p.statsLock.Unlock()

	updateStats(&p.bandwidthStats, sentBytes, receivedBytes)
}
