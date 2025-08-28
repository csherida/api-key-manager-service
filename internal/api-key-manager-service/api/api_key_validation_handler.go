package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ApiKeyValidationHandler struct {
	apiKeyValidator ApiKeyValidator
}

type ApiKeyValidationResponse struct {
	Valid            bool   `json:"valid"`
	ApiId            string `json:"api_id,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
	Message          string `json:"message,omitempty"`
}

func NewApiKeyValidationHandler(apiKeyValidator ApiKeyValidator) ApiKeyValidationHandler {
	return ApiKeyValidationHandler{apiKeyValidator: apiKeyValidator}
}

func (a ApiKeyValidationHandler) ValidateApiKey(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received a request to validate an API Key")

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	// Extract the private key from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithValidation(w, false, "", "", "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	// Expected format: "Bearer <private_key>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		respondWithValidation(w, false, "", "", "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}

	privateKey := parts[1]

	// Get client IP address
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}

	// Validate the API key
	apiKey, err := a.apiKeyValidator.ValidateApiKey(ctx, privateKey, ipAddress)
	if err != nil {
		respondWithValidation(w, false, "", "", err.Error(), http.StatusUnauthorized)
		return
	}

	// Return successful validation response
	respondWithValidation(w, true, apiKey.ApiId, apiKey.OrganizationName, "API key is valid", http.StatusOK)
}

func respondWithValidation(w http.ResponseWriter, valid bool, apiId, orgName, message string, statusCode int) {
	response := ApiKeyValidationResponse{
		Valid:            valid,
		ApiId:            apiId,
		OrganizationName: orgName,
		Message:          message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "   ")
	if err := enc.Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
