package act_test

import (
	"flag"
	"net/url"
	"reflect"
	"testing"
	"time"

	"go.ectobit.com/act"
)

var (
	_ flag.Getter = (*act.StringSlice)(nil)
	_ flag.Getter = (*act.IntSlice)(nil)
	_ flag.Getter = (*act.URL)(nil)
	_ flag.Getter = (*act.Time)(nil)
)

func TestStringSlice(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in         string
		wantString string
		wantGet    []string
	}{
		"empty": {
			"",
			"[]",
			[]string{},
		},
		"one": {
			"foo",
			"['foo']",
			[]string{"foo"},
		},
		"multi": {
			"foo,bar,baz",
			"['foo','bar','baz']",
			[]string{"foo", "bar", "baz"},
		},
	}

	for n, tt := range tests { //nolint:paralleltest
		n := n
		tt := tt

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			f := &act.StringSlice{}

			if err := f.Set(tt.in); err != nil {
				t.Error(err)
			}

			stringWant := f.String()
			if stringWant != tt.wantString {
				t.Errorf("want %s got %s", tt.wantString, stringWant)
			}

			getWant := f.Get()
			if !reflect.DeepEqual(getWant, tt.wantGet) {
				t.Errorf("want %v got %v", tt.wantGet, getWant)
			}
		})
	}
}

func TestIntSlice(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := map[string]struct {
		in         string
		wantString string
		wantGet    []int
		wantErr    bool
	}{
		"empty": {
			"",
			"[]",
			[]int{},
			false,
		},
		"valid-one": {
			"1",
			"[1]",
			[]int{1},
			false,
		},
		"multi": {
			"1,2,3",
			"[1,2,3]",
			[]int{1, 2, 3},
			false,
		},
		"invalid-one": {
			"foo",
			"[]",
			[]int{},
			true,
		},
		"invalid multi": {
			"1,foo,2",
			"[]",
			[]int{},
			true,
		},
	}

	for n, tt := range tests { //nolint:paralleltest
		n := n
		tt := tt

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			f := &act.IntSlice{}

			err := f.Set(tt.in)
			if err != nil {
				if !tt.wantErr {
					t.Error("want error got no error")
				}

				return
			}

			gotString := f.String()
			if gotString != tt.wantString {
				t.Errorf("want %s got %s", tt.wantString, gotString)
			}

			gotGet := f.Get()
			if !reflect.DeepEqual(gotGet, tt.wantGet) {
				t.Errorf("want %v got %v", tt.wantGet, gotGet)
			}
		})
	}
}

func TestURL(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := map[string]struct {
		in         string
		wantString string
		wantGet    url.URL
		wantErr    bool
	}{
		"empty": {
			"",
			"",
			url.URL{}, //nolint:exhaustruct
			false,
		},
		"path": {
			"foo.bar",
			"foo.bar",
			url.URL{Path: "foo.bar"}, //nolint:exhaustruct
			false,
		},
		"host": {
			"//foo.bar",
			"//foo.bar",
			url.URL{Host: "foo.bar"}, //nolint:exhaustruct
			false,
		},
		"full": {
			"https://foo.bar/baz?qux=1",
			"https://foo.bar/baz?qux=1",
			url.URL{Scheme: "https", Host: "foo.bar", Path: "/baz", RawQuery: "qux=1"}, //nolint:exhaustruct
			false,
		},
		"invalid": {
			"%",
			"",
			url.URL{}, //nolint:exhaustruct
			true,
		},
	}

	for n, tt := range tests { //nolint:paralleltest
		n := n
		tt := tt

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			f := &act.URL{} //nolint:exhaustruct

			err := f.Set(tt.in)
			if err != nil {
				if !tt.wantErr {
					t.Error("want error got no error")
				}

				return
			}

			gotString := f.String()
			if gotString != tt.wantString {
				t.Errorf("want %s got %s", tt.wantString, gotString)
			}

			gotGet := f.Get()
			if !reflect.DeepEqual(gotGet, tt.wantGet) {
				t.Errorf("want %v got %v", tt.wantGet, gotGet)
			}
		})
	}
}

func TestTime(t *testing.T) { //nolint:funlen
	t.Parallel()

	at1, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	if err != nil {
		t.Error(err)
	}

	at2, err := time.Parse(time.RFC3339, "2002-10-02T15:00:00Z")
	if err != nil {
		t.Error(err)
	}

	at3, err := time.Parse(time.RFC3339, "2002-10-02T15:00:00.05Z")
	if err != nil {
		t.Error(err)
	}

	tests := map[string]struct {
		in         string
		wantString string
		wantGet    time.Time
		wantErr    bool
	}{
		"empty": {
			"",
			"",
			time.Time{},
			true,
		},
		"format_1": {
			"2002-10-02T10:00:00-05:00",
			"2002-10-02T10:00:00-05:00",
			at1,
			false,
		},
		"format_2": {
			"2002-10-02T15:00:00Z",
			"2002-10-02T15:00:00Z",
			at2,
			false,
		},
		"format_3": {
			"2002-10-02T15:00:00.05Z",
			"2002-10-02T15:00:00Z",
			at3,
			false,
		},
	}

	for n, tt := range tests { //nolint:paralleltest
		n := n
		tt := tt

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			f := &act.Time{} //nolint:exhaustruct

			err := f.Set(tt.in)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("want no error, got error %v", err)
				}

				return
			}

			gotString := f.String()
			if gotString != tt.wantString {
				t.Errorf("want %s got %s", tt.wantString, gotString)
			}

			gotGet := f.Get()
			if !reflect.DeepEqual(gotGet, tt.wantGet) {
				t.Errorf("want %v got %v", tt.wantGet, gotGet)
			}
		})
	}
}

func TestNils(t *testing.T) {
	t.Parallel()

	ss := (*act.StringSlice)(nil)
	_ = ss.String()

	is := (*act.IntSlice)(nil)
	_ = is.String()

	u := (*act.URL)(nil)
	_ = u.String()

	at := (*act.Time)(nil)
	_ = at.String()
}
