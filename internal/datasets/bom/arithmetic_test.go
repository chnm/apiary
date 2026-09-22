package bom

import (
	"net/http/httptest"
	"testing"
)

func TestParseArithmeticParameters(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    arithmeticParameters
		wantErr bool
	}{
		{name: "defaults", query: "", want: arithmeticParameters{StartYear: 1636, EndYear: 1754}},
		{name: "single year", query: "start-year=1665&end-year=1665", want: arithmeticParameters{StartYear: 1665, EndYear: 1665}},
		{name: "count type is case-insensitive", query: "count-type=Plague", want: arithmeticParameters{StartYear: 1636, EndYear: 1754, CountType: "plague"}},
		{name: "legible false", query: "legible=false", want: arithmeticParameters{StartYear: 1636, EndYear: 1754, Legible: false}},
		{name: "bad start year", query: "start-year=abc", wantErr: true},
		{name: "bad end year", query: "end-year=1665.5", wantErr: true},
		{name: "reversed range", query: "start-year=1700&end-year=1650", wantErr: true},
		{name: "bad count type", query: "count-type=christened", wantErr: true},
		{name: "bad legible", query: "legible=maybe", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArithmeticParameters(httptest.NewRequest("GET", "/bom/arithmetic?"+tt.query, nil))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestArithmeticHandlerRejectsBadInput(t *testing.T) {
	response := httptest.NewRecorder()
	New(nil).ArithmeticHandler()(response, httptest.NewRequest("GET", "/bom/arithmetic?count-type=christened", nil))
	if response.Code != 400 {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}
