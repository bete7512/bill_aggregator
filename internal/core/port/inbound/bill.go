// internal/core/port/inbound/bill.go
package inbound

import (
	"context"
	
	"github.com/bete7512/bill-aggregator/internal/core/domain/model"
)

// BillService defines the primary port for bill operations
type BillService interface {
	// GetAggregatedBills retrieves all bills for a user, aggregated by provider
	GetAggregatedBills(ctx context.Context, userID string) (*model.AggregatedBills, error)
	
	// GetBillsByProvider retrieves bills for a specific provider
	GetBillsByProvider(ctx context.Context, userID string, providerID string) ([]model.Bill, error)
	
	// RefreshBills triggers a refresh of all bills for a user
	RefreshBills(ctx context.Context, userID string) error
	
	// RefreshBillsForProvider triggers a refresh of bills for a specific provider
	RefreshBillsForProvider(ctx context.Context, userID string, providerID string) error
}

