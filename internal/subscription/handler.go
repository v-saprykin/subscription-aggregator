package subscription

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) RegisterRoutes(api *gin.RouterGroup) {
	routes := api.Group("/subscriptions")
	routes.POST("", h.create)
	routes.GET("", h.list)
	routes.GET("/total", h.total)
	routes.GET("/:id", h.get)
	routes.PUT("/:id", h.update)
	routes.DELETE("/:id", h.delete)
}

// create godoc
// @Summary Create a subscription
// @Description Creates a subscription with a monthly price in integer rubles. Dates use MM-YYYY.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body upsertSubscriptionRequest true "Subscription data"
// @Success 201 {object} subscriptionResponse
// @Failure 400 {object} errorResponse "Invalid request body or field value"
// @Failure 500 {object} errorResponse "Unexpected server error"
// @Router /api/v1/subscriptions [post]
func (h *Handler) create(c *gin.Context) {
	input, ok := h.readUpsertInput(c)
	if !ok {
		return
	}

	item, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		h.writeServiceError(c, "create", err)
		return
	}

	c.JSON(http.StatusCreated, toSubscriptionResponse(item))
}

// get godoc
// @Summary Get a subscription
// @Description Returns one subscription by its UUID.
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription UUID" format(uuid)
// @Success 200 {object} subscriptionResponse
// @Failure 400 {object} errorResponse "Invalid subscription UUID"
// @Failure 404 {object} errorResponse "Subscription not found"
// @Failure 500 {object} errorResponse "Unexpected server error"
// @Router /api/v1/subscriptions/{id} [get]
func (h *Handler) get(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	item, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		h.writeServiceError(c, "get", err)
		return
	}

	c.JSON(http.StatusOK, toSubscriptionResponse(item))
}

// list godoc
// @Summary List subscriptions
// @Description Returns subscriptions using the current offset-based list behavior and optional exact-match filters.
// @Tags subscriptions
// @Produce json
// @Param limit query int false "Maximum number of subscriptions to return" default(20) minimum(1) maximum(100)
// @Param offset query int false "Number of subscriptions to skip" default(0) minimum(0)
// @Param user_id query string false "User UUID filter" format(uuid)
// @Param service_name query string false "Exact service name filter"
// @Success 200 {object} listSubscriptionsResponse
// @Failure 400 {object} errorResponse "Invalid query parameter"
// @Failure 500 {object} errorResponse "Unexpected server error"
// @Router /api/v1/subscriptions [get]
func (h *Handler) list(c *gin.Context) {
	filter, err := parseListSubscriptionsFilter(c.Request.URL.Query())
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	items, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		h.writeServiceError(c, "list", err)
		return
	}

	c.JSON(http.StatusOK, listSubscriptionsResponse{
		Items:  toSubscriptionResponses(items),
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
}

// total godoc
// @Summary Calculate total subscription cost
// @Description Calculates the total cost in integer rubles for all overlapping months in the inclusive requested period.
// @Tags subscriptions
// @Produce json
// @Param from query string true "Inclusive period start in MM-YYYY format" example(07-2025)
// @Param to query string true "Inclusive period end in MM-YYYY format" example(12-2025)
// @Param user_id query string false "User UUID filter" format(uuid)
// @Param service_name query string false "Exact service name filter"
// @Success 200 {object} totalPriceResponse
// @Failure 400 {object} errorResponse "Invalid or missing query parameter"
// @Failure 500 {object} errorResponse "Unexpected server error"
// @Router /api/v1/subscriptions/total [get]
func (h *Handler) total(c *gin.Context) {
	filter, err := parseTotalPriceFilter(c.Request.URL.Query())
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	totalPrice, err := h.service.CalculateTotalPrice(c.Request.Context(), filter)
	if err != nil {
		h.writeServiceError(c, "calculate_total_price", err)
		return
	}

	c.JSON(http.StatusOK, toTotalPriceResponse(totalPrice, filter))
}

// update godoc
// @Summary Update a subscription
// @Description Replaces all mutable subscription fields. Dates use MM-YYYY.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription UUID" format(uuid)
// @Param subscription body upsertSubscriptionRequest true "Replacement subscription data"
// @Success 200 {object} subscriptionResponse
// @Failure 400 {object} errorResponse "Invalid subscription UUID, request body, or field value"
// @Failure 404 {object} errorResponse "Subscription not found"
// @Failure 500 {object} errorResponse "Unexpected server error"
// @Router /api/v1/subscriptions/{id} [put]
func (h *Handler) update(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	input, ok := h.readUpsertInput(c)
	if !ok {
		return
	}

	item, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		h.writeServiceError(c, "update", err)
		return
	}

	c.JSON(http.StatusOK, toSubscriptionResponse(item))
}

// delete godoc
// @Summary Delete a subscription
// @Description Deletes one subscription by its UUID.
// @Tags subscriptions
// @Param id path string true "Subscription UUID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} errorResponse "Invalid subscription UUID"
// @Failure 404 {object} errorResponse "Subscription not found"
// @Failure 500 {object} errorResponse "Unexpected server error"
// @Router /api/v1/subscriptions/{id} [delete]
func (h *Handler) delete(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.writeServiceError(c, "delete", err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) readUpsertInput(c *gin.Context) (UpsertSubscription, bool) {
	var req upsertSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return UpsertSubscription{}, false
	}

	input, err := validateUpsertSubscriptionRequest(req)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		return UpsertSubscription{}, false
	}

	return input, true
}

func (h *Handler) writeServiceError(c *gin.Context, operation string, err error) {
	if errors.Is(err, ErrNotFound) {
		writeError(c, http.StatusNotFound, "not_found", "subscription not found")
		return
	}

	h.logger.Error("subscription request failed", "operation", operation, "error", err)
	writeError(c, http.StatusInternalServerError, "internal_error", "unexpected server error")
}

func parseIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	rawID := strings.TrimSpace(c.Param(name))
	id, err := uuid.Parse(rawID)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", name+" must be UUID")
		return uuid.Nil, false
	}

	return id, true
}

type errorResponse struct {
	// Error is a stable machine-readable error code.
	Error string `json:"error" example:"validation_error"`
	// Message describes the error for API clients.
	Message string `json:"message" example:"start_date must use MM-YYYY format"`
}

type subscriptionResponse struct {
	// ID is the subscription UUID.
	ID string `json:"id" format:"uuid" example:"2f8d9b27-5b9e-4d6f-83ef-cf4fef0c9fc2"`
	// ServiceName is the subscription service name.
	ServiceName string `json:"service_name" example:"Yandex Plus"`
	// Price is the monthly price in integer rubles.
	Price int32 `json:"price" example:"400"`
	// UserID is the UUID of the user who owns the subscription.
	UserID string `json:"user_id" format:"uuid" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	// StartDate is the first active month in MM-YYYY format.
	StartDate string `json:"start_date" example:"07-2025"`
	// EndDate is the optional last active month in MM-YYYY format.
	EndDate *string `json:"end_date" extensions:"x-nullable" example:"12-2025"`
	// CreatedAt is the record creation time in RFC 3339 format.
	CreatedAt string `json:"created_at" format:"date-time" example:"2026-06-23T10:00:00Z"`
	// UpdatedAt is the record update time in RFC 3339 format.
	UpdatedAt string `json:"updated_at" format:"date-time" example:"2026-06-23T10:00:00Z"`
}

type listSubscriptionsResponse struct {
	// Items contains the subscriptions in the current result page.
	Items []subscriptionResponse `json:"items"`
	// Limit is the requested maximum number of returned subscriptions.
	Limit int32 `json:"limit"`
	// Offset is the requested number of skipped subscriptions.
	Offset int32 `json:"offset"`
}

type totalPriceResponse struct {
	// TotalPrice is the aggregate cost in integer rubles.
	TotalPrice int64 `json:"total_price" example:"2400"`
	// PeriodFrom is the inclusive period start in MM-YYYY format.
	PeriodFrom string `json:"period_from" example:"07-2025"`
	// PeriodTo is the inclusive period end in MM-YYYY format.
	PeriodTo string `json:"period_to" example:"12-2025"`
	// UserID is the applied user UUID filter, or null when omitted.
	UserID *string `json:"user_id" format:"uuid" extensions:"x-nullable" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	// ServiceName is the applied exact service filter, or null when omitted.
	ServiceName *string `json:"service_name" extensions:"x-nullable" example:"Yandex Plus"`
}

func writeError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, errorResponse{
		Error:   code,
		Message: message,
	})
}

func toSubscriptionResponses(items []Subscription) []subscriptionResponse {
	responses := make([]subscriptionResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toSubscriptionResponse(item))
	}

	return responses
}

func toTotalPriceResponse(totalPrice int64, filter TotalPriceFilter) totalPriceResponse {
	var userID *string
	if filter.UserID != nil {
		formatted := filter.UserID.String()
		userID = &formatted
	}

	var serviceName *string
	if filter.ServiceName != nil {
		formatted := *filter.ServiceName
		serviceName = &formatted
	}

	return totalPriceResponse{
		TotalPrice:  totalPrice,
		PeriodFrom:  formatAPIMonth(filter.PeriodFrom),
		PeriodTo:    formatAPIMonth(filter.PeriodTo),
		UserID:      userID,
		ServiceName: serviceName,
	}
}

func toSubscriptionResponse(item Subscription) subscriptionResponse {
	var endDate *string
	if item.EndDate != nil {
		formatted := formatAPIMonth(*item.EndDate)
		endDate = &formatted
	}

	return subscriptionResponse{
		ID:          item.ID.String(),
		ServiceName: item.ServiceName,
		Price:       item.Price,
		UserID:      item.UserID.String(),
		StartDate:   formatAPIMonth(item.StartDate),
		EndDate:     endDate,
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}
