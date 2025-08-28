package api

import (
	"context"
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/domain"
)

type ApiKeyGenerator interface {
	GenerateApiKey(ctx context.Context, organizationName string) (string, string, error)
}

type ApiKeyValidator interface {
	ValidateApiKey(ctx context.Context, privateKey string, ipAddress string) (*domain.ApiKey, error)
}

type ApiKeyDeleter interface {
	ExpireApiKey(_ context.Context, apiId string) error
}

type ApiKeyManager interface {
	GetAllApiKeyUsage(ctx context.Context) []domain.ApiKey
	ExpireApiKey(ctx context.Context, apiId string) error
}
