package usecase

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/domain"
	"github.com/ethereum/go-ethereum/crypto"
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
	keyPair, err := generateKeyPair()
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

func generateKeyPair() (*KeyPair, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Printf("Failed to generate private key: %v\n", err)
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("failed to cast public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &KeyPair{
		PublicKey:  fmt.Sprintf("%x", crypto.FromECDSA(privateKey)),
		PrivateKey: address.Hex(),
	}, nil
}
