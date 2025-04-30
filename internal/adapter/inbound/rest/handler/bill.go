// internal/adapter/inbound/rest/handler/bill.go
package handler

import (
	"net/http"
	"time"

	"github.com/bete7512/bill-aggregator/internal/adapter/inbound/rest/dto"
	"github.com/bete7512/bill-aggregator/internal/adapter/inbound/rest/middleware"
	"github.com/bete7512/bill-aggregator/internal/core/port/inbound"
	"github.com/bete7512/bill-aggregator/internal/core/port/outbound/provider"
	"github.com/bete7512/bill-aggregator/pkg/util/logger"
	"github.com/gin-gonic/gin"
)

// BillHandler handles bill-related HTTP requests
type BillHandler struct {
	billService    inbound.BillService
	providerClient provider.ProviderClient
	logger         *logger.Logger
}

// NewBillHandler creates a new bill handler
func NewBillHandler(
	billService inbound.BillService,
	providerClient provider.ProviderClient,
	logger *logger.Logger,
) *BillHandler {
	return &BillHandler{
		billService:    billService,
		providerClient: providerClient,
		logger:         logger,
	}
}

// GetAggregatedBills handles the request to get all bills for a user
func (h *BillHandler) GetAggregatedBills(c *gin.Context) {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}

	h.logger.Info("Getting aggregated bills", "userID", userID)

	aggregatedBills, err := h.billService.GetAggregatedBills(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get aggregated bills", "error", err.Error(), "userID", userID)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Convert to response DTO
	billsByProvider := make(map[string][]dto.BillResponse)
	providerSummaries := make([]dto.ProviderSummary, 0)

	// Process bills by provider
	for providerID, bills := range aggregatedBills.BillsByProvider {
		// Get provider info
		providerInfo, err := h.providerClient.GetProviderInfo(c.Request.Context(), providerID)
		providerName := "Unknown Provider"
		providerType := "Unknown"
		if err == nil && providerInfo != nil {
			providerName = providerInfo.Name
			providerType = providerInfo.Type
		} else {
			h.logger.Warn("Failed to get provider info", "error", err.Error(), "providerID", providerID)
		}

		// Convert bills to response DTOs
		billResponses := make([]dto.BillResponse, 0, len(bills))
		var providerTotal float64
		var nextDueDate *time.Time
		var unpaidCount, overdueCount int

		for _, bill := range bills {
			billResponses = append(billResponses, dto.MapToBillResponse(bill, providerName, providerType))
			providerTotal += bill.Amount

			// Update counts and next due date
			if bill.Status == "unpaid" {
				unpaidCount++
				if nextDueDate == nil || bill.DueDate.Before(*nextDueDate) {
					nextDueDate = &bill.DueDate
				}
			} else if bill.Status == "overdue" {
				overdueCount++
				if nextDueDate == nil || bill.DueDate.Before(*nextDueDate) {
					nextDueDate = &bill.DueDate
				}
			}
		}

		billsByProvider[providerID] = billResponses

		// Add provider summary
		providerSummaries = append(providerSummaries, dto.ProviderSummary{
			ProviderID:   providerID,
			ProviderName: providerName,
			ProviderType: providerType,
			TotalAmount:  providerTotal,
			Currency:     aggregatedBills.Currency,
			BillCount:    len(bills),
			UnpaidCount:  unpaidCount,
			OverdueCount: overdueCount,
			NextDueDate:  nextDueDate,
		})
	}

	response := dto.MapToAggregatedBillsResponse(aggregatedBills, billsByProvider, providerSummaries)

	c.JSON(http.StatusOK, response)
}

// GetBillsByProvider handles the request to get bills for a specific provider
func (h *BillHandler) GetBillsByProvider(c *gin.Context) {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}

	providerID := c.Param("providerID")
	if providerID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Provider ID is required"})
		return
	}

	h.logger.Info("Getting bills by provider", "userID", userID, "providerID", providerID)

	bills, err := h.billService.GetBillsByProvider(c.Request.Context(), userID, providerID)
	if err != nil {
		h.logger.Error("Failed to get bills by provider", "error", err.Error(), "userID", userID, "providerID", providerID)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Get provider info
	providerInfo, err := h.providerClient.GetProviderInfo(c.Request.Context(), providerID)
	providerName := "Unknown Provider"
	providerType := "Unknown"
	if err == nil && providerInfo != nil {
		providerName = providerInfo.Name
		providerType = providerInfo.Type
	} else {
		h.logger.Warn("Failed to get provider info", "error", err.Error(), "providerID", providerID)
	}

	// Convert to response DTOs
	billResponses := make([]dto.BillResponse, 0, len(bills))
	for _, bill := range bills {
		billResponses = append(billResponses, dto.MapToBillResponse(bill, providerName, providerType))
	}

	response := dto.BillsResponse{
		Bills: billResponses,
		Count: len(billResponses),
	}

	c.JSON(http.StatusOK, response)
}

// RefreshBills handles the request to refresh all bills for a user
func (h *BillHandler) RefreshBills(c *gin.Context) {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req dto.RefreshBillsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Non-fatal error, we can proceed with refreshing all bills
		h.logger.Warn("Failed to bind request, refreshing all bills", "error", err.Error())
	}

	// If provider ID is specified, refresh bills for that provider only
	if req.ProviderID != "" {
		h.logger.Info("Refreshing bills for provider", "userID", userID, "providerID", req.ProviderID)

		err = h.billService.RefreshBillsForProvider(c.Request.Context(), userID, req.ProviderID)
		if err != nil {
			h.logger.Error("Failed to refresh bills for provider", "error", err.Error(), "userID", userID, "providerID", req.ProviderID)
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, dto.RefreshBillsResponse{
			Success: true,
			Message: "Bills refreshed successfully for provider",
		})
		return
	}

	// Refresh all bills
	h.logger.Info("Refreshing all bills", "userID", userID)

	err = h.billService.RefreshBills(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to refresh bills", "error", err.Error(), "userID", userID)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.RefreshBillsResponse{
		Success: true,
		Message: "All bills refreshed successfully",
	})
}

// GetBillDetails handles the request to get details for a specific bill
func (h *BillHandler) GetBillDetails(c *gin.Context) {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}

	billID := c.Param("billID")
	if billID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Bill ID is required"})
		return
	}

	h.logger.Info("Getting bill details", "userID", userID, "billID", billID)

	// This endpoint would require an additional method in the BillService interface
	// Since it's not part of the original requirements, we'll just return a placeholder response
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse{Error: "This endpoint is not implemented yet"})
}

// GetUpcomingBills handles the request to get upcoming bills for a user
func (h *BillHandler) GetUpcomingBills(c *gin.Context) {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}

	h.logger.Info("Getting upcoming bills", "userID", userID)

	// This endpoint would require an additional method in the BillService interface
	// Since it's not part of the original requirements, we'll just return a placeholder response
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse{Error: "This endpoint is not implemented yet"})
}
