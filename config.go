package vistar

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var UserAgent = "VistarGoClient"

var DefaultMimeTypes = []string{
	"image/gif", "image/jpeg", "image/png", "video/mp4",
}

type Dimension struct {
	Width  int64
	Height int64
}

type DeviceParams map[string]string

type DisplayArea struct {
	AllowAudio     bool     `json:"allow_audio"`
	Height         int64    `json:"height"`
	Id             string   `json:"id"`
	MaxDuration    int64    `json:"max_duration,omitempty"`
	MinDuration    int64    `json:"min_duration,omitempty"`
	StaticDuration int64    `json:"static_duration,omitempty"`
	SupportedMedia []string `json:"supported_media"`
	Width          int64    `json:"width"`
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
	Duration           int64             `json:"duration,omitempty"`
	Interval           int64             `json:"interval,omitempty"`
}

type AdConfig interface {
	NewAdRequest() *AdRequest
	UpdateAdRequest(*AdRequest)
	ServerUrl() string
	AssetEndpointUrl() string
	AssetEndpointDisplayAreas() []DisplayArea
}

type adConfig struct {
	baseRequest               *AdRequest
	url                       string
	assetEndpointUrl          string
	assetEndpointDisplayAreas []DisplayArea
	logEnabled                bool
	logLevel                  int64
}

func NewAdConfig(params DeviceParams) *adConfig {
	c := &adConfig{}
	c.parse(params)
	return c
}

func (c *adConfig) NewAdRequest() *AdRequest {
	return c.baseRequest
}

func (c adConfig) LogEnabled() bool {
	return c.logEnabled
}

func (c adConfig) LogLevel() int64 {
	return c.logLevel
}

func (c adConfig) ServerUrl() string {
	return c.url
}

func (c adConfig) AssetEndpointUrl() string {
	return c.assetEndpointUrl
}

func (c adConfig) AssetEndpointDisplayAreas() []DisplayArea {
	return c.assetEndpointDisplayAreas
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
	req.Duration = c.baseRequest.Duration
	req.Interval = c.baseRequest.Interval
	req.DisplayTime = time.Now().Unix()

	if len(req.DisplayAreas) == 0 {
		req.DisplayAreas = c.baseRequest.DisplayAreas
	} else {
		for i := 0; i < len(req.DisplayAreas); i++ {
			if len(req.DisplayAreas[i].SupportedMedia) == 0 {
				req.DisplayAreas[i].SupportedMedia = make([]string,
					len(c.baseRequest.DisplayAreas[0].SupportedMedia))
				copy(req.DisplayAreas[i].SupportedMedia,
					c.baseRequest.DisplayAreas[0].SupportedMedia)
			}

			if req.DisplayAreas[i].Width <= 0 {
				req.DisplayAreas[i].Width =
					c.baseRequest.DisplayAreas[0].Width
			}

			if req.DisplayAreas[i].Height <= 0 {
				req.DisplayAreas[i].Height =
					c.baseRequest.DisplayAreas[0].Height
			}

			if !req.DisplayAreas[i].AllowAudio {
				req.DisplayAreas[i].AllowAudio =
					c.baseRequest.DisplayAreas[0].AllowAudio
			}
		}
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
	c.assetEndpointUrl = params["vistar.asset_cache_endpoint"]
	c.logEnabled = c.parseBool(params, "cortex.log_enabled", false)
	// 0 = debug, 1 = info, 2 = warn, 3 = error
	c.logLevel = c.parseInt(params, "cortex.log_level", 2)

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
	req.Duration = c.parseInt(params, "vistar.duration", 0)
	req.Interval = c.parseInt(params, "vistar.interval", 0)

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

	req.DisplayAreas[0].MinDuration = c.parseInt(params,
		"vistar.min_duration", 0)

	req.DisplayAreas[0].MaxDuration = c.parseInt(params,
		"vistar.max_duration", 0)

	req.DisplayAreas[0].StaticDuration = c.parseInt(params,
		"vistar.static_duration", 0)

	c.assetEndpointDisplayAreas = c.parseAssetEndpointDisplayAreas(params,
		req.DisplayAreas[0].AllowAudio, mimeTypes)

	c.baseRequest = &req
}

func (c adConfig) parseAssetEndpointDisplayAreas(params DeviceParams,
	allowAudio bool, mimeTypes []string) []DisplayArea {
	areas := make([]DisplayArea, 0, 0)
	sval, ok := params["vistar.creative_sizes"]
	if !ok {
		return areas
	}

	for idx, dimension := range c.parseDimensionString(sval) {
		areas = append(areas, DisplayArea{
			Id:             fmt.Sprintf("display-%d", idx),
			Width:          dimension.Width,
			Height:         dimension.Height,
			AllowAudio:     allowAudio,
			SupportedMedia: mimeTypes})
	}
	return areas
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

func (c adConfig) parseDimensionString(val string) []Dimension {
	res := make([]Dimension, 0, 0)
	for _, part := range strings.Split(val, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		widthAndHeight := strings.Split(part, "x")
		if len(widthAndHeight) != 2 {
			continue
		}

		w, err := strconv.ParseInt(strings.TrimSpace(widthAndHeight[0]), 10, 64)
		if err != nil {
			continue
		}

		h, err := strconv.ParseInt(strings.TrimSpace(widthAndHeight[1]), 10, 64)
		if err != nil {
			continue
		}

		res = append(res, Dimension{Width: w, Height: h})
	}

	return res
}
