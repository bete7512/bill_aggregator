// internal/adapter/outbound/repository/postgres/account.go
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

// AccountRepository implements the account repository interface using PostgreSQL
type AccountRepository struct {
	db *sql.DB
}

// NewAccountRepository creates a new account repository
func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{
		db: db,
	}
}

// SaveLinkedAccount persists a linked account
func (r *AccountRepository) SaveLinkedAccount(ctx context.Context, account *model.LinkedAccount) error {
	// Convert credentials to JSON for storage
	credentialsJSON, err := json.Marshal(account.Credentials)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	query := `
		INSERT INTO linked_accounts (
			id, user_id, provider_id, account_number, credentials, status, last_synced, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		ON CONFLICT (id) DO UPDATE
		SET 
			user_id = EXCLUDED.user_id,
			provider_id = EXCLUDED.provider_id,
			account_number = EXCLUDED.account_number,
			credentials = EXCLUDED.credentials,
			status = EXCLUDED.status,
			last_synced = EXCLUDED.last_synced,
			updated_at = EXCLUDED.updated_at
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		account.ID,
		account.UserID,
		account.ProviderID,
		account.AccountNumber,
		credentialsJSON,
		account.Status,
		account.LastSynced,
		account.CreatedAt,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save linked account: %w", err)
	}

	return nil
}

// GetLinkedAccountByID retrieves a linked account by ID
func (r *AccountRepository) GetLinkedAccountByID(ctx context.Context, accountID string) (*model.LinkedAccount, error) {
	query := `
		SELECT 
			id, user_id, provider_id, account_number, credentials, status, last_synced, created_at, updated_at
		FROM 
			linked_accounts
		WHERE 
			id = $1
	`

	var (
		account        model.LinkedAccount
		credentialsJSON []byte
	)

	err := r.db.QueryRowContext(ctx, query, accountID).Scan(
		&account.ID,
		&account.UserID,
		&account.ProviderID,
		&account.AccountNumber,
		&credentialsJSON,
		&account.Status,
		&account.LastSynced,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No account found
		}
		return nil, fmt.Errorf("failed to get linked account: %w", err)
	}

	// Unmarshal credentials
	err = json.Unmarshal(credentialsJSON, &account.Credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return &account, nil
}

// GetLinkedAccountsByUserID retrieves all linked accounts for a user
func (r *AccountRepository) GetLinkedAccountsByUserID(ctx context.Context, userID string) ([]model.LinkedAccount, error) {
	query := `
		SELECT 
			id, user_id, provider_id, account_number, credentials, status, last_synced, created_at, updated_at
		FROM 
			linked_accounts
		WHERE 
			user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query linked accounts: %w", err)
	}
	defer rows.Close()

	var accounts []model.LinkedAccount

	for rows.Next() {
		var (
			account        model.LinkedAccount
			credentialsJSON []byte
		)

		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.ProviderID,
			&account.AccountNumber,
			&credentialsJSON,
			&account.Status,
			&account.LastSynced,
			&account.CreatedAt,
			&account.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan linked account: %w", err)
		}

		// Unmarshal credentials
		err = json.Unmarshal(credentialsJSON, &account.Credentials)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
		}

		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating linked accounts: %w", err)
	}

	return accounts, nil
}

// GetLinkedAccountsByUserAndProvider retrieves linked accounts for a user and provider
func (r *AccountRepository) GetLinkedAccountsByUserAndProvider(ctx context.Context, userID string, providerID string) ([]model.LinkedAccount, error) {
	query := `
		SELECT 
			id, user_id, provider_id, account_number, credentials, status, last_synced, created_at, updated_at
		FROM 
			linked_accounts
		WHERE 
			user_id = $1 AND provider_id = $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query linked accounts: %w", err)
	}
	defer rows.Close()

	var accounts []model.LinkedAccount

	for rows.Next() {
		var (
			account        model.LinkedAccount
			credentialsJSON []byte
		)

		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.ProviderID,
			&account.AccountNumber,
			&credentialsJSON,
			&account.Status,
			&account.LastSynced,
			&account.CreatedAt,
			&account.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan linked account: %w", err)
		}

		// Unmarshal credentials
		err = json.Unmarshal(credentialsJSON, &account.Credentials)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
		}

		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating linked accounts: %w", err)
	}

	return accounts, nil
}

// DeleteLinkedAccount removes a linked account
func (r *AccountRepository) DeleteLinkedAccount(ctx context.Context, accountID string) error {
	// Use a transaction to ensure both account and bills are deleted atomically
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Delete bills first (foreign key constraint)
	_, err = tx.ExecContext(ctx, "DELETE FROM bills WHERE linked_account_id = $1", accountID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete bills: %w", err)
	}

	// Delete the account
	_, err = tx.ExecContext(ctx, "DELETE FROM linked_accounts WHERE id = $1", accountID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete linked account: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateLinkedAccountStatus updates the status of a linked account
func (r *AccountRepository) UpdateLinkedAccountStatus(ctx context.Context, accountID string, status string) error {
	query := `
		UPDATE linked_accounts
		SET 
			status = $1,
			updated_at = $2
		WHERE 
			id = $3
	`

	_, err := r.db.ExecContext(ctx, query, status, time.Now(), accountID)
	if err != nil {
		return fmt.Errorf("failed to update linked account status: %w", err)
	}

	return nil
}

// UpdateLastSynced updates the last synced timestamp of a linked account
func (r *AccountRepository) UpdateLastSynced(ctx context.Context, accountID string) error {
	query := `
		UPDATE linked_accounts
		SET 
			last_synced = $1,
			updated_at = $2
		WHERE 
			id = $3
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, now, accountID)
	if err != nil {
		return fmt.Errorf("failed to update last synced: %w", err)
	}

	return nil
}
