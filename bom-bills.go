package apiary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v4"
)

// ParishByYear describes a parish's canoncial name, count type, total count, start day,
// start month, end day, end month, year, week number, and week ID.
type ParishByYear struct {
	ParishName string     `json:"name"`
	BillType   string     `json:"bill_type"`
	CountType  string     `json:"count_type"`
	TotalCount NullInt64  `json:"count"`
	StartDay   NullInt64  `json:"start_day"`
	StartMonth NullString `json:"start_month"`
	EndDay     NullInt64  `json:"end_day"`
	EndMonth   NullString `json:"end_month"`
	Year       int        `json:"year"`
	SplitYear  string     `json:"split_year"`
	WeekNo     int        `json:"week_no"`
	WeekID     string     `json:"week_id"`
}

// TotalBills returns to the total number of records in the database. We need this
// number to get pagination working.
type TotalBills struct {
	TotalRecords NullInt64 `json:"total_records"`
}

// BillsHandler returns the bills for a given range of years. It expects a start year and
// an end year. It returns a JSON array of ParishByYear objects.
func (s *Server) BillsHandler() http.HandlerFunc {

	// Query for specific bill types and count
	query := `
	SELECT
		p.canonical_name,
		b.bill_type,
		b.count_type,
		b.count,
		w.start_day,
		w.start_month,
		w.end_day,
		w.end_month,
		y.year,
		y.split_year,
		w.week_no,
		b.week_id
	FROM
		bom.bill_of_mortality b
	JOIN
		bom.parishes p ON p.id = b.parish_id
	JOIN
		bom.year y ON y.year_id = b.year_id
	JOIN
		bom.week w ON w.week_id = b.week_id
	WHERE
		year >= $1
		AND year <= $2
		AND b.bill_type = $3
		AND b.count_type = $4
	ORDER BY
		year ASC,
		week_no ASC,
		canonical_name ASC
	LIMIT $5
	OFFSET $6;
	`
	// Query for all bills (weekly and general) and a specific count type (plague or buried).
	queryAllBillTypes := `
	SELECT
		p.canonical_name,
		'All' AS bill_type,
		b.count_type,
		b.count,
		w.start_day,
		w.start_month,
		w.end_day,
		w.end_month,
		y.year,
		y.split_year,
		w.week_no,
		b.week_id
	FROM
		bom.bill_of_mortality b
	JOIN
		bom.parishes p ON p.id = b.parish_id
	JOIN
		bom.year y ON y.year_id = b.year_id
	JOIN
		bom.week w ON w.week_id = b.week_id
	WHERE
		year >= $1
		AND year <= $2
		AND count_type = $3
	ORDER BY
		year ASC,
		week_no ASC,
		canonical_name ASC
	LIMIT $4
	OFFSET $5;
	`
	// Query for all count types (plague and buried) and a specific bill type (plague or buried).
	queryAllCountTypes := `
	SELECT
		p.canonical_name,
		b.bill_type,
		'All' AS count_type,
		b.count,
		w.start_day,
		w.start_month,
		w.end_day,
		w.end_month,
		y.year,
		y.split_year,
		w.week_no,
		b.week_id
	FROM
		bom.bill_of_mortality b
	JOIN
		bom.parishes p ON p.id = b.parish_id
	JOIN
		bom.year y ON y.year_id = b.year_id
	JOIN
		bom.week w ON w.week_id = b.week_id
	WHERE
		year >= $1
		AND year <= $2
		AND bill_type = $3
	ORDER BY
		year ASC,
		week_no ASC,
		canonical_name ASC
	LIMIT $4
	OFFSET $5;
	`

	// Default query, return all data with limit and offset.
	queryAll := `
	SELECT
		p.canonical_name,
		b.bill_type,
		b.count_type,
		b.count,
		w.start_day,
		w.start_month,
		w.end_day,
		w.end_month,
		y.year,
		y.split_year,
		w.week_no,
		b.week_id
	FROM
		bom.bill_of_mortality b
	JOIN
		bom.parishes p ON p.id = b.parish_id
	JOIN
		bom.year y ON y.year_id = b.year_id
	JOIN
		bom.week w ON w.week_id = b.week_id
	WHERE
		year >= $1
		AND year <= $2
	ORDER BY
		year ASC,
		week_no ASC,
		canonical_name ASC
	LIMIT $3
	OFFSET $4;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		startYear := r.URL.Query().Get("startYear")
		endYear := r.URL.Query().Get("endYear")
		billType := r.URL.Query().Get("bill_type")
		countType := r.URL.Query().Get("count_type")
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		if startYear == "" || endYear == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		startYearInt, err := strconv.Atoi(startYear)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		endYearInt, err := strconv.Atoi(endYear)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if limit == "" {
			limit = "10000"
		}
		if offset == "" {
			offset = "0"
		}

		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// If billType is supplied, it can only be one of the following:
		// "Weekly", "General", or "All".
		if billType != "" && billType != "Weekly" && billType != "General" && billType != "All" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// If countType is supplied, it can only be one of the following:
		// "All records", "Buried", "Plague", or "All".
		if countType != "" && countType != "All" && countType != "Total" && countType != "Buried" && countType != "Plague" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// TODO: We want the ability to sort the following columns:
		// 1. Parish name (canonical_name)
		// 2. Week number (week_no)
		// 3. Year (year)
		// 4. Count (count)

		results := make([]ParishByYear, 0)
		var row ParishByYear
		var rows pgx.Rows

		switch {
		// The following returns the data based on user choices:

		// 1. Bill type (weekly or general) and count type (buried or plague) are specifically set
		// 		GET /bom/bills?startYear=1669&endYear=1754&bill_type=Weekly&count_type=Buried&limit=50&offset=0
		// 		GET /bom/bills?startYear=1669&endYear=1754&bill_type=General&count_type=Total&limit=50&offset=0

		// 2. Bill type (weekly or general) is set, but count type is not
		// 		GET /bom/bills?startYear=1669&endYear=1754&bill_type=All&limit=50&offset=0
		// 	 	GET/bom/bills?startYear=1669&endYear=1754&bill_type=General&limit=50&offset=0

		// 3. Count type (buried or plague) is set, but bill type is not
		// 		GET /bom/bills?startYear=1669&endYear=1754&count_type=All&limit=50&offset=0
		// 		GET /bom/bills?startYear=1669&endYear=1754&count_type=Buried&limit=50&offset=0

		// 4. Bill type and count type are not set, returns all data -- this is the default
		// 		GET /bom/bills?startYear=1669&endYear=1754&limit=50&offset=0

		case billType != "" && countType != "":
			rows, err = s.DB.Query(context.TODO(), query, startYearInt, endYearInt, billType, countType, limitInt, offsetInt)
		case billType != "" && countType == "":
			rows, err = s.DB.Query(context.TODO(), queryAllCountTypes, startYearInt, endYearInt, billType, limitInt, offsetInt)
		case countType != "" && billType == "":
			rows, err = s.DB.Query(context.TODO(), queryAllBillTypes, startYearInt, endYearInt, countType, limitInt, offsetInt)
		default:
			rows, err = s.DB.Query(context.TODO(), queryAll, startYearInt, endYearInt, limitInt, offsetInt)
		}

		if err != nil {
			log.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(
				&row.ParishName,
				&row.BillType,
				&row.CountType,
				&row.TotalCount,
				&row.StartDay,
				&row.StartMonth,
				&row.EndDay,
				&row.EndMonth,
				&row.Year,
				&row.SplitYear,
				&row.WeekNo,
				&row.WeekID)
			if err != nil {
				log.Println(err)
			}
			results = append(results, row)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, _ := json.Marshal(results)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}

}

// TotalBillsHandler returns the total number of bills in the database.
// This number is required for pagination in the web application.
func (s *Server) TotalBillsHandler() http.HandlerFunc {

	query := `
	SELECT
		COUNT(*)
	FROM
		bom.bill_of_mortality;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		var totalBills TotalBills

		row := s.DB.QueryRow(context.TODO(), query)
		err := row.Scan(&totalBills.TotalRecords)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, _ := json.Marshal(totalBills)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}
}
