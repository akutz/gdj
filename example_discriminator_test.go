// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	json "github.com/akutz/gdj"
)

func ExampleNewEncoder() {
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
	group := ColorGroup{
		ID:   1,
		Name: "Reds",
		Colors: []interface{}{
			[3]int{220, 20, 60}, // Crimson
			"Red",
			CMYK{Cyan: 0, Magenta: 92, Yellow: 58, Key: 12}, // Ruby
			0x800000, // Maroon
		},
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetDiscriminator("_t", "_v", 0)
	if err := enc.Encode(group); err != nil {
		fmt.Println("error:", err)
	}
	// Output:
	// {"ID":1,"Name":"Reds","Colors":[{"_t":"[3]int","_v":[220,20,60]},{"_t":"string","_v":"Red"},{"_t":"CMYK","Cyan":0,"Magenta":92,"Yellow":58,"Key":12},{"_t":"int","_v":8388608}]}
}

func ExampleNewDecoder() {
	var jsonBlob = `{
	"ID":1,
	"Name":"Reds",
	"Colors":[
		{"_t":"[3]int","_v":[220,20,60]},
		{"_t":"string","_v":"Red"},
		{"_t":"CMYK","Cyan":0,"Magenta":92,"Yellow":58,"Key":12},
		{"_t":"int","_v":8388608}
	]}`
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
	dec := json.NewDecoder(strings.NewReader(jsonBlob))
	dec.SetDiscriminator("_t", "_v", func(s string) (reflect.Type, bool) {
		switch s {
		case "CMYK":
			return reflect.TypeOf(CMYK{}), true
		case "ColorGroup":
			return reflect.TypeOf(ColorGroup{}), true
		}
		return nil, false
	})
	var group ColorGroup
	if err := dec.Decode(&group); err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v", group)
	// Output:
	// {ID:1 Name:Reds Colors:[[220 20 60] Red {Cyan:0 Magenta:92 Yellow:58 Key:12} 8388608]}
}

type Person struct {
	Name       string        `json:"name"`
	Attributes []interface{} `json:"attributes,omitempty"`
}

func (p Person) GetName() string {
	return p.Name
}
func (p *Person) SetName(s string) {
	p.Name = s
}

type Spouse struct {
	Person
}

type CanGetName interface {
	GetName() string
}
type CanSetName interface {
	SetName(string)
}

func ExampleNewEncoder_empty_interface() {
	enc := json.NewEncoder(os.Stdout)
	enc.SetDiscriminator("type", "value", 0)

	enc.Encode(Person{"Andrew", []interface{}{"Austin", uint8(42)}})
	enc.Encode(Person{"Mandy", []interface{}{Spouse{Person{"Andrew", nil}}}})
	// Output:
	// {"name":"Andrew","attributes":[{"type":"string","value":"Austin"},{"type":"uint8","value":42}]}
	// {"name":"Mandy","attributes":[{"type":"Spouse","name":"Andrew"}]}
}

func ExampleNewDecoder_empty_interface() {
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
	// Output:
	// uint8(42)
	// string(Andrew)
}
