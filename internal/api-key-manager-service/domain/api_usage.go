package domain

import "time"

type ApiUsage struct {
	ApiId             string    `json:"api_id"`
	IpAddress         string    `json:"ip_address"`
	CumulativeRequest uint64    `json:"cumulative_request"`
	ValidatedAt       time.Time `json:"validated_at"`
}
