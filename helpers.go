package dataapi

import (
	"os"
	"time"
)

// getEnv either returns the value of an environment variable or, if that
// environment variables does not exist, returns the fallback value provided.
func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

// dateInRange takes in a string which should be parsed to a date. That date is
// then kept within the range of the min and max dates passed as arguments.
func dateInRange(d string, min, max time.Time) (time.Time, error) {
	format := "2006-01-02"
	date, err := time.Parse(format, d)
	if err != nil {
		return time.Time{}, err
	}
	if date.Before(min) {
		date = min
	} else if date.After(max) {
		date = max
	}
	return date, nil
}
