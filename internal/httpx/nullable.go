// Package httpx contains HTTP and JSON helpers shared by Apiary datasets.
package httpx

import (
	"database/sql"
	"encoding/json"
)

// NullInt64 marshals a nullable SQL integer as a JSON number or null.
type NullInt64 struct {
	sql.NullInt64
}

// MarshalJSON implements json.Marshaler.
func (v NullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *NullInt64) UnmarshalJSON(data []byte) error {
	var value *int64
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value == nil {
		v.Valid = false
		return nil
	}
	v.Valid = true
	v.Int64 = *value
	return nil
}

// NullString marshals a nullable SQL string as a JSON string or null.
type NullString struct {
	sql.NullString
}

// MarshalJSON implements json.Marshaler.
func (v NullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *NullString) UnmarshalJSON(data []byte) error {
	var value *string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value == nil {
		v.Valid = false
		return nil
	}
	v.Valid = true
	v.String = *value
	return nil
}
