package subscription

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	testSubscriptionID = "2f8d9b27-5b9e-4d6f-83ef-cf4fef0c9fc2"
	testUserID         = "60601fee-2bf1-4721-ae6f-7636e79a0cba"
)

type fakeSubscriptionService struct {
	t *testing.T

	createFn              func(context.Context, UpsertSubscription) (Subscription, error)
	getFn                 func(context.Context, uuid.UUID) (Subscription, error)
	listFn                func(context.Context, ListSubscriptionsFilter) ([]Subscription, error)
	calculateTotalPriceFn func(context.Context, TotalPriceFilter) (int64, error)
	updateFn              func(context.Context, uuid.UUID, UpsertSubscription) (Subscription, error)
	deleteFn              func(context.Context, uuid.UUID) error
}

func (f *fakeSubscriptionService) Create(ctx context.Context, input UpsertSubscription) (Subscription, error) {
	if f.createFn == nil {
		f.t.Helper()
		f.t.Fatal("unexpected call to Create")
	}
	return f.createFn(ctx, input)
}

func (f *fakeSubscriptionService) Get(ctx context.Context, id uuid.UUID) (Subscription, error) {
	if f.getFn == nil {
		f.t.Helper()
		f.t.Fatal("unexpected call to Get")
	}
	return f.getFn(ctx, id)
}

func (f *fakeSubscriptionService) List(ctx context.Context, filter ListSubscriptionsFilter) ([]Subscription, error) {
	if f.listFn == nil {
		f.t.Helper()
		f.t.Fatal("unexpected call to List")
	}
	return f.listFn(ctx, filter)
}

func (f *fakeSubscriptionService) CalculateTotalPrice(ctx context.Context, filter TotalPriceFilter) (int64, error) {
	if f.calculateTotalPriceFn == nil {
		f.t.Helper()
		f.t.Fatal("unexpected call to CalculateTotalPrice")
	}
	return f.calculateTotalPriceFn(ctx, filter)
}

func (f *fakeSubscriptionService) Update(ctx context.Context, id uuid.UUID, input UpsertSubscription) (Subscription, error) {
	if f.updateFn == nil {
		f.t.Helper()
		f.t.Fatal("unexpected call to Update")
	}
	return f.updateFn(ctx, id, input)
}

func (f *fakeSubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	if f.deleteFn == nil {
		f.t.Helper()
		f.t.Fatal("unexpected call to Delete")
	}
	return f.deleteFn(ctx, id)
}

func TestHandlerCreate(t *testing.T) {
	want := testSubscription()
	service := &fakeSubscriptionService{
		t: t,
		createFn: func(_ context.Context, input UpsertSubscription) (Subscription, error) {
			assertUpsertInput(t, input, "Yandex Plus", 400)
			return want, nil
		},
	}

	recorder := performRequest(t, service, http.MethodPost, "/api/v1/subscriptions", `{
		"service_name":"Yandex Plus",
		"price":400,
		"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba",
		"start_date":"07-2025",
		"end_date":"12-2025"
	}`)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	assertSubscriptionResponse(t, recorder.Body.Bytes(), want)
}

func TestHandlerCreateValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantMessage string
	}{
		{
			name:        "invalid JSON",
			body:        `{"service_name":`,
			wantMessage: "request body must be valid JSON",
		},
		{
			name: "invalid field value",
			body: `{
				"service_name":"Yandex Plus",
				"price":0,
				"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba",
				"start_date":"07-2025"
			}`,
			wantMessage: "price must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &fakeSubscriptionService{t: t}
			recorder := performRequest(t, service, http.MethodPost, "/api/v1/subscriptions", tt.body)

			assertErrorResponse(t, recorder, http.StatusBadRequest, "validation_error", tt.wantMessage)
		})
	}
}

func TestHandlerList(t *testing.T) {
	want := testSubscription()
	var gotFilter ListSubscriptionsFilter
	service := &fakeSubscriptionService{
		t: t,
		listFn: func(_ context.Context, filter ListSubscriptionsFilter) ([]Subscription, error) {
			gotFilter = filter
			return []Subscription{want}, nil
		},
	}
	query := url.Values{
		"limit":        {"7"},
		"offset":       {"3"},
		"user_id":      {testUserID},
		"service_name": {"Yandex Plus"},
	}

	recorder := performRequest(t, service, http.MethodGet, "/api/v1/subscriptions?"+query.Encode(), "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if gotFilter.Limit != 7 || gotFilter.Offset != 3 {
		t.Fatalf("pagination filter = (%d, %d), want (7, 3)", gotFilter.Limit, gotFilter.Offset)
	}
	if gotFilter.UserID == nil || gotFilter.UserID.String() != testUserID {
		t.Fatalf("UserID filter = %v, want %s", gotFilter.UserID, testUserID)
	}
	if gotFilter.ServiceName == nil || *gotFilter.ServiceName != "Yandex Plus" {
		t.Fatalf("ServiceName filter = %v, want Yandex Plus", gotFilter.ServiceName)
	}

	var body struct {
		Items  []json.RawMessage `json:"items"`
		Limit  int32             `json:"limit"`
		Offset int32             `json:"offset"`
	}
	decodeJSON(t, recorder.Body.Bytes(), &body)
	if body.Limit != 7 || body.Offset != 3 {
		t.Fatalf("response pagination = (%d, %d), want (7, 3)", body.Limit, body.Offset)
	}
	if len(body.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(body.Items))
	}
	assertSubscriptionResponse(t, body.Items[0], want)
}

func TestHandlerListInvalidQuery(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantMessage string
	}{
		{name: "invalid limit", query: "limit=101", wantMessage: "limit must be an integer between 1 and 100"},
		{name: "invalid offset", query: "offset=-1", wantMessage: "offset must be a non-negative integer"},
		{name: "invalid user id", query: "user_id=not-a-uuid", wantMessage: "user_id must be UUID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &fakeSubscriptionService{t: t}
			recorder := performRequest(t, service, http.MethodGet, "/api/v1/subscriptions?"+tt.query, "")

			assertErrorResponse(t, recorder, http.StatusBadRequest, "validation_error", tt.wantMessage)
		})
	}
}

func TestHandlerGet(t *testing.T) {
	want := testSubscription()
	service := &fakeSubscriptionService{
		t: t,
		getFn: func(_ context.Context, id uuid.UUID) (Subscription, error) {
			if id.String() != testSubscriptionID {
				t.Fatalf("id = %s, want %s", id, testSubscriptionID)
			}
			return want, nil
		},
	}

	recorder := performRequest(t, service, http.MethodGet, "/api/v1/subscriptions/"+testSubscriptionID, "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	assertSubscriptionResponse(t, recorder.Body.Bytes(), want)
}

func TestHandlerGetInvalidID(t *testing.T) {
	service := &fakeSubscriptionService{t: t}
	recorder := performRequest(t, service, http.MethodGet, "/api/v1/subscriptions/not-a-uuid", "")

	assertErrorResponse(t, recorder, http.StatusBadRequest, "validation_error", "id must be UUID")
}

func TestHandlerGetNotFound(t *testing.T) {
	service := &fakeSubscriptionService{
		t: t,
		getFn: func(context.Context, uuid.UUID) (Subscription, error) {
			return Subscription{}, ErrNotFound
		},
	}
	recorder := performRequest(t, service, http.MethodGet, "/api/v1/subscriptions/"+testSubscriptionID, "")

	assertErrorResponse(t, recorder, http.StatusNotFound, "not_found", "subscription not found")
}

func TestHandlerUpdate(t *testing.T) {
	want := testSubscription()
	want.Price = 500
	service := &fakeSubscriptionService{
		t: t,
		updateFn: func(_ context.Context, id uuid.UUID, input UpsertSubscription) (Subscription, error) {
			if id.String() != testSubscriptionID {
				t.Fatalf("id = %s, want %s", id, testSubscriptionID)
			}
			assertUpsertInput(t, input, "Yandex Plus", 500)
			return want, nil
		},
	}

	recorder := performRequest(t, service, http.MethodPut, "/api/v1/subscriptions/"+testSubscriptionID, `{
		"service_name":"Yandex Plus",
		"price":500,
		"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba",
		"start_date":"07-2025",
		"end_date":"12-2025"
	}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	assertSubscriptionResponse(t, recorder.Body.Bytes(), want)
}

func TestHandlerUpdateInvalidID(t *testing.T) {
	service := &fakeSubscriptionService{t: t}
	recorder := performRequest(t, service, http.MethodPut, "/api/v1/subscriptions/not-a-uuid", `{}`)

	assertErrorResponse(t, recorder, http.StatusBadRequest, "validation_error", "id must be UUID")
}

func TestHandlerDelete(t *testing.T) {
	service := &fakeSubscriptionService{
		t: t,
		deleteFn: func(_ context.Context, id uuid.UUID) error {
			if id.String() != testSubscriptionID {
				t.Fatalf("id = %s, want %s", id, testSubscriptionID)
			}
			return nil
		},
	}

	recorder := performRequest(t, service, http.MethodDelete, "/api/v1/subscriptions/"+testSubscriptionID, "")

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusNoContent, recorder.Body.String())
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty", recorder.Body.String())
	}
}

func TestHandlerDeleteNotFound(t *testing.T) {
	service := &fakeSubscriptionService{
		t: t,
		deleteFn: func(context.Context, uuid.UUID) error {
			return ErrNotFound
		},
	}
	recorder := performRequest(t, service, http.MethodDelete, "/api/v1/subscriptions/"+testSubscriptionID, "")

	assertErrorResponse(t, recorder, http.StatusNotFound, "not_found", "subscription not found")
}

func TestHandlerTotal(t *testing.T) {
	var gotFilter TotalPriceFilter
	service := &fakeSubscriptionService{
		t: t,
		calculateTotalPriceFn: func(_ context.Context, filter TotalPriceFilter) (int64, error) {
			gotFilter = filter
			return 2400, nil
		},
	}
	query := url.Values{
		"from":         {"07-2025"},
		"to":           {"12-2025"},
		"user_id":      {testUserID},
		"service_name": {"Yandex Plus"},
	}

	recorder := performRequest(t, service, http.MethodGet, "/api/v1/subscriptions/total?"+query.Encode(), "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if got := gotFilter.PeriodFrom.Format("01-2006"); got != "07-2025" {
		t.Fatalf("PeriodFrom = %s, want 07-2025", got)
	}
	if got := gotFilter.PeriodTo.Format("01-2006"); got != "12-2025" {
		t.Fatalf("PeriodTo = %s, want 12-2025", got)
	}
	if gotFilter.UserID == nil || gotFilter.UserID.String() != testUserID {
		t.Fatalf("UserID filter = %v, want %s", gotFilter.UserID, testUserID)
	}
	if gotFilter.ServiceName == nil || *gotFilter.ServiceName != "Yandex Plus" {
		t.Fatalf("ServiceName filter = %v, want Yandex Plus", gotFilter.ServiceName)
	}

	var body struct {
		TotalPrice  int64   `json:"total_price"`
		PeriodFrom  string  `json:"period_from"`
		PeriodTo    string  `json:"period_to"`
		UserID      *string `json:"user_id"`
		ServiceName *string `json:"service_name"`
	}
	decodeJSON(t, recorder.Body.Bytes(), &body)
	if body.TotalPrice != 2400 || body.PeriodFrom != "07-2025" || body.PeriodTo != "12-2025" {
		t.Fatalf("total response = %+v, want total 2400 for 07-2025 through 12-2025", body)
	}
	if body.UserID == nil || *body.UserID != testUserID {
		t.Fatalf("response user_id = %v, want %s", body.UserID, testUserID)
	}
	if body.ServiceName == nil || *body.ServiceName != "Yandex Plus" {
		t.Fatalf("response service_name = %v, want Yandex Plus", body.ServiceName)
	}
}

func TestHandlerTotalMissingRequiredQuery(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		wantMessage string
	}{
		{name: "missing from", target: "/api/v1/subscriptions/total?to=12-2025", wantMessage: "from is required"},
		{name: "missing to", target: "/api/v1/subscriptions/total?from=07-2025", wantMessage: "to is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &fakeSubscriptionService{t: t}
			recorder := performRequest(t, service, http.MethodGet, tt.target, "")

			assertErrorResponse(t, recorder, http.StatusBadRequest, "validation_error", tt.wantMessage)
		})
	}
}

func performRequest(t *testing.T, service subscriptionService, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHandler(service, logger)
	router := gin.New()
	handler.RegisterRoutes(router.Group("/api/v1"))

	request := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	return recorder
}

func testSubscription() Subscription {
	endDate := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)
	return Subscription{
		ID:          uuid.MustParse(testSubscriptionID),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.MustParse(testUserID),
		StartDate:   time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     &endDate,
		CreatedAt:   time.Date(2026, time.June, 23, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, time.June, 23, 11, 30, 0, 0, time.UTC),
	}
}

func assertUpsertInput(t *testing.T, input UpsertSubscription, serviceName string, price int32) {
	t.Helper()
	if input.ServiceName != serviceName || input.Price != price {
		t.Fatalf("input service and price = (%q, %d), want (%q, %d)", input.ServiceName, input.Price, serviceName, price)
	}
	if input.UserID.String() != testUserID {
		t.Fatalf("input UserID = %s, want %s", input.UserID, testUserID)
	}
	if got := input.StartDate.Format("01-2006"); got != "07-2025" {
		t.Fatalf("input StartDate = %s, want 07-2025", got)
	}
	if input.EndDate == nil || input.EndDate.Format("01-2006") != "12-2025" {
		t.Fatalf("input EndDate = %v, want 12-2025", input.EndDate)
	}
}

func assertSubscriptionResponse(t *testing.T, data []byte, want Subscription) {
	t.Helper()
	var got struct {
		ID          string  `json:"id"`
		ServiceName string  `json:"service_name"`
		Price       int32   `json:"price"`
		UserID      string  `json:"user_id"`
		StartDate   string  `json:"start_date"`
		EndDate     *string `json:"end_date"`
		CreatedAt   string  `json:"created_at"`
		UpdatedAt   string  `json:"updated_at"`
	}
	decodeJSON(t, data, &got)

	if got.ID != want.ID.String() || got.ServiceName != want.ServiceName || got.Price != want.Price {
		t.Fatalf("subscription identity fields = (%q, %q, %d), want (%q, %q, %d)", got.ID, got.ServiceName, got.Price, want.ID, want.ServiceName, want.Price)
	}
	if got.UserID != want.UserID.String() || got.StartDate != formatAPIMonth(want.StartDate) {
		t.Fatalf("subscription owner and start = (%q, %q), want (%q, %q)", got.UserID, got.StartDate, want.UserID, formatAPIMonth(want.StartDate))
	}
	if want.EndDate == nil {
		if got.EndDate != nil {
			t.Fatalf("end_date = %v, want nil", got.EndDate)
		}
	} else if got.EndDate == nil || *got.EndDate != formatAPIMonth(*want.EndDate) {
		t.Fatalf("end_date = %v, want %s", got.EndDate, formatAPIMonth(*want.EndDate))
	}
	if got.CreatedAt != want.CreatedAt.Format(time.RFC3339Nano) || got.UpdatedAt != want.UpdatedAt.Format(time.RFC3339Nano) {
		t.Fatalf("timestamps = (%q, %q), want (%q, %q)", got.CreatedAt, got.UpdatedAt, want.CreatedAt.Format(time.RFC3339Nano), want.UpdatedAt.Format(time.RFC3339Nano))
	}
}

func assertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int, wantCode, wantMessage string) {
	t.Helper()
	if recorder.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, wantStatus, recorder.Body.String())
	}

	var body map[string]any
	decodeJSON(t, recorder.Body.Bytes(), &body)
	if got, ok := body["error"].(string); !ok || got != wantCode {
		t.Fatalf("error = %v, want %q", body["error"], wantCode)
	}
	if got, ok := body["message"].(string); !ok || got != wantMessage {
		t.Fatalf("message = %v, want %q", body["message"], wantMessage)
	}
}

func decodeJSON(t *testing.T, data []byte, target any) {
	t.Helper()
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("decode response JSON: %v; body = %s", err, data)
	}
}
