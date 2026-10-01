package handlers

import (
	"net/url"
	"testing"
	"time"
)

func TestParseCalendarRange_Default(t *testing.T) {
	from, to, err := parseCalendarRange(url.Values{})
	if err != nil {
		t.Fatalf("parseCalendarRange() error = %v", err)
	}

	if from.Weekday() != time.Sunday {
		t.Errorf("from = %v, want a Sunday", from)
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	jan1 := time.Date(today.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	if from.Before(jan1) || from.After(jan1.AddDate(0, 0, 6)) {
		t.Errorf("from = %v, want within the first 7 days of %d", from, today.Year())
	}
	if want := from.AddDate(0, 0, 370); !to.Equal(want) {
		t.Errorf("to = %v, want %v (371 days after from)", to, want)
	}
}

func TestParseCalendarRange_ExplicitRange(t *testing.T) {
	q := url.Values{"from": {"2026-01-01"}, "to": {"2026-01-31"}}
	from, to, err := parseCalendarRange(q)
	if err != nil {
		t.Fatalf("parseCalendarRange() error = %v", err)
	}
	if from.Format(calendarDateLayout) != "2026-01-01" {
		t.Errorf("from = %v, want 2026-01-01", from)
	}
	if to.Format(calendarDateLayout) != "2026-01-31" {
		t.Errorf("to = %v, want 2026-01-31", to)
	}
}

func TestParseCalendarRange_InvalidFormat(t *testing.T) {
	q := url.Values{"from": {"2026/01/01"}}
	if _, _, err := parseCalendarRange(q); err == nil {
		t.Error("parseCalendarRange() with malformed from should return an error")
	}
}

func TestParseCalendarRange_FromAfterTo(t *testing.T) {
	q := url.Values{"from": {"2026-02-01"}, "to": {"2026-01-01"}}
	if _, _, err := parseCalendarRange(q); err == nil {
		t.Error("parseCalendarRange() with from after to should return an error")
	}
}
