// Package params contains shared request-parameter parsing helpers.
package params

import "time"

// DateInRange parses an ISO date and clamps it to the supplied range.
func DateInRange(value string, minDate, maxDate time.Time) (time.Time, error) {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, err
	}
	if date.Before(minDate) {
		return minDate, nil
	}
	if date.After(maxDate) {
		return maxDate, nil
	}
	return date, nil
}
