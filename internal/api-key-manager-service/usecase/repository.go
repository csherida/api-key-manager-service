package usecase

import "github.com/csherida/api-key-manager-service/internal/api-key-manager-service/domain"

type Repository interface {
	StoreApiKey(apiKey *domain.ApiKey) error
	GetApiKey(apiId string) (*domain.ApiKey, bool, error)
	GetAllActiveApiKeys() ([]*domain.ApiKey, error)
	DeleteApiKey(apiId string) error
	StoreApiUsage(usage *domain.ApiUsage) error
	GetApiUsage(apiId string) ([]*domain.ApiUsage, error)
	GetLatestApiUsage(apiId string) (*domain.ApiUsage, error)
	GetAllApiUsages() (map[string][]*domain.ApiUsage, error)
}
