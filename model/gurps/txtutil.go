// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"github.com/richardwilkes/toolbox"
)

// LineBuilder is an interface that can be used to build a line of text.
type LineBuilder interface {
	Len() int
	WriteByte(ch byte) error
	WriteString(s string) (int, error)
	String() string
}

// AppendBufferOntoNewLine appends the contents of the 'from' buffer onto the 'to' buffer, starting on a new line.
func AppendBufferOntoNewLine(to, from LineBuilder) {
	if toolbox.IsNil(to) || toolbox.IsNil(from) || from.Len() == 0 {
		return
	}
	if to.Len() != 0 {
		to.WriteByte('\n') //nolint:errcheck // Writing a byte to a buffer can't fail.
	}
	to.WriteString(from.String()) //nolint:errcheck // Writing a byte to a buffer can't fail.
}

// AppendStringOntoNewLine appends the contents of the 'from' string onto the 'to' buffer, starting on a new line.
func AppendStringOntoNewLine(to LineBuilder, from string) {
	if toolbox.IsNil(to) || from == "" {
		return
	}
	if to.Len() != 0 {
		to.WriteByte('\n') //nolint:errcheck // Writing a byte to a buffer can't fail.
	}
	to.WriteString(from) //nolint:errcheck // Writing a byte to a buffer can't fail.
}
