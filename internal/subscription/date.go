package subscription

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const apiMonthLayout = "01-2006"

var apiMonthPattern = regexp.MustCompile(`^\d{2}-\d{4}$`)

func parseAPIMonth(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if !apiMonthPattern.MatchString(value) {
		return time.Time{}, fmt.Errorf("must use MM-YYYY format")
	}

	parsed, err := time.Parse(apiMonthLayout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("must use MM-YYYY format")
	}

	return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

func formatAPIMonth(value time.Time) string {
	return value.UTC().Format(apiMonthLayout)
}

func pgDate(value time.Time) pgtype.Date {
	return pgtype.Date{
		Time:  time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

func nullablePGDate(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}

	return pgDate(*value)
}

func timeFromPGDate(value pgtype.Date) (time.Time, bool) {
	if !value.Valid {
		return time.Time{}, false
	}

	return time.Date(value.Time.Year(), value.Time.Month(), 1, 0, 0, 0, 0, time.UTC), true
}

func timeFromPGTimestamptz(value pgtype.Timestamptz) (time.Time, bool) {
	if !value.Valid {
		return time.Time{}, false
	}

	return value.Time.UTC(), true
}
