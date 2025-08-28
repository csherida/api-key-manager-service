package usecase

import (
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/domain"
	"time"
)

type Repository interface {
	StoreApiKey(apiKey *domain.ApiKey) error
	GetApiKey(apiId string) (*domain.ApiKey, bool, error)
	GetApiKeyByPublicKey(privateKey string) (*domain.ApiKey, error)
	GetAllApiKeys() ([]*domain.ApiKey, error)
	GetAllActiveApiKeys() ([]*domain.ApiKey, error)
	DeleteApiKey(apiId string) error
	ExpireApiKey(apiId string, expirationDate *time.Time) error
	StoreApiUsage(usage *domain.ApiUsage) error
	GetApiUsage(apiId string) ([]*domain.ApiUsage, error)
	GetLatestApiUsage(apiId string) (*domain.ApiUsage, error)
	GetAllApiUsages() (map[string][]*domain.ApiUsage, error)
}
