package vistar

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	try "gopkg.in/matryer/try.v1"
)

var PoPRequestTimeout = 30 * time.Second
var PoPRetryDelaySecs = 10 * time.Second
var PoPNumRetries = 3
var RetryInterval = 1 * time.Minute

type ProofOfPlayRequest struct {
	DisplayTime int64 `json:"display_time"`
}

type ProofOfPlay interface {
	Expire(Ad) error
	Confirm(Ad, int64) error
	Stop()
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
	AdId    string
}

func (e *PoPError) Error() string {
	return fmt.Sprintf("%d: %s (%s)", e.Status, e.Message, e.AdId)
}

type testProofOfPlay struct {
	requests   []*PoPRequest
	retryQueue []*PoPRequest
}

func NewTestProofOfPlay() *testProofOfPlay {
	return &testProofOfPlay{requests: make([]*PoPRequest, 0, 0)}
}

func (t *testProofOfPlay) Expire(ad Ad) {
	t.requests = append(t.requests, &PoPRequest{Ad: ad, Status: false})
}

func (t *testProofOfPlay) Stop() {
}

func (t *testProofOfPlay) Confirm(ad Ad, displayTime int64) {
	t.requests = append(t.requests, &PoPRequest{
		Ad:          ad,
		Status:      true,
		DisplayTime: displayTime})
}

type proofOfPlay struct {
	eventFn    EventFunc
	httpClient *http.Client
	requests   chan *PoPRequest
	retryQueue chan *PoPRequest
}

func NewProofOfPlay(eventFn EventFunc) *proofOfPlay {
	httpClient := &http.Client{Timeout: PoPRequestTimeout}
	pop := &proofOfPlay{
		eventFn:    eventFn,
		httpClient: httpClient,
		requests:   make(chan *PoPRequest, 100),
		retryQueue: make(chan *PoPRequest, 100),
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
	err := p.expire(req)
	return err
}

func (p *proofOfPlay) Confirm(ad Ad, displayTime int64) error {
	req := &PoPRequest{Ad: ad, Status: true, DisplayTime: displayTime}
	err := p.confirm(req)
	return err
}

func (p *proofOfPlay) start() {
	for req := range p.requests {
		if req.Status {
			err := p.confirm(req)
			if err != nil {
				p.processRequestFailure(req, err)
			}
		} else {
			err := p.expire(req)
			if err != nil {
				p.processRequestFailure(req, err)
			}
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

func (p *proofOfPlay) processResponse(popType string, adId string,
	resp *http.Response) (bool, error) {
	code := resp.StatusCode

	// Response was OK: 1xx - 3xx
	if code >= http.StatusContinue && code < http.StatusBadRequest {
		return false, nil
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
				AdId:    adId,
				Message: fmt.Sprintf("%s", body),
				Status:  code,
			}
		}
		return false, err
	}

	// Server error 5xx - we should retry
	return true, &PoPError{AdId: adId, Status: code, Message: "Ad server error"}
}

func (p *proofOfPlay) confirm(popReq *PoPRequest) error {
	ad := popReq.Ad
	displayTime := popReq.DisplayTime
	confirmUrl, ok := ad["proof_of_play_url"].(string)
	if !ok {
		return errors.New("Invalid proof of play url")
	}

	adId, ok := ad["id"].(string)
	if !ok {
		return errors.New("Invalid ad id")
	}

	data, err := json.Marshal(&ProofOfPlayRequest{DisplayTime: displayTime})
	if err != nil {
		return err
	}

	err = try.Do(func(attempt int) (bool, error) {
		req, err := http.NewRequest("POST", confirmUrl, bytes.NewBuffer(data))
		if err != nil {
			return false, nil
		}

		resp, err := p.httpClient.Do(req)
		if err != nil {
			time.Sleep(PoPRetryDelaySecs)
			if attempt < PoPNumRetries {
				time.Sleep(PoPRetryDelaySecs)
				return true, err
			}

			p.processRequestFailure(popReq, err)
			return false, nil
		}

		shouldRetry, err := p.processResponse("pop", adId, resp)
		if shouldRetry {
			if attempt < PoPNumRetries {
				time.Sleep(PoPRetryDelaySecs)
				return true, err
			}

			p.processRequestFailure(popReq, err)
			return false, nil
		}

		defer resp.Body.Close()
		return false, err
	})

	return err
}

func (p *proofOfPlay) expire(popReq *PoPRequest) error {
	ad := popReq.Ad
	expUrl, ok := ad["expiration_url"].(string)
	if !ok {
		return errors.New("Invalid expire url")
	}

	adId, ok := ad["id"].(string)
	if !ok {
		return errors.New("Invalid ad id")
	}

	err := try.Do(func(attempt int) (bool, error) {
		req, err := http.NewRequest("GET", expUrl, nil)
		if err != nil {
			return false, nil
		}

		resp, err := p.httpClient.Do(req)
		if err != nil {
			time.Sleep(PoPRetryDelaySecs)
			if attempt < PoPNumRetries {
				time.Sleep(PoPRetryDelaySecs)
				return true, err
			}

			p.processRequestFailure(popReq, err)
			return false, nil
		}

		shouldRetry, err := p.processResponse("expire", adId, resp)
		if shouldRetry {
			if attempt < PoPNumRetries {
				time.Sleep(PoPRetryDelaySecs)
				return true, err
			}

			p.processRequestFailure(popReq, err)
			return false, nil
		}

		defer resp.Body.Close()
		return false, err
	})

	return err
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
		p.retryPoP(req)
	}
}

func (p *proofOfPlay) retryPoP(req *PoPRequest) {
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
