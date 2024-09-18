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
	CanonicalName string     `json:"name"`
	BillType      string     `json:"bill_type"`
	CountType     string     `json:"count_type"`
	Count         NullInt64  `json:"count"`
	StartDay      NullInt64  `json:"start_day"`
	StartMonth    NullString `json:"start_month"`
	EndDay        NullInt64  `json:"end_day"`
	EndMonth      NullString `json:"end_month"`
	Year          NullInt64  `json:"year"`
	SplitYear     string     `json:"split_year"`
	WeekNo        int        `json:"week_no"`
	WeekID        string     `json:"week_id"`
	TotalRecords  int        `json:"totalrecords"`
}

type APIParameters struct {
	StartYear int
	EndYear   int
	Parish    []int
	BillType  string
	CountType string
	Sort      string
}

// TotalBills returns to the total number of records in the database. We need this
// number to get pagination working.
type TotalBills struct {
	TotalRecords NullInt64 `json:"total_records"`
}

// BillsHandler returns the bills for a given range of years. It expects a start year and
// an end year. It returns a JSON array of ParishByYear objects.
func (s *Server) BillsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startYear := r.URL.Query().Get("start-year")
		endYear := r.URL.Query().Get("end-year")
		parish := r.URL.Query().Get("parish")
		billType := r.URL.Query().Get("bill-type")
		countType := r.URL.Query().Get("count-type")
		sortData := r.URL.Query().Get("sort")
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")
		page := r.URL.Query().Get("page")

		// Now we can modify the query if a user has supplied one of the following:
		// 1. A start and end year. If neither are provided, we use the defaults of
		// start year: 1648 and end year of 1750.
		// 2. A count type ("buried" or "plague"). If a user selects "All", we provide all.
		// 3. A bill type ("weekly", or "general", or "total", or "All")
		// 4. A parish name, which we fetch by the parish ID. If a user selects "All", we provide all.

		// Create the default API parameters
		apiParams := APIParameters{
			StartYear: 1648,
			EndYear:   1750,
			Parish:    []int{},
			BillType:  "",
			CountType: "",
			Sort:      "year, week_no, canonical_name",
		}

		// If a start year is provided, update the API parameters
		if startYear != "" {
			startYearInt, err := strconv.Atoi(startYear)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				log.Println("start year is not an integer", err)
				return
			}

			apiParams.StartYear = startYearInt
		}

		// If an end year is provided, update the API parameters
		if endYear != "" {
			endYearInt, err := strconv.Atoi(endYear)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				log.Println("end year is not an integer", err)
				return
			}

			apiParams.EndYear = endYearInt
		}

		// if a parish ID is provided, update the API parameters
		if parish != "" {
			parishList := strings.Split(parish, ",")
			var parishInts []int

			for _, p := range parishList {
				parishInt, err := strconv.Atoi(strings.TrimSpace(p))
				if err != nil {
					http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
					log.Println("parish is not an integer", err)
					return
				}
				parishInts = append(parishInts, parishInt)
			}

			apiParams.Parish = parishInts
		}

		// If a bill type is provided, update the API parameters
		if billType != "" {
			// billtype can only be "Weekly", "General", or "Total"
			if billType != "Weekly" && billType != "General" && billType != "Total" {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				log.Println("bill type is invalid")
				return
			}

			apiParams.BillType = billType
		}

		// If a count type is provided, update the API parameters
		if countType != "" {
			// counttype can only be "Buried" or "Plague"
			if countType != "Buried" && countType != "Plague" {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				log.Println("count type is invalid")
				return
			}

			apiParams.CountType = countType
		}

		// If a sort is provided, update the API parameters
		if sortData != "" {
			apiParams.Sort = sortData
		}

		// Create the query
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
		`

		// Now we can deal with the parameters.
		// If a start year is provided, add it to the query
		if apiParams.StartYear != 0 {
			query += " WHERE b.year_id >= " + strconv.Itoa(apiParams.StartYear)
			// params = append(params, apiParams.StartYear)
		}

		// If an end year is provided, add it to the query
		if apiParams.EndYear != 0 {
			query += " AND b.year_id <= " + strconv.Itoa(apiParams.EndYear)
			// params = append(params, apiParams.EndYear)
		}

		// If a parish is provided, add it to the query
		if len(apiParams.Parish) > 0 {
			query += " AND b.parish_id IN (" + strings.Trim(strings.Join(strings.Fields(fmt.Sprint(apiParams.Parish)), ","), "[]") + ")"
			// params = append(params, apiParams.Parish)
		}

		// If a bill type is provided, add it to the query
		if apiParams.BillType != "" {
			query += " AND b.bill_type = " + "'" + apiParams.BillType + "'"
			// params = append(params, apiParams.BillType)
		}

		// If a count type is provided, add it to the query
		if apiParams.CountType != "" {
			query += " AND b.count_type = " + "'" + apiParams.CountType + "'"
			// params = append(params, apiParams.CountType)
		}

		// If a sort is provided, add it to the query
		if apiParams.Sort != "" {
			query += " ORDER BY " + apiParams.Sort + " ASC"
			// params = append(params, apiParams.Sort)
		}

		// log the query and parameters
		log.Println("query", query)
		log.Println("api parameters ", apiParams)
		// log.Println("params", params)

		// We handle limit and offset for pagination, and allow a page parameter which will be converted to limit and offset
		// If a limit is provided, add it to the query
		if limit != "" {
			limitInt, err := strconv.Atoi(limit)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				log.Println("limit is not an integer", err)
				return
			}

			query += " LIMIT " + strconv.Itoa(limitInt)
		}

		// If an offset is provided, add it to the query
		if offset != "" {
			offsetInt, err := strconv.Atoi(offset)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				log.Println("offset is not an integer", err)
				return
			}

			query += " OFFSET " + strconv.Itoa(offsetInt)
		}

		// If a page is provided, add it to the query
		if page != "" {
			pageInt, err := strconv.Atoi(page)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				log.Println("page is not an integer", err)
				return
			}

			// We implement logic here to figure out the limit and offset that
			// corresponds to a particular page.
			// If the page is 1, the limit is 25 and the offset is 0
			// If the page is 2, the limit is 25 and the offset is 25
			// If the page is 15, the limit is 25 and the offset is 350
			// And so on...
			limitInt := 25 // default limit
			offsetInt := (pageInt - 1) * limitInt

			query += " LIMIT " + strconv.Itoa(limitInt) + " OFFSET " + strconv.Itoa(offsetInt)
		}

		// now we can query the database
		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Fatal("Error preparing statement", err)
			return
		}

		defer rows.Close()

		// create a slice to hold the results
		results := []ParishByYear{}

		// iterate through the rows
		for rows.Next() {
			// create a variable to hold the result
			var result ParishByYear

			// scan the row into the result
			err := rows.Scan(
				&result.CanonicalName,
				&result.BillType,
				&result.CountType,
				&result.Count,
				&result.StartDay,
				&result.StartMonth,
				&result.EndDay,
				&result.EndMonth,
				&result.Year,
				&result.SplitYear,
				&result.WeekNo,
				&result.WeekID,
				&result.TotalRecords,
			)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				log.Fatal("Error scanning row", err)
				return
			}

			// append the result to the results slice
			results = append(results, result)
		}

		// check for errors after iterating through the rows
		err = rows.Err()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Fatal("Error iterating through rows", err)
			return
		}

		// if no results are returned, return a 404
		if len(results) == 0 {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			log.Println("404", err)
			return
		}

		// if results are returned, marshal them into JSON
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
