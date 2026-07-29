package apiary

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
	emptyInt := NullInt64{sql.NullInt64{Int64: 0, Valid: false}}
	out, _ := json.Marshal(emptyInt)
	if string(out) != "null" {
		t.Errorf("Want: null. Got: %s.", out)
	}
}

func TestNullString(t *testing.T) {
	emptyString := NullString{sql.NullString{String: "", Valid: false}}
	out, _ := json.Marshal(emptyString)
	if string(out) != "null" {
		t.Errorf("Want: null. Got: %s.", out)
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
