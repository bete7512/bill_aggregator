// internal/core/domain/service/bill.go
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bete7512/bill-aggregator/internal/core/domain/model"
	"github.com/bete7512/bill-aggregator/internal/core/port/outbound/provider"
	"github.com/bete7512/bill-aggregator/internal/core/port/outbound/repository"
)

// BillService implements the bill service interface
type BillService struct {
	billRepo       repository.BillRepository
	accountRepo    repository.AccountRepository
	providerClient provider.ProviderClient
}

// NewBillService creates a new bill service
func NewBillService(
	billRepo repository.BillRepository,
	accountRepo repository.AccountRepository,
	providerClient provider.ProviderClient,
) *BillService {
	return &BillService{
		billRepo:       billRepo,
		accountRepo:    accountRepo,
		providerClient: providerClient,
	}
}

// GetAggregatedBills retrieves all bills for a user, aggregated by provider
func (s *BillService) GetAggregatedBills(ctx context.Context, userID string) (*model.AggregatedBills, error) {
	// Get all bills for the user
	bills, err := s.billRepo.GetBillsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bills: %w", err)
	}

	// Prepare aggregated bills
	aggregated := &model.AggregatedBills{
		UserID:          userID,
		TotalAmount:     0,
		Currency:        "USD", // Simplified for this example
		BillsByProvider: make(map[string][]model.Bill),
		UnpaidCount:     0,
		OverdueCount:    0,
	}

	// Process bills
	var nextDueDate *time.Time
	for _, bill := range bills {
		// Add to provider group
		if _, exists := aggregated.BillsByProvider[bill.ProviderID]; !exists {
			aggregated.BillsByProvider[bill.ProviderID] = []model.Bill{}
		}
		aggregated.BillsByProvider[bill.ProviderID] = append(aggregated.BillsByProvider[bill.ProviderID], bill)

		// Update total amount
		aggregated.TotalAmount += bill.Amount

		// Update unpaid/overdue counts
		if bill.Status == "unpaid" {
			aggregated.UnpaidCount++

			// Update next due date
			if nextDueDate == nil || bill.DueDate.Before(*nextDueDate) {
				nextDueDate = &bill.DueDate
			}
		} else if bill.Status == "overdue" {
			aggregated.OverdueCount++

			// Overdue bills take precedence for next due date
			if nextDueDate == nil {
				nextDueDate = &bill.DueDate
			}
		}
	}

	aggregated.NextDueDate = nextDueDate

	return aggregated, nil
}

// GetBillsByProvider retrieves bills for a specific provider
func (s *BillService) GetBillsByProvider(ctx context.Context, userID string, providerID string) ([]model.Bill, error) {
	return s.billRepo.GetBillsByUserAndProvider(ctx, userID, providerID)
}

// RefreshBills triggers a refresh of all bills for a user
func (s *BillService) RefreshBills(ctx context.Context, userID string) error {
	// Get all linked accounts for the user
	accounts, err := s.accountRepo.GetLinkedAccountsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get linked accounts: %w", err)
	}

	// Create error channel and wait group for concurrent operations
	errCh := make(chan error, len(accounts))
	var wg sync.WaitGroup

	// For each linked account, fetch bills in parallel
	for _, account := range accounts {
		wg.Add(1)
		go func(acc model.LinkedAccount) {
			defer wg.Done()

			// Skip inactive accounts
			if acc.Status != "active" {
				return
			}

			// Fetch bills from provider
			bills, err := s.providerClient.FetchBills(ctx, &acc)
			if err != nil {
				errCh <- fmt.Errorf("failed to fetch bills for account %s: %w", acc.ID, err)
				return
			}

			// Save bills
			err = s.billRepo.SaveBills(ctx, bills)
			if err != nil {
				errCh <- fmt.Errorf("failed to save bills for account %s: %w", acc.ID, err)
				return
			}

			// Update last synced timestamp
			err = s.accountRepo.UpdateLastSynced(ctx, acc.ID)
			if err != nil {
				errCh <- fmt.Errorf("failed to update last synced for account %s: %w", acc.ID, err)
				return
			}
		}(account)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errCh)

	// Check for errors
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	// If there were any errors, return a combined error
	if len(errs) > 0 {
		return fmt.Errorf("failed to refresh bills for %d/%d accounts", len(errs), len(accounts))
	}

	return nil
}

// RefreshBillsForProvider triggers a refresh of bills for a specific provider
func (s *BillService) RefreshBillsForProvider(ctx context.Context, userID string, providerID string) error {
	// Get linked accounts for the user and provider
	accounts, err := s.accountRepo.GetLinkedAccountsByUserAndProvider(ctx, userID, providerID)
	if err != nil {
		return fmt.Errorf("failed to get linked accounts: %w", err)
	}

	// For each linked account, fetch bills
	for _, account := range accounts {
		// Skip inactive accounts
		if account.Status != "active" {
			continue
		}

		// Fetch bills from provider
		bills, err := s.providerClient.FetchBills(ctx, &account)
		if err != nil {
			return fmt.Errorf("failed to fetch bills: %w", err)
		}

		// Save bills
		err = s.billRepo.SaveBills(ctx, bills)
		if err != nil {
			return fmt.Errorf("failed to save bills: %w", err)
		}

		// Update last synced timestamp
		err = s.accountRepo.UpdateLastSynced(ctx, account.ID)
		if err != nil {
			return fmt.Errorf("failed to update last synced: %w", err)
		}
	}

	return nil
}
