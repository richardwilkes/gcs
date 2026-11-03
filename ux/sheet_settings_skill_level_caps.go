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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/role"
)

// skillLevelCapColumns is the number of columns in the skill level cap list: the cap's description and a button to
// remove it.
const skillLevelCapColumns = 2

// confirmSkillLevelCapRemoval asks whether to remove the cap with the description and reports whether the user agreed.
// It is a variable so that tests can stand in for the modal dialog.
var confirmSkillLevelCapRemoval = func(description string) bool {
	return unison.QuestionDialog(i18n.Text("Are you sure you want to remove this skill level cap?"), description) ==
		unison.ModalResponseOK
}

// createSkillLevelCapsSection adds the section that lists the caps on skill and spell levels, each opening its editor
// when clicked, with the buttons that add a cap and import or export caps below the list, where its rebuilds leave
// them be.
func (d *sheetSettingsDockable) createSkillLevelCapsSection(content *unison.Panel) {
	title := i18n.Text("Skill Level Caps")
	section := newSettingsSection(content, title, 1)
	d.skillLevelCapsPanel = unison.NewPanel()
	d.skillLevelCapsPanel.SetLayout(&unison.FlexLayout{
		Columns:  skillLevelCapColumns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	d.skillLevelCapsPanel.SetBorder(newSettingsListBorder())
	d.skillLevelCapsPanel.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	d.skillLevelCapsPanel.Tooltip = newWrappedTooltip(i18n.Text("Each cap holds the skills or spells it selects to at most the level it resolves to on the sheet. Where several caps select the same skill or spell, the lowest of them applies. Click a cap to edit it."))
	installListBanding(d.skillLevelCapsPanel, func() [][]unison.Paneler { return d.skillLevelCapRows })
	section.AddChild(d.skillLevelCapsPanel)

	d.skillLevelCapAddButton = unison.NewSVGButton(unison.CircledAddSVG)
	d.skillLevelCapAddButton.Accessibility.Name = i18n.Text("Add a skill level cap…")
	d.skillLevelCapAddButton.Tooltip = newWrappedTooltip(d.skillLevelCapAddButton.Accessibility.Name)
	d.skillLevelCapAddButton.ClickCallback = d.addSkillLevelCap
	d.skillLevelCapsMenu = &settingsFileMenu{
		title:      title,
		extensions: []string{gurps.SkillLevelCapsExt},
		load:       d.importSkillLevelCaps,
		saver:      func(filePath string) error { return gurps.SaveSkillLevelCaps(d.settings().SkillLevelCaps, filePath) },
		canSave:    func() bool { return len(d.settings().SkillLevelCaps) != 0 },
	}
	d.skillLevelCapsMenuButton = unison.NewSVGButton(svg.Menu)
	// Named apart from the toolbar's menu, which loads and saves the whole of the sheet settings.
	d.skillLevelCapsMenuButton.Accessibility.Name = i18n.Text("Skill Level Caps Menu")
	d.skillLevelCapsMenuButton.Tooltip = newWrappedTooltip(d.skillLevelCapsMenuButton.Accessibility.Name)
	d.skillLevelCapsMenuButton.ClickCallback = func() { d.skillLevelCapsMenu.show(d.skillLevelCapsMenuButton) }
	buttons := unison.NewPanel()
	buttons.SetLayout(&unison.FlexLayout{Columns: 2, HSpacing: unison.StdHSpacing})
	buttons.AddChild(d.skillLevelCapAddButton)
	buttons.AddChild(d.skillLevelCapsMenuButton)
	section.AddChild(buttons)
	d.rebuildSkillLevelCaps()
}

// rebuildSkillLevelCaps fills the skill level cap list from the settings: a row for each cap, in the order held,
// showing its description, which opens its editor, with a button that removes it, or a hint saying how to add one when
// there are none.
func (d *sheetSettingsDockable) rebuildSkillLevelCaps() {
	p := d.skillLevelCapsPanel
	p.RemoveAllChildren()
	d.skillLevelCapRows = nil
	caps := d.settings().SkillLevelCaps
	if len(caps) == 0 {
		hint := NewWrappingLabel()
		hint.Font = fonts.FieldSecondary
		hint.setText(i18n.Text("No skill or spell levels are capped. Use the add button below to add a cap."),
			dimmedTextColor)
		hint.SetLayoutData(&unison.FlexLayoutData{
			HSpan:  skillLevelCapColumns,
			HAlign: align.Fill,
		})
		p.AddChild(hint)
	}
	entity := d.entity()
	for _, one := range caps {
		var sentence *sentenceButton
		sentence = newSentenceButton(one.Describe(entity, emphasize), func() { d.editSkillLevelCap(one, sentence) })
		// The editor opens in a dialog of its own rather than in place, so the sentence is a plain button.
		sentence.Accessibility.Role = role.Button
		addAccessibilityCallback(sentence, func(node *accessibility.Node) {
			node.Expandable = false
			node.Actions = node.Actions.Without(accessibility.Expand, accessibility.Collapse)
		})
		sentence.setInk(unison.ThemeOnBelowSurface)
		sentence.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Middle, HGrab: true})
		p.AddChild(sentence)
		remove := unison.NewSVGButton(unison.TrashSVG)
		remove.Accessibility.Name = i18n.Text("Remove cap: %s", sentence.plainText())
		remove.Tooltip = newWrappedTooltip(remove.Accessibility.Name)
		remove.ClickCallback = func() { d.removeSkillLevelCap(one, remove) }
		remove.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
		p.AddChild(remove)
		d.skillLevelCapRows = append(d.skillLevelCapRows, []unison.Paneler{sentence, remove})
	}
	p.MarkForLayoutRecursivelyUpward()
	p.MarkForRedraw()
}

// addSkillLevelCap puts up the editor for a new cap and adds it to the end of the list when it is accepted.
func (d *sheetSettingsDockable) addSkillLevelCap() {
	one := gurps.NewSkillLevelCap()
	if !showSkillLevelCapEditor(i18n.Text("New Skill Level Cap"), d.entity(), one) {
		return
	}
	s := d.settings()
	s.SkillLevelCaps = slices.Concat(s.SkillLevelCaps, gurps.SkillLevelCaps{one})
	d.rebuildSkillLevelCaps()
	d.syncSheet(gurps.SheetSettingsGeneralPart, false)
}

// editSkillLevelCap puts up the editor for a copy of the cap the sentence shows, so that a cancel leaves the cap as it
// was, and puts the copy in its place when it is accepted with changes. The focus stays on the cap's sentence when it
// was there.
func (d *sheetSettingsDockable) editSkillLevelCap(one *gurps.SkillLevelCap, sentence *sentenceButton) {
	edited := one.Clone()
	if !showSkillLevelCapEditor(i18n.Text("Edit Skill Level Cap"), d.entity(), edited) ||
		gurps.Hash64(edited) == gurps.Hash64(one) {
		return
	}
	s := d.settings()
	i := slices.Index(s.SkillLevelCaps, one)
	if i < 0 {
		return
	}
	caps := slices.Clone(s.SkillLevelCaps)
	caps[i] = edited
	s.SkillLevelCaps = caps
	hadFocus := hasWindowFocus(sentence)
	d.rebuildSkillLevelCaps()
	d.syncSheet(gurps.SheetSettingsGeneralPart, false)
	if hadFocus {
		d.skillLevelCapRows[i][0].AsPanel().RequestFocus()
	}
}

// removeSkillLevelCap removes the cap whose row holds the remove button, once the user has confirmed it, since a cap
// can hold a tree built by hand and this view has no undo. The rebuild takes every button off the page, this one
// included, and the focus doesn't leave a removed panel of its own accord, so it is moved on when it was on the button.
func (d *sheetSettingsDockable) removeSkillLevelCap(one *gurps.SkillLevelCap, remove *unison.Button) {
	s := d.settings()
	i := slices.Index(s.SkillLevelCaps, one)
	if i < 0 {
		return
	}
	hadFocus := hasWindowFocus(remove)
	if !confirmSkillLevelCapRemoval(one.Describe(d.entity(), func(text string) string { return text })) {
		return
	}
	s.SkillLevelCaps = slices.Concat(s.SkillLevelCaps[:i], s.SkillLevelCaps[i+1:])
	d.rebuildSkillLevelCaps()
	d.syncSheet(gurps.SheetSettingsGeneralPart, false)
	if hadFocus {
		d.focusSkillLevelCapRemoveButton(i)
	}
}

// hasWindowFocus reports whether the panel is the keyboard focus of its window, whether or not that window is active.
func hasWindowFocus(p unison.Paneler) bool {
	wnd := p.AsPanel().Window()
	return wnd != nil && wnd.CurrentFocus() == p.AsPanel()
}

// focusSkillLevelCapRemoveButton gives the keyboard focus to the button that removes the cap at index, or the last cap
// when there are fewer, or to the add button when there are none.
func (d *sheetSettingsDockable) focusSkillLevelCapRemoveButton(index int) {
	if len(d.skillLevelCapRows) == 0 {
		d.skillLevelCapAddButton.RequestFocus()
		return
	}
	d.skillLevelCapRows[min(index, len(d.skillLevelCapRows)-1)][1].AsPanel().RequestFocus()
}

// importSkillLevelCaps adds the caps in the file to the end of the list, in the file's order, leaving out each that is
// the same as one already listed or one earlier in the file.
func (d *sheetSettingsDockable) importSkillLevelCaps(ref *library.NamedFileRef) error {
	imported, err := gurps.NewSkillLevelCapsFromFile(ref.FileSystem, ref.FilePath)
	if err != nil {
		return err
	}
	s := d.settings()
	seen := make(map[uint64]bool, len(s.SkillLevelCaps)+len(imported))
	for _, one := range s.SkillLevelCaps {
		seen[gurps.Hash64(one)] = true
	}
	caps := slices.Clone(s.SkillLevelCaps)
	for _, one := range imported {
		if h := gurps.Hash64(one); !seen[h] {
			seen[h] = true
			caps = append(caps, one)
		}
	}
	if len(caps) == len(s.SkillLevelCaps) {
		return nil
	}
	s.SkillLevelCaps = caps
	d.rebuildSkillLevelCaps()
	d.syncSheet(gurps.SheetSettingsGeneralPart, false)
	return nil
}
