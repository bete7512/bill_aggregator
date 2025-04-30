// internal/core/domain/service/aggregation.go
package service

import (
	"context"
	"fmt"

	"github.com/bete7512/bill-aggregator/internal/core/domain/model"
	"github.com/bete7512/bill-aggregator/internal/core/port/outbound/provider"
	"github.com/bete7512/bill-aggregator/internal/core/port/outbound/repository"
)

// AggregationService implements the aggregation service interface
type AggregationService struct {
	billRepo       repository.BillRepository
	accountRepo    repository.AccountRepository
	providerClient provider.ProviderClient
}

// NewAggregationService creates a new aggregation service
func NewAggregationService(
	billRepo repository.BillRepository,
	accountRepo repository.AccountRepository,
	providerClient provider.ProviderClient,
) *AggregationService {
	return &AggregationService{
		billRepo:       billRepo,
		accountRepo:    accountRepo,
		providerClient: providerClient,
	}
}

// GetAggregatedBills retrieves aggregated bills for a user
func (s *AggregationService) GetAggregatedBills(ctx context.Context, userID string) (*model.AggregatedBills, error) {
	// Get all linked accounts for the user
	linkedAccounts, err := s.accountRepo.GetLinkedAccountsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked accounts: %w", err)
	}
	// Fetch bills for each linked account
	billsByProvider := make(map[string][]model.Bill)
	for _, linkedAccount := range linkedAccounts {
		// Fetch bills from the provider
		_, err := s.providerClient.GetProviderInfo(ctx, linkedAccount.ProviderID)
		if err != nil {
			return nil, fmt.Errorf("failed to get provider info: %w", err)
		}

		// Fetch bills from the provider
		bills, err := s.providerClient.FetchBills(ctx, &linkedAccount)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch bills: %w", err)
		}
		// Add bills to the map
		billsByProvider[linkedAccount.ProviderID] = bills
	}

	// Aggregate bills by provider
	aggregatedBills := &model.AggregatedBills{
		UserID:          userID,
		TotalAmount:     0,
		Currency:        "USD", // Default currency, should be configurable
		BillsByProvider: billsByProvider,
		UnpaidCount:     0,
		OverdueCount:    0,
		NextDueDate:     nil,
	}

	// Calculate aggregated values
	for _, bills := range billsByProvider {
		for _, bill := range bills {
			aggregatedBills.TotalAmount += bill.Amount
		}
	}

	// Calculate next due date
	aggregatedBills.NextDueDate = nil
	for _, bills := range billsByProvider {
		for _, bill := range bills {
			if aggregatedBills.NextDueDate == nil || bill.DueDate.Before(*aggregatedBills.NextDueDate) {
				aggregatedBills.NextDueDate = &bill.DueDate
			}
		}
	}

	return aggregatedBills, nil
}

// GetBillsByProvider retrieves bills for a user by provider
func (s *AggregationService) GetBillsByProvider(ctx context.Context, userID string, providerID string) ([]model.Bill, error) {
	// Get all linked accounts for the user
	linkedAccounts, err := s.accountRepo.GetLinkedAccountsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked accounts: %w", err)
	}

	// Filter bills by provider
	bills := make([]model.Bill, 0)
	for _, linkedAccount := range linkedAccounts {
		if linkedAccount.ProviderID == providerID {
			bills = append(bills, linkedAccount.Bills...)
		}
	}

	return bills, nil
}

// // RefreshBills refreshes bills for a user
// func (s *AggregationService) RefreshBills(ctx context.Context, userID string) error {
// 	// Get all linked accounts for the user
// 	linkedAccounts, err := s.accountRepo.GetLinkedAccountsByUserID(ctx, userID)
// 	if err != nil {
// 		return fmt.Errorf("failed to get linked accounts: %w", err)
// 	}

// 	// Refresh bills for each linked account
// 	for _, linkedAccount := range linkedAccounts {
// 		// Refresh bills from the provider
// 		err := s.providerClient.RefreshBills(ctx, &linkedAccount)
// 		if err != nil {
// 			return fmt.Errorf("failed to refresh bills: %w", err)
// 		}
// 	}

// 	return nil
// }

// // RefreshBillsForProvider refreshes bills for a user by provider
// func (s *AggregationService) RefreshBillsForProvider(ctx context.Context, userID string, providerID string) error {
// 	// Get all linked accounts for the user
// 	linkedAccounts, err := s.accountRepo.GetLinkedAccountsByUserID(ctx, userID)
// 	if err != nil {
// 		return fmt.Errorf("failed to get linked accounts: %w", err)
// 	}

// 	// Refresh bills for the specified provider
// 	for _, linkedAccount := range linkedAccounts {
// 		if linkedAccount.ProviderID == providerID {
// 			err := s.providerClient.RefreshBills(ctx, &linkedAccount)
// 			if err != nil {
// 				return fmt.Errorf("failed to refresh bills: %w", err)
// 			}
// 		}
// 	}

// 	return nil
// }
