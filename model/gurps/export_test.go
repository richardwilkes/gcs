// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/frequency"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/uti"
)

func TestTemplateFuncs(t *testing.T) {
	c := check.New(t)
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
		c.NoError(err, "Test %d", i)
		var buffer strings.Builder
		c.NoError(tmpl.Execute(&buffer, values), "Test %d", i)
		c.Equal(data.out, buffer.String(), "Test %d", i)
	}
}

func TestExportTraitSelfControlAndFrequency(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()

	plain := NewTrait(entity, nil, false)
	plain.Name = "Plain"

	rolls := NewTrait(entity, nil, false)
	rolls.Name = "Rolls"
	rolls.SelfControl = selfctrl.CR12
	rolls.Frequency = frequency.FR9
	entity.Traits = append(entity.Traits, plain, rolls)

	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "tmpl.txt")
	const tmpl = "GCS Text Template v1\n" +
		"{{range .Traits}}" +
		"<<{{.Description}}|{{.CR}}|{{.CRFull}}|{{.FR}}|{{.FRFull}}|" +
		"{{.ModifierNotes}}|{{.ModifierNotesNoCR}}|{{.ModifierNotesNoFR}}|{{.ModifierNotesNoRolls}}>>\n" +
		"{{end}}"
	c.NoError(os.WriteFile(tmplPath, []byte(tmpl), 0o600))
	outPath := filepath.Join(dir, "out.txt")
	c.NoError(Export(entity, tmplPath, outPath))
	data, err := os.ReadFile(outPath)
	c.NoError(err)
	out := string(data)

	// A trait with no self-control or frequency roll emits nothing for any of the new fields.
	c.Contains(out, "<<Plain|0||0|||||>>")

	// A trait with both rolls set emits the numeric value, the full descriptor, and the modifier-notes variants that
	// individually suppress each roll line.
	cr := "Self-Control Roll (CR): 12 or less (Resist quite often)"
	fr := "Frequency Roll (FR): 9 or less (Fairly often)"
	c.Contains(out, "<<Rolls|12|12 or less (Resist quite often)|9|9 or less (Fairly often)|"+
		cr+"<br>"+fr+"|"+fr+"|"+cr+"|>>")
}

func TestExportSheetsNoExportableFiles(t *testing.T) {
	c := check.New(t)

	// The model tests don't run the ux-layer file-type registration, so register a non-exportable type for the
	// extension used below.
	(&FileInfo{
		Name:         "Test Non-Exportable",
		UTI:          &uti.DataType{Extensions: []string{".unsupported"}},
		IsExportable: false,
	}).Register()

	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "tmpl.txt")
	c.NoError(os.WriteFile(tmplPath, []byte("GCS Text Template v1\n"), 0o600))

	// A non-exportable file must not be reported as a successful export.
	notExportable := filepath.Join(dir, "data.unsupported")
	c.NoError(os.WriteFile(notExportable, []byte("nope"), 0o600))
	c.HasError(ExportSheets(tmplPath, []string{notExportable}))

	// An empty file list also exports nothing and must surface an error rather than a silent success.
	c.HasError(ExportSheets(tmplPath, nil))
}
