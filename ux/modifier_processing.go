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

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/promptstep"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// The modifier prompts are held in variables so that tests can substitute non-interactive implementations. Each reports
// whether any modifier was changed and whether the prompt was canceled.
var (
	promptForTraitModifiers     = showModifiersDialog[*gurps.TraitModifier]
	promptForEquipmentModifiers = showModifiersDialog[*gurps.EquipmentModifier]
)

// modifierPromptInfo is what the modifier prompt shows about the row whose modifiers it asks about.
type modifierPromptInfo struct {
	// op is the operation the prompt is part of (see promptOperation).
	op promptOperation
	// name is the row's name.
	name string
	// location is the row's kind and the containers above it (see rowLocation). It is empty for a top-level row.
	location string
	// step is the prompt's place among the rows being asked about, counting from 1, and steps is how many rows there
	// are. A steps of 1 or less shows no count.
	step, steps int
	// requirePicks is true when the rows are headed for a character or loot sheet, where a mandatory modifier choice
	// must have its pick made; elsewhere one may be left without.
	requirePicks bool
}

// modifierTargets returns the rows the modifier prompt asks about: each of the given rows, and every row below them,
// that has modifiers to ask about (see modifierPromptOf).
func modifierTargets[T gurps.Node[T]](rows []T) []T {
	var targets []T
	for _, row := range rows {
		gurps.Traverse(func(row T) bool {
			if modifierPromptOf(row) != nil {
				targets = append(targets, row)
			}
			return false
		}, false, false, row)
	}
	return targets
}

// promptForModifierTargets puts up the modifier prompt for each of the targets (see modifierTargets). The prompts are
// counted as following the given number already done, out of total, since one transfer may ask about the rows of
// several lists. Returns false if the user canceled a prompt, in which case no further prompts are shown.
func promptForModifierTargets[T gurps.Node[T]](op promptOperation, targets []T, done, total int) bool {
	for i, row := range targets {
		if ask := modifierPromptOf(row); ask != nil && ask(&modifierPromptInfo{
			op:           op,
			name:         row.String(),
			location:     rowLocation(row),
			step:         done + i + 1,
			steps:        total,
			requirePicks: gurps.IsOnSheet(row),
		}) {
			return false
		}
	}
	return true
}

// modifierPromptOf returns what puts up the prompt for the row's modifiers and reports whether it was canceled, or nil
// if the row has none to ask about (see modifiersToAskAbout).
func modifierPromptOf[T gurps.Node[T]](row T) func(info *modifierPromptInfo) bool {
	switch t := any(row).(type) {
	case *gurps.Trait:
		return modifierPromptFor(row, t.Modifiers, promptForTraitModifiers)
	case *gurps.Equipment:
		return modifierPromptFor(row, t.Modifiers, promptForEquipmentModifiers)
	default:
		return nil
	}
}

func modifierPromptFor[T gurps.Node[T], M gurps.Node[M]](row T, modifiers []M, prompt func(*modifierPromptInfo, []M) (changed, canceled bool)) func(*modifierPromptInfo) bool {
	mods := modifiersToAskAbout(row, modifiers)
	if len(mods) == 0 {
		return nil
	}
	return func(info *modifierPromptInfo) bool {
		_, canceled := prompt(info, mods)
		return canceled
	}
}

// modifiersToAskAbout returns the modifiers of the row to ask about: all of them for a row that isn't preconfigured,
// and for one on a sheet that is, just the outermost of the mandatory choices it has left unresolved.
func modifiersToAskAbout[T gurps.Node[T], M gurps.Node[M]](row T, modifiers []M) []M {
	if !gurps.IsNodePreconfigured(row) {
		return modifiers
	}
	if !gurps.IsOnSheet(row) {
		return nil
	}
	// A choice within another is shown along with it.
	unresolved := gurps.UnresolvedModifierChoices(modifiers...)
	return slices.DeleteFunc(slices.Clone(unresolved), func(one M) bool {
		for parent := one.Parent(); !xreflect.IsNil(parent); parent = parent.Parent() {
			if slices.Contains(unresolved, parent) {
				return true
			}
		}
		return false
	})
}

// showModifiersDialog asks which of the modifiers to enable. Modifiers that aren't options of a choice each get a check
// box. The options of a choice get radio buttons, since no more than one of them may be enabled: an optional choice
// also offers "None". When info.requirePicks is true, a mandatory choice offers no way out of picking and is flagged
// until one is picked, the prompt refusing to finish until then; otherwise it is offered just as an optional one is.
func showModifiersDialog[T gurps.Node[T]](info *modifierPromptInfo, modifiers []T) (changed, canceled bool) {
	selection := newModifierSelection(modifiers, info.requirePicks)
	if selection == nil {
		return false, false
	}
	header := i18n.Text("Select Modifiers for:")
	if info.steps > 1 {
		header = fmt.Sprintf(i18n.Text("Select Modifiers (%d of %d) for:"), info.step, info.steps)
	}
	extraHeaders := []*unison.Label{newTruncatedLabel(info.name, maxRowNameLength, unison.SystemFont)}
	if info.location != "" {
		extraHeaders = append(extraHeaders, newTruncatedLabel(info.location, maxContextLineLength, fonts.FieldSecondary))
	}
	op := info.op.at(promptstep.Modifiers)
	dialog, err := newPromptDialog(op, nil, nil, newListQuestionPanel(op, header, selection.list, extraHeaders...),
		unison.NewCancelButtonInfo(), unison.NewOKButtonInfo())
	if err != nil {
		errs.Log(err)
		return false, true
	}
	selection.onChange = func() { dialog.Button(unison.ModalResponseOK).SetEnabled(selection.complete()) }
	selection.onChange()
	if dialog.RunModal() != unison.ModalResponseOK {
		return false, true
	}
	return selection.apply(), false
}

// modifierSelection is the content of the prompt asking which modifiers to enable, along with what is needed to read
// the answers back out of it.
type modifierSelection struct {
	list     *unison.Panel
	boxes    map[*unison.CheckBox]gurps.GeneralModifier
	choices  []*choiceRadioGroup
	onChange func()
}

// choiceRadioGroup holds the radio buttons of the options of one modifier choice.
type choiceRadioGroup struct {
	group   *unison.Group
	options map[*unison.RadioButton]gurps.GeneralModifier
	// mandatory is true for a choice whose pick has to be made before the prompt can finish.
	mandatory bool
	// status and updateStatus are the flag a mandatory choice shows and what refreshes it; they are nil for any other.
	status       *unison.Label
	updateStatus func(made bool)
}

// made returns true if the choice has been made, which an optional choice always has, "None" being an answer.
func (c *choiceRadioGroup) made() bool {
	return !c.mandatory || c.hasPickedOption()
}

// newModifierSelection builds the content of the prompt asking which of the modifiers to enable, or returns nil if
// there is nothing to ask about. A mandatory choice must have its pick made only when requirePicks is true. Each option
// of a choice starts out picked if it is enabled, the first one enabled winning should an older version have left more
// than one so.
func newModifierSelection[T gurps.Node[T]](modifiers []T, requirePicks bool) *modifierSelection {
	s := &modifierSelection{
		list:  unison.NewPanel(),
		boxes: make(map[*unison.CheckBox]gurps.GeneralModifier),
	}
	s.list.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing)))
	s.list.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: 2 * unison.StdVSpacing,
	})
	changed := func() {
		for _, choice := range s.choices {
			if choice.updateStatus != nil {
				choice.updateStatus(choice.made())
			}
		}
		if s.onChange != nil {
			s.onChange()
		}
	}
	choices := make(map[T]*choiceRadioGroup)
	gurps.Traverse(func(m T) bool {
		gm, ok := any(m).(gurps.GeneralModifier)
		if !ok {
			return false
		}
		text := gm.FullDescription()
		if cost := gm.FullCostDescription(); cost != "" {
			text += "; **" + cost + "**"
		}
		if m.Container() {
			if !gurps.IsModifierChoice(m) {
				s.addRow(gm.Depth(), text, nil, nil)
				return false
			}
			choice := &choiceRadioGroup{
				group:   unison.NewGroup(),
				options: make(map[*unison.RadioButton]gurps.GeneralModifier),
				// A choice with no options has nothing to pick from, so it can't hold the prompt open, and isn't
				// flagged.
				mandatory: requirePicks && gurps.IsMandatoryModifierChoice(m) &&
					len(gurps.ModifierChoiceOptions(m)) != 0,
			}
			choices[m] = choice
			s.choices = append(s.choices, choice)
			text += "; *" + gurps.ModifierChoiceDescription(m) + "*"
			if choice.mandatory {
				choice.status, choice.updateStatus = newModifierChoiceStatus()
				s.addRow(gm.Depth(), text, nil, choice.status)
				return false
			}
			s.addRow(gm.Depth(), text, nil, nil)
			none := unison.NewRadioButton()
			none.Accessibility.Name = i18n.Text("None")
			none.ClickCallback = changed
			choice.group.Add(none)
			choice.group.Select(none)
			s.addRow(gm.Depth()+1, "*"+i18n.Text("None")+"*", none, nil)
			return false
		}
		if owner, found := gurps.ModifierChoiceFor(m); found && choices[owner] != nil {
			choice := choices[owner]
			rb := unison.NewRadioButton()
			rb.Accessibility.Name = gm.NameWithReplacements()
			rb.ClickCallback = changed
			choice.group.Add(rb)
			choice.options[rb] = gm
			if gm.Enabled() && !choice.hasPickedOption() {
				choice.group.Select(rb)
			}
			s.addRow(gm.Depth(), text, rb, nil)
			return false
		}
		cb := unison.NewCheckBox()
		cb.Accessibility.Name = gm.NameWithReplacements()
		cb.State = check.FromBool(gm.Enabled())
		s.boxes[cb] = gm
		s.addRow(gm.Depth(), text, cb, nil)
		return false
	}, false, false, modifiers...)
	children := s.list.Children()
	if len(children) == 0 {
		return nil
	}
	if border, ok := children[len(children)-1].Border().(*unison.EmptyBorder); ok {
		insets := border.Insets()
		insets.Bottom = 0
		children[len(children)-1].SetBorder(unison.NewEmptyBorder(insets))
	}
	changed()
	return s
}

// hasPickedOption returns true if one of the options, rather than "None", is picked.
func (c *choiceRadioGroup) hasPickedOption() bool {
	for rb := range c.options {
		if c.group.Selected(rb) {
			return true
		}
	}
	return false
}

// addRow adds a row to the list, indented for its depth: the control, if there is one, the text, and the status, if
// there is one. Clicking the text clicks the control, as clicking a control's own title would.
func (s *modifierSelection) addRow(depth int, text string, control interface {
	unison.Paneler
	Click()
}, status *unison.Label,
) {
	indent := float32(depth) * xmath.Ceil(unison.DefaultMarkdownTheme.Font.Baseline()*1.5)
	wrapper := unison.NewPanel()
	md := unison.NewMarkdown(false)
	md.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Start,
		HGrab:  true,
	})
	md.SetContent(text, 800-indent)
	md.MouseDownCallback = func(_ geom.Point, _, _ int, _ mod.Modifiers) bool {
		return true
	}
	md.UpdateCursorCallback = func(_ geom.Point) *unison.Cursor {
		return unison.PointingCursor()
	}
	if control != nil {
		panel := control.AsPanel()
		panel.SetLayoutData(&unison.FlexLayoutData{
			VAlign: align.Start,
		})
		panel.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 2}))
		switch c := panel.Self.(type) {
		case *unison.CheckBox:
			c.Font = unison.DefaultMarkdownTheme.Font
		case *unison.RadioButton:
			c.Font = unison.DefaultMarkdownTheme.Font
		}
		md.MouseUpCallback = func(where geom.Point, _ int, _ mod.Modifiers) bool {
			if where.In(md.ContentRect(false)) {
				control.Click()
			}
			return true
		}
		wrapper.AddChild(control)
	}
	wrapper.AddChild(md)
	if status != nil {
		status.SetLayoutData(&unison.FlexLayoutData{
			HAlign: align.End,
			VAlign: align.Start,
		})
		wrapper.AddChild(status)
	}
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  len(wrapper.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Start,
	})
	wrapper.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: indent}))
	s.list.AddChild(wrapper)
}

// newModifierChoiceStatus returns the flag a mandatory choice shows, marking it as required until it is made, along
// with the function that refreshes it.
func newModifierChoiceStatus() (status *unison.Label, update func(made bool)) {
	status = unison.NewLabel()
	status.Font = fonts.FieldSecondary
	status.SetBorder(unison.NewEmptyBorder(geom.NewHorizontalInsets(unison.StdHSpacing)))
	background := pickerMatchStateColor(false)
	status.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		r := status.ContentRect(true)
		gc.DrawRoundedRect(r, geom.NewUniformSize(8), background.Paint(gc, r, paintstyle.Fill))
		status.DefaultDraw(gc, r)
	}
	update = func(made bool) {
		img := svg.Not
		title := i18n.Text("Required")
		if made {
			img = unison.CheckmarkSVG
			title = i18n.Text("Picked")
		}
		size := max(status.Font.Baseline()-2, 6)
		status.Drawable = &unison.DrawableSVG{
			SVG:  img,
			Size: geom.NewSize(size, size),
		}
		status.SetTitle(title)
		background = pickerMatchStateColor(made)
		status.OnBackgroundInk = background.On()
		status.MarkForLayoutRecursivelyUpward()
		status.MarkForRedraw()
	}
	return status, update
}

// complete returns true if every mandatory choice has been made.
func (s *modifierSelection) complete() bool {
	for _, choice := range s.choices {
		if !choice.made() {
			return false
		}
	}
	return true
}

// apply enables the modifiers the answers pick and disables the rest, reporting whether any of them changed.
func (s *modifierSelection) apply() (changed bool) {
	set := func(gm gurps.GeneralModifier, on bool) {
		if gm.Enabled() != on {
			gm.SetEnabled(on)
			changed = true
		}
	}
	for cb, gm := range s.boxes {
		set(gm, cb.State == check.On)
	}
	for _, choice := range s.choices {
		for rb, gm := range choice.options {
			set(gm, choice.group.Selected(rb))
		}
	}
	return changed
}
