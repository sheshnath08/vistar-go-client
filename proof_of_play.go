package vistar

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type ProofOfPlayRequest struct {
	DisplayTime int64 `json:"display_time"`
}

type ProofOfPlay interface {
	Expire(Ad) error
	Confirm(Ad, int64) error
}

type PoPRequest struct {
	Ad          Ad
	AdId        string
	Url         string
	Status      bool
	DisplayTime int64
}

type PoPError struct {
	Status  int
	Message string
}

func (e *PoPError) Error() string {
	return fmt.Sprintf("%d: %s", e.Status, e.Message)
}

type proofOfPlay struct {
	eventFn EventFunc
	popFunc PoPFunc
}

func NewProofOfPlay(eventFn EventFunc, popFunc PoPFunc) *proofOfPlay {
	pop := &proofOfPlay{
		eventFn: eventFn,
		popFunc: popFunc,
	}

	return pop
}

func (p *proofOfPlay) Expire(ad Ad) error {
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

	return p.expire(&PoPRequest{
		Ad:     ad,
		AdId:   adId,
		Status: false,
		Url:    expUrl,
	})
}

func (p *proofOfPlay) Confirm(ad Ad, displayTime int64) error {
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

	return p.confirm(&PoPRequest{
		Ad:          ad,
		AdId:        adId,
		Status:      true,
		DisplayTime: displayTime,
		Url:         confirmUrl,
	})
}

func (p *proofOfPlay) publishEvent(name string, message string, level string) {
	if p.eventFn == nil {
		return
	}

	p.eventFn(name, message, "", level)
}

func (p *proofOfPlay) confirm(popReq *PoPRequest) error {
	data := &ProofOfPlayRequest{DisplayTime: popReq.DisplayTime}

	resp, err := p.popFunc(http.MethodPost, popReq.Url, data)
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		body, readErr := ioutil.ReadAll(resp.Body)
		if readErr == nil {
			p.publishEvent("ad-pop-failed", fmt.Sprintf("adId: %s, error: %s",
				popReq.AdId, body), "warning")
		}
	}

	return err
}

func (p *proofOfPlay) expire(popReq *PoPRequest) error {
	resp, err := p.popFunc(http.MethodGet, popReq.Url, nil)
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		body, readErr := ioutil.ReadAll(resp.Body)
		if readErr == nil {
			p.publishEvent("ad-expire-failed",
				fmt.Sprintf("adId: %s, error: %s", popReq.AdId, body),
				"warning")
		}
	}

	return err
}

type testProofOfPlay struct {
	requests []*PoPRequest
}

func NewTestProofOfPlay() *testProofOfPlay {
	return &testProofOfPlay{
		requests: make([]*PoPRequest, 0, 0),
	}
}

func (t *testProofOfPlay) Confirm(ad Ad, displayTime int64) error {
	t.requests = append(
		t.requests,
		&PoPRequest{Ad: ad, Status: true, DisplayTime: displayTime})
	return nil
}

func (t *testProofOfPlay) Expire(ad Ad) error {
	t.requests = append(t.requests, &PoPRequest{Ad: ad, Status: false})
	return nil
}
