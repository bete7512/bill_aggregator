package model

// Provider represents a utility provider
type Provider struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Type           string   `json:"type"` // electricity, water, internet, etc.
	AuthTypes      []string `json:"auth_types"`
	RequiredFields []string `json:"required_fields"`
	BaseURL        string   `json:"base_url"`
	IconURL        string   `json:"icon_url"`
	Active         bool     `json:"active"`
}
