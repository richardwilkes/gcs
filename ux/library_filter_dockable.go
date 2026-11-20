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
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

var (
	_ unison.Dockable              = &LibraryFilterDockable{}
	_ unison.TabCloser             = &LibraryFilterDockable{}
	_ unison.UndoManagerProvider   = &LibraryFilterDockable{}
	_ Rebuildable                  = &LibraryFilterDockable{}
	_ readOnlyView                 = &LibraryFilterDockable{}
	_ listFilterObserver           = &LibraryFilterDockable{}
	_ gurps.SheetSettingsResponder = &LibraryFilterDockable{}
	_ boundsDeferredDockable       = &LibraryFilterDockable{}
	_ openStateSharer              = &LibraryFilterDockable{}
)

// libraryFilterBoundsWait is how long a library filter view that is to have a window of its own waits for its first
// rows before taking one without them (see boundsDeferredDockable).
const libraryFilterBoundsWait = 500 * time.Millisecond

// Seams so that tests can supply the listing the views resolve their sets against and shorten the wait for a reload.
var (
	// libraryFilterListing returns the Library Explorer's last listing of the libraries, or nil when there is none
	// yet.
	libraryFilterListing = func() *navigatorListing {
		if Workspace.Navigator == nil {
			return nil
		}
		return Workspace.Navigator.libraryListing()
	}
	// libraryFilterReloadDelay is how long a view waits after hearing that one of its files changed before it reads
	// the files again, so that a file written in pieces, or several files written together, are read once.
	libraryFilterReloadDelay = 250 * time.Millisecond
)

// LibraryFilterDockable shows the rows of a source set of library list files, all of one list type, together in one
// table, with a column naming the file each came from, and filters them as a library list is filtered. It never
// changes what it shows, and the editors opened from it are read-only. Its content, which is built for the list type,
// is replaced when a set of another type is chosen.
type LibraryFilterDockable struct {
	unison.Panel
	kind    libraryFilterKind
	set     *gurps.LibrarySourceSet // nil for every file of the list type
	content libraryFilterView
	undoMgr *unison.UndoManager
	// resolved and missing are what resolving the set against the Library Explorer's listing found last: the files it
	// covers, and its entries that name nothing.
	resolved             []libraryFilterFile
	missing              []gurps.LibraryFile
	boundsKnownCallbacks []func()
	created              time.Time
	loads                sync.WaitGroup // The loads in flight; see waitForLoads
	scale                int
	adHoc                bool // The set isn't one of the saved ones
	listed               bool // The set has been resolved against a listing since the content was built
	reloadPending        bool
	closed               bool
}

// NewLibraryFilter opens a new library filter view on the list type and source set last used.
func NewLibraryFilter() {
	key, set := lastLibraryFilterSet()
	openLibraryFilter(key, set, false)
}

// lastLibraryFilterSet returns the key of the list type last used and its saved source set last used, or nil, standing
// for every file of the type, when none was or the one that was no longer exists.
func lastLibraryFilterSet() (key string, set *gurps.LibrarySourceSet) {
	settings := gurps.GlobalSettings()
	key = settings.LibraryFilter.LastKey
	if name := settings.LibraryFilter.LastSets[key]; name != "" {
		set = settings.LibrarySourceSetNamed(key, name)
	}
	return key, set
}

// openLibraryFilter opens a new library filter view on the set of the list type with the key, which is unsaved when
// adHoc is true. A nil set stands for every file of the type.
func openLibraryFilter(key string, set *gurps.LibrarySourceSet, adHoc bool) *LibraryFilterDockable {
	d := newLibraryFilterDockable(key, set, adHoc)
	PlaceInDock(d, dgroup.Libraries, false)
	if d.Window() != nil {
		d.content.focusFilter()
	}
	return d
}

// newLibraryFilterDockable returns a new library filter view, as openLibraryFilter does, without placing it. A key that
// names no list type stands for the first one there is.
func newLibraryFilterDockable(key string, set *gurps.LibrarySourceSet, adHoc bool) *LibraryFilterDockable {
	d := &LibraryFilterDockable{
		undoMgr: unison.NewUndoManager(100, func(err error) { errs.Log(err) }),
		scale:   gurps.GlobalSettings().General.InitialListUIScale,
		created: time.Now(),
	}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{Columns: 1})
	if _, ok := libraryFilterKindFor(key); !ok {
		key = libraryFilterKinds()[0].key
		set = nil
		adHoc = false
	}
	d.chooseSet(key, set, adHoc)
	return d
}

// chooseSet has the view show the set as showSet does and, unless it is unsaved, remembers it as the set a new view
// starts on. Every file of the type is remembered by recording no set for it.
func (d *LibraryFilterDockable) chooseSet(key string, set *gurps.LibrarySourceSet, adHoc bool) {
	if !d.showSet(key, set, adHoc) || adHoc {
		return
	}
	settings := &gurps.GlobalSettings().LibraryFilter
	settings.LastKey = key
	switch {
	case set == nil:
		delete(settings.LastSets, key)
	case settings.LastSets == nil:
		settings.LastSets = map[string]string{key: set.Name}
	default:
		settings.LastSets[key] = set.Name
	}
}

// showSet has the view show the set of the list type with the key, which is unsaved when adHoc is true; a nil set
// stands for every file of the type. A set of another type than the one shown replaces the content, once the editors
// opened from the rows it shows have closed, carrying the quick filter's text over. It returns false, changing nothing,
// when there is no such list type or an editor refused to close. The source set popup is part of the content, so a
// caller within the popup's callbacks must not use the popup afterward.
func (d *LibraryFilterDockable) showSet(key string, set *gurps.LibrarySourceSet, adHoc bool) bool {
	kind, ok := libraryFilterKindFor(key)
	if !ok {
		return false
	}
	newKind := d.content == nil || kind.key != d.kind.key
	if newKind && !CloseGroup(d) {
		return false
	}
	d.set = set
	d.adHoc = adHoc
	if newKind {
		var text string
		focused := false
		if d.content != nil {
			text = d.content.quickFilterText()
			focused = d.content.AsPanel().HasInSelfOrDescendants(func(p *unison.Panel) bool { return p.Focused() })
			d.content.dispose()
			d.RemoveChild(d.content)
		}
		d.kind = kind
		d.listed = false
		d.resolved = nil
		d.missing = nil
		d.undoMgr.Clear() // Its edits are to the scale field of the content just discarded
		d.content = kind.newContent(d)
		d.content.setQuickFilterText(text)
		d.content.AsPanel().MarkForLayoutRecursivelyUpward()
		d.MarkForRedraw()
		if focused {
			d.content.focusFilter()
		}
	} else {
		d.content.syncSetPopup()
	}
	UpdateTitleForDockable(d)
	d.resolveAndLoad()
	return true
}

// shownEntries returns the source entries of what the view shows: those of its set, or for the set of every file, one
// for each whole library.
func (d *LibraryFilterDockable) shownEntries() []gurps.LibraryFile {
	if d.set != nil {
		return slices.Clone(d.set.Entries)
	}
	libs := gurps.GlobalSettings().Libraries.List()
	entries := make([]gurps.LibraryFile, 0, len(libs))
	for _, lib := range libs {
		entries = append(entries, gurps.LibraryFile{Library: lib.Key()})
	}
	return gurps.NormalizeLibrarySourceEntries(entries)
}

// libraryFilterLibraryRekeyed brings the source sets into line with the library oldKey now going by newKey: the saved
// sets, which are then saved, and the unsaved sets of the open views. The views find their files again once the
// Library Explorer has listed the libraries anew, and read them afresh, since their rows record their library.
func libraryFilterLibraryRekeyed(oldKey, newKey string) {
	gurps.GlobalSettings().RekeyLibrarySourceSets(oldKey, newKey)
	for _, d := range allLibraryFilters() {
		d.rekeyUnsavedSet(oldKey, newKey)
	}
	saveGlobalSettings()
}

// rekeyUnsavedSet changes the library of the entries of the view's unsaved set, if it has one, from oldKey to newKey,
// normalizing them afterward, since newKey may already have been there.
func (d *LibraryFilterDockable) rekeyUnsavedSet(oldKey, newKey string) {
	if d.adHoc && d.set != nil {
		gurps.RekeyLibraryEntries(d.set.Entries, oldKey, newKey)
		d.set.Entries = gurps.NormalizeLibrarySourceEntries(d.set.Entries)
	}
}

// libraryFilterPathRenamed brings the source sets into line with the file or folder at oldPath in the library with the
// key now being at newPath, both library-relative and slash-separated: the saved sets, which are then saved if that
// changed them, and the unsaved sets of the open views. The views find their files again once the Library Explorer has
// listed the libraries anew.
func libraryFilterPathRenamed(key, oldPath, newPath string) {
	for _, d := range allLibraryFilters() {
		d.renameInUnsavedSet(key, oldPath, newPath)
	}
	if gurps.GlobalSettings().RenameLibrarySourcePaths(key, oldPath, newPath) {
		saveGlobalSettings()
	}
}

// renameInUnsavedSet does what gurps.RenameLibraryEntries does to the entries of the view's unsaved set, if it has one,
// normalizing them afterward if that changed them, since newPath may already have been there.
func (d *LibraryFilterDockable) renameInUnsavedSet(key, oldPath, newPath string) {
	if d.adHoc && d.set != nil && gurps.RenameLibraryEntries(d.set.Entries, key, oldPath, newPath) {
		d.set.Entries = gurps.NormalizeLibrarySourceEntries(d.set.Entries)
	}
}

// resolveAndLoad resolves the set against the Library Explorer's listing and has the content load the files it
// covers, which reads only those that changed since they were last read, even when they are the files it covered last
// time, since nothing may have been watching them change. Until there is a listing, the footer says the view is waiting
// for one.
func (d *LibraryFilterDockable) resolveAndLoad() {
	listing := libraryFilterListing()
	if listing == nil {
		d.content.updateStatus()
		return
	}
	d.resolved, d.missing = resolveLibraryFilterFiles(listing, gurps.GlobalSettings().Libraries, d.kind.ext, d.set)
	d.listed = true
	d.content.load(d.resolved)
}

// reload reads every file of the set again, whether or not it has changed.
func (d *LibraryFilterDockable) reload() {
	d.content.dropCache()
	d.resolveAndLoad()
}

// scheduleReload loads the files again once libraryFilterReloadDelay has passed, collapsing the requests made in the
// meantime into one. Only the files that changed are read. Nothing is loaded while the views are held off (see
// libraryFilterHeld).
func (d *LibraryFilterDockable) scheduleReload() {
	if d.reloadPending || libraryFilterHeld() {
		return
	}
	d.reloadPending = true
	unison.InvokeTaskAfter(func() {
		d.reloadPending = false
		if !d.closed && !libraryFilterHeld() {
			d.resolveAndLoad()
		}
	}, libraryFilterReloadDelay)
}

// libraryFilterHeld returns true while the views are to leave their files alone: while a library update is replacing
// them, with the Library Explorer holding off its content cache prewarm for as long (see suspendContentCachePrewarm),
// since they would otherwise read every file part-written, over and over. Lifting the hold reloads the Library
// Explorer, whose new listing has every view load its files once (see libraryFilterListingChanged).
func libraryFilterHeld() bool {
	return Workspace.Navigator != nil && Workspace.Navigator.prewarmSuspensions > 0
}

// waitForLoads waits for the loads in flight to finish reading their files. What they read is shown by a task they
// leave for the UI thread, which may not have run yet.
func (d *LibraryFilterDockable) waitForLoads() {
	d.loads.Wait()
}

// contentInstalled is told by the content when it has shown the rows of a load.
func (d *LibraryFilterDockable) contentInstalled() {
	d.notifyBoundsKnown()
}

// libraryFilterListingChanged has every library filter view resolve its set against the Library Explorer's new
// listing and load its files, which reads those that changed, unless the views are held off (see libraryFilterHeld).
func libraryFilterListingChanged() {
	if libraryFilterHeld() {
		return
	}
	for _, d := range allLibraryFilters() {
		d.resolveAndLoad()
	}
}

// libraryFilterPathChanged has every library filter view showing the file at the path load its files again shortly
// (see scheduleReload).
func libraryFilterPathChanged(fullPath string) {
	fullPath = filepath.Clean(fullPath)
	for _, d := range allLibraryFilters() {
		if slices.ContainsFunc(d.resolved, func(f libraryFilterFile) bool { return f.fullPath == fullPath }) {
			d.scheduleReload()
		}
	}
}

// allLibraryFilters returns the open library filter views, those waiting for a window among them.
func allLibraryFilters() []*LibraryFilterDockable {
	var list []*LibraryFilterDockable
	for _, one := range AllDockables() {
		if d, ok := one.AsPanel().Self.(*LibraryFilterDockable); ok {
			list = append(list, d)
		}
	}
	return list
}

// TitleIcon implements unison.Dockable.
func (d *LibraryFilterDockable) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  d.kind.icon,
		Size: suggestedSize,
	}
}

// Title implements unison.Dockable.
func (d *LibraryFilterDockable) Title() string {
	return libraryFilterSetTitle(d.kind, d.set, d.adHoc)
}

// String implements fmt.Stringer.
func (d *LibraryFilterDockable) String() string {
	return d.Title()
}

// Tooltip implements unison.Dockable.
func (d *LibraryFilterDockable) Tooltip() string {
	return ""
}

// Modified implements unison.Dockable. The view changes nothing, so it is never modified.
func (d *LibraryFilterDockable) Modified() bool {
	return false
}

// MayAttemptClose implements unison.TabCloser.
func (d *LibraryFilterDockable) MayAttemptClose() bool {
	return MayAttemptCloseOfGroup(d)
}

// AttemptClose implements unison.TabCloser. The editors opened from the view close with it, and the load in flight,
// if any, is discarded.
func (d *LibraryFilterDockable) AttemptClose() bool {
	if !CloseGroup(d) {
		return false
	}
	d.content.dispose()
	d.closed = true
	d.boundsKnownCallbacks = nil
	return AttemptCloseForDockable(d)
}

// UndoManager implements unison.UndoManagerProvider. Only the scale field makes edits.
func (d *LibraryFilterDockable) UndoManager() *unison.UndoManager {
	return d.undoMgr
}

// Rebuild implements Rebuildable.
func (d *LibraryFilterDockable) Rebuild(_ bool) {
	d.content.rebuild()
	UpdateTitleForDockable(d)
}

// isReadOnly implements readOnlyView. The view never changes the rows it shows: they are copies of what its files held
// when it read them, which nothing would save.
func (d *LibraryFilterDockable) isReadOnly() bool {
	return true
}

// listFiltersChanged implements listFilterObserver.
func (d *LibraryFilterDockable) listFiltersChanged(key string, source *listFilterPopup) {
	d.content.listFiltersChanged(key, source)
}

// syncSharedOpenStates implements openStateSharer.
func (d *LibraryFilterDockable) syncSharedOpenStates(origin *unison.Panel, ids map[tid.TID]bool) {
	d.content.syncSharedOpenStates(origin, ids)
}

// SheetSettingsUpdated implements gurps.SheetSettingsResponder.
func (d *LibraryFilterDockable) SheetSettingsUpdated(entity *gurps.Entity, parts gurps.SheetSettingsParts,
	fullRebuild bool,
) {
	d.content.sheetSettingsUpdated(entity, parts, fullRebuild)
}

// BoundsKnown implements boundsDeferredDockable. The view's size is known once its first rows are shown.
func (d *LibraryFilterDockable) BoundsKnown() bool {
	return d.content.hasInstalled()
}

// WhenBoundsKnown implements boundsDeferredDockable. If the bounds are already known, the callback is invoked before
// returning.
func (d *LibraryFilterDockable) WhenBoundsKnown(callback func()) {
	if d.BoundsKnown() {
		callback()
		return
	}
	d.boundsKnownCallbacks = append(d.boundsKnownCallbacks, callback)
	if len(d.boundsKnownCallbacks) == 1 {
		unison.InvokeTaskAfter(d.notifyBoundsKnown, max(libraryFilterBoundsWait-time.Since(d.created), 0))
	}
}

// notifyBoundsKnown invokes the callbacks waiting for the bounds. One of them gives the view its window, in which the
// quick filter then takes the focus, as it does when the view is placed in a window straight away.
func (d *LibraryFilterDockable) notifyBoundsKnown() {
	callbacks := d.boundsKnownCallbacks
	d.boundsKnownCallbacks = nil
	if d.closed || len(callbacks) == 0 {
		return
	}
	for _, callback := range callbacks {
		callback()
	}
	if d.Window() != nil {
		d.content.focusFilter()
	}
}
