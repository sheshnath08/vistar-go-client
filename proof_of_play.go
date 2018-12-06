package vistar

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	try "gopkg.in/matryer/try.v1"
)

var PoPRequestTimeout = 2000 * time.Millisecond
var PoPRetryDelaySecs = 10 * time.Second
var PoPNumRetries = 3

type ProofOfPlayRequest struct {
	DisplayTime int64 `json:"display_time"`
}

type ProofOfPlay interface {
	Expire(Ad)
	Confirm(Ad, int64)
}

type PoPRequest struct {
	Ad          Ad
	Status      bool
	DisplayTime int64
}

type testProofOfPlay struct {
	requests []*PoPRequest
}

func NewTestProofOfPlay() *testProofOfPlay {
	return &testProofOfPlay{requests: make([]*PoPRequest, 0, 0)}
}

func (t *testProofOfPlay) Expire(ad Ad) {
	t.requests = append(t.requests, &PoPRequest{Ad: ad, Status: false})
}

func (t *testProofOfPlay) Confirm(ad Ad, displayTime int64) {
	t.requests = append(t.requests, &PoPRequest{
		Ad:          ad,
		Status:      true,
		DisplayTime: displayTime})
}

type proofOfPlay struct {
	httpClient *http.Client
	requests   chan *PoPRequest
	eventFn    EventFunc
}

func NewProofOfPlay(eventFn EventFunc) *proofOfPlay {
	httpClient := &http.Client{Timeout: PoPRequestTimeout}
	pop := &proofOfPlay{
		eventFn:    eventFn,
		requests:   make(chan *PoPRequest, 100),
		httpClient: httpClient,
	}
	go pop.start()
	return pop
}

func (p *proofOfPlay) Expire(ad Ad) {
	p.requests <- &PoPRequest{Ad: ad, Status: false}
}

func (p *proofOfPlay) Confirm(ad Ad, displayTime int64) {
	p.requests <- &PoPRequest{Ad: ad, Status: true, DisplayTime: displayTime}
}

func (p *proofOfPlay) start() {
	for req := range p.requests {
		if req.Status {
			err := p.confirm(req.Ad, req.DisplayTime)
			if err != nil {
				p.publishEvent("ad-pop-failed",
					fmt.Sprintf("adId: %s, error: %s", req.Ad["id"], err.Error()),
					"critical")
			}
		} else {
			err := p.expire(req.Ad)
			if err != nil {
				p.publishEvent("ad-expire-failed",
					fmt.Sprintf("adId: %s, error: %s", req.Ad["id"], err.Error()),
					"critical")
			}
		}
	}
}

func (p proofOfPlay) publishEvent(name string, message string, level string) {
	if p.eventFn == nil {
		return
	}

	p.eventFn(name, message, "", level)
}

func (p *proofOfPlay) confirm(ad Ad, displayTime int64) error {
	confirmUrl, ok := ad["proof_of_play_url"].(string)
	if !ok {
		return errors.New("Invalid proof of play url")
	}

	data, err := json.Marshal(&ProofOfPlayRequest{DisplayTime: displayTime})
	if err != nil {
		return err
	}

	return try.Do(func(attempt int) (bool, error) {
		req, err := http.NewRequest("POST", confirmUrl, bytes.NewBuffer(data))
		if err != nil {
			return false, nil
		}

		resp, err := p.httpClient.Do(req)
		if err != nil {
			time.Sleep(PoPRetryDelaySecs)
			return attempt < PoPNumRetries, err
		}
		defer resp.Body.Close()
		return false, nil
	})
}

func (p *proofOfPlay) expire(ad Ad) error {
	expUrl, ok := ad["expiration_url"].(string)
	if !ok {
		return errors.New("Invalid expire url")
	}

	return try.Do(func(attempt int) (bool, error) {
		req, err := http.NewRequest("GET", expUrl, nil)
		if err != nil {
			return false, nil
		}

		resp, err := p.httpClient.Do(req)
		if err != nil {
			time.Sleep(PoPRetryDelaySecs)
			return attempt < PoPNumRetries, err
		}
		defer resp.Body.Close()
		return false, nil
	})
}
