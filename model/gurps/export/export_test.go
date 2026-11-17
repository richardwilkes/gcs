// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package export

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"text/template"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/frequency"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/uti"
)

// runTemplateExport runs Export over the given template text, format line included, and returns its output.
func runTemplateExport(t *testing.T, c check.Checker, entity *gurps.Entity, tmpl string) string {
	t.Helper()
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "tmpl.txt")
	c.NoError(os.WriteFile(tmplPath, []byte(tmpl), 0o600))
	outPath := filepath.Join(dir, "out.txt")
	c.NoError(Export(entity, tmplPath, outPath))
	data, err := os.ReadFile(outPath)
	c.NoError(err)
	return string(data)
}

// newGroupedModifiersEntity returns a sheet with the reactions A, B and C and the conditional modifiers D and E, some
// of them in the "Combat" group, with the "group containers when sorting" setting on for the rest of the test.
func newGroupedModifiersEntity(t *testing.T) *gurps.Entity {
	gurpstest.WithGroupContainersOnSort(t, true)
	entity := gurps.NewEntity()
	gurpstest.AddReaction(entity, "A", "from everyone", "", fxp.One)
	gurpstest.AddReaction(entity, "B", "from foes", "Combat", fxp.Two)
	gurpstest.AddReaction(entity, "C", "from allies", "Combat", fxp.Three)
	gurpstest.AddGroupedConditionalModifier(entity, "D", "to hit", "Combat", fxp.One)
	gurpstest.AddGroupedConditionalModifier(entity, "E", "to dodge", "", fxp.Two)
	return entity
}

// newAttributeKindsEntity returns a sheet with two extra attributes: "forced", whose base would make it secondary but
// whose placement makes it primary, and "mana", a hidden pool.
func newAttributeKindsEntity() *gurps.Entity {
	entity := gurps.NewEntity()
	addDef := func(def *gurps.AttributeDef) {
		def.Order = len(entity.SheetSettings.Attributes.Set)
		entity.SheetSettings.Attributes.Set[def.DefID] = def
		entity.Attributes.Set[def.DefID] = gurps.NewAttribute(entity, def.DefID, len(entity.Attributes.Set))
	}
	addDef(&gurps.AttributeDef{DefID: "forced", Type: attribute.Integer, Name: "Forced", Base: "$iq", Placement: attribute.Primary})
	addDef(&gurps.AttributeDef{DefID: "mana", Type: attribute.Pool, Name: "Mana", Base: "10", Placement: attribute.Hidden})
	return entity
}

// newContainerTypesEntity returns a sheet carrying the physical container Backpack, the group Kit and the item Rope.
func newContainerTypesEntity() *gurps.Entity {
	entity := gurps.NewEntity()
	backpack := gurps.NewEquipment(entity, nil, true)
	backpack.Name = "Backpack"
	kit := gurps.NewEquipmentGroup(entity, nil)
	kit.Name = "Kit"
	rope := gurps.NewEquipment(entity, nil, false)
	rope.Name = "Rope"
	entity.CarriedEquipment = []*gurps.Equipment{backpack, kit, rope}
	return entity
}

// TestTemplateFuncs verifies that the number functions and the Number methods templates can call give the documented
// results.
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
	for i, data := range []struct {
		in, out string
		fails   bool
	}{
		{in: `{{numberFrom 22}}`, out: "22"},
		{in: `{{numberFrom 23.45}}`, out: "23.45"},
		{in: `{{numberFrom "1.23456"}}`, out: "1.2345"},
		{in: `{{numberFrom "15.5U"}}`, out: "15.5"},
		{in: `{{numberFrom true}}`, fails: true},
		{in: `{{numberToInt .OnePointOne}}`, out: "1"},
		{in: `{{numberToFloat .OnePointOne}}`, out: "1.1"},
		{in: `{{.One.Add .OnePointOne}}`, out: "2.1"},
		{in: `{{.One.Sub .OnePointOne}}`, out: "-0.1"},
	} {
		tmpl, err := tmplBase.Parse(data.in)
		c.NoError(err, "Test %d", i)
		var buffer strings.Builder
		err = tmpl.Execute(&buffer, values)
		if data.fails {
			c.HasError(err, "Test %d", i)
			continue
		}
		c.NoError(err, "Test %d", i)
		c.Equal(data.out, buffer.String(), "Test %d", i)
	}
}

// newGrantedRowsEntity returns a sheet whose trait Training grants the skill Brawling and the spell Light, both at
// level 3, for which the grants give them 4 points.
func newGrantedRowsEntity() *gurps.Entity {
	entity := gurps.NewEntity()
	brawling := gurps.NewSkill(nil, nil, false)
	brawling.Name = "Brawling"
	brawling.Difficulty.Difficulty = difficulty.Easy
	skillGrant := gurps.NewGrant(feature.GrantSkill)
	skillGrant.Skill = brawling
	skillGrant.Amount = fxp.Three
	light := gurps.NewSpell(nil, nil, false)
	light.Name = "Light"
	spellGrant := gurps.NewGrant(feature.GrantSpell)
	spellGrant.Spell = light
	spellGrant.Amount = fxp.Three
	training := gurps.NewTrait(entity, nil, false)
	training.Name = "Training"
	training.Features = gurps.Features{skillGrant, spellGrant}
	entity.Traits = append(entity.Traits, training)
	return entity
}

// TestExportGrantedRowsCostNothing verifies that both exporters write a granted skill's or spell's points as nothing,
// as the sheet shows them, rather than as the points its grant gives it.
func TestExportGrantedRowsCostNothing(t *testing.T) {
	c := check.New(t)
	entity := newGrantedRowsEntity()
	c.Equal("<<Brawling|0|12>><<Light|0|10>>", runTemplateExport(t, c, entity, "GCS Text Template v1\n"+
		"{{range .Skills}}<<{{.Description}}|{{.Points}}|{{.Level}}>>{{end}}"+
		"{{range .Spells}}<<{{.Description}}|{{.Points}}|{{.Level}}>>{{end}}"))
	c.Equal("<Brawling|0|12><Light|0|10>", runLegacyExport(t, c, entity,
		"@SKILLS_LOOP_START<@DESCRIPTION|@POINTS|@SL>@SKILLS_LOOP_END"+
			"@SPELLS_LOOP_START<@DESCRIPTION_PRIMARY|@POINTS|@SL>@SPELLS_LOOP_END"))
}

// TestExportTraitSelfControlAndFrequency verifies that the template exporter writes a trait's self-control and
// frequency rolls as numbers and descriptions, and offers its modifier notes with either or both rolls left out.
func TestExportTraitSelfControlAndFrequency(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()

	plain := gurps.NewTrait(entity, nil, false)
	plain.Name = "Plain"

	rolls := gurps.NewTrait(entity, nil, false)
	rolls.Name = "Rolls"
	rolls.SelfControl = selfctrl.CR12
	rolls.Frequency = frequency.FR9
	entity.Traits = append(entity.Traits, plain, rolls)

	out := runTemplateExport(t, c, entity, "GCS Text Template v1\n"+
		"{{range .Traits}}"+
		"<<{{.Description}}|{{.CR}}|{{.CRFull}}|{{.FR}}|{{.FRFull}}|"+
		"{{.ModifierNotes}}|{{.ModifierNotesNoCR}}|{{.ModifierNotesNoFR}}|{{.ModifierNotesNoRolls}}>>\n"+
		"{{end}}")

	// A trait with neither roll emits 0 for the numeric fields and nothing for the rest.
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
	c := check.New(t)
	c.Equal("<<from allies|Combat|3>><<from foes|Combat|2>><<from everyone||1>>\n"+
		"<<to hit|Combat|1>><<to dodge||2>>\n", runTemplateExport(t, c, newGroupedModifiersEntity(t),
		"GCS Text Template v1\n"+
			"{{range .Reactions}}<<{{.Situation}}|{{.Group}}|{{.Total}}>>{{end}}\n"+
			"{{range .ConditionalModifiers}}<<{{.Situation}}|{{.Group}}|{{.Total}}>>{{end}}\n"))
}

var registerTestFileTypesOnce sync.Once

// registerTestFileTypes registers the file types TestSheets uses, since the ux package registers the real ones: an
// exportable type for sheets and one that isn't exportable.
func registerTestFileTypes() {
	registerTestFileTypesOnce.Do(func() {
		(&gurps.FileInfo{
			Name:         "Test Sheet",
			UTI:          &uti.DataType{Extensions: []string{gurps.SheetExt}},
			IsExportable: true,
		}).Register()
		(&gurps.FileInfo{
			Name: "Test Non-Exportable",
			UTI:  &uti.DataType{Extensions: []string{".unsupported"}},
		}).Register()
	})
}

// TestSheets verifies that Sheets exports each exportable file beside it, named after it with the template's extension,
// skips the files that aren't exportable, and fails when it exports nothing or can't export a file.
func TestSheets(t *testing.T) {
	registerTestFileTypes()
	for _, tc := range []struct {
		name            string
		files           []string
		missingTemplate bool
		wantErr         bool
		errIs           error
	}{
		{name: "only non-exportable files", files: []string{"data.unsupported"}, wantErr: true},
		{name: "no files", wantErr: true},
		{name: "sheet after a non-exportable file", files: []string{"data.unsupported", "Dai.gcs"}},
		// The loadable sheet is listed too, so that skipping the unloadable one would not fail for exporting nothing.
		{name: "unloadable sheet", files: []string{"garbage.gcs", "Dai.gcs"}, wantErr: true},
		{name: "missing template", files: []string{"Dai.gcs"}, missingTemplate: true, errIs: fs.ErrNotExist},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			dir := t.TempDir()
			entity := gurps.NewEntity()
			entity.Profile.Name = "Dai"
			c.NoError(entity.Save(filepath.Join(dir, "Dai.gcs")))
			c.NoError(os.WriteFile(filepath.Join(dir, "garbage.gcs"), []byte("garbage"), 0o600))
			c.NoError(os.WriteFile(filepath.Join(dir, "data.unsupported"), []byte("nope"), 0o600))
			// The template sits apart from the sheets, so that the export is seen to land beside its source.
			tmplPath := filepath.Join(t.TempDir(), "tmpl.txt")
			if !tc.missingTemplate {
				c.NoError(os.WriteFile(tmplPath, []byte("GCS Text Template v1\n{{.Name}}"), 0o600))
			}
			files := make([]string, 0, len(tc.files))
			for _, one := range tc.files {
				files = append(files, filepath.Join(dir, one))
			}
			err := Sheets(tmplPath, files)
			switch {
			case tc.errIs != nil:
				c.True(errors.Is(err, tc.errIs), "got %v", err)
			case tc.wantErr:
				c.HasError(err)
			default:
				c.NoError(err)
				data, readErr := os.ReadFile(filepath.Join(dir, "Dai.txt"))
				c.NoError(readErr)
				c.Equal("Dai", string(data))
			}
		})
	}
}

// TestExportErrors verifies that Export fails when it can't read, parse or execute the template, or can't create the
// output file.
func TestExportErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		tmpl    string // When empty, no template file is written.
		outPath string // Relative to the test's directory.
		errIs   error
	}{
		{name: "missing template", outPath: "out.txt", errIs: fs.ErrNotExist},
		{name: "text template parse failure", tmpl: "GCS Text Template v1\n{{if}}", outPath: "out.txt"},
		{name: "HTML template parse failure", tmpl: "GCS HTML Template v1\n{{if}}", outPath: "out.txt"},
		{name: "execution failure", tmpl: "GCS Text Template v1\n{{.NoSuchField}}", outPath: "out.txt"},
		{
			name:    "output path not creatable",
			tmpl:    "GCS Text Template v1\n{{.Name}}",
			outPath: filepath.Join("missing", "out.txt"),
			errIs:   fs.ErrNotExist,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			dir := t.TempDir()
			tmplPath := filepath.Join(dir, "tmpl.txt")
			if tc.tmpl != "" {
				c.NoError(os.WriteFile(tmplPath, []byte(tc.tmpl), 0o600))
			}
			err := Export(gurps.NewEntity(), tmplPath, filepath.Join(dir, tc.outPath))
			if tc.errIs != nil {
				c.True(errors.Is(err, tc.errIs), "got %v", err)
			} else {
				c.HasError(err)
			}
		})
	}
}

// TestExportModifierNotesLineBreaks verifies that a multi-line field such as a trait's modifier notes carries plain
// newlines rather than embedded HTML, so an HTML template escapes only the text and the htmlLines function is what
// turns the newlines into real line breaks.
func TestExportModifierNotesLineBreaks(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Greed"
	trait.SelfControl = selfctrl.CR12
	mod := gurps.NewTraitModifier(entity, nil, false)
	mod.Name = `Mitigator <"&">`
	trait.AddModifiers(mod)
	entity.Traits = []*gurps.Trait{trait}
	c.Equal("|A|Self-Control Roll (CR): 12 or less (Resist quite often)\nMitigator &lt;&#34;&amp;&#34;&gt;"+
		"|B|Self-Control Roll (CR): 12 or less (Resist quite often)<br>\nMitigator &lt;&#34;&amp;&#34;&gt;|C|",
		runTemplateExport(t, c, entity, "GCS HTML Template v1\n"+
			"{{range .Traits}}|A|{{.ModifierNotes}}|B|{{htmlLines .ModifierNotes}}|C|{{end}}"))
}

// TestExportAttributeKinds verifies that the Go-template export sorts attributes into the primary, secondary and pool
// lists by their resolved kind: an explicit placement overrides the base-derived classification, and a hidden pool is
// omitted.
func TestExportAttributeKinds(t *testing.T) {
	c := check.New(t)
	c.Equal("<st><dx><iq><ht><forced>|"+
		"<will><fright_check><per><vision><hearing><taste_smell><touch><basic_speed><basic_move>|"+
		"<fp:10/10><hp:10/10>", runTemplateExport(t, c, newAttributeKindsEntity(), "GCS Text Template v1\n"+
		"{{range .Attributes.Primary}}<{{.ID}}>{{end}}|{{range .Attributes.Secondary}}<{{.ID}}>{{end}}|"+
		"{{range .Attributes.Pools}}<{{.ID}}:{{.Current}}/{{.Maximum}}>{{end}}"))
}

// TestExportTraitPrereqContradiction verifies that the template export carries what the trait table shows for traits
// caught in a contradiction among the prerequisites the sheet enforces: a trait whose own prerequisites are unmet has
// the contradiction explained within its unsatisfied reason, and one whose own prerequisites are met has it explained
// in a field of its own. A trait the sheet disabled is left out.
func TestExportTraitPrereqContradiction(t *testing.T) {
	c := check.New(t)
	gurpstest.CountLogs(t, slog.LevelWarn)
	entity := gurps.NewEntity()
	entity.SheetSettings.EnforceTraitPrereqs = true
	requires := gurpstest.NewTraitRequiring(entity, "Requires", "Excludes")
	excludes := gurps.NewTrait(entity, nil, false)
	excludes.Name = "Excludes"
	excludes.Prereq = gurpstest.NewPrereqListForbiddingTrait("Requires")
	unrelated := gurpstest.NewTraitNeedingMissingTrait(entity, "Unrelated")
	entity.Traits = []*gurps.Trait{requires, excludes, unrelated}
	out := runTemplateExport(t, c, entity, "GCS Text Template v1\n"+
		"{{range .Traits}}<<{{.Description}}|{{.UnsatisfiedReason}}|{{.PrereqContradiction}}>>\n{{end}}")
	_, contradiction := requires.PrereqStatus()
	c.NotEqual("", contradiction, "precondition: Requires is caught in the contradiction")
	excludesReason, _ := excludes.PrereqStatus()
	c.True(len(excludesReason) > len(excludes.UnsatisfiedReason),
		"precondition: the reason Excludes is unsatisfied explains the contradiction")
	c.Equal("<<Requires||"+contradiction+">>\n<<Excludes|"+excludesReason+"|>>\n", out)
}

// TestExportEquipmentContainerType verifies that the template exporter tells a group from a physical container, as it
// tells trait containers apart by their type.
func TestExportEquipmentContainerType(t *testing.T) {
	c := check.New(t)
	c.Equal("<<Backpack|container>><<Kit|group>><<Rope|item>>", runTemplateExport(t, c, newContainerTypesEntity(),
		"GCS Text Template v1\n{{range .Equipment.Carried}}<<{{.Description}}|{{.Type}}>>{{end}}"))
}

// TestExportFormatLineTolerance verifies that a byte order mark or whitespace around the format line, which some
// editors add, doesn't keep it from being recognized, while a line that differs in anything else still selects the
// legacy processor.
func TestExportFormatLineTolerance(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	entity.Profile.Name = "Dai"
	for i, data := range []struct {
		firstLine string
		expected  string
	}{
		{firstLine: "\uFEFFGCS Text Template v1", expected: "Dai"},
		{firstLine: "GCS Text Template v1 ", expected: "Dai"},
		{firstLine: " \tGCS Text Template v1\t", expected: "Dai"},
		{firstLine: "\uFEFFGCS HTML Template v1 ", expected: "Dai"},
		{firstLine: "GCS Text Template v2", expected: "GCS Text Template v2\n{{.Name}}"},
	} {
		c.Equal(data.expected, runTemplateExport(t, c, entity, data.firstLine+"\n{{.Name}}"), "Test %d", i)
	}
}
