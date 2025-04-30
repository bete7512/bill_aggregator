
// internal/core/port/outbound/provider/client.go
package provider

import (
	"context"
	
	"github.com/bete7512/bill-aggregator/internal/core/domain/model"
)

// ProviderClient defines the secondary port for interacting with utility providers
type ProviderClient interface {
	// Authenticate authenticates with the provider using the given credentials
	Authenticate(ctx context.Context, providerID string, credentials map[string]string) (*model.Credential, error)
	
	// FetchBills retrieves bills from the provider
	FetchBills(ctx context.Context, linkedAccount *model.LinkedAccount) ([]model.Bill, error)
	
	// ValidateCredentials validates the credentials for a provider
	ValidateCredentials(ctx context.Context, providerID string, credentials map[string]string) (bool, error)
	
	// GetProviderInfo retrieves information about a provider
	GetProviderInfo(ctx context.Context, providerID string) (*model.Provider, error)
	
	// GetAllProviders retrieves all available providers
	GetAllProviders(ctx context.Context) ([]model.Provider, error)
}