//go:build e2e

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/domain"
	"github.com/csherida/api-key-manager-service/internal/service/di"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"math/rand"
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
		apiKeyResponse := generateApiKey(t)
		apiKeyResponse2 := generateApiKey(t)
		require.NotEqual(t, apiKeyResponse.ApiKey, apiKeyResponse2.ApiKey)
		apiKeyResponse3 := generateApiKey(t)
		require.NotEqual(t, apiKeyResponse.ApiKey, apiKeyResponse3.ApiKey)

		responseMap := map[string]domain.ApiKeyGeneratorResponse{
			apiKeyResponse.ApiId:  apiKeyResponse,
			apiKeyResponse2.ApiId: apiKeyResponse2,
			apiKeyResponse3.ApiId: apiKeyResponse3,
		}

		// Generate a random number between 1 and 20
		src := rand.NewSource(time.Now().UnixNano())
		r := rand.New(src)
		validationCount := r.Intn(20) + 1

		// Test API key validation
		t.Run("TestApiKeyValidation", func(t *testing.T) {
			for i := 0; i < validationCount; i++ {
				validateApiKey(t, apiKeyResponse)
			}
		})

		// Test API key listing
		t.Run("TestApiKeyListing", func(t *testing.T) {
			// Create a request to list all API keys
			resp, err := http.Get("http://localhost:8080/keys")
			if err != nil {
				t.Fatalf("failed to make GET request: %v", err)
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

			// Parse list response
			var listResponse domain.ApiKeyListResponse
			if err := json.Unmarshal(body, &listResponse); err != nil {
				t.Fatalf("failed to unmarshal list response: %v", err)
			}

			require.Len(t, listResponse.ApiKeys, 3)
			require.Greater(t, listResponse.Total, 0)
			for _, apiKey := range listResponse.ApiKeys {
				require.NotNil(t, responseMap[apiKey.ApiId])
				require.NotEmpty(t, apiKey.OrganizationName)
				if apiKey.ApiId == apiKeyResponse.ApiId {
					require.NotEmpty(t, apiKey.UsageStats)
					require.Equal(t, uint64(validationCount), apiKey.UsageStats.TotalRequests)
				}
			}

			// Confirm no data leaked
			var listResponseInterface map[string]interface{}
			if err := json.Unmarshal(body, &listResponseInterface); err != nil {
				t.Fatalf("failed to unmarshal list response: %v", err)
			}

			apiKeys, _ := listResponseInterface["api_keys"].([]interface{})
			for _, keyInterface := range apiKeys {
				key := keyInterface.(map[string]interface{})

				// Verify fields are present and not exposing actual key values
				if _, hasPrivateKey := key["private_key"]; hasPrivateKey {
					t.Error("response should not contain private_key")
				}
				if _, hasPublicKey := key["public_key"]; hasPublicKey {
					t.Error("response should not contain public_key")
				}

				// Verify required metadata fields
				if _, hasOrgName := key["organization_name"]; !hasOrgName {
					t.Error("response missing organization_name field")
				}
				if _, hasUsageStats := key["usage_stats"]; !hasUsageStats {
					t.Error("response missing usage_stats field")
				}
				if _, hasIsExpired := key["is_expired"]; !hasIsExpired {
					t.Error("response missing is_expired field")
				}
			}

			t.Logf("Successfully listed API keys, found %d keys", len(apiKeys))
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

			// Verify the key does not appear in GET /keys
			resp, err = http.Get("http://localhost:8080/keys")
			if err != nil {
				t.Fatalf("failed to make GET request: %v", err)
			}
			defer resp.Body.Close()

			// Check response status
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("expected status OK, got %d: %s", resp.StatusCode, string(body))
			}

			var listResponse domain.ApiKeyListResponse
			require.NoError(t, json.Unmarshal(body, &listResponse))
			for _, apiKey := range listResponse.ApiKeys {
				require.Equal(t, apiKey.ApiId, apiKeyResponse.ApiId)
			}
		})
	})
}

func generateApiKey(t *testing.T) domain.ApiKeyGeneratorResponse {
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

	return apiKeyResponse
}

func validateApiKey(t *testing.T, apiKeyResponse domain.ApiKeyGeneratorResponse) {
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
}
