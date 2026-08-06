package bom

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestTotalBillsHandlerRejectsInvalidType(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/bom/totalbills?type=unknown", nil)
	response := httptest.NewRecorder()

	(&Handler{}).TotalBillsHandler().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if body := strings.TrimSpace(response.Body.String()); !strings.HasPrefix(body, "Invalid type parameter.") {
		t.Fatalf("body = %q, want invalid type error", body)
	}
}

func TestBillsHandlerRejectsInvalidSort(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/bom/bills", nil)
	query := request.URL.Query()
	query.Set("sort", "year; select pg_sleep(10)")
	request.URL.RawQuery = query.Encode()
	response := httptest.NewRecorder()

	(&Handler{}).BillsHandler().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if body := strings.TrimSpace(response.Body.String()); body != "invalid sort: supported values are year, week_number, and canonical_name" {
		t.Fatalf("body = %q, want fixed invalid sort error", body)
	}
}

func TestParseAPIParametersParish(t *testing.T) {
	// Non-integer parish IDs are rejected by the parser.
	r := httptest.NewRequest("GET", "/bom/bills?parish=abc", nil)
	if _, err := parseAPIParameters(r); err == nil {
		t.Error("expected error for non-integer parish ID, got nil")
	}

	// Comma-separated parish IDs are parsed in order, with whitespace trimmed.
	r = httptest.NewRequest("GET", "/bom/bills?parish=1,%205,151", nil)
	params, err := parseAPIParameters(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(params.Parish, []int{1, 5, 151}) {
		t.Errorf("expected [1 5 151], got %v", params.Parish)
	}

	// The parser does not range-check IDs; existence is validated against
	// the database by Handler.invalidParishIDs.
	r = httptest.NewRequest("GET", "/bom/bills?parish=0,99999", nil)
	params, err = parseAPIParameters(r)
	if err != nil {
		t.Fatalf("unexpected error for out-of-range IDs: %v", err)
	}
	if !reflect.DeepEqual(params.Parish, []int{0, 99999}) {
		t.Errorf("expected [0 99999], got %v", params.Parish)
	}
}

func TestAPIParametersGetQueryOptions(t *testing.T) {
	tests := []struct {
		name   string
		params APIParameters
		want   QueryOptions
	}{
		{
			name:   "direct limit and offset",
			params: APIParameters{Limit: 50, Offset: 10},
			want:   QueryOptions{Limit: 50, Offset: 10},
		},
		{
			name:   "page overrides direct pagination",
			params: APIParameters{Page: 3, Limit: 50, Offset: 10},
			want:   QueryOptions{Limit: 25, Offset: 50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.params.GetQueryOptions(); got != tt.want {
				t.Fatalf("GetQueryOptions() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCursorRoundTrip(t *testing.T) {
	cursor, err := generateCursor(1665, 12, "All Hallows")
	if err != nil {
		t.Fatalf("generateCursor: %v", err)
	}

	year, week, name, err := parseCursor(cursor)
	if err != nil {
		t.Fatalf("parseCursor: %v", err)
	}
	if year != 1665 || week != 12 || name != "All Hallows" {
		t.Fatalf(
			"parseCursor() = (%d, %d, %q), want (1665, 12, %q)",
			year,
			week,
			name,
			"All Hallows",
		)
	}
}

func TestParseCursorRejectsMalformedValues(t *testing.T) {
	tests := []struct {
		name   string
		cursor string
	}{
		{name: "invalid base64", cursor: "%%%"},
		{
			name:   "wrong field count",
			cursor: base64.URLEncoding.EncodeToString([]byte("1665|12")),
		},
		{
			name:   "invalid year",
			cursor: base64.URLEncoding.EncodeToString([]byte("year|12|Parish")),
		},
		{
			name:   "invalid week",
			cursor: base64.URLEncoding.EncodeToString([]byte("1665|week|Parish")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, _, err := parseCursor(tt.cursor); err == nil {
				t.Fatal("parseCursor returned nil error")
			}
		})
	}
}

func TestGetEffectiveLimit(t *testing.T) {
	tests := []struct {
		name   string
		params APIParameters
		want   int
	}{
		{name: "default", params: APIParameters{}, want: 100},
		{name: "direct", params: APIParameters{Limit: 25}, want: 25},
		{name: "page", params: APIParameters{Page: 2, Limit: 25}, want: 100},
		{name: "cursor", params: APIParameters{Cursor: "cursor", Limit: 25}, want: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getEffectiveLimit(tt.params); got != tt.want {
				t.Fatalf("getEffectiveLimit() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseAPIParameters(t *testing.T) {
	cursor, err := generateCursor(1665, 12, "All Hallows")
	if err != nil {
		t.Fatalf("generate cursor: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/bom/bills?start-year=1664&end-year=1666&start-week=2&end-week=50"+
			"&parish=1,2&bill-type=Weekly&count-type=Buried&missing=true"+
			"&illegible=false&limit=25&offset=5&page=2&sort=canonical_name"+
			"&cursor="+cursor,
		nil,
	)
	params, err := parseAPIParameters(request)
	if err != nil {
		t.Fatalf("parseAPIParameters: %v", err)
	}

	if params.StartYear != 1664 || params.EndYear != 1666 {
		t.Fatalf("year range = %d-%d, want 1664-1666", params.StartYear, params.EndYear)
	}
	if params.StartWeek != 2 || params.EndWeek != 50 {
		t.Fatalf("week range = %d-%d, want 2-50", params.StartWeek, params.EndWeek)
	}
	if !reflect.DeepEqual(params.Parish, []int{1, 2}) {
		t.Fatalf("parishes = %v, want [1 2]", params.Parish)
	}
	if params.BillType != "Weekly" || params.CountType != "Buried" {
		t.Fatalf("types = (%q, %q), want (Weekly, Buried)", params.BillType, params.CountType)
	}
	if params.Missing == nil || !*params.Missing {
		t.Fatal("missing parameter was not parsed as true")
	}
	if params.Illegible == nil || *params.Illegible {
		t.Fatal("illegible parameter was not parsed as false")
	}
	if params.Limit != 25 || params.Offset != 5 || params.Page != 2 {
		t.Fatalf(
			"pagination = limit %d offset %d page %d, want 25, 5, 2",
			params.Limit,
			params.Offset,
			params.Page,
		)
	}
	if params.Sort != "canonical_name" {
		t.Fatalf("sort = %q, want canonical_name", params.Sort)
	}
	if params.CursorYear != 1665 || params.CursorWeek != 12 || params.CursorName != "All Hallows" {
		t.Fatalf(
			"cursor = (%d, %d, %q), want (1665, 12, All Hallows)",
			params.CursorYear,
			params.CursorWeek,
			params.CursorName,
		)
	}
}

func TestParseAPIParametersSort(t *testing.T) {
	tests := []struct {
		name    string
		sort    string
		want    string
		wantErr bool
	}{
		{name: "default", want: defaultBillsSort},
		{name: "year", sort: "year", want: "year"},
		{name: "week number", sort: "week_number", want: "week_number"},
		{name: "canonical name", sort: "canonical_name", want: "canonical_name"},
		{name: "SQL expression", sort: "lower(canonical_name)", wantErr: true},
		{name: "stacked statement", sort: "year; select pg_sleep(10)", wantErr: true},
		{name: "unsupported direction", sort: "year desc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/bom/bills", nil)
			query := request.URL.Query()
			if tt.sort != "" {
				query.Set("sort", tt.sort)
				request.URL.RawQuery = query.Encode()
			}

			params, err := parseAPIParameters(request)
			if tt.wantErr {
				if err == nil {
					t.Fatal("parseAPIParameters returned nil error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseAPIParameters: %v", err)
			}
			if params.Sort != tt.want {
				t.Fatalf("sort = %q, want %q", params.Sort, tt.want)
			}
		})
	}
}

func TestParseAPIParametersRejectsMalformedValues(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "start year", query: "start-year=year"},
		{name: "end year", query: "end-year=year"},
		{name: "start week type", query: "start-week=week"},
		{name: "start week range", query: "start-week=0"},
		{name: "end week type", query: "end-week=week"},
		{name: "end week range", query: "end-week=54"},
		{name: "bill type", query: "bill-type=monthly"},
		{name: "count type", query: "count-type=unknown"},
		{name: "missing", query: "missing=sometimes"},
		{name: "illegible", query: "illegible=sometimes"},
		{name: "limit", query: "limit=many"},
		{name: "offset", query: "offset=later"},
		{name: "page", query: "page=next"},
		{name: "cursor", query: "cursor=invalid!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/bom/bills?"+tt.query, nil)
			if _, err := parseAPIParameters(request); err == nil {
				t.Fatal("parseAPIParameters returned nil error")
			}
		})
	}
}

func TestBuildBillsQueryPaginationModes(t *testing.T) {
	tests := []struct {
		name          string
		params        APIParameters
		queryContains []string
		wantParams    []interface{}
	}{
		{
			name:          "default",
			params:        APIParameters{},
			queryContains: []string{"COUNT(*) OVER()", "LIMIT $1"},
			wantParams:    []interface{}{100},
		},
		{
			name:          "direct",
			params:        APIParameters{Limit: 25, Offset: 10},
			queryContains: []string{"LIMIT $1 OFFSET $2"},
			wantParams:    []interface{}{25, 10},
		},
		{
			name:          "page",
			params:        APIParameters{Page: 2},
			queryContains: []string{"LIMIT $1 OFFSET $2"},
			wantParams:    []interface{}{100, 100},
		},
		{
			name: "cursor",
			params: APIParameters{
				Cursor:     "cursor",
				CursorYear: 1665,
				CursorWeek: 12,
				CursorName: "All Hallows",
			},
			queryContains: []string{
				"0 AS totalrecords",
				"(b.year, w.week_number, p.canonical_name) > ($1, $2, $3)",
				"LIMIT $4",
			},
			wantParams: []interface{}{1665, 12, "All Hallows", 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := buildBillsQueryWithParams(tt.params)
			if err != nil {
				t.Fatalf("buildBillsQueryWithParams: %v", err)
			}
			for _, fragment := range tt.queryContains {
				if !strings.Contains(query.Query, fragment) {
					t.Errorf("query does not contain %q", fragment)
				}
			}
			if !reflect.DeepEqual(query.Params, tt.wantParams) {
				t.Fatalf("params = %#v, want %#v", query.Params, tt.wantParams)
			}
		})
	}
}

func TestBuildBillsQuerySorts(t *testing.T) {
	tests := []struct {
		name    string
		params  APIParameters
		want    string
		wantErr bool
	}{
		{
			name:   "default",
			params: APIParameters{},
			want:   "ORDER BY b.year, w.week_number, p.canonical_name ASC",
		},
		{
			name:   "year",
			params: APIParameters{Sort: "year"},
			want:   "ORDER BY b.year, w.week_number, p.canonical_name ASC",
		},
		{
			name:   "week number",
			params: APIParameters{Sort: "week_number"},
			want:   "ORDER BY w.week_number, b.year, p.canonical_name ASC",
		},
		{
			name:   "canonical name",
			params: APIParameters{Sort: "canonical_name"},
			want:   "ORDER BY p.canonical_name, b.year, w.week_number ASC",
		},
		{
			name: "cursor retains continuation order",
			params: APIParameters{
				Sort:       "canonical_name",
				Cursor:     "cursor",
				CursorYear: 1665,
				CursorWeek: 12,
				CursorName: "All Hallows",
			},
			want: "ORDER BY b.year, w.week_number, p.canonical_name ASC",
		},
		{
			name:    "rejects unvalidated caller",
			params:  APIParameters{Sort: "year; select pg_sleep(10)"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := buildBillsQueryWithParams(tt.params)
			if tt.wantErr {
				if err == nil {
					t.Fatal("buildBillsQueryWithParams returned nil error")
				}
				return
			}
			if err != nil {
				t.Fatalf("buildBillsQueryWithParams: %v", err)
			}
			if !strings.Contains(query.Query, tt.want) {
				t.Fatalf("query does not contain %q", tt.want)
			}
		})
	}
}

func TestBuildStatisticsQueries(t *testing.T) {
	if query := buildWeeklyStatsQuery(); !strings.Contains(query, "year_week_range") {
		t.Fatalf("weekly statistics query missing generated year/week range")
	}

	allParishes, err := buildParishYearlyStatsQuery("")
	if err != nil {
		t.Fatalf("build all-parish statistics query: %v", err)
	}
	if len(allParishes.Params) != 0 {
		t.Fatalf("all-parish params = %#v, want none", allParishes.Params)
	}

	filtered, err := buildParishYearlyStatsQuery("All Hallows")
	if err != nil {
		t.Fatalf("build filtered parish statistics query: %v", err)
	}
	if !strings.Contains(filtered.Query, "LOWER(p.canonical_name) = LOWER($1)") {
		t.Fatal("filtered parish query does not use a parameter placeholder")
	}
	if !reflect.DeepEqual(filtered.Params, []interface{}{"All Hallows"}) {
		t.Fatalf("filtered parish params = %#v, want All Hallows", filtered.Params)
	}
}
