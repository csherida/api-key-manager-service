package usecase

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/domain"
	"github.com/ethereum/go-ethereum/crypto"
	"time"
)

type ApiKeyValidation struct {
	repo Repository
}

func NewApiKeyValidation(repo Repository) ApiKeyValidation {
	return ApiKeyValidation{repo: repo}
}

func (a ApiKeyValidation) ValidateApiKey(ctx context.Context, privateKeyHex string, ipAddress string) (*domain.ApiKey, error) {
	// Parse the private key from hex string
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key format: %w", err)
	}

	// Convert bytes to ECDSA private key
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Get public key from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("failed to cast public key to ECDSA")
	}

	// Generate the address (which was stored as PrivateKey in ApiKey)
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	storedAddress := address.Hex()

	// Find the API key by matching the stored "PrivateKey" (which is actually the address)
	apiKey, err := a.repo.GetApiKeyByPublicKey(storedAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve API key: %w", err)
	}
	if apiKey == nil {
		return nil, errors.New("invalid API key")
	}

	// Check if the key has expired
	if apiKey.ExpirationDate != nil && apiKey.ExpirationDate.Before(time.Now()) {
		return nil, errors.New("API key has expired")
	}

	// Store API usage
	usage := &domain.ApiUsage{
		ApiId:       apiKey.ApiId,
		IpAddress:   ipAddress,
		ValidatedAt: time.Now(),
	}

	if err := a.repo.StoreApiUsage(usage); err != nil {
		// Log but don't fail validation if we can't store usage
		fmt.Printf("Failed to store API usage: %v\n", err)
	}

	return apiKey, nil
}
