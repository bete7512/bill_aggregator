// internal/adapter/outbound/provider/client/http.go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bete7512/bill-aggregator/internal/core/domain/model"
	"github.com/bete7512/bill-aggregator/pkg/util/resilience"
)

// ProviderClient implements the provider client interface using HTTP
type ProviderClient struct {
	httpClient      *http.Client
	baseURL         string
	retryPolicy     *resilience.RetryPolicy
	circuitBreaker  *resilience.CircuitBreaker
	providerInfoCache map[string]*model.Provider
}

// NewProviderClient creates a new provider client
func NewProviderClient(baseURL string) *ProviderClient {
	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create retry policy
	retryPolicy := resilience.NewRetryPolicy(
		3,                   // Max retries
		1*time.Second,       // Initial delay
		5*time.Second,       // Max delay
		2.0,                 // Backoff factor
		[]int{408, 429, 500, 502, 503, 504}, // Retryable status codes
	)

	// Create circuit breaker
	circuitBreaker := resilience.NewCircuitBreaker(
		5,                   // Failure threshold
		30*time.Second,      // Reset timeout
		10,                  // Concurrent requests
	)

	return &ProviderClient{
		httpClient:      httpClient,
		baseURL:         baseURL,
		retryPolicy:     retryPolicy,
		circuitBreaker:  circuitBreaker,
		providerInfoCache: make(map[string]*model.Provider),
	}
}

// Authenticate authenticates with the provider using the given credentials
func (c *ProviderClient) Authenticate(ctx context.Context, providerID string, credentials map[string]string) (*model.Credential, error) {
	// Prepare request URL
	url := fmt.Sprintf("%s/providers/%s/authenticate", c.baseURL, providerID)

	// Prepare request body
	_, err := json.Marshal(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request with retry and circuit breaker
	var authResponse struct {
		Token     string    `json:"token"`
		Type      string    `json:"type"`
		ExpiresAt time.Time `json:"expires_at"`
	}

	err = c.executeRequest(ctx, req, &authResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	// Create credential
	credential := &model.Credential{
		Type: authResponse.Type,
		Details: map[string]string{
			"token": authResponse.Token,
		},
		ExpiresAt: &authResponse.ExpiresAt,
		IsExpired: time.Now().After(authResponse.ExpiresAt),
	}

	return credential, nil
}

// FetchBills retrieves bills from the provider
func (c *ProviderClient) FetchBills(ctx context.Context, linkedAccount *model.LinkedAccount) ([]model.Bill, error) {
	// Prepare request URL
	url := fmt.Sprintf("%s/providers/%s/accounts/%s/bills", 
		c.baseURL, 
		linkedAccount.ProviderID, 
		linkedAccount.AccountNumber,
	)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header based on credential type
	c.setAuthHeader(req, linkedAccount.Credentials)

	// Execute request with retry and circuit breaker
	var billsResponse struct {
		Bills []struct {
			ID           string                 `json:"id"`
			BillNumber   string                 `json:"bill_number"`
			Amount       float64                `json:"amount"`
			Currency     string                 `json:"currency"`
			Status       string                 `json:"status"`
			DueDate      time.Time              `json:"due_date"`
			IssuedDate   time.Time              `json:"issued_date"`
			PaidDate     *time.Time             `json:"paid_date,omitempty"`
			PeriodStart  time.Time              `json:"period_start"`
			PeriodEnd    time.Time              `json:"period_end"`
			Details      map[string]interface{} `json:"details"`
		} `json:"bills"`
	}

	err = c.executeRequest(ctx, req, &billsResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bills: %w", err)
	}

	// Convert response to domain models
	now := time.Now()
	bills := make([]model.Bill, 0, len(billsResponse.Bills))
	for _, b := range billsResponse.Bills {
		bill := model.Bill{
			ID:              b.ID,
			UserID:          linkedAccount.UserID,
			LinkedAccountID: linkedAccount.ID,
			ProviderID:      linkedAccount.ProviderID,
			BillNumber:      b.BillNumber,
			Amount:          b.Amount,
			Currency:        b.Currency,
			Status:          b.Status,
			DueDate:         b.DueDate,
			IssuedDate:      b.IssuedDate,
			PaidDate:        b.PaidDate,
			PeriodStart:     b.PeriodStart,
			PeriodEnd:       b.PeriodEnd,
			Details:         b.Details,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		bills = append(bills, bill)
	}

	return bills, nil
}

// ValidateCredentials validates the credentials for a provider
func (c *ProviderClient) ValidateCredentials(ctx context.Context, providerID string, credentials map[string]string) (bool, error) {
	// Prepare request URL
	url := fmt.Sprintf("%s/providers/%s/validate", c.baseURL, providerID)

	// Prepare request body
	_, err := json.Marshal(credentials)
	if err != nil {
		return false, fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request with retry and circuit breaker
	var validateResponse struct {
		Valid bool `json:"valid"`
	}

	err = c.executeRequest(ctx, req, &validateResponse)
	if err != nil {
		return false, fmt.Errorf("failed to validate credentials: %w", err)
	}

	return validateResponse.Valid, nil
}

// GetProviderInfo retrieves information about a provider
func (c *ProviderClient) GetProviderInfo(ctx context.Context, providerID string) (*model.Provider, error) {
	// Check cache first
	if provider, exists := c.providerInfoCache[providerID]; exists {
		return provider, nil
	}

	// Prepare request URL
	url := fmt.Sprintf("%s/providers/%s", c.baseURL, providerID)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request with retry and circuit breaker
	var providerResponse model.Provider

	err = c.executeRequest(ctx, req, &providerResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider info: %w", err)
	}

	// Cache the result
	c.providerInfoCache[providerID] = &providerResponse

	return &providerResponse, nil
}

// GetAllProviders retrieves all available providers
func (c *ProviderClient) GetAllProviders(ctx context.Context) ([]model.Provider, error) {
	// Prepare request URL
	url := fmt.Sprintf("%s/providers", c.baseURL)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request with retry and circuit breaker
	var providersResponse struct {
		Providers []model.Provider `json:"providers"`
	}

	err = c.executeRequest(ctx, req, &providersResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}

	// Cache the providers
	for i := range providersResponse.Providers {
		c.providerInfoCache[providersResponse.Providers[i].ID] = &providersResponse.Providers[i]
	}

	return providersResponse.Providers, nil
}

// executeRequest executes an HTTP request with retry and circuit breaker
func (c *ProviderClient) executeRequest(ctx context.Context, req *http.Request, response interface{}) error {
	// Execute with circuit breaker
	err := c.circuitBreaker.Execute(func() error {
		// Execute with retry
		return c.retryPolicy.Execute(func() error {
			// Execute the request
			resp, err := c.httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			}

			// Decode response
			if response != nil {
				err = json.NewDecoder(resp.Body).Decode(response)
				if err != nil {
					return fmt.Errorf("failed to decode response: %w", err)
				}
			}

			return nil
		})
	})

	return err
}

// setAuthHeader sets the authentication header based on credential type
func (c *ProviderClient) setAuthHeader(req *http.Request, credential model.Credential) {
	switch credential.Type {
	case "token", "api_key":
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", credential.Details["token"]))
	case "basic":
		// In a real application, this would be base64 encoded username:password
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", credential.Details["token"]))
	case "oauth":
		req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", credential.Details["token"]))
	}
}


