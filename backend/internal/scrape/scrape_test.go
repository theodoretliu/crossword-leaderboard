package scrape

import (
	"reflect"
	"testing"
)

func TestUniqueByUser(t *testing.T) {
	in := []Result{{NYTUserID: 3, Name: "c"}, {NYTUserID: 1, Name: "a"}, {NYTUserID: 3, Name: "dup"}, {NYTUserID: 2, Name: "b"}}
	got := uniqueByUser(in)
	want := []Result{{NYTUserID: 1, Name: "a"}, {NYTUserID: 2, Name: "b"}, {NYTUserID: 3, Name: "c"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("uniqueByUser = %+v, want %+v", got, want)
	}
	if in[0].NYTUserID != 3 {
		t.Error("uniqueByUser modified its input")
	}
}

func TestPlaceholders(t *testing.T) {
	if got, want := placeholders(2, 3), "($1, $2, $3), ($4, $5, $6)"; got != want {
		t.Errorf("placeholders(2, 3) = %q, want %q", got, want)
	}
}
