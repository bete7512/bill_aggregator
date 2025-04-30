
// internal/adapter/outbound/repository/postgres/bill.go
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bete7512/bill-aggregator/internal/core/domain/model"
)

// BillRepository implements the bill repository interface using PostgreSQL
type BillRepository struct {
	db *sql.DB
}

// NewBillRepository creates a new bill repository
func NewBillRepository(db *sql.DB) *BillRepository {
	return &BillRepository{
		db: db,
	}
}

// SaveBill persists a bill
func (r *BillRepository) SaveBill(ctx context.Context, bill *model.Bill) error {
	// Convert details to JSON for storage
	detailsJSON, err := json.Marshal(bill.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	query := `
		INSERT INTO bills (
			id, user_id, linked_account_id, provider_id, bill_number, 
			amount, currency, status, due_date, issued_date, 
			paid_date, period_start, period_end, details, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
		ON CONFLICT (id) DO UPDATE
		SET 
			amount = EXCLUDED.amount,
			currency = EXCLUDED.currency,
			status = EXCLUDED.status,
			due_date = EXCLUDED.due_date,
			issued_date = EXCLUDED.issued_date,
			paid_date = EXCLUDED.paid_date,
			period_start = EXCLUDED.period_start,
			period_end = EXCLUDED.period_end,
			details = EXCLUDED.details,
			updated_at = EXCLUDED.updated_at
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		bill.ID,
		bill.UserID,
		bill.LinkedAccountID,
		bill.ProviderID,
		bill.BillNumber,
		bill.Amount,
		bill.Currency,
		bill.Status,
		bill.DueDate,
		bill.IssuedDate,
		bill.PaidDate,
		bill.PeriodStart,
		bill.PeriodEnd,
		detailsJSON,
		bill.CreatedAt,
		bill.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save bill: %w", err)
	}

	return nil
}

// SaveBills persists multiple bills
func (r *BillRepository) SaveBills(ctx context.Context, bills []model.Bill) error {
	// Use a transaction for batch insertion
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	query := `
		INSERT INTO bills (
			id, user_id, linked_account_id, provider_id, bill_number, 
			amount, currency, status, due_date, issued_date, 
			paid_date, period_start, period_end, details, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
		ON CONFLICT (id) DO UPDATE
		SET 
			amount = EXCLUDED.amount,
			currency = EXCLUDED.currency,
			status = EXCLUDED.status,
			due_date = EXCLUDED.due_date,
			issued_date = EXCLUDED.issued_date,
			paid_date = EXCLUDED.paid_date,
			period_start = EXCLUDED.period_start,
			period_end = EXCLUDED.period_end,
			details = EXCLUDED.details,
			updated_at = EXCLUDED.updated_at
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, bill := range bills {
		// Convert details to JSON for storage
		detailsJSON, err := json.Marshal(bill.Details)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to marshal details: %w", err)
		}

		_, err = stmt.ExecContext(
			ctx,
			bill.ID,
			bill.UserID,
			bill.LinkedAccountID,
			bill.ProviderID,
			bill.BillNumber,
			bill.Amount,
			bill.Currency,
			bill.Status,
			bill.DueDate,
			bill.IssuedDate,
			bill.PaidDate,
			bill.PeriodStart,
			bill.PeriodEnd,
			detailsJSON,
			bill.CreatedAt,
			bill.UpdatedAt,
		)

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to save bill: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetBillByID retrieves a bill by ID
func (r *BillRepository) GetBillByID(ctx context.Context, billID string) (*model.Bill, error) {
	query := `
		SELECT 
			id, user_id, linked_account_id, provider_id, bill_number, 
			amount, currency, status, due_date, issued_date, 
			paid_date, period_start, period_end, details, created_at, updated_at
		FROM 
			bills
		WHERE 
			id = $1
	`

	var (
		bill        model.Bill
		detailsJSON []byte
		paidDate    sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, billID).Scan(
		&bill.ID,
		&bill.UserID,
		&bill.LinkedAccountID,
		&bill.ProviderID,
		&bill.BillNumber,
		&bill.Amount,
		&bill.Currency,
		&bill.Status,
		&bill.DueDate,
		&bill.IssuedDate,
		&paidDate,
		&bill.PeriodStart,
		&bill.PeriodEnd,
		&detailsJSON,
		&bill.CreatedAt,
		&bill.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No bill found
		}
		return nil, fmt.Errorf("failed to get bill: %w", err)
	}

	// Handle nullable paid date
	if paidDate.Valid {
		bill.PaidDate = &paidDate.Time
	}

	// Unmarshal details
	err = json.Unmarshal(detailsJSON, &bill.Details)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal details: %w", err)
	}

	return &bill, nil
}

// GetBillsByUserID retrieves all bills for a user
func (r *BillRepository) GetBillsByUserID(ctx context.Context, userID string) ([]model.Bill, error) {
	query := `
		SELECT 
			id, user_id, linked_account_id, provider_id, bill_number, 
			amount, currency, status, due_date, issued_date, 
			paid_date, period_start, period_end, details, created_at, updated_at
		FROM 
			bills
		WHERE 
			user_id = $1
		ORDER BY
			due_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query bills: %w", err)
	}
	defer rows.Close()

	var bills []model.Bill

	for rows.Next() {
		var (
			bill        model.Bill
			detailsJSON []byte
			paidDate    sql.NullTime
		)

		err := rows.Scan(
			&bill.ID,
			&bill.UserID,
			&bill.LinkedAccountID,
			&bill.ProviderID,
			&bill.BillNumber,
			&bill.Amount,
			&bill.Currency,
			&bill.Status,
			&bill.DueDate,
			&bill.IssuedDate,
			&paidDate,
			&bill.PeriodStart,
			&bill.PeriodEnd,
			&detailsJSON,
			&bill.CreatedAt,
			&bill.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan bill: %w", err)
		}

		// Handle nullable paid date
		if paidDate.Valid {
			bill.PaidDate = &paidDate.Time
		}

		// Unmarshal details
		err = json.Unmarshal(detailsJSON, &bill.Details)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal details: %w", err)
		}

		bills = append(bills, bill)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bills: %w", err)
	}

	return bills, nil
}

// GetBillsByUserAndProvider retrieves bills for a user and provider
func (r *BillRepository) GetBillsByUserAndProvider(ctx context.Context, userID string, providerID string) ([]model.Bill, error) {
	query := `
		SELECT 
			id, user_id, linked_account_id, provider_id, bill_number, 
			amount, currency, status, due_date, issued_date, 
			paid_date, period_start, period_end, details, created_at, updated_at
		FROM 
			bills
		WHERE 
			user_id = $1 AND provider_id = $2
		ORDER BY
			due_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query bills: %w", err)
	}
	defer rows.Close()

	var bills []model.Bill

	for rows.Next() {
		var (
			bill        model.Bill
			detailsJSON []byte
			paidDate    sql.NullTime
		)

		err := rows.Scan(
			&bill.ID,
			&bill.UserID,
			&bill.LinkedAccountID,
			&bill.ProviderID,
			&bill.BillNumber,
			&bill.Amount,
			&bill.Currency,
			&bill.Status,
			&bill.DueDate,
			&bill.IssuedDate,
			&paidDate,
			&bill.PeriodStart,
			&bill.PeriodEnd,
			&detailsJSON,
			&bill.CreatedAt,
			&bill.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan bill: %w", err)
		}

		// Handle nullable paid date
		if paidDate.Valid {
			bill.PaidDate = &paidDate.Time
		}

		// Unmarshal details
		err = json.Unmarshal(detailsJSON, &bill.Details)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal details: %w", err)
		}

		bills = append(bills, bill)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bills: %w", err)
	}

	return bills, nil
}

// GetBillsByLinkedAccount retrieves bills for a linked account
func (r *BillRepository) GetBillsByLinkedAccount(ctx context.Context, linkedAccountID string) ([]model.Bill, error) {
	query := `
		SELECT 
			id, user_id, linked_account_id, provider_id, bill_number, 
			amount, currency, status, due_date, issued_date, 
			paid_date, period_start, period_end, details, created_at, updated_at
		FROM 
			bills
		WHERE 
			linked_account_id = $1
		ORDER BY
			due_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, linkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query bills: %w", err)
	}
	defer rows.Close()

	var bills []model.Bill

	for rows.Next() {
		var (
			bill        model.Bill
			detailsJSON []byte
			paidDate    sql.NullTime
		)

		err := rows.Scan(
			&bill.ID,
			&bill.UserID,
			&bill.LinkedAccountID,
			&bill.ProviderID,
			&bill.BillNumber,
			&bill.Amount,
			&bill.Currency,
			&bill.Status,
			&bill.DueDate,
			&bill.IssuedDate,
			&paidDate,
			&bill.PeriodStart,
			&bill.PeriodEnd,
			&detailsJSON,
			&bill.CreatedAt,
			&bill.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan bill: %w", err)
		}

		// Handle nullable paid date
		if paidDate.Valid {
			bill.PaidDate = &paidDate.Time
		}

		// Unmarshal details
		err = json.Unmarshal(detailsJSON, &bill.Details)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal details: %w", err)
		}

		bills = append(bills, bill)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bills: %w", err)
	}

	return bills, nil
}

// GetUnpaidBills retrieves unpaid bills for a user
func (r *BillRepository) GetUnpaidBills(ctx context.Context, userID string) ([]model.Bill, error) {
	query := `
		SELECT 
			id, user_id, linked_account_id, provider_id, bill_number, 
			amount, currency, status, due_date, issued_date, 
			paid_date, period_start, period_end, details, created_at, updated_at
		FROM 
			bills
		WHERE 
			user_id = $1 AND status IN ('unpaid', 'overdue')
		ORDER BY
			due_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query unpaid bills: %w", err)
	}
	defer rows.Close()

	var bills []model.Bill

	for rows.Next() {
		var (
			bill        model.Bill
			detailsJSON []byte
			paidDate    sql.NullTime
		)

		err := rows.Scan(
			&bill.ID,
			&bill.UserID,
			&bill.LinkedAccountID,
			&bill.ProviderID,
			&bill.BillNumber,
			&bill.Amount,
			&bill.Currency,
			&bill.Status,
			&bill.DueDate,
			&bill.IssuedDate,
			&paidDate,
			&bill.PeriodStart,
			&bill.PeriodEnd,
			&detailsJSON,
			&bill.CreatedAt,
			&bill.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan bill: %w", err)
		}

		// Handle nullable paid date
		if paidDate.Valid {
			bill.PaidDate = &paidDate.Time
		}

		// Unmarshal details
		err = json.Unmarshal(detailsJSON, &bill.Details)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal details: %w", err)
		}

		bills = append(bills, bill)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bills: %w", err)
	}

	return bills, nil
}

// GetBillsDueBefore retrieves bills due before a certain date
func (r *BillRepository) GetBillsDueBefore(ctx context.Context, userID string, date time.Time) ([]model.Bill, error) {
	query := `
		SELECT 
			id, user_id, linked_account_id, provider_id, bill_number, 
			amount, currency, status, due_date, issued_date, 
			paid_date, period_start, period_end, details, created_at, updated_at
		FROM 
			bills
		WHERE 
			user_id = $1 AND due_date <= $2 AND status IN ('unpaid', 'overdue')
		ORDER BY
			due_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to query bills due before: %w", err)
	}
	defer rows.Close()

	var bills []model.Bill

	for rows.Next() {
		var (
			bill        model.Bill
			detailsJSON []byte
			paidDate    sql.NullTime
		)

		err := rows.Scan(
			&bill.ID,
			&bill.UserID,
			&bill.LinkedAccountID,
			&bill.ProviderID,
			&bill.BillNumber,
			&bill.Amount,
			&bill.Currency,
			&bill.Status,
			&bill.DueDate,
			&bill.IssuedDate,
			&paidDate,
			&bill.PeriodStart,
			&bill.PeriodEnd,
			&detailsJSON,
			&bill.CreatedAt,
			&bill.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan bill: %w", err)
		}

		// Handle nullable paid date
		if paidDate.Valid {
			bill.PaidDate = &paidDate.Time
		}

		// Unmarshal details
		err = json.Unmarshal(detailsJSON, &bill.Details)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal details: %w", err)
		}

		bills = append(bills, bill)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bills: %w", err)
	}

	return bills, nil
}

// UpdateBillStatus updates the status of a bill
func (r *BillRepository) UpdateBillStatus(ctx context.Context, billID string, status string) error {
	query := `
		UPDATE bills
		SET 
			status = $1,
			paid_date = CASE WHEN $1 = 'paid' THEN NOW() ELSE NULL END,
			updated_at = NOW()
		WHERE 
			id = $2
	`

	_, err := r.db.ExecContext(ctx, query, status, billID)
	if err != nil {
		return fmt.Errorf("failed to update bill status: %w", err)
	}

	return nil
}

// DeleteBillsByLinkedAccount removes all bills for a linked account
func (r *BillRepository) DeleteBillsByLinkedAccount(ctx context.Context, linkedAccountID string) error {
	query := `DELETE FROM bills WHERE linked_account_id = $1`

	_, err := r.db.ExecContext(ctx, query, linkedAccountID)
	if err != nil {
		return fmt.Errorf("failed to delete bills: %w", err)
	}

	return nil
}