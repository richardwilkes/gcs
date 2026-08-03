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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/toolbox/v2/check"
)

// runLegacyExport runs the legacy (non-Go-template) exporter over the given template text and returns the resulting
// output.
func runLegacyExport(t *testing.T, c check.Checker, entity *Entity, tmpl string) string {
	t.Helper()
	entity.Recalculate()
	outPath := filepath.Join(t.TempDir(), "out.txt")
	c.NoError(legacyTextExport(entity, []byte(tmpl), outPath))
	data, err := os.ReadFile(outPath)
	c.NoError(err)
	return string(data)
}

// TestLegacyExportPerceptionPoints verifies that @PERCEPTION_POINTS resolves against the "per" attribute ID rather than
// a non-existent "perception" ID, which would always yield 0.
func TestLegacyExportPerceptionPoints(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.Attributes.Set["per"].Adjustment = fxp.Two // "per" costs 5/point, so 2 levels == 10 points
	c.Equal("10", runLegacyExport(t, c, e, "@PERCEPTION_POINTS"))
	c.Equal(e.Attributes.Cost("per").String(), runLegacyExport(t, c, e, "@PERCEPTION_POINTS"))
}

// TestLegacyExportOptionalParensEncoding verifies that the DESCRIPTION_NOTES* family encodes its text like every other
// text-emitting key rather than dumping it raw into the output.
func TestLegacyExportOptionalParensEncoding(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	s := NewSkill(e, nil, false)
	s.Name = "Brawling"
	s.LocalNotes = "a<b & \"c\"\nd"
	e.Skills = append(e.Skills, s)
	c.Equal("[ (a&lt;b &amp; &quot;c&quot;<br>d)]",
		runLegacyExport(t, c, e, "@SKILLS_LOOP_START[@DESCRIPTION_NOTES_PAREN]@SKILLS_LOOP_END"))

	// With encoding turned off, the text passes through untouched.
	c.Equal("[ [a<b & \"c\"\nd]]",
		runLegacyExport(t, c, e, "@ENCODING_OFF@SKILLS_LOOP_START[@DESCRIPTION_NOTES_BRACKET]@SKILLS_LOOP_END"))
}

// TestLegacyExportSkillsLoopCount verifies that @SKILLS_LOOP_COUNT counts the same nodes that @SKILLS_LOOP_START
// iterates over, i.e. containers are included in both.
func TestLegacyExportSkillsLoopCount(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	group := NewSkill(e, nil, true)
	group.Name = "Group"
	child := NewSkill(e, group, false)
	child.Name = "Child"
	group.Children = append(group.Children, child)
	e.Skills = append(e.Skills, group)
	c.Equal("2|<GROUP><ITEM>", runLegacyExport(t, c, e, "@SKILLS_LOOP_COUNT|@SKILLS_LOOP_START<@TYPE>@SKILLS_LOOP_END"))
}

// TestLegacyExportLoopBodyKeyDetection verifies that the decision to treat '@' as the start of a key inside a loop body
// is made from the loop body itself, not from a fixed byte in the outer template.
func TestLegacyExportLoopBodyKeyDetection(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	n := NewNote(e, nil, false)
	n.MarkDown = "Note1"
	e.Notes = append(e.Notes, n)

	// "@7" is literal text at the top level, so it must be literal text inside a loop body too.
	c.Equal("[@7]", runLegacyExport(t, c, e, "@NOTES_LOOP_START[@7]@NOTES_LOOP_END"))

	// A digit following the loop-end marker must not turn every key in the loop body into literal text.
	c.Equal("[Note1]7", runLegacyExport(t, c, e, "@NOTES_LOOP_START[@NOTE]@NOTES_LOOP_END7"))
}

// TestLegacyExportNotesSeparator verifies that the separator between top-level notes is encoded along with the notes
// themselves, so that HTML exports keep the blank line between them.
func TestLegacyExportNotesSeparator(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	first := NewNote(e, nil, false)
	first.MarkDown = "First"
	second := NewNote(e, nil, false)
	second.MarkDown = "Second"
	e.Notes = append(e.Notes, first, second)
	c.Equal("First<br><br>Second", runLegacyExport(t, c, e, "@NOTES"))
	c.Equal("First\n\nSecond", runLegacyExport(t, c, e, "@ENCODING_OFF@NOTES"))
}

// TestLegacyExportOtherEquipmentLoopMarker verifies that the other-equipment loop terminates on its own loop-end marker
// rather than on the carried-equipment one, which is a substring of it.
func TestLegacyExportOtherEquipmentLoopMarker(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	eqp := NewEquipment(e, nil, false)
	eqp.Name = "Rock"
	e.OtherEquipment = append(e.OtherEquipment, eqp)
	c.Equal("|EQUIPMENT_LOOP_END|Rock|", runLegacyExport(t, c, e,
		"@OTHER_EQUIPMENT_LOOP_START|EQUIPMENT_LOOP_END|@DESCRIPTION|@OTHER_EQUIPMENT_LOOP_END"))
}

// TestLegacyExportAttributeNameSuffixes verifies that the more specific attribute key suffixes are matched before the
// less specific ones, and that a suffix match that doesn't resolve to an attribute falls through to the next candidate.
func TestLegacyExportAttributeNameSuffixes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	c.Equal("ST|Strength|Strength (ST)|10", runLegacyExport(t, c, e,
		"@ST_NAME|@ST_FULL_NAME|@ST_COMBINED_NAME|@ST_CURRENT"))

	// An attribute whose own ID ends with "_full" must still resolve via the "_name" suffix.
	e.SheetSettings.Attributes.Set["hit_full"] = &AttributeDef{
		AttributeDefData: AttributeDefData{
			DefID: "hit_full",
			Type:  attribute.Integer,
			Name:  "Hit Full",
			Base:  "3",
		},
		Order: len(e.SheetSettings.Attributes.Set),
	}
	e.Attributes.Set["hit_full"] = NewAttribute(e, "hit_full", len(e.Attributes.Set))
	c.Equal("Hit Full|3", runLegacyExport(t, c, e, "@HIT_FULL_NAME|@HIT_FULL_CURRENT"))
}
