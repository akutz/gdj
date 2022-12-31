//go:build go1.19 && !go1.20

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import (
	"io"

	json "github.com/akutz/gdj/go1/1.19/1.19.4"
)

// DiscriminatorEncodeMode is a mask that describes the different encode
// options.
type DiscriminatorEncodeMode = json.DiscriminatorEncodeMode

const (
	// DiscriminatorEncodeTypeNameIfRequired is the default behavior when
	// the discriminator is set, and the type name is only encoded if required.
	DiscriminatorEncodeTypeNameIfRequired = json.DiscriminatorEncodeTypeNameIfRequired

	// DiscriminatorEncodeTypeNameRootValue causes the type name to be encoded
	// for the root value.
	DiscriminatorEncodeTypeNameRootValue = json.DiscriminatorEncodeTypeNameRootValue

	// DiscriminatorEncodeTypeNameAllObjects causes the type name to be encoded
	// for all struct and map values. Please note this specifically does not
	// apply to the root value.
	DiscriminatorEncodeTypeNameAllObjects = json.DiscriminatorEncodeTypeNameAllObjects

	// DiscriminatorEncodeTypeNameWithPath causes the type name to be encoded
	// prefixed with the type's full package path.
	DiscriminatorEncodeTypeNameWithPath = json.DiscriminatorEncodeTypeNameWithPath
)

// DiscriminatorToTypeFunc is used to get a reflect.Type from its
// discriminator.
type DiscriminatorToTypeFunc = json.DiscriminatorToTypeFunc

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *json.Encoder {
	return json.NewEncoder(w)
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may
// read data from r beyond the JSON values requested.
func NewDecoder(r io.Reader) *json.Decoder {
	return json.NewDecoder(r)
}
