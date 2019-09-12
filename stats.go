package vistar

import (
	"net/http"
	"net/http/httputil"
	"strings"
)

type Stats struct {
	Average       float64 `json:"average_per_request"`
	BytesReceived int64   `json:"bytes_received"`
	BytesSent     int64   `json:"bytes_sent"`
	Count         int64   `json:"count"`
	Total         int64   `json:"total_bytes"`
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

	headerLength := 0

	for k, v := range req.Header {
		headerLength += len(k)
		headerLength += len(strings.Join(v[:], ","))
	}

	return req.ContentLength + int64(headerLength)
}

func updateStats(stats *Stats, bytesSent int64, bytesReceived int64) {
	stats.BytesSent += bytesSent
	stats.BytesReceived += bytesReceived
	stats.Count += 1
	stats.Total = stats.BytesSent + stats.BytesReceived
	stats.Average = float64(stats.Total) / float64(stats.Count)
}
