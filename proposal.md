# JSON Discriminator Proposal (:construction: work-in-progress :construction:)

Andrew Kutz<br />
January 1st, 2023


## Status

This is the design for adding support for a JSON discriminator to Go's [`encoding/json`](https://pkg.go.dev/encoding/json) package. This design has not yet been proposed to the Go community.


## Abstract

This proposal suggests enhancing the `encoding/json` package with support for an optional, [type discriminator](https://www.rfc-editor.org/rfc/rfc8927#name-discriminator) when encoding/decoding values to/from JSON. Values stored in interfaces are always encoded with their type name. A `map` or `struct` value is encoded with the type name directly as a member of the JSON object where all other values are wrapped in a JSON object that specifies the type and value. Pointers are indirected prior to being encoded. A discriminator does not enable the encoding or decoding of `chan`, `func`, `complex64`, or `complex128` values. Values may be decoded into empty interfaces or interfaces with methods as long as the decoded value or its address is assignable to the interface. The design is fully backward compatible with Go 1.


## How to read this proposal

This document is long; here is some guidance on how to read it:

* The first section is a [high level overview](#very-high-level-overview) which briefly describes the concepts.
* The second section provides some [background](#background) on the issue and why support for a JSON discriminator is needed.
* Next, the [full design](#design) is explained from scratch, introducing details as necessary, with simple examples.
* After the design is completely described, the [implementation](#implementation) is reviewed.
* There are also several, complete [examples](#examples) of how this design would be used in practice.
* Following the examples, some minor details are discussed in an [appendix](#appendix).


## Very high level overview

This section explains the changes suggested by the design very briefly. This section is intended for people who are already familiar with how `encoding/json` works. Concepts touched upon here will be explained in detail in the following sections.

* A discriminator may be used to encode/decode type information for Go types that correspond to the following `relfect.Kind` values: `Bool`, `Int`, `Int8`, `Int16`, `Int32`, `Int64`, `Uint`, `Uint8`, `Uint16`, `Uint32`, `Uint64`, `Uintptr`, `Float32`, `Float64`, `Array`, `Interface`, `Map`, `Pointer`, `Slice`, `String`, `Struct`.

* A discriminator may **not** be used to encode/decode type information for Go types that correspond to the following `relfect.Kind` values: `Complex64`, `Complex128`, `Chan`, `Func`, `UnsafePointer`.

* If `T` is a `Struct` or `Map` then the type name is encoded as an additional field inside of the encoded JSON object.

* If `T` is any other of the supported types then the value is wrapped in a JSON object with two fields, one that specifies the Go type for the value and another field that is the value.

* If a value stored in an interface is `*T`, the type name for `T` is encoded.

* If a decoded value of type `T` cannot be assigned to an interface but `*T` _can_, then the address of the value is stored in the interface.

* There is a new type, `type DiscriminatorEncodeMode uint8`, which is used by the mask that controls different options during encode operations when the discriminator is set:

    | Name                                    | Value | Description                                                                                                                |
    | --------------------------------------- | ----- | -------------------------------------------------------------------------------------------------------------------------- |
    | `DiscriminatorEncodeTypeNameIfRequired` | `0`   | The type name is only encoded for values stored in interfaces. This is the default behavior when the discriminator is set. |
    | `DiscriminatorEncodeTypeNameRootValue`  | `2`   | The type name is encoded for the root value.                                                                               |
    | `DiscriminatorEncodeTypeNameAllObjects` | `4`   | The type name is encoded for all `Struct` and `Map` values. Please note this does not include the root value.              |

* The `Encoder` type has a new function used to enable/disable the discriminator, `SetDiscriminator(typeFieldName, valueFieldName string, mode DiscriminatorEncodeMode)`:

  * The `typeFieldName` parameter specifies the name of the key used to encode the value's type in a JSON object.
  * The `valueFieldName` parameter specifies the name of the key used to encode the value in a JSON object.
  * The `mode` parameter specifies the aforementioned `DiscriminatorEncodeMode`.

* There is a new type, `type DiscriminatorToTypeFunc func(typeName string) (reflect.Type, bool)`, which refers to the function used by the `Decoder` to look up a `reflect.Type` based on encoded type names.

* The `Decoder` type has a new function used to enable/disable the discriminator, `SetDiscriminator(typeFieldName, valueFieldName string, typeFn DiscriminatorToTypeFunc)`:

  * The `typeFieldName` and `valueFieldName` parameters specify the same information as they do in `Encoder.SetDiscriminator`.
  * The `typeFn` parameter specifies the aforementioned `DiscriminatorToTypeFunc`.

The following sections work through each of these changes in great detail. Readers may prefer to skip ahead to the [examples](#examples) to see what code written to this design will look like in practice.


## Background

JSON was created in 2000-2001 to fill a gap that required a light-weight, stateless data exchange model for server-to-browser communication without relying on Flash or Java applets and that could be easily used by JavaScript. Unlike XML, which saw the creation of schemas alongside XML 1.0, the first draft of the JSON Schema did not occur [until 2013](https://datatracker.ietf.org/doc/html/draft-zyp-json-schema-00), over a decade after JSON became widely used. 

Golang will happily encode a `struct` with one or more fields of type `interface`. For example, take the following ([Golang playground](https://go.dev/play/p/kXY1fKPwB1C)):

```go
package main

import (
	"encoding/json"
	"os"
)

type CMYK struct {
	Cyan    int
	Magenta int
	Yellow  int
	Key     int
}

type ColorGroup struct {
	ID     int
	Name   string
	Colors []interface{}
}

func main() {
    cg := ColorGroup{
		ID:   1,
		Name: "Reds",
		Colors: []interface{}{
			[3]int{220, 20, 60}, // Crimson
			"Red",
			CMYK{Cyan: 0, Magenta: 92, Yellow: 58, Key: 12}, // Ruby
			0x800000, // Maroon
		},
	}
	for i := range cg.Colors {
		fmt.Printf("%T\n", cg.Colors[i])
	}
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(cg)
}
```

The above program will output:

* the types of the elements in the `Colors` field:

    ```
    [3]int
    string
    main.CMYK
    int
    ```

* the object encoded as JSON:

    ```json
    {"ID":1,"Name":"Reds","Colors":[[220,20,60],"Red",{"Cyan":0,"Magenta":92,"Yellow":58,"Key":12},8388608]}
    ```

However, decoding the above JSON back into a `ColorGroup` will not work ([Golang playground](https://go.dev/play/p/1JM4riw6XR2)) as expected:

```go
package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type CMYK struct {
	Cyan    int
	Magenta int
	Yellow  int
	Key     int
}

type ColorGroup struct {
	ID     int
	Name   string
	Colors []interface{}
}

func main() {
	var cg ColorGroup
	dec := json.NewDecoder(strings.NewReader(`{"ID":1,"Name":"Reds","Colors":[[220,20,60],"Red",{"Cyan":0,"Magenta":92,"Yellow":58,"Key":12},8388608]}`))
	dec.Decode(&cg)
	for i := range cg.Colors {
		fmt.Printf("%T\n", cg.Colors[i])
	}
}
```

The program emits the following:

```
[]interface {}
string
map[string]interface {}
float64
```

However, all but one of these types do not match those from the original object:

| Original Type | New Type                  |
| ------------- | ------------------------- |
| `[3]int`      | `[]interface{}`           |
| `string`      | `string`                  |
| `main.CMYK`   | `map[string]interface {}` |
| `int`         | `float64`                 |

This is because Golang's `encoding/json` package, by design, does not include type information when encoding a value _to_ JSON or look for type information when decoding a value _from_ JSON. And that is _fine_! In fact, that is the way JSON was conceived -- a light-weight data transport that did not carry the overbearing schemas used by other data models such as XML. However, the [OpenAPI 3 specification](https://swagger.io/docs/specification/data-models/inheritance-and-polymorphism/) and the proposed, [RFC 8927](https://www.rfc-editor.org/rfc/rfc8927#name-discriminator) both support the concept of a discriminator. Kubernetes, a massively popular container orchestration platform, has an extensible API based on the OpenAPI specification that is often extended by generating Custom Resource Definitions (CRD) from Go types. However, CRDs generated from Go types will not have support for OpenAPI3 discriminators since Go lacks the same support.

There are many different ways to tackle this issue, but they all, ultimately, boil down to two variations of the same design:

1. Types implement the `encoding/json` package's `Marshaler` and `Unmarshaler` interfaces to control their own encoding/decoding
1. Reflection is used to avoid defining custom encoding/decoding logic for every involved type

The first option is


## Design

Lorem ipsum.

## Implementation

Lorem ipsum.

## Examples

Lorem ipsum.

## Appendix

Lorem ipsum.
