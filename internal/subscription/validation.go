package subscription

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

type upsertSubscriptionRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

func validateUpsertSubscriptionRequest(req upsertSubscriptionRequest) (UpsertSubscription, error) {
	serviceName := strings.TrimSpace(req.ServiceName)
	if serviceName == "" {
		return UpsertSubscription{}, fmt.Errorf("service_name is required")
	}

	if req.Price <= 0 {
		return UpsertSubscription{}, fmt.Errorf("price must be greater than 0")
	}
	if req.Price > math.MaxInt32 {
		return UpsertSubscription{}, fmt.Errorf("price is too large")
	}

	userIDValue := strings.TrimSpace(req.UserID)
	if userIDValue == "" {
		return UpsertSubscription{}, fmt.Errorf("user_id is required")
	}
	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		return UpsertSubscription{}, fmt.Errorf("user_id must be UUID")
	}

	startDateValue := strings.TrimSpace(req.StartDate)
	if startDateValue == "" {
		return UpsertSubscription{}, fmt.Errorf("start_date is required")
	}
	startDate, err := parseAPIMonth(startDateValue)
	if err != nil {
		return UpsertSubscription{}, fmt.Errorf("start_date must use MM-YYYY format")
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsedEndDate, err := parseAPIMonth(*req.EndDate)
		if err != nil {
			return UpsertSubscription{}, fmt.Errorf("end_date must use MM-YYYY format")
		}
		if parsedEndDate.Before(startDate) {
			return UpsertSubscription{}, fmt.Errorf("end_date must not be earlier than start_date")
		}
		endDate = &parsedEndDate
	}

	input := UpsertSubscription{
		ServiceName: serviceName,
		Price:       int32(req.Price),
		UserID:      userID,
		StartDate:   startDate,
	}
	if endDate != nil {
		input.EndDate = endDate
	}

	return input, nil
}

func parseListSubscriptionsFilter(values url.Values) (ListSubscriptionsFilter, error) {
	filter := ListSubscriptionsFilter{
		Limit:  defaultListLimit,
		Offset: 0,
	}

	if rawLimit, ok := firstQueryValue(values, "limit"); ok {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit <= 0 || limit > maxListLimit {
			return ListSubscriptionsFilter{}, fmt.Errorf("limit must be an integer between 1 and %d", maxListLimit)
		}
		filter.Limit = int32(limit)
	}

	if rawOffset, ok := firstQueryValue(values, "offset"); ok {
		offset, err := strconv.Atoi(rawOffset)
		if err != nil || offset < 0 || offset > math.MaxInt32 {
			return ListSubscriptionsFilter{}, fmt.Errorf("offset must be a non-negative integer")
		}
		filter.Offset = int32(offset)
	}

	if rawUserID, ok := firstQueryValue(values, "user_id"); ok {
		rawUserID = strings.TrimSpace(rawUserID)
		if rawUserID == "" {
			return ListSubscriptionsFilter{}, fmt.Errorf("user_id must be UUID")
		}
		userID, err := uuid.Parse(rawUserID)
		if err != nil {
			return ListSubscriptionsFilter{}, fmt.Errorf("user_id must be UUID")
		}
		filter.UserID = &userID
	}

	if rawServiceName, ok := firstQueryValue(values, "service_name"); ok {
		serviceName := strings.TrimSpace(rawServiceName)
		if serviceName == "" {
			return ListSubscriptionsFilter{}, fmt.Errorf("service_name must not be blank")
		}
		filter.ServiceName = &serviceName
	}

	return filter, nil
}

func parseTotalPriceFilter(values url.Values) (TotalPriceFilter, error) {
	rawFrom, ok := firstQueryValue(values, "from")
	if !ok || strings.TrimSpace(rawFrom) == "" {
		return TotalPriceFilter{}, fmt.Errorf("from is required")
	}
	periodFrom, err := parseAPIMonth(rawFrom)
	if err != nil {
		return TotalPriceFilter{}, fmt.Errorf("from must use MM-YYYY format")
	}

	rawTo, ok := firstQueryValue(values, "to")
	if !ok || strings.TrimSpace(rawTo) == "" {
		return TotalPriceFilter{}, fmt.Errorf("to is required")
	}
	periodTo, err := parseAPIMonth(rawTo)
	if err != nil {
		return TotalPriceFilter{}, fmt.Errorf("to must use MM-YYYY format")
	}
	if periodTo.Before(periodFrom) {
		return TotalPriceFilter{}, fmt.Errorf("to must not be earlier than from")
	}

	filter := TotalPriceFilter{
		PeriodFrom: periodFrom,
		PeriodTo:   periodTo,
	}

	if rawUserID, ok := firstQueryValue(values, "user_id"); ok {
		rawUserID = strings.TrimSpace(rawUserID)
		if rawUserID == "" {
			return TotalPriceFilter{}, fmt.Errorf("user_id must be UUID")
		}
		userID, err := uuid.Parse(rawUserID)
		if err != nil {
			return TotalPriceFilter{}, fmt.Errorf("user_id must be UUID")
		}
		filter.UserID = &userID
	}

	if rawServiceName, ok := firstQueryValue(values, "service_name"); ok {
		serviceName := strings.TrimSpace(rawServiceName)
		if serviceName == "" {
			return TotalPriceFilter{}, fmt.Errorf("service_name must not be blank")
		}
		filter.ServiceName = &serviceName
	}

	return filter, nil
}

func firstQueryValue(values url.Values, name string) (string, bool) {
	rawValues, ok := values[name]
	if !ok {
		return "", false
	}
	if len(rawValues) == 0 {
		return "", true
	}

	return rawValues[0], true
}
