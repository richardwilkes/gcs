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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xrand"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ FileBackedDockable           = &LootSheet{}
	_ ExportDockable               = &LootSheet{}
	_ unison.UndoManagerProvider   = &LootSheet{}
	_ ModifiableRoot               = &LootSheet{}
	_ Rebuildable                  = &LootSheet{}
	_ unison.TabCloser             = &LootSheet{}
	_ gurps.SheetSettingsResponder = &LootSheet{}
	_ KeyedDockable                = &LootSheet{}
)

// LootSheet holds the view for a loot sheet.
type LootSheet struct {
	pageView
	content   *unison.Panel
	loot      *gurps.Loot
	Equipment *PageList[*gurps.Equipment]
	Notes     *PageList[*gurps.Note]
}

// OpenLootSheets returns the currently open loot sheets.
func OpenLootSheets(exclude *LootSheet) []*LootSheet {
	var result []*LootSheet
	for _, one := range AllDockables() {
		if loot, ok := one.(*LootSheet); ok && loot != exclude {
			result = append(result, loot)
		}
	}
	return result
}

// NewLootSheetFromFile loads a loot sheet file and creates a new unison.Dockable for it.
func NewLootSheetFromFile(filePath string) (unison.Dockable, error) {
	return openDockableFromFile(filePath, gurps.NewLootFromFile, NewLootSheet)
}

// NewLootSheet creates a new unison.Dockable for loot sheet files.
func NewLootSheet(filePath string, loot *gurps.Loot) *LootSheet {
	l := &LootSheet{
		content: unison.NewPanel(),
		loot:    loot,
	}
	l.initPageDockable(l, filePath, gurps.LootExt, loot.Save, loot)

	l.content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: 1,
	})
	l.content.AddChild(createLootTopBlock(l.loot, l.targetMgr))
	l.createLists()
	l.finishPageDockable(l, l.content)

	installListItemCmdHandlers(l, listItemCreators{
		otherEquipment: func() itemCreator { return l.Equipment },
		notes:          func() itemCreator { return l.Notes },
	})
	InstallExportCmdHandlers(l)

	prepareForPage(l.loot)
	return l
}

func (l *LootSheet) createToolbar() {
	l.toolbar = newToolbar()
	l.AddChild(l.toolbar)
	l.toolbar.AddChild(NewDefaultInfoPop())
	addUIScaleField(l.toolbar, func() int { return gurps.GlobalSettings().General.InitialSheetUIScale },
		func() int { return l.scale }, func(scale int) { l.scale = scale }, true, l.scroll)

	hierarchyButton := unison.NewSVGButton(svg.Hierarchy)
	hierarchyButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all hierarchical rows"))
	hierarchyButton.ClickCallback = l.toggleHierarchy
	l.toolbar.AddChild(hierarchyButton)

	noteToggleButton := unison.NewSVGButton(svg.NotesToggle)
	noteToggleButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all embedded notes"))
	noteToggleButton.ClickCallback = l.toggleNotes
	l.toolbar.AddChild(noteToggleButton)

	syncSourceButton := unison.NewSVGButton(svg.DownToBracket)
	syncSourceButton.Tooltip = newWrappedTooltip(i18n.Text("Sync with all sources in this sheet"))
	syncSourceButton.ClickCallback = func() { l.syncWithAllSources() }
	l.toolbar.AddChild(syncSourceButton)

	treasureButton := unison.NewSVGButton(svg.MagicWand)
	treasureButton.Tooltip = newWrappedTooltip(i18n.Text("Generate a treasure horde from this loot sheet"))
	treasureButton.ClickCallback = func() { l.generateTreasure() }
	l.toolbar.AddChild(treasureButton)

	l.searchTracker = installListSearchTracker(l.toolbar, l.lists)

	finishToolbarLayout(l.toolbar)
}

const (
	lootPanelFieldPrefix         = "loot:"
	lootPanelNameFieldRefKey     = lootPanelFieldPrefix + "name"
	lootPanelLocationFieldRefKey = lootPanelFieldPrefix + "location"
	lootPanelSessionFieldRefKey  = lootPanelFieldPrefix + "session"
)

func createLootTopBlock(loot *gurps.Loot, targetMgr *TargetMgr) *Page {
	page := NewPage(loot)
	top := unison.NewPanel()
	_, layoutData := initTitledPagePanel(top, i18n.Text("Loot"), 2, false, nil)
	layoutData.HGrab = true
	addLootTextField(top, targetMgr, lootPanelNameFieldRefKey, i18n.Text("Name"), &loot.Name)
	addLootTextField(top, targetMgr, lootPanelLocationFieldRefKey, i18n.Text("Location"), &loot.Location)
	addLootTextField(top, targetMgr, lootPanelSessionFieldRefKey, i18n.Text("Session"), &loot.Session)
	page.AddChild(top)
	return page
}

func addLootTextField(parent *unison.Panel, targetMgr *TargetMgr, targetKey, title string, field *string) {
	addLabeledStringPageField(parent, targetMgr, targetKey, title, NewPageLabel,
		func() string { return *field },
		func(s string) { *field = s })
}

func (l *LootSheet) keyToPanel(key *uti.DataType) *unison.Panel {
	var p unison.Paneler
	switch key {
	case equipmentDragKey:
		p = l.Equipment.Table
	case noteDragKey:
		p = l.Notes.Table
	default:
		return nil
	}
	return p.AsPanel()
}

// Entity implements EntityPanel. A loot sheet has no entity, so nil is always returned.
func (l *LootSheet) Entity() *gurps.Entity {
	return nil
}

// BackingFilePath implements FileBackedDockable. A loot sheet that has never been saved goes by its name.
func (l *LootSheet) BackingFilePath() string {
	if l.needsSaveAsPrompt {
		return unsavedFileName(l.loot.Name, i18n.Text("Unnamed Loot"), gurps.LootExt)
	}
	return l.path
}

// MarkModified implements ModifiableRoot.
func (l *LootSheet) MarkModified(_ unison.Paneler) {
	l.markModified(l, l.bumpModificationTimestamp)
}

// bumpModificationTimestamp implements modificationTimestampBumper.
func (l *LootSheet) bumpModificationTimestamp() {
	l.loot.ModifiedOn = jio.Now()
}

func (l *LootSheet) syncWithAllSources() {
	syncWithAllSources(l, l.loot, l.Equipment, l.Notes)
}

// Rebuild implements Rebuildable.
func (l *LootSheet) Rebuild(full bool) {
	gurps.DiscardGlobalResolveCache()
	prepareForPage(l.loot)
	state := l.captureViewState()
	if full {
		defer preserveSelections(l.lists)()
		l.createLists()
	}
	l.resync(l, state)
}

func (l *LootSheet) createLists() {
	children := l.content.Children()
	if len(children) == 0 {
		return
	}
	page, ok := children[0].Self.(*Page)
	if !ok {
		return
	}
	if children = page.Children(); len(children) == 0 {
		return
	}
	for i := len(children) - 1; i > 0; i-- {
		page.RemoveChildAtIndex(i)
	}
	syncOrRebuildList(&l.Equipment, func() *PageList[*gurps.Equipment] { return NewOtherEquipmentPageList(l, l.loot) })
	syncOrRebuildList(&l.Notes, func() *PageList[*gurps.Note] { return NewNotesPageList(l, l.loot) })
	for _, list := range l.lists() {
		p := list.AsPanel()
		p.SetLayoutData(&unison.FlexLayoutData{
			HAlign: align.Fill,
			HGrab:  true,
		})
		page.AddChild(p)
	}
	page.ApplyPreferredSize()
}

// lists returns the loot sheet's two lists, in the order they appear on the page. While the sheet is first being put
// together, the lists it hasn't built yet are present but nil.
func (l *LootSheet) lists() []sheetList {
	return []sheetList{l.Equipment, l.Notes}
}

// PageInfoProvider returns the page info provider for this sheet.
func (l *LootSheet) PageInfoProvider() gurps.PageInfoProvider {
	return l.loot
}

// SheetSettingsUpdated implements gurps.SheetSettingsResponder. A loot sheet has no entity of its own and reads the
// global sheet settings (see Loot.WeightUnit and Loot.PageSettings), so only a change to those -- reported with a nil
// entity -- concerns it, just as with a template. Responding to one character's per-sheet settings would bump the loot
// sheet's modification timestamp for an edit that was never made to it.
func (l *LootSheet) SheetSettingsUpdated(entity *gurps.Entity, fullRebuild bool) {
	if entity == nil {
		// A single rebuild both reports the change and refreshes everything the settings affect; marking the sheet as
		// modified first would only perform the same update a second time (see rebuildAsModified).
		rebuildAsModified(l, fullRebuild)
	}
}

func (l *LootSheet) toggleHierarchy() {
	toggleHierarchy(l.lists()...)
	l.Rebuild(true)
}

func (l *LootSheet) toggleNotes() {
	if toggleNotes(l.lists()...) {
		l.Rebuild(true)
	}
}

func (l *LootSheet) generateTreasure() {
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	markdown := unison.NewMarkdown(false)
	markdown.SetContent(i18n.Text(`# Treasure Generation

This will generate a new Loot Sheet with items from the contents of this one.
Each top-level item in this sheet will be treated as a potential item to select from.
The quantity of that top-level item will be used to determine the likelihood of it
being selected, with larger numbers increasing the chance it is chosen.`), 400)
	content.AddChild(markdown)
	settings := gurps.GlobalSettings()
	gen := newTreasureGenPanel(settings.LootGenMinValue, settings.LootGenMaxValue)
	content.AddChild(gen.panel)
	icon := &unison.DrawableSVG{
		SVG:  svg.MagicWand,
		Size: geom.Size{Width: 48, Height: 48},
	}
	var err error
	if gen.dialog, err = unison.NewDialog(icon, unison.DefaultDialogTheme.QuestionIconInk, content,
		[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()},
		unison.FloatingWindowOption(), unison.NotResizableWindowOption()); err != nil {
		errs.Log(err)
		return
	}
	gen.validateOK() // The fields couldn't set the button state while they were being created, so do it now
	if gen.dialog.RunModal() == unison.ModalResponseOK {
		minValue := gen.minValue
		maxValue := gen.maxValue
		var current fxp.Int
		r := xrand.New()
		settings.LootGenMinValue = minValue
		settings.LootGenMaxValue = maxValue
		m := make(map[*gurps.Equipment]int)
		for range 10 {
			choices, total, highest := pruneEquipmentList(maxValue-current, l.loot.Equipment)
			for len(choices) > 0 && current < minValue {
				found := false
				choice := fxp.Int(r.Intn(int(total)))
				for _, item := range choices {
					if item.Quantity >= choice {
						singleValue := item.ExtendedValueOfJustOne()
						switch {
						case len(choices) == 1:
							count := (minValue - current).Div(singleValue).Ceil()
							if current+count.Mul(singleValue) > maxValue {
								count = count.Dec()
							}
							m[item] += count.AsInteger[int]()
							current += count.Mul(singleValue)
						case singleValue < fxp.OneHundredth && minValue-current > fxp.One:
							count := fxp.One.Div(singleValue).Ceil()
							if current+count.Mul(singleValue) > maxValue {
								count = count.Dec()
							}
							m[item] += count.AsInteger[int]()
							current += count.Mul(singleValue)
						default:
							m[item]++
							current += singleValue
						}
						found = true
						break
					}
					choice -= item.Quantity
				}
				if !found || current >= minValue {
					break
				}
				remaining := maxValue - current
				if highest > remaining {
					choices, total, highest = pruneEquipmentList(remaining, choices)
				}
			}
			if current >= minValue {
				break
			}
			current = 0
			clear(m)
		}
		if current < minValue {
			unison.ErrorDialogWithMessage(i18n.Text("Unable to generate treasure!"),
				fmt.Sprintf(i18n.Text(`The minimum value of $%s could not be reached while staying at
or under the maximum value of $%s with the available items.`),
					minValue.Comma(), maxValue.Comma()))
			return
		}
		loot := gurps.NewLoot()
		for item, quantity := range m {
			clone := item.Clone(gurps.LibraryFile{}, gurps.EntityFromNode(item), nil, gurps.Reference)
			clone.Quantity = fxp.FromInteger(quantity)
			loot.Equipment = append(loot.Equipment, clone)
		}
		loot.EnsureAttachments()
		sheet := NewLootSheet("untitled"+gurps.LootExt, loot)
		sheet.hash = 0 // Force it to be recognized as unsaved
		DisplayNewDockable(sheet)
	}
}

type treasureGenPanel struct {
	panel    *unison.Panel
	dialog   *unison.Dialog
	minField *DecimalField
	maxField *DecimalField
	minValue fxp.Int
	maxValue fxp.Int
}

func newTreasureGenPanel(minValue, maxValue fxp.Int) *treasureGenPanel {
	p := &treasureGenPanel{
		panel:    unison.NewPanel(),
		minValue: minValue,
		maxValue: maxValue,
	}
	p.panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	p.panel.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Middle})
	p.panel.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: unison.StdVSpacing * 4}))
	label := i18n.Text("Target (Minimum) Value")
	p.panel.AddChild(NewFieldLeadingLabel(label, false))
	p.minField = NewDecimalField(nil, "", label,
		func() fxp.Int { return p.minValue },
		func(value fxp.Int) {
			p.minValue = value
			p.validateOK()
		},
		fxp.One, fxp.TenMillionMinusOne, false, false)
	p.panel.AddChild(p.minField)
	label = i18n.Text("Maximum Value")
	p.panel.AddChild(NewFieldLeadingLabel(label, false))
	p.maxField = NewDecimalField(nil, "", label,
		func() fxp.Int { return p.maxValue },
		func(value fxp.Int) {
			p.maxValue = value
			p.validateOK()
		},
		fxp.One, fxp.TenMillionMinusOne, false, false)
	p.panel.AddChild(p.maxField)
	return p
}

func (p *treasureGenPanel) validateOK() {
	if p.dialog == nil || p.minField == nil || p.maxField == nil {
		return
	}
	p.dialog.Button(unison.ModalResponseOK).SetEnabled(p.minValue <= p.maxValue && !p.minField.Invalid() &&
		!p.maxField.Invalid())
}

func pruneEquipmentList(remaining fxp.Int, items []*gurps.Equipment) (revisedItems []*gurps.Equipment, total, highest fxp.Int) {
	for _, item := range items {
		if item.Quantity > 0 {
			one := item.ExtendedValueOfJustOne()
			if one > 0 && one <= remaining {
				revisedItems = append(revisedItems, item)
				total += item.Quantity
				if highest < one {
					highest = one
				}
			}
		}
	}
	return revisedItems, total, highest
}
