package apiary

import (
	"database/sql"
	"encoding/json"
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

// NullInt64 embeds the sql.NullInt64 type, so that it can be extended
// to change the JSON marshaling.
type NullInt64 struct {
	sql.NullInt64
}

// MarshalJSON marshals a null integer as `{"int": null}` rather than
// using an object inside the field.
// See https://stackoverflow.com/questions/33072172/how-can-i-work-with-sql-null-values-and-json-in-a-good-way
func (v NullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals a null integer represented in JSON as `{"int": null}`
// into our embedded type that allows nulls.
// See https://stackoverflow.com/questions/33072172/how-can-i-work-with-sql-null-values-and-json-in-a-good-way
func (v *NullInt64) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *int64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Int64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

// NullString embeds the sql.NullString type, so that it can be extended
// to change the JSON marshaling.
type NullString struct {
	sql.NullString
}

// MarshalJSON marshals a null string as `{"string": null}` rather than
// using an object inside the field.
// See https://stackoverflow.com/questions/33072172/how-can-i-work-with-sql-null-values-and-json-in-a-good-way
func (v NullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals a null string represented in JSON as `{"string": null}`
// into our embedded type that allows nulls.
// See https://stackoverflow.com/questions/33072172/how-can-i-work-with-sql-null-values-and-json-in-a-good-way
func (v *NullString) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *string
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.String = *x
	} else {
		v.Valid = false
	}
	return nil
}
