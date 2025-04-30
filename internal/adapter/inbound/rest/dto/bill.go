// internal/adapter/inbound/rest/dto/bill.go
package dto

import (
	"time"

	"github.com/bete7512/bill-aggregator/internal/core/domain/model"
)

// BillResponse represents a bill response
type BillResponse struct {
	ID              string                 `json:"id"`
	ProviderID      string                 `json:"provider_id"`
	ProviderName    string                 `json:"provider_name"`
	ProviderType    string                 `json:"provider_type,omitempty"`
	LinkedAccountID string                 `json:"linked_account_id"`
	BillNumber      string                 `json:"bill_number"`
	Amount          float64                `json:"amount"`
	Currency        string                 `json:"currency"`
	Status          string                 `json:"status"`
	DueDate         time.Time              `json:"due_date"`
	IssuedDate      time.Time              `json:"issued_date"`
	PaidDate        *time.Time             `json:"paid_date,omitempty"`
	PeriodStart     time.Time              `json:"period_start"`
	PeriodEnd       time.Time              `json:"period_end"`
	Details         map[string]interface{} `json:"details,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// AggregatedBillsResponse represents an aggregated bills response
type AggregatedBillsResponse struct {
	TotalAmount      float64                        `json:"total_amount"`
	Currency         string                         `json:"currency"`
	UnpaidCount      int                            `json:"unpaid_count"`
	OverdueCount     int                            `json:"overdue_count"`
	NextDueDate      *time.Time                     `json:"next_due_date,omitempty"`
	BillsByProvider  map[string][]BillResponse      `json:"bills_by_provider"`
	ProviderSummary  []ProviderSummary              `json:"provider_summary"`
}

// ProviderSummary represents a summary of bills for a provider
type ProviderSummary struct {
	ProviderID    string    `json:"provider_id"`
	ProviderName  string    `json:"provider_name"`
	ProviderType  string    `json:"provider_type"`
	TotalAmount   float64   `json:"total_amount"`
	Currency      string    `json:"currency"`
	BillCount     int       `json:"bill_count"`
	UnpaidCount   int       `json:"unpaid_count"`
	OverdueCount  int       `json:"overdue_count"`
	NextDueDate   *time.Time `json:"next_due_date,omitempty"`
}

// BillsResponse represents a collection of bills
type BillsResponse struct {
	Bills []BillResponse `json:"bills"`
	Count int            `json:"count"`
}

// RefreshBillsRequest represents a request to refresh bills
type RefreshBillsRequest struct {
	ProviderID string `json:"provider_id,omitempty"`
}

// RefreshBillsResponse represents a response to a refresh bills request
type RefreshBillsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Count   int    `json:"count,omitempty"`
}

// MapToBillResponse converts a domain bill to a response DTO
func MapToBillResponse(bill model.Bill, providerName, providerType string) BillResponse {
	return BillResponse{
		ID:              bill.ID,
		ProviderID:      bill.ProviderID,
		ProviderName:    providerName,
		ProviderType:    providerType,
		LinkedAccountID: bill.LinkedAccountID,
		BillNumber:      bill.BillNumber,
		Amount:          bill.Amount,
		Currency:        bill.Currency,
		Status:          bill.Status,
		DueDate:         bill.DueDate,
		IssuedDate:      bill.IssuedDate,
		PaidDate:        bill.PaidDate,
		PeriodStart:     bill.PeriodStart,
		PeriodEnd:       bill.PeriodEnd,
		Details:         bill.Details,
		CreatedAt:       bill.CreatedAt,
		UpdatedAt:       bill.UpdatedAt,
	}
}

// MapToAggregatedBillsResponse converts a domain aggregated bills to a response DTO
func MapToAggregatedBillsResponse(
	aggregated *model.AggregatedBills, 
	billsByProvider map[string][]BillResponse,
	providerSummaries []ProviderSummary,
) AggregatedBillsResponse {
	return AggregatedBillsResponse{
		TotalAmount:     aggregated.TotalAmount,
		Currency:        aggregated.Currency,
		UnpaidCount:     aggregated.UnpaidCount,
		OverdueCount:    aggregated.OverdueCount,
		NextDueDate:     aggregated.NextDueDate,
		BillsByProvider: billsByProvider,
		ProviderSummary: providerSummaries,
	}
}