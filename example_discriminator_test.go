// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/akutz/gdj"
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
