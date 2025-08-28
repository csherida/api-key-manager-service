package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ApiKeyListHandler struct {
	apiKeyLister ApiKeyLister
}

func NewApiKeyListHandler(apiKeyLister ApiKeyLister) ApiKeyListHandler {
	return ApiKeyListHandler{apiKeyLister: apiKeyLister}
}

func (a ApiKeyListHandler) ListApiKeys(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received a request to list API Keys")

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	// Get the list of API keys with their usage stats
	apiKeyList, err := a.apiKeyLister.ListApiKeys(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the list
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "   ")
	if err := enc.Encode(apiKeyList); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
