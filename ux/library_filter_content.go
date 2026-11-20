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
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

// libraryFileColumnID is the ID of the File column a library filter view adds after the columns of its list type. It
// is negative so that it can't be taken for one of theirs (see extraColumnModel); -1 is the slot of the time column
// widths were saved (see saveColumnWidths) and -10 is gurps.PageRefCellAlias.
const libraryFileColumnID = -20

// maxSourceFilesWithoutAsking is the most files Show in Source File opens without asking first.
const maxSourceFilesWithoutAsking = 10

// Seams so that tests can exercise Show in Source File without a modal dialog.
var (
	// confirmShowInSourceFiles asks whether Show in Source File may open the given number of files.
	confirmShowInSourceFiles = func(count int) bool {
		return unison.QuestionDialog(i18n.Text("Are you sure you want to open all of these?"),
			i18n.Text("%d files will be opened.", count)) == unison.ModalResponseOK
	}
	// warnRowsNotInSourceFiles tells the user that Show in Source File didn't find the given number of the selected rows
	// in the lists it opened.
	warnRowsNotInSourceFiles = func(count int) {
		detail := i18n.Text("1 item was not found.")
		if count != 1 {
			detail = i18n.Text("%d items were not found.", count)
		}
		unison.WarningDialogWithMessage(i18n.Text("Some items could not be found in their files"), detail)
	}
)

// libraryFilterKind is a list type a library filter view can show: the key and title its saved filters go by (see
// listFilterKind), the extension of its files, its icon, and how to build a view's content for it.
type libraryFilterKind struct {
	listFilterKind
	ext        string
	icon       *unison.SVG
	newContent func(d *LibraryFilterDockable) libraryFilterView
}

// libraryFilterKinds returns the list types a library filter view can show, which are those that can have saved
// filters, with the same titles and in the same order (see listFilterKinds).
func libraryFilterKinds() []libraryFilterKind {
	bases := listFilterKinds()
	kinds := make([]libraryFilterKind, 0, len(bases))
	for _, base := range bases {
		switch base.key {
		case gurps.ListFilterKeyForExtension(gurps.TraitsExt):
			kinds = append(kinds, newLibraryFilterKind(base, gurps.TraitsExt, svg.GCSTraits, gurps.NewTraitsFromFile,
				func() TableProvider[*gurps.Trait] { return NewTraitsProvider(&traitListProvider{}, false) }, nil))
		case gurps.ListFilterKeyForExtension(gurps.TraitModifiersExt):
			kinds = append(kinds, newLibraryFilterKind(base, gurps.TraitModifiersExt, svg.GCSTraitModifiers,
				gurps.NewTraitModifiersFromFile,
				func() TableProvider[*gurps.TraitModifier] {
					return NewTraitModifiersProvider(&traitModifierListProvider{}, false)
				},
				func(target unison.Paneler, table *unison.Table[*Node[*gurps.TraitModifier]]) {
					installApplyModifierHandler(target, table, traitModifierTargetKind())
				}))
		case gurps.ListFilterKeyForExtension(gurps.SkillsExt):
			kinds = append(kinds, newLibraryFilterKind(base, gurps.SkillsExt, svg.GCSSkills, gurps.NewSkillsFromFile,
				func() TableProvider[*gurps.Skill] { return NewSkillsProvider(&skillListProvider{}, false) }, nil))
		case gurps.ListFilterKeyForExtension(gurps.SpellsExt):
			kinds = append(kinds, newLibraryFilterKind(base, gurps.SpellsExt, svg.GCSSpells, gurps.NewSpellsFromFile,
				func() TableProvider[*gurps.Spell] { return NewSpellsProvider(&spellListProvider{}, false) }, nil))
		case gurps.ListFilterKeyForExtension(gurps.EquipmentExt):
			kinds = append(kinds, newLibraryFilterKind(base, gurps.EquipmentExt, svg.GCSEquipment,
				gurps.NewEquipmentFromFile,
				func() TableProvider[*gurps.Equipment] {
					return NewEquipmentProvider(&equipmentListProvider{}, false, false)
				},
				installBuyEquipmentHandler))
		case gurps.ListFilterKeyForExtension(gurps.EquipmentModifiersExt):
			kinds = append(kinds, newLibraryFilterKind(base, gurps.EquipmentModifiersExt, svg.GCSEquipmentModifiers,
				gurps.NewEquipmentModifiersFromFile,
				func() TableProvider[*gurps.EquipmentModifier] {
					return NewEquipmentModifiersProvider(&equipmentModifierListProvider{}, false)
				},
				func(target unison.Paneler, table *unison.Table[*Node[*gurps.EquipmentModifier]]) {
					installApplyModifierHandler(target, table, equipmentModifierTargetKind())
				}))
		case gurps.ListFilterKeyForExtension(gurps.NotesExt):
			kinds = append(kinds, newLibraryFilterKind(base, gurps.NotesExt, svg.GCSNotes, gurps.NewNotesFromFile,
				func() TableProvider[*gurps.Note] { return NewNotesProvider(&noteListProvider{}, false) }, nil))
		}
	}
	return kinds
}

// libraryFilterKindFor returns the list type with the key, or false if there is none.
func libraryFilterKindFor(key string) (libraryFilterKind, bool) {
	for _, kind := range libraryFilterKinds() {
		if kind.key == key {
			return kind, true
		}
	}
	return libraryFilterKind{}, false
}

// newLibraryFilterKind returns the list type of the files with the extension, whose saved filters go by base. load
// reads such a file, newProvider returns a provider for a list of them, and installExtra, when set, installs on the
// content the commands the type's rows offer beyond those every list offers (see installStandardTableCmdHandlers).
func newLibraryFilterKind[T gurps.Node[T]](base listFilterKind, ext string, icon *unison.SVG,
	load func(fileSystem fs.FS, filePath string) ([]T, error), newProvider func() TableProvider[T],
	installExtra func(target unison.Paneler, table *unison.Table[*Node[T]]),
) libraryFilterKind {
	return libraryFilterKind{
		listFilterKind: base,
		ext:            ext,
		icon:           icon,
		newContent: func(d *LibraryFilterDockable) libraryFilterView {
			return newLibraryFilterContent(d, base.key, icon, load, newProvider(), installExtra)
		},
	}
}

// libraryFilterSetTitle returns what a set of the list type is called: its name, or for the set of every file and for
// an unsaved set, what they hold.
func libraryFilterSetTitle(kind libraryFilterKind, set *gurps.LibrarySourceSet, adHoc bool) string {
	switch {
	case set == nil:
		return i18n.Text("All %s", kind.title)
	case adHoc:
		return i18n.Text("Selected %s", kind.title)
	default:
		return set.Name
	}
}

// libraryFilterView is the content of a library filter view, which shows the rows of one list type: what the
// LibraryFilterDockable holding it needs of it, whatever the type is. Every method is called on the UI thread.
type libraryFilterView interface {
	unison.Paneler
	// load reads the files in the background, then shows their rows in place of those it shows now. A file that hasn't
	// changed since it was last read isn't read again, and when none of the files shown has, and no other file is to be
	// shown, the rows shown are left as they are.
	load(files []libraryFilterFile)
	// dropCache forgets what the files read so far held, so that the next load reads every file afresh.
	dropCache()
	// hasInstalled returns true once the rows of a load are shown.
	hasInstalled() bool
	// rebuild builds the rows anew, keeping the filters in force, the selection and the scroll position.
	rebuild()
	sheetSettingsUpdated(entity *gurps.Entity, parts gurps.SheetSettingsParts, fullRebuild bool)
	listFiltersChanged(key string, source *listFilterPopup)
	focusFilter()
	quickFilterText() string
	setQuickFilterText(text string)
	// syncSetPopup brings the source set popup into line with the set the view shows and the saved sets.
	syncSetPopup()
	sourceSetPopup() *librarySourceSetPopup
	// updateStatus brings the footer into line with what the view shows.
	updateStatus()
	// dispose discards the load in flight, if any, and saves the column widths, for when the content goes away.
	dispose()
}

// libraryFileProvider decorates the provider of a library filter view's list type with the File column, which names
// the file each row came from, and keeps the rows it hands the table from one call for them to the next.
type libraryFileProvider[T gurps.Node[T]] struct {
	TableProvider[T]
	table    *unison.Table[*Node[T]]
	files    []libraryFilterFile
	rowFiles map[T]int  // Every row, at any depth, to the index of its file in files
	roots    []*Node[T] // The rows RootRows last returned
}

// SetTable implements TableProvider.
func (p *libraryFileProvider[T]) SetTable(table *unison.Table[*Node[T]]) {
	p.TableProvider.SetTable(table)
	p.table = table
	p.roots = nil
}

// RootRows implements unison.TableModel. Each item keeps the row it was last given until a load replaces the items,
// where the wrapped provider builds every row anew on every call: the table asks each time it is filtered, and rows
// built anew derive the data of every cell, and build the panels of those shown, over again.
// Nothing but a load changes a row of the view, and a change to how rows are displayed moves the display epoch on,
// which the rows kept go by (see cellStamp).
func (p *libraryFileProvider[T]) RootRows() []*Node[T] {
	p.roots = nodesFor(p.table, nil, p.RootData(), p.roots, false, nil)
	return p.roots
}

// SetRootData implements TableProvider. No row is kept, even for an item that stays, since what the File column shows
// for it turns on the other files shown.
func (p *libraryFileProvider[T]) SetRootData(data []T) {
	p.TableProvider.SetRootData(data)
	p.roots = nil
}

// ColumnIDs implements TableProvider.
func (p *libraryFileProvider[T]) ColumnIDs() []int {
	return append(p.TableProvider.ColumnIDs(), libraryFileColumnID)
}

// Headers implements TableProvider.
func (p *libraryFileProvider[T]) Headers() []unison.TableColumnHeader[*Node[T]] {
	return append(p.TableProvider.Headers(), headerFromData[T](gurps.HeaderData{
		Title:  i18n.Text("File"),
		Detail: i18n.Text("The library file the row is from"),
	}, false))
}

// ContextMenuItems implements TableProvider. Show in Source File follows Open Detail Editor.
func (p *libraryFileProvider[T]) ContextMenuItems() []ContextMenuItem {
	items := p.TableProvider.ContextMenuItems()
	if i := slices.IndexFunc(items, func(item ContextMenuItem) bool { return item.ID == OpenEditorItemID }); i != -1 {
		return slices.Insert(items, i+1, contextMenuItemFor(showInSourceFileAction))
	}
	return items
}

// extraCellData implements extraColumnModel. The tooltip names the file's library and its path within it, then its
// path on disk. Every row has one, so it is plain text, since a sync builds a cell for every row and a Markdown tooltip
// for each would take as long again as the rest of the sync.
func (p *libraryFileProvider[T]) extraCellData(columnID int, row T, data *gurps.CellData) bool {
	if columnID != libraryFileColumnID {
		return false
	}
	data.Type = cell.Text
	if i, ok := p.rowFiles[row]; ok {
		f := &p.files[i]
		data.Primary = f.display
		data.Tooltip = f.tooltip
		data.PlainTooltip = true
	}
	return true
}

// libraryFilterParse is what reading one of a library filter view's files yielded. A file that is gone yields neither
// rows nor an error. Once a parse has been handed to the UI thread, the loads that follow read only modTime, size and
// statted on their own goroutines; the rest is the UI thread's.
type libraryFilterParse[T gurps.Node[T]] struct {
	modTime time.Time
	err     error
	rows    []T
	// inFile maps the IDs given to rows to keep them apart from the rows before them (see gurps.MakeRowIDsUnique) to
	// the IDs they have in the file.
	inFile map[tid.TID]tid.TID
	size   int64
	// statted is true when the modification time and size were read, which makes the parse fit to be reused while the
	// file stays as it is, even when reading the file failed.
	statted bool
}

// libraryFilterFailure is a file a library filter view couldn't read.
type libraryFilterFailure struct {
	file libraryFilterFile
	err  error
}

// libraryFilterLoad is one reading of a library filter view's files (see libraryFilterContent.load).
type libraryFilterLoad[T gurps.Node[T]] struct {
	// prior is the load this one superseded while it was in flight. This one waits for it to stop, then takes over what
	// it knew (see read).
	prior *libraryFilterLoad[T]
	// known holds the latest parse of each file: when the load begins, the view's cache, and once it has read its
	// files, those it read too. It belongs to the load's goroutine, and once done is closed, to the load that
	// superseded it, if any.
	known map[libraryFilterCacheKey]*libraryFilterParse[T]
	done  chan struct{} // Closed once the load has read its files and handed them to the UI thread
	gen   int64         // The view's count of loads when this one began
}

// read reads the files as parseLibraryFilterFiles does, reusing the parses in known, and adds the parses it made to
// known. When the load superseded another, it first waits for that one to stop and takes over what it knew, rather than
// read those files again. It runs on a background goroutine.
func (l *libraryFilterLoad[T]) read(files []libraryFilterFile, parse func(file libraryFilterFile) ([]T, error),
	canceled func() bool,
) []*libraryFilterParse[T] {
	if l.prior != nil {
		// Superseding the prior load canceled it, and it stops at the next file boundary, so this wait is brief. What it
		// knew includes the cache as it was when this load began, since no load was shown in between.
		<-l.prior.done
		l.known = l.prior.known
		l.prior = nil
	}
	parsed := parseLibraryFilterFiles(files, l.known, parse, canceled)
	for i, p := range parsed {
		if p != nil && p.statted {
			l.known[files[i].cacheKey()] = p
		}
	}
	return parsed
}

// libraryFilterContent is the content of a library filter view showing rows of type T: its toolbar, the notice that it
// is read-only, its table and the footer beneath it reporting on what it shows. It filters its rows as a library list
// does (see listFilterControls).
type libraryFilterContent[T gurps.Node[T]] struct {
	unison.Panel
	listFilterControls[T]
	shell     *LibraryFilterDockable
	decorator *libraryFileProvider[T]
	scroll    *unison.ScrollPanel
	header    *unison.TableHeader[*Node[T]]
	setPopup  *librarySourceSetPopup
	footer    *unison.Label
	// parse reads a file's rows. It runs on a background goroutine, and is a field so that tests can count the reads,
	// fail them or hold them back.
	parse func(file libraryFilterFile) ([]T, error)
	// cache holds the parses of the files shown that could be checked for changes (see libraryFilterParse.statted).
	cache map[libraryFilterCacheKey]*libraryFilterParse[T]
	// lastLoad is the load begun last, while it is yet to be shown, for the next load to take over from.
	lastLoad   *libraryFilterLoad[T]
	inFile     []map[tid.TID]tid.TID // Parallel to decorator.files; see libraryFilterParse.inFile
	unreadable []libraryFilterFailure
	widthsKey  string
	gen        atomic.Int64 // Counts the loads, so that one that has been superseded is discarded
	rowCount   int
	matched    int
	// savedWidths is true when the column widths were saved before, which keeps the first load from sizing the
	// columns to fit.
	savedWidths bool
	installed   bool
	loading     bool
	filtered    bool
	closed      bool
}

// newLibraryFilterContent builds the content of the view d for the list type with the key and icon, whose files load
// reads, and adds it to d. See newLibraryFilterKind for the rest.
func newLibraryFilterContent[T gurps.Node[T]](d *LibraryFilterDockable, key string, icon *unison.SVG,
	load func(fileSystem fs.FS, filePath string) ([]T, error), provider TableProvider[T],
	installExtra func(target unison.Paneler, table *unison.Table[*Node[T]]),
) *libraryFilterContent[T] {
	c := &libraryFilterContent[T]{
		shell:     d,
		decorator: &libraryFileProvider[T]{TableProvider: provider},
		scroll:    unison.NewScrollPanel(),
		footer:    unison.NewLabel(),
		cache:     make(map[libraryFilterCacheKey]*libraryFilterParse[T]),
		widthsKey: "library_filter/" + key,
		parse: func(file libraryFilterFile) ([]T, error) {
			return load(os.DirFS(filepath.Dir(file.fullPath)), filepath.Base(file.fullPath))
		},
	}
	c.Self = c
	var table *unison.Table[*Node[T]]
	c.header, table = NewNodeTable[T](c.decorator, nil)
	c.initListFilterControls(table, c.decorator,
		func(row *Node[T], text string) bool {
			// A word of a file's name would otherwise match every row of the file.
			return row.matchColumns(text, func(columnID int) bool { return columnID == libraryFileColumnID })
		},
		func(filtered bool, matched int) {
			c.filtered = filtered
			c.matched = matched
			c.updateStatus()
		})
	c.SetLayout(&unison.FlexLayout{Columns: 1})
	c.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Fill, HGrab: true, VGrab: true})

	c.table.SyncToModel()
	c.table.SizeColumnsToFit(true)
	c.savedWidths = applySavedColumnWidths(c.table, c.widthsKey)
	c.scroll.SetColumnHeader(c.header)
	c.scroll.SetContent(c.table, behavior.Fill, behavior.Fill)
	c.scroll.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Fill, HGrab: true, VGrab: true})

	c.footer.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.Insets{Top: 1}, false), unison.NewEmptyBorder(unison.StdInsets())))
	c.footer.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})

	// The scale field installs its commands on the dockable that holds the scroll panel once the scroll panel has been
	// added, so the content is added to the dockable first.
	d.AddChild(c)
	c.AddChild(c.createToolbar(icon))
	c.AddChild(newLibraryFilterNotice())
	c.AddChild(c.scroll)
	c.AddChild(c.footer)

	installStandardTableCmdHandlers(c, c.table, c.decorator, func() Rebuildable { return d }, false)
	if installExtra != nil {
		installExtra(c, c.table)
	}
	c.InstallCmdHandlers(JumpToSearchFilterItemID,
		func(any) bool { return !c.filterField.Focused() },
		func(any) { c.filterField.RequestFocus() })
	c.InstallCmdHandlers(ShowInSourceFileItemID,
		func(any) bool { return c.table.HasSelection() },
		func(any) { c.showInSourceFiles() })
	c.updateStatus()
	return c
}

func (c *libraryFilterContent[T]) createToolbar(icon *unison.SVG) *unison.Panel {
	d := c.shell
	toolbar := NewToolbar()
	toolbar.AddChild(NewDefaultInfoPop())
	addHelpButton(toolbar, "md:User%20Guide/Library%20Filter")
	AddUIScaleField(toolbar, func() int { return gurps.GlobalSettings().General.InitialListUIScale },
		func() int { return d.scale }, func(scale int) { d.scale = scale }, false, c.scroll)

	c.setPopup = newLibrarySourceSetPopup(d, icon)
	toolbar.AddChild(c.setPopup.popup)

	reloadButton := unison.NewSVGButton(svg.Refresh)
	reloadButton.Tooltip = newWrappedTooltip(i18n.Text("Reads the library files again"))
	reloadButton.ClickCallback = d.reload
	toolbar.AddChild(reloadButton)

	sizeToFitButton := unison.NewSVGButton(svg.SizeToFit)
	sizeToFitButton.Tooltip = newWrappedTooltip(i18n.Text("Sets the width of each column to fit its contents"))
	sizeToFitButton.ClickCallback = func() {
		c.table.SizeColumnsToFit(true)
		c.table.MarkForRedraw()
	}
	c.addListFilterControls(toolbar, sizeToFitButton)
	c.finishListFilterToolbar(toolbar, []*filterPopupButton{c.setPopup.button}, c.setPopup.syncTooltip)
	return toolbar
}

// newLibraryFilterNotice returns the line beneath the toolbar that says the view is read-only and names the command
// that leads to where an item can be changed. Its minimum width is zero, so that it never makes the view wider than
// its toolbar and table would on their own; when narrower than its text, it clips it.
func newLibraryFilterNotice() *unison.Label {
	notice := unison.NewLabel()
	notice.SetTitle(i18n.Text("Read-only view. Use Show in Source File to change an item."))
	notice.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.Insets{Bottom: 1}, false), unison.NewEmptyBorder(unison.StdInsets())))
	notice.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	notice.SetSizer(func(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
		minSize, prefSize, maxSize = notice.DefaultSizes(hint)
		minSize.Width = 0
		return minSize, prefSize, maxSize
	})
	return notice
}

// load implements libraryFilterView. A file whose modification time and size are as they were when it was last read,
// by this load or one it superseded, isn't read again.
func (c *libraryFilterContent[T]) load(files []libraryFilterFile) {
	l := c.startLoad()
	parse := c.parse
	c.shell.loads.Go(func() {
		// Closed after the install is posted, so that it runs ahead of that of the load that superseded this one.
		defer close(l.done)
		parsed := l.read(files, parse, func() bool { return c.gen.Load() != l.gen })
		unison.InvokeTask(func() { c.install(l, files, parsed) })
	})
}

// startLoad records that a load has begun, superseding any in flight, and returns it.
func (c *libraryFilterContent[T]) startLoad() *libraryFilterLoad[T] {
	c.loading = true
	c.updateStatus()
	l := &libraryFilterLoad[T]{prior: c.lastLoad, done: make(chan struct{}), gen: c.gen.Add(1)}
	if l.prior == nil {
		l.known = make(map[libraryFilterCacheKey]*libraryFilterParse[T], len(c.cache))
		maps.Copy(l.known, c.cache)
	}
	c.lastLoad = l
	return l
}

// parseLibraryFilterFiles reads the files on as many goroutines as there are CPUs, returning what each yielded, in the
// order given. A file whose modification time and size are those of its parse in cached yields that parse. The rows
// of each file read are given their library source (see gurps.StampLibrarySource). A file that can't be read, even
// because reading it panics, yields the error, which is logged, unless the file is gone. canceled is polled between
// files, so that a load that has been superseded stops early, leaving nil for the files not reached.
func parseLibraryFilterFiles[T gurps.Node[T]](files []libraryFilterFile,
	cached map[libraryFilterCacheKey]*libraryFilterParse[T], parse func(file libraryFilterFile) ([]T, error),
	canceled func() bool,
) []*libraryFilterParse[T] {
	parsed := make([]*libraryFilterParse[T], len(files))
	var next atomic.Int64
	var wg sync.WaitGroup
	for range min(runtime.NumCPU(), len(files)) {
		wg.Go(func() {
			for {
				i := int(next.Add(1) - 1)
				if i >= len(files) || canceled() {
					return
				}
				parsed[i] = parseLibraryFilterFile(&files[i], cached[files[i].cacheKey()], parse)
			}
		})
	}
	wg.Wait()
	return parsed
}

// parseLibraryFilterFile is what parseLibraryFilterFiles does for each file.
func parseLibraryFilterFile[T gurps.Node[T]](file *libraryFilterFile, previous *libraryFilterParse[T],
	parse func(file libraryFilterFile) ([]T, error),
) *libraryFilterParse[T] {
	fi, err := os.Stat(file.fullPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &libraryFilterParse[T]{} // Removed since it was listed; the next listing drops it
		}
		errs.Log(errs.NewWithCause("unable to read library file", err), "path", file.fullPath)
		return &libraryFilterParse[T]{err: err}
	}
	if previous != nil && previous.statted && previous.size == fi.Size() && previous.modTime.Equal(fi.ModTime()) {
		return previous
	}
	result := &libraryFilterParse[T]{modTime: fi.ModTime(), size: fi.Size(), statted: true}
	xos.SafeCall(func() {
		if result.rows, result.err = parse(*file); result.err == nil {
			gurps.StampLibrarySource(file.LibraryFile, result.rows)
		}
	}, func(panicErr error) { result.err = panicErr })
	if result.err != nil {
		if errors.Is(result.err, fs.ErrNotExist) {
			return &libraryFilterParse[T]{}
		}
		result.rows = nil
		errs.Log(errs.NewWithCause("unable to read library file", result.err), "path", file.fullPath)
	}
	return result
}

// install shows the rows of the files, which the load l read, in file order, unless the view has gone away or another
// load has begun since this one. When the files are those shown and none of them was read again, the rows shown are
// left as they are. A file that couldn't be read contributes no rows and is reported in the footer. Rows that share an
// ID with a row of a file before theirs are given new ones, since the table keys its filter and selection by ID. The
// filters in force are applied to the new rows, or the sort, if any, when none is; the selection and scroll position
// are kept, a selected row given a new ID staying selected; and the editors open on rows that are gone, even where a
// file read again has a row with the same ID, are closed. The first rows shown size the columns, unless their widths
// were saved before, to fit every row, whatever the filters in force let through.
func (c *libraryFilterContent[T]) install(l *libraryFilterLoad[T], files []libraryFilterFile,
	parsed []*libraryFilterParse[T],
) {
	if c.closed || l.gen != c.gen.Load() {
		return
	}
	c.lastLoad = nil // What it knew of the files shown is about to be in the cache
	c.loading = false
	if c.installed && slices.Equal(files, c.decorator.files) && c.showsParses(files, parsed) {
		c.updateStatus()
		return
	}
	sel := c.table.CopySelectionMap()
	moved := make(map[tid.TID]tid.TID) // The selected rows given new IDs, old to new
	cache := make(map[libraryFilterCacheKey]*libraryFilterParse[T], len(files))
	inFile := make([]map[tid.TID]tid.TID, len(files))
	rowFiles := make(map[T]int)
	seen := make(map[tid.TID]bool)
	var roots []T
	c.unreadable = nil
	for i, file := range files {
		p := parsed[i]
		key := file.cacheKey()
		shown := c.cache[key] == p
		if p.statted {
			cache[key] = p
		}
		if p.err != nil {
			c.unreadable = append(c.unreadable, libraryFilterFailure{file: file, err: p.err})
			continue
		}
		replaced := gurps.MakeRowIDsUnique(file.LibraryFile, seen, p.rows)
		if shown {
			// The rows are those shown, so a selected one given a new ID takes the selection with it. The rows of a file
			// read again are new, and take the selection of the rows shown with their IDs.
			for newID, oldID := range replaced {
				if sel[oldID] {
					moved[oldID] = newID
				}
			}
		}
		p.inFile = composeInFileIDs(p.inFile, replaced)
		inFile[i] = p.inFile
		roots = append(roots, p.rows...)
		gurps.Traverse(func(row T) bool {
			rowFiles[row] = i
			return false
		}, false, false, p.rows...)
	}
	for oldID := range moved {
		delete(sel, oldID)
	}
	for _, newID := range moved {
		sel[newID] = true
	}
	c.cache = cache
	c.inFile = inFile
	c.decorator.files = files
	c.decorator.rowFiles = rowFiles
	c.rowCount = len(rowFiles)

	h, v := c.scroll.Position()
	c.provider.SetRootData(roots)
	// A row read again has the ID it had, so the heights the table keeps by ID from one filter to the next are dropped.
	c.table.DiscardRowHeights()
	c.dropUnofferedTags()
	sizeColumns := !c.installed && !c.savedWidths
	if sizeColumns && c.table.IsFiltered() {
		// A filter, such as the quick filter's text carried over from another list type, may let few rows through, or
		// none, which would leave each column as narrow as its header.
		c.table.ApplyHierarchicalFilter(nil)
		c.table.SizeColumnsToFit(true)
		sizeColumns = false
	}
	// Applying the filter syncs the table and keeps the sort, so the sort is applied here only when it didn't.
	if !c.applyFilter() {
		if c.header.HasSort() {
			c.header.ApplySort()
		} else {
			c.table.SyncToModel()
		}
	}
	c.table.SetSelectionMap(sel)
	if sizeColumns {
		c.table.SizeColumnsToFit(true)
	}
	c.installed = true
	c.scroll.SetPosition(h, v)
	c.closeEditorsOfGoneRows()
	c.updateStatus()
	c.shell.contentInstalled()
}

// showsParses returns true if each of the files, which are those shown, has the parse shown for it in parsed.
func (c *libraryFilterContent[T]) showsParses(files []libraryFilterFile, parsed []*libraryFilterParse[T]) bool {
	for i, p := range parsed {
		if c.cache[files[i].cacheKey()] != p {
			return false
		}
	}
	return true
}

// composeInFileIDs returns previous, which maps the IDs a file's rows were given to the IDs they have in the file,
// updated for the rows replaced has since given new IDs, mapping their new IDs to their old ones.
func composeInFileIDs(previous, replaced map[tid.TID]tid.TID) map[tid.TID]tid.TID {
	if len(replaced) == 0 {
		return previous
	}
	next := maps.Clone(previous)
	if next == nil {
		next = make(map[tid.TID]tid.TID, len(replaced))
	}
	for _, oldID := range replaced {
		delete(next, oldID)
	}
	for newID, oldID := range replaced {
		original := oldID
		if earlier, ok := previous[oldID]; ok {
			original = earlier
		}
		if newID != original {
			next[newID] = original
		}
	}
	return next
}

// closeEditorsOfGoneRows closes the editors opened from the view on rows it no longer shows. A row of a file read again
// is a new object, whose editor would otherwise go on showing the old one; a file that was not read again keeps its
// rows, and they their editors.
func (c *libraryFilterContent[T]) closeEditorsOfGoneRows() {
	var gone []GroupedCloser
	traverseGroup(c.shell, func(target GroupedCloser) bool {
		if e, ok := target.(interface{ editTarget() any }); ok {
			if row, isRow := e.editTarget().(T); isRow {
				if _, shown := c.decorator.rowFiles[row]; !shown {
					gone = append(gone, target)
				}
			}
		}
		return false
	})
	for _, one := range gone {
		if one.MayAttemptClose() {
			one.AttemptClose()
		}
	}
}

// fileOf returns the file the row came from, or false for a row the view doesn't show.
func (c *libraryFilterContent[T]) fileOf(row T) (libraryFilterFile, bool) {
	i, ok := c.decorator.rowFiles[row]
	if !ok {
		return libraryFilterFile{}, false
	}
	return c.decorator.files[i], true
}

// inFileID returns the ID the row has in the file it came from, which is its own unless it was given a new one to keep
// it apart from a row of another file (see gurps.MakeRowIDsUnique).
func (c *libraryFilterContent[T]) inFileID(row T) tid.TID {
	id := row.ID()
	if i, ok := c.decorator.rowFiles[row]; ok {
		if original, replaced := c.inFile[i][id]; replaced {
			return original
		}
	}
	return id
}

// showInSourceFiles opens the list file each selected row came from and selects there the rows it came from (see
// TableDockable.selectRowsByID). The file of the first selected row is shown last, so that it ends up in front with the
// focus. More than maxSourceFilesWithoutAsking files are opened only once the user has confirmed it. The user is told
// how many rows the lists opened don't hold, as when a file has changed since the view read it, or has rows without
// IDs of their own, which are given new ones each time the file is read.
func (c *libraryFilterContent[T]) showInSourceFiles() {
	type source struct {
		ids      map[tid.TID]bool
		fullPath string
	}
	var sources []*source
	byPath := make(map[string]*source)
	for _, row := range c.table.SelectedRows(false) {
		data := row.Data()
		file, ok := c.fileOf(data)
		if !ok {
			continue
		}
		// A library rooted within another's folder lists the file a second time, which opens as the same list.
		one := byPath[file.fullPath]
		if one == nil {
			one = &source{ids: make(map[tid.TID]bool), fullPath: file.fullPath}
			byPath[file.fullPath] = one
			sources = append(sources, one)
		}
		one.ids[c.inFileID(data)] = true
	}
	if len(sources) == 0 ||
		(len(sources) > maxSourceFilesWithoutAsking && !confirmShowInSourceFiles(len(sources))) {
		return
	}
	notFound := 0
	for i, one := range slices.Concat(sources[1:], sources[:1]) {
		d, _ := OpenFile(one.fullPath, gurps.PageInfo{})
		if xreflect.IsNil(d) {
			continue // Opening the file has reported why it couldn't
		}
		if list, ok := d.(*TableDockable[T]); ok {
			notFound += len(one.ids) - list.selectRowsByID(one.ids)
		}
		if i == len(sources)-1 && d.AsPanel().Window() != nil {
			// The list may be in a window of its own, which selecting its rows doesn't bring to the front.
			ActivateDockable(d)
		}
	}
	if notFound != 0 {
		warnRowsNotInSourceFiles(notFound)
	}
}

// dropCache implements libraryFilterView. The load in flight, if any, is no longer taken over by the next one either.
func (c *libraryFilterContent[T]) dropCache() {
	c.cache = make(map[libraryFilterCacheKey]*libraryFilterParse[T])
	c.lastLoad = nil
}

// hasInstalled implements libraryFilterView.
func (c *libraryFilterContent[T]) hasInstalled() bool {
	return c.installed
}

// rebuild implements libraryFilterView.
func (c *libraryFilterContent[T]) rebuild() {
	c.rebuildKeepingView(c.scroll)
}

// sheetSettingsUpdated implements libraryFilterView.
func (c *libraryFilterContent[T]) sheetSettingsUpdated(entity *gurps.Entity, parts gurps.SheetSettingsParts,
	fullRebuild bool,
) {
	listSheetSettingsUpdated(c.table, func(bool) { c.rebuild() }, entity, parts, fullRebuild)
}

// focusFilter implements libraryFilterView.
func (c *libraryFilterContent[T]) focusFilter() {
	c.filterField.RequestFocus()
}

// quickFilterText implements libraryFilterView.
func (c *libraryFilterContent[T]) quickFilterText() string {
	return c.filterField.Text()
}

// setQuickFilterText implements libraryFilterView.
func (c *libraryFilterContent[T]) setQuickFilterText(text string) {
	if text != c.filterField.Text() {
		c.filterField.SetText(text) // Its modified callback applies the filter.
	}
}

// syncSetPopup implements libraryFilterView.
func (c *libraryFilterContent[T]) syncSetPopup() {
	c.setPopup.rebuildItems()
}

// sourceSetPopup implements libraryFilterView.
func (c *libraryFilterContent[T]) sourceSetPopup() *librarySourceSetPopup {
	return c.setPopup
}

// dispose implements libraryFilterView. The column widths are saved only once rows have been shown, since until then
// they fit the headers alone. The rows and what was read are let go of, as are the rows being dragged from the table,
// since a task or timer yet to run may keep the content reachable for a while.
func (c *libraryFilterContent[T]) dispose() {
	c.gen.Add(1)
	c.closed = true
	if c.installed {
		saveColumnWidths(c.table, c.widthsKey)
	}
	if data, ok := draggedTableData.(*unison.TableDragData[*Node[T]]); ok && data.Table == c.table {
		draggedTableData = nil
	}
	c.cache = nil
	c.lastLoad = nil
	c.inFile = nil
	c.decorator.files = nil
	c.decorator.rowFiles = nil
	c.provider.SetRootData(nil)
	c.table.SyncToModel()
}

// updateStatus implements libraryFilterView. The footer's tooltip names the sources that can't be found and the files
// that couldn't be read.
func (c *libraryFilterContent[T]) updateStatus() {
	c.footer.SetTitle(libraryFilterStatus{
		files:      len(c.decorator.files),
		rows:       c.rowCount,
		matched:    c.matched,
		missing:    len(c.shell.missing),
		unreadable: len(c.unreadable),
		waiting:    !c.shell.listed,
		loading:    c.loading,
		installed:  c.installed,
		filtered:   c.filtered,
	}.String())
	var lines []string
	if c.installed {
		if len(c.shell.missing) != 0 {
			lines = append(lines, i18n.Text("Can't be found:"))
			for _, entry := range c.shell.missing {
				lines = append(lines, "• "+describeLibraryEntry(gurps.GlobalSettings().Libraries, entry))
			}
		}
		if len(c.unreadable) != 0 {
			lines = append(lines, i18n.Text("Couldn't be read:"))
			for _, one := range c.unreadable {
				lines = append(lines, "• "+one.file.libTitle+": "+one.file.Path)
			}
		}
	}
	if len(lines) == 0 {
		c.footer.Tooltip = nil
	} else {
		c.footer.Tooltip = newWrappedTooltip(strings.Join(lines, "\n"))
	}
	c.footer.MarkForLayoutAndRedraw()
}

// describeLibraryEntry returns the title of the entry's library among the libraries, or its key when none has it,
// followed by the entry's path within the library, if it has one.
func describeLibraryEntry(libs *library.Libraries, entry gurps.LibraryFile) string {
	title := entry.Library
	if lib := libs.Lookup(entry.Library); lib != nil {
		title = lib.Data().Title
	}
	if entry.Path == "" {
		return title
	}
	return title + ": " + entry.Path
}

// libraryFilterStatus is what the footer of a library filter view reports.
type libraryFilterStatus struct {
	files      int
	rows       int // Every row, at any depth
	matched    int // The rows the filters in force let through
	missing    int // The entries of the set that name nothing
	unreadable int // The files that couldn't be read
	waiting    bool
	loading    bool
	installed  bool
	filtered   bool
}

// String returns the footer's text. Until the first load is shown, it says the view is loading, or waiting for the
// Library Explorer to list the libraries, without which there are no files to load.
func (s libraryFilterStatus) String() string {
	if !s.installed {
		if s.waiting {
			return i18n.Text("Waiting for the Library Explorer to finish scanning the libraries…")
		}
		return i18n.Text("Loading…")
	}
	parts := make([]string, 0, 5)
	if s.files == 1 {
		parts = append(parts, i18n.Text("1 file"))
	} else {
		parts = append(parts, i18n.Text("%d files", s.files))
	}
	switch {
	case s.filtered && s.rows == 1:
		parts = append(parts, i18n.Text("%d of 1 row matches", s.matched))
	case s.filtered && s.matched == 1:
		parts = append(parts, i18n.Text("1 of %d rows matches", s.rows))
	case s.filtered:
		parts = append(parts, i18n.Text("%d of %d rows match", s.matched, s.rows))
	case s.rows == 1:
		parts = append(parts, i18n.Text("1 row"))
	default:
		parts = append(parts, i18n.Text("%d rows", s.rows))
	}
	switch s.missing {
	case 0:
	case 1:
		parts = append(parts, i18n.Text("1 source can't be found"))
	default:
		parts = append(parts, i18n.Text("%d sources can't be found", s.missing))
	}
	switch s.unreadable {
	case 0:
	case 1:
		parts = append(parts, i18n.Text("1 file couldn't be read"))
	default:
		parts = append(parts, i18n.Text("%d files couldn't be read", s.unreadable))
	}
	if s.loading {
		parts = append(parts, i18n.Text("reloading…"))
	}
	return strings.Join(parts, ", ")
}
