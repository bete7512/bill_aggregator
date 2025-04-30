// internal/adapter/inbound/rest/handler/account.go
package handler

import (
	"net/http"

	"github.com/bete7512/bill-aggregator/internal/adapter/inbound/rest/dto"
	"github.com/bete7512/bill-aggregator/internal/adapter/inbound/rest/middleware"
	"github.com/bete7512/bill-aggregator/internal/core/domain/model"
	"github.com/bete7512/bill-aggregator/internal/core/port/inbound"
	"github.com/bete7512/bill-aggregator/internal/core/port/outbound/provider"
	"github.com/bete7512/bill-aggregator/pkg/util/logger"
	"github.com/gin-gonic/gin"
)

// AccountHandler handles account-related HTTP requests
type AccountHandler struct {
	accountService inbound.AccountService
	providerClient provider.ProviderClient
	logger         *logger.Logger
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(
	accountService inbound.AccountService,
	providerClient provider.ProviderClient,
	logger *logger.Logger,
) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		providerClient: providerClient,
		logger:         logger,
	}
}

// LinkAccount handles the request to link a new account
func (h *AccountHandler) LinkAccount(c *gin.Context) {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req dto.LinkAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", "error", err.Error())
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info("Linking account", "userID", userID, "providerID", req.ProviderID)

	linkedAccount, err := h.accountService.LinkAccount(c.Request.Context(), userID, req.ProviderID, req.AccountDetails)
	if err != nil {
		h.logger.Error("Failed to link account", "error", err.Error(), "userID", userID, "providerID", req.ProviderID)

		// Handle domain errors
		if domainErr, ok := err.(*model.Error); ok {
			statusCode := getStatusCodeForErrorCode(domainErr.Code)
			c.JSON(statusCode, dto.NewErrorResponse(err))
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Get provider info for response
	providerInfo, err := h.providerClient.GetProviderInfo(c.Request.Context(), linkedAccount.ProviderID)
	if err != nil {
		h.logger.Warn("Failed to get provider info", "error", err.Error(), "providerID", linkedAccount.ProviderID)
		// Non-fatal error, we can still return the account without provider details
		c.JSON(http.StatusOK, dto.MapToLinkedAccountResponse(linkedAccount, "Unknown Provider", "Unknown"))
		return
	}

	c.JSON(http.StatusOK, dto.MapToLinkedAccountResponse(linkedAccount, providerInfo.Name, providerInfo.Type))
}

// GetLinkedAccounts handles the request to get all linked accounts
func (h *AccountHandler) GetLinkedAccounts(c *gin.Context) {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}

	h.logger.Info("Getting linked accounts", "userID", userID)

	accounts, err := h.accountService.GetLinkedAccounts(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get linked accounts", "error", err.Error(), "userID", userID)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Convert to response DTOs
	responses := make([]dto.LinkedAccountResponse, 0, len(accounts))
	for _, account := range accounts {
		providerInfo, err := h.providerClient.GetProviderInfo(c.Request.Context(), account.ProviderID)
		providerName := "Unknown Provider"
		providerType := "Unknown"

		if err == nil && providerInfo != nil {
			providerName = providerInfo.Name
			providerType = providerInfo.Type
		} else {
			h.logger.Warn("Failed to get provider info", "error", err.Error(), "providerID", account.ProviderID)
		}

		responses = append(responses, dto.MapToLinkedAccountResponse(&account, providerName, providerType))
	}

	response := dto.LinkedAccountsResponse{
		Accounts: responses,
		Count:    len(responses),
	}

	c.JSON(http.StatusOK, response)
}

// GetLinkedAccount handles the request to get a specific linked account
func (h *AccountHandler) GetLinkedAccount(c *gin.Context) {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}

	accountID := c.Param("accountID")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Account ID is required"})
		return
	}

	h.logger.Info("Getting linked account", "userID", userID, "accountID", accountID)

	account, err := h.accountService.GetLinkedAccount(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get linked account", "error", err.Error(), "userID", userID, "accountID", accountID)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	if account == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Account not found"})
		return
	}

	// Verify the account belongs to the authenticated user
	if account.UserID != userID {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "Access denied"})
		return
	}

	// Get provider info for response
	providerInfo, err := h.providerClient.GetProviderInfo(c.Request.Context(), account.ProviderID)
	if err != nil {
		h.logger.Warn("Failed to get provider info", "error", err.Error(), "providerID", account.ProviderID)
		// Non-fatal error, we can still return the account without provider details
		c.JSON(http.StatusOK, dto.MapToLinkedAccountResponse(account, "Unknown Provider", "Unknown"))
		return
	}

	c.JSON(http.StatusOK, dto.MapToLinkedAccountResponse(account, providerInfo.Name, providerInfo.Type))
}

// DeleteLinkedAccount handles the request to delete a linked account
func (h *AccountHandler) DeleteLinkedAccount(c *gin.Context) {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}

	accountID := c.Param("accountID")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Account ID is required"})
		return
	}

	h.logger.Info("Deleting linked account", "userID", userID, "accountID", accountID)

	// Get the account first to verify ownership
	account, err := h.accountService.GetLinkedAccount(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get linked account", "error", err.Error(), "userID", userID, "accountID", accountID)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	if account == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Account not found"})
		return
	}

	// Verify the account belongs to the authenticated user
	if account.UserID != userID {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "Access denied"})
		return
	}

	// Delete the account
	err = h.accountService.DeleteLinkedAccount(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to delete linked account", "error", err.Error(), "userID", userID, "accountID", accountID)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

// RefreshAccountStatus handles the request to refresh the status of a linked account
func (h *AccountHandler) RefreshAccountStatus(c *gin.Context) {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}

	accountID := c.Param("accountID")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Account ID is required"})
		return
	}

	h.logger.Info("Refreshing account status", "userID", userID, "accountID", accountID)

	// Get the account first to verify ownership
	account, err := h.accountService.GetLinkedAccount(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get linked account", "error", err.Error(), "userID", userID, "accountID", accountID)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	if account == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Account not found"})
		return
	}

	// Verify the account belongs to the authenticated user
	if account.UserID != userID {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "Access denied"})
		return
	}

	// Refresh the account status
	err = h.accountService.RefreshAccountStatus(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to refresh account status", "error", err.Error(), "userID", userID, "accountID", accountID)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account status refreshed successfully"})
}

// getStatusCodeForErrorCode maps domain error codes to HTTP status codes
func getStatusCodeForErrorCode(code string) int {
	switch code {
	case "PROVIDER_INACTIVE":
		return http.StatusNotFound
	case "ACCESS_DENIED":
		return http.StatusForbidden
	case "PROVIDER_ERROR":
		return http.StatusBadGateway
	case "VALIDATION_ERROR":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
