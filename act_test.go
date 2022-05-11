package act_test

import (
	"bytes"
	"flag"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"go.ectobit.com/act"
)

func TestParse_errors(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in      interface{}
		flags   []string
		wantErr string
	}{
		"config-not-pointer-to-struct": {
			in:      []string{},
			flags:   []string{""},
			wantErr: "invalid config type",
		},
		"unsupported-field-type": {
			in: &struct {
				Port int16
			}{},
			flags:   []string{""},
			wantErr: "Port def: parsing value: type not supported: int16",
		},
		"---help": {
			in: &struct {
				Port int
			}{},
			flags:   []string{"---help"},
			wantErr: "bad flag syntax: ---help",
		},
	}

	for n, tt := range tests { //nolint:paralleltest
		n := n
		tt := tt

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			a := act.New("test", act.WithErrorHandling(flag.ContinueOnError))

			if err := a.Parse(tt.in, tt.flags); err != nil && err.Error() != tt.wantErr {
				t.Errorf("want error %v got error %v", tt.wantErr, err.Error())
			}
		})
	}
}

func TestParse_usage(t *testing.T) { //nolint:funlen
	t.Parallel()

	ws := regexp.MustCompile(`\s+`)

	tests := map[string]struct {
		config  interface{}
		want    string
		wantErr string
	}{
		"string-help-without-def": {
			config: &struct {
				LogLevel string
			}{},
			want: "Usage of test: -log-level string log level (env TEST_LOG_LEVEL)",
		},
		"string-help-with-def": {
			config: &struct {
				LogLevel string `def:"debug"`
			}{},
			want: `Usage of test: -log-level string log level (env TEST_LOG_LEVEL) (default "debug")`,
		},
		"bool-help-without-def": {
			config: &struct {
				Verbose bool
			}{},
			want: "Usage of test: -verbose verbose (env TEST_VERBOSE)",
		},
		"bool-help-with-def": {
			config: &struct {
				Verbose bool `def:"true"`
			}{},
			want: "Usage of test: -verbose verbose (env TEST_VERBOSE) (default true)",
		},
		"bool-help-with-invalid-def": {
			config: &struct {
				Verbose bool `def:"a"`
			}{},
			wantErr: `Verbose def: parsing bool "a": strconv.ParseBool: parsing "a": invalid syntax`,
		},
		"number-help-without-def": {
			config: &struct {
				Port       uint
				Distance   uint64
				Degrees    int
				Difference int64
				Balance    float64
			}{},
			want: `Usage of test:
-balance float balance (env TEST_BALANCE)
-degrees int degrees (env TEST_DEGREES)
-difference int difference (env TEST_DIFFERENCE)
-distance uint distance (env TEST_DISTANCE)
-port uint port (env TEST_PORT)`,
		},
		"number-help-with-def": {
			config: &struct {
				Port       uint    `def:"1"`
				Distance   uint64  `def:"1"`
				Degrees    int     `def:"1"`
				Difference int64   `def:"1"`
				Balance    float64 `def:"1"`
			}{},
			want: `Usage of test:
-balance float balance (env TEST_BALANCE) (default 1)
-degrees int degrees (env TEST_DEGREES) (default 1)
-difference int difference (env TEST_DIFFERENCE) (default 1)
-distance uint distance (env TEST_DISTANCE) (default 1)
-port uint port (env TEST_PORT) (default 1)`,
		},
		"uint-help-with-invalid-def": {
			config: &struct {
				Port uint `def:"a"`
			}{},
			wantErr: `Port def: parsing uint "a": strconv.ParseUint: parsing "a": invalid syntax`,
		},
		"uint64-help-with-invalid-def": {
			config: &struct {
				Difference uint64 `def:"a"`
			}{},
			wantErr: `Difference def: parsing uint64 "a": strconv.ParseUint: parsing "a": invalid syntax`,
		},
		"int-help-with-invalid-def": {
			config: &struct {
				Temperature int `def:"a"`
			}{},
			wantErr: `Temperature def: parsing int "a": strconv.Atoi: parsing "a": invalid syntax`,
		},
		"int64-help-with-invalid-def": {
			config: &struct {
				Distance int64 `def:"a"`
			}{},
			wantErr: `Distance def: parsing int64 "a": strconv.ParseInt: parsing "a": invalid syntax`,
		},
		"float64-help-with-invalid-def": {
			config: &struct {
				BlackHole float64 `def:"a"`
			}{},
			wantErr: `BlackHole def: parsing float64 "a": strconv.ParseFloat: parsing "a": invalid syntax`,
		},
		"duration-help-without-def": {
			config: &struct {
				Timeout time.Duration
			}{},
			want: "Usage of test: -timeout duration timeout (env TEST_TIMEOUT)",
		},
		"duration-help-with-def": {
			config: &struct {
				Timeout time.Duration `def:"1s"`
			}{},
			want: "Usage of test: -timeout duration timeout (env TEST_TIMEOUT) (default 1s)",
		},
		"duration-help-with-invalid-def": {
			config: &struct {
				Timeout time.Duration `def:"a"`
			}{},
			wantErr: `Timeout def: parsing duration "a": time: invalid duration "a"`,
		},
		"url-help-without-def": {
			config: &struct {
				Endpoint act.URL
			}{},
			want: "Usage of test: -endpoint value endpoint (env TEST_ENDPOINT)",
		},
		"url-help-with-def": {
			config: &struct {
				Endpoint act.URL `def:"http://localhost"`
			}{},
			want: "Usage of test: -endpoint value endpoint (env TEST_ENDPOINT) (default http://localhost)",
		},
		"url-help-with-invalid-def": {
			config: &struct {
				Endpoint act.URL `def:"%"`
			}{},
			wantErr: `Endpoint def: parsing url: parse "%": invalid URL escape "%"`,
		},
		"string-slice-help-without-def": {
			config: &struct {
				Buckets act.StringSlice
			}{},
			want: "Usage of test: -buckets value buckets (env TEST_BUCKETS)",
		},
		"string-slice-help-with-def": {
			config: &struct {
				Buckets act.StringSlice `def:"foo,bar,baz"`
			}{},
			want: "Usage of test: -buckets value buckets (env TEST_BUCKETS) (default ['foo','bar','baz'])",
		},
		"int-slice-help-without-def": {
			config: &struct {
				DailyTemperatures act.IntSlice
			}{},
			want: "Usage of test: -daily-temperatures value daily temperatures (env TEST_DAILY_TEMPERATURES)",
		},
		"int-slice-help-with-def": {
			config: &struct {
				DailyTemperatures act.IntSlice `def:"10,-5,0"`
			}{},
			want: "Usage of test: -daily-temperatures value daily temperatures (env TEST_DAILY_TEMPERATURES) (default [10,-5,0])", //nolint:lll
		},
		"int-slice-help-with-invalid-def": {
			config: &struct {
				DailyTemperatures act.IntSlice `def:"10,-5,a"`
			}{},
			wantErr: `DailyTemperatures def: parsing int: strconv.Atoi: parsing "a": invalid syntax`,
		},
		"number-in-child-struct-help-with-invalid-def": {
			config: &struct {
				DB struct {
					Port uint `def:"a"`
				}
			}{},
			wantErr: `Port def: parsing uint "a": strconv.ParseUint: parsing "a": invalid syntax`,
		},
		"override-flag-and-env": {
			config: &struct {
				Log struct {
					LogLevel  string `flag:"foo"`
					Verbose   bool   `env:"FOO"`
					Something struct {
						Why act.StringSlice
						Not act.IntSlice
						URL act.URL
					}
				}
				Expiration time.Duration `help:"bar"`
				Number1    int
				Number2    int64 `def:"5"`
				Number3    float64
				Number4    uint
				Number5    uint64
			}{},
			want: `Usage of test:
-expiration duration bar (env TEST_EXPIRATION)
-foo string log log level (env TEST_LOG_LOG_LEVEL)
-log-something-not value log something not (env TEST_LOG_SOMETHING_NOT)
-log-something-url value log something url (env TEST_LOG_SOMETHING_URL)
-log-something-why value log something why (env TEST_LOG_SOMETHING_WHY)
-log-verbose log verbose (env FOO) -number-1 int number 1 (env TEST_NUMBER_1)
-number-2 int number 2 (env TEST_NUMBER_2) (default 5)
-number-3 float number 3 (env TEST_NUMBER_3)
-number-4 uint number 4 (env TEST_NUMBER_4)
-number-5 uint number 5 (env TEST_NUMBER_5)`,
		},
		"time-invalid-def": {
			config: &struct {
				Start act.Time `def:"a"`
			}{},
			wantErr: `Start def: parsing time: parsing time "a" as "2006-01-02T15:04:05Z07:00": cannot parse "a" as "2006"`,
		},
		"time-valid-def": {
			config: &struct {
				Start act.Time `def:"2002-10-02T10:00:00-05:00"`
			}{},
			want: "Usage of test: -start value start (env TEST_START) (default 2002-10-02T10:00:00-05:00)",
		},
	}

	for n, tt := range tests { //nolint:paralleltest
		n := n
		tt := tt

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			b := &bytes.Buffer{}

			a := act.New("test", act.WithErrorHandling(flag.ContinueOnError), act.WithOutput(b))

			if err := a.Parse(tt.config, []string{"-h"}); err != nil && err.Error() != tt.wantErr {
				t.Errorf("want error %q got error %q", tt.wantErr, err.Error())
			}

			got := strings.TrimSpace(ws.ReplaceAllString(b.String(), " "))
			want := ws.ReplaceAllString(tt.want, " ")
			if got != want {
				t.Errorf("\ngot %q\nwant %q\n", got, want)
			}
		})
	}
}

func TestParse_valid(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := map[string]struct {
		in interface{}
	}{
		"empty-config": {
			in: &struct{}{},
		},
		"without-tags": {
			in: &struct {
				Host string
				Port int
				DB   struct {
					Kind     string
					Postgres struct {
						Host string
					}
					Mongo struct {
						Host act.StringSlice
					}
				}
				Start act.Time
			}{},
		},
		"with-tags": {
			in: &struct {
				Env   string `help:"environment [development|production]" def:"development"`
				Port  uint   `def:"3000"`
				Mongo struct {
					Hosts             act.StringSlice `def:"mongo"`
					ConnectionTimeout time.Duration   `def:"10s"`
					ReplicaSet        string
					MaxPoolSize       uint64 `def:"100"`
					TLS               bool
					Username          string
					Password          string
					Database          string `def:"cool"`
				}
				JWT struct {
					Secret                 string
					TokenExpiration        time.Duration `def:"24h"`
					RefreshTokenExpiration time.Duration `def:"168h"`
				}
				AWS struct {
					Region string `def:"eu-central-1"`
				}
				Start act.Time `def:"2002-10-02T10:00:00-05:00"`
			}{},
		},
	}

	for n, tt := range tests { //nolint:paralleltest
		n := n
		tt := tt

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			a := act.New("test")

			err := a.Parse(tt.in, []string{})
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestParse_environment_errors(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config        interface{}
		lookupEnvFunc func(string) (string, bool)
		wantErr       string
	}{
		"number-invalid-env": {
			config: &struct {
				Port uint `def:"1"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "a", true
			},
			wantErr: `Port env: parsing uint "a": strconv.ParseUint: parsing "a": invalid syntax`,
		},
		"duration-invalid-env": {
			config: &struct {
				Timeout time.Duration `def:"1s"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "a", true
			},
			wantErr: `Timeout env: parsing duration "a": time: invalid duration "a"`,
		},
		"url-invalid-env": {
			config: &struct {
				Endpoint act.URL `def:"http://localhost"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "%", true
			},
			wantErr: `Endpoint env: parsing url: parse "%": invalid URL escape "%"`,
		},
		"int-slice-invalid-env": {
			config: &struct {
				DailyTemperatures act.IntSlice `def:"10,-5,0"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "a,-2,3", true
			},
			wantErr: `DailyTemperatures env: parsing int: strconv.Atoi: parsing "a": invalid syntax`,
		},
	}

	for n, tt := range tests { //nolint:paralleltest
		n := n
		tt := tt

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			a := act.New("test", act.WithErrorHandling(flag.ContinueOnError), act.WithLookupEnvFunc(tt.lookupEnvFunc))

			if err := a.Parse(tt.config, []string{}); err != nil && err.Error() != tt.wantErr {
				t.Errorf("want error %v got error %v", tt.wantErr, err.Error())
			}
		})
	}
}

func TestParse_environment(t *testing.T) { //nolint:cyclop,gocognit,funlen,maintidx
	t.Parallel()

	tests := map[string]struct {
		config          interface{}
		lookupEnvFunc   func(string) (string, bool)
		wantString      string
		wantBool        bool
		wantUint        uint
		wantUint64      uint64
		wantInt         int
		wantInt64       int64
		wantFloat64     float64
		wantDuration    time.Duration
		wantURL         act.URL
		wantStringSlice act.StringSlice
		wantIntSlice    act.IntSlice
	}{
		"string-env-not-set-def-not-set": {
			config: &struct {
				LogLevel string
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantString: "",
		},
		"string-env-not-set-def-set": {
			config: &struct {
				LogLevel string `def:"debug"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantString: "debug",
		},
		"string-env-set": {
			config: &struct {
				LogLevel string `def:"debug"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "info", true
			},
			wantString: "info",
		},
		"bool-env-not-set-def-not-set": {
			config: &struct {
				Verbose bool
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantBool: false,
		},
		"bool-env-not-set-def-set": {
			config: &struct {
				Verbose bool `def:"true"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantBool: true,
		},
		"bool-env-set": {
			config: &struct {
				Verbose bool `def:"true"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "false", true
			},
			wantBool: false,
		},
		"uint-env-not-set-def-not-set": {
			config: &struct {
				Port uint
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantUint: 0,
		},
		"uint-env-not-set-def-set": {
			config: &struct {
				Port uint `def:"1"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantUint: 1,
		},
		"uint-env-set": {
			config: &struct {
				Port uint `def:"1"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "2", true
			},
			wantUint: 2,
		},
		"uint64-env-not-set-def-not-set": {
			config: &struct {
				Distance uint64
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantUint64: 0,
		},
		"uint64-env-not-set-def-set": {
			config: &struct {
				Distance uint64 `def:"1"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantUint64: 1,
		},
		"uint64-env-set": {
			config: &struct {
				Distance uint64 `def:"1"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "2", true
			},
			wantUint64: 2,
		},
		"int-env-not-set-def-not-set": {
			config: &struct {
				Degrees int
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantInt: 0,
		},
		"int-env-not-set-def-set": {
			config: &struct {
				Degrees int `def:"1"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantInt: 1,
		},
		"int-env-set": {
			config: &struct {
				Degrees int `def:"1"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "2", true
			},
			wantInt: 2,
		},
		"int64-env-not-set-def-not-set": {
			config: &struct {
				Difference int64
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantInt64: 0,
		},
		"int64-env-not-set-def-set": {
			config: &struct {
				Difference int64 `def:"1"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantInt64: 1,
		},
		"int64-env-set": {
			config: &struct {
				Difference int64 `def:"1"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "2", true
			},
			wantInt64: 2,
		},
		"float64-env-not-set-def-not-set": {
			config: &struct {
				Balance float64
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantFloat64: 0,
		},
		"float64-env-not-set-def-set": {
			config: &struct {
				Balance float64 `def:"1.2"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantFloat64: 1.2,
		},
		"float64-env-set": {
			config: &struct {
				Balance float64 `def:"1.2"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "2.3", true
			},
			wantFloat64: 2.3,
		},
		"duration-env-not-set-def-not-set": {
			config: &struct {
				Timeout time.Duration
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantDuration: time.Duration(0),
		},
		"duration-env-not-set-def-set": {
			config: &struct {
				Timeout time.Duration `def:"1s"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantDuration: time.Second,
		},
		"duration-env-set": {
			config: &struct {
				Timeout time.Duration `def:"1s"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "2s", true
			},
			wantDuration: 2 * time.Second,
		},
		"url-env-not-set-def-not-set": {
			config: &struct {
				Endpoint act.URL
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantURL: act.URL{}, //nolint:exhaustruct
		},
		"url-env-not-set-def-set": {
			config: &struct {
				Endpoint act.URL `def:"http://localhost"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantURL: act.URL{URL: &url.URL{Scheme: "http", Host: "localhost"}}, //nolint:exhaustruct
		},
		"url-env-set": {
			config: &struct {
				Endpoint act.URL `def:"http://localhost"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "https://api.example.com/v1", true
			},
			wantURL: act.URL{URL: &url.URL{Scheme: "https", Host: "api.example.com", Path: "/v1"}}, //nolint:exhaustruct
		},
		"string-slice-env-not-set-def-not-set": {
			config: &struct {
				Buckets act.StringSlice
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantStringSlice: act.StringSlice{},
		},
		"string-slice-env-not-set-def-set": {
			config: &struct {
				Buckets act.StringSlice `def:"foo,bar,baz"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantStringSlice: act.StringSlice{"foo", "bar", "baz"},
		},
		"string-slice-env-set": {
			config: &struct {
				Buckets act.StringSlice `def:"foo,bar,baz"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "qux,bar", true
			},
			wantStringSlice: act.StringSlice{"qux", "bar"},
		},
		"int-slice-env-not-set-def-not-set": {
			config: &struct {
				DailyTemperatures act.IntSlice
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantIntSlice: act.IntSlice{},
		},
		"int-slice-env-not-set-def-set": {
			config: &struct {
				DailyTemperatures act.IntSlice `def:"10,-5,0"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "", false
			},
			wantIntSlice: act.IntSlice{10, -5, 0},
		},
		"int-slice-env-set": {
			config: &struct {
				DailyTemperatures act.IntSlice `def:"10,-5,0"`
			}{},
			lookupEnvFunc: func(env string) (string, bool) {
				return "-2,3,-1", true
			},
			wantIntSlice: act.IntSlice{-2, 3, -1},
		},
	}

	for n, tt := range tests { //nolint:paralleltest
		n := n
		tt := tt

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			a := act.New("test", act.WithErrorHandling(flag.ContinueOnError), act.WithLookupEnvFunc(tt.lookupEnvFunc))

			if err := a.Parse(tt.config, []string{}); err != nil {
				t.Error(err)
			}

			v := reflect.ValueOf(tt.config).Elem()

			p := v.FieldByIndex([]int{0}).Addr().Interface()

			switch p := p.(type) {
			case *string:
				if *p != tt.wantString {
					t.Errorf("want %q, got %q", tt.wantString, *p)
				}
			case *bool:
				if *p != tt.wantBool {
					t.Errorf("want %t, got %t", tt.wantBool, *p)
				}
			case *uint:
				if *p != tt.wantUint {
					t.Errorf("want %d, got %d", tt.wantUint, *p)
				}
			case *uint64:
				if *p != tt.wantUint64 {
					t.Errorf("want %d, got %d", tt.wantUint64, *p)
				}
			case *int:
				if *p != tt.wantInt {
					t.Errorf("want %d, got %d", tt.wantInt, *p)
				}
			case *int64:
				if *p != tt.wantInt64 {
					t.Errorf("want %d, got %d", tt.wantInt64, *p)
				}
			case *float64:
				if *p != tt.wantFloat64 {
					t.Errorf("want %f, got %f", tt.wantFloat64, *p)
				}
			case *time.Duration:
				if *p != tt.wantDuration {
					t.Errorf("want %v, got %v", tt.wantDuration, *p)
				}
			case *act.URL:
				if !reflect.DeepEqual(*p, tt.wantURL) {
					t.Errorf("want %#v, got %#v", tt.wantURL, *p)
				}
			case *act.StringSlice:
				if !reflect.DeepEqual(*p, tt.wantStringSlice) {
					t.Errorf("want %#v, got %#v", tt.wantStringSlice, *p)
				}
			case *act.IntSlice:
				if !reflect.DeepEqual(*p, tt.wantIntSlice) {
					t.Errorf("want %#v, got %#v", tt.wantIntSlice, *p)
				}
			default:
				t.Errorf("type %T not supported", p)
			}
		})
	}
}

func TestWithUsage(t *testing.T) {
	t.Parallel()

	b := &bytes.Buffer{}

	a := act.New("me", act.WithErrorHandling(flag.ContinueOnError), act.WithOutput(b), act.WithUsage("test"))
	_ = a.Parse(&struct{}{}, []string{"-h"})

	if want := "Usage of test me:\n"; want != b.String() {
		t.Errorf("want %q got %q", want, b.String())
	}
}
