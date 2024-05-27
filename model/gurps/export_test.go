// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
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
	tmplBase := template.New("").Funcs(createTemplateFuncs())
	tmpl, err := tmplBase.Parse(`
{{numberFrom 22}}
{{numberFrom 23.45}}
{{numberFrom "1"}}
{{numberFrom "1.23456"}}
{{numberToInt .One}}
{{numberToFloat .One}}
{{numberToInt .OnePointOne}}
{{numberToFloat .OnePointOne}}
{{.One.Add .OnePointOne}}
{{.One.Sub .OnePointOne}}
{{(numberFrom 22).Add (numberFrom 44.4)}}
`)
	check.NoError(t, err)
	var buffer strings.Builder
	input := struct {
		One         fxp.Int
		OnePointOne fxp.Int
	}{
		One:         fxp.One,
		OnePointOne: fxp.OnePointOne,
	}
	check.NoError(t, tmpl.Execute(&buffer, input))
	check.Equal(t, `
22
23.45
1
1.2345
1
1
1
1.1
2.1
-0.1
66.4
`, buffer.String())

	buffer.Reset()
	tmpl, err = tmplBase.Parse(`{{numberFrom "x"}}`)
	check.NoError(t, err)
	check.Error(t, tmpl.Execute(&buffer, nil))
}
