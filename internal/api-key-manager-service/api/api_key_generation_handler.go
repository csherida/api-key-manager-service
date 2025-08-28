package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/domain"
	"net/http"
	"time"
)

type ApiKeyGeneratorHandler struct {
	apiKeyGenerator ApiKeyGenerator
}

type ApiKeyGeneratorHandlerType func(w http.ResponseWriter, r *http.Request)

func NewApiKeyGeneratorHandler(apiKeyGenerator ApiKeyGenerator) ApiKeyGeneratorHandler {
	return ApiKeyGeneratorHandler{apiKeyGenerator: apiKeyGenerator}
}

func (a ApiKeyGeneratorHandler) ApiKeyGenerator(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received a request to create an API Key")

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	request := domain.ApiKeyGeneratorRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	apiId, apiKey, err := a.apiKeyGenerator.GenerateApiKey(ctx, request.OrganizationName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := domain.ApiKeyGeneratorResponse{
		ApiId:  apiId,
		ApiKey: apiKey,
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "   ")
	if err = enc.Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}
