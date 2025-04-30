// internal/core/port/outbound/repository/account.go
package repository

import (
	"context"
	
	"github.com/bete7512/bill-aggregator/internal/core/domain/model"
)

// AccountRepository defines the secondary port for account persistence
type AccountRepository interface {
	// SaveLinkedAccount persists a linked account
	SaveLinkedAccount(ctx context.Context, account *model.LinkedAccount) error
	
	// GetLinkedAccountByID retrieves a linked account by ID
	GetLinkedAccountByID(ctx context.Context, accountID string) (*model.LinkedAccount, error)
	
	// GetLinkedAccountsByUserID retrieves all linked accounts for a user
	GetLinkedAccountsByUserID(ctx context.Context, userID string) ([]model.LinkedAccount, error)
	
	// GetLinkedAccountsByUserAndProvider retrieves linked accounts for a user and provider
	GetLinkedAccountsByUserAndProvider(ctx context.Context, userID string, providerID string) ([]model.LinkedAccount, error)
	
	// DeleteLinkedAccount removes a linked account
	DeleteLinkedAccount(ctx context.Context, accountID string) error
	
	// UpdateLinkedAccountStatus updates the status of a linked account
	UpdateLinkedAccountStatus(ctx context.Context, accountID string, status string) error
	
	// UpdateLastSynced updates the last synced timestamp of a linked account
	UpdateLastSynced(ctx context.Context, accountID string) error
}

