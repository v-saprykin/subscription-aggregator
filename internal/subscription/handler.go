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
	Error   string `json:"error"`
	Message string `json:"message"`
}

type subscriptionResponse struct {
	ID          string  `json:"id"`
	ServiceName string  `json:"service_name"`
	Price       int32   `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type listSubscriptionsResponse struct {
	Items  []subscriptionResponse `json:"items"`
	Limit  int32                  `json:"limit"`
	Offset int32                  `json:"offset"`
}

type totalPriceResponse struct {
	TotalPrice  int64   `json:"total_price"`
	PeriodFrom  string  `json:"period_from"`
	PeriodTo    string  `json:"period_to"`
	UserID      *string `json:"user_id"`
	ServiceName *string `json:"service_name"`
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
