package dataapi

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

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
