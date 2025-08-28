package domain

import "time"

type ApiKeyListResponse struct {
	ApiKeys []ApiKeyWithStats `json:"api_keys"`
	Total   int               `json:"total"`
}

type ApiKeyWithStats struct {
	ApiId            string     `json:"api_id"`
	OrganizationName string     `json:"organization_name"`
	ExpirationDate   *time.Time `json:"expiration_date"`
	IsExpired        bool       `json:"is_expired"`
	UsageStats       UsageStats `json:"usage_stats"`
}

type UsageStats struct {
	TotalRequests    uint64     `json:"total_requests"`
	LastUsed         *time.Time `json:"last_used,omitempty"`
	UniqueIPCount    int        `json:"unique_ip_count"`
	MostRecentIP     string     `json:"most_recent_ip,omitempty"`
}