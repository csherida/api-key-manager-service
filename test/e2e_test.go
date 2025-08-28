//go:build e2e

package test

import (
	"bytes"
	"encoding/json"
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/domain"
	"github.com/csherida/api-key-manager-service/internal/service/di"
	"io"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestApiKeyManager(t *testing.T) {
	application, err := di.SetupApplication()
	if err != nil {
		t.Fatalf("failed to setup application: %v", err)
	}
	t.Cleanup(application.CancelContext)

	go func() {
		if err := application.Run(); err != nil {
			log.Fatalf("failed to start application: %v", err)
		}
	}()
	time.Sleep(3 * time.Second)

	t.Run("TestApiKeyGeneration", func(t *testing.T) {
		// Create request payload
		request := domain.ApiKeyGeneratorRequest{
			OrganizationName: "TestOrganization",
		}

		// Marshal request to JSON
		jsonData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		// Make POST request to generate API key
		resp, err := http.Post("http://localhost:8080/keys", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("failed to make POST request: %v", err)
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected status OK, got %d: %s", resp.StatusCode, string(body))
		}

		// Read and validate response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		// Parse response (expecting the API key response structure)
		var apiKeyResponse domain.ApiKeyGeneratorResponse
		if err := json.Unmarshal(body, &apiKeyResponse); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		// Validate response contains expected fields
		if apiKeyResponse.ApiId == "" {
			t.Error("response missing api_id field")
		}
		if apiKeyResponse.ApiKey == "" {
			t.Error("response missing api_key field")
		}

		t.Logf("Successfully generated API key with ID: %v", apiKeyResponse.ApiId)
	})
}
