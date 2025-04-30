// internal/core/port/outbound/repository/bill.go
package repository

import (
	"context"
	"time"
	
	"github.com/bete7512/bill-aggregator/internal/core/domain/model"
)

// BillRepository defines the secondary port for bill persistence
type BillRepository interface {
	// SaveBill persists a bill
	SaveBill(ctx context.Context, bill *model.Bill) error
	
	// SaveBills persists multiple bills
	SaveBills(ctx context.Context, bills []model.Bill) error
	
	// GetBillByID retrieves a bill by ID
	GetBillByID(ctx context.Context, billID string) (*model.Bill, error)
	
	// GetBillsByUserID retrieves all bills for a user
	GetBillsByUserID(ctx context.Context, userID string) ([]model.Bill, error)
	
	// GetBillsByUserAndProvider retrieves bills for a user and provider
	GetBillsByUserAndProvider(ctx context.Context, userID string, providerID string) ([]model.Bill, error)
	
	// GetBillsByLinkedAccount retrieves bills for a linked account
	GetBillsByLinkedAccount(ctx context.Context, linkedAccountID string) ([]model.Bill, error)
	
	// GetUnpaidBills retrieves unpaid bills for a user
	GetUnpaidBills(ctx context.Context, userID string) ([]model.Bill, error)
	
	// GetBillsDueBefore retrieves bills due before a certain date
	GetBillsDueBefore(ctx context.Context, userID string, date time.Time) ([]model.Bill, error)
	
	// UpdateBillStatus updates the status of a bill
	UpdateBillStatus(ctx context.Context, billID string, status string) error
	
	// DeleteBillsByLinkedAccount removes all bills for a linked account
	DeleteBillsByLinkedAccount(ctx context.Context, linkedAccountID string) error
}

