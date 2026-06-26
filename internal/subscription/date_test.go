package subscription

import "testing"

func TestParseAPIMonth(t *testing.T) {
	t.Parallel()

	parsed, err := parseAPIMonth("07-2025")
	if err != nil {
		t.Fatalf("parseAPIMonth() error = %v", err)
	}

	if got := parsed.Format("2006-01-02"); got != "2025-07-01" {
		t.Fatalf("parseAPIMonth() = %s, want 2025-07-01", got)
	}
}

func TestParseAPIMonthRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []string{
		"",
		"7-2025",
		"13-2025",
		"2025-07",
	}

	for _, value := range tests {
		value := value
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			if _, err := parseAPIMonth(value); err == nil {
				t.Fatalf("parseAPIMonth(%q) error = nil, want error", value)
			}
		})
	}
}

func TestFormatAPIMonth(t *testing.T) {
	t.Parallel()

	parsed, err := parseAPIMonth("12-2025")
	if err != nil {
		t.Fatalf("parseAPIMonth() error = %v", err)
	}

	if got := formatAPIMonth(parsed); got != "12-2025" {
		t.Fatalf("formatAPIMonth() = %s, want 12-2025", got)
	}
}
