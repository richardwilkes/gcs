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
	"cmp"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/slant"
	"github.com/richardwilkes/unison/enums/weight"
)

var (
	_ unison.Dockable       = &SettingsDockable{}
	_ unison.TabCloser      = &SettingsDockable{}
	_ ownWindowMinimumSizer = &SettingsDockable{}
)

// SettingsDockable holds common settings dockable data.
type SettingsDockable struct {
	unison.Panel
	TabTitle   string
	TabIcon    *unison.SVG
	Extensions []string
	Loader     func(fileSystem fs.FS, filePath string) error
	// RefLoader is an alternative to Loader for dockables that need to know where the file lives on disk. When set, it
	// is used in preference to Loader.
	RefLoader func(ref *library.NamedFileRef) error
	// LoadItemTitle is the title of the toolbar menu's item that loads a file chosen in a dialog. When empty, it is
	// Import…, which suits a dockable that takes the file's contents into itself; the file editors, which open the
	// chosen file in an editor of its own, call it Open… instead.
	LoadItemTitle     string
	Saver             func(filePath string) error
	Resetter          func()
	ModifiedCallback  func() bool
	WillCloseCallback func() bool
	// SelectionSaver, when set along with Saver, adds Export Selected… to the toolbar menu, which writes only what is
	// selected, and Export… is then called Export All…. HasSelection must be set with it.
	SelectionSaver func(filePath string) error
	// HasSelection reports whether anything is selected, which Export Selected… needs to be enabled.
	HasSelection func() bool
	// ContentWidthBehavior is how the content's width follows the view's. Left Unmodified, it is Fill: the content is
	// at least as wide as the view. Follow makes it exactly as wide, never wider, so that it doesn't scroll sideways
	// however narrow the view becomes, and it must then fit its content to whatever width it is given. HintedFill gives
	// it the view's width to lay itself out within, for content that wraps or shrinks to fit, but lets it be wider when
	// it cannot.
	ContentWidthBehavior behavior.Enum
	// ContentMinSize is the smallest size the view lets its content be shrunk to, below which the content would be
	// hidden or too cramped to use. The view scrolls rather than shrinking it further.
	ContentMinSize geom.Size
	scroller       *unison.ScrollPanel
}

// settingsSpec is what distinguishes one of the views that edit a global setting in place (the colors, fonts, menu keys,
// general settings, page reference mappings, saved filters, sheet defaults and library settings) from another;
// everything else about them is in SettingsDockable. Their tab icon is always the settings icon, so the spec does not name one. Any of
// the functions but initContent may be nil, in which case the base omits what it would have done with it: with no
// loader, saver or resetter the toolbar has no menu or reset button, and with no willClose the view closes without
// being asked.
type settingsSpec struct {
	title             string
	ext               string
	loader            func(fileSystem fs.FS, filePath string) error
	saver             func(filePath string) error
	resetter          func()
	willClose         func() bool
	addToStartToolbar func(toolbar *unison.Panel)
	initContent       func(content *unison.Panel)
}

// initSettings fills in the base from the spec, with self, the outer view, as what the dock resolves the panel to, then
// builds the toolbar and content and places the view in the dock. The caller must already have checked that the view is
// not open, with ActivateDockable or a predicate of its own, since the base cannot tell one view from another.
func (d *SettingsDockable) initSettings(self unison.Paneler, spec *settingsSpec) {
	d.Self = self
	d.TabTitle = spec.title
	d.TabIcon = svg.Settings
	if spec.ext != "" {
		d.Extensions = []string{spec.ext}
	}
	d.Loader = spec.loader
	d.Saver = spec.saver
	d.Resetter = spec.resetter
	d.WillCloseCallback = spec.willClose
	d.Setup(spec.addToStartToolbar, nil, spec.initContent)
}

// initSettingsContent gives a settings view's content panel the layout most of them share, a grid of the given number
// of columns with the standard spacing, and returns it so that a view keeping a reference to its content can take it
// from the same call.
func initSettingsContent(content *unison.Panel, columns int) *unison.Panel {
	content.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	return content
}

// initSettingsSections gives a settings view's content panel the layout the views built from titled sections share: a
// single column of the sections newSettingsSection adds, with a line of space between them.
func initSettingsSections(content *unison.Panel) {
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.DefaultLabelTheme.Font.LineHeight(),
	})
}

// newSettingsSection adds a section to content: a grid of the given number of columns under a header with the title,
// as wide as the content when it is laid out by initSettingsSections. It returns the grid for the section's rows.
func newSettingsSection(content *unison.Panel, title string, columns int) *unison.Panel {
	section := unison.NewPanel()
	section.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	section.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	addSettingsSectionHeader(section, title, columns)
	content.AddChild(section)
	return section
}

// addSettingsSectionHeader adds a bold title and a separator, each spanning hspan columns of the panel, to head a
// section of a settings view.
func addSettingsSectionHeader(panel *unison.Panel, title string, hspan int) {
	label := unison.NewLabel()
	desc := label.Font.Descriptor()
	desc.Weight = weight.Bold
	label.Font = desc.Font()
	label.SetTitle(title)
	label.SetLayoutData(&unison.FlexLayoutData{HSpan: hspan})
	panel.AddChild(label)
	sep := unison.NewSeparator()
	sep.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  hspan,
		HAlign: align.Fill,
		HGrab:  true,
	})
	panel.AddChild(sep)
}

// newSettingsSubgroup adds a group of related rows to a section, spanning all of its columns: a heading with the
// title, then an indented grid of the given number of columns, which is returned. The heading is static text, which
// nothing reads as naming the grid, so the grid is named with it as well.
func newSettingsSubgroup(section *unison.Panel, title string, columns int) *unison.Panel {
	span := flexColumns(section)
	heading := unison.NewLabel()
	heading.SetTitle(title)
	heading.SetLayoutData(&unison.FlexLayoutData{HSpan: span})
	section.AddChild(heading)
	grid := unison.NewPanel()
	grid.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: unison.StdHSpacing * 2}))
	grid.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	grid.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  span,
		HAlign: align.Fill,
	})
	grid.Accessibility.Name = title
	section.AddChild(grid)
	return grid
}

// flexColumns returns the number of columns the panel's layout has, which is one for anything but a FlexLayout.
func flexColumns(panel *unison.Panel) int {
	if layout, ok := panel.Layout().(*unison.FlexLayout); ok && layout.Columns > 0 {
		return layout.Columns
	}
	return 1
}

// addSettingsRow adds a labeled row to a section: a label with the title, then the field spanning the next hspan
// columns. Any trailing panels follow the field on the same row, in a wrapper that separates the field from its label,
// so the field is pointed at the label explicitly either way. It returns the label.
func addSettingsRow(section *unison.Panel, title string, field unison.Paneler, hspan int, trailing ...unison.Paneler) *unison.Label {
	label := NewFieldLeadingLabel(title, false)
	section.AddChild(label)
	labelControl(field, label)
	if len(trailing) == 0 {
		field.AsPanel().SetLayoutData(&unison.FlexLayoutData{HSpan: hspan})
		section.AddChild(field)
	} else {
		section.AddChild(WrapWithSpan(hspan, append([]unison.Paneler{field}, trailing...)...))
	}
	return label
}

// addStretchedSettingsRow adds a labeled row to a section whose field, spanning hspan columns, takes whatever width
// the section has left.
func addStretchedSettingsRow(section *unison.Panel, title string, field unison.Paneler, hspan int) {
	label := NewFieldLeadingLabel(title, false)
	section.AddChild(label)
	labelControl(field, label)
	field.AsPanel().SetLayoutData(&unison.FlexLayoutData{
		HSpan:  hspan,
		HAlign: align.Fill,
		HGrab:  true,
	})
	section.AddChild(field)
}

// settingsNoteFont returns the font a note in a settings view is set in: the label font, italic and a size smaller.
func settingsNoteFont() unison.Font {
	fd := unison.DefaultLabelTheme.Font.Descriptor()
	fd.Slant = slant.Italic
	fd.Size--
	return fd.Font()
}

// newSettingsContent returns the content panel of a settings view, filled by initContent within the standard margin.
func newSettingsContent(initContent func(*unison.Panel)) *unison.Panel {
	content := unison.NewPanel()
	content.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing * 2)))
	initContent(content)
	return content
}

// Setup the dockable and display it.
func (d *SettingsDockable) Setup(addToStartToolbar, addToEndToolbar, initContent func(*unison.Panel)) {
	d.SetLayout(&unison.FlexLayout{Columns: 1})
	toolbar := d.createToolbar(addToStartToolbar, addToEndToolbar)
	d.AddChild(toolbar)
	content := newSettingsContent(initContent)
	d.scroller = d.newContentScroller(content)
	d.AddChild(d.scroller)
	PlaceInDock(d, dgroup.Settings, false)
	FocusFirstContent(toolbar, content)
}

// newContentScroller returns the scroll panel that shows the view's content, sized as ContentWidthBehavior and
// ContentMinSize say.
func (d *SettingsDockable) newContentScroller(content *unison.Panel) *unison.ScrollPanel {
	scroller := unison.NewScrollPanel()
	widthBehavior := d.ContentWidthBehavior
	if widthBehavior == behavior.Unmodified {
		widthBehavior = behavior.Fill
	}
	scroller.SetContent(content, widthBehavior, behavior.Fill)
	scroller.SetLayoutData(&unison.FlexLayoutData{
		MinSize: d.ContentMinSize,
		HAlign:  align.Fill,
		VAlign:  align.Fill,
		HGrab:   true,
		VGrab:   true,
	})
	return scroller
}

// settingsOwnWindowContentMinHeight is how much of its content's height a settings view keeps in view in a window of
// its own, or all of it when the content is shorter.
const settingsOwnWindowContentMinHeight = 200

// ownWindowMinimumSize implements ownWindowMinimumSizer. The view scrolls its content, so its layout lets it shrink to
// little more than its toolbar. A window of its own is instead kept big enough to show, below the toolbar, the content
// at its narrowest across and settingsOwnWindowContentMinHeight of it down. A view with a ContentMinSize is already
// held to that by its layout.
func (d *SettingsDockable) ownWindowMinimumSize() geom.Size {
	if d.ContentMinSize != (geom.Size{}) {
		return geom.Size{}
	}
	size, _, _ := d.scroller.Content().AsPanel().Sizes(geom.Size{})
	if size.Height > settingsOwnWindowContentMinHeight {
		size.Height = settingsOwnWindowContentMinHeight
		// The content then scrolls vertically, and is not to lose the scroll bar's width from its own.
		size.Width += d.scroller.Bar(false).MinimumThickness
	}
	_, toolbar, _ := d.Children()[0].Sizes(geom.Size{})
	return geom.NewSize(max(toolbar.Width, size.Width), toolbar.Height+size.Height).Ceil()
}

// TitleIcon implements unison.Dockable.
func (d *SettingsDockable) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  d.TabIcon,
		Size: suggestedSize,
	}
}

// Title implements unison.Dockable.
func (d *SettingsDockable) Title() string {
	return d.TabTitle
}

// Tooltip implements unison.Dockable.
func (d *SettingsDockable) Tooltip() string {
	return ""
}

// Modified implements unison.Dockable.
func (d *SettingsDockable) Modified() bool {
	if d.ModifiedCallback == nil {
		return false
	}
	return d.ModifiedCallback()
}

// MarkModified implements ModifiableRoot.
func (d *SettingsDockable) MarkModified(_ unison.Paneler) {
	invalidateDisplayData()
	d.Modified()
	UpdateTitleForDockable(d)
	DeepSync(d)
}

// MayAttemptClose implements unison.TabCloser.
func (d *SettingsDockable) MayAttemptClose() bool {
	return MayAttemptCloseOfGroup(d)
}

// AttemptClose implements unison.TabCloser.
func (d *SettingsDockable) AttemptClose() bool {
	if !CloseGroup(d) {
		return false
	}
	if d.WillCloseCallback != nil {
		if !d.WillCloseCallback() {
			return false
		}
	}
	return AttemptCloseForDockable(d)
}

func (d *SettingsDockable) createToolbar(addToStartToolbar, addToEndToolbar func(*unison.Panel)) *unison.Panel {
	toolbar := NewToolbar()
	if addToStartToolbar != nil {
		addToStartToolbar(toolbar)
	}
	index := len(toolbar.Children())
	if addToEndToolbar != nil {
		addToEndToolbar(toolbar)
	}
	if d.Resetter != nil {
		b := unison.NewSVGButton(svg.Reset)
		b.Tooltip = newWrappedTooltip(i18n.Text("Reset"))
		b.ClickCallback = d.handleReset
		toolbar.AddChild(b)
	}
	if d.canLoad() || d.Saver != nil {
		b := unison.NewSVGButton(svg.Menu)
		b.Tooltip = newWrappedTooltip(i18n.Text("Menu"))
		b.ClickCallback = func() { d.fileMenu().show(b) }
		toolbar.AddChild(b)
	}
	if len(toolbar.Children()) != index {
		spacer := unison.NewPanel()
		spacer.SetLayoutData(&unison.FlexLayoutData{HGrab: true})
		toolbar.AddChildAtIndex(spacer, index)
	}
	FinishToolbarLayout(toolbar)
	return toolbar
}

func (d *SettingsDockable) handleReset() {
	if unison.QuestionDialog(i18n.Text("Are you sure you want to reset the\n%s?", d.TabTitle), "") == unison.ModalResponseOK {
		d.Resetter()
	}
}

// fileMenu returns the toolbar's menu of files to load and save, made from the dockable's fields as they are now.
func (d *SettingsDockable) fileMenu() *settingsFileMenu {
	m := &settingsFileMenu{
		title:          d.TabTitle,
		extensions:     d.Extensions,
		loadItemTitle:  d.LoadItemTitle,
		saver:          d.Saver,
		selectionSaver: d.SelectionSaver,
		hasSelection:   d.HasSelection,
	}
	if d.canLoad() {
		m.load = d.loadRef
	}
	return m
}

// canLoad reports whether the dockable was given either form of loader.
func (d *SettingsDockable) canLoad() bool {
	return d.Loader != nil || d.RefLoader != nil
}

// loadRef hands the whole file reference to RefLoader when one is set, and otherwise its file system and path to
// Loader.
func (d *SettingsDockable) loadRef(ref *library.NamedFileRef) error {
	if d.RefLoader != nil {
		return d.RefLoader(ref)
	}
	return d.Loader(ref.FileSystem, ref.FilePath)
}

// settingsFileMenu is the menu that loads and saves a settings file: Import… (or loadItemTitle), Export… (Export All…
// and Export Selected… when there is a selectionSaver), and the files of the extensions found in the libraries'
// Settings folders. Either of load and saver may be nil, leaving its items out.
type settingsFileMenu struct {
	title          string // what is loaded or saved, for error messages
	extensions     []string
	loadItemTitle  string // "" means Import…
	load           func(ref *library.NamedFileRef) error
	saver          func(filePath string) error
	selectionSaver func(filePath string) error
	hasSelection   func() bool
	canSave        func() bool // nil means always; otherwise Export… is enabled only when it returns true
}

// show pops up the menu below the button.
func (m *settingsFileMenu) show(b *unison.Button) {
	f := unison.DefaultMenuFactory()
	id := unison.ContextMenuIDFlag
	menu := f.NewMenu(id, "", nil)
	id++
	if m.load != nil {
		menu.InsertItem(-1, f.NewItem(id, cmp.Or(m.loadItemTitle, i18n.Text("Import…")), unison.KeyBinding{}, nil,
			m.handleImport))
		id++
	}
	if m.saver != nil {
		title := i18n.Text("Export…")
		if m.selectionSaver != nil {
			title = i18n.Text("Export All…")
		}
		var canSave func(unison.MenuItem) bool
		if m.canSave != nil {
			canSave = func(_ unison.MenuItem) bool { return m.canSave() }
		}
		menu.InsertItem(-1, f.NewItem(id, title, unison.KeyBinding{}, canSave,
			func(_ unison.MenuItem) { m.export(m.saver) }))
		id++
		if m.selectionSaver != nil {
			menu.InsertItem(-1, f.NewItem(id, i18n.Text("Export Selected…"), unison.KeyBinding{},
				func(_ unison.MenuItem) bool { return m.hasSelection() },
				func(_ unison.MenuItem) { m.export(m.selectionSaver) }))
			id++
		}
	}
	if m.load != nil {
		libraries := gurps.GlobalSettings().Libraries
		sets := library.ScanForNamedFileSets(nil, "", false, libraries, m.extensions...)
		if len(sets) != 0 {
			menu.InsertSeparator(-1, false)
			for _, lib := range sets {
				menu.InsertItem(-1, f.NewItem(id, lib.Name, unison.KeyBinding{},
					func(_ unison.MenuItem) bool { return false }, nil))
				id++
				for _, one := range lib.List {
					m.insertFileToLoad(menu, id, one)
					id++
				}
			}
		}
	}
	menu.Popup(b.RectToRoot(b.ContentRect(true)), 0)
}

func (m *settingsFileMenu) insertFileToLoad(menu unison.Menu, id int, ref *library.NamedFileRef) {
	menu.InsertItem(-1, menu.Factory().NewItem(id, "    "+ref.Name, unison.KeyBinding{}, nil, func(_ unison.MenuItem) {
		m.doLoad(ref)
	}))
}

// doLoad loads the file, reporting a failure through the workspace's error handler.
func (m *settingsFileMenu) doLoad(ref *library.NamedFileRef) {
	if err := m.load(ref); err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to load ")+m.title, err)
	}
}

func (m *settingsFileMenu) handleImport(_ unison.MenuItem) {
	if filePath, ok := chooseFileToOpen(gurps.SettingsLastDirKey, m.extensions...); ok {
		m.doLoad(diskFileRef(filePath))
	}
}

func (m *settingsFileMenu) export(saver func(filePath string) error) {
	// The title is not a path, so it is offered whole; chooseFileToSave makes it safe for the file system.
	if filePath, ok := chooseFileToSave(gurps.GlobalSettings().LastDir(gurps.SettingsLastDirKey),
		m.title, m.extensions[0], gurps.SettingsLastDirKey); ok {
		if err := saver(filePath); err != nil {
			Workspace.ErrorHandler(i18n.Text("Unable to save ")+m.title, err)
		}
	}
}
