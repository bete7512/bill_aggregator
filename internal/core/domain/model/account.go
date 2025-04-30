// internal/core/domain/model/account.go
package model

import (
	"time"
)

// User represents a user of the bill aggregation service
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LinkedAccount represents a user's account with a utility provider
type LinkedAccount struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	ProviderID    string     `json:"provider_id"`
	AccountNumber string     `json:"account_number"`
	Credentials   Credential `json:"credentials"`
	Status        string     `json:"status"` // active, inactive, error
	LastSynced    time.Time  `json:"last_synced"`
	Bills         []Bill     `json:"bills"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Credential stores authentication information for provider APIs
// This is a simplified version - in a real system, credentials should be encrypted
type Credential struct {
	Type      string            `json:"type"` // token, api_key, oauth, etc.
	Details   map[string]string `json:"details"`
	ExpiresAt *time.Time        `json:"expires_at,omitempty"`
	IsExpired bool              `json:"is_expired"`
}
