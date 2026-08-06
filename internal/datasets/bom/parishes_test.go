package bom

import (
	"reflect"
	"testing"
)

func TestMissingIDs(t *testing.T) {
	found := map[int]bool{1: true, 5: true}
	cases := []struct {
		name      string
		requested []int
		want      []int
	}{
		{"all found", []int{1, 5}, nil},
		{"some missing", []int{1, 2, 5, 9}, []int{2, 9}},
		{"duplicates reported once", []int{2, 2}, []int{2}},
		{"empty request", nil, nil},
	}
	for _, c := range cases {
		got := missingIDs(c.requested, found)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: missingIDs(%v) = %v, want %v", c.name, c.requested, got, c.want)
		}
	}
}
