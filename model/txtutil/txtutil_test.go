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

func TestAppendStringOntoNewLine(t *testing.T) {
	c := check.New(t)
	var buffer strings.Builder
	txtutil.AppendStringOntoNewLine(&buffer, "")
	c.Equal("", buffer.String())
	txtutil.AppendStringOntoNewLine(&buffer, "first")
	c.Equal("first", buffer.String())
	txtutil.AppendStringOntoNewLine(&buffer, "")
	c.Equal("first", buffer.String())
	txtutil.AppendStringOntoNewLine(&buffer, "second")
	c.Equal("first\nsecond", buffer.String())
	var nilBuffer *strings.Builder
	txtutil.AppendStringOntoNewLine(nilBuffer, "ignored")
	c.Nil(nilBuffer)
}

func TestAppendBufferOntoNewLine(t *testing.T) {
	c := check.New(t)
	var to, from strings.Builder
	txtutil.AppendBufferOntoNewLine(&to, &from)
	c.Equal("", to.String())
	from.WriteString("first")
	txtutil.AppendBufferOntoNewLine(&to, &from)
	c.Equal("first", to.String())
	from.Reset()
	txtutil.AppendBufferOntoNewLine(&to, &from)
	c.Equal("first", to.String())
	from.WriteString("second")
	txtutil.AppendBufferOntoNewLine(&to, &from)
	c.Equal("first\nsecond", to.String())
	var nilBuffer *strings.Builder
	txtutil.AppendBufferOntoNewLine(nilBuffer, &from)
	txtutil.AppendBufferOntoNewLine(&to, nilBuffer)
	c.Equal("first\nsecond", to.String())
}
