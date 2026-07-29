package apiary

import (
	"encoding/json"
	"math"
	"testing"
)

func TestNullFloat64JSON(t *testing.T) {
	tests := []struct {
		name  string
		value NullFloat64
		want  string
	}{
		{name: "valid", value: NullFloat64{Float64: 1.25, Valid: true}, want: "1.25"},
		{name: "null", value: NullFloat64{}, want: "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("marshal NullFloat64: %v", err)
			}
			if got := string(encoded); got != tt.want {
				t.Fatalf("marshal NullFloat64 = %s, want %s", got, tt.want)
			}
		})
	}

	var value NullFloat64
	if err := json.Unmarshal([]byte("2.5"), &value); err != nil {
		t.Fatalf("unmarshal valid NullFloat64: %v", err)
	}
	if !value.Valid || value.Float64 != 2.5 {
		t.Fatalf("unmarshal valid NullFloat64 = %+v, want 2.5 and valid", value)
	}
	if err := json.Unmarshal([]byte("null"), &value); err != nil {
		t.Fatalf("unmarshal null NullFloat64: %v", err)
	}
	if value.Valid {
		t.Fatalf("unmarshal null NullFloat64 = %+v, want invalid", value)
	}
	if err := json.Unmarshal([]byte(`"invalid"`), &value); err == nil {
		t.Fatal("unmarshal invalid NullFloat64 returned nil error")
	}
}

func TestNullFloat64Scan(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  float64
		valid bool
	}{
		{name: "null", input: nil, valid: false},
		{name: "float64", input: float64(1.5), want: 1.5, valid: true},
		{name: "float32", input: float32(2.5), want: 2.5, valid: true},
		{name: "int64", input: int64(3), want: 3, valid: true},
		{name: "int", input: 4, want: 4, valid: true},
		{name: "string", input: "5.5", want: 5.5, valid: true},
		{name: "bytes", input: []byte("6.5"), want: 6.5, valid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var value NullFloat64
			if err := value.Scan(tt.input); err != nil {
				t.Fatalf("Scan(%v): %v", tt.input, err)
			}
			if value.Valid != tt.valid {
				t.Fatalf("Valid = %t, want %t", value.Valid, tt.valid)
			}
			if math.Abs(value.Float64-tt.want) > 0.000001 {
				t.Fatalf("Float64 = %f, want %f", value.Float64, tt.want)
			}
		})
	}

	for _, input := range []interface{}{"invalid", []byte("invalid"), true} {
		var value NullFloat64
		if err := value.Scan(input); err == nil {
			t.Fatalf("Scan(%v) returned nil error", input)
		}
	}
}
