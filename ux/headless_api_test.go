// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//go:build headlessapi

package ux

import (
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

func TestNegotiateDocFormat(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name   string
		accept string
		want   docFormat
	}{
		{name: "missing", want: docFormatMD},
		{name: "anything", accept: "*/*", want: docFormatMD},
		{name: "unknown type", accept: "text/html", want: docFormatMD},
		{name: "json", accept: "application/json", want: docFormatJSON},
		{name: "yaml", accept: "application/yaml", want: docFormatYAML},
		{name: "yaml, the other spelling", accept: "application/x-yaml", want: docFormatYAML},
		{name: "markdown", accept: "text/markdown", want: docFormatMD},
		{name: "highest q wins", accept: "application/json;q=0.2, application/yaml;q=0.8", want: docFormatYAML},
		{name: "a tie goes to the first listed", accept: "application/yaml, application/json", want: docFormatYAML},
		{name: "q defaults to 1", accept: "application/json;q=0.9, application/yaml", want: docFormatYAML},
		{name: "spaces and other params", accept: " application/json ; charset=utf-8 ; q=0.9 ", want: docFormatJSON},
		// A q of 0 means "not acceptable", so it must not win merely by being the only format listed.
		{name: "q of 0 is refused", accept: "application/json;q=0", want: docFormatMD},
		{
			name: "q of 0 loses to a lower listed format", accept: "application/json;q=0, application/yaml;q=0.1",
			want: docFormatYAML,
		},
	} {
		c.Equal(tc.want, negotiateDocFormat(tc.accept), tc.name)
	}
}
