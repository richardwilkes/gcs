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
	"time"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	gsettings "github.com/richardwilkes/gcs/v5/model/gurps/settings"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	wsettings "github.com/richardwilkes/gcs/v5/ui/workspace/settings"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
)

var (
	_ workspace.FileBackedDockable = &Sheet{}
	_ unison.UndoManagerProvider   = &Sheet{}
	_ widget.ModifiableRoot        = &Sheet{}
	_ widget.Rebuildable           = &Sheet{}
	_ widget.DockableKind          = &Sheet{}
	_ unison.TabCloser             = &Sheet{}
)

type itemCreator interface {
	CreateItem(widget.Rebuildable, ntable.ItemVariant)
}

// Sheet holds the view for a GURPS character sheet.
type Sheet struct {
	unison.Panel
	path                 string
	undoMgr              *unison.UndoManager
	scroll               *unison.ScrollPanel
	entity               *gurps.Entity
	crc                  uint64
	scale                int
	scaleField           *widget.PercentageField
	pages                *unison.Panel
	PortraitPanel        *PortraitPanel
	IdentityPanel        *IdentityPanel
	MiscPanel            *MiscPanel
	DescriptionPanel     *DescriptionPanel
	PointsPanel          *PointsPanel
	PrimaryAttrPanel     *PrimaryAttrPanel
	SecondaryAttrPanel   *SecondaryAttrPanel
	PointPoolsPanel      *PointPoolsPanel
	BodyPanel            *BodyPanel
	EncumbrancePanel     *EncumbrancePanel
	LiftingPanel         *LiftingPanel
	DamagePanel          *DamagePanel
	Reactions            *PageList[*gurps.ConditionalModifier]
	ConditionalModifiers *PageList[*gurps.ConditionalModifier]
	MeleeWeapons         *PageList[*gurps.Weapon]
	RangedWeapons        *PageList[*gurps.Weapon]
	Traits               *PageList[*gurps.Trait]
	Skills               *PageList[*gurps.Skill]
	Spells               *PageList[*gurps.Spell]
	CarriedEquipment     *PageList[*gurps.Equipment]
	OtherEquipment       *PageList[*gurps.Equipment]
	Notes                *PageList[*gurps.Note]
	awaitingUpdate       bool
	needsSaveAsPrompt    bool
}

// ActiveSheet returns the currently active sheet.
func ActiveSheet() *Sheet {
	d := workspace.ActiveDockable()
	if d == nil {
		return nil
	}
	if s, ok := d.(*Sheet); ok {
		return s
	}
	return nil
}

// NewSheetFromFile loads a GURPS character sheet file and creates a new unison.Dockable for it.
func NewSheetFromFile(filePath string) (unison.Dockable, error) {
	entity, err := gurps.NewEntityFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	s := NewSheet(filePath, entity)
	s.needsSaveAsPrompt = false
	return s, nil
}

// NewSheet creates a new unison.Dockable for GURPS character sheet files.
func NewSheet(filePath string, entity *gurps.Entity) *Sheet {
	s := &Sheet{
		path:              filePath,
		undoMgr:           unison.NewUndoManager(200, func(err error) { jot.Error(err) }),
		scroll:            unison.NewScrollPanel(),
		entity:            entity,
		crc:               entity.CRC64(),
		scale:             settings.Global().General.InitialSheetUIScale,
		pages:             unison.NewPanel(),
		needsSaveAsPrompt: true,
	}
	s.Self = s
	s.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  unison.FillAlignment,
		VAlign:  unison.FillAlignment,
	})

	s.pages.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: 1,
	})
	s.pages.AddChild(s.createTopBlock())
	s.createLists()
	s.scroll.SetContent(s.pages, unison.UnmodifiedBehavior, unison.UnmodifiedBehavior)
	s.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})
	s.scroll.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, theme.PageVoidColor.Paint(gc, rect, unison.Fill))
	}

	sheetSettingsButton := unison.NewSVGButton(res.SettingsSVG)
	sheetSettingsButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Sheet Settings"))
	sheetSettingsButton.ClickCallback = func() { wsettings.ShowSheetSettings(s) }

	scaleTitle := i18n.Text("Scale")
	s.scaleField = widget.NewPercentageField(scaleTitle, func() int { return s.scale }, func(v int) {
		s.scale = v
		s.applyScale()
	}, gsettings.InitialUIScaleMin, gsettings.InitialUIScaleMax, false, false)
	s.scaleField.SetMarksModified(false)
	s.scaleField.Tooltip = unison.NewTooltipWithText(scaleTitle)

	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	toolbar.AddChild(sheetSettingsButton)
	toolbar.AddChild(s.scaleField)
	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})

	s.AddChild(toolbar)
	s.AddChild(s.scroll)

	s.applyScale()

	s.InstallCmdHandlers(constants.SaveItemID, func(_ any) bool { return s.Modified() }, func(_ any) { s.save(false) })
	s.InstallCmdHandlers(constants.SaveAsItemID, unison.AlwaysEnabled, func(_ any) { s.save(true) })
	s.installNewItemCmdHandlers(constants.NewTraitItemID, constants.NewTraitContainerItemID, s.Traits)
	s.installNewItemCmdHandlers(constants.NewSkillItemID, constants.NewSkillContainerItemID, s.Skills)
	s.installNewItemCmdHandlers(constants.NewTechniqueItemID, -1, s.Skills)
	s.installNewItemCmdHandlers(constants.NewSpellItemID, constants.NewSpellContainerItemID, s.Spells)
	s.installNewItemCmdHandlers(constants.NewRitualMagicSpellItemID, -1, s.Spells)
	s.installNewItemCmdHandlers(constants.NewCarriedEquipmentItemID, constants.NewCarriedEquipmentContainerItemID,
		s.CarriedEquipment)
	s.installNewItemCmdHandlers(constants.NewOtherEquipmentItemID, constants.NewOtherEquipmentContainerItemID,
		s.OtherEquipment)
	s.installNewItemCmdHandlers(constants.NewNoteItemID, constants.NewNoteContainerItemID, s.Notes)
	s.InstallCmdHandlers(constants.AddNaturalAttacksItemID, unison.AlwaysEnabled, func(_ any) {
		ntable.InsertItem[*gurps.Trait](s, s.Traits.table, gurps.NewNaturalAttacks(s.entity, nil),
			s.entity.TraitList, s.entity.SetTraitList,
			func(_ *unison.Table[*ntable.Node[*gurps.Trait]]) []*ntable.Node[*gurps.Trait] {
				return s.Traits.provider.RootRows()
			})
	})

	return s
}

func (s *Sheet) installNewItemCmdHandlers(itemID, containerID int, creator itemCreator) {
	variant := ntable.NoItemVariant
	if containerID == -1 {
		variant = ntable.AlternateItemVariant
	} else {
		s.InstallCmdHandlers(containerID, unison.AlwaysEnabled,
			func(_ any) { creator.CreateItem(s, ntable.ContainerItemVariant) })
	}
	s.InstallCmdHandlers(itemID, unison.AlwaysEnabled, func(_ any) { creator.CreateItem(s, variant) })
}

// DockableKind implements widget.DockableKind
func (s *Sheet) DockableKind() string {
	return widget.SheetDockableKind
}

// Entity returns the entity this is displaying information for.
func (s *Sheet) Entity() *gurps.Entity {
	return s.entity
}

// UndoManager implements undo.Provider
func (s *Sheet) UndoManager() *unison.UndoManager {
	return s.undoMgr
}

func (s *Sheet) applyScale() {
	s.pages.SetScale(float32(s.scale) / 100)
	s.scroll.Sync()
}

// TitleIcon implements workspace.FileBackedDockable
func (s *Sheet) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  library.FileInfoFor(s.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements workspace.FileBackedDockable
func (s *Sheet) Title() string {
	return fs.BaseName(s.path)
}

func (s *Sheet) String() string {
	return s.Title()
}

// Tooltip implements workspace.FileBackedDockable
func (s *Sheet) Tooltip() string {
	return s.path
}

// BackingFilePath implements workspace.FileBackedDockable
func (s *Sheet) BackingFilePath() string {
	return s.path
}

// Modified implements workspace.FileBackedDockable
func (s *Sheet) Modified() bool {
	return s.crc != s.entity.CRC64()
}

// MarkModified implements widget.ModifiableRoot.
func (s *Sheet) MarkModified() {
	if !s.awaitingUpdate {
		s.awaitingUpdate = true
		unison.InvokeTaskAfter(func() {
			s.MiscPanel.UpdateModified()
			// TODO: This is still too slow when the lists have more than a few rows of content.
			//       It impinges on interactive typing. Looks like most of the time is spent in updating the tables.
			//       Unfortunately, there isn't a fast way to determine that the content doesn't need to be refreshed.
			widget.DeepSync(s)
			if dc := unison.Ancestor[*unison.DockContainer](s); dc != nil {
				dc.UpdateTitle(s)
			}
			s.awaitingUpdate = false
		}, time.Millisecond*100)
	}
}

// MayAttemptClose implements unison.TabCloser
func (s *Sheet) MayAttemptClose() bool {
	return workspace.MayAttemptCloseOfGroup(s)
}

// AttemptClose implements unison.TabCloser
func (s *Sheet) AttemptClose() bool {
	if !workspace.CloseGroup(s) {
		return false
	}
	if s.Modified() {
		switch unison.YesNoCancelDialog(fmt.Sprintf(i18n.Text("Save changes made to\n%s?"), s.Title()), "") {
		case unison.ModalResponseDiscard:
		case unison.ModalResponseOK:
			if !s.save(false) {
				return false
			}
		case unison.ModalResponseCancel:
			return false
		}
	}
	if dc := unison.Ancestor[*unison.DockContainer](s); dc != nil {
		dc.Close(s)
	}
	return true
}

func (s *Sheet) save(forceSaveAs bool) bool {
	success := false
	if forceSaveAs || s.needsSaveAsPrompt {
		success = workspace.SaveDockableAs(s, library.SheetExt, s.entity.Save, func(path string) {
			s.crc = s.entity.CRC64()
			s.path = path
		})
	} else {
		success = workspace.SaveDockable(s, s.entity.Save, func() { s.crc = s.entity.CRC64() })
	}
	if success {
		s.needsSaveAsPrompt = false
	}
	return success
}

func (s *Sheet) createTopBlock() *Page {
	p := NewPage(s.entity)
	p.AddChild(s.createFirstRow())
	p.AddChild(s.createSecondRow())
	return p
}

func (s *Sheet) createFirstRow() *unison.Panel {
	s.PortraitPanel = NewPortraitPanel(s.entity)
	s.IdentityPanel = NewIdentityPanel(s.entity)
	s.MiscPanel = NewMiscPanel(s.entity)
	s.DescriptionPanel = NewDescriptionPanel(s.entity)
	s.PointsPanel = NewPointsPanel(s.entity)

	right := unison.NewPanel()
	right.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: 1,
		VSpacing: 1,
		HAlign:   unison.FillAlignment,
		VAlign:   unison.FillAlignment,
	})

	right.AddChild(s.IdentityPanel)
	right.AddChild(s.MiscPanel)
	right.AddChild(s.PointsPanel)
	right.AddChild(s.DescriptionPanel)

	p := unison.NewPanel()
	p.SetLayout(&portraitLayout{
		portrait: s.PortraitPanel,
		rest:     right,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
	})
	p.AddChild(s.PortraitPanel)
	p.AddChild(right)

	return p
}

func (s *Sheet) createSecondRow() *unison.Panel {
	p := unison.NewPanel()
	p.SetLayout(&unison.FlexLayout{
		Columns:  4,
		HSpacing: 1,
		VSpacing: 1,
		HAlign:   unison.FillAlignment,
		VAlign:   unison.FillAlignment,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
	})

	s.PrimaryAttrPanel = NewPrimaryAttrPanel(s.entity)
	s.SecondaryAttrPanel = NewSecondaryAttrPanel(s.entity)
	s.PointPoolsPanel = NewPointPoolsPanel(s.entity)
	s.BodyPanel = NewBodyPanel(s.entity)
	s.EncumbrancePanel = NewEncumbrancePanel(s.entity)
	s.LiftingPanel = NewLiftingPanel(s.entity)
	s.DamagePanel = NewDamagePanel(s.entity)

	endWrapper := unison.NewPanel()
	endWrapper.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: 1,
	})
	endWrapper.SetLayoutData(&unison.FlexLayoutData{
		VSpan:  3,
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
	})
	endWrapper.AddChild(s.EncumbrancePanel)
	endWrapper.AddChild(s.LiftingPanel)

	p.AddChild(s.PrimaryAttrPanel)
	p.AddChild(s.SecondaryAttrPanel)
	p.AddChild(s.BodyPanel)
	p.AddChild(endWrapper)
	p.AddChild(s.DamagePanel)
	p.AddChild(s.PointPoolsPanel)

	return p
}

func (s *Sheet) createLists() {
	children := s.pages.Children()
	if len(children) == 0 {
		return
	}
	page, ok := children[0].Self.(*Page)
	if !ok {
		return
	}
	children = page.Children()
	if len(children) < 2 {
		return
	}
	h, v := s.scroll.Position()
	var refocusOnKey string
	var refocusOn unison.Paneler
	if wnd := s.Window(); wnd != nil {
		if focus := wnd.Focus(); focus != nil {
			// For page lists, the focus will be the table, so we need to look up a level
			if focus = focus.Parent(); focus != nil {
				switch focus.Self {
				case s.Reactions:
					refocusOnKey = gurps.BlockLayoutReactionsKey
				case s.ConditionalModifiers:
					refocusOnKey = gurps.BlockLayoutConditionalModifiersKey
				case s.MeleeWeapons:
					refocusOnKey = gurps.BlockLayoutMeleeKey
				case s.RangedWeapons:
					refocusOnKey = gurps.BlockLayoutRangedKey
				case s.Traits:
					refocusOnKey = gurps.BlockLayoutTraitsKey
				case s.Skills:
					refocusOnKey = gurps.BlockLayoutSkillsKey
				case s.Spells:
					refocusOnKey = gurps.BlockLayoutSpellsKey
				case s.CarriedEquipment:
					refocusOnKey = gurps.BlockLayoutEquipmentKey
				case s.OtherEquipment:
					refocusOnKey = gurps.BlockLayoutOtherEquipmentKey
				case s.Notes:
					refocusOnKey = gurps.BlockLayoutNotesKey
				}
			}
		}
	}
	for i := len(children) - 1; i > 1; i-- {
		page.RemoveChildAtIndex(i)
	}
	// Add the various blocks, based on the layout preference.
	for _, col := range s.entity.SheetSettings.BlockLayout.ByRow() {
		rowPanel := unison.NewPanel()
		rowPanel.SetLayout(&unison.FlexLayout{
			Columns:      len(col),
			HSpacing:     1,
			HAlign:       unison.FillAlignment,
			VAlign:       unison.FillAlignment,
			EqualColumns: true,
		})
		rowPanel.SetLayoutData(&unison.FlexLayoutData{
			HAlign: unison.FillAlignment,
			VAlign: unison.StartAlignment,
			HGrab:  true,
		})
		for _, c := range col {
			switch c {
			case gurps.BlockLayoutReactionsKey:
				s.Reactions = NewReactionsPageList(s.entity)
				rowPanel.AddChild(s.Reactions)
				if c == refocusOnKey {
					refocusOn = s.Reactions.table
				}
			case gurps.BlockLayoutConditionalModifiersKey:
				s.ConditionalModifiers = NewConditionalModifiersPageList(s.entity)
				rowPanel.AddChild(s.ConditionalModifiers)
				if c == refocusOnKey {
					refocusOn = s.ConditionalModifiers.table
				}
			case gurps.BlockLayoutMeleeKey:
				s.MeleeWeapons = NewMeleeWeaponsPageList(s.entity)
				rowPanel.AddChild(s.MeleeWeapons)
				if c == refocusOnKey {
					refocusOn = s.MeleeWeapons.table
				}
			case gurps.BlockLayoutRangedKey:
				s.RangedWeapons = NewRangedWeaponsPageList(s.entity)
				rowPanel.AddChild(s.RangedWeapons)
				if c == refocusOnKey {
					refocusOn = s.RangedWeapons.table
				}
			case gurps.BlockLayoutTraitsKey:
				s.Traits = NewTraitsPageList(s, s.entity)
				rowPanel.AddChild(s.Traits)
				if c == refocusOnKey {
					refocusOn = s.Traits.table
				}
			case gurps.BlockLayoutSkillsKey:
				s.Skills = NewSkillsPageList(s, s.entity)
				rowPanel.AddChild(s.Skills)
				if c == refocusOnKey {
					refocusOn = s.Skills.table
				}
			case gurps.BlockLayoutSpellsKey:
				s.Spells = NewSpellsPageList(s, s.entity)
				rowPanel.AddChild(s.Spells)
				if c == refocusOnKey {
					refocusOn = s.Spells.table
				}
			case gurps.BlockLayoutEquipmentKey:
				s.CarriedEquipment = NewCarriedEquipmentPageList(s, s.entity)
				rowPanel.AddChild(s.CarriedEquipment)
				if c == refocusOnKey {
					refocusOn = s.CarriedEquipment.table
				}
			case gurps.BlockLayoutOtherEquipmentKey:
				s.OtherEquipment = NewOtherEquipmentPageList(s, s.entity)
				rowPanel.AddChild(s.OtherEquipment)
				if c == refocusOnKey {
					refocusOn = s.OtherEquipment.table
				}
			case gurps.BlockLayoutNotesKey:
				s.Notes = NewNotesPageList(s, s.entity)
				rowPanel.AddChild(s.Notes)
				if c == refocusOnKey {
					refocusOn = s.Notes.table
				}
			}
		}
		page.AddChild(rowPanel)
	}
	page.ApplyPreferredSize()
	if refocusOn != nil {
		refocusOn.AsPanel().RequestFocus()
	}
	s.scroll.SetPosition(h, v)
}

// SheetSettingsUpdated implements gurps.SheetSettingsResponder.
func (s *Sheet) SheetSettingsUpdated(entity *gurps.Entity, blockLayout bool) {
	if s.entity == entity {
		s.Rebuild(blockLayout)
	}
}

// Rebuild implements widget.Rebuildable.
func (s *Sheet) Rebuild(full bool) {
	s.entity.Recalculate()
	if full {
		reactionsSelMap := s.Reactions.RecordSelection()
		conditionalModifiersSelMap := s.ConditionalModifiers.RecordSelection()
		meleeWeaponsSelMap := s.MeleeWeapons.RecordSelection()
		rangedWeaponsSelMap := s.RangedWeapons.RecordSelection()
		traitsSelMap := s.Traits.RecordSelection()
		skillsSelMap := s.Skills.RecordSelection()
		spellsSelMap := s.Spells.RecordSelection()
		carriedEquipmentSelMap := s.CarriedEquipment.RecordSelection()
		otherEquipmentSelMap := s.OtherEquipment.RecordSelection()
		notesSelMap := s.Notes.RecordSelection()
		defer func() {
			s.Reactions.ApplySelection(reactionsSelMap)
			s.ConditionalModifiers.ApplySelection(conditionalModifiersSelMap)
			s.MeleeWeapons.ApplySelection(meleeWeaponsSelMap)
			s.RangedWeapons.ApplySelection(rangedWeaponsSelMap)
			s.Traits.ApplySelection(traitsSelMap)
			s.Skills.ApplySelection(skillsSelMap)
			s.Spells.ApplySelection(spellsSelMap)
			s.CarriedEquipment.ApplySelection(carriedEquipmentSelMap)
			s.OtherEquipment.ApplySelection(otherEquipmentSelMap)
			s.Notes.ApplySelection(notesSelMap)
		}()
		s.createLists()
	}
	widget.DeepSync(s)
	if dc := unison.Ancestor[*unison.DockContainer](s); dc != nil {
		dc.UpdateTitle(s)
	}
}

func drawBandedBackground(p unison.Paneler, gc *unison.Canvas, rect unison.Rect, start, step int) {
	gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
	children := p.AsPanel().Children()
	for i := start; i < len(children); i += step {
		var ink unison.Ink
		if ((i-start)/step)&1 == 1 {
			ink = unison.BandingColor
		} else {
			ink = unison.ContentColor
		}
		r := children[i].FrameRect()
		r.X = rect.X
		r.Width = rect.Width
		gc.DrawRect(r, ink.Paint(gc, r, unison.Fill))
	}
}
