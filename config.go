package vistar

import (
	"strconv"
	"strings"
	"time"
)

var UserAgent = "VistarGoClient"

var DefaultMimeTypes = []string{
	"image/gif", "image/jpeg", "image/png", "video/mp4",
}

type DeviceParams map[string]string

type DisplayArea struct {
	Id             string   `json:"id"`
	Width          int64    `json:"width"`
	Height         int64    `json:"height"`
	AllowAudio     bool     `json:"allow_audio"`
	SupportedMedia []string `json:"supported_media"`
	StaticDuration int64    `json:"static_duration,omitempty"`
}

type DeviceAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AdRequest struct {
	ApiKey             string            `json:"api_key"`
	NetworkId          string            `json:"network_id"`
	DeviceId           string            `json:"device_id"`
	VenueId            string            `json:"venue_id"`
	RequiredCompletion float64           `json:"required_completion"`
	DirectConnection   bool              `json:"direct_connection"`
	Latitude           float64           `json:"latitude,omitempty"`
	Longitude          float64           `json:"longitude,omitempty"`
	DisplayTime        int64             `json:"display_time"`
	NumberOfScreens    int64             `json:"number_of_screens"`
	DisplayAreas       []DisplayArea     `json:"display_area"`
	DeviceAttributes   []DeviceAttribute `json:"device_attribute"`
}

type AdConfig interface {
	NewAdRequest() *AdRequest
	UpdateAdRequest(*AdRequest)
	ServerUrl() string
}

type adConfig struct {
	baseRequest *AdRequest
	url         string
}

func NewAdConfig(params DeviceParams) *adConfig {
	c := &adConfig{}
	c.parse(params)
	return c
}

func (c *adConfig) NewAdRequest() *AdRequest {
	return c.baseRequest
}

func (c adConfig) ServerUrl() string {
	return c.url
}

func (c *adConfig) UpdateAdRequest(req *AdRequest) {
	req.ApiKey = c.baseRequest.ApiKey
	req.NetworkId = c.baseRequest.NetworkId
	req.DeviceId = c.baseRequest.DeviceId
	req.VenueId = c.baseRequest.VenueId
	req.DirectConnection = c.baseRequest.DirectConnection
	req.Latitude = c.baseRequest.Latitude
	req.Longitude = c.baseRequest.Longitude
	req.NumberOfScreens = c.baseRequest.NumberOfScreens
	req.RequiredCompletion = c.baseRequest.RequiredCompletion
	req.DisplayTime = time.Now().Unix()

	if len(req.DisplayAreas) == 0 {
		req.DisplayAreas = c.baseRequest.DisplayAreas
	}

	attrs := make(map[string]string)
	for _, attr := range c.baseRequest.DeviceAttributes {
		attrs[attr.Name] = attr.Value
	}

	for _, attr := range req.DeviceAttributes {
		attrs[attr.Name] = attr.Value
	}

	attrArr := make([]DeviceAttribute, 0, 0)
	for k, v := range attrs {
		attrArr = append(attrArr, DeviceAttribute{Name: k, Value: v})
	}

	req.DeviceAttributes = attrArr
}

func (c *adConfig) parse(params DeviceParams) {
	c.url = params["vistar.url"]

	req := AdRequest{}
	req.ApiKey = params["vistar.api_key"]
	req.NetworkId = params["vistar.network_id"]
	req.DeviceId = params["vistar.venue_id"]
	req.VenueId = params["vistar.venue_id"]
	req.DirectConnection = c.parseBool(params, "vistar.direct_connection", false)
	req.Latitude = c.parseFloat(params, "vistar.latitude", 0)
	req.Longitude = c.parseFloat(params, "vistar.longitude", 0)
	req.RequiredCompletion = c.parseFloat(params, "vistar.required_completion", 1)
	req.NumberOfScreens = 1

	mimeTypes := c.parseArray(params, "vistar.mime_types", DefaultMimeTypes)
	req.DeviceAttributes = []DeviceAttribute{
		DeviceAttribute{Name: "UserAgent", Value: UserAgent},
		DeviceAttribute{Name: "MimeTypes", Value: strings.Join(mimeTypes, ",")},
	}

	req.DisplayAreas = []DisplayArea{
		DisplayArea{
			Id:             "display-0",
			Width:          c.parseInt(params, "vistar.width", 0),
			Height:         c.parseInt(params, "vistar.height", 0),
			AllowAudio:     c.parseBool(params, "vistar.allow_audio", false),
			SupportedMedia: mimeTypes,
		},
	}

	req.DisplayAreas[0].StaticDuration = c.parseInt(params,
		"vistar.static_duration", 0)

	c.baseRequest = &req
}

func (c adConfig) parseArray(params DeviceParams, name string,
	def []string) []string {
	sval, ok := params[name]
	if !ok {
		return def
	}

	res := make([]string, 0, 0)
	for _, part := range strings.Split(sval, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			res = append(res, trimmed)
		}
	}

	if len(res) > 0 {
		return res
	}

	return def
}

func (c adConfig) parseInt(params DeviceParams, name string, def int64) int64 {
	sval, ok := params[name]
	if !ok {
		return def
	}

	val, err := strconv.ParseInt(sval, 10, 64)
	if err != nil {
		return def
	}

	return val
}

func (c adConfig) parseFloat(params DeviceParams, name string,
	def float64) float64 {
	sval, ok := params[name]
	if !ok {
		return def
	}

	val, err := strconv.ParseFloat(sval, 64)
	if err != nil {
		return def
	}

	return val
}

func (c adConfig) parseBool(params DeviceParams, name string, def bool) bool {
	sval, ok := params[name]
	if !ok {
		return def
	}

	val, err := strconv.ParseBool(sval)
	if err != nil {
		return def
	}

	return val
}
