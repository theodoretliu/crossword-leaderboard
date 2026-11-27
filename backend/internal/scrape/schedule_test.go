package scrape

import (
	"testing"
	"time"
)

// oldHoursToSleepTime is the original cmd/scraper formula, kept to pin Interval
// to identical behavior.
func oldHoursToSleepTime(hours float64) int64 {
	return int64(5*hours*hours/576 + 5*hours/24 + 10)
}

func TestInterval(t *testing.T) {
	tests := []struct {
		age  time.Duration
		want time.Duration
	}{
		{0, 10 * time.Second},
		{time.Hour, 10 * time.Second}, // 10.2s truncates to 10s
		{24 * time.Hour, 20 * time.Second},
		{7 * 24 * time.Hour, 290 * time.Second},
		{365 * 24 * time.Hour, 667960 * time.Second},
	}
	for _, tt := range tests {
		if got := Interval(tt.age); got != tt.want {
			t.Errorf("Interval(%v) = %v, want %v", tt.age, got, tt.want)
		}
	}
}

func TestIntervalMatchesOriginalFormula(t *testing.T) {
	for age := time.Duration(0); age < 7*365*24*time.Hour; age += 37*time.Hour + 13*time.Minute {
		want := time.Duration(oldHoursToSleepTime(age.Hours())) * time.Second
		if got := Interval(age); got != want {
			t.Fatalf("Interval(%v) = %v, original formula gives %v", age, got, want)
		}
	}
}

func TestStartOfDay(t *testing.T) {
	est := time.FixedZone("EST", -5*3600)
	got := StartOfDay(time.Date(2026, 11, 27, 21, 30, 0, 0, est)) // 02:30 UTC on the 28th
	want := time.Date(2026, 11, 28, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) || got.Location() != time.UTC {
		t.Errorf("StartOfDay = %v, want %v", got, want)
	}
}
