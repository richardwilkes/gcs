// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package txtutil_test

import (
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/txtutil"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestAppendOntoNewLine(t *testing.T) {
	for _, tc := range []struct {
		name string
		to   string
		from string
		want string
	}{
		{name: "empty onto empty"},
		{name: "text onto empty", from: "first", want: "first"},
		{name: "empty onto text", to: "first", want: "first"},
		{name: "text onto text", to: "first", from: "second", want: "first\nsecond"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			var to, from strings.Builder
			to.WriteString(tc.to)
			txtutil.AppendStringOntoNewLine(&to, tc.from)
			c.Equal(tc.want, to.String(), "AppendStringOntoNewLine")
			to.Reset()
			to.WriteString(tc.to)
			from.WriteString(tc.from)
			txtutil.AppendBufferOntoNewLine(&to, &from)
			c.Equal(tc.want, to.String(), "AppendBufferOntoNewLine")
		})
	}
}

func TestAppendOntoNewLineIgnoresNilBuffers(t *testing.T) {
	c := check.New(t)
	var nilBuffer *strings.Builder
	var to, from strings.Builder
	to.WriteString("first")
	from.WriteString("second")
	c.NotPanics(func() { txtutil.AppendStringOntoNewLine(nilBuffer, "second") }, "string onto a nil buffer")
	c.NotPanics(func() { txtutil.AppendBufferOntoNewLine(nilBuffer, &from) }, "buffer onto a nil buffer")
	c.NotPanics(func() { txtutil.AppendBufferOntoNewLine(&to, nilBuffer) }, "nil buffer onto a buffer")
	c.Equal("first", to.String())
}
