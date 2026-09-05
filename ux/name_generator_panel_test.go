// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/namegen"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	uncheck "github.com/richardwilkes/unison/enums/check"
)

// hasWidget reports whether the editor has a widget with the given reference key.
func hasWidget(d structuralEditor, key string) bool {
	return d.targetManager().Find(key) != nil
}

// trainingNamesPanel returns the panel editing the generator's training names, or nil when there is none.
func trainingNamesPanel(d *nameGeneratorEditorDockable, g *gurps.NameGenerator) *weightedStringOptionsPanel {
	for _, p := range panelsOfType[*weightedStringOptionsPanel](d.AsPanel()) {
		if p.spec.list == &g.Entries {
			return p
		}
	}
	return nil
}

// TestNameGeneratorPanelFieldsFollowType verifies that the fields shown depend on the type: every type has the type
// popup and the two case checkboxes; a compound generator adds the separator and the list of generators; a Markov
// letter generator adds the depth; and every type but compound has the built-in training data popup and the training
// names. Changing the type through the popup writes the model, rebuilds the content and is undoable.
func TestNameGeneratorPanelFieldsFollowType(t *testing.T) {
	c := check.New(t)
	g := gurps.NewNameGenerator()
	d := newTestNameGeneratorEditorDockable(g)
	key := g.KeyPrefix
	c.True(hasWidget(d, key+"type"))
	c.True(hasWidget(d, key+"lowered"))
	c.True(hasWidget(d, key+"first_upper"))
	c.True(hasWidget(d, key+"built_in"), "a simple generator takes training data")
	c.NotNil(trainingNamesPanel(d, g), "a simple generator with no built-in set lists its training names")
	c.False(hasWidget(d, key+"separator"), "a simple generator has no separator")
	c.False(hasWidget(d, key+"depth"), "a simple generator has no depth")

	widgetFor[*Popup[namegen.Type]](t, d, key+"type").Select(namegen.Compound)
	c.Equal(namegen.Compound, d.model.Type, "the popup writes the model")
	c.True(d.Modified())
	c.True(hasWidget(d, key+"separator"), "a compound generator has a separator")
	c.Nil(rootNameGeneratorPanel(t, d).rows, "a compound generator with no children has no rows container")
	c.False(hasWidget(d, key+"built_in"), "a compound generator takes no training data")
	c.Nil(trainingNamesPanel(d, g))
	c.False(hasWidget(d, key+"depth"))
	c.Equal(namegen.Compound, selectedType(widgetFor[*Popup[namegen.Type]](t, d, key+"type")),
		"the rebuilt popup shows the new type")

	widgetFor[*Popup[namegen.Type]](t, d, key+"type").Select(namegen.MarkovLetter)
	c.Equal(namegen.MarkovLetter, d.model.Type)
	c.True(hasWidget(d, key+"depth"), "a Markov letter generator has a depth")
	c.True(hasWidget(d, key+"built_in"))
	c.NotNil(trainingNamesPanel(d, g))
	c.False(hasWidget(d, key+"separator"))

	widgetFor[*Popup[namegen.Type]](t, d, key+"type").Select(namegen.MarkovRun)
	c.Equal(namegen.MarkovRun, d.model.Type)
	c.False(hasWidget(d, key+"depth"), "only the Markov letter generator has a depth")
	c.True(hasWidget(d, key+"built_in"))
	c.NotNil(trainingNamesPanel(d, g))

	d.undoMgr.Undo()
	c.Equal(namegen.MarkovLetter, d.model.Type, "undo restores the previous type")
	c.True(hasWidget(d, key+"depth"), "and rebuilds the content for it")
	d.undoMgr.Undo()
	d.undoMgr.Undo()
	c.Equal(namegen.Simple, d.model.Type)
	c.False(d.undoMgr.CanUndo(), "each type change is its own edit")
	c.False(d.Modified())
}

func selectedType(p *Popup[namegen.Type]) namegen.Type {
	value, _ := p.Selected()
	return value
}

// TestNameGeneratorPanelBuiltInHidesTrainingNames verifies that choosing a built-in training set removes the training
// names list, since the list would not be used, and that choosing None brings it back with the names intact.
func TestNameGeneratorPanelBuiltInHidesTrainingNames(t *testing.T) {
	c := check.New(t)
	g := simpleNameGenerator("Alice", "Bob")
	d := newTestNameGeneratorEditorDockable(g)
	c.NotNil(trainingNamesPanel(d, g))

	widgetFor[*Popup[namegen.Builtin]](t, d, g.KeyPrefix+"built_in").Select(namegen.AmericanLast)
	c.Equal(namegen.AmericanLast, d.model.BuiltIn, "the popup writes the model")
	c.Nil(trainingNamesPanel(d, g), "a built-in set hides the training names")
	c.Equal(2, len(d.model.Entries), "but keeps them")

	widgetFor[*Popup[namegen.Builtin]](t, d, g.KeyPrefix+"built_in").Select(namegen.None)
	c.Equal(namegen.None, d.model.BuiltIn)
	c.NotNil(trainingNamesPanel(d, g), "None shows the training names again")
	c.Equal([]string{"Alice", "Bob"}, optionValues(d.model.Entries))
}

// TestNameGeneratorPanelCaseCheckboxes verifies that the positively phrased checkboxes are bound to the negated model
// fields in both directions: a checked box means the option is on, which the model records as the "No" field being
// false.
func TestNameGeneratorPanelCaseCheckboxes(t *testing.T) {
	c := check.New(t)
	g := gurps.NewNameGenerator()
	g.NoFirstToUpper = true
	d := newTestNameGeneratorEditorDockable(g)
	lowered := widgetFor[*CheckBox](t, d, g.KeyPrefix+"lowered")
	firstUpper := widgetFor[*CheckBox](t, d, g.KeyPrefix+"first_upper")
	c.Equal(uncheck.On, lowered.State, "lowering is on when NoLowered is false")
	c.Equal(uncheck.Off, firstUpper.State, "capitalizing is off when NoFirstToUpper is true")

	lowered.State = uncheck.Off
	lowered.ClickCallback()
	c.True(g.NoLowered, "unchecking turns the option off")
	firstUpper.State = uncheck.On
	firstUpper.ClickCallback()
	c.False(g.NoFirstToUpper, "checking turns the option on")
	c.True(d.Modified())

	d.undoMgr.Undo()
	c.True(g.NoFirstToUpper, "undo restores the model")
	c.Equal(uncheck.Off, widgetFor[*CheckBox](t, d, g.KeyPrefix+"first_upper").State, "and the checkbox")
}

// TestNameGeneratorPanelDepthField verifies that the depth field is bound to the model and accepts zero, which stands
// for the default, so that a file which leaves the depth out is not altered by being opened.
func TestNameGeneratorPanelDepthField(t *testing.T) {
	c := check.New(t)
	g := gurps.NewNameGenerator()
	g.Type = namegen.MarkovLetter
	d := newTestNameGeneratorEditorDockable(g)
	field := integerFieldFor(t, d, g.KeyPrefix+"depth")
	c.Equal("0", field.Text(), "an omitted depth shows as zero")
	field.SetText("4")
	c.Equal(4, g.Depth)
	field.hasFocus = true
	field.SetText("6")
	c.Equal(5, g.Depth, "the depth is held at its maximum")
	c.Equal("Value must be no more than 5", tooltipText(field.Tooltip))
	field.SetText("0")
	c.Equal(0, g.Depth, "zero is accepted")
	c.Contains(tooltipText(field.Tooltip), "default of 3", "zero is not flagged")
}

// TestNameGeneratorPanelCompoundAddRemove verifies that adding a generator to a compound generator appends a simple one
// with a key prefix of its own, builds a row with a nested editor for it, and focuses nothing it cannot find; that
// removing works by identity; and that both are undoable.
func TestNameGeneratorPanelCompoundAddRemove(t *testing.T) {
	c := check.New(t)
	g := gurps.NewNameGenerator()
	g.Type = namegen.Compound
	d := newTestNameGeneratorEditorDockable(g)
	root := rootNameGeneratorPanel(t, d)
	c.Nil(root.rows, "an empty compound generator has no rows container")

	root.addGenerator()
	c.Equal(1, len(d.model.Compound))
	first := d.model.Compound[0]
	c.Equal(namegen.Simple, first.Type, "a new generator is simple")
	c.NotEqual("", first.KeyPrefix, "the new generator has a key prefix")
	c.NotEqual(g.KeyPrefix, first.KeyPrefix)
	root = rootNameGeneratorPanel(t, d)
	c.NotNil(root.rows, "the rows container appears with the first generator")
	c.Equal(1, len(root.rows.Children()))
	c.True(hasWidget(d, first.KeyPrefix+"type"), "the nested editor has a type popup")
	c.True(hasWidget(d, first.KeyPrefix+"built_in"), "the nested editor takes training data")
	c.True(d.Modified())

	rootNameGeneratorPanel(t, d).addGenerator()
	c.Equal(2, len(d.model.Compound))
	second := d.model.Compound[1]
	widgetFor[*Popup[namegen.Type]](t, d, second.KeyPrefix+"type").Select(namegen.MarkovRun)
	c.Equal(namegen.MarkovRun, second.Type, "a nested type popup writes the nested generator")
	c.Equal(namegen.Simple, d.model.Compound[0].Type, "and not its sibling")

	rows := panelsOfType[*compoundGeneratorPanel](d.AsPanel())
	c.Equal(2, len(rows))
	c.True(rows[0].deleteButton.Enabled())
	rows[0].remove()
	c.Equal(1, len(d.model.Compound), "the first generator is removed")
	c.Equal(namegen.MarkovRun, d.model.Compound[0].Type, "leaving the second")
	panelsOfType[*compoundGeneratorPanel](d.AsPanel())[0].remove()
	c.Equal(0, len(d.model.Compound), "the last generator may be removed")
	c.Nil(rootNameGeneratorPanel(t, d).rows, "the rows container is gone again")

	d.undoMgr.Undo()
	c.Equal(1, len(d.model.Compound), "undo restores the last removal")
	d.undoMgr.Undo()
	c.Equal(2, len(d.model.Compound), "and the one before it")
	d.undoMgr.Undo()
	d.undoMgr.Undo()
	d.undoMgr.Undo()
	c.Equal(0, len(d.model.Compound), "each edit is its own undo")
	c.False(d.undoMgr.CanUndo())
	c.False(d.Modified())
}

// TestNameGeneratorPanelTrainingNamesAddRemoveReorder verifies that the shared weighted string list panel edits the
// training names: adding, removing and drag reordering all reach the model through the editor's undo.
func TestNameGeneratorPanelTrainingNamesAddRemoveReorder(t *testing.T) {
	c := check.New(t)
	g := simpleNameGenerator("Alice", "Bob")
	d := newTestNameGeneratorEditorDockable(g)
	list := trainingNamesPanel(d, g)
	c.NotNil(list)
	c.Equal(2, len(list.rows.Children()))

	list.addOption()
	c.Equal(3, len(d.model.Entries))
	added := d.model.Entries[2]
	c.Equal(1, added.Weight)
	stringFieldFor(t, d, added.KeyPrefix+"value").SetText("Carol")
	c.Equal([]string{"Alice", "Bob", "Carol"}, optionValues(d.model.Entries))
	integerFieldFor(t, d, added.KeyPrefix+"weight").SetText("3")
	c.Equal(3, added.Weight)

	list = trainingNamesPanel(d, g)
	dd := dragDataForRow(t, list.rows.Children()[2])
	c.Equal("Training Name Drag", dd.title)
	beginDragOver(&d.rowDragState, list.rows, 0)
	d.dataDragDrop(geom.Point{}, dd)
	c.Equal([]string{"Carol", "Alice", "Bob"}, optionValues(d.model.Entries), "the drop reorders the names")
	c.False(d.inDragOver, "the drag state is cleared")

	trainingNamesPanel(d, g).removeOption(d.model.Entries[1])
	c.Equal([]string{"Carol", "Bob"}, optionValues(d.model.Entries), "removal is by identity")

	d.undoMgr.Undo()
	c.Equal([]string{"Carol", "Alice", "Bob"}, optionValues(d.model.Entries), "undo restores the removed name")
	d.undoMgr.Undo()
	c.Equal([]string{"Alice", "Bob", "Carol"}, optionValues(d.model.Entries), "undo restores the order")
}

// TestNameGeneratorPanelCompoundReorder verifies that the rows of a compound generator can be reordered by dragging.
func TestNameGeneratorPanelCompoundReorder(t *testing.T) {
	c := check.New(t)
	g := gurps.NewNameGenerator()
	g.Type = namegen.Compound
	g.Compound = []*gurps.NameGenerator{simpleNameGenerator("Alice"), simpleNameGenerator("Bob")}
	d := newTestNameGeneratorEditorDockable(g)
	rows := rootNameGeneratorPanel(t, d).rows
	dd := dragDataForRow(t, rows.Children()[0])
	c.Equal("Generator Drag", dd.title)
	beginDragOver(&d.rowDragState, rows, 2)
	d.dataDragDrop(geom.Point{}, dd)
	c.Equal("Bob", d.model.Compound[0].Entries[0].Value, "the drop reorders the generators")
	c.Equal("Alice", d.model.Compound[1].Entries[0].Value)
	d.undoMgr.Undo()
	c.Equal("Alice", d.model.Compound[0].Entries[0].Value, "undo restores the order")
	c.False(d.undoMgr.CanUndo(), "the drop is a single edit")
}

// TestParseTrainingNames verifies the text format imports accept: one name per line, a leading byte order mark, blank
// lines and surrounding whitespace ignored, an optional trailing ": N" giving the weight, and anything else kept whole
// as the name.
func TestParseTrainingNames(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name     string
		text     string
		expected []*gurps.WeightedStringOption
	}{
		{name: "empty", text: "", expected: nil},
		{name: "only blank lines", text: "\n  \n\t\n", expected: nil},
		{name: "plain names", text: "Alice\nBob\n", expected: []*gurps.WeightedStringOption{
			{Weight: 1, Value: "Alice"}, {Weight: 1, Value: "Bob"},
		}},
		{name: "no trailing newline", text: "Alice", expected: []*gurps.WeightedStringOption{{Weight: 1, Value: "Alice"}}},
		{name: "blank lines and whitespace", text: "\n  Alice  \n\n\tBob\r\n", expected: []*gurps.WeightedStringOption{
			{Weight: 1, Value: "Alice"}, {Weight: 1, Value: "Bob"},
		}},
		{name: "byte order mark", text: "\ufeffAlice: 3\nBob\n", expected: []*gurps.WeightedStringOption{
			{Weight: 3, Value: "Alice"}, {Weight: 1, Value: "Bob"},
		}},
		{name: "weights", text: "Alice: 3\nBob:2\nCarol :  10", expected: []*gurps.WeightedStringOption{
			{Weight: 3, Value: "Alice"}, {Weight: 2, Value: "Bob"}, {Weight: 10, Value: "Carol"},
		}},
		{
			name: "malformed weight keeps the whole line", text: "Alice: three\nBob: 2x\nCarol: 0\nDave: -1\nEve:",
			expected: []*gurps.WeightedStringOption{
				{Weight: 1, Value: "Alice: three"},
				{Weight: 1, Value: "Bob: 2x"},
				{Weight: 1, Value: "Carol: 0"},
				{Weight: 1, Value: "Dave: -1"},
				{Weight: 1, Value: "Eve:"},
			},
		},
		{name: "colon within the name", text: "Ka:ren: 2\nRe:becca", expected: []*gurps.WeightedStringOption{
			{Weight: 2, Value: "Ka:ren"}, {Weight: 1, Value: "Re:becca"},
		}},
		{name: "weight with no name", text: ": 3", expected: []*gurps.WeightedStringOption{{Weight: 1, Value: ": 3"}}},
	} {
		c.Equal(one.expected, parseTrainingNames(one.text), one.name)
	}
}

// TestNameGeneratorPanelImportDedupes verifies that imported names are added after the existing ones in a single undo
// edit, that a name already present or repeated within the import is added once, and that an import with nothing new
// posts nothing.
func TestNameGeneratorPanelImportDedupes(t *testing.T) {
	c := check.New(t)
	g := simpleNameGenerator("Alice")
	d := newTestNameGeneratorEditorDockable(g)
	root := rootNameGeneratorPanel(t, d)
	root.addTrainingNames(parseTrainingNames("Bob: 2\nAlice\nCarol\nBob\n"))
	c.Equal([]string{"Alice", "Bob", "Carol"}, optionValues(d.model.Entries))
	c.Equal(2, d.model.Entries[1].Weight, "the first occurrence's weight is kept")
	for _, one := range d.model.Entries {
		c.NotEqual("", one.KeyPrefix, "every entry has a key prefix")
	}
	c.Equal(3, len(trainingNamesPanel(d, g).rows.Children()), "the rows are rebuilt")
	c.True(d.Modified())

	rootNameGeneratorPanel(t, d).addTrainingNames(parseTrainingNames("Alice\nCarol"))
	c.Equal([]string{"Alice", "Bob", "Carol"}, optionValues(d.model.Entries), "nothing new is added")

	d.undoMgr.Undo()
	c.Equal([]string{"Alice"}, optionValues(d.model.Entries), "the import is a single edit")
	c.False(d.undoMgr.CanUndo(), "and an import with nothing new posted none")
}

// TestNameGeneratorSamplesPanel verifies that the samples panel shows ten names from a working definition, the reason
// when the definition cannot generate names, and that the samples follow edits and can be regenerated on demand.
func TestNameGeneratorSamplesPanel(t *testing.T) {
	c := check.New(t)
	d := newTestNameGeneratorEditorDockable(simpleNameGenerator("Alice"))
	c.NotNil(d.samples)
	c.Equal(strings.Repeat("Alice, ", 9)+"Alice", d.samples.label.text, "ten names, joined by commas")
	c.Equal(unison.DefaultLabelTheme.OnBackgroundInk, d.samples.label.ink, "names are drawn normally")

	trainingNamesPanel(d, d.model).removeOption(d.model.Entries[0])
	c.Contains(d.samples.label.text, "no training data has been provided", "the rebuilt samples say what is missing")
	c.Equal(unison.ThemeError, d.samples.label.ink, "the reason is drawn as an error")

	d.undoMgr.Undo()
	c.Equal(strings.Repeat("Alice, ", 9)+"Alice", d.samples.label.text, "undo brings the names back")

	// Typing into a field is not a structural edit, yet the samples must follow it, which they do through Sync.
	stringFieldFor(t, d, d.model.Entries[0].KeyPrefix+"value").SetText("Bob")
	c.Equal(strings.Repeat("Bob, ", 9)+"Bob", d.samples.label.text, "the samples follow a typed change")

	d.model.Entries[0].Value = "Carol"
	button := buttonWithSVG(d.samples.AsPanel(), svg.Randomize)
	c.NotNil(button, "the header has a regenerate button")
	button.ClickCallback()
	c.Equal(strings.Repeat("Carol, ", 9)+"Carol", d.samples.label.text, "the button regenerates the samples")

	compound := gurps.NewNameGenerator()
	compound.Type = namegen.Compound
	d = newTestNameGeneratorEditorDockable(compound)
	c.Contains(d.samples.label.text, "at least one name generator", "an empty compound generator explains itself")
	c.NotContains(d.samples.label.text, "\n", "the message carries no stack trace")
}
