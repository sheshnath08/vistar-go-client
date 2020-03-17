package vistar

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

type Data struct {
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

type Request interface {
	Data() *Data
	ServerUrl() string
	AssetEndpointUrl() string
	AssetEndpointDisplayAreas() []DisplayArea
	LogLevel() int64
	LogEnabled() bool
}

type request struct {
	data                      *Data
	assetEndpointUrl          string
	assetEndpointDisplayAreas []DisplayArea
	logEnabled                bool
	logLevel                  int64
	url                       string
}

func NewRequest(url string, assetEndpointUrl string, data *Data,
	logEnabled bool, logLevel int64) *request {
	return &request{
		url:              url,
		assetEndpointUrl: assetEndpointUrl,
		data:             data,
		logEnabled:       logEnabled,
		logLevel:         logLevel,
	}
}

func (r request) Data() *Data {
	return r.data
}

func (r request) LogEnabled() bool {
	return r.logEnabled
}

func (r request) LogLevel() int64 {
	return r.logLevel
}

func (r request) ServerUrl() string {
	return r.url
}

func (r request) AssetEndpointUrl() string {
	return r.assetEndpointUrl
}

func (r request) AssetEndpointDisplayAreas() []DisplayArea {
	return r.assetEndpointDisplayAreas
}
