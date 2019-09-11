package vistar

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
)

type Stats struct {
	Average       string `json:"average_per_request"`
	BytesReceived int64  `json:"bytes_received"`
	BytesSent     int64  `json:"bytes_sent"`
	Count         int64  `json:"count"`
	Total         int64  `json:"total_bytes"`
}

func getResponseLength(resp *http.Response) int64 {
	if resp == nil {
		return 0
	}
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return 0
	}

	return int64(len(dump))
}

func getRequestLength(req *http.Request) int64 {
	if req == nil {
		return 0
	}

	content := req.ContentLength
	header, err := json.Marshal(req.Header)
	if err != nil {
		return content
	}

	return content + int64(len(header))
}

func calcAverage(total int64, count int64) string {
	if count == 0 {
		return "0"
	}

	return fmt.Sprintf("%.2f", float64(total)/float64(count))
}
