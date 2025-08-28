package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type ApiKeyDeletionHandler struct {
	apiKeyDeleter ApiKeyDeleter
}

type ApiKeyDeletionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ApiId   string `json:"api_id,omitempty"`
}

func NewApiKeyDeletionHandler(apiKeyDeleter ApiKeyDeleter) ApiKeyDeletionHandler {
	return ApiKeyDeletionHandler{apiKeyDeleter: apiKeyDeleter}
}

func (a ApiKeyDeletionHandler) DeleteApiKey(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received a request to delete/expire an API Key")

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	// Extract keyId from URL path
	vars := mux.Vars(r)
	keyId := vars["keyId"]

	if keyId == "" {
		respondWithDeletion(w, false, "", "Missing API key ID", http.StatusBadRequest)
		return
	}

	// Expire the API key
	if err := a.apiKeyDeleter.ExpireApiKey(ctx, keyId); err != nil {
		// Check if it's a not found error
		if err.Error() == fmt.Sprintf("API key with ID %s not found", keyId) {
			respondWithDeletion(w, false, keyId, "API key not found", http.StatusNotFound)
			return
		}
		respondWithDeletion(w, false, keyId, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return successful deletion response
	respondWithDeletion(w, true, keyId, "API key successfully expired", http.StatusOK)
}

func respondWithDeletion(w http.ResponseWriter, success bool, apiId, message string, statusCode int) {
	response := ApiKeyDeletionResponse{
		Success: success,
		Message: message,
		ApiId:   apiId,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "   ")
	if err := enc.Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
