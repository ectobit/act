package act

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// StringSlice implements flag.Getter interface for []string type.
type StringSlice []string

// Set sets flag's value by splitting provided comma separated string.
func (f *StringSlice) Set(s string) error {
	if s == "" {
		return nil
	}

	*f = strings.Split(s, ",")

	return nil
}

// String formats flag's value.
func (f *StringSlice) String() string {
	if f != nil {
		if len(*f) == 0 {
			return "[]"
		}

		return fmt.Sprintf("['%s']", strings.Join(*f, `','`))
	}

	return ""
}

// Get returns flag's value.
func (f *StringSlice) Get() interface{} {
	return []string(*f)
}

// IntSlice implements flag.Getter interface for []int type.
type IntSlice []int

// Set sets flag's value by splitting provided comma separated string.
func (f *IntSlice) Set(s string) error {
	if s == "" {
		return nil
	}

	vs := strings.Split(s, ",")
	*f = make([]int, 0, len(vs))

	for _, v := range vs {
		i, err := strconv.Atoi(v)
		if err != nil {
			*f = []int{}

			return fmt.Errorf("parsing int: %w", err)
		}

		*f = append(*f, i)
	}

	return nil
}

// String formats flag's value.
func (f *IntSlice) String() string {
	if f != nil {
		if len(*f) == 0 {
			return "[]"
		}

		s := make([]string, 0, len(*f))
		for _, i := range *f {
			s = append(s, fmt.Sprintf("%d", i))
		}

		return fmt.Sprintf("[%s]", strings.Join(s, `,`))
	}

	return ""
}

// Get returns flag's value.
func (f *IntSlice) Get() interface{} {
	return []int(*f)
}

// URL implements flag.Getter interface for url.URL type.
type URL struct {
	*url.URL
}

// Set sets flag's value by parsing provided url.
func (f *URL) Set(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("parsing url: %w", err)
	}

	f.URL = u

	return nil
}

// String formats flag's value.
func (f *URL) String() string {
	if f != nil && f.URL != nil {
		return f.URL.String()
	}

	return ""
}

// Get returns flag's value.
func (f *URL) Get() interface{} {
	return *f.URL
}
