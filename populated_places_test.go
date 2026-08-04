package apiary

import (
	"encoding/json"
	"testing"
)

func TestPopulatedPlaceJSONContract(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{
			name: "county place",
			value: Place{
				PlaceID: 611119,
				Place:   "Groton",
				Lat:     42.6112,
				Lon:     -71.5745,
			},
		},
		{
			name: "place details",
			value: PlaceDetails{
				PlaceID:    611119,
				Place:      "Groton",
				Lat:        42.6112,
				Lon:        -71.5745,
				County:     "Middlesex",
				CountyAHCB: "mas_middlesex",
				State:      "MA",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("marshal populated place: %v", err)
			}

			var fields map[string]any
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatalf("unmarshal populated place: %v", err)
			}
			if _, ok := fields["map_name"]; ok {
				t.Fatalf("populated place contains removed map_name field: %s", encoded)
			}
			if got := fields["lat"]; got != 42.6112 {
				t.Errorf("lat = %v, want 42.6112", got)
			}
			if got := fields["lon"]; got != -71.5745 {
				t.Errorf("lon = %v, want -71.5745", got)
			}
		})
	}
}
