//go:build e2e

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
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

		// Test API key validation
		t.Run("TestApiKeyValidation", func(t *testing.T) {
			// Create a request to validate the API key
			client := &http.Client{}
			req, err := http.NewRequest("POST", "http://localhost:8080/keys/validate", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			// Add the API key to the Authorization header
			// Note: In a real scenario, this would be the private key hex
			// For now, we're using the returned key from generation
			req.Header.Add("Authorization", "Bearer "+apiKeyResponse.ApiKey)

			// Make the request
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("failed to make validation request: %v", err)
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

			// Parse validation response
			var validationResponse map[string]interface{}
			if err := json.Unmarshal(body, &validationResponse); err != nil {
				t.Fatalf("failed to unmarshal validation response: %v", err)
			}

			// Check validation response
			if valid, ok := validationResponse["valid"].(bool); !ok || !valid {
				t.Error("API key validation failed")
			}

			if validationResponse["api_id"] != apiKeyResponse.ApiId {
				t.Errorf("API ID mismatch: got %v, want %v", validationResponse["api_id"], apiKeyResponse.ApiId)
			}

			t.Logf("Successfully validated API key with ID: %v", validationResponse["api_id"])
		})

		// Test API key deletion/expiration
		t.Run("TestApiKeyDeletion", func(t *testing.T) {
			// Create a request to delete/expire the API key
			client := &http.Client{}
			deleteURL := fmt.Sprintf("http://localhost:8080/keys/%s", apiKeyResponse.ApiId)
			req, err := http.NewRequest("DELETE", deleteURL, nil)
			if err != nil {
				t.Fatalf("failed to create delete request: %v", err)
			}

			// Make the delete request
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("failed to make delete request: %v", err)
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

			// Parse deletion response
			var deletionResponse map[string]interface{}
			if err := json.Unmarshal(body, &deletionResponse); err != nil {
				t.Fatalf("failed to unmarshal deletion response: %v", err)
			}

			// Check deletion response
			if success, ok := deletionResponse["success"].(bool); !ok || !success {
				t.Error("API key deletion failed")
			}

			if deletionResponse["api_id"] != apiKeyResponse.ApiId {
				t.Errorf("API ID mismatch: got %v, want %v", deletionResponse["api_id"], apiKeyResponse.ApiId)
			}

			t.Logf("Successfully deleted/expired API key with ID: %v", deletionResponse["api_id"])

			// Verify the key is now expired by trying to validate it again
			validateReq, err := http.NewRequest("POST", "http://localhost:8080/keys/validate", nil)
			if err != nil {
				t.Fatalf("failed to create validation request: %v", err)
			}
			validateReq.Header.Add("Authorization", "Bearer "+apiKeyResponse.ApiKey)

			validateResp, err := client.Do(validateReq)
			if err != nil {
				t.Fatalf("failed to make validation request: %v", err)
			}
			defer validateResp.Body.Close()

			// Should return unauthorized now since key is expired
			if validateResp.StatusCode != http.StatusUnauthorized {
				t.Errorf("expected validation to fail with unauthorized status after deletion, got %d", validateResp.StatusCode)
			}
		})
	})
}
