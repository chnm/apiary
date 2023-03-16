package apiary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
)

// ParishByYear describes a parish's canoncial name, count type, total count, start day,
// start month, end day, end month, year, week number, and week ID.
type ParishByYear struct {
	ParishName   string     `json:"name"`
	BillType     string     `json:"bill_type"`
	CountType    string     `json:"count_type"`
	TotalCount   NullInt64  `json:"count"`
	StartDay     NullInt64  `json:"start_day"`
	StartMonth   NullString `json:"start_month"`
	EndDay       NullInt64  `json:"end_day"`
	EndMonth     NullString `json:"end_month"`
	Year         NullInt64  `json:"year"`
	SplitYear    string     `json:"split_year"`
	WeekNo       int        `json:"week_no"`
	WeekID       string     `json:"week_id"`
	TotalRecords int        `json:"totalrecords"`
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
		w.split_year,
		w.week_no,
		b.week_id,
		COUNT(*) OVER() AS totalrecords
	FROM
		bom.bill_of_mortality b
	JOIN
		bom.parishes p ON p.id = b.parish_id
	JOIN
		bom.year y ON y.year = b.year_id
	JOIN
		bom.week w ON w.joinid = b.week_id
	WHERE
		y.year >= $1::int
		AND y.year <= $2::int
		AND bill_type = $3
		AND count_type = $4
		AND (
			$5::int[] IS NULL
			OR parish_id = ANY($5::int[])
		)
	ORDER BY
		year ASC,
		week_no ASC,
		canonical_name ASC
	LIMIT $6
	OFFSET $7;
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
		w.split_year,
		w.week_no,
		b.week_id,
		COUNT(*) OVER() AS totalrecords
	FROM
		bom.bill_of_mortality b
	JOIN
		bom.parishes p ON p.id = b.parish_id
	JOIN
		bom.year y ON y.year = b.year_id
	JOIN
		bom.week w ON w.joinid = b.week_id
	WHERE
		y.year >= $1::int
		AND y.year <= $2::int
		AND count_type = $3
		AND (
			$4::int[] IS NULL
			OR parish_id = ANY($4::int[])
		)
	ORDER BY
		year ASC,
		week_no ASC,
		canonical_name ASC
	LIMIT $5
	OFFSET $6;
	`

	// Query for all count types (plague and buried) and a specific bill type (plague or buried).
	queryAllCountTypes := `
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
		w.split_year,
		w.week_no,
		b.week_id,
		COUNT(*) OVER() AS totalrecords
	FROM
		bom.bill_of_mortality b
	JOIN
		bom.parishes p ON p.id = b.parish_id
	JOIN
		bom.year y ON y.year = b.year_id
	JOIN
		bom.week w ON w.joinid = b.week_id
	WHERE
		y.year >= $1::int
		AND y.year <= $2::int
		AND bill_type = $3
		AND (
			$4::int[] IS NULL
			OR parish_id = ANY($4::int[])
		)
	ORDER BY
		year ASC,
		week_no ASC,
		canonical_name ASC
	LIMIT $5
	OFFSET $6;
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
		w.split_year,
		w.week_no,
		b.week_id,
		COUNT(*) OVER() AS totalrecords
	FROM
		bom.bill_of_mortality b
	JOIN
		bom.parishes p ON p.id = b.parish_id
	JOIN
		bom.year y ON y.year = b.year_id
	JOIN
		bom.week w ON w.joinid = b.week_id
	WHERE
		y.year >= $1::int
		AND y.year <= $2::int
		AND (
			$3::int[] IS NULL
			OR parish_id = ANY($3::int[])
		)
	ORDER BY
		year ASC,
		week_no ASC,
		canonical_name ASC
	LIMIT $4
	OFFSET $5;
	`

	// Query for specific bill types and count but no parishes.
	queryNoParishes := `
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
		w.split_year,
		w.week_no,
		b.week_id,
		COUNT(*) OVER() AS totalrecords
	FROM
		bom.bill_of_mortality b
	JOIN
		bom.parishes p ON p.id = b.parish_id
	JOIN
		bom.year y ON y.year = b.year_id
	JOIN
		bom.week w ON w.joinid = b.week_id
	WHERE
		y.year >= $1::int
		AND y.year <= $2::int
		AND bill_type = $3
		AND count_type = $4
	ORDER BY
		year ASC,
		week_no ASC,
		canonical_name ASC
	LIMIT $5
	OFFSET $6;
	`

	// Query for all bills (weekly and general) and a specific count type (plague or buried) and no parishes.
	queryAllBillTypesNoParishes := `
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
		w.split_year,
		w.week_no,
		b.week_id,
		COUNT(*) OVER() AS totalrecords
	FROM
		bom.bill_of_mortality b
	JOIN
		bom.parishes p ON p.id = b.parish_id
	JOIN
		bom.year y ON y.year = b.year_id
	JOIN
		bom.week w ON w.joinid = b.week_id
	WHERE
		y.year >= $1::int
		AND y.year <= $2::int
		AND count_type = $3
	ORDER BY
		year ASC,
		week_no ASC,
		canonical_name ASC
	LIMIT $4
	OFFSET $5;
	`

	// Query for all count types (plague and buried) and a specific bill type (plague or buried) and no parishes.
	queryAllCountTypesNoParishes := `
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
		w.split_year,
		w.week_no,
		b.week_id,
		COUNT(*) OVER() AS totalrecords
	FROM
		bom.bill_of_mortality b
	JOIN
		bom.parishes p ON p.id = b.parish_id
	JOIN
		bom.year y ON y.year = b.year_id
	JOIN
		bom.week w ON w.joinid = b.week_id
	WHERE
		y.year >= $1::int
		AND y.year <= $2::int
		AND bill_type = $3
	ORDER BY
		year ASC,
		week_no ASC,
		canonical_name ASC
	LIMIT $4
	OFFSET $5;
	`

	// Default query, return all data with limit and offset.
	queryAllNoParishes := `
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
		w.split_year,
		w.week_no,
		b.week_id,
		COUNT(*) OVER() AS totalrecords
	FROM
		bom.bill_of_mortality b
	JOIN
		bom.parishes p ON p.id = b.parish_id
	JOIN
		bom.year y ON y.year = b.year_id
	JOIN
		bom.week w ON w.joinid = b.week_id
	WHERE
		y.year >= $1::int
		AND y.year <= $2::int
	ORDER BY
		year ASC,
		week_no ASC,
		canonical_name ASC
	LIMIT $3
	OFFSET $4;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		// We use hyphen-separated strings since URLs can be case-sensitive.
		// https://www.rfc-editor.org/rfc/rfc3986
		// https://developers.google.com/search/docs/advanced/guidelines/url-structure?hl=en&visit_id=637937657362879240-859683351&rd=1
		startYear := r.URL.Query().Get("start-year")
		endYear := r.URL.Query().Get("end-year")
		billType := r.URL.Query().Get("bill-type")
		countType := r.URL.Query().Get("count-type")
		parishIDs := r.URL.Query().Get("parishes")
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		// parishIDs needs to be a postgres array of integers to send to the query. This
		// returns '{1, 2, 3}' to give the query a literal array.
		parishIDs = fmt.Sprintf("{%s}", strings.TrimSpace(parishIDs))

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
			limit = "25"
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
		// "Total", "Buried", "Plague", or "All".
		if countType != "" && countType != "All" && countType != "Total" && countType != "Buried" && countType != "Plague" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// ParishID must be between the minParishID and maxParishID range. Otherwise, it's a bad request.
		// func validateParishIDs(parishIDs string) (string, error) {
		// 	// We need to check that the parishIDs are within the range of the min and max parish IDs.
		// 	// We do this by checking that the parishIDs are within the range of the min and max parish IDs.
		// 	// currently in the database.
		// 	maxValueQuery := `
		// 	SELECT
		// 		MAX(id)
		// 	FROM
		// 		bom.parishes;
		// 	`
		// 	minValueQuery := `
		// 	SELECT
		// 		MIN(id)
		// 	FROM
		// 		bom.parishes;
		// 	`

		// 	var maxValue int
		// 	var minValue int

		// 	err := db.QueryRow(maxValueQuery).Scan(&maxValue)
		// 	if err != nil {
		// 		return "", err
		// 	}

		// 	err = db.QueryRow(minValueQuery).Scan(&minValue)
		// 	if err != nil {
		// 		return "", err
		// 	}
		// }

		// TODO: We want the ability to sort the following columns:
		// 1. Parish name (canonical_name)
		// 2. Week number (week_no)
		// 3. Year (year)
		// 4. Count (count)
		// Sorting is selected in the frontend table.

		// sortBy := r.URL.Query().Get("sort")
		// if sortBy == "" {
		// 	sortBy = "canonical_name"
		// }
		// sortQuery, err := validateAndReturnSortQuery(sortBy)
		// if err != nil {
		// 	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		// 	return
		// }

		results := make([]ParishByYear, 0)
		var row ParishByYear
		var rows pgx.Rows

		switch {
		// The following returns the data based on user choices:

		// 1. Bill type and count type are not set, returns all data
		// 		GET /bom/bills?start-year=1669&end-year=1754&limit=50&offset=0 --- this returns [] as expected
		// 		GET /bom/bills?start-year=1669&end-year=1754&bill-type=All&count-type=All&parishes=5&limit=50&offset=0
		// 		GET /bom/bills?start-year=1669&end-year=1754&bill-type=All&count-type=All&limit=50&offset=0
		//
		// 2. Bill type (weekly or general) is set, but count type is not
		// 		GET /bom/bills?start-year=1669&end-year=1754&bill-type=Weekly&limit=50&offset=0
		// 		GET /bom/bills?start-year=1669&end-year=1754&bill-type=Weekly&parishes=1,5,19,67,77&limit=50&offset=0
		//
		// 3. Count type (buried or plague) is set, but bill type is not
		// 		GET /bom/bills?start-year=1669&end-year=1754&count-type=Buried&limit=50&offset=0
		// 		GET /bom/bills?start-year=1669&end-year=1754&count-type=Buried&parishes=2&limit=50&offset=0
		//
		// 4. Bill type (weekly or general) and count type (buried or plague) are specifically set
		// 		GET /bom/bills?start-year=1669&end-year=1754&bill-type=Weekly&count-type=Buried&limit=50&offset=0
		// 		GET /bom/bills?start-year=1669&end-year=1754&bill-type=General&count-type=Total&limit=50&offset=0
		// 		GET /bom/bills?start-year=1669&end-year=1754&bill-type=Weekly&count-type=Plague&parishes=1,5,9&limit=50&offset=0

		// If parish ids are provided:
		case billType == "All" && countType == "All" && parishIDs != "{}":
			rows, err = s.DB.Query(context.TODO(), queryAll, startYearInt, endYearInt, parishIDs, limitInt, offsetInt)
		case countType != "" && billType != "" && parishIDs != "{}":
			rows, err = s.DB.Query(context.TODO(), query, startYearInt, endYearInt, billType, countType, parishIDs, limitInt, offsetInt)
		case billType != "" && countType == "" && parishIDs != "{}":
			rows, err = s.DB.Query(context.TODO(), queryAllCountTypes, startYearInt, endYearInt, billType, parishIDs, limitInt, offsetInt)
		case countType != "" && billType == "" && parishIDs != "{}":
			rows, err = s.DB.Query(context.TODO(), queryAllBillTypes, startYearInt, endYearInt, countType, parishIDs, limitInt, offsetInt)

		// If parish ids are not provided:
		case billType == "All" && countType == "All" && parishIDs == "{}":
			rows, err = s.DB.Query(context.TODO(), queryAllNoParishes, startYearInt, endYearInt, limitInt, offsetInt)
		case countType != "" && billType != "" && parishIDs == "{}":
			rows, err = s.DB.Query(context.TODO(), queryNoParishes, startYearInt, endYearInt, billType, countType, limitInt, offsetInt)
		case billType != "" && countType == "" && parishIDs == "{}":
			rows, err = s.DB.Query(context.TODO(), queryAllCountTypesNoParishes, startYearInt, endYearInt, billType, limitInt, offsetInt)
		case countType != "" && billType == "" && parishIDs == "{}":
			rows, err = s.DB.Query(context.TODO(), queryAllBillTypesNoParishes, startYearInt, endYearInt, countType, limitInt, offsetInt)

		default:
			rows, err = s.DB.Query(context.TODO(), query, startYearInt, endYearInt, billType, countType, parishIDs, limitInt, offsetInt)
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
				&row.WeekID,
				&row.TotalRecords)
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

	queryWeekly := `
	SELECT
		COUNT(*)
	FROM
		bom.bill_of_mortality
	WHERE 
		bill_type = 'Weekly';
	`

	queryGeneral := `
	SELECT
		COUNT(*)
	FROM
		bom.bill_of_mortality
	WHERE	
		bill_type = 'General';
	`

	queryChristenings := `
	SELECT
		COUNT(*)
	FROM	
		bom.christenings;
	`

	queryCauses := `
	SELECT
		COUNT(*)
	FROM 
		bom.causes_of_death;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		totalValues := r.URL.Query().Get("type")

		if totalValues == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		results := make([]TotalBills, 0)
		var row TotalBills
		var rows pgx.Rows
		var err error

		switch {
		case totalValues == "Weekly":
			rows, err = s.DB.Query(context.TODO(), queryWeekly)
		case totalValues == "General":
			rows, err = s.DB.Query(context.TODO(), queryGeneral)
		case totalValues == "Christenings":
			rows, err = s.DB.Query(context.TODO(), queryChristenings)
		case totalValues == "Causes":
			rows, err = s.DB.Query(context.TODO(), queryCauses)
		}
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&row.TotalRecords)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			results = append(results, row)
		}

		response, _ := json.Marshal(results)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}
}
