// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package canaries

import (
	"time"
)

func mustParseTime(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

func addrOfMustParseTime(layout, value string) *time.Time {
	t := mustParseTime(layout, value)
	return &t
}

func addrOfBool(v bool) *bool {
	return &v
}

func addrOfInt32(v int32) *int32 {
	return &v
}

func addrOfInt64(v int64) *int64 {
	return &v
}
