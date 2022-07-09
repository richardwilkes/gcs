/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package sheet

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	gsettings "github.com/richardwilkes/gcs/v5/model/gurps/settings"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
)

var (
	_ workspace.FileBackedDockable = &Template{}
	_ unison.UndoManagerProvider   = &Template{}
	_ widget.ModifiableRoot        = &Template{}
	_ widget.Rebuildable           = &Template{}
	_ widget.DockableKind          = &Template{}
	_ unison.TabCloser             = &Template{}
)

// Template holds the view for a GURPS character template.
type Template struct {
	unison.Panel
	path              string
	undoMgr           *unison.UndoManager
	scroll            *unison.ScrollPanel
	template          *gurps.Template
	crc               uint64
	scale             int
	content           *templateContent
	scaleField        *widget.PercentageField
	Traits            *PageList[*gurps.Trait]
	Skills            *PageList[*gurps.Skill]
	Spells            *PageList[*gurps.Spell]
	Equipment         *PageList[*gurps.Equipment]
	Notes             *PageList[*gurps.Note]
	needsSaveAsPrompt bool
}

// OpenTemplates returns the currently open templates.
func OpenTemplates() []*Template {
	var templates []*Template
	ws := workspace.FromWindowOrAny(unison.ActiveWindow())
	ws.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
		for _, one := range dc.Dockables() {
			if template, ok := one.(*Template); ok {
				templates = append(templates, template)
			}
		}
		return false
	})
	return templates
}

// NewTemplateFromFile loads a GURPS template file and creates a new unison.Dockable for it.
func NewTemplateFromFile(filePath string) (unison.Dockable, error) {
	template, err := gurps.NewTemplateFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	t := NewTemplate(filePath, template)
	t.needsSaveAsPrompt = false
	return t, nil
}

// NewTemplate creates a new unison.Dockable for GURPS template files.
func NewTemplate(filePath string, template *gurps.Template) *Template {
	d := &Template{
		path:              filePath,
		undoMgr:           unison.NewUndoManager(200, func(err error) { jot.Error(err) }),
		scroll:            unison.NewScrollPanel(),
		template:          template,
		scale:             settings.Global().General.InitialSheetUIScale,
		crc:               template.CRC64(),
		needsSaveAsPrompt: true,
	}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  unison.FillAlignment,
		VAlign:  unison.FillAlignment,
	})

	d.scroll.SetContent(d.createContent(), unison.UnmodifiedBehavior, unison.UnmodifiedBehavior)
	d.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})
	d.scroll.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, theme.PageVoidColor.Paint(gc, rect, unison.Fill))
	}

	scaleTitle := i18n.Text("Scale")
	d.scaleField = widget.NewPercentageField(scaleTitle, func() int { return d.scale }, func(v int) {
		d.scale = v
		d.applyScale()
	}, gsettings.InitialUIScaleMin, gsettings.InitialUIScaleMax, false, false)
	d.scaleField.SetMarksModified(false)
	d.scaleField.Tooltip = unison.NewTooltipWithText(scaleTitle)

	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	toolbar.AddChild(d.scaleField)
	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})

	d.AddChild(toolbar)
	d.AddChild(d.scroll)

	d.applyScale()

	d.InstallCmdHandlers(constants.SaveItemID, func(_ any) bool { return d.Modified() }, func(_ any) { d.save(false) })
	d.InstallCmdHandlers(constants.SaveAsItemID, unison.AlwaysEnabled, func(_ any) { d.save(true) })
	d.installNewItemCmdHandlers(constants.NewTraitItemID, constants.NewTraitContainerItemID, d.Traits)
	d.installNewItemCmdHandlers(constants.NewSkillItemID, constants.NewSkillContainerItemID, d.Skills)
	d.installNewItemCmdHandlers(constants.NewTechniqueItemID, -1, d.Skills)
	d.installNewItemCmdHandlers(constants.NewSpellItemID, constants.NewSpellContainerItemID, d.Spells)
	d.installNewItemCmdHandlers(constants.NewRitualMagicSpellItemID, -1, d.Spells)
	d.installNewItemCmdHandlers(constants.NewCarriedEquipmentItemID,
		constants.NewCarriedEquipmentContainerItemID, d.Equipment)
	d.installNewItemCmdHandlers(constants.NewNoteItemID, constants.NewNoteContainerItemID, d.Notes)
	d.InstallCmdHandlers(constants.AddNaturalAttacksItemID, unison.AlwaysEnabled, func(_ any) {
		ntable.InsertItems[*gurps.Trait](d, d.Traits.Table, d.template.TraitList, d.template.SetTraitList,
			func(_ *unison.Table[*ntable.Node[*gurps.Trait]]) []*ntable.Node[*gurps.Trait] {
				return d.Traits.provider.RootRows()
			}, gurps.NewNaturalAttacks(nil, nil))
	})

	return d
}

func (d *Template) installNewItemCmdHandlers(itemID, containerID int, creator itemCreator) {
	variant := ntable.NoItemVariant
	if containerID == -1 {
		variant = ntable.AlternateItemVariant
	} else {
		d.InstallCmdHandlers(containerID, unison.AlwaysEnabled,
			func(_ any) { creator.CreateItem(d, ntable.ContainerItemVariant) })
	}
	d.InstallCmdHandlers(itemID, unison.AlwaysEnabled, func(_ any) { creator.CreateItem(d, variant) })
}

// Entity implements gurps.EntityProvider
func (d *Template) Entity() *gurps.Entity {
	return nil
}

// DockableKind implements widget.DockableKind
func (d *Template) DockableKind() string {
	return widget.TemplateDockableKind
}

func (d *Template) applyScale() {
	d.scroll.Content().AsPanel().SetScale(float32(d.scale) / 100)
	d.scroll.Sync()
}

// UndoManager implements undo.Provider
func (d *Template) UndoManager() *unison.UndoManager {
	return d.undoMgr
}

// TitleIcon implements workspace.FileBackedDockable
func (d *Template) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  library.FileInfoFor(d.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements workspace.FileBackedDockable
func (d *Template) Title() string {
	return fs.BaseName(d.path)
}

func (d *Template) String() string {
	return d.Title()
}

// Tooltip implements workspace.FileBackedDockable
func (d *Template) Tooltip() string {
	return d.path
}

// BackingFilePath implements workspace.FileBackedDockable
func (d *Template) BackingFilePath() string {
	return d.path
}

// Modified implements workspace.FileBackedDockable
func (d *Template) Modified() bool {
	return d.crc != d.template.CRC64()
}

// MarkModified implements widget.ModifiableRoot.
func (d *Template) MarkModified() {
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
}

// MayAttemptClose implements unison.TabCloser
func (d *Template) MayAttemptClose() bool {
	return workspace.MayAttemptCloseOfGroup(d)
}

// AttemptClose implements unison.TabCloser
func (d *Template) AttemptClose() bool {
	if !workspace.CloseGroup(d) {
		return false
	}
	if d.Modified() {
		switch unison.YesNoCancelDialog(fmt.Sprintf(i18n.Text("Save changes made to\n%s?"), d.Title()), "") {
		case unison.ModalResponseDiscard:
		case unison.ModalResponseOK:
			if !d.save(false) {
				return false
			}
		case unison.ModalResponseCancel:
			return false
		}
	}
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.Close(d)
	}
	return true
}

func (d *Template) createContent() unison.Paneler {
	d.content = newTemplateContent()
	d.createLists()
	return d.content
}

func (d *Template) save(forceSaveAs bool) bool {
	success := false
	if forceSaveAs || d.needsSaveAsPrompt {
		success = workspace.SaveDockableAs(d, library.TemplatesExt, d.template.Save, func(path string) {
			d.crc = d.template.CRC64()
			d.path = path
		})
	} else {
		success = workspace.SaveDockable(d, d.template.Save, func() { d.crc = d.template.CRC64() })
	}
	if success {
		d.needsSaveAsPrompt = false
	}
	return success
}

func (d *Template) createLists() {
	h, v := d.scroll.Position()
	var refocusOnKey string
	var refocusOn unison.Paneler
	if wnd := d.Window(); wnd != nil {
		if focus := wnd.Focus(); focus != nil {
			// For page lists, the focus will be the table, so we need to look up a level
			if focus = focus.Parent(); focus != nil {
				switch focus.Self {
				case d.Traits:
					refocusOnKey = gurps.BlockLayoutTraitsKey
				case d.Skills:
					refocusOnKey = gurps.BlockLayoutSkillsKey
				case d.Spells:
					refocusOnKey = gurps.BlockLayoutSpellsKey
				case d.Equipment:
					refocusOnKey = gurps.BlockLayoutEquipmentKey
				case d.Notes:
					refocusOnKey = gurps.BlockLayoutNotesKey
				}
			}
		}
	}
	d.content.RemoveAllChildren()
	for _, col := range settings.Global().Sheet.BlockLayout.ByRow() {
		rowPanel := unison.NewPanel()
		for _, c := range col {
			switch c {
			case gurps.BlockLayoutTraitsKey:
				if d.Traits == nil {
					d.Traits = NewTraitsPageList(d, d.template)
				} else {
					d.Traits.Sync()
				}
				rowPanel.AddChild(d.Traits)
				if c == refocusOnKey {
					refocusOn = d.Traits.Table
				}
			case gurps.BlockLayoutSkillsKey:
				if d.Skills == nil {
					d.Skills = NewSkillsPageList(d, d.template)
				} else {
					d.Skills.Sync()
				}
				rowPanel.AddChild(d.Skills)
				if c == refocusOnKey {
					refocusOn = d.Skills.Table
				}
			case gurps.BlockLayoutSpellsKey:
				if d.Spells == nil {
					d.Spells = NewSpellsPageList(d, d.template)
				} else {
					d.Spells.Sync()
				}
				rowPanel.AddChild(d.Spells)
				if c == refocusOnKey {
					refocusOn = d.Spells.Table
				}
			case gurps.BlockLayoutEquipmentKey:
				if d.Equipment == nil {
					d.Equipment = NewCarriedEquipmentPageList(d, d.template)
				} else {
					d.Equipment.Sync()
				}
				rowPanel.AddChild(d.Equipment)
				if c == refocusOnKey {
					refocusOn = d.Equipment.Table
				}
			case gurps.BlockLayoutNotesKey:
				if d.Notes == nil {
					d.Notes = NewNotesPageList(d, d.template)
				} else {
					d.Notes.Sync()
				}
				rowPanel.AddChild(d.Notes)
				if c == refocusOnKey {
					refocusOn = d.Notes.Table
				}
			}
		}
		if len(rowPanel.Children()) != 0 {
			rowPanel.SetLayout(&unison.FlexLayout{
				Columns:      len(rowPanel.Children()),
				HSpacing:     1,
				HAlign:       unison.FillAlignment,
				EqualColumns: true,
			})
			rowPanel.SetLayoutData(&unison.FlexLayoutData{
				HAlign: unison.FillAlignment,
				VAlign: unison.StartAlignment,
				HGrab:  true,
			})
			d.content.AddChild(rowPanel)
		}
	}
	d.content.ApplyPreferredSize()
	if refocusOn != nil {
		refocusOn.AsPanel().RequestFocus()
	}
	d.scroll.SetPosition(h, v)
}

// SheetSettingsUpdated implements gurps.SheetSettingsResponder.
func (d *Template) SheetSettingsUpdated(entity *gurps.Entity, blockLayout bool) {
	if entity == nil {
		d.Rebuild(blockLayout)
	}
}

// Rebuild implements widget.Rebuildable.
func (d *Template) Rebuild(full bool) {
	h, v := d.scroll.Position()
	if full {
		traitsSelMap := d.Traits.RecordSelection()
		skillsSelMap := d.Skills.RecordSelection()
		spellsSelMap := d.Spells.RecordSelection()
		equipmentSelMap := d.Equipment.RecordSelection()
		notesSelMap := d.Notes.RecordSelection()
		defer func() {
			d.Traits.ApplySelection(traitsSelMap)
			d.Skills.ApplySelection(skillsSelMap)
			d.Spells.ApplySelection(spellsSelMap)
			d.Equipment.ApplySelection(equipmentSelMap)
			d.Notes.ApplySelection(notesSelMap)
		}()
		d.createLists()
	}
	widget.DeepSync(d)
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
	d.scroll.SetPosition(h, v)
}
