package vistar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArray(t *testing.T) {
	params := map[string]string{
		"p1": "v1",
		"p2": "   First,  Second, Third Value , ,, Fourth  ",
		"p3": ",,, ,, ,   ,,, ",
	}

	conf := &adConfig{}
	def := []string{"a", "b"}
	res := conf.parseArray(params, "unknown", def)
	assert.Equal(t, res, def)

	res = conf.parseArray(params, "p2", def)
	assert.Equal(t, res, []string{"First", "Second", "Third Value", "Fourth"})

	res = conf.parseArray(params, "p3", def)
	assert.Equal(t, res, def)
}

func TestParseInt(t *testing.T) {
	params := map[string]string{
		"p1": "0",
		"p2": "100",
		"p3": "not-number",
	}

	conf := &adConfig{}
	res := conf.parseInt(params, "unknown", 6)
	assert.Equal(t, res, int64(6))

	res = conf.parseInt(params, "p1", 6)
	assert.Equal(t, res, int64(0))

	res = conf.parseInt(params, "p2", 6)
	assert.Equal(t, res, int64(100))

	res = conf.parseInt(params, "p3", 6)
	assert.Equal(t, res, int64(6))
}

func TestParseFloat(t *testing.T) {
	params := map[string]string{
		"p1": "0.0",
		"p2": "100.9",
		"p3": "not-number",
	}

	conf := &adConfig{}
	res := conf.parseFloat(params, "unknown", 6.6)
	assert.Equal(t, res, 6.6)

	res = conf.parseFloat(params, "p1", 6.6)
	assert.Equal(t, res, float64(0))

	res = conf.parseFloat(params, "p2", 6.6)
	assert.Equal(t, res, 100.9)

	res = conf.parseFloat(params, "p3", 6.6)
	assert.Equal(t, res, 6.6)
}

func TestParseBool(t *testing.T) {
	params := map[string]string{
		"p1": "false",
		"p2": "true",
		"p3": "not-number",
	}

	conf := &adConfig{}
	res := conf.parseBool(params, "unknown", true)
	assert.True(t, res)

	res = conf.parseBool(params, "p1", true)
	assert.False(t, res)

	res = conf.parseBool(params, "p2", false)
	assert.True(t, res)

	res = conf.parseBool(params, "p3", false)
	assert.False(t, res)
}

func TestParse(t *testing.T) {
	params := map[string]string{
		"vistar.url":                 "staging-url",
		"vistar.api_key":             "api-key",
		"vistar.network_id":          "network-id",
		"vistar.venue_id":            "venue-id",
		"vistar.direct_connection":   "true",
		"vistar.latitude":            "45.5",
		"vistar.longitude":           "44.4",
		"vistar.mime_types":          "a,b,c",
		"vistar.width":               "100",
		"vistar.height":              "200",
		"vistar.allow_audio":         "true",
		"vistar.static_duration":     "9",
		"vistar.max_duration":        "10",
		"vistar.required_completion": "9",
	}

	conf := &adConfig{}
	conf.parse(params)

	assert.Equal(t, conf.url, "staging-url")
	assert.Equal(t, conf.baseRequest.ApiKey, "api-key")
	assert.Equal(t, conf.baseRequest.NetworkId, "network-id")
	assert.Equal(t, conf.baseRequest.DeviceId, "venue-id")
	assert.Equal(t, conf.baseRequest.VenueId, "venue-id")
	assert.True(t, conf.baseRequest.DirectConnection)
	assert.Equal(t, conf.baseRequest.RequiredCompletion, 9.0)
	assert.Equal(t, conf.baseRequest.Latitude, 45.5)
	assert.Equal(t, conf.baseRequest.Longitude, 44.4)
	assert.Equal(t, conf.baseRequest.DisplayTime, int64(0))
	assert.Equal(t, conf.baseRequest.NumberOfScreens, int64(1))
	assert.Equal(t, conf.baseRequest.NumberOfScreens, int64(1))
	assert.Len(t, conf.baseRequest.DisplayAreas, 1)
	assert.Len(t, conf.baseRequest.DeviceAttributes, 2)
	assert.Equal(t, conf.baseRequest.Duration, int64(0))
	assert.Equal(t, conf.baseRequest.Interval, int64(0))

	assert.Equal(t, conf.baseRequest.DisplayAreas[0].Width, int64(100))
	assert.Equal(t, conf.baseRequest.DisplayAreas[0].Height, int64(200))
	assert.True(t, conf.baseRequest.DisplayAreas[0].AllowAudio)
	assert.Equal(t, conf.baseRequest.DisplayAreas[0].SupportedMedia,
		[]string{"a", "b", "c"})
	assert.Equal(t, conf.baseRequest.DisplayAreas[0].StaticDuration, int64(9))
	assert.Equal(t, conf.baseRequest.DisplayAreas[0].MaxDuration, int64(10))
	assert.Equal(t, conf.baseRequest.DisplayAreas[0].MinDuration, int64(0))
}

func TestParseBulk(t *testing.T) {
	params := map[string]string{
		"vistar.url":                 "staging-url",
		"vistar.api_key":             "api-key",
		"vistar.network_id":          "network-id",
		"vistar.venue_id":            "venue-id",
		"vistar.direct_connection":   "true",
		"vistar.latitude":            "45.5",
		"vistar.longitude":           "44.4",
		"vistar.mime_types":          "a,b,c",
		"vistar.width":               "100",
		"vistar.height":              "200",
		"vistar.allow_audio":         "true",
		"vistar.static_duration":     "9",
		"vistar.required_completion": "9",
		"vistar.duration":            "300",
		"vistar.interval":            "60",
	}

	conf := &adConfig{}
	conf.parse(params)

	assert.Equal(t, conf.url, "staging-url")
	assert.Equal(t, conf.baseRequest.ApiKey, "api-key")
	assert.Equal(t, conf.baseRequest.NetworkId, "network-id")
	assert.Equal(t, conf.baseRequest.DeviceId, "venue-id")
	assert.Equal(t, conf.baseRequest.VenueId, "venue-id")
	assert.True(t, conf.baseRequest.DirectConnection)
	assert.Equal(t, conf.baseRequest.RequiredCompletion, 9.0)
	assert.Equal(t, conf.baseRequest.Latitude, 45.5)
	assert.Equal(t, conf.baseRequest.Longitude, 44.4)
	assert.Equal(t, conf.baseRequest.DisplayTime, int64(0))
	assert.Equal(t, conf.baseRequest.NumberOfScreens, int64(1))
	assert.Equal(t, conf.baseRequest.NumberOfScreens, int64(1))
	assert.Len(t, conf.baseRequest.DisplayAreas, 1)
	assert.Len(t, conf.baseRequest.DeviceAttributes, 2)
	assert.Equal(t, conf.baseRequest.Duration, int64(300))
	assert.Equal(t, conf.baseRequest.Interval, int64(60))

	assert.Equal(t, conf.baseRequest.DisplayAreas[0].Width, int64(100))
	assert.Equal(t, conf.baseRequest.DisplayAreas[0].Height, int64(200))
	assert.True(t, conf.baseRequest.DisplayAreas[0].AllowAudio)
	assert.Equal(t, conf.baseRequest.DisplayAreas[0].SupportedMedia,
		[]string{"a", "b", "c"})
	assert.Equal(t, conf.baseRequest.DisplayAreas[0].StaticDuration, int64(9))
}

func TestUpdateAdRequest(t *testing.T) {
	params := map[string]string{
		"vistar.url":                 "staging-url",
		"vistar.api_key":             "api-key",
		"vistar.network_id":          "network-id",
		"vistar.venue_id":            "venue-id",
		"vistar.direct_connection":   "true",
		"vistar.latitude":            "45.5",
		"vistar.longitude":           "44.4",
		"vistar.mime_types":          "a,b,c",
		"vistar.width":               "100",
		"vistar.height":              "200",
		"vistar.allow_audio":         "false",
		"vistar.static_duration":     "9",
		"vistar.max_duration":        "10",
		"vistar.required_completion": "9",
		"vistar.duration":            "300",
		"vistar.interval":            "60",
	}

	conf := &adConfig{}
	conf.parse(params)

	req := &AdRequest{
		DisplayAreas: []DisplayArea{
			DisplayArea{Id: "d1", Width: 500, Height: 500, AllowAudio: true,
				SupportedMedia: []string{"image"}},
		},
		DeviceAttributes: []DeviceAttribute{
			DeviceAttribute{Name: "attr1", Value: "value1"},
			DeviceAttribute{Name: "attr2", Value: "value2"},
		},
	}

	conf.UpdateAdRequest(req)

	assert.Equal(t, req.ApiKey, "api-key")
	assert.Equal(t, req.NetworkId, "network-id")
	assert.Equal(t, req.DeviceId, "venue-id")
	assert.Equal(t, req.VenueId, "venue-id")
	assert.True(t, req.DirectConnection)
	assert.Equal(t, req.Latitude, 45.5)
	assert.Equal(t, req.Longitude, 44.4)
	assert.NotEqual(t, req.DisplayTime, int64(0))
	assert.Equal(t, req.RequiredCompletion, 9.0)
	assert.Equal(t, req.NumberOfScreens, int64(1))
	assert.Len(t, req.DisplayAreas, 1)
	assert.Len(t, req.DeviceAttributes, 4)
	assert.Equal(t, req.Duration, int64(300))
	assert.Equal(t, req.Interval, int64(60))

	assert.Equal(t, req.DisplayAreas[0].Width, int64(500))
	assert.Equal(t, req.DisplayAreas[0].Height, int64(500))
	assert.True(t, req.DisplayAreas[0].AllowAudio)
	assert.Equal(t, req.DisplayAreas[0].SupportedMedia, []string{"image"})
	assert.Equal(t, req.DisplayAreas[0].StaticDuration, int64(9))
	assert.Equal(t, req.DisplayAreas[0].MaxDuration, int64(10))
	assert.Equal(t, req.DisplayAreas[0].MinDuration, int64(0))
}

func TestUpdateAdRequestIncompleteDisplayAreaParams(t *testing.T) {
	params := map[string]string{
		"vistar.url":                 "staging-url",
		"vistar.api_key":             "api-key",
		"vistar.network_id":          "network-id",
		"vistar.venue_id":            "venue-id",
		"vistar.direct_connection":   "true",
		"vistar.latitude":            "45.5",
		"vistar.longitude":           "44.4",
		"vistar.mime_types":          "a,b,c",
		"vistar.width":               "100",
		"vistar.height":              "200",
		"vistar.allow_audio":         "true",
		"vistar.static_duration":     "9",
		"vistar.max_duration":        "19",
		"vistar.min_duration":        "9",
		"vistar.required_completion": "9",
		"vistar.duration":            "300",
		"vistar.interval":            "60",
	}

	conf := &adConfig{}
	conf.parse(params)

	req := &AdRequest{
		DisplayAreas: []DisplayArea{
			DisplayArea{Id: "d1", AllowAudio: false},
			DisplayArea{
				Id:             "d2",
				Width:          500,
				Height:         500,
				AllowAudio:     true,
				StaticDuration: 5,
				MinDuration:    5,
				MaxDuration:    15,
				SupportedMedia: []string{"image"}},
		},
		DeviceAttributes: []DeviceAttribute{
			DeviceAttribute{Name: "attr1", Value: "value1"},
			DeviceAttribute{Name: "attr2", Value: "value2"},
		},
	}

	conf.UpdateAdRequest(req)

	assert.Equal(t, req.ApiKey, "api-key")
	assert.Equal(t, req.NetworkId, "network-id")
	assert.Equal(t, req.DeviceId, "venue-id")
	assert.Equal(t, req.VenueId, "venue-id")
	assert.True(t, req.DirectConnection)
	assert.Equal(t, req.Latitude, 45.5)
	assert.Equal(t, req.Longitude, 44.4)
	assert.NotEqual(t, req.DisplayTime, int64(0))
	assert.Equal(t, req.RequiredCompletion, 9.0)
	assert.Equal(t, req.NumberOfScreens, int64(1))
	assert.Len(t, req.DisplayAreas, 2)
	assert.Len(t, req.DeviceAttributes, 4)
	assert.Equal(t, req.Duration, int64(300))
	assert.Equal(t, req.Interval, int64(60))

	assert.Equal(t, req.DisplayAreas[0].Width, int64(100))
	assert.Equal(t, req.DisplayAreas[0].Height, int64(200))
	assert.True(t, req.DisplayAreas[0].AllowAudio)
	assert.Equal(t, req.DisplayAreas[0].SupportedMedia, []string{"a", "b", "c"})
	assert.Equal(t, req.DisplayAreas[0].StaticDuration, int64(9))
	assert.Equal(t, req.DisplayAreas[0].MinDuration, int64(9))
	assert.Equal(t, req.DisplayAreas[0].MaxDuration, int64(19))

	assert.Equal(t, req.DisplayAreas[1].Width, int64(500))
	assert.Equal(t, req.DisplayAreas[1].Height, int64(500))
	assert.True(t, req.DisplayAreas[1].AllowAudio)
	assert.Equal(t, req.DisplayAreas[1].SupportedMedia, []string{"image"})
	assert.Equal(t, req.DisplayAreas[1].StaticDuration, int64(5))
	assert.Equal(t, req.DisplayAreas[1].MinDuration, int64(5))
	assert.Equal(t, req.DisplayAreas[1].MaxDuration, int64(15))
}

func TestParseDimensionString(t *testing.T) {
	conf := &adConfig{}
	dims := conf.parseDimensionString("")
	assert.Equal(t, len(dims), 0)
	dims = conf.parseDimensionString("1,2")
	assert.Equal(t, len(dims), 0)
	dims = conf.parseDimensionString(",")
	assert.Equal(t, len(dims), 0)
	dims = conf.parseDimensionString("axb")
	assert.Equal(t, len(dims), 0)
	dims = conf.parseDimensionString("100x200")
	assert.Equal(t, len(dims), 1)
	assert.Equal(t, dims[0], Dimension{Width: 100, Height: 200})
	dims = conf.parseDimensionString("100x200,")
	assert.Equal(t, len(dims), 1)
	assert.Equal(t, dims[0], Dimension{Width: 100, Height: 200})
	dims = conf.parseDimensionString("100x 200  , ,,")
	assert.Equal(t, len(dims), 1)
	assert.Equal(t, dims[0], Dimension{Width: 100, Height: 200})
	dims = conf.parseDimensionString("100x 200  , ,100x,")
	assert.Equal(t, len(dims), 1)
	assert.Equal(t, dims[0], Dimension{Width: 100, Height: 200})
	dims = conf.parseDimensionString("100x 200  , ,300x500,")
	assert.Equal(t, len(dims), 2)
	assert.Equal(t, dims[0], Dimension{Width: 100, Height: 200})
	assert.Equal(t, dims[1], Dimension{Width: 300, Height: 500})
}

func TestParseAssetEndpointDisplayAreas(t *testing.T) {
	conf := &adConfig{}
	params := map[string]string{
		"vistar.url":        "staging-url",
		"vistar.mime_types": "a,b,c",
	}
	areas := conf.parseAssetEndpointDisplayAreas(params, true, []string{"a"})
	assert.Equal(t, len(areas), 0)

	params = map[string]string{
		"vistar.url":            "staging-url",
		"vistar.creative_sizes": "a,b,c",
	}
	areas = conf.parseAssetEndpointDisplayAreas(params, true, []string{"a"})
	assert.Equal(t, len(areas), 0)

	params = map[string]string{
		"vistar.url":            "staging-url",
		"vistar.creative_sizes": "100x200,300x  500",
	}
	areas = conf.parseAssetEndpointDisplayAreas(params, true, []string{"a"})
	assert.Equal(t, len(areas), 2)
	assert.Equal(t, areas[0], DisplayArea{
		Width:          100,
		Height:         200,
		Id:             "display-0",
		AllowAudio:     true,
		SupportedMedia: []string{"a"}})
	assert.Equal(t, areas[1], DisplayArea{
		Width:          300,
		Height:         500,
		Id:             "display-1",
		AllowAudio:     true,
		SupportedMedia: []string{"a"}})
}
