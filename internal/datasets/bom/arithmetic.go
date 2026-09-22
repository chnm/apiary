package bom

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// WeeklyArithmetic compares a week's printed subtotals with the sum of its
// parish counts. MixedCopies is true when the week's figures come from more
// than one transcribed copy of the bill.
type WeeklyArithmetic struct {
	Year        int    `json:"year"`
	WeekNumber  int    `json:"week_number"`
	WeekID      string `json:"week_id"`
	CountType   string `json:"count_type"`
	SubtotalSum int64  `json:"subtotal_sum"`
	ParishSum   int64  `json:"parish_sum"`
	Difference  int64  `json:"difference"`
	Legible     bool   `json:"legible"`
	MixedCopies bool   `json:"mixed_copies"`
}

type arithmeticParameters struct {
	StartYear int
	EndYear   int
	CountType any
	Legible   any
}

func parseArithmeticParameters(r *http.Request) (arithmeticParameters, error) {
	params := arithmeticParameters{StartYear: 1636, EndYear: 1754}
	query := r.URL.Query()

	if value := query.Get("start-year"); value != "" {
		year, err := strconv.Atoi(value)
		if err != nil {
			return params, fmt.Errorf("invalid start year: %q", value)
		}
		params.StartYear = year
	}
	if value := query.Get("end-year"); value != "" {
		year, err := strconv.Atoi(value)
		if err != nil {
			return params, fmt.Errorf("invalid end year: %q", value)
		}
		params.EndYear = year
	}
	if params.StartYear > params.EndYear {
		return params, fmt.Errorf("start year must not be after end year")
	}
	if value := query.Get("count-type"); value != "" {
		if !IsValidCountType(value) {
			return params, fmt.Errorf("count type must be 'buried' or 'plague'")
		}
		params.CountType = strings.ToLower(value)
	}
	if value := query.Get("legible"); value != "" {
		legible, err := strconv.ParseBool(value)
		if err != nil {
			return params, fmt.Errorf("legible must be true or false")
		}
		params.Legible = legible
	}
	return params, nil
}

// ArithmeticHandler returns printed weekly subtotals compared with the sum of
// parish counts for each week. Optional query parameters: start-year and
// end-year (inclusive), count-type (buried or plague), and legible
// (true or false).
func (h *Handler) ArithmeticHandler() http.HandlerFunc {
	query := `
	SELECT
		year,
		week_number,
		week_id,
		count_type,
		subtotal_sum,
		parish_sum,
		difference,
		legible,
		mixed_copies
	FROM
		bom.weekly_arithmetic
	WHERE
		year BETWEEN $1::int AND $2::int
		AND ($3::text IS NULL OR count_type = $3::text)
		AND ($4::boolean IS NULL OR legible = $4::boolean)
	ORDER BY
		year, week_number, week_id, count_type;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		params, err := parseArithmeticParameters(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rows, err := h.db.Query(r.Context(), query, params.StartYear, params.EndYear, params.CountType, params.Legible)
		if err != nil {
			internalServerError(w, "error querying weekly arithmetic", err)
			return
		}
		defer rows.Close()

		results := make([]WeeklyArithmetic, 0)
		var row WeeklyArithmetic
		for rows.Next() {
			if err := rows.Scan(
				&row.Year,
				&row.WeekNumber,
				&row.WeekID,
				&row.CountType,
				&row.SubtotalSum,
				&row.ParishSum,
				&row.Difference,
				&row.Legible,
				&row.MixedCopies,
			); err != nil {
				internalServerError(w, "error scanning weekly arithmetic", err)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			internalServerError(w, "error iterating weekly arithmetic", err)
			return
		}

		writeJSONResponse(w, results)
	}
}
