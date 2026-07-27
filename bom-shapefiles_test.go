package apiary

import (
	"reflect"
	"strings"
	"testing"
)

func TestBuildSeparateFiltersParameterizesValues(t *testing.T) {
	injection := "City' OR 1=1 --"

	billFilters, parishFilters, params, err := buildSeparateFilters(
		"1665", "", "", injection, "London", "Weekly", "Buried", "1, 2",
	)
	if err != nil {
		t.Fatalf("buildSeparateFilters returned an unexpected error: %v", err)
	}

	wantBillFilters := "AND b.year = $1 AND b.bill_type = $4 AND b.count_type = $5 AND b.parish_id = ANY($6)"
	if billFilters != wantBillFilters {
		t.Errorf("bill filters = %q, want %q", billFilters, wantBillFilters)
	}

	wantParishFilters := "AND parishes_shp.subunit = $2 AND parishes_shp.city_cnty = $3 AND parishes_shp.id = ANY($6)"
	if parishFilters != wantParishFilters {
		t.Errorf("parish filters = %q, want %q", parishFilters, wantParishFilters)
	}

	if strings.Contains(billFilters, injection) || strings.Contains(parishFilters, injection) {
		t.Fatal("user input was interpolated into SQL filters")
	}

	wantParams := []any{1665, injection, "London", "weekly", "buried", []int{1, 2}}
	if !reflect.DeepEqual(params, wantParams) {
		t.Errorf("params = %#v, want %#v", params, wantParams)
	}
}

func TestBuildSeparateFiltersRejectsMalformedValues(t *testing.T) {
	tests := []struct {
		name      string
		year      string
		startYear string
		endYear   string
		billType  string
		countType string
		parish    string
	}{
		{name: "year", year: "abc"},
		{name: "start year", startYear: "abc"},
		{name: "end year", endYear: "abc"},
		{name: "bill type", billType: "invalid"},
		{name: "count type", countType: "invalid"},
		{name: "parish", parish: "abc"},
		{name: "empty parish", parish: "1,,2"},
		{name: "non-positive parish", parish: "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := buildSeparateFilters(
				tt.year, tt.startYear, tt.endYear, "", "", tt.billType, tt.countType, tt.parish,
			)
			if err == nil {
				t.Fatal("buildSeparateFilters returned nil error")
			}
		})
	}
}
