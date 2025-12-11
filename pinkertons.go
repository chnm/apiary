package apiary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Activity represents a detective activity from the database
type Activity struct {
	ID              int        `json:"id"`
	Source          NullString `json:"source"`
	Operative       NullString `json:"operative"`
	Date            NullString `json:"date"`
	Time            NullString `json:"time"`
	Duration        NullString `json:"duration"`
	Roping          NullString `json:"roping"`
	Mode            NullString `json:"mode"`
	ActivityNotes   NullString `json:"activity_notes"`
	Subject         NullString `json:"subject"`
	Information     NullString `json:"information"`
	InformationType NullString `json:"information_type"`
	Edited          NullString `json:"edited"`
	EditType        NullString `json:"edit_type"`
	Locations       []Location `json:"locations,omitempty"`
}

// Location represents a location in the database
type Location struct {
	ID            int         `json:"id"`
	Locality      NullString  `json:"locality"`
	StreetAddress NullString  `json:"street_address"`
	LocationName  NullString  `json:"location_name"`
	LocationType  NullString  `json:"location_type"`
	LocationNotes NullString  `json:"location_notes"`
	Latitude      NullFloat64 `json:"latitude"`
	Longitude     NullFloat64 `json:"longitude"`
}

// NullFloat64 handles nullable float64 values for JSON marshaling
type NullFloat64 struct {
	Float64 float64
	Valid   bool
}

// MarshalJSON marshals a null float64 as null instead of 0
func (v NullFloat64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Float64)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals a null float64 from JSON
func (v *NullFloat64) UnmarshalJSON(data []byte) error {
	var x *float64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Float64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

// ActivitiesHandler returns a list of all detective activities with optional filtering and location inclusion
// Query parameters:
//   - include_locations: if "true", includes location data with coordinates
//   - operative: filter by operative name
//   - subject: filter by subject name
//   - start_date: filter by start date (YYYY-MM-DD)
//   - end_date: filter by end date (YYYY-MM-DD)
func (s *Server) ActivitiesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		includeLocations := r.URL.Query().Get("include_locations") == "true"
		operative := r.URL.Query().Get("operative")
		subject := r.URL.Query().Get("subject")
		startDate := r.URL.Query().Get("start_date")
		endDate := r.URL.Query().Get("end_date")

		// Build query dynamically based on parameters
		baseQuery := `
		SELECT
			a.id, a.source, a.operative, a.date, a.time, a.duration,
			a.roping, a.mode, a.activity_notes, a.subject, a.information,
			a.information_type, a.edited, a.edit_type
		FROM detectives.activities a
		WHERE 1=1
		`

		// Add filters
		args := make([]interface{}, 0)
		argCount := 1

		if operative != "" {
			baseQuery += fmt.Sprintf(" AND a.operative = $%d", argCount)
			args = append(args, operative)
			argCount++
		}

		if subject != "" {
			baseQuery += fmt.Sprintf(" AND a.subject = $%d", argCount)
			args = append(args, subject)
			argCount++
		}

		if startDate != "" {
			baseQuery += fmt.Sprintf(" AND a.date >= $%d", argCount)
			args = append(args, startDate)
			argCount++
		}

		if endDate != "" {
			baseQuery += fmt.Sprintf(" AND a.date <= $%d", argCount)
			args = append(args, endDate)
			argCount++
		}

		baseQuery += " ORDER BY a.date, a.time;"

		results := make([]Activity, 0)

		rows, err := s.DB.Query(context.TODO(), baseQuery, args...)
		if err != nil {
			log.Println("Error querying activities:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var row Activity
			err := rows.Scan(
				&row.ID, &row.Source, &row.Operative, &row.Date, &row.Time,
				&row.Duration, &row.Roping, &row.Mode, &row.ActivityNotes,
				&row.Subject, &row.Information, &row.InformationType,
				&row.Edited, &row.EditType,
			)
			if err != nil {
				log.Println("Error scanning activity row:", err)
				continue
			}
			results = append(results, row)
		}

		if err = rows.Err(); err != nil {
			log.Println("Error iterating activities:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// If include_locations is true, fetch locations for each activity
		if includeLocations {
			locationsQuery := `
			SELECT
				l.id, l.locality, l.street_address, l.location_name,
				l.location_type, l.location_notes, l.latitude, l.longitude
			FROM detectives.locations l
			INNER JOIN detectives.activity_locations al ON l.id = al.location_id
			WHERE al.activity_id = $1;
			`

			for i := range results {
				results[i].Locations = make([]Location, 0)
				locRows, err := s.DB.Query(context.TODO(), locationsQuery, results[i].ID)
				if err != nil {
					log.Println("Error querying locations for activity", results[i].ID, ":", err)
					continue
				}

				for locRows.Next() {
					var loc Location
					err := locRows.Scan(
						&loc.ID, &loc.Locality, &loc.StreetAddress, &loc.LocationName,
						&loc.LocationType, &loc.LocationNotes, &loc.Latitude, &loc.Longitude,
					)
					if err != nil {
						log.Println("Error scanning location row:", err)
						continue
					}
					results[i].Locations = append(results[i].Locations, loc)
				}
				locRows.Close()
			}
		}

		response, err := json.Marshal(results)
		if err != nil {
			log.Println("Error marshaling JSON:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}
}

// ActivityByIDHandler returns a single activity with its locations
func (s *Server) ActivityByIDHandler() http.HandlerFunc {
	activityQuery := `
	SELECT
		a.id, a.source, a.operative, a.date, a.time, a.duration,
		a.roping, a.mode, a.activity_notes, a.subject, a.information,
		a.information_type, a.edited, a.edit_type
	FROM detectives.activities a
	WHERE a.id = $1;
	`

	locationsQuery := `
	SELECT
		l.id, l.locality, l.street_address, l.location_name,
		l.location_type, l.location_notes, l.latitude, l.longitude
	FROM detectives.locations l
	INNER JOIN detectives.activity_locations al ON l.id = al.location_id
	WHERE al.activity_id = $1;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid activity ID", http.StatusBadRequest)
			return
		}

		var activity Activity

		// Get activity
		err = s.DB.QueryRow(context.TODO(), activityQuery, id).Scan(
			&activity.ID, &activity.Source, &activity.Operative, &activity.Date,
			&activity.Time, &activity.Duration, &activity.Roping, &activity.Mode,
			&activity.ActivityNotes, &activity.Subject, &activity.Information,
			&activity.InformationType, &activity.Edited, &activity.EditType,
		)
		if err != nil {
			log.Println("Error querying activity:", err)
			http.Error(w, "Activity not found", http.StatusNotFound)
			return
		}

		// Get locations for this activity
		activity.Locations = make([]Location, 0)
		rows, err := s.DB.Query(context.TODO(), locationsQuery, id)
		if err != nil {
			log.Println("Error querying locations:", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var loc Location
				err := rows.Scan(
					&loc.ID, &loc.Locality, &loc.StreetAddress, &loc.LocationName,
					&loc.LocationType, &loc.LocationNotes, &loc.Latitude, &loc.Longitude,
				)
				if err != nil {
					log.Println("Error scanning location row:", err)
					continue
				}
				activity.Locations = append(activity.Locations, loc)
			}
		}

		response, err := json.Marshal(activity)
		if err != nil {
			log.Println("Error marshaling JSON:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}
}

// LocationsHandler returns all locations with coordinates
func (s *Server) LocationsHandler() http.HandlerFunc {
	query := `
	SELECT
		l.id, l.locality, l.street_address, l.location_name,
		l.location_type, l.location_notes, l.latitude, l.longitude
	FROM detectives.locations l
	ORDER BY l.locality, l.location_name;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]Location, 0)

		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			log.Println("Error querying locations:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var row Location
			err := rows.Scan(
				&row.ID, &row.Locality, &row.StreetAddress, &row.LocationName,
				&row.LocationType, &row.LocationNotes, &row.Latitude, &row.Longitude,
			)
			if err != nil {
				log.Println("Error scanning location row:", err)
				continue
			}
			results = append(results, row)
		}

		if err = rows.Err(); err != nil {
			log.Println("Error iterating locations:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(results)
		if err != nil {
			log.Println("Error marshaling JSON:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}
}

// OperativesHandler returns a list of unique operatives
func (s *Server) OperativesHandler() http.HandlerFunc {
	query := `
	SELECT DISTINCT operative
	FROM detectives.activities
	WHERE operative IS NOT NULL
	ORDER BY operative;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]string, 0)

		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			log.Println("Error querying operatives:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var operative string
			err := rows.Scan(&operative)
			if err != nil {
				log.Println("Error scanning operative:", err)
				continue
			}
			results = append(results, operative)
		}

		if err = rows.Err(); err != nil {
			log.Println("Error iterating operatives:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(results)
		if err != nil {
			log.Println("Error marshaling JSON:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}
}

// SubjectsHandler returns a list of unique subjects
func (s *Server) SubjectsHandler() http.HandlerFunc {
	query := `
	SELECT DISTINCT subject
	FROM detectives.activities
	WHERE subject IS NOT NULL
	ORDER BY subject;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]string, 0)

		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			log.Println("Error querying subjects:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var subject string
			err := rows.Scan(&subject)
			if err != nil {
				log.Println("Error scanning subject:", err)
				continue
			}
			results = append(results, subject)
		}

		if err = rows.Err(); err != nil {
			log.Println("Error iterating subjects:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(results)
		if err != nil {
			log.Println("Error marshaling JSON:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}
}
