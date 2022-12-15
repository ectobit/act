# act

Microservices oriented [12-factor](https://12factor.net) Go library for parsing environment variables and
command line flags to arbitrary config struct using struct tags to define default values and to override flag names and
environment variables' names.

[![Build Status](https://github.com/ectobit/act/workflows/build/badge.svg)](https://github.com/ectobit/act/actions)
![Go Coverage](https://img.shields.io/badge/coverage-97.5%25-brightgreen?style=flat&logo=go)
[![Go Reference](https://pkg.go.dev/badge/go.ectobit.com/act.svg)](https://pkg.go.dev/go.ectobit.com/act)
[![Go Report](https://goreportcard.com/badge/go.ectobit.com/act)](https://goreportcard.com/report/go.ectobit.com/act)

This package in intended to be used to parse command line arguments and environment variables into an arbitrary config struct.
This struct may contain multiple nested structs, they all will be processed recursively. Names of the flags and environment
variables are automatically generated. Flags will be kebab case of the field name eventually preceded by parent fields
in case of nested structs. Names of environment variables will be similar, but additionally prefixed with command name
and then snake and upper cased. Description of each flag will also be automatically generated in a human friendly way
as much as possible. Additionally, you may override these auto-generated names using the struct tags and you also may
define default value.

- **flag** - override generated flag name
- **env** - override generated environment variable name
- **help** - override generated flag description
- **def** - override default (zero) value

## Important: all struct fields should be exported.

## Custom flag types

Besides the types supported by flag package, this package provides additional types:

- **act.StringSlice** - doesn't support multiple flags but instead supports comma separated strings, i.e. "foo,bar"
- **act.IntSlice** - doesn't support multiple flags but instead supports comma separated integers, i.e. "5,-8,0"
- **act.URL**
- **act.Time** - RFC3339 time

## Order of precedence:

- command line options
- environment variables
- default values

## [Examples](example_test.go)

Run `make test-verbose` to see examples output.

## Subcommands

These are handled just like by standard library's flag package.

```go
package main

import (
	"log"
	"os"

	"go.ectobit.com/act"
)

func main() {
	subCmd := os.Args[1]
	switch subCmd {
	case "create":
		config := &struct{}{}
		createCmd := act.New("create")

		if err := createCmd.Parse(config, os.Args[2:]); err != nil {
			log.Println(err)
		}

		// Implementation

	case "delete":
		config := &struct{}{}
		deleteCmd := act.New("create")

		if err := deleteCmd.Parse(config, os.Args[2:]); err != nil {
			log.Println(err)
		}

		// Implementation
	}
}
```

## TODO

- support req struct tag to mark required values

## License

Licensed under either of

- Apache License, Version 2.0
  ([LICENSE-APACHE](LICENSE-APACHE) or http://www.apache.org/licenses/LICENSE-2.0)
- MIT license
  ([LICENSE-MIT](LICENSE-MIT) or http://opensource.org/licenses/MIT)

at your option.

## Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted
for inclusion in the work by you, as defined in the Apache-2.0 license, shall be
dual licensed as above, without any additional terms or conditions.
