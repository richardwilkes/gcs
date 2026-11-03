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
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/levelcap"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/role"
	"github.com/richardwilkes/unison/enums/side"
)

// newTestSkillLevelCap returns a cap holding the skills with the given name to a fixed level.
func newTestSkillLevelCap(name string, level int) *gurps.SkillLevelCap {
	one := gurps.NewSkillLevelCap()
	one.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: name}
	one.Cap.Children = gurps.LevelCaps{gurps.NewFixedLevelCap(one.Cap, fxp.FromInteger(level))}
	return one
}

// setFixedLevelCap makes the cap hold what it selects to a fixed level, as its editor would.
func setFixedLevelCap(levelCap *gurps.SkillLevelCap, level int) {
	levelCap.Cap.Children = gurps.LevelCaps{gurps.NewFixedLevelCap(levelCap.Cap, fxp.FromInteger(level))}
}

// plainDescriptions returns the description of each cap, without emphasis, as its row should show it.
func plainDescriptions(entity *gurps.Entity, caps ...*gurps.SkillLevelCap) []string {
	texts := make([]string, 0, len(caps))
	for _, one := range caps {
		texts = append(texts, one.Describe(entity, func(s string) string { return s }))
	}
	return texts
}

// skillLevelCapHashes returns the hash of each cap, in order.
func skillLevelCapHashes(caps gurps.SkillLevelCaps) []uint64 {
	hashes := make([]uint64, 0, len(caps))
	for _, one := range caps {
		hashes = append(hashes, gurps.Hash64(one))
	}
	return hashes
}

// skillLevelCapSentences returns the text of the sentence of each row of the dockable's skill level cap list, in order.
func skillLevelCapSentences(d *sheetSettingsDockable) []string {
	buttons := uxtest.PanelsOfType[*sentenceButton](d.skillLevelCapsPanel)
	texts := make([]string, 0, len(buttons))
	for _, b := range buttons {
		texts = append(texts, b.plainText())
	}
	return texts
}

// selfAs returns what the panel resolves to as a T, failing the test when it isn't one.
func selfAs[T any](t *testing.T, p unison.Paneler) T {
	t.Helper()
	typed, ok := p.AsPanel().Self.(T)
	if !ok {
		t.Fatalf("expected a %T, got a %T", typed, p.AsPanel().Self)
	}
	return typed
}

// skillLevelCapRemoveNames returns the accessible name of the remove button of each row of the dockable's skill level
// cap list, in order.
func skillLevelCapRemoveNames(d *sheetSettingsDockable) []string {
	buttons := uxtest.PanelsOfType[*unison.Button](d.skillLevelCapsPanel)
	names := make([]string, 0, len(buttons))
	for _, b := range buttons {
		names = append(names, b.Accessibility.Name)
	}
	return names
}

// TestSheetSettingsSkillLevelCapsListFollowsTheSettings checks that the skill level cap list shows a row for each cap
// of the settings, in order, with its description as a plain button and a remove button named after it, or a hint when
// there are none, that sync() and reset() show the list of replaced settings, sync() without telling the sheets
// anything, and that a change reported to the sheet's attribute definitions, which the rows name, describes them anew.
func TestSheetSettingsSkillLevelCapsListFollowsTheSettings(t *testing.T) {
	c := check.New(t)
	owner := newEntityPanelWithFlaggedSettings()
	stealth := newTestSkillLevelCap("Stealth", 12)
	magic := newTestSkillLevelCap("Thaumatology", 14)
	iq := gurps.NewReferenceLevelCap(owner.entity, magic.Cap, levelcap.Attribute)
	iq.Attribute = gurps.IntelligenceID
	iq.Modifier = fxp.Two
	magic.Cap.Children = gurps.LevelCaps{iq}
	owner.entity.SheetSettings.SkillLevelCaps = gurps.SkillLevelCaps{stealth, magic}
	d, recorder := newTestSheetSettingsDockable(t, owner)
	// Docked, as it is when open, where the reports of changes to the settings reach it.
	Workspace.DocumentDock.DockTo(d, nil, side.Left)
	sentences := plainDescriptions(owner.entity, stealth, magic)
	c.Equal(sentences, skillLevelCapSentences(d), "a row for each cap, in order, showing its description")
	c.Equal([]string{"Remove cap: " + sentences[0], "Remove cap: " + sentences[1]}, skillLevelCapRemoveNames(d),
		"each remove button is named after its own cap")
	c.Equal(0, len(uxtest.LabelTexts(d.skillLevelCapsPanel)), "no hint while there are caps")
	for _, b := range uxtest.PanelsOfType[*sentenceButton](d.skillLevelCapsPanel) {
		var node accessibility.Node
		b.Accessibility.Callback(&node)
		c.Equal(role.Button, b.Accessibility.Role, "a sentence opens its editor in a dialog, so it is a plain button")
		c.False(node.Expandable, "rather than a disclosure")
		c.False(node.Actions.Has(accessibility.Expand) || node.Actions.Has(accessibility.Collapse),
			"and offers nothing to expand or collapse")
	}

	replacement := owner.entity.SheetSettings.Clone(owner.entity)
	replacement.SkillLevelCaps = gurps.SkillLevelCaps{replacement.SkillLevelCaps[1]}
	owner.entity.SheetSettings = replacement
	replacement.SetOwningEntity(owner.entity)
	d.sync()
	c.Equal(sentences[1:], skillLevelCapSentences(d), "sync shows the replacement's caps")
	c.Equal(0, len(recorder.updates), "sync tells the sheets nothing")

	// Renamed as the attribute settings rename it.
	defs := owner.entity.SheetSettings.Attributes.Clone()
	defs.Set[gurps.IntelligenceID].Name = "Smarts"
	owner.entity.SheetSettings.Attributes = defs
	notifySheetSettingsUpdated(owner.entity, gurps.SheetSettingsAttributesPart, true)
	renamed := plainDescriptions(owner.entity, owner.entity.SheetSettings.SkillLevelCaps...)
	c.True(strings.HasSuffix(renamed[0], "Smarts +2"), "the cap names the attribute as defined: %q", renamed[0])
	c.Equal(renamed, skillLevelCapSentences(d), "a change to the attribute definitions describes the rows anew")
	c.Equal([]string{"Remove cap: " + renamed[0]}, skillLevelCapRemoveNames(d), "and names their remove buttons anew")

	d.reset()
	c.Equal(0, len(owner.entity.SheetSettings.SkillLevelCaps), "reset restores the defaults, which have no caps")
	c.Equal(0, len(skillLevelCapSentences(d)), "the list follows the reset")
	c.Equal([]string{"No skill or spell levels are capped. Use the add button below to add a cap."},
		uxtest.LabelTexts(d.skillLevelCapsPanel), "an empty list explains how to add to it")
}

// TestSheetSettingsSkillLevelCapChanges checks adding, editing and removing a cap through the list: the editor is put
// up with the title for the change and handed a cap the settings don't hold, so that canceling it changes nothing; a
// removal is asked about first, naming the cap, so that declining it changes nothing; an accepted change is written to
// a new list, leaving the one held before as it was, shown, and reported to the sheets once as a change to a general
// setting without a full rebuild; and nothing is reported when nothing changed.
func TestSheetSettingsSkillLevelCapChanges(t *testing.T) {
	add := func(d *sheetSettingsDockable) { d.skillLevelCapAddButton.ClickCallback() }
	edit := func(row int) func(d *sheetSettingsDockable) {
		return func(d *sheetSettingsDockable) { selfAs[*sentenceButton](t, d.skillLevelCapRows[row][0]).onClick() }
	}
	removeFirst := func(d *sheetSettingsDockable) {
		selfAs[*unison.Button](t, d.skillLevelCapRows[0][1]).ClickCallback()
	}
	cases := []struct {
		name    string
		act     func(d *sheetSettingsDockable)
		title   string                          // the title the editor is put up with; empty when it isn't
		edit    func(*gurps.SkillLevelCap) bool // what the editor does with the cap it is handed
		asks    bool                            // whether removing the first cap is asked about
		confirm bool                            // the answer to that
		want    func() gurps.SkillLevelCaps
		report  bool
	}{
		{
			name:  "adding a cap",
			act:   add,
			title: "New Skill Level Cap",
			edit: func(levelCap *gurps.SkillLevelCap) bool {
				levelCap.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Climbing"}
				setFixedLevelCap(levelCap, 11)
				return true
			},
			want: func() gurps.SkillLevelCaps {
				return gurps.SkillLevelCaps{
					newTestSkillLevelCap("Stealth", 12), newTestSkillLevelCap("Thaumatology", 14),
					newTestSkillLevelCap("Climbing", 11),
				}
			},
			report: true,
		},
		{
			name:  "canceling a new cap",
			act:   add,
			title: "New Skill Level Cap",
			edit: func(levelCap *gurps.SkillLevelCap) bool {
				setFixedLevelCap(levelCap, 11)
				return false
			},
		},
		{
			name:  "editing a cap",
			act:   edit(0),
			title: "Edit Skill Level Cap",
			edit: func(levelCap *gurps.SkillLevelCap) bool {
				setFixedLevelCap(levelCap, 9)
				return true
			},
			want: func() gurps.SkillLevelCaps {
				return gurps.SkillLevelCaps{newTestSkillLevelCap("Stealth", 9), newTestSkillLevelCap("Thaumatology", 14)}
			},
			report: true,
		},
		{
			name:  "accepting a cap's editor without a change",
			act:   edit(1),
			title: "Edit Skill Level Cap",
			edit:  func(*gurps.SkillLevelCap) bool { return true },
		},
		{
			name:  "canceling a cap's edit",
			act:   edit(0),
			title: "Edit Skill Level Cap",
			edit: func(levelCap *gurps.SkillLevelCap) bool {
				setFixedLevelCap(levelCap, 9)
				return false
			},
		},
		{
			name:    "removing a cap",
			act:     removeFirst,
			asks:    true,
			confirm: true,
			want: func() gurps.SkillLevelCaps {
				return gurps.SkillLevelCaps{newTestSkillLevelCap("Thaumatology", 14)}
			},
			report: true,
		},
		{
			name: "declining to remove a cap",
			act:  removeFirst,
			asks: true,
		},
	}
	for _, one := range cases {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			owner := newEntityPanelWithFlaggedSettings()
			s := owner.entity.SheetSettings
			s.SkillLevelCaps = gurps.SkillLevelCaps{
				newTestSkillLevelCap("Stealth", 12), newTestSkillLevelCap("Thaumatology", 14),
			}
			held := s.SkillLevelCaps
			heldHashes := skillLevelCapHashes(held)
			d, recorder := newTestSheetSettingsDockable(t, owner)
			var titles []string
			uxtest.SwapForTest(t, &showSkillLevelCapEditor,
				func(title string, entity *gurps.Entity, levelCap *gurps.SkillLevelCap) bool {
					titles = append(titles, title)
					c.True(entity == owner.entity, "the editor is handed the sheet's entity")
					c.False(slices.Contains(s.SkillLevelCaps, levelCap), "the editor is handed a cap the settings don't hold")
					return one.edit != nil && one.edit(levelCap)
				})
			var asked []string
			uxtest.SwapForTest(t, &confirmSkillLevelCapRemoval, func(description string) bool {
				asked = append(asked, description)
				return one.confirm
			})
			one.act(d)
			if one.title != "" {
				c.Equal([]string{one.title}, titles, "the editor is put up once, with the title for the change")
			} else {
				c.Equal(0, len(titles), "the editor is not put up")
			}
			if one.asks {
				c.Equal(plainDescriptions(owner.entity, held[0]), asked,
					"the removal is asked about once, naming the cap")
			} else {
				c.Equal(0, len(asked), "no removal is asked about")
			}
			want := held
			if one.want != nil {
				want = one.want()
			}
			c.Equal(skillLevelCapHashes(want), skillLevelCapHashes(s.SkillLevelCaps), "the settings hold the caps")
			c.Equal(plainDescriptions(owner.entity, want...), skillLevelCapSentences(d), "the list shows them")
			c.Equal(heldHashes, skillLevelCapHashes(held), "the list held before is left as it was")
			if one.report {
				c.Equal([]gurps.SheetSettingsParts{gurps.SheetSettingsGeneralPart}, recorder.parts,
					"the change is reported once, as one to a general setting")
				c.Equal([]bool{false}, recorder.updates, "without a full rebuild")
			} else {
				c.Equal(0, len(recorder.updates), "nothing changed, so nothing is reported")
			}
		})
	}
}

// TestSheetSettingsSkillLevelCapListKeepsTheFocus checks that the list's rebuild leaves the keyboard focus where the
// user expects rather than on a widget the rebuilt list no longer holds: an accepted edit keeps it on the edited cap's
// sentence, a change to the attribute definitions keeps it on the widget that takes its place, and removing a cap with
// its button, once the question naming the cap is answered, moves it to the button that takes its place, or to the add
// button once no cap is left.
func TestSheetSettingsSkillLevelCapListKeepsTheFocus(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	screen.Do(func() {
		sheet.Entity().SheetSettings.SkillLevelCaps = gurps.SkillLevelCaps{
			newTestSkillLevelCap("Stealth", 12), newTestSkillLevelCap("Thaumatology", 14),
		}
	})
	d := openSheetSettingsFor(t, screen, sheet)
	uxtest.SwapForTest(t, &showSkillLevelCapEditor, func(_ string, _ *gurps.Entity, levelCap *gurps.SkillLevelCap) bool {
		setFixedLevelCap(levelCap, 9)
		return true
	})
	var focused, expected *unison.Panel
	var sentences []string
	screen.Do(func() {
		sentence := selfAs[*sentenceButton](t, d.skillLevelCapRows[0][0])
		sentence.RequestFocus()
		sentence.onClick()
		focused = wnd.CurrentFocus()
		expected = d.skillLevelCapRows[0][0].AsPanel()
		sentences = skillLevelCapSentences(d)
	})
	c.Equal(plainDescriptions(sheet.Entity(), newTestSkillLevelCap("Stealth", 9), newTestSkillLevelCap("Thaumatology", 14)),
		sentences, "the edit is accepted")
	c.True(focused != nil && focused == expected, "the focus must stay on the edited cap's sentence")

	screen.Do(func() {
		d.skillLevelCapRows[1][1].AsPanel().RequestFocus()
		notifySheetSettingsUpdated(sheet.Entity(), gurps.SheetSettingsAttributesPart, true)
		focused = wnd.CurrentFocus()
		expected = d.skillLevelCapRows[1][1].AsPanel()
	})
	c.True(focused != nil && focused == expected,
		"the focus must move to the widget that takes its place when the attribute definitions change")

	remove := func() {
		var button *unison.Button
		var description string
		screen.Do(func() {
			button = selfAs[*unison.Button](t, d.skillLevelCapRows[0][1])
			button.RequestFocus()
			description = skillLevelCapSentences(d)[0]
		})
		// The question runs a nested modal loop, so the click is posted rather than run through Do.
		c.True(screen.Post(button.Click))
		screen.Sync()
		dialogWnd, dialog := modalDialog(t, screen, wnd)
		var texts []string
		screen.Do(func() { texts = uxtest.LabelTexts(dialogWnd.Content()) })
		c.True(slices.Contains(texts, description), "the question names the cap: %q", texts)
		screen.Do(func() { dialog.StopModal(unison.ModalResponseOK) })
		screen.Sync()
		screen.Do(func() {
			focused = wnd.CurrentFocus()
			expected = d.skillLevelCapAddButton.AsPanel()
			if len(d.skillLevelCapRows) != 0 {
				expected = d.skillLevelCapRows[0][1].AsPanel()
			}
			sentences = skillLevelCapSentences(d)
		})
	}
	remove()
	c.Equal(plainDescriptions(sheet.Entity(), newTestSkillLevelCap("Thaumatology", 14)), sentences, "the cap is removed")
	c.True(focused != nil && focused == expected,
		"the focus must move to the button that takes the removed one's place")
	remove()
	c.Equal(0, len(sentences), "the last cap is removed")
	c.True(focused == d.skillLevelCapAddButton.AsPanel(), "the focus must move to the add button once no cap is left")
}

// TestSheetSettingsSkillLevelCapLowersASheetSkill checks, end to end, that a cap added through a sheet's settings
// takes effect on the sheet at once: the skill it selects is held to it, and the skill's level cell says why.
func TestSheetSettingsSkillLevelCapLowersASheetSkill(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var skill *gurps.Skill
	var before fxp.Int
	screen.Do(func() {
		entity := sheet.Entity()
		skill = gurps.NewSkill(entity, nil, false)
		skill.Name = "Stealth"
		entity.Skills = append(entity.Skills, skill)
		entity.Recalculate()
		sheet.Rebuild(true)
		sheet.undoMgr.Clear()
		before = skill.LevelData.Level
	})
	c.True(before > fxp.Five, "the skill must start above the cap")
	d := openSheetSettingsFor(t, screen, sheet)
	uxtest.SwapForTest(t, &showSkillLevelCapEditor, func(_ string, _ *gurps.Entity, levelCap *gurps.SkillLevelCap) bool {
		levelCap.Cap.Children = append(levelCap.Cap.Children, gurps.NewFixedLevelCap(levelCap.Cap, fxp.Five))
		return true
	})
	var level fxp.Int
	var tooltip string
	screen.Do(func() {
		d.skillLevelCapAddButton.Click()
		level = skill.LevelData.Level
		table := sheet.Skills.Table
		if col := table.ColumnIndexForID(gurps.SkillLevelColumn); col >= 0 && table.RowFromIndex(0) != nil {
			tooltip = table.RowFromIndex(0).cellData(col, true).Tooltip
		}
	})
	c.Equal(fxp.Five, level, "the skill is held to the cap")
	c.True(strings.Contains(tooltip, "Level cap of 5"), "the level cell's tooltip says the cap lowered it: %q", tooltip)
}

// TestSheetSettingsSkillLevelCapsImport checks that importing a caps file through the list's menu adds its caps after
// those already listed, in the file's order, leaving out each that is the same as one already listed or one earlier in
// the file, and tells the sheets once; and that an import that adds nothing tells them nothing.
func TestSheetSettingsSkillLevelCapsImport(t *testing.T) {
	c := check.New(t)
	owner := newEntityPanelWithFlaggedSettings()
	stealth := newTestSkillLevelCap("Stealth", 12)
	owner.entity.SheetSettings.SkillLevelCaps = gurps.SkillLevelCaps{stealth}
	d, recorder := newTestSheetSettingsDockable(t, owner)
	path := filepath.Join(t.TempDir(), "Campaign"+gurps.SkillLevelCapsExt)
	c.NoError(gurps.SaveSkillLevelCaps(gurps.SkillLevelCaps{
		newTestSkillLevelCap("Thaumatology", 14), newTestSkillLevelCap("Stealth", 12),
		newTestSkillLevelCap("Thaumatology", 14), newTestSkillLevelCap("Stealth", 11),
	}, path))

	c.NoError(d.skillLevelCapsMenu.load(diskFileRef(path)))
	want := gurps.SkillLevelCaps{stealth, newTestSkillLevelCap("Thaumatology", 14), newTestSkillLevelCap("Stealth", 11)}
	caps := owner.entity.SheetSettings.SkillLevelCaps
	c.Equal(skillLevelCapHashes(want), skillLevelCapHashes(caps), "the file's caps that differ are added after the rest")
	c.True(len(caps) != 0 && caps[0] == stealth, "the cap already listed stays")
	c.Equal(plainDescriptions(owner.entity, want...), skillLevelCapSentences(d), "the list shows them")
	c.Equal([]gurps.SheetSettingsParts{gurps.SheetSettingsGeneralPart}, recorder.parts,
		"the import is reported once, as a change to a general setting")
	c.Equal([]bool{false}, recorder.updates, "without a full rebuild")

	c.NoError(d.skillLevelCapsMenu.load(diskFileRef(path)))
	c.Equal(skillLevelCapHashes(want), skillLevelCapHashes(owner.entity.SheetSettings.SkillLevelCaps),
		"importing the file again adds nothing")
	c.Equal(1, len(recorder.updates), "an import that adds nothing is not reported")
}

// TestSheetSettingsSkillLevelCapsExport checks that the list's menu offers to export only when there are caps, and
// that the file it writes loads back as the same caps, in order.
func TestSheetSettingsSkillLevelCapsExport(t *testing.T) {
	c := check.New(t)
	owner := newEntityPanelWithFlaggedSettings()
	d, _ := newTestSheetSettingsDockable(t, owner)
	menu := d.skillLevelCapsMenu
	canExport := func() bool { return menu.canSave == nil || menu.canSave() }
	c.False(canExport(), "there is nothing to export while there are no caps")

	caps := gurps.SkillLevelCaps{newTestSkillLevelCap("Stealth", 12), newTestSkillLevelCap("Thaumatology", 14)}
	owner.entity.SheetSettings.SkillLevelCaps = caps
	c.True(canExport(), "the caps can be exported")
	dir := t.TempDir()
	name := "Campaign" + gurps.SkillLevelCapsExt
	c.NoError(menu.saver(filepath.Join(dir, name)))
	loaded, err := gurps.NewSkillLevelCapsFromFile(os.DirFS(dir), name)
	c.NoError(err)
	c.Equal(skillLevelCapHashes(caps), skillLevelCapHashes(loaded), "the file holds the caps, in order")
}

// TestSheetSettingsSkillLevelCapsMenuOffersLibraryFiles checks that the list's menu button is named apart from the
// view's toolbar menu, which loads and saves other files, and that the menu offers the caps files in the libraries'
// Settings folders, under their library's title, after its own commands, and that choosing one imports it.
func TestSheetSettingsSkillLevelCapsMenuOffersLibraryFiles(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	user := gurps.GlobalSettings().Libraries.User()
	dir := filepath.Join(user.Path(false), library.SettingsDirName)
	c.NoError(os.MkdirAll(dir, 0o750))
	magic := newTestSkillLevelCap("Thaumatology", 14)
	c.NoError(gurps.SaveSkillLevelCaps(gurps.SkillLevelCaps{magic}, filepath.Join(dir, "Campaign"+gurps.SkillLevelCapsExt)))
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	d := openSheetSettingsFor(t, screen, sheet)
	var toolbarMenu *unison.Button
	screen.Do(func() { toolbarMenu = buttonWithSVG(d.Children()[0], svg.Menu) })
	if toolbarMenu == nil {
		t.Fatal("the view's toolbar must have a menu")
	}
	c.NotNil(screen.AccessibilityTree(wnd))
	var names []string
	for _, b := range []*unison.Button{d.skillLevelCapsMenuButton, toolbarMenu} {
		if node := screen.AccessibilityNodeFor(b); node != nil {
			names = append(names, node.Name)
		}
	}
	c.Equal([]string{"Skill Level Caps Menu", "Menu"}, names,
		"the list's menu button is named apart from the toolbar's")
	screen.Do(func() { d.skillLevelCapsMenuButton.Click() })
	c.Equal([]string{"Import…", "Export…", user.Data().Title, "    Campaign"}, contextMenuTitles(t, screen, wnd),
		"the menu lists the library's caps file after its commands")
	chooseOpenMenuItem(t, screen, wnd, "    Campaign")
	var caps gurps.SkillLevelCaps
	var sentences []string
	screen.Do(func() {
		caps = sheet.Entity().SheetSettings.SkillLevelCaps
		sentences = skillLevelCapSentences(d)
	})
	c.Equal(skillLevelCapHashes(gurps.SkillLevelCaps{magic}), skillLevelCapHashes(caps), "choosing the file imports it")
	c.Equal(plainDescriptions(sheet.Entity(), magic), sentences, "the list shows the imported cap")
}
