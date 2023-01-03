# Go's Discriminating JSON (gdj)

This repository contains a fork of Golang's [`encoding/json`](https://pkg.go.dev/encoding/json) package, enhanced with support for encoding type information into objects via a [discriminator](https://www.rfc-editor.org/rfc/rfc8927#name-discriminator).

## Overview

This project enhances the Go package `encoding/json` with support for an optional, [type discriminator](https://www.rfc-editor.org/rfc/rfc8927#name-discriminator) when encoding/decoding values to/from JSON. For more information please see the associated, [JSON Discriminator Proposal](./proposal.md) that will be submitted to Go.

## Goals

This project is intended to help drive the following goals:

* Introduce support for a JSON discriminator in Go's `encoding/json` package
* Make it easy for people to use JSON discriminators _today_ with minimal changes to their existing types/code
* Support all versions of Go newer than or equal to 1.17.13, 1.18.9, 1.19.4, 1.20rc1

## Getting started

Using this project is quite simple, just:

1. import `github.com/akutz/gdj`
1. create a new `json.Encoder` or `json.Encoder`
1. set the discriminator on the new encoder/decoder

The following example illustrates a encoding and decoding JSON with a discriminator:

```go
import "github.com/akutz/gdj" // imports as the "json" package

// ...

type Person struct {
	Name       string        `json:"name"`
	Attributes []interface{} `json:"attributes,omitempty"`
}

enc := json.NewEncoder(os.Stdout)
enc.SetDiscriminator("type", "value", 0)

enc.Encode(Person{"Andrew", []interface{}{"Austin", uint8(42)}})
enc.Encode(Person{"Mandy", []interface{}{Spouse{Person{"Andrew", nil}}}})
```

The above program will emit the following output:

```
{"name":"Andrew","attributes":[{"type":"string","value":"Austin"},{"type":"uint8","value":42}]}
{"name":"Mandy","attributes":[{"type":"Spouse","name":"Andrew"}]}
```

The type information is encoded alongside the values for the elements in the field `Person.Attributes`. It is also possible to decode the information back to a `Person` while maintaining the same type information:

```go
var jsonBlob = `{
	"name":"Andrew",
	"attributes":[
		{"type":"string", "value": "Austin"},
		{"type":"uint8", "value":42}
	]
}
{
	"name":"Mandy",
	"attributes":[
		{"type":"Spouse", "name": "Andrew"}
	]
}`

dec := json.NewDecoder(strings.NewReader(jsonBlob))
dec.SetDiscriminator("type", "value", func(s string) (reflect.Type, bool) {
	switch s {
	case "Person":
		return reflect.TypeOf(Person{}), true
	case "Spouse":
		return reflect.TypeOf(Spouse{}), true
	}
	return nil, false
})

var p Person
dec.Decode(&p)
fmt.Printf("%[1]T(%[1]d)\n", p.Attributes[1])

dec.Decode(&p)
fmt.Printf("%[1]T(%[1]s)\n", p.Attributes[0].(Spouse).Name)
```

The above program emits the following:

```
uint8(42)
string(Andrew)
```

The output indicates the original type and value information was respected when the data was decoded. For more examples, please see:

* [`discriminator_test.go`](./discriminator_test.go)
* [`example_discriminator_test.go`](./example_discriminator_test.go)
* [`govmomi_test.go`](./canaries/govmomi_test.go)


## Type support

The discriminator supports encoding and decoding the following, built-in types:

* `uint`
* `uint8`
* `uint16`
* `uint32`
* `uint64`
* `uintptr`
* `int`
* `int8`
* `int16`
* `int32`
* `int64`
* `float32`
* `float64`
* `bool`
* `string`

Encoding custom types is supported as well, with decoding custom types dependent on the type lookup function provided to the decoder's `SetDiscriminator` function.


## Testing

The discriminator functionality is thoroughly tested with:

* the tests from the `encoding/json` package
* ~300 encoding/decoding tests in [`discriminator_test.go`](./discriminator_test.go)
* canary testing of complex type models in the [`canaries`] directory, ex. the GoVmomi `VirtualMachineConfigInfo` structure ([`govmomi_test.go`](./canaries/govmomi_test.go))

All of the above tests are executed:

* on every pull request
* push to the `main` branch
* via the GitHub action, [`test` workflow](./.github/workflows/test.yml)
* for all supported versions of Go, ex. 1.17.13, 1.18.9, 1.19.4, 1.20rc1

It is also possible to run the tests locally with `make test`. This depends on either:

* an environment variable, `GO_<VERSION>_BIN`, for each version of Go tested that points to the Go installation's `go` binary, ex. `GO_1.17.3_BIN="${HOME}/.go/1.17.3/bin/go"`
* or Docker, which is used to run the tests with the official Golang container images

The command `make test` will attempt to use a local Go installation for a given version of Golang, and if one cannot be found, default to using Docker. It is also possible to force the use of Docker with `DOCKER=1 make test`.

## License

This is a fork of Go's `encoding/json` package and so uses the same license as Golang.
