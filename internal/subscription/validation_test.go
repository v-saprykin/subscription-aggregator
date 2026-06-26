package subscription

import (
	"net/url"
	"testing"
)

func TestValidateUpsertSubscriptionRequest(t *testing.T) {
	t.Parallel()

	endDate := "12-2025"
	input, err := validateUpsertSubscriptionRequest(upsertSubscriptionRequest{
		ServiceName: " Yandex Plus ",
		Price:       400,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "07-2025",
		EndDate:     &endDate,
	})
	if err != nil {
		t.Fatalf("validateUpsertSubscriptionRequest() error = %v", err)
	}

	if input.ServiceName != "Yandex Plus" {
		t.Fatalf("ServiceName = %q, want Yandex Plus", input.ServiceName)
	}
	if input.Price != 400 {
		t.Fatalf("Price = %d, want 400", input.Price)
	}
	if got := input.StartDate.Format("2006-01-02"); got != "2025-07-01" {
		t.Fatalf("StartDate = %s, want 2025-07-01", got)
	}
	if input.EndDate == nil {
		t.Fatal("EndDate = nil, want value")
	}
	if got := input.EndDate.Format("2006-01-02"); got != "2025-12-01" {
		t.Fatalf("EndDate = %s, want 2025-12-01", got)
	}
}

func TestValidateUpsertSubscriptionRequestErrors(t *testing.T) {
	t.Parallel()

	base := upsertSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "07-2025",
	}

	tests := []struct {
		name    string
		patch   func(*upsertSubscriptionRequest)
		wantErr string
	}{
		{
			name: "blank service name",
			patch: func(req *upsertSubscriptionRequest) {
				req.ServiceName = " "
			},
			wantErr: "service_name is required",
		},
		{
			name: "non-positive price",
			patch: func(req *upsertSubscriptionRequest) {
				req.Price = 0
			},
			wantErr: "price must be greater than 0",
		},
		{
			name: "invalid user id",
			patch: func(req *upsertSubscriptionRequest) {
				req.UserID = "not-a-uuid"
			},
			wantErr: "user_id must be UUID",
		},
		{
			name: "invalid start date",
			patch: func(req *upsertSubscriptionRequest) {
				req.StartDate = "2025-07"
			},
			wantErr: "start_date must use MM-YYYY format",
		},
		{
			name: "end date before start date",
			patch: func(req *upsertSubscriptionRequest) {
				endDate := "06-2025"
				req.EndDate = &endDate
			},
			wantErr: "end_date must not be earlier than start_date",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := base
			tt.patch(&req)

			_, err := validateUpsertSubscriptionRequest(req)
			if err == nil {
				t.Fatal("validateUpsertSubscriptionRequest() error = nil, want error")
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestParseListSubscriptionsFilter(t *testing.T) {
	t.Parallel()

	values := url.Values{
		"limit":        {"50"},
		"offset":       {"10"},
		"user_id":      {"60601fee-2bf1-4721-ae6f-7636e79a0cba"},
		"service_name": {" Yandex Plus "},
	}

	filter, err := parseListSubscriptionsFilter(values)
	if err != nil {
		t.Fatalf("parseListSubscriptionsFilter() error = %v", err)
	}

	if filter.Limit != 50 {
		t.Fatalf("Limit = %d, want 50", filter.Limit)
	}
	if filter.Offset != 10 {
		t.Fatalf("Offset = %d, want 10", filter.Offset)
	}
	if filter.UserID == nil {
		t.Fatal("UserID = nil, want value")
	}
	if filter.ServiceName == nil || *filter.ServiceName != "Yandex Plus" {
		t.Fatalf("ServiceName = %v, want Yandex Plus", filter.ServiceName)
	}
}

func TestParseListSubscriptionsFilterDefaults(t *testing.T) {
	t.Parallel()

	filter, err := parseListSubscriptionsFilter(url.Values{})
	if err != nil {
		t.Fatalf("parseListSubscriptionsFilter() error = %v", err)
	}

	if filter.Limit != defaultListLimit {
		t.Fatalf("Limit = %d, want %d", filter.Limit, defaultListLimit)
	}
	if filter.Offset != 0 {
		t.Fatalf("Offset = %d, want 0", filter.Offset)
	}
}

func TestParseListSubscriptionsFilterErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values url.Values
	}{
		{
			name:   "limit too large",
			values: url.Values{"limit": {"101"}},
		},
		{
			name:   "negative offset",
			values: url.Values{"offset": {"-1"}},
		},
		{
			name:   "invalid user id",
			values: url.Values{"user_id": {"not-a-uuid"}},
		},
		{
			name:   "blank service name",
			values: url.Values{"service_name": {" "}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := parseListSubscriptionsFilter(tt.values); err == nil {
				t.Fatal("parseListSubscriptionsFilter() error = nil, want error")
			}
		})
	}
}
