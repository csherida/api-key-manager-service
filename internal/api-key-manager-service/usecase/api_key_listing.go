package usecase

import (
	"context"
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/domain"
	"github.com/samber/lo"
	"time"
)

type ApiKeyListing struct {
	repo Repository
}

func NewApiKeyListing(repo Repository) ApiKeyListing {
	return ApiKeyListing{repo: repo}
}

func (a ApiKeyListing) ListApiKeys(ctx context.Context) (*domain.ApiKeyListResponse, error) {
	// Get all API keys
	allApiKeys, err := a.repo.GetAllApiKeys()
	if err != nil {
		return nil, err
	}

	// Get all usage data
	allUsages, err := a.repo.GetAllApiUsages()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var apiKeysWithStats []domain.ApiKeyWithStats

	for _, apiKey := range allApiKeys {
		// Calculate usage stats for this API key
		usages := allUsages[apiKey.ApiId]
		stats := calculateUsageStats(usages)

		// Check if expired
		isExpired := apiKey.ExpirationDate != nil && apiKey.ExpirationDate.Before(now)

		apiKeyWithStats := domain.ApiKeyWithStats{
			ApiId:            apiKey.ApiId,
			OrganizationName: apiKey.OrganizationName,
			ExpirationDate:   apiKey.ExpirationDate,
			IsExpired:        isExpired,
			UsageStats:       stats,
		}

		apiKeysWithStats = append(apiKeysWithStats, apiKeyWithStats)
	}

	// Sort by creation order (newest first based on API ID)
	apiKeysWithStats = lo.Reverse(apiKeysWithStats)

	return &domain.ApiKeyListResponse{
		ApiKeys: apiKeysWithStats,
		Total:   len(apiKeysWithStats),
	}, nil
}

func calculateUsageStats(usages []*domain.ApiUsage) domain.UsageStats {
	if len(usages) == 0 {
		return domain.UsageStats{
			TotalRequests: 0,
			UniqueIPCount: 0,
		}
	}

	// Find the highest cumulative request count (latest usage)
	maxUsage := lo.MaxBy(usages, func(a, b *domain.ApiUsage) bool {
		return a.CumulativeRequest > b.CumulativeRequest
	})

	// Find the most recent usage by timestamp
	mostRecentUsage := lo.MaxBy(usages, func(a, b *domain.ApiUsage) bool {
		return a.ValidatedAt.After(b.ValidatedAt)
	})

	// Count unique IP addresses
	uniqueIPs := lo.Uniq(lo.Map(usages, func(usage *domain.ApiUsage, _ int) string {
		return usage.IpAddress
	}))

	return domain.UsageStats{
		TotalRequests: maxUsage.CumulativeRequest,
		LastUsed:      &mostRecentUsage.ValidatedAt,
		UniqueIPCount: len(uniqueIPs),
		MostRecentIP:  mostRecentUsage.IpAddress,
	}
}