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
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/promptstep"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
)

// The prompts a transfer puts to the user are each shown on their own, with nothing around them to say what they are
// part of. So each one is titled with the operation's short name and the step it is, such as "Apply Template:
// Modifiers", is given a line describing the operation under way, such as "Applying template Knight to Sir Bob", and
// places the rows it asks about by their kind and the containers above them.

const (
	// maxNameLength is how much of a name goes into a description or a list of names. Names are cut down before the
	// text holding them is put together, so that a long one can't push the rest of that text out of sight.
	maxNameLength = 40
	// maxLocationLength is how long the containers above a row may run before the ones in the middle are left out.
	maxLocationLength = 60
)

// promptOperation describes the operation a prompt is part of. Any of its fields may be left at its zero value.
type promptOperation struct {
	// name is the operation's short name, such as "Apply Template", which starts the prompt's title.
	name string
	// description says what the operation is doing, such as "Applying template Knight to Sir Bob", and is shown above
	// the prompt's question (see newOperationLabel).
	description string
	// step identifies the prompt within the operation, and ends the prompt's title.
	step promptstep.Step
}

// at returns the operation with its step set to the given one.
func (op promptOperation) at(step promptstep.Step) promptOperation {
	op.step = step
	return op
}

// title returns the title for a prompt of the operation: its name and step, or whichever of them it has.
func (op promptOperation) title() string {
	switch {
	case op.step == promptstep.None:
		return op.name
	case op.name == "":
		return op.step.String()
	default:
		return fmt.Sprintf(i18n.Text("%s: %s"), op.name, op.step.String())
	}
}

// shortNames returns the names cut down to maxNameLength, ready to be handed to fmt.Sprintf.
func shortNames(names ...string) []any {
	short := make([]any, len(names))
	for i, name := range names {
		short[i] = xstrings.Truncate(name, maxNameLength, true)
	}
	return short
}

// describeRows names the rows a transfer is moving: the row itself when there is only the one, or else how many there
// are, counted by their kind.
func describeRows[T gurps.Node[T]](rows []T) string {
	if len(rows) == 1 {
		return rows[0].String()
	}
	var format string
	switch any(rows).(type) {
	case []*gurps.Trait:
		format = i18n.Text("%d traits")
	case []*gurps.Skill:
		format = i18n.Text("%d skills")
	case []*gurps.Spell:
		format = i18n.Text("%d spells")
	case []*gurps.Equipment:
		format = i18n.Text("%d pieces of equipment")
	case []*gurps.Note:
		format = i18n.Text("%d notes")
	case []*gurps.TraitModifier, []*gurps.EquipmentModifier:
		format = i18n.Text("%d modifiers")
	default:
		format = i18n.Text("%d rows")
	}
	return fmt.Sprintf(format, len(rows))
}

// dockableTitle returns the title of the dockable holding the panel, or an empty string if there isn't one.
func dockableTitle(panel unison.Paneler) string {
	if xreflect.IsNil(panel) {
		return ""
	}
	if d := unison.AncestorOrSelf[unison.Dockable](panel); !xreflect.IsNil(d) {
		return d.Title()
	}
	return ""
}

// rowLocation describes where a row being asked about sits: its kind, followed by the containers above it, outermost
// first. A row at the top level has nothing worth saying about where it is, so an empty string is returned for it.
// When the containers run long, the ones in the middle are left out, keeping the outermost, which says where the rows
// came from, and the innermost, which is the one the row is actually in.
func rowLocation[T gurps.Node[T]](row T) string {
	var path []string
	for parent := row.Parent(); !xreflect.IsNil(parent); parent = parent.Parent() {
		path = append(path, parent.String())
	}
	if len(path) == 0 {
		return ""
	}
	slices.Reverse(path)
	separator := i18n.Text(" › ")
	joined := strings.Join(path, separator)
	if len(path) > 2 && len([]rune(joined)) > maxLocationLength {
		joined = strings.Join([]string{path[0], "…", path[len(path)-1]}, separator)
	}
	return fmt.Sprintf(i18n.Text("%s in %s"), row.Kind(), joined)
}

// nameList returns the names one per line, each cut down to maxNameLength, and cut short with a count of the rest when
// there are too many to show.
func nameList(names []string) string {
	const maxShown = 10
	lines := make([]string, 0, min(len(names), maxShown))
	for i, name := range names {
		if i == maxShown-1 && len(names) > maxShown {
			lines = append(lines, fmt.Sprintf(i18n.Text("and %d more"), len(names)-i))
			break
		}
		lines = append(lines, xstrings.Truncate(name, maxNameLength, true))
	}
	return strings.Join(lines, "\n")
}

// joinNames returns the names on one line, each cut down to maxNameLength.
func joinNames(names []string) string {
	short := make([]string, len(names))
	for i, name := range names {
		short[i] = xstrings.Truncate(name, maxNameLength, true)
	}
	return strings.Join(short, i18n.Text(", "))
}

// newPromptDialog returns a dialog for a prompt of the operation, titled for it (see promptOperation.title). A nil icon
// leaves the dialog without one, which is what a prompt that is simply asking a question wants, since an icon would
// only take room from it.
func newPromptDialog(op promptOperation, icon unison.Drawable, iconInk unison.Ink, content unison.Paneler, buttons ...*unison.DialogButtonInfo) (*unison.Dialog, error) {
	dialog, err := unison.NewDialog(icon, iconInk, content, buttons)
	if err != nil {
		return nil, err
	}
	dialog.Window().SetTitle(op.title())
	return dialog, nil
}

// runPromptDialog shows the content in a modal dialog without an icon (see newPromptDialog) and returns the response,
// or unison.ModalResponseCancel if the dialog couldn't be made.
func runPromptDialog(op promptOperation, content unison.Paneler, buttons ...*unison.DialogButtonInfo) int {
	dialog, err := newPromptDialog(op, nil, nil, content, buttons...)
	if err != nil {
		errs.Log(err)
		return unison.ModalResponseCancel
	}
	return dialog.RunModal()
}

// newTruncatedLabel returns a label showing text, cut down to maxLen characters, with the whole of it in a tooltip when
// it had to be cut.
func newTruncatedLabel(text string, maxLen int, font unison.Font) *unison.Label {
	label := unison.NewLabel()
	label.Font = font
	title := xstrings.Truncate(text, maxLen, true)
	label.SetTitle(title)
	if title != text {
		label.Tooltip = newWrappedTooltip(text)
	}
	return label
}

// newOperationLabel returns the label a prompt shows above its question to describe the operation it is part of, or
// nil if the operation has no description.
func newOperationLabel(op promptOperation) *unison.Label {
	if op.description == "" {
		return nil
	}
	return newTruncatedLabel(op.description, 80, fonts.FieldSecondary)
}

// newOperationMessagePanel returns a message panel with the primary and detail text, beneath the operation's
// description when it has one (see newOperationLabel).
func newOperationMessagePanel(op promptOperation, primary, detail string) *unison.Panel {
	message := unison.NewMessagePanel(primary, detail)
	opLabel := newOperationLabel(op)
	if opLabel == nil {
		return message
	}
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
	})
	panel.AddChild(opLabel)
	panel.AddChild(message)
	return panel
}
