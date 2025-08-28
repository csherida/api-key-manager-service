package usecase

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/domain"
	"github.com/google/uuid"
	"log"
)

type ApiKeyGeneration struct {
	repo Repository
}

type KeyPair struct {
	PublicKey  string
	PrivateKey string
}

func NewApiKeyGeneration(repo Repository) ApiKeyGeneration {
	return ApiKeyGeneration{repo: repo}
}

func (a ApiKeyGeneration) GenerateApiKey(_ context.Context, organizationName string) (string, string, error) {
	apiId := uuid.NewString()
	keyPair, err := a.generateKeyPair()
	if err != nil {
		return "", "", err
	}

	apiKey := domain.ApiKey{
		ApiId:            apiId,
		PrivateKey:       keyPair.PrivateKey,
		OrganizationName: organizationName,
	}
	if err := a.repo.StoreApiKey(&apiKey); err != nil {
		log.Printf("Failed to store api key for organization %s: %v", organizationName, err)
		return "", "", err
	}

	return apiId, keyPair.PublicKey, nil
}

func (a ApiKeyGeneration) generateKeyPair() (*KeyPair, error) {
	// Generate RSA key pair with 4096 bits
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	// Encode private key to PKCS8 PEM format
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key to SPKI PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return &KeyPair{
		PublicKey:  string(publicKeyPEM),
		PrivateKey: string(privateKeyPEM),
	}, nil
}
