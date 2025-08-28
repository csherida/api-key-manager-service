package infra

import (
	"sync"
	"time"

	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/domain"
	"github.com/samber/lo"
)

type DataStore struct {
	mu              sync.RWMutex
	apiKeys         map[string]*domain.ApiKey     // keyed by ApiId
	apiKeysByPublic map[string]*domain.ApiKey     // keyed by public key (PrivateKey field)
	apiUsages       map[string][]*domain.ApiUsage // keyed by ApiId
}

func NewDataStore() *DataStore {
	return &DataStore{
		apiKeys:         make(map[string]*domain.ApiKey),
		apiKeysByPublic: make(map[string]*domain.ApiKey),
		apiUsages:       make(map[string][]*domain.ApiUsage),
	}
}

// StoreApiKey stores an API key in the data store
func (ds *DataStore) StoreApiKey(apiKey *domain.ApiKey) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.apiKeys[apiKey.ApiId] = apiKey
	// Also store by public key for fast lookup
	if apiKey.PrivateKey != "" {
		ds.apiKeysByPublic[apiKey.PrivateKey] = apiKey
	}
	return nil
}

// GetApiKey retrieves an API key by ID
func (ds *DataStore) GetApiKey(apiId string) (*domain.ApiKey, bool, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	apiKey, exists := ds.apiKeys[apiId]
	return apiKey, exists, nil
}

// GetApiKeyByPublicKey retrieves an API key by its public key (stored in PrivateKey field)
func (ds *DataStore) GetApiKeyByPublicKey(publicKey string) (*domain.ApiKey, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	
	apiKey, exists := ds.apiKeysByPublic[publicKey]
	if !exists {
		return nil, nil
	}
	return apiKey, nil
}

// GetAllActiveApiKeys returns all API keys that haven't expired yet or nil if none exist
func (ds *DataStore) GetAllActiveApiKeys() ([]*domain.ApiKey, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	now := time.Now()
	var activeKeys []*domain.ApiKey

	for _, apiKey := range ds.apiKeys {
		// Check if the key hasn't expired
		if apiKey.ExpirationDate.After(now) {
			activeKeys = append(activeKeys, apiKey)
		}
	}

	if len(activeKeys) == 0 {
		return nil, nil
	}

	return activeKeys, nil
}

// DeleteApiKey removes an API key from the store
func (ds *DataStore) DeleteApiKey(apiId string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	
	// Get the key first to find its public key
	if apiKey, exists := ds.apiKeys[apiId]; exists {
		delete(ds.apiKeys, apiId)
		// Also delete from public key map
		if apiKey.PrivateKey != "" {
			delete(ds.apiKeysByPublic, apiKey.PrivateKey)
		}
	}
	return nil
}

// StoreApiUsage stores API usage data with auto-incremented CumulativeRequest
func (ds *DataStore) StoreApiUsage(usage *domain.ApiUsage) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Find the highest CumulativeRequest for this API ID using lo.MaxBy
	existingUsages := ds.apiUsages[usage.ApiId]
	if len(existingUsages) > 0 {
		maxUsage := lo.MaxBy(existingUsages, func(a, b *domain.ApiUsage) bool {
			return a.CumulativeRequest > b.CumulativeRequest
		})
		usage.CumulativeRequest = maxUsage.CumulativeRequest + 1
	} else {
		usage.CumulativeRequest = 1
	}

	ds.apiUsages[usage.ApiId] = append(ds.apiUsages[usage.ApiId], usage)
	return nil
}

// GetApiUsage retrieves all usage records for a specific API ID
func (ds *DataStore) GetApiUsage(apiId string) ([]*domain.ApiUsage, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	usages, exists := ds.apiUsages[apiId]
	if !exists {
		return nil, nil
	}

	// Return a copy to prevent external modification
	result := make([]*domain.ApiUsage, len(usages))
	copy(result, usages)
	return result, nil
}

// GetLatestApiUsage retrieves the most recent usage record for a specific API ID
func (ds *DataStore) GetLatestApiUsage(apiId string) (*domain.ApiUsage, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	usages, exists := ds.apiUsages[apiId]
	if !exists || len(usages) == 0 {
		return nil, nil
	}

	// Find the most recent usage based on ValidatedAt
	latest := usages[0]
	for _, usage := range usages[1:] {
		if usage.ValidatedAt.After(latest.ValidatedAt) {
			latest = usage
		}
	}

	return latest, nil
}

// GetAllApiUsages returns all API usage records
func (ds *DataStore) GetAllApiUsages() (map[string][]*domain.ApiUsage, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string][]*domain.ApiUsage)
	for apiId, usages := range ds.apiUsages {
		copiedUsages := make([]*domain.ApiUsage, len(usages))
		copy(copiedUsages, usages)
		result[apiId] = copiedUsages
	}

	return result, nil
}
