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

package lists

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/crc"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	gsettings "github.com/richardwilkes/gcs/v5/model/gurps/settings"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	"github.com/richardwilkes/gcs/v5/ui/workspace/editors"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
)

var (
	_ workspace.FileBackedDockable = &TableDockable[*gurps.Trait]{}
	_ unison.UndoManagerProvider   = &TableDockable[*gurps.Trait]{}
	_ widget.ModifiableRoot        = &TableDockable[*gurps.Trait]{}
	_ widget.Rebuildable           = &TableDockable[*gurps.Trait]{}
	_ widget.DockableKind          = &TableDockable[*gurps.Trait]{}
	_ unison.TabCloser             = &TableDockable[*gurps.Trait]{}
)

// TableDockable holds the view for a file that contains a (potentially hierarchical) list of data.
type TableDockable[T gurps.NodeConstraint[T]] struct {
	unison.Panel
	path              string
	extension         string
	undoMgr           *unison.UndoManager
	provider          ntable.TableProvider[T]
	saver             func(path string) error
	canCreateIDs      map[int]bool
	hierarchyButton   *unison.Button
	sizeToFitButton   *unison.Button
	scale             int
	scaleField        *widget.PercentageField
	backButton        *unison.Button
	forwardButton     *unison.Button
	searchField       *unison.Field
	matchesLabel      *unison.Label
	scroll            *unison.ScrollPanel
	tableHeader       *unison.TableHeader[*ntable.Node[T]]
	table             *unison.Table[*ntable.Node[T]]
	crc               uint64
	searchResult      []*ntable.Node[T]
	searchIndex       int
	needsSaveAsPrompt bool
}

type traitListProvider struct {
	traits []*gurps.Trait
}

func (p *traitListProvider) Entity() *gurps.Entity {
	return nil
}

func (p *traitListProvider) TraitList() []*gurps.Trait {
	return p.traits
}

func (p *traitListProvider) SetTraitList(list []*gurps.Trait) {
	p.traits = list
}

// NewTraitTableDockableFromFile loads a list of traits from a file and creates a new unison.Dockable for them.
func NewTraitTableDockableFromFile(filePath string) (unison.Dockable, error) {
	traits, err := gurps.NewTraitsFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewTraitTableDockable(filePath, traits)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewTraitTableDockable creates a new unison.Dockable for trait list files.
func NewTraitTableDockable(filePath string, traits []*gurps.Trait) *TableDockable[*gurps.Trait] {
	provider := &traitListProvider{traits: traits}
	return NewTableDockable(filePath, library.TraitsExt, editors.NewTraitsProvider(provider, false),
		func(path string) error { return gurps.SaveTraits(provider.TraitList(), path) },
		constants.NewTraitItemID, constants.NewTraitContainerItemID)
}

type traitModifierListProvider struct {
	modifiers []*gurps.TraitModifier
}

func (p *traitModifierListProvider) Entity() *gurps.Entity {
	return nil
}

func (p *traitModifierListProvider) TraitModifierList() []*gurps.TraitModifier {
	return p.modifiers
}

func (p *traitModifierListProvider) SetTraitModifierList(list []*gurps.TraitModifier) {
	p.modifiers = list
}

// NewTraitModifierTableDockableFromFile loads a list of trait modifiers from a file and creates a new
// unison.Dockable for them.
func NewTraitModifierTableDockableFromFile(filePath string) (unison.Dockable, error) {
	modifiers, err := gurps.NewTraitModifiersFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewTraitModifierTableDockable(filePath, modifiers)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewTraitModifierTableDockable creates a new unison.Dockable for trait modifier list files.
func NewTraitModifierTableDockable(filePath string, modifiers []*gurps.TraitModifier) *TableDockable[*gurps.TraitModifier] {
	provider := &traitModifierListProvider{modifiers: modifiers}
	return NewTableDockable(filePath, library.TraitModifiersExt,
		editors.NewTraitModifiersProvider(provider, false),
		func(path string) error { return gurps.SaveTraitModifiers(provider.TraitModifierList(), path) },
		constants.NewTraitModifierItemID, constants.NewTraitContainerModifierItemID)
}

type equipmentListProvider struct {
	carried []*gurps.Equipment
	other   []*gurps.Equipment
}

func (p *equipmentListProvider) Entity() *gurps.Entity {
	return nil
}

func (p *equipmentListProvider) CarriedEquipmentList() []*gurps.Equipment {
	return p.carried
}

func (p *equipmentListProvider) SetCarriedEquipmentList(list []*gurps.Equipment) {
	p.carried = list
}

func (p *equipmentListProvider) OtherEquipmentList() []*gurps.Equipment {
	return p.other
}

func (p *equipmentListProvider) SetOtherEquipmentList(list []*gurps.Equipment) {
	p.other = list
}

// NewEquipmentTableDockableFromFile loads a list of equipment from a file and creates a new unison.Dockable for them.
func NewEquipmentTableDockableFromFile(filePath string) (unison.Dockable, error) {
	equipment, err := gurps.NewEquipmentFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewEquipmentTableDockable(filePath, equipment)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewEquipmentTableDockable creates a new unison.Dockable for equipment list files.
func NewEquipmentTableDockable(filePath string, equipment []*gurps.Equipment) *TableDockable[*gurps.Equipment] {
	provider := &equipmentListProvider{other: equipment}
	return NewTableDockable(filePath, library.EquipmentExt, editors.NewEquipmentProvider(provider, false, false),
		func(path string) error { return gurps.SaveEquipment(provider.OtherEquipmentList(), path) },
		constants.NewCarriedEquipmentItemID, constants.NewCarriedEquipmentContainerItemID)
}

type equipmentModifierListProvider struct {
	modifiers []*gurps.EquipmentModifier
}

func (p *equipmentModifierListProvider) Entity() *gurps.Entity {
	return nil
}

func (p *equipmentModifierListProvider) EquipmentModifierList() []*gurps.EquipmentModifier {
	return p.modifiers
}

func (p *equipmentModifierListProvider) SetEquipmentModifierList(list []*gurps.EquipmentModifier) {
	p.modifiers = list
}

// NewEquipmentModifierTableDockableFromFile loads a list of equipment modifiers from a file and creates a new
// unison.Dockable for them.
func NewEquipmentModifierTableDockableFromFile(filePath string) (unison.Dockable, error) {
	modifiers, err := gurps.NewEquipmentModifiersFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewEquipmentModifierTableDockable(filePath, modifiers)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewEquipmentModifierTableDockable creates a new unison.Dockable for equipment modifier list files.
func NewEquipmentModifierTableDockable(filePath string, modifiers []*gurps.EquipmentModifier) *TableDockable[*gurps.EquipmentModifier] {
	provider := &equipmentModifierListProvider{modifiers: modifiers}
	return NewTableDockable(filePath, library.EquipmentModifiersExt,
		editors.NewEquipmentModifiersProvider(provider, false),
		func(path string) error { return gurps.SaveEquipmentModifiers(provider.EquipmentModifierList(), path) },
		constants.NewEquipmentModifierItemID, constants.NewEquipmentContainerModifierItemID)
}

type skillListProvider struct {
	skills []*gurps.Skill
}

func (p *skillListProvider) Entity() *gurps.Entity {
	return nil
}

func (p *skillListProvider) SkillList() []*gurps.Skill {
	return p.skills
}

func (p *skillListProvider) SetSkillList(list []*gurps.Skill) {
	p.skills = list
}

// NewSkillTableDockableFromFile loads a list of skills from a file and creates a new unison.Dockable for them.
func NewSkillTableDockableFromFile(filePath string) (unison.Dockable, error) {
	skills, err := gurps.NewSkillsFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewSkillTableDockable(filePath, skills)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewSkillTableDockable creates a new unison.Dockable for skill list files.
func NewSkillTableDockable(filePath string, skills []*gurps.Skill) *TableDockable[*gurps.Skill] {
	provider := &skillListProvider{skills: skills}
	return NewTableDockable(filePath, library.SkillsExt, editors.NewSkillsProvider(provider, false),
		func(path string) error { return gurps.SaveSkills(provider.SkillList(), path) },
		constants.NewSkillItemID, constants.NewSkillContainerItemID, constants.NewTechniqueItemID)
}

type spellListProvider struct {
	spells []*gurps.Spell
}

func (p *spellListProvider) Entity() *gurps.Entity {
	return nil
}

func (p *spellListProvider) SpellList() []*gurps.Spell {
	return p.spells
}

func (p *spellListProvider) SetSpellList(list []*gurps.Spell) {
	p.spells = list
}

// NewSpellTableDockableFromFile loads a list of spells from a file and creates a new unison.Dockable for them.
func NewSpellTableDockableFromFile(filePath string) (unison.Dockable, error) {
	spells, err := gurps.NewSpellsFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewSpellTableDockable(filePath, spells)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewSpellTableDockable creates a new unison.Dockable for spell list files.
func NewSpellTableDockable(filePath string, spells []*gurps.Spell) *TableDockable[*gurps.Spell] {
	provider := &spellListProvider{spells: spells}
	return NewTableDockable(filePath, library.SpellsExt, editors.NewSpellsProvider(provider, false),
		func(path string) error { return gurps.SaveSpells(provider.SpellList(), path) },
		constants.NewSpellItemID, constants.NewSpellContainerItemID, constants.NewRitualMagicSpellItemID)
}

type noteListProvider struct {
	notes []*gurps.Note
}

func (p *noteListProvider) Entity() *gurps.Entity {
	return nil
}

func (p *noteListProvider) NoteList() []*gurps.Note {
	return p.notes
}

func (p *noteListProvider) SetNoteList(list []*gurps.Note) {
	p.notes = list
}

// NewNoteTableDockableFromFile loads a list of notes from a file and creates a new unison.Dockable for them.
func NewNoteTableDockableFromFile(filePath string) (unison.Dockable, error) {
	notes, err := gurps.NewNotesFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewNoteTableDockable(filePath, notes)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewNoteTableDockable creates a new unison.Dockable for note list files.
func NewNoteTableDockable(filePath string, notes []*gurps.Note) *TableDockable[*gurps.Note] {
	provider := &noteListProvider{notes: notes}
	return NewTableDockable(filePath, library.NotesExt, editors.NewNotesProvider(provider, false),
		func(path string) error { return gurps.SaveNotes(provider.NoteList(), path) },
		constants.NewNoteItemID, constants.NewNoteContainerItemID)
}

// NewTableDockable creates a new TableDockable for list data files.
func NewTableDockable[T gurps.NodeConstraint[T]](filePath, extension string, provider ntable.TableProvider[T], saver func(path string) error, canCreateIDs ...int) *TableDockable[T] {
	header, table := ntable.NewNodeTable[T](provider, nil)
	d := &TableDockable[T]{
		path:              filePath,
		extension:         extension,
		undoMgr:           unison.NewUndoManager(200, func(err error) { jot.Error(err) }),
		provider:          provider,
		saver:             saver,
		canCreateIDs:      make(map[int]bool),
		scroll:            unison.NewScrollPanel(),
		tableHeader:       header,
		table:             table,
		scale:             settings.Global().General.InitialListUIScale,
		needsSaveAsPrompt: true,
	}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	for _, id := range canCreateIDs {
		d.canCreateIDs[id] = true
	}

	d.table.SyncToModel()
	d.table.SizeColumnsToFit(true)
	ntable.InstallTableDropSupport(d.table, d.provider)

	d.scroll.SetColumnHeader(d.tableHeader)
	d.scroll.SetContent(d.table, unison.FillBehavior, unison.FillBehavior)
	d.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})

	d.hierarchyButton = unison.NewSVGButton(res.HierarchySVG)
	d.hierarchyButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Opens/closes all hierarchical rows"))
	d.hierarchyButton.ClickCallback = d.toggleHierarchy

	d.sizeToFitButton = unison.NewSVGButton(res.SizeToFitSVG)
	d.sizeToFitButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Sets the width of each column to fit its contents"))
	d.sizeToFitButton.ClickCallback = d.sizeToFit

	scaleTitle := i18n.Text("Scale")
	d.scaleField = widget.NewPercentageField(scaleTitle, func() int { return d.scale }, func(v int) {
		d.scale = v
		d.applyScale()
	}, gsettings.InitialUIScaleMin, gsettings.InitialUIScaleMax, false, false)
	d.scaleField.Tooltip = unison.NewTooltipWithText(scaleTitle)

	d.backButton = unison.NewSVGButton(res.BackSVG)
	d.backButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Previous Match"))
	d.backButton.ClickCallback = d.previousMatch
	d.backButton.SetEnabled(false)

	d.forwardButton = unison.NewSVGButton(res.ForwardSVG)
	d.forwardButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Next Match"))
	d.forwardButton.ClickCallback = d.nextMatch
	d.forwardButton.SetEnabled(false)

	d.searchField = unison.NewField()
	search := i18n.Text("Search")
	d.searchField.Watermark = search
	d.searchField.Tooltip = unison.NewTooltipWithText(search)
	d.searchField.ModifiedCallback = d.searchModified
	d.searchField.KeyDownCallback = func(keyCode unison.KeyCode, mod unison.Modifiers, repeat bool) bool {
		if keyCode == unison.KeyReturn || keyCode == unison.KeyNumPadEnter {
			if mod.ShiftDown() {
				d.previousMatch()
			} else {
				d.nextMatch()
			}
			return true
		}
		return d.searchField.DefaultKeyDown(keyCode, mod, repeat)
	}
	d.searchField.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
		HGrab:  true,
	})

	d.matchesLabel = unison.NewLabel()
	d.matchesLabel.Text = "-"
	d.matchesLabel.Tooltip = unison.NewTooltipWithText(i18n.Text("Number of matches found"))

	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	toolbar.AddChild(d.hierarchyButton)
	toolbar.AddChild(d.sizeToFitButton)
	toolbar.AddChild(d.scaleField)
	toolbar.AddChild(d.backButton)
	toolbar.AddChild(d.forwardButton)
	toolbar.AddChild(d.searchField)
	toolbar.AddChild(d.matchesLabel)
	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})

	d.AddChild(toolbar)
	d.AddChild(d.scroll)

	d.applyScale()

	d.InstallCmdHandlers(constants.OpenEditorItemID,
		func(_ any) bool { return d.table.HasSelection() },
		func(_ any) { d.provider.OpenEditor(d, d.table) })
	d.InstallCmdHandlers(constants.OpenOnePageReferenceItemID,
		func(_ any) bool { return editors.CanOpenPageRef(d.table) },
		func(_ any) { editors.OpenPageRef(d.table) })
	d.InstallCmdHandlers(constants.OpenEachPageReferenceItemID,
		func(_ any) bool { return editors.CanOpenPageRef(d.table) },
		func(_ any) { editors.OpenEachPageRef(d.table) })
	d.InstallCmdHandlers(constants.SaveItemID,
		func(_ any) bool { return d.Modified() },
		func(_ any) { d.save(false) })
	d.InstallCmdHandlers(constants.SaveAsItemID, unison.AlwaysEnabled, func(_ any) { d.save(true) })
	d.InstallCmdHandlers(unison.DeleteItemID,
		func(_ any) bool { return d.table.HasSelection() },
		func(_ any) { ntable.DeleteSelection(d.table) })
	d.InstallCmdHandlers(constants.DuplicateItemID,
		func(_ any) bool { return d.table.HasSelection() },
		func(_ any) { ntable.DuplicateSelection(d.table) })
	for _, id := range canCreateIDs {
		variant := ntable.ItemVariant(-1)
		switch {
		case id > constants.FirstNonContainerMarker && id < constants.LastNonContainerMarker:
			variant = ntable.NoItemVariant
		case id > constants.FirstContainerMarker && id < constants.LastContainerMarker:
			variant = ntable.ContainerItemVariant
		case id > constants.FirstAlternateNonContainerMarker && id < constants.LastAlternateNonContainerMarker:
			variant = ntable.AlternateItemVariant
		}
		if variant != -1 {
			d.InstallCmdHandlers(id, unison.AlwaysEnabled,
				func(_ any) { d.provider.CreateItem(d, d.table, variant) })
		}
	}

	d.crc = d.crc64()
	return d
}

// UndoManager implements undo.Provider
func (d *TableDockable[T]) UndoManager() *unison.UndoManager {
	return d.undoMgr
}

// DockableKind implements widget.DockableKind
func (d *TableDockable[T]) DockableKind() string {
	return widget.ListDockableKind
}

func (d *TableDockable[T]) applyScale() {
	s := float32(d.scale) / 100
	d.tableHeader.SetScale(s)
	d.table.SetScale(s)
	d.scroll.Sync()
}

// TitleIcon implements workspace.FileBackedDockable
func (d *TableDockable[T]) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  library.FileInfoFor(d.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements workspace.FileBackedDockable
func (d *TableDockable[T]) Title() string {
	return fs.BaseName(d.path)
}

func (d *TableDockable[T]) String() string {
	return d.Title()
}

// Tooltip implements workspace.FileBackedDockable
func (d *TableDockable[T]) Tooltip() string {
	return d.path
}

// BackingFilePath implements workspace.FileBackedDockable
func (d *TableDockable[T]) BackingFilePath() string {
	return d.path
}

// Modified implements workspace.FileBackedDockable
func (d *TableDockable[T]) Modified() bool {
	return d.crc != d.crc64()
}

// MarkModified implements widget.ModifiableRoot.
func (d *TableDockable[T]) MarkModified() {
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
}

// MayAttemptClose implements unison.TabCloser
func (d *TableDockable[T]) MayAttemptClose() bool {
	return workspace.MayAttemptCloseOfGroup(d)
}

// AttemptClose implements unison.TabCloser
func (d *TableDockable[T]) AttemptClose() bool {
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

func (d *TableDockable[T]) save(forceSaveAs bool) bool {
	success := false
	if forceSaveAs || d.needsSaveAsPrompt {
		success = workspace.SaveDockableAs(d, d.extension, d.saver, func(path string) {
			d.crc = d.crc64()
			d.path = path
		})
	} else {
		success = workspace.SaveDockable(d, d.saver, func() { d.crc = d.crc64() })
	}
	if success {
		d.needsSaveAsPrompt = false
	}
	return success
}

func (d *TableDockable[T]) toggleHierarchy() {
	first := true
	open := false
	for _, row := range d.table.RootRows() {
		if row.CanHaveChildren() {
			if first {
				first = false
				open = !row.IsOpen()
			}
			setRowOpen(row, open)
		}
	}
	d.table.SyncToModel()
}

func setRowOpen[T gurps.NodeConstraint[T]](row *ntable.Node[T], open bool) {
	row.SetOpen(open)
	for _, child := range row.Children() {
		if child.CanHaveChildren() {
			setRowOpen(child, open)
		}
	}
}

func (d *TableDockable[T]) sizeToFit() {
	d.table.SizeColumnsToFit(true)
	d.table.MarkForRedraw()
}

func (d *TableDockable[T]) searchModified() {
	d.searchIndex = 0
	d.searchResult = nil
	text := strings.ToLower(d.searchField.Text())
	for _, row := range d.table.RootRows() {
		d.search(text, row)
	}
	d.adjustForMatch()
}

func (d *TableDockable[T]) search(text string, row *ntable.Node[T]) {
	if row.Match(text) {
		d.searchResult = append(d.searchResult, row)
	}
	if row.CanHaveChildren() {
		for _, child := range row.Children() {
			d.search(text, child)
		}
	}
}

func (d *TableDockable[T]) previousMatch() {
	if d.searchIndex > 0 {
		d.searchIndex--
		d.adjustForMatch()
	}
}

func (d *TableDockable[T]) nextMatch() {
	if d.searchIndex < len(d.searchResult)-1 {
		d.searchIndex++
		d.adjustForMatch()
	}
}

func (d *TableDockable[T]) adjustForMatch() {
	d.backButton.SetEnabled(d.searchIndex != 0)
	d.forwardButton.SetEnabled(len(d.searchResult) != 0 && d.searchIndex != len(d.searchResult)-1)
	if len(d.searchResult) != 0 {
		d.matchesLabel.Text = fmt.Sprintf(i18n.Text("%d of %d"), d.searchIndex+1, len(d.searchResult))
		row := d.searchResult[d.searchIndex]
		d.table.DiscloseRow(row, false)
		d.table.ClearSelection()
		rowIndex := d.table.RowToIndex(row)
		d.table.SelectByIndex(rowIndex)
		d.table.ScrollRowIntoView(rowIndex)
	} else {
		d.matchesLabel.Text = "-"
	}
	d.matchesLabel.Parent().MarkForLayoutAndRedraw()
}

// Rebuild implements widget.Rebuildable.
func (d *TableDockable[T]) Rebuild(_ bool) {
	h, v := d.scroll.Position()
	sel := d.table.CopySelectionMap()
	d.table.SyncToModel()
	d.table.SetSelectionMap(sel)
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
	d.scroll.SetPosition(h, v)
}

func (d *TableDockable[T]) crc64() uint64 {
	var buffer bytes.Buffer
	rows := d.provider.RootRows()
	data := make([]any, 0, len(rows))
	for _, row := range rows {
		data = append(data, row.Data())
	}
	if err := jio.Save(context.Background(), &buffer, data); err != nil {
		return 0
	}
	return crc.Bytes(0, buffer.Bytes())
}
