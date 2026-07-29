package apiary

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

type failingJSONMarshaler struct{}

func (failingJSONMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errors.New("test encoding failure")
}

func TestWriteJSONResponse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		response := httptest.NewRecorder()

		writeJSONResponse(response, struct {
			Status string `json:"status"`
		}{Status: "ok"})

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
		if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", contentType)
		}
		if body := response.Body.String(); body != `{"status":"ok"}` {
			t.Fatalf(`body = %q, want {"status":"ok"}`, body)
		}
	})

	t.Run("encoding failure", func(t *testing.T) {
		response := httptest.NewRecorder()

		writeJSONResponse(response, failingJSONMarshaler{})

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
		}
		if body := strings.TrimSpace(response.Body.String()); body != http.StatusText(http.StatusInternalServerError) {
			t.Fatalf(
				"body = %q, want %q",
				body,
				http.StatusText(http.StatusInternalServerError),
			)
		}
	})
}

func Test_dateInRange(t *testing.T) {
	minDate, _ := time.Parse("2006-01-02", "1783-09-03")
	maxDate, _ := time.Parse("2006-01-02", "2000-12-31")

	type args struct {
		d   string
		min time.Time
		max time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "Test date in range",
			args: args{
				d:   "1900-06-09",
				min: minDate,
				max: maxDate,
			},
			want:    time.Date(1900, 6, 9, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name: "Test date after range",
			args: args{
				d:   "2020-06-09",
				min: minDate,
				max: maxDate,
			},
			want:    maxDate,
			wantErr: false,
		},
		{
			name: "Test date before range",
			args: args{
				d:   "1620-06-09",
				min: minDate,
				max: maxDate,
			},
			want:    minDate,
			wantErr: false,
		},
		{
			name: "Test invalid date",
			args: args{
				d:   "1920-15-40",
				min: minDate,
				max: maxDate,
			},
			want:    time.Time{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dateInRange(tt.args.d, tt.args.min, tt.args.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("dateInRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dateInRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullInt64(t *testing.T) {
	tests := []struct {
		name  string
		value NullInt64
		want  string
	}{
		{
			name:  "valid",
			value: NullInt64{NullInt64: sql.NullInt64{Int64: 42, Valid: true}},
			want:  "42",
		},
		{
			name:  "null",
			value: NullInt64{NullInt64: sql.NullInt64{Valid: false}},
			want:  "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("marshal NullInt64: %v", err)
			}
			if got := string(out); got != tt.want {
				t.Fatalf("marshal NullInt64 = %s, want %s", got, tt.want)
			}
		})
	}

	var value NullInt64
	if err := json.Unmarshal([]byte("27"), &value); err != nil {
		t.Fatalf("unmarshal valid NullInt64: %v", err)
	}
	if !value.Valid || value.Int64 != 27 {
		t.Fatalf("unmarshal valid NullInt64 = %+v, want 27 and valid", value)
	}
	if err := json.Unmarshal([]byte("null"), &value); err != nil {
		t.Fatalf("unmarshal null NullInt64: %v", err)
	}
	if value.Valid {
		t.Fatalf("unmarshal null NullInt64 = %+v, want invalid", value)
	}
	if err := json.Unmarshal([]byte(`"invalid"`), &value); err == nil {
		t.Fatal("unmarshal invalid NullInt64 returned nil error")
	}
}

func TestNullString(t *testing.T) {
	tests := []struct {
		name  string
		value NullString
		want  string
	}{
		{
			name:  "valid",
			value: NullString{NullString: sql.NullString{String: "value", Valid: true}},
			want:  `"value"`,
		},
		{
			name:  "null",
			value: NullString{NullString: sql.NullString{Valid: false}},
			want:  "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("marshal NullString: %v", err)
			}
			if got := string(out); got != tt.want {
				t.Fatalf("marshal NullString = %s, want %s", got, tt.want)
			}
		})
	}

	var value NullString
	if err := json.Unmarshal([]byte(`"text"`), &value); err != nil {
		t.Fatalf("unmarshal valid NullString: %v", err)
	}
	if !value.Valid || value.String != "text" {
		t.Fatalf("unmarshal valid NullString = %+v, want text and valid", value)
	}
	if err := json.Unmarshal([]byte("null"), &value); err != nil {
		t.Fatalf("unmarshal null NullString: %v", err)
	}
	if value.Valid {
		t.Fatalf("unmarshal null NullString = %+v, want invalid", value)
	}
	if err := json.Unmarshal([]byte("42"), &value); err == nil {
		t.Fatal("unmarshal invalid NullString returned nil error")
	}
}

func TestIntsToString(t *testing.T) {
	if got := intsToString([]int{152, 999}); got != "152, 999" {
		t.Errorf(`intsToString([152 999]) = %q, want "152, 999"`, got)
	}
	if got := intsToString([]int{7}); got != "7" {
		t.Errorf(`intsToString([7]) = %q, want "7"`, got)
	}
}

func TestGetEnv(t *testing.T) {
	const key = "APIARY_HELPERS_TEST_ENV"

	original, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset test environment variable: %v", err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, original)
			return
		}
		_ = os.Unsetenv(key)
	})

	if got := getEnv(key, "fallback"); got != "fallback" {
		t.Fatalf("getEnv missing = %q, want fallback", got)
	}

	if err := os.Setenv(key, "configured"); err != nil {
		t.Fatalf("set test environment variable: %v", err)
	}
	if got := getEnv(key, "fallback"); got != "configured" {
		t.Fatalf("getEnv configured = %q, want configured", got)
	}
}
