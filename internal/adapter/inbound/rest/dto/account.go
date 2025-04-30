// internal/adapter/inbound/rest/dto/account.go
package dto

import (
	"time"

	"github.com/bete7512/bill-aggregator/internal/core/domain/model"
)

// LinkAccountRequest represents the request to link an account
type LinkAccountRequest struct {
	ProviderID     string            `json:"provider_id" validate:"required"`
	AccountDetails map[string]string `json:"account_details" validate:"required"`
}

// LinkedAccountResponse represents a linked account response
type LinkedAccountResponse struct {
	ID             string    `json:"id"`
	ProviderID     string    `json:"provider_id"`
	ProviderName   string    `json:"provider_name"`
	ProviderType   string    `json:"provider_type,omitempty"`
	AccountNumber  string    `json:"account_number"`
	Status         string    `json:"status"`
	LastSynced     time.Time `json:"last_synced"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// LinkedAccountsResponse represents a collection of linked accounts
type LinkedAccountsResponse struct {
	Accounts []LinkedAccountResponse `json:"accounts"`
	Count    int                     `json:"count"`
}

// MapToLinkedAccountResponse converts a domain model to a response DTO
func MapToLinkedAccountResponse(account *model.LinkedAccount, providerName, providerType string) LinkedAccountResponse {
	return LinkedAccountResponse{
		ID:            account.ID,
		ProviderID:    account.ProviderID,
		ProviderName:  providerName,
		ProviderType:  providerType,
		AccountNumber: account.AccountNumber,
		Status:        account.Status,
		LastSynced:    account.LastSynced,
		CreatedAt:     account.CreatedAt,
		UpdatedAt:     account.UpdatedAt,
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response from a domain error
func NewErrorResponse(err error) ErrorResponse {
	// Check if it's a domain error
	if domainErr, ok := err.(*model.Error); ok {
		return ErrorResponse{
			Error:   domainErr.Message,
			Code:    domainErr.Code,
			Details: domainErr.Details,
		}
	}

	// Generic error
	return ErrorResponse{
		Error: err.Error(),
	}
}