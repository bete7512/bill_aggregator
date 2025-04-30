// internal/core/domain/service/account.go
package service

import (
	"context"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	"github.com/bete7512/bill-aggregator/internal/core/domain/model"
	"github.com/bete7512/bill-aggregator/internal/core/port/outbound/provider"
	"github.com/bete7512/bill-aggregator/internal/core/port/outbound/repository"
)

// AccountService implements the account service interface
type AccountService struct {
	accountRepo repository.AccountRepository
	providerClient provider.ProviderClient
}

// NewAccountService creates a new account service
func NewAccountService(accountRepo repository.AccountRepository, providerClient provider.ProviderClient) *AccountService {
	return &AccountService{
		accountRepo:    accountRepo,
		providerClient: providerClient,
	}
}

// LinkAccount links a new utility provider account to a user
func (s *AccountService) LinkAccount(ctx context.Context, userID string, providerID string, accountDetails map[string]string) (*model.LinkedAccount, error) {
	// Validate provider exists
	providerInfo, err := s.providerClient.GetProviderInfo(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider info: %w", err)
	}
	
	if !providerInfo.Active {
		return nil, model.NewError("PROVIDER_INACTIVE", "Provider is currently inactive")
	}
	
	// Validate credentials with provider
	valid, err := s.providerClient.ValidateCredentials(ctx, providerID, accountDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to validate credentials: %w", err)
	}
	
	if !valid {
		return nil, model.NewError("INVALID_CREDENTIALS", "Invalid credentials for provider")
	}
	
	// Create credentials
	credential := model.Credential{
		Type:    "api_key", // Simplified for this example
		Details: accountDetails,
	}
	
	// Create linked account
	linkedAccount := &model.LinkedAccount{
		ID:            uuid.New().String(),
		UserID:        userID,
		ProviderID:    providerID,
		AccountNumber: accountDetails["account_number"],
		Credentials:   credential,
		Status:        "active",
		LastSynced:    time.Time{}, // Never synced
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	// Save linked account
	err = s.accountRepo.SaveLinkedAccount(ctx, linkedAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to save linked account: %w", err)
	}
	
	return linkedAccount, nil
}

// GetLinkedAccounts retrieves all linked accounts for a user
func (s *AccountService) GetLinkedAccounts(ctx context.Context, userID string) ([]model.LinkedAccount, error) {
	return s.accountRepo.GetLinkedAccountsByUserID(ctx, userID)
}

// GetLinkedAccount retrieves a specific linked account
func (s *AccountService) GetLinkedAccount(ctx context.Context, accountID string) (*model.LinkedAccount, error) {
	return s.accountRepo.GetLinkedAccountByID(ctx, accountID)
}

// DeleteLinkedAccount removes a linked account
func (s *AccountService) DeleteLinkedAccount(ctx context.Context, accountID string) error {
	// Get the linked account first to ensure it exists
	account, err := s.accountRepo.GetLinkedAccountByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get linked account: %w", err)
	}
	
	if account == nil {
		return model.NewError("ACCOUNT_NOT_FOUND", "Linked account not found")
	}
	
	return s.accountRepo.DeleteLinkedAccount(ctx, accountID)
}

// RefreshAccountStatus updates the status of a linked account
func (s *AccountService) RefreshAccountStatus(ctx context.Context, accountID string) error {
	// Get the linked account
	account, err := s.accountRepo.GetLinkedAccountByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get linked account: %w", err)
	}
	
	if account == nil {
		return model.NewError("ACCOUNT_NOT_FOUND", "Linked account not found")
	}
	
	// Validate credentials with provider
	valid, err := s.providerClient.ValidateCredentials(ctx, account.ProviderID, account.Credentials.Details)
	if err != nil {
		// Set status to error if we can't validate
		_ = s.accountRepo.UpdateLinkedAccountStatus(ctx, accountID, "error")
		return fmt.Errorf("failed to validate credentials: %w", err)
	}
	
	// Update status based on validation result
	status := "active"
	if !valid {
		status = "invalid_credentials"
	}
	
	return s.accountRepo.UpdateLinkedAccountStatus(ctx, accountID, status)
}
