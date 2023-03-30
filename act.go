// Package act is a library for parsing command line flags, environment variables and default values defined by
// struct tag "def" respecting the order of precedence according to the 12-factor principles.
// It generates predefined flag names and environment variables names by the fields path of the supplied config,
// but allows this to be overridden by struct tags "flag" and "env".
package act

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/iancoleman/strcase"
)

// Errors.
var (
	ErrInvalidConfigType = errors.New("invalid config type")
	ErrUnsupportedType   = errors.New("type not supported")
)

// Act is an abstraction of a CLI command.
type Act struct {
	flagSet       *flag.FlagSet
	output        io.Writer
	lookupEnvFunc func(string) (string, bool)
	name          string
	errorHandling flag.ErrorHandling
	help          bool
}

// New creates new act command.
func New(name string, opts ...Option) *Act {
	a := &Act{ //nolint:exhaustruct
		flagSet:       flag.NewFlagSet(name, flag.ContinueOnError),
		output:        os.Stderr,
		lookupEnvFunc: os.LookupEnv,
		name:          name,
		errorHandling: flag.ExitOnError,
	}

	for _, opt := range opts {
		opt(a)
	}

	a.flagSet.SetOutput(a.output)

	return a
}

// Parse parses command line flags, environment variables and default values.
// It populates supplied pointer to configuration struct with values according to the order of precedence.
func (a *Act) Parse(config interface{}, flags []string) error {
	if err := a.parse(config, flags, ""); err != nil {
		return a.exit(err)
	}

	return a.exit(a.flagSet.Parse(flags))
}

func (a *Act) parse(config interface{}, flags []string, prefix string) error { //nolint:cyclop
	a.parseHelp(flags)

	v := reflect.ValueOf(config)
	t := reflect.TypeOf(config)

	if v.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return ErrInvalidConfigType
	}

	v = v.Elem()

	for i := 0; i < v.NumField(); i++ {
		field := t.Elem().Field(i)

		flagName := a.flagName(field, prefix)

		envVarName := a.envVarName(field, prefix)

		usage := a.usage(field, envVarName, prefix)

		p := v.FieldByName(field.Name).Addr().Interface()

		// Recurse if got struct which is not of URL type
		_, oku := p.(*URL)
		_, okt := p.(*Time)

		if field.Type.Kind() == reflect.Struct && !oku && !okt {
			if err := a.parse(p, flags, a.newPrefix(field, prefix)); err != nil {
				return err
			}

			continue
		}

		envVarValue, ok := a.lookupEnvFunc(envVarName)
		if ok && !a.help {
			if err := a.parseValue(field.Type.Kind(), p, flagName, envVarValue, usage); err != nil {
				return fmt.Errorf("%s env: %w", field.Name, err)
			}

			continue
		}

		if err := a.parseValue(field.Type.Kind(), p, flagName, field.Tag.Get("def"), usage); err != nil {
			return fmt.Errorf("%s def: %w", field.Name, err)
		}
	}

	return nil
}

func (*Act) flagName(sf reflect.StructField, prefix string) string {
	if f := sf.Tag.Get("flag"); f != "" {
		return f
	}

	n := sf.Name

	if prefix != "" {
		n = fmt.Sprintf("%s-%s", prefix, n)
	}

	return strcase.ToKebab(n)
}

func (a *Act) envVarName(sf reflect.StructField, prefix string) string {
	if e := sf.Tag.Get("env"); e != "" {
		return e
	}

	n := fmt.Sprintf("%s_%s", a.name, sf.Name)
	if prefix != "" {
		n = fmt.Sprintf("%s_%s_%s", a.name, prefix, sf.Name)
	}

	return strcase.ToScreamingSnake(n)
}

func (*Act) usage(sf reflect.StructField, env string, prefix string) string {
	if u := sf.Tag.Get("help"); u != "" {
		return fmt.Sprintf("%s (env %s)", u, env)
	}

	n := sf.Name
	if prefix != "" {
		n = fmt.Sprintf("%s %s", prefix, sf.Name)
	}

	return fmt.Sprintf("%s (env %s)", strcase.ToDelimited(n, ' '), env)
}

func (a *Act) parseHelp(flags []string) {
	if len(flags) == 0 {
		return
	}

	for _, f := range flags {
		if f == "--help" || f == "-help" || f == "--h" || f == "-h" {
			a.help = true

			return
		}
	}
}

func (*Act) newPrefix(sf reflect.StructField, prefix string) string {
	if prefix != "" {
		return fmt.Sprintf("%s-%s", prefix, sf.Name)
	}

	return sf.Name
}

func (a *Act) parseValue(kind reflect.Kind, varPointer interface{}, flag, value, usage string) error { //nolint:cyclop
	switch kind { //nolint:exhaustive
	case reflect.Bool:
		return a.parseBool(varPointer.(*bool), flag, value, usage) //nolint:forcetypeassert
	case reflect.String:
		a.flagSet.StringVar(varPointer.(*string), flag, value, usage) //nolint:forcetypeassert

		return nil
	case reflect.Uint:
		return a.parseUint(varPointer.(*uint), flag, value, usage) //nolint:forcetypeassert
	case reflect.Uint64:
		return a.parseUint64(varPointer.(*uint64), flag, value, usage) //nolint:forcetypeassert
	case reflect.Int:
		return a.parseInt(varPointer.(*int), flag, value, usage) //nolint:forcetypeassert
	case reflect.Int64:
		switch varPointer := varPointer.(type) {
		case *time.Duration:
			return a.parseDuration(varPointer, flag, value, usage)
		case *int64:
			return a.parseInt64(varPointer, flag, value, usage)
		}
	case reflect.Float64:
		return a.parseFloat64(varPointer.(*float64), flag, value, usage) //nolint:forcetypeassert
	case reflect.Struct:
		switch varPointer := varPointer.(type) {
		case *URL:
			return a.parseURL(varPointer, flag, value, usage)
		case *Time:
			return a.parseTime(varPointer, flag, value, usage)
		}
	case reflect.Slice:
		switch varPointer := varPointer.(type) {
		case *StringSlice:
			return a.parseStringSlice(varPointer, flag, value, usage)
		case *IntSlice:
			return a.parseIntSlice(varPointer, flag, value, usage)
		}
	}

	return fmt.Errorf("parsing value: %w: %v", ErrUnsupportedType, kind)
}

func (a *Act) parseBool(p *bool, flag, value, usage string) error {
	if value == "" {
		a.flagSet.BoolVar(p, flag, false, usage)

		return nil
	}

	val, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("parsing bool %q: %w", value, err)
	}

	a.flagSet.BoolVar(p, flag, val, usage)

	return nil
}

func (a *Act) parseUint(p *uint, flag, value, usage string) error {
	if value == "" {
		a.flagSet.UintVar(p, flag, 0, usage)

		return nil
	}

	val, err := strconv.ParseUint(value, 10, 32) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("parsing uint %q: %w", value, err)
	}

	a.flagSet.UintVar(p, flag, uint(val), usage)

	return nil
}

func (a *Act) parseUint64(p *uint64, flag, value, usage string) error {
	if value == "" {
		a.flagSet.Uint64Var(p, flag, 0, usage)

		return nil
	}

	val, err := strconv.ParseUint(value, 10, 64) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("parsing uint64 %q: %w", value, err)
	}

	a.flagSet.Uint64Var(p, flag, val, usage)

	return nil
}

func (a *Act) parseInt(p *int, flag, value, usage string) error {
	if value == "" {
		a.flagSet.IntVar(p, flag, 0, usage)

		return nil
	}

	val, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("parsing int %q: %w", value, err)
	}

	a.flagSet.IntVar(p, flag, val, usage)

	return nil
}

func (a *Act) parseInt64(p *int64, flag, value, usage string) error {
	if value == "" {
		a.flagSet.Int64Var(p, flag, 0, usage)

		return nil
	}

	val, err := strconv.ParseInt(value, 10, 64) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("parsing int64 %q: %w", value, err)
	}

	a.flagSet.Int64Var(p, flag, val, usage)

	return nil
}

func (a *Act) parseDuration(p *time.Duration, flag, value, usage string) error {
	if value == "" {
		a.flagSet.DurationVar(p, flag, 0, usage)

		return nil
	}

	val, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf("parsing duration %q: %w", value, err)
	}

	a.flagSet.DurationVar(p, flag, val, usage)

	return nil
}

func (a *Act) parseFloat64(p *float64, flag, value, usage string) error {
	if value == "" {
		a.flagSet.Float64Var(p, flag, 0, usage)

		return nil
	}

	val, err := strconv.ParseFloat(value, 64) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("parsing float64 %q: %w", value, err)
	}

	a.flagSet.Float64Var(p, flag, val, usage)

	return nil
}

func (a *Act) parseStringSlice(p *StringSlice, flag, value, usage string) error {
	if value == "" {
		*p = StringSlice{}
		a.flagSet.Var(p, flag, usage)

		return nil
	}

	ss := &StringSlice{}

	_ = ss.Set(value)

	*p = *ss
	a.flagSet.Var(p, flag, usage)

	return nil
}

func (a *Act) parseIntSlice(p *IntSlice, flag, value, usage string) error {
	if value == "" {
		*p = IntSlice{}
		a.flagSet.Var(p, flag, usage)

		return nil
	}

	is := &IntSlice{}

	if err := is.Set(value); err != nil {
		return err
	}

	*p = *is
	a.flagSet.Var(p, flag, usage)

	return nil
}

func (a *Act) parseURL(p *URL, flag, value, usage string) error {
	if value == "" {
		*p = URL{} //nolint:exhaustruct
		a.flagSet.Var(p, flag, usage)

		return nil
	}

	u := &URL{} //nolint:exhaustruct

	if err := u.Set(value); err != nil {
		return err
	}

	*p = *u
	a.flagSet.Var(p, flag, usage)

	return nil
}

func (a *Act) parseTime(p *Time, flag, value, usage string) error {
	if value == "" {
		*p = Time{} //nolint:exhaustruct
		a.flagSet.Var(p, flag, usage)

		return nil
	}

	t := &Time{} //nolint:exhaustruct

	if err := t.Set(value); err != nil {
		return err
	}

	*p = *t
	a.flagSet.Var(p, flag, usage)

	return nil
}

func (a *Act) exit(err error) error {
	if err == nil {
		return nil
	}

	switch a.errorHandling {
	case flag.ContinueOnError:
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}

		return err
	case flag.ExitOnError:
		if a.help || errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}

		fmt.Fprintf(a.output, "act: %v\n", err)
		os.Exit(2) //nolint:gomnd
	case flag.PanicOnError:
		panic(err)
	}

	return nil
}

// Option defines optional parameters to the constructor.
type Option func(a *Act)

// WithErrorHandling is an option to change error handling similar to flag package.
func WithErrorHandling(errorHandling flag.ErrorHandling) Option {
	return func(a *Act) {
		a.errorHandling = errorHandling
	}
}

// WithOutput is an option to change the output writer similar as flag.SetOutput does.
func WithOutput(w io.Writer) Option {
	return func(a *Act) {
		a.output = w
	}
}

// WithLookupEnvFunc may be used to override default os.LookupEnv function to read environment variables values.
func WithLookupEnvFunc(fn func(string) (string, bool)) Option {
	return func(a *Act) {
		a.lookupEnvFunc = fn
	}
}

// WithUsage allows to prefix your command name with a parent command name.
func WithUsage(parentCmdName string) Option {
	return func(a *Act) {
		a.flagSet.Usage = func() {
			fmt.Fprintf(a.output, "Usage of %s %s:\n", parentCmdName, a.name)
			a.flagSet.PrintDefaults()
		}
	}
}
