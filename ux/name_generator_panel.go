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
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/namegen"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
)

// nameGeneratorPanel edits one name generator: its type, its case handling, and then whatever the type calls for. A
// compound generator lists the generators it combines, each edited by a nested nameGeneratorPanel; the others take
// their training data from a built-in set or from a list of names. The fields shown depend on the type and the
// built-in choice, so changing either rebuilds the editor's content.
type nameGeneratorPanel struct {
	unison.Panel
	dockable  *nameGeneratorEditorDockable
	generator *gurps.NameGenerator
	// parent is the compound generator this one belongs to, or nil for the generator the file defines.
	parent *gurps.NameGenerator
	rows   *unison.Panel // the compound generator's rows; nil unless there are any
}

func newNameGeneratorPanel(d *nameGeneratorEditorDockable, g, parent *gurps.NameGenerator) *nameGeneratorPanel {
	p := &nameGeneratorPanel{
		dockable:  d,
		generator: g,
		parent:    parent,
	}
	p.Self = p
	if parent == nil {
		p.SetBorder(unison.NewEmptyBorder(geom.Insets{
			Top:    unison.StdVSpacing,
			Left:   unison.StdHSpacing,
			Bottom: unison.StdVSpacing,
			Right:  unison.StdHSpacing * 2,
		}))
	}
	p.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	p.addTypeField()
	p.addCaseFields()
	switch g.Type {
	case namegen.Compound:
		p.addCompoundFields()
	case namegen.MarkovLetter:
		p.addDepthField()
		p.addTrainingDataFields()
	default:
		p.addTrainingDataFields()
	}
	return p
}

// addTypeField adds the popup that chooses how names are generated. The fields below it depend on the choice, so
// changing it rebuilds the content.
func (p *nameGeneratorPanel) addTypeField() {
	d := p.dockable
	g := p.generator
	text := i18n.Text("Type")
	p.AddChild(NewFieldLeadingLabel(text, false))
	popup := NewPopup(d.targetMgr, g.KeyPrefix+"type", text,
		func() namegen.Type { return g.Type },
		func(t namegen.Type) {
			g.Type = t
			d.sync()
		}, namegen.Types...)
	popup.Tooltip = newWrappedTooltip(i18n.Text("Simple picks one of the training names at random. Markov Letter builds new names letter by letter from patterns in the training names. Markov Run builds new names from the runs of vowels and consonants in the training names. Compound joins the output of several generators with a separator."))
	p.AddChild(popup)
}

// addCaseFields adds the two case-handling checkboxes. The model records the options as negations, since the files do,
// while the checkboxes are phrased positively, so the two are inverted on the way through.
func (p *nameGeneratorPanel) addCaseFields() {
	d := p.dockable
	g := p.generator
	lowered := NewCheckBox(d.targetMgr, g.KeyPrefix+"lowered", i18n.Text("Lowercase training names before use"),
		func() check.Enum { return check.FromBool(!g.NoLowered) },
		func(state check.Enum) { g.NoLowered = state != check.On })
	lowered.Tooltip = newWrappedTooltip(i18n.Text("The training names are converted to lowercase before being used, so the case they were entered in does not matter"))
	lowered.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	p.AddChild(lowered)

	firstUpper := NewCheckBox(d.targetMgr, g.KeyPrefix+"first_upper",
		i18n.Text("Capitalize the first letter of each generated name"),
		func() check.Enum { return check.FromBool(!g.NoFirstToUpper) },
		func(state check.Enum) { g.NoFirstToUpper = state != check.On })
	firstUpper.Tooltip = newWrappedTooltip(i18n.Text("Each generated name starts with an uppercase letter, whatever case the training names produce"))
	firstUpper.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	p.AddChild(firstUpper)
}

// addCompoundFields adds the separator and the list of generators a compound generator combines.
func (p *nameGeneratorPanel) addCompoundFields() {
	d := p.dockable
	g := p.generator
	text := i18n.Text("Separator")
	p.AddChild(NewFieldLeadingLabel(text, false))
	field := NewStringField(d.targetMgr, g.KeyPrefix+"separator", text,
		func() string { return g.Separator },
		func(s string) { g.Separator = s })
	field.SetMinimumTextWidthUsing(prototypeMinIDWidth)
	field.Tooltip = newWrappedTooltip(i18n.Text("Placed between the outputs of the combined generators, such as a space"))
	p.AddChild(field)

	addButton := unison.NewSVGButton(unison.CircledAddSVG)
	addButton.Tooltip = newWrappedTooltip(i18n.Text("Add generator"))
	addButton.ClickCallback = p.addGenerator
	p.AddChild(newEditorSectionHeader(i18n.Text("Generators"),
		i18n.Text("The generators whose outputs are joined, in order, to form a name"), addButton))
	if len(g.Compound) != 0 {
		p.rows = unison.NewPanel()
		p.rows.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
		p.rows.SetLayout(&unison.FlexLayout{Columns: 1})
		p.rows.SetLayoutData(&unison.FlexLayoutData{HSpan: 2, HAlign: align.Fill, HGrab: true})
		for _, child := range g.Compound {
			p.rows.AddChild(newCompoundGeneratorPanel(d, g, child))
		}
		p.AddChild(p.rows)
	}
}

// addGenerator appends a simple generator to the compound list and moves the focus to its type popup.
func (p *nameGeneratorPanel) addGenerator() {
	d := p.dockable
	g := p.generator
	child := &gurps.NameGenerator{Type: namegen.Simple, KeyPrefix: d.targetMgr.NextPrefix()}
	d.editStructure(i18n.Text("Add Generator"), func() { g.Compound = append(g.Compound, child) }, child.KeyPrefix+"type")
}

// addDepthField adds the Markov letter depth. Zero is allowed so that a file which leaves the depth out, and so uses
// the default, is not altered by being opened.
func (p *nameGeneratorPanel) addDepthField() {
	d := p.dockable
	g := p.generator
	text := i18n.Text("Depth")
	p.AddChild(NewFieldLeadingLabel(text, false))
	field := NewIntegerField(d.targetMgr, g.KeyPrefix+"depth", text,
		func() int { return g.Depth },
		func(v int) { g.Depth = v }, 0, 5, false, false)
	field.SetBaseTooltip(newWrappedTooltip(i18n.Text("How many preceding letters are considered when choosing the next one, from 1 to 5. Lower values produce more random names; higher values stay closer to the training names. 0 uses the default of 3.")))
	p.AddChild(field)
}

// addTrainingDataFields adds the built-in training data popup and, when no built-in set is chosen, the list of training
// names. The list is only shown when it would be used, so choosing a built-in set rebuilds the content.
func (p *nameGeneratorPanel) addTrainingDataFields() {
	d := p.dockable
	g := p.generator
	text := i18n.Text("Built-in Training Data")
	p.AddChild(NewFieldLeadingLabel(text, false))
	popup := NewPopup(d.targetMgr, g.KeyPrefix+"built_in", text,
		func() namegen.Builtin { return g.BuiltIn },
		func(b namegen.Builtin) {
			g.BuiltIn = b
			d.sync()
		}, namegen.Builtins...)
	popup.Tooltip = newWrappedTooltip(i18n.Text("A built-in set of training names. Choose None to provide your own list below."))
	p.AddChild(popup)
	if g.BuiltIn == namegen.None {
		list := newWeightedStringOptionsPanel(d, p.trainingNamesSpec())
		list.SetLayoutData(&unison.FlexLayoutData{HSpan: 2, HAlign: align.Fill, HGrab: true})
		p.AddChild(list)
	}
}

// trainingNamesSpec describes the training names list for the shared weighted string list panel.
func (p *nameGeneratorPanel) trainingNamesSpec() *weightedStringListSpec {
	return &weightedStringListSpec{
		title:                 i18n.Text("Training Names"),
		tooltip:               i18n.Text("The names the generator learns from. Weights make some names more likely; leave them all at 1 for an unweighted list. A name listed more than once has its weights added together."),
		valueTooltip:          i18n.Text("A name for the generator to learn from"),
		addTooltip:            i18n.Text("Add training name"),
		removeTooltip:         i18n.Text("Remove training name"),
		removeSelectedTooltip: i18n.Text("Remove the selected training names"),
		importTooltip:         i18n.Text("Import training names from a text file, one name per line"),
		addTitle:              i18n.Text("Add Training Name"),
		removeTitle:           i18n.Text("Remove Training Name"),
		removeSelectedTitle:   i18n.Text("Remove Selected Training Names"),
		dragTitle:             i18n.Text("Training Name Drag"),
		importer:              p.importTrainingNames,
		list:                  &p.generator.Entries,
	}
}

// importTrainingNames asks for a text file and adds the names it holds to the training names. Any file may be chosen,
// since name lists come in whatever form they were found in.
func (p *nameGeneratorPanel) importTrainingNames() {
	dialog := unison.NewOpenDialog()
	dialog.SetAllowsMultipleSelection(false)
	dialog.SetResolvesAliases(true)
	dialog.SetCanChooseDirectories(false)
	dialog.SetCanChooseFiles(true)
	global := gurps.GlobalSettings()
	dialog.SetInitialDirectory(global.LastDir(gurps.DefaultLastDirKey))
	if !dialog.RunModal() {
		return
	}
	filePath := dialog.Path()
	global.SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(filePath))
	data, err := os.ReadFile(filePath)
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to import training names"), errs.Wrap(err))
		return
	}
	names := parseTrainingNames(string(data))
	if len(names) > largeTrainingNameCount && !confirmLargeImport(filePath, len(names)) {
		return
	}
	p.addTrainingNames(names)
}

// largeTrainingNameCount is the number of training names beyond which an import asks for confirmation. Every training
// name is a row of fields, and every structural edit rebuilds them all, so a list of many thousands makes the editor
// slow to respond; the built-in training sets are the way to use a list of that size.
const largeTrainingNameCount = 2000

// confirmLargeImport asks whether to go ahead with importing count names from the file.
func confirmLargeImport(filePath string, count int) bool {
	return unison.QuestionDialog(fmt.Sprintf(i18n.Text("Import %d names from %s?"), count, filepath.Base(filePath)),
		xstrings.Wrap("", fmt.Sprintf(i18n.Text("Every training name is a row in the editor, which becomes slow to respond with more than %d of them. For a list this size, consider one of the built-in training sets instead."),
			largeTrainingNameCount), 100)) == unison.ModalResponseOK
}

// addTrainingNames appends the names that are not already in the list, in a single undo edit. A name that appears more
// than once among the new names is added once. Nothing is posted when there is nothing new.
func (p *nameGeneratorPanel) addTrainingNames(names []*gurps.WeightedStringOption) {
	d := p.dockable
	g := p.generator
	present := make(map[string]bool, len(g.Entries)+len(names))
	for _, one := range g.Entries {
		present[one.Value] = true
	}
	var added []*gurps.WeightedStringOption
	for _, one := range names {
		if present[one.Value] {
			continue
		}
		present[one.Value] = true
		one.KeyPrefix = d.targetMgr.NextPrefix()
		added = append(added, one)
	}
	if len(added) == 0 {
		return
	}
	d.editStructure(i18n.Text("Import Training Names"), func() { g.Entries = append(g.Entries, added...) }, "")
}

// parseTrainingNames extracts training names from text, one per line. Blank lines and the whitespace around a name are
// ignored, as is a byte order mark at the start of the text, which TrimSpace does not treat as whitespace. A line may
// end in ": N", N being a positive integer, to give the name that weight; any other line is used whole with a weight of
// 1, so a name that happens to contain a colon is not cut short.
func parseTrainingNames(text string) []*gurps.WeightedStringOption {
	var result []*gurps.WeightedStringOption
	for line := range strings.Lines(strings.TrimPrefix(text, "\ufeff")) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		weight := 1
		if i := strings.LastIndexByte(line, ':'); i != -1 {
			if n, err := strconv.Atoi(strings.TrimSpace(line[i+1:])); err == nil && n > 0 {
				if name := strings.TrimSpace(line[:i]); name != "" {
					line = name
					weight = n
				}
			}
		}
		result = append(result, &gurps.WeightedStringOption{Weight: weight, Value: line})
	}
	return result
}

// compoundGeneratorPanel is one row of a compound generator's list: a drag handle, a remove button and the nested
// editor for the generator itself.
type compoundGeneratorPanel struct {
	unison.Panel
	dockable     *nameGeneratorEditorDockable
	parent       *gurps.NameGenerator
	generator    *gurps.NameGenerator
	deleteButton *unison.Button
}

func newCompoundGeneratorPanel(d *nameGeneratorEditorDockable, parent, child *gurps.NameGenerator) *compoundGeneratorPanel {
	p := &compoundGeneratorPanel{
		dockable:  d,
		parent:    parent,
		generator: child,
	}
	p.Self = p
	configureEditorRow(p.AsPanel(), 3)
	p.AddChild(NewDragHandle(editorRowDragKey, &editorRowDragData{
		editor: d,
		row:    p.AsPanel(),
		title:  i18n.Text("Generator Drag"),
		move:   func(to int) bool { return moveEntry(&parent.Compound, slices.Index(parent.Compound, child), to) },
	}))

	p.deleteButton = unison.NewSVGButton(unison.TrashSVG)
	p.deleteButton.ClickCallback = p.remove
	p.deleteButton.Tooltip = newWrappedTooltip(i18n.Text("Remove generator"))
	p.AddChild(p.deleteButton)

	p.AddChild(newNameGeneratorPanel(d, child, parent))
	return p
}

// remove takes this generator out of the compound list. Removing the last one is allowed, though the samples will then
// say that a compound generator needs at least one.
func (p *compoundGeneratorPanel) remove() {
	i := slices.Index(p.parent.Compound, p.generator)
	if i == -1 {
		return
	}
	p.dockable.editStructure(i18n.Text("Remove Generator"), func() {
		p.parent.Compound = slices.Delete(p.parent.Compound, i, i+1)
	}, "")
}
