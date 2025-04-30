// internal/adapter/outbound/provider/mock/mock.go
package mock

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/bete7512/bill-aggregator/internal/core/domain/model"
)

// MockProviderClient implements the provider client interface for testing
type MockProviderClient struct {
	providers map[string]*model.Provider
	bills     map[string][]model.Bill
}

// NewMockProviderClient creates a new mock provider client
func NewMockProviderClient() *MockProviderClient {
	// Initialize with some mock providers
	providers := map[string]*model.Provider{
		"electricity-provider": {
			ID:             "electricity-provider",
			Name:           "PowerCo",
			Type:           "electricity",
			AuthTypes:      []string{"api_key"},
			RequiredFields: []string{"account_number", "api_key"},
			BaseURL:        "https://api.powerco.example.com",
			IconURL:        "https://powerco.example.com/icon.png",
			Active:         true,
		},
		"water-provider": {
			ID:             "water-provider",
			Name:           "WaterWorks",
			Type:           "water",
			AuthTypes:      []string{"api_key"},
			RequiredFields: []string{"account_number", "api_key"},
			BaseURL:        "https://api.waterworks.example.com",
			IconURL:        "https://waterworks.example.com/icon.png",
			Active:         true,
		},
		"internet-provider": {
			ID:             "internet-provider",
			Name:           "NetConnect",
			Type:           "internet",
			AuthTypes:      []string{"oauth"},
			RequiredFields: []string{"account_number", "client_id", "client_secret"},
			BaseURL:        "https://api.netconnect.example.com",
			IconURL:        "https://netconnect.example.com/icon.png",
			Active:         true,
		},
	}

	return &MockProviderClient{
		providers: providers,
		bills:     make(map[string][]model.Bill),
	}
}

// Authenticate authenticates with the provider using the given credentials
func (c *MockProviderClient) Authenticate(ctx context.Context, providerID string, credentials map[string]string) (*model.Credential, error) {
	// Check if provider exists
	if _, exists := c.providers[providerID]; !exists {
		return nil, fmt.Errorf("provider not found: %s", providerID)
	}

	// Generate mock bills if they don't exist yet for this account
	accountKey := fmt.Sprintf("%s-%s", providerID, credentials["account_number"])
	if _, exists := c.bills[accountKey]; !exists {
		linkedAccount := &model.LinkedAccount{
			ProviderID:    providerID,
			AccountNumber: credentials["account_number"],
		}
		c.generateMockBills(linkedAccount)
	}

	// Create and return mock credentials
	return &model.Credential{
		Type: "api_key",
		Details: map[string]string{
			"token": "mock-api-key",
		},
	}, nil
}

// ValidateCredentials validates the credentials for a provider
func (c *MockProviderClient) ValidateCredentials(ctx context.Context, providerID string, credentials map[string]string) (bool, error) {
	// Check if provider exists
	if _, exists := c.providers[providerID]; !exists {
		return false, fmt.Errorf("provider not found: %s", providerID)
	}

	// Simulate validation
	if accountNumber, exists := credentials["account_number"]; !exists || accountNumber == "" {
		return false, nil
	}

	// Simulate API key validation
	if apiKey, exists := credentials["api_key"]; !exists || apiKey == "" {
		// If the provider requires an API key
		if providerID == "electricity-provider" || providerID == "water-provider" {
			return false, nil
		}
	}

	// Simulate OAuth validation
	if providerID == "internet-provider" {
		if clientID, exists := credentials["client_id"]; !exists || clientID == "" {
			return false, nil
		}
		if clientSecret, exists := credentials["client_secret"]; !exists || clientSecret == "" {
			return false, nil
		}
	}

	return true, nil
}

// GetProviderInfo retrieves information about a provider
func (c *MockProviderClient) GetProviderInfo(ctx context.Context, providerID string) (*model.Provider, error) {
	// Check if provider exists
	provider, exists := c.providers[providerID]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", providerID)
	}

	return provider, nil
}

// GetAllProviders retrieves all available providers
func (c *MockProviderClient) GetAllProviders(ctx context.Context) ([]model.Provider, error) {
	providers := make([]model.Provider, 0, len(c.providers))
	for _, provider := range c.providers {
		providers = append(providers, *provider)
	}

	return providers, nil
}

// generateMockBills generates mock bills for a linked account
func (c *MockProviderClient) generateMockBills(linkedAccount *model.LinkedAccount) {
	now := time.Now()
	rand.Seed(time.Now().UnixNano())

	// Generate 6 months of bills
	bills := make([]model.Bill, 0, 6)
	for i := 0; i < 6; i++ {
		// Calculate period and due dates
		periodEnd := now.AddDate(0, -i, 0)
		periodEnd = time.Date(periodEnd.Year(), periodEnd.Month(), 1, 23, 59, 59, 999999999, periodEnd.Location())
		periodStart := periodEnd.AddDate(0, -1, 0)
		periodStart = periodStart.Add(time.Second)
		issuedDate := periodEnd.AddDate(0, 0, 5)
		dueDate := periodEnd.AddDate(0, 0, 20)

		// Generate amount based on provider type
		var amount float64
		var status string
		switch linkedAccount.ProviderID {
		case "electricity-provider":
			amount = 50.0 + rand.Float64()*100.0 // Between $50 and $150
		case "water-provider":
			amount = 30.0 + rand.Float64()*50.0 // Between $30 and $80
		case "internet-provider":
			amount = 60.0 + rand.Float64()*40.0 // Between $60 and $100
		default:
			amount = 40.0 + rand.Float64()*60.0 // Between $40 and $100
		}

		// Round to 2 decimal places
		amount = float64(int(amount*100)) / 100

		// Determine status based on due date
		if dueDate.After(now) {
			status = "unpaid"
		} else if i == 0 && rand.Float64() < 0.2 {
			// 20% chance for the most recent bill to be overdue
			status = "overdue"
		} else {
			status = "paid"
		}

		// Create paid date if paid
		var paidDate *time.Time
		if status == "paid" {
			pd := dueDate.AddDate(0, 0, -rand.Intn(10)) // Paid up to 10 days before due date
			paidDate = &pd
		}

		// Create bill
		bill := model.Bill{
			ID:              uuid.New().String(),
			UserID:          linkedAccount.UserID,
			LinkedAccountID: linkedAccount.ID,
			ProviderID:      linkedAccount.ProviderID,
			BillNumber:      fmt.Sprintf("BILL-%d-%d", i, rand.Intn(10000)),
			Amount:          amount,
			Currency:        "USD",
			Status:          status,
			DueDate:         dueDate,
			IssuedDate:      issuedDate,
			PaidDate:        paidDate,
			PeriodStart:     periodStart,
			PeriodEnd:       periodEnd,
			Details: map[string]interface{}{
				"usageAmount": amount * 0.8,
				"usageUnit":   getUsageUnit(linkedAccount.ProviderID),
				"taxAmount":   amount * 0.2,
			},
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		bills = append(bills, bill)
	}

	// Store bills
	accountKey := fmt.Sprintf("%s-%s", linkedAccount.ProviderID, linkedAccount.AccountNumber)
	c.bills[accountKey] = bills
}

// getUsageUnit returns the usage unit for a provider
func getUsageUnit(providerID string) string {
	switch providerID {
	case "electricity-provider":
		return "kWh"
	case "water-provider":
		return "gallons"
	case "internet-provider":
		return "GB"
	default:
		return "units"
	}
}

// FetchBills retrieves bills from the provider
func (c *MockProviderClient) FetchBills(ctx context.Context, linkedAccount *model.LinkedAccount) ([]model.Bill, error) {
	// Check if provider exists
	if _, exists := c.providers[linkedAccount.ProviderID]; !exists {
		return nil, fmt.Errorf("provider not found: %s", linkedAccount.ProviderID)
	}

	// Generate mock bills if they don't exist yet for this account
	accountKey := fmt.Sprintf("%s-%s", linkedAccount.ProviderID, linkedAccount.AccountNumber)
	if _, exists := c.bills[accountKey]; !exists {
		c.generateMockBills(linkedAccount)
	}

	// Return the mock bills
	bills, exists := c.bills[accountKey]
	if !exists {
		// This should ideally not happen if generateMockBills worked correctly
		return []model.Bill{}, nil
	}

	return bills, nil
}
