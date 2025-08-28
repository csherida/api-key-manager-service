package usecase

import (
	"context"
	"fmt"
	"time"
)

type ApiKeyDeletion struct {
	repo Repository
}

func NewApiKeyDeletion(repo Repository) ApiKeyDeletion {
	return ApiKeyDeletion{repo: repo}
}

func (a ApiKeyDeletion) ExpireApiKey(_ context.Context, apiId string) error {
	// First check if the API key exists
	apiKey, exists, err := a.repo.GetApiKey(apiId)
	if err != nil {
		return fmt.Errorf("failed to retrieve API key: %w", err)
	}
	if !exists || apiKey == nil {
		return fmt.Errorf("API key with ID %s not found", apiId)
	}

	// Set expiration date to now (immediately expired)
	now := time.Now()
	if err := a.repo.ExpireApiKey(apiId, &now); err != nil {
		return fmt.Errorf("failed to expire API key: %w", err)
	}

	return nil
}
