package vistar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdRequestData(t *testing.T) {
	adRequestData := &Data{}
	request := NewRequest("url.com", adRequestData, true, int64(0))

	assert.Equal(t, adRequestData, request.Data())
}

func TestServerUrl(t *testing.T) {
	request := NewRequest("ad-server-url.com", nil, true, int64(0))
	assert.Equal(t, request.ServerUrl(), "ad-server-url.com")
}

func TestLogEnabled(t *testing.T) {
	request := NewRequest("", nil, true, int64(0))
	assert.True(t, request.LogEnabled())

	request = NewRequest("", nil, false, int64(0))
	assert.False(t, request.LogEnabled())
}

func TestLogLevel(t *testing.T) {
	request := NewRequest("", nil, true, int64(0))
	assert.Equal(t, request.LogLevel(), int64(0))

	request = NewRequest("", nil, true, int64(2))
	assert.Equal(t, request.LogLevel(), int64(2))
}

func TestSetAssetEndpointUrl(t *testing.T) {
	request := NewRequest("", nil, true, int64(0))
	request.SetAssetEndpointUrl("asset-url.com")

	assert.Equal(t, request.AssetEndpointUrl(), "asset-url.com")
}

func TestSetAssetEndpointDisplayAreas(t *testing.T) {
	var displayAreas []DisplayArea
	request := NewRequest("", nil, true, int64(0))
	request.SetAssetEndpointDisplayAreas(displayAreas)

	assert.Equal(t, request.AssetEndpointDisplayAreas(), displayAreas)
}
