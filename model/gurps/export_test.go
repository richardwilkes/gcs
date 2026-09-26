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
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/frequency"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/jio"
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
		cr+"\n"+fr+"|"+fr+"|"+cr+"|>>")
}

// TestExportConditionalModifierGroupsAreFlat verifies that the template exporter writes the reactions and conditional
// modifiers out flat -- the group containers are skipped and their members take their place, each naming its group --
// so that templates written before groups existed keep working.
func TestExportConditionalModifierGroupsAreFlat(t *testing.T) {
	withGroupContainersOnSort(t, true)
	c := check.New(t)
	entity := NewEntity()
	addReaction(entity, "A", "from everyone", "", fxp.One)
	addReaction(entity, "B", "from foes", "Combat", fxp.Two)
	addReaction(entity, "C", "from allies", "Combat", fxp.Three)
	addGroupedConditionalModifier(entity, "D", "to hit", "Combat", fxp.One)
	addGroupedConditionalModifier(entity, "E", "to dodge", "", fxp.Two)
	entity.Recalculate()

	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "tmpl.txt")
	const tmpl = "GCS Text Template v1\n" +
		"{{range .Reactions}}<<{{.Situation}}|{{.Group}}|{{.Total}}>>{{end}}\n" +
		"{{range .ConditionalModifiers}}<<{{.Situation}}|{{.Group}}|{{.Total}}>>{{end}}\n"
	c.NoError(os.WriteFile(tmplPath, []byte(tmpl), 0o600))
	outPath := filepath.Join(dir, "out.txt")
	c.NoError(Export(entity, tmplPath, outPath))
	data, err := os.ReadFile(outPath)
	c.NoError(err)
	c.Equal("<<from allies|Combat|3>><<from foes|Combat|2>><<from everyone||1>>\n"+
		"<<to hit|Combat|1>><<to dodge||2>>\n", string(data))
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

// TestExportModifierNotesLineBreaks verifies that a multi-line field such as a trait's modifier notes carries plain
// newlines rather than embedded HTML, so an HTML template escapes only the text and the htmlLines function is what
// turns the newlines into real line breaks.
func TestExportModifierNotesLineBreaks(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	trait := NewTrait(entity, nil, false)
	trait.Name = "Greed"
	trait.SelfControl = selfctrl.CR12
	mod := NewTraitModifier(entity, nil, false)
	mod.Name = `Mitigator <"&">`
	trait.AddModifiers(mod)
	entity.Traits = append(entity.Traits, trait)

	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "tmpl.html")
	const tmpl = "GCS HTML Template v1\n" +
		"{{range .Traits}}|A|{{.ModifierNotes}}|B|{{htmlLines .ModifierNotes}}|C|{{end}}"
	c.NoError(os.WriteFile(tmplPath, []byte(tmpl), 0o600))
	outPath := filepath.Join(dir, "out.html")
	c.NoError(Export(entity, tmplPath, outPath))
	data, err := os.ReadFile(outPath)
	c.NoError(err)
	out := string(data)

	// The raw field must never emit an escaped <br>; the separator is a plain newline.
	c.Equal(0, strings.Count(out, "&lt;br&gt;"), "no escaped line-break markup is emitted")
	c.Contains(out, "|A|Self-Control Roll (CR): 12 or less (Resist quite often)\nMitigator")

	// htmlLines turns that newline into a real line break, while still escaping the text around it.
	c.Contains(out, "|B|Self-Control Roll (CR): 12 or less (Resist quite often)<br>\nMitigator")
	c.Contains(out, "&lt;&#34;&amp;&#34;&gt;|C|")
}

// TestExportAttributeKinds verifies that the Go-template export sorts attributes into the primary, secondary and pool
// lists by their resolved kind: an explicit placement overrides the base-derived classification, and a hidden pool is
// omitted.
func TestExportAttributeKinds(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	addDef := func(def *AttributeDef) {
		def.Order = len(entity.SheetSettings.Attributes.Set)
		entity.SheetSettings.Attributes.Set[def.DefID] = def
		entity.Attributes.Set[def.DefID] = NewAttribute(entity, def.DefID, len(entity.Attributes.Set))
	}
	addDef(&AttributeDef{DefID: "forced", Type: attribute.Integer, Name: "Forced", Base: "$iq", Placement: attribute.Primary})
	addDef(&AttributeDef{DefID: "mana", Type: attribute.Pool, Name: "Mana", Base: "10", Placement: attribute.Hidden})
	entity.Recalculate()

	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "tmpl.txt")
	const tmpl = "GCS Text Template v1\n" +
		"{{range .Attributes.Primary}}<{{.ID}}>{{end}}|{{range .Attributes.Secondary}}<{{.ID}}>{{end}}|" +
		"{{range .Attributes.Pools}}<{{.ID}}:{{.Current}}/{{.Maximum}}>{{end}}"
	c.NoError(os.WriteFile(tmplPath, []byte(tmpl), 0o600))
	outPath := filepath.Join(dir, "out.txt")
	c.NoError(Export(entity, tmplPath, outPath))
	data, err := os.ReadFile(outPath)
	c.NoError(err)
	c.Equal("<st><dx><iq><ht><forced>|"+
		"<will><fright_check><per><vision><hearing><taste_smell><touch><basic_speed><basic_move>|"+
		"<fp:10/10><hp:10/10>", string(data))
}

// TestExportTraitPrereqContradiction verifies that the template export and the "calc" object written to disk carry
// what the trait table shows for a trait caught in a contradiction among the prerequisites, when the sheet enforces
// them: a trait whose own prerequisites are unmet has the contradiction explained within its unsatisfied reason, and
// one whose own prerequisites are met has it explained in a field of its own. The "calc" object is written for every
// trait, so it alone also covers a trait the sheet disabled, which has that explained within its reason; the export
// leaves disabled traits out, as it always has.
func TestExportTraitPrereqContradiction(t *testing.T) {
	c := check.New(t)
	countLogs(t, slog.LevelWarn)
	entity := NewEntity()
	entity.SheetSettings.EnforceTraitPrereqs = true
	requires := newTraitRequiring(entity, "Requires", "Excludes")
	excludes := NewTrait(entity, nil, false)
	excludes.Name = "Excludes"
	excludes.Prereq = newPrereqListForbiddingTrait("Requires")
	unrelated := newTraitNeedingMissingTrait(entity, "Unrelated")
	entity.Traits = append(entity.Traits, requires, excludes, unrelated)
	entity.Recalculate()
	c.True(requires.ContradictedPrereqs(), "precondition: the traits are caught in the contradiction")
	c.True(unrelated.DisabledByPrereqs(), "precondition: the unrelated trait is disabled")

	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "tmpl.txt")
	const tmpl = "GCS Text Template v1\n" +
		"{{range .Traits}}<<{{.Description}}|{{.UnsatisfiedReason}}|{{.PrereqContradiction}}>>\n{{end}}"
	c.NoError(os.WriteFile(tmplPath, []byte(tmpl), 0o600))
	outPath := filepath.Join(dir, "out.txt")
	c.NoError(Export(entity, tmplPath, outPath))
	data, err := os.ReadFile(outPath)
	c.NoError(err)
	out := string(data)
	requiresReason, requiresContradiction := requires.prereqStatus()
	c.Equal("", requiresReason, "the trait whose own prerequisites are met has no unsatisfied reason")
	c.NotEqual("", requiresContradiction, "but is caught in the contradiction")
	c.Contains(out, "<<Requires||"+requiresContradiction+">>", "which the export says in a field of its own")
	excludesReason, excludesContradiction := excludes.prereqStatus()
	c.True(strings.HasPrefix(excludesReason, excludes.UnsatisfiedReason) &&
		len(excludesReason) > len(excludes.UnsatisfiedReason),
		"the trait whose own prerequisites are unmet has the contradiction explained within its reason")
	c.Equal("", excludesContradiction, "and not in the field of its own")
	c.Contains(out, "<<Excludes|"+excludesReason+"|>>", "which the export carries")

	// The "calc" object written to disk says the same for each of them, and alone covers the trait the sheet disabled.
	c.False(unrelated.ContradictedPrereqs(), "precondition: the trait the sheet disabled is not caught in the contradiction")
	unrelatedReason, unrelatedContradiction := unrelated.prereqStatus()
	c.True(strings.HasPrefix(unrelatedReason, unrelated.UnsatisfiedReason) &&
		len(unrelatedReason) > len(unrelated.UnsatisfiedReason),
		"the trait the sheet disabled has that explained within its reason")
	c.Equal("", unrelatedContradiction, "and nothing in the field of its own")
	requiresCalc := savedPrereqStatus(c, requires)
	c.Equal("", requiresCalc.UnsatisfiedReason, "the trait whose own prerequisites are met records no reason")
	c.Equal(requiresContradiction, requiresCalc.PrereqContradiction, "and records the contradiction")
	excludesCalc := savedPrereqStatus(c, excludes)
	c.Equal(excludesReason, excludesCalc.UnsatisfiedReason,
		"the trait whose own prerequisites are unmet records the reason, with the contradiction within it")
	c.Equal("", excludesCalc.PrereqContradiction, "rather than alongside it")
	unrelatedCalc := savedPrereqStatus(c, unrelated)
	c.Equal(unrelatedReason, unrelatedCalc.UnsatisfiedReason,
		"the trait the sheet disabled records the reason, with the sheet's disabling of it explained within")
	c.Equal("", unrelatedCalc.PrereqContradiction, "and nothing in the field of its own")
}

// traitPrereqCalc is the part of a trait's "calc" object that records what the trait table shows for its prerequisites.
type traitPrereqCalc struct {
	UnsatisfiedReason   string `json:"unsatisfied_reason"`
	PrereqContradiction string `json:"prereq_contradiction"`
}

// savedPrereqStatus returns the prerequisite status the "calc" object written to disk for the trait records.
func savedPrereqStatus(c check.Checker, t *Trait) traitPrereqCalc {
	c.Helper()
	saved, err := jio.Marshal(t)
	c.NoError(err)
	var data struct {
		Calc traitPrereqCalc `json:"calc"`
	}
	c.NoError(jio.Unmarshal(saved, &data))
	return data.Calc
}
