package model

import "time"

// Bill represents a bill from a utility provider
type Bill struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	LinkedAccountID string                 `json:"linked_account_id"`
	ProviderID      string                 `json:"provider_id"`
	BillNumber      string                 `json:"bill_number"`
	Amount          float64                `json:"amount"`
	Currency        string                 `json:"currency"`
	Status          string                 `json:"status"` // paid, unpaid, overdue
	DueDate         time.Time              `json:"due_date"`
	IssuedDate      time.Time              `json:"issued_date"`
	PaidDate        *time.Time             `json:"paid_date,omitempty"`
	PeriodStart     time.Time              `json:"period_start"`
	PeriodEnd       time.Time              `json:"period_end"`
	Details         map[string]interface{} `json:"details"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// AggregatedBills represents a collection of bills grouped by provider
type AggregatedBills struct {
	UserID          string            `json:"user_id"`
	TotalAmount     float64           `json:"total_amount"`
	Currency        string            `json:"currency"` // Assuming same currency for simplicity
	BillsByProvider map[string][]Bill `json:"bills_by_provider"`
	UnpaidCount     int               `json:"unpaid_count"`
	OverdueCount    int               `json:"overdue_count"`
	NextDueDate     *time.Time        `json:"next_due_date,omitempty"`
}

// Error represents a domain error
type Error struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

// NewError creates a new domain error
func NewError(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithDetails adds details to an error
func (e *Error) WithDetails(details map[string]interface{}) *Error {
	e.Details = details
	return e
}
