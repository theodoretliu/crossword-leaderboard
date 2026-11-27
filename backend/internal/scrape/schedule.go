package scrape

import "time"

// FirstDate is the earliest puzzle date the scrapers cover.
var FirstDate = time.Date(2020, time.August, 21, 0, 0, 0, 0, time.UTC)

// StartOfDay returns midnight UTC of t's UTC calendar day, which is how puzzle
// dates are represented throughout the scrapers.
func StartOfDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// Interval is how long to wait between scrapes of a puzzle whose date is age in
// the past. The wait grows quadratically with age, so today's puzzle is
// re-scraped every few seconds while old puzzles are left nearly alone:
//
//	age      interval
//	0        10s
//	1 day    20s
//	1 week   ~5m
//	1 year   ~7.7 days
func Interval(age time.Duration) time.Duration {
	h := age.Hours()
	seconds := 5*h*h/576 + 5*h/24 + 10
	return time.Duration(seconds) * time.Second
}
