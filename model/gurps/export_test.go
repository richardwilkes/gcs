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
	"strings"
	"testing"
	"text/template"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/check"
)

func TestTemplateFuncs(t *testing.T) {
	values := struct {
		One         fxp.Int
		OnePointOne fxp.Int
	}{
		One:         fxp.One,
		OnePointOne: fxp.OnePointOne,
	}
	tmplBase := template.New("").Funcs(createTemplateFuncs())
	for i, data := range []struct{ in, out string }{
		{in: `{{numberFrom 22}}`, out: "22"},
		{in: `{{numberFrom 23.45}}`, out: "23.45"},
		{in: `{{numberFrom "1"}}`, out: "1"},
		{in: `{{numberFrom "1.23456"}}`, out: "1.2345"},
		{in: `{{numberFrom "15U"}}`, out: "15"},
		{in: `{{numberFrom "15.5U"}}`, out: "15.5"},
		{in: `{{numberToInt .One}}`, out: "1"},
		{in: `{{numberToFloat .One}}`, out: "1"},
		{in: `{{numberToInt .OnePointOne}}`, out: "1"},
		{in: `{{numberToFloat .OnePointOne}}`, out: "1.1"},
		{in: `{{.One.Add .OnePointOne}}`, out: "2.1"},
		{in: `{{.One.Sub .OnePointOne}}`, out: "-0.1"},
		{in: `{{(numberFrom 22).Add (numberFrom 44.4)}}`, out: "66.4"},
	} {
		tmpl, err := tmplBase.Parse(data.in)
		check.NoError(t, err, "Test %d", i)
		var buffer strings.Builder
		check.NoError(t, tmpl.Execute(&buffer, values), "Test %d", i)
		check.Equal(t, data.out, buffer.String(), "Test %d", i)
	}
}
