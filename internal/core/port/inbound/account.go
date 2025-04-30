// internal/core/port/inbound/account.go
package inbound

import (
	"context"
	
	"github.com/bete7512/bill-aggregator/internal/core/domain/model"
)

// AccountService defines the primary port for account operations
type AccountService interface {
	// LinkAccount links a new utility provider account to a user
	LinkAccount(ctx context.Context, userID string, providerID string, accountDetails map[string]string) (*model.LinkedAccount, error)
	
	// GetLinkedAccounts retrieves all linked accounts for a user
	GetLinkedAccounts(ctx context.Context, userID string) ([]model.LinkedAccount, error)
	
	// GetLinkedAccount retrieves a specific linked account
	GetLinkedAccount(ctx context.Context, accountID string) (*model.LinkedAccount, error)
	
	// DeleteLinkedAccount removes a linked account
	DeleteLinkedAccount(ctx context.Context, accountID string) error
	
	// RefreshAccountStatus updates the status of a linked account
	RefreshAccountStatus(ctx context.Context, accountID string) error
}

