// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"sync"
)

// DiscriminatorToTypeFunc is used to get a reflect.Type from its
// discriminator.
type DiscriminatorToTypeFunc func(discriminator string) (reflect.Type, bool)

// DiscriminatorEncodeMode is a mask that describes the different encode
// options.
type DiscriminatorEncodeMode uint8

const (
	// DiscriminatorEncodeTypeNameIfRequired is the default behavior when
	// the discriminator is set, and the type name is only encoded if required.
	DiscriminatorEncodeTypeNameIfRequired DiscriminatorEncodeMode = 0

	// DiscriminatorEncodeTypeNameRootValue causes the type name to be encoded
	// for the root value.
	DiscriminatorEncodeTypeNameRootValue DiscriminatorEncodeMode = 1 << iota 

	// DiscriminatorEncodeTypeNameAllObjects causes the type name to be encoded
	// for all struct and map values. Please note this specifically does not
	// apply to the root value.
	DiscriminatorEncodeTypeNameAllObjects
)

func (m DiscriminatorEncodeMode) root() bool {
	return m & DiscriminatorEncodeTypeNameRootValue > 0
}

func (m DiscriminatorEncodeMode) all() bool {
	return m & DiscriminatorEncodeTypeNameAllObjects > 0
}

func (d *decodeState) isDiscriminatorSet() bool {
	return d.discriminatorTypeFieldName != "" &&
		d.discriminatorValueFieldName != ""
}

// discriminatorOpType describes the current operation related to
// discriminators when reading a JSON object's fields.
type discriminatorOpType uint8

const (
	// discriminatorOpTypeNameField indicates the discriminator type name
	// field was discovered.
	discriminatorOpTypeNameField = iota + 1

	// discriminatorOpValueField indicates the discriminator value field
	// was discovered.
	discriminatorOpValueField
)

func (d *decodeState) discriminatorGetValue() (reflect.Value, error) {
	// Record the current offset so we know where the data starts.
	offset := d.readIndex()

	// Create a temporary decodeState used to inspect the current object
	// and determine its discriminator type and decode its value.
	dd := &decodeState{
		disallowUnknownFields:       d.disallowUnknownFields,
		useNumber:                   d.useNumber,
		discriminatorToTypeFn:       d.discriminatorToTypeFn,
		discriminatorTypeFieldName:  d.discriminatorTypeFieldName,
		discriminatorValueFieldName: d.discriminatorValueFieldName,
	}
	dd.init(append([]byte{}, d.data[offset:]...))
	defer freeScanner(&dd.scan)
	dd.scan.reset()

	var (
		typeName string // the discriminator type name
		valueOff = -1   // the offset of a possible discriminator value
	)

	dd.scanWhile(scanSkipSpace)
	if dd.opcode != scanBeginObject {
		panic(phasePanicMsg)
	}

	for {
		dd.scanWhile(scanSkipSpace)
		if dd.opcode == scanEndObject {
			// closing } - can only happen on first iteration.
			break
		}
		if dd.opcode != scanBeginLiteral {
			panic(phasePanicMsg)
		}

		// Read key.
		start := dd.readIndex()
		dd.rescanLiteral()
		item := dd.data[start:dd.readIndex()]
		key, ok := unquote(item)
		if !ok {
			panic(phasePanicMsg)
		}

		// Check to see if the key is related to the discriminator.
		var discriminatorOp discriminatorOpType
		switch key {
		case d.discriminatorTypeFieldName:
			discriminatorOp = discriminatorOpTypeNameField
		case d.discriminatorValueFieldName:
			discriminatorOp = discriminatorOpValueField
		}

		// Read : before value.
		if dd.opcode == scanSkipSpace {
			dd.scanWhile(scanSkipSpace)
		}

		if dd.opcode != scanObjectKey {
			panic(phasePanicMsg)
		}
		dd.scanWhile(scanSkipSpace)

		// Read value.
		valOff := dd.readIndex()
		val := dd.valueInterface()

		switch discriminatorOp {
		case discriminatorOpTypeNameField:
			if valStr, ok := val.(string); !ok {
				return reflect.Value{}, fmt.Errorf(
					"json: discriminator type at offset %d is not string",
					offset+valOff)
			} else {
				typeName = valStr
			}
		case discriminatorOpValueField:
			valueOff = valOff
		}

		// Next token must be , or }.
		if dd.opcode == scanSkipSpace {
			dd.scanWhile(scanSkipSpace)
		}
		if dd.opcode == scanEndObject {
			break
		}
		if dd.opcode != scanObjectValue {
			panic(phasePanicMsg)
		}
	}

	// If there is not a type discriminator then return early.
	if typeName == "" {
		return reflect.Value{}, fmt.Errorf("json: missing discriminator")
	}

	t, err := discriminatorParseTypeName(typeName, d.discriminatorToTypeFn)
	if err != nil {
		return reflect.Value{}, err
	}

	// Instantiate a new instance of the discriminated type.
	var v reflect.Value
	switch t.Kind() {
	case reflect.Slice:
		// MakeSlice returns a value that is not addressable.
		// Instead, use MakeSlice to get the type, then use
		// reflect.New to create an addressable value.
		v = reflect.New(reflect.MakeSlice(t, 0, 0).Type()).Elem()
	case reflect.Map:
		// MakeMap returns a value that is not addressable.
		// Instead, use MakeMap to get the type, then use
		// reflect.New to create an addressable value.
		v = reflect.New(reflect.MakeMap(t).Type()).Elem()
	case reflect.Complex64, reflect.Complex128:
		return reflect.Value{}, fmt.Errorf("json: unsupported discriminator type: %s", t.Kind())
	default:
		v = reflect.New(t)
	}

	// Reset the decode state to prepare for decoding the data.
	dd.scan.reset()

	switch t.Kind() {
	case reflect.Map, reflect.Struct:
		// Set the offset to zero since the entire object will be decoded
		// into v.
		dd.off = 0
	default:
		// Set the offset to what it was before the discriminator value was
		// read so only the discriminator value is decoded into v.
		dd.off = valueOff
	}

	// This will initialize the correct scan step and op code.
	dd.scanWhile(scanSkipSpace)

	// Decode the data into the value.
	if err := dd.value(v); err != nil {
		return reflect.Value{}, err
	}

	return v, nil
}

func (d *decodeState) discriminatorInterfaceDecode(t reflect.Type, v reflect.Value) error {

	defer func() {
		// Advance the decode state, throwing away the value.
		_ = d.objectInterface()
	}()

	dv, err := d.discriminatorGetValue()
	if err != nil {
		return err
	}

	switch dv.Kind() {
	case reflect.Map, reflect.Slice:
		if dv.Type().AssignableTo(t) {
			v.Set(dv)
			return nil
		}
		if pdv := dv.Addr(); pdv.Type().AssignableTo(t) {
			v.Set(pdv)
			return nil
		}
	case reflect.Ptr:
		if dve := dv.Elem(); dve.Type().AssignableTo(t) {
			v.Set(dve)
			return nil
		}
		if dv.Type().AssignableTo(t) {
			v.Set(dv)
			return nil
		}
	}

	return fmt.Errorf("json: unsupported discriminator kind: %s", dv.Kind())
}

func (o encOpts) isDiscriminatorSet() bool {
	return o.discriminatorTypeFieldName != "" &&
		o.discriminatorValueFieldName != ""
}

func discriminatorInterfaceEncode(e *encodeState, v reflect.Value, opts encOpts) {
	v = v.Elem()
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Invalid:
		e.error(&UnsupportedValueError{v, fmt.Sprintf("invalid kind: %s", v.Kind())})
	case reflect.Map:
		e.discriminatorEncodeTypeName = true
		newMapEncoder(v.Type())(e, v, opts)
	case reflect.Struct:
		e.discriminatorEncodeTypeName = true
		newStructEncoder(v.Type())(e, v, opts)
	case reflect.Ptr:
		discriminatorInterfaceEncode(e, v, opts)
	default:
		typeName := v.Type().Name()
		if typeName == "" {
			typeName = v.Type().String()
		}
		e.WriteString(`{"`)
		e.WriteString(opts.discriminatorTypeFieldName)
		e.WriteString(`":"`)
		e.WriteString(typeName)
		e.WriteString(`","`)
		e.WriteString(opts.discriminatorValueFieldName)
		e.WriteString(`":`)
		e.reflectValue(v, opts)
		e.WriteByte('}')
	}
}

func discriminatorMapEncode(e *encodeState, v reflect.Value, opts encOpts) {
	if !e.discriminatorEncodeTypeName && !opts.discriminatorEncodeMode.all() {
		return
	}
	typeName := v.Type().Name()
	if typeName == "" {
		typeName = v.Type().String()
	}
	e.WriteByte('"')
	e.WriteString(opts.discriminatorTypeFieldName)
	e.WriteString(`":"`)
	e.WriteString(typeName)
	e.WriteByte('"')
	if v.Len() > 0 {
		e.WriteByte(',')
	}
	e.discriminatorEncodeTypeName = false
}

func discriminatorStructEncode(e *encodeState, v reflect.Value, opts encOpts) byte {
	if !e.discriminatorEncodeTypeName && !opts.discriminatorEncodeMode.all() {
		return '{'
	}
	e.WriteString(`{"`)
	typeName := v.Type().Name()
	if typeName == "" {
		typeName = v.Type().String()
	}
	e.WriteString(opts.discriminatorTypeFieldName)
	e.WriteString(`":"`)
	e.WriteString(typeName)
	e.WriteByte('"')
	e.discriminatorEncodeTypeName = false
	return ','
}

var discriminatorTypeRegistry = map[string]reflect.Type{
	"uint":         reflect.TypeOf(uint(0)),
	"uint8":        reflect.TypeOf(uint8(0)),
	"uint16":       reflect.TypeOf(uint16(0)),
	"uint32":       reflect.TypeOf(uint32(0)),
	"uint64":       reflect.TypeOf(uint64(0)),
	"uintptr":      reflect.TypeOf(uintptr(0)),
	"int":          reflect.TypeOf(int(0)),
	"int8":         reflect.TypeOf(int8(0)),
	"int16":        reflect.TypeOf(int16(0)),
	"int32":        reflect.TypeOf(int32(0)),
	"int64":        reflect.TypeOf(int64(0)),
	"float32":      reflect.TypeOf(float32(0)),
	"float64":      reflect.TypeOf(float64(0)),
	"bool":         reflect.TypeOf(true),
	"string":       reflect.TypeOf(""),
	"any":          reflect.TypeOf((*interface{})(nil)).Elem(),
	"interface{}":  reflect.TypeOf((*interface{})(nil)).Elem(),
	"interface {}": reflect.TypeOf((*interface{})(nil)).Elem(),

	// Not supported, but here to prevent the decoder from panicing
	// if encountered.
	"complex64":  reflect.TypeOf(complex64(0)),
	"complex128": reflect.TypeOf(complex128(0)),
}

// discriminatorPointerTypeCache caches the pointer type for another type.
// For example, a key that was the int type would have a value that is the
// *int type.
var discriminatorPointerTypeCache sync.Map // map[reflect.Type]reflect.Type

// cachedPointerType returns the pointer type for another and avoids repeated
// work by using a cache.
func cachedPointerType(t reflect.Type) reflect.Type {
	if value, ok := discriminatorPointerTypeCache.Load(t); ok {
		return value.(reflect.Type)
	}
	pt := reflect.New(t).Type()
	value, _ := discriminatorPointerTypeCache.LoadOrStore(t, pt)
	return value.(reflect.Type)
}

var (
	mapPatt   = regexp.MustCompile(`^\*?map\[([^\]]+)\](.+)$`)
	arrayPatt = regexp.MustCompile(`^\*?\[(\d+)\](.+)$`)
	slicePatt = regexp.MustCompile(`^\*?\[\](.+)$`)
)

// discriminatorParseTypeName returns a reflect.Type for the given type name.
func discriminatorParseTypeName(
	typeName string,
	typeFn DiscriminatorToTypeFunc) (reflect.Type, error) {

	// Check to see if the type is an array, map, or slice.
	var (
		aln = -1   // array length
		etn string // map or slice element type name
		ktn string // map key type name
	)
	if m := arrayPatt.FindStringSubmatch(typeName); len(m) > 0 {
		i, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, err
		}
		aln = i
		etn = m[2]
	} else if m := slicePatt.FindStringSubmatch(typeName); len(m) > 0 {
		etn = m[1]
	} else if m := mapPatt.FindStringSubmatch(typeName); len(m) > 0 {
		ktn = m[1]
		etn = m[2]
	}

	// indirectTypeName checks to see if the type name begins with a
	// "*" characters. If it does, then the type name sans the "*"
	// character is returned along with a true value indicating the
	// type is a pointer. Otherwise the original type name is returned
	// along with a false value.
	indirectTypeName := func(tn string) (string, bool) {
		if len(tn) > 1 && tn[0] == '*' {
			return tn[1:], true
		}
		return tn, false
	}

	lookupType := func(tn string) (reflect.Type, bool) {
		// Get the actual type name and a flag indicating whether the
		// type is a pointer.
		n, p := indirectTypeName(tn)

		// First look up the type in the built-in type registry.
		t, ok := discriminatorTypeRegistry[n]
		if !ok {
			// If not found in the type registry then see if the type
			// is returne from the optional type function.
			if typeFn == nil {
				return nil, false
			}
			if t, ok = typeFn(n); !ok {
				return nil, false
			}
		}
		// If the type was a pointer then get the type's pointer type.
		if p {
			t = cachedPointerType(t)
		}
		return t, true
	}

	var t reflect.Type

	if ktn == "" && etn != "" {
		et, ok := lookupType(etn)
		if !ok {
			return nil, fmt.Errorf("json: invalid array/slice element type: %s", etn)
		}
		if aln > -1 {
			// Array
			t = reflect.ArrayOf(aln, et)
		} else {
			// Slice
			t = reflect.SliceOf(et)
		}
	} else if ktn != "" && etn != "" {
		// Map
		kt, ok := lookupType(ktn)
		if !ok {
			return nil, fmt.Errorf("json: invalid map key type: %s", ktn)
		}
		et, ok := lookupType(etn)
		if !ok {
			return nil, fmt.Errorf("json: invalid map element type: %s", etn)
		}
		t = reflect.MapOf(kt, et)
	} else {
		var ok bool
		if t, ok = lookupType(typeName); !ok {
			return nil, fmt.Errorf("json: invalid discriminator type: %s", typeName)
		}
	}

	return t, nil
}
