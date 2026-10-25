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
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/zeebo/xxh3"
)

// The picker of an open grant row shows grantPickerRows library items at once, with columns of these widths for the
// points or difficulty and the library file.
const (
	grantPickerRows        = 6
	grantPickerPointsWidth = 48
	grantPickerFileWidth   = 200
)

// grantLibraryItem is an item a grant can give, from a library or a table: a trait, skill or spell.
type grantLibraryItem struct {
	node   any
	from   gurps.LibraryFile
	name   string
	points string
	// file names the library file the item came from, which tells apart items of the same name.
	file  string
	words []string
}

// newGrantLibraryItem returns the node, from the library file, as an item a grant can give, or nil for a technique or
// an item that grants something itself, among its own features or those of a trait's modifiers, which no grant can
// give.
func newGrantLibraryItem(node any, from gurps.LibraryFile) *grantLibraryItem {
	item := &grantLibraryItem{node: node, from: from, file: xfilepath.TrimExtension(filepath.Base(from.Path))}
	var features []gurps.Features
	switch one := node.(type) {
	case *gurps.Trait:
		item.name, item.points, features = one.NameWithReplacements(), one.AdjustedPoints(nil).String(),
			[]gurps.Features{one.Features}
		gurps.Traverse(func(mod *gurps.TraitModifier) bool {
			features = append(features, mod.Features)
			return false
		}, false, false, one.Modifiers...)
	case *gurps.Skill:
		if one.IsTechnique() {
			return nil
		}
		diff := one.AdjustedDifficulty()
		item.name, item.points, features = one.String(), diff.Description(nil), []gurps.Features{one.Features}
	case *gurps.Spell:
		item.name, item.points, features = one.String(), one.Difficulty.Description(nil), []gurps.Features{one.Features}
	}
	for _, list := range features {
		if slices.ContainsFunc(list, func(f gurps.Feature) bool { return gurps.IsGrant(f.FeatureType()) }) {
			return nil
		}
	}
	item.words = grantFilterWords(item.name)
	return item
}

// collectGrantItems returns the items a grant can give among the nodes and their descendants, containers left out.
func collectGrantItems[T gurps.Node[T]](from gurps.LibraryFile, nodes []T) []*grantLibraryItem {
	var items []*grantLibraryItem
	gurps.Traverse(func(one T) bool {
		if item := newGrantLibraryItem(one, from); item != nil {
			items = append(items, item)
		}
		return false
	}, false, true, nodes...)
	return items
}

// grantPickerState is what the picker of an open grant row holds that its widgets would lose when the panel is rebuilt,
// as any edit of the row's other controls rebuilds it: the grant it is for, the filter typed and the list, whose
// selection the picker built next takes up. built records that the build in progress made the picker, so that the
// state goes once no build does.
type grantPickerState struct {
	grant  *gurps.Grant
	filter string
	list   *unison.List[*grantLibraryItem]
	built  bool
}

// grant adds the controls of an open grant row: the item granted, its level, what happens when the sheet already has
// it and, as a chip, whether the item's prerequisites are checked. A zone to drop the item on follows, which with an
// item chosen replaces it, then, while no item is chosen or once the button naming the item or the zone is clicked, a
// picker of the libraries' items of the kind granted.
func (p *featuresPanel) grant(fields, box, chips *unison.Panel, path string, g *gurps.Grant) {
	key := func(name string) string { return path + ":" + name }
	chosen := g.Key() != ""
	if chosen {
		button := unison.NewButton()
		button.SetTitle(g.ItemName())
		button.RefKey = key("item")
		button.Tooltip = newWrappedTooltip(i18n.Text("Choose a different item from the libraries"))
		button.ClickCallback = func() {
			if p.grantPicking == g.Key() {
				p.grantPicking = ""
				p.rebuild(key("item"))
			} else {
				p.grantPicking = g.Key()
				p.rebuild(key("filter"))
			}
		}
		addCentered(fields, button)
	} else {
		addJoiningWords(fields, i18n.Text("(nothing chosen yet)"))
	}
	if chosen && g.Leveled() {
		title := i18n.Text("Level")
		addJoiningWords(fields, i18n.Text("at level"))
		field := NewDecimalField(p.targetMgr, key("level"), title, func() fxp.Int { return g.Level },
			func(v fxp.Int) { p.edit(title, key("level"), "", func() { g.Level = v }) }, fxp.Min, fxp.Max, false, false)
		// A skill or spell is granted at a tier, which is whole; a trait's levels may be fractional.
		if g.Type != feature.GrantTrait {
			refuseFractions(field)
		}
		field.Tooltip = newWrappedTooltip(grantLevelTooltip(g.Type))
		p.addCompact(fields, field.withoutUndo())
	}
	if merges := g.AllowedMerges(); chosen && len(merges) > 1 {
		addJoiningWords(fields, i18n.Text("and if already present"))
		addCentered(fields, compactPopup(&p.sentenceRows, key("merge"), i18n.Text("When Already Present"), merges,
			g.EffectiveMerge(), nil, func(v grantmerge.Mode) { g.Merge = v }))
	}
	p.optional(chips, path, "checkprereqs", i18n.Text("+ check prerequisites"), i18n.Text("Check Prerequisites"),
		i18n.Text("Stop Checking Prerequisites"), g.CheckPrereqs,
		func() { g.CheckPrereqs = true },
		func() { g.CheckPrereqs = false },
		func(chip *unison.Panel) { addJoiningWords(chip, i18n.Text("checking its prerequisites")) })
	picking := !chosen || p.grantPicking == g.Key()
	box.AddChild(newEmptyPlaceholder(key("drop"), grantDropText(g.Type, chosen), func() {
		if picking {
			p.focusOn(key("filter"))
			return
		}
		p.grantPicking = g.Key()
		p.rebuild(key("filter"))
	}))
	if picking {
		box.AddChild(p.grantPicker(path, g))
	}
}

// grantDropText invites dropping an item on a grant of the given type, to replace the item chosen or to fill the grant.
func grantDropText(t feature.Type, chosen bool) string {
	var what string
	switch t {
	case feature.GrantSkill:
		what = i18n.Text("skill")
	case feature.GrantSpell:
		what = i18n.Text("spell")
	default:
		what = i18n.Text("trait")
	}
	if chosen {
		return i18n.Text("Drop a %s here from a library or sheet to replace the one chosen.", what)
	}
	return i18n.Text("Drop a %s here from a library or sheet, or choose one below.", what)
}

// grantFocusKey returns the reference key of the control of the grant row at the path that takes the focus once the
// grant has its item: its level field, or its first control when the level means nothing for the item.
func grantFocusKey(path string, g *gurps.Grant) string {
	if g.Leveled() {
		return path + ":level"
	}
	return path + keyFirst
}

// grantLevelTooltip explains what the level of a grant of the given type means.
func grantLevelTooltip(t feature.Type) string {
	if t == feature.GrantTrait {
		return i18n.Text("The level of the trait. A negative level only takes levels away from a trait already present.")
	}
	return i18n.Text(`The level is counted in steps of points: level 1 is what 1 point buys, level 2 what 2 points buy, level 3 what 4 points buy, and each level after that 4 points more. A negative level only takes levels or points away from one already present.`)
}

// grantPicker returns the filter field and list of library items to choose the granted item from, along with the
// buttons that make a blank one and open the editor on the one chosen. The items are those the panel's owner can be
// given (see grantable). The filter and the item chosen are those of the picker last built for the grant, if any
// (see grantPickerState).
func (p *featuresPanel) grantPicker(path string, g *gurps.Grant) *unison.Panel {
	key := func(name string) string { return path + ":" + name }
	panel := newColumn()
	items := p.grantable(grantLibraryItems(g.Type))
	if p.picker.grant != g {
		p.picker = grantPickerState{grant: g}
	}
	var chosen *grantLibraryItem
	if last := p.picker.list; last != nil && last.Selection.Count() != 0 {
		chosen = last.DataAtIndex(last.Selection.FirstSet())
	}
	p.picker.built = true
	filter := unison.NewField()
	filter.RefKey = key("filter")
	filter.SetText(p.picker.filter)
	filter.Watermark = i18n.Text("Filter the library")
	filter.Accessibility.Name = i18n.Text("Library Filter")
	filter.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	list := unison.NewList[*grantLibraryItem]()
	list.Factory = &grantItemCellFactory{}
	list.Accessibility.Name = i18n.Text("Library Items")
	// The focus stays in the filter, where Return chooses the selected item, of which there is only ever one.
	list.SetFocusable(false)
	list.SetAllowMultipleSelection(false)
	// A list sizes itself by building the cell of every row, which for a whole library is slow enough to stall each
	// layout of the editor; its rows all have the same height, so the size follows from their count alone.
	list.SetSizer(func(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
		prefSize = geom.NewSize(hint.Width, xmath.Ceil(list.Factory.CellHeight())*float32(list.Count()))
		return prefSize, prefSize, unison.MaxSize(prefSize)
	})
	choose := func(title string, item *grantLibraryItem) {
		p.sentenceRows.restructure(title, key("filter"), "", func() string {
			g.SetItem(item.node, item.from)
			p.grantPicking = ""
			return grantFocusKey(path, g)
		})
	}
	list.DoubleClickCallback = func() {
		if list.Selection.Count() != 0 {
			choose(i18n.Text("Choose Granted Item"), list.DataAtIndex(list.Selection.FirstSet()))
		}
	}
	scroll := unison.NewScrollPanel()
	scroll.SetContent(list, behavior.Follow, behavior.Fill)
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	p.picker.list = list
	// refill fills the list with the items the filter leaves, selecting keep among them, or else the first.
	refill := func(keep *grantLibraryItem) {
		p.picker.filter = filter.Text()
		query := grantFilterWords(filter.Text())
		list.Clear()
		index := -1
		for _, one := range items {
			if grantItemMatches(query, one.words) {
				if one == keep {
					index = list.Count()
				}
				list.Append(one)
			}
		}
		if list.Count() != 0 {
			index = max(index, 0)
			list.Select(false, index)
			list.ScrollRectIntoView(list.RowRect(index))
		}
		// The list marks only itself for layout, which doesn't resize it: the scroll panel sizes its content, so it
		// is the one laid out again, or the list would keep the frame of the rows it showed before.
		scroll.MarkForLayoutAndRedraw()
	}
	filter.ModifiedCallback = func(_, _ *unison.FieldState) { refill(nil) }
	keyDown := filter.KeyDownCallback
	filter.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
		switch keyCode {
		case unison.KeyUp, unison.KeyDown:
			// With a modifier down, the key is the filter's, as Shift-Down selects to the end of its text.
			if noModifiersDown(mods) {
				return list.KeyDownCallback(keyCode, mods, repeat)
			}
			return keyDown(keyCode, mods, repeat)
		case unison.KeyReturn, unison.KeyNumPadEnter:
			// With a modifier down, the key is the editor's, as Cmd-Return applies the changes.
			if noModifiersDown(mods) {
				list.DoubleClickCallback()
				return true
			}
			return keyDown(keyCode, mods, repeat)
		default:
			return keyDown(keyCode, mods, repeat)
		}
	}
	refill(chosen)
	panel.AddChild(filter)
	height := xmath.Ceil(list.Factory.CellHeight()) * grantPickerRows
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign:   align.Fill,
		HGrab:    true,
		MinSize:  geom.NewSize(0, height),
		SizeHint: geom.NewSize(0, height),
	})
	panel.AddChild(scroll)
	buttons := newFlow()
	blank := unison.NewButton()
	blank.SetTitle(i18n.Text("New Blank"))
	blank.RefKey = key("blank")
	blank.ClickCallback = func() { choose(i18n.Text("New Blank Granted Item"), &grantLibraryItem{}) }
	buttons.AddChild(blank)
	editButton := unison.NewButton()
	editButton.SetTitle(i18n.Text("Edit"))
	editButton.RefKey = key("edititem")
	editButton.SetEnabled(g.Key() != "")
	editButton.ClickCallback = func() { p.editGrantedItem(g) }
	buttons.AddChild(editButton)
	panel.AddChild(buttons)
	return panel
}

// grantItemOwner owns the editor of a grant's item. The editor works on a copy, which each apply, undo and redo hands
// to the grant with the item's key, since undoing the panel's own edits replaces its grants with copies. hash is the
// hash of the grant's item as the editor last saw it, by which the panel tells that the item changed under the editor;
// see closeStaleGrantItemEditors. The panel is known by the features it edits (see editsFeaturesOf), since the editor
// holding it may build a new panel for the same features.
type grantItemOwner struct {
	Rebuildable // the editor holding the panel
	panel       *featuresPanel
	key         tid.TID
	hash        uint64
	item        fmt.Stringer
	commit      func(g *gurps.Grant)
}

// String names the item, for the title of its editor.
func (o *grantItemOwner) String() string {
	return o.item.String()
}

// Rebuild hands the copy to the grant.
func (o *grantItemOwner) Rebuild(_ bool) {
	if g := o.grant(); g != nil {
		o.commit(g)
		o.hash = grantItemHash(g)
		MarkModified(o.panel)
		o.panel.rebuild("")
	}
}

// editsFeaturesOf returns true if the owner's panel edits the same features as p, taking p for its panel from here on
// when it does, since the editor holding the panel builds a new one when its content is rebuilt.
func (o *grantItemOwner) editsFeaturesOf(p *featuresPanel) bool {
	if o.panel.features != p.features {
		return false
	}
	o.panel = p
	return true
}

// grant returns the panel's grant with the item's key, or nil once there is none.
func (o *grantItemOwner) grant() *gurps.Grant {
	for _, f := range *o.panel.features {
		if g, ok := f.(*gurps.Grant); ok && g.Key() == o.key {
			return g
		}
	}
	return nil
}

// current returns true if the panel's grant still holds the item the editor was opened on.
func (o *grantItemOwner) current() bool {
	g := o.grant()
	return g != nil && grantItemHash(g) == o.hash
}

// grantItemHash returns a hash of the grant's copy of its item, modifiers and all. The item is the one of the grant's
// type, as for the grant itself, whatever else a loaded grant holds.
func grantItemHash(g *gurps.Grant) uint64 {
	h := xxh3.New()
	switch g.Type {
	case feature.GrantSkill:
		if g.Skill != nil {
			gurps.HashJSON(h, g.Skill)
		}
	case feature.GrantSpell:
		if g.Spell != nil {
			gurps.HashJSON(h, g.Spell)
		}
	default:
		if g.Trait != nil {
			gurps.HashJSON(h, g.Trait)
		}
	}
	return h.Sum64()
}

// closeStaleGrantItemEditors closes, without asking, the editors open on the items of the panel's grants whose grant
// is gone, gives another item or had its item changed, as by a sync, since what such an editor would save has nowhere
// to go or would overwrite the new item. The editors opened from one, on the item's modifiers and weapons, close with
// it the same way, since what they would save goes into the same data; see editorShell.discardAndCloseAll.
func (p *featuresPanel) closeStaleGrantItemEditors() {
	for _, d := range AllDockables() {
		owned, ok := d.AsPanel().Self.(Owned)
		if !ok {
			continue
		}
		if o, isGrant := owned.Owner().(*grantItemOwner); isGrant && o.editsFeaturesOf(p) && !o.current() {
			if closer, canDiscard := d.(interface{ discardAndCloseAll() }); canDiscard {
				closer.discardAndCloseAll()
			}
		}
	}
}

// editGrantedItem opens the normal editor on a copy of the grant's item, or brings forward the one already open. A
// grant with no item chosen yet has nothing to edit.
func (p *featuresPanel) editGrantedItem(g *gurps.Grant) {
	outer := p.AncestorOrSelf[Rebuildable]()
	if g.Key() == "" || outer == nil || Activate(func(d unison.Dockable) bool {
		owned, ok := d.AsPanel().Self.(Owned)
		if !ok {
			return false
		}
		o, isGrant := owned.Owner().(*grantItemOwner)
		return isGrant && o.editsFeaturesOf(p) && o.key == g.Key()
	}) {
		return
	}
	o := &grantItemOwner{Rebuildable: outer, panel: p, key: g.Key(), hash: grantItemHash(g)}
	// The item is the one of the grant's type, as for the grant itself; see grantItemHash.
	switch g.Type {
	case feature.GrantSkill:
		item := g.Skill.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy)
		o.item, o.commit = item, func(g *gurps.Grant) { g.Skill = item.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy) }
		EditSkill(o, item)
	case feature.GrantSpell:
		item := g.Spell.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy)
		o.item, o.commit = item, func(g *gurps.Grant) { g.Spell = item.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy) }
		EditSpell(o, item)
	default:
		item := g.Trait.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy)
		o.item, o.commit = item, func(g *gurps.Grant) { g.Trait = item.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy) }
		EditTrait(o, item)
	}
}

// grantFileCache is what a library file held for the grant pickers when it was last read.
type grantFileCache struct {
	modTime time.Time
	items   []*grantLibraryItem
}

// grantFileItems holds the items the grant pickers found in each library file, by library key and path, until the file
// changes or is no longer found under that key.
var grantFileItems = make(map[gurps.LibraryFile]grantFileCache)

// grantItemList is the sorted list of items the pickers last showed for one kind of grant, along with the files and
// modification times it came from, so that a picker rebuilt while the libraries are unchanged shows the same list again
// rather than gathering and sorting it afresh.
type grantItemList struct {
	stamps []grantFileStamp
	items  []*grantLibraryItem
}

// grantFileStamp identifies a library file as it was when read.
type grantFileStamp struct {
	file    gurps.LibraryFile
	modTime time.Time
}

func (s grantFileStamp) equal(other grantFileStamp) bool {
	return s.file == other.file && s.modTime.Equal(other.modTime)
}

// grantItemLists holds the list the pickers last showed for each kind of grant; see grantItemList.
var grantItemLists = make(map[feature.Type]grantItemList)

// grantLibraryItems returns the items of the libraries a grant of the given type can give, reading only the files
// changed since they were last read, and the list last returned when none has. Hidden files and folders are skipped and
// symlinked folders followed, as the library navigator skips and follows them.
func grantLibraryItems(t feature.Type) []*grantLibraryItem {
	ext, load := grantItemLoader(t)
	var stamps []grantFileStamp
	var caches []grantFileCache
	seen := make(map[gurps.LibraryFile]bool)
	for _, lib := range gurps.GlobalSettings().Libraries.List() {
		root := lib.Path(false)
		if root == "" {
			continue
		}
		fileSystem := os.DirFS(root)
		// WalkDir doesn't follow the symlinks it finds, but does walk the target of the one it starts from, so each
		// symlink to a folder is walked on its own once the library's own folders have been, so that an item in one of
		// those is offered by its path there, whatever symlinks also lead to it. visited holds the real path of each
		// folder walked, so that a folder a symlink leads to again is not walked again: it would offer its items twice,
		// and a symlink back up the tree would never end.
		var links []string
		visited := make(map[string]bool)
		pointsAtFolder := func(path string) bool {
			info, statErr := fs.Stat(fileSystem, path)
			return statErr == nil && info.IsDir()
		}
		walk := func(path string, d fs.DirEntry, err error) error {
			switch {
			case err != nil:
				return nil
			case path != "." && strings.HasPrefix(d.Name(), "."):
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			case d.Type()&fs.ModeSymlink != 0 && pointsAtFolder(path):
				links = append(links, path)
				return nil
			case d.IsDir():
				realDir, realErr := filepath.EvalSymlinks(filepath.Join(root, filepath.FromSlash(path)))
				if realErr != nil {
					return nil
				}
				if visited[realDir] {
					return fs.SkipDir
				}
				visited[realDir] = true
				return nil
			case !strings.EqualFold(filepath.Ext(path), ext):
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			file := gurps.LibraryFile{Library: lib.Key(), Path: path}
			seen[file] = true
			cached, ok := grantFileItems[file]
			if !ok || !cached.modTime.Equal(info.ModTime()) {
				cached = grantFileCache{
					modTime: info.ModTime(),
					items:   load(fileSystem, file),
				}
				grantFileItems[file] = cached
			}
			stamps = append(stamps, grantFileStamp{file: file, modTime: cached.modTime})
			caches = append(caches, cached)
			return nil
		}
		_ = fs.WalkDir(fileSystem, ".", walk) //nolint:errcheck // A file that can't be read is skipped
		// The walks of the symlinked folders may find more.
		for i := 0; i < len(links); i++ {
			_ = fs.WalkDir(fileSystem, links[i], walk) //nolint:errcheck // A file that can't be read is skipped
		}
	}
	// The files of this kind no longer found, deleted, renamed or under a library that was rekeyed or removed, are
	// forgotten.
	for file := range grantFileItems {
		if !seen[file] && strings.EqualFold(filepath.Ext(file.Path), ext) {
			delete(grantFileItems, file)
		}
	}
	if last, ok := grantItemLists[t]; ok && slices.EqualFunc(last.stamps, stamps, grantFileStamp.equal) {
		return last.items
	}
	var items []*grantLibraryItem
	for _, cached := range caches {
		items = append(items, cached.items...)
	}
	slices.SortStableFunc(items, func(a, b *grantLibraryItem) int {
		return xstrings.NaturalCmp(a.name, b.name, true)
	})
	grantItemLists[t] = grantItemList{stamps: stamps, items: items}
	return items
}

// grantItemLoader returns the extension of the library files holding items of the kind a grant of the given type
// gives, and what reads the items a grant can give from one.
func grantItemLoader(t feature.Type) (ext string, load func(fs.FS, gurps.LibraryFile) []*grantLibraryItem) {
	switch t {
	case feature.GrantSkill:
		return gurps.SkillsExt, loadGrantItems(gurps.NewSkillsFromFile)
	case feature.GrantSpell:
		return gurps.SpellsExt, loadGrantItems(gurps.NewSpellsFromFile)
	default:
		return gurps.TraitsExt, loadGrantItems(gurps.NewTraitsFromFile)
	}
}

func loadGrantItems[T gurps.Node[T]](load func(fs.FS, string) ([]T, error)) func(fs.FS, gurps.LibraryFile) []*grantLibraryItem {
	return func(fileSystem fs.FS, from gurps.LibraryFile) []*grantLibraryItem {
		list, err := load(fileSystem, from.Path)
		if err != nil {
			return nil
		}
		return collectGrantItems(from, list)
	}
}

// grantFilterWords returns the lowercase words of the text.
func grantFilterWords(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

// grantItemMatches returns true if every word of the query matches a word of the item's name: as its start, or within
// a typo or two of it, more being allowed the longer the word is.
func grantItemMatches(query, words []string) bool {
	for _, q := range query {
		if !slices.ContainsFunc(words, func(w string) bool { return grantWordMatches(q, w) }) {
			return false
		}
	}
	return true
}

func grantWordMatches(q, w string) bool {
	if strings.HasPrefix(w, q) {
		return true
	}
	var allowed int
	switch n := len([]rune(q)); {
	case n < 4:
		return false
	case n < 8:
		allowed = 1
	default:
		allowed = 2
	}
	if editDistance(q, w) <= allowed {
		return true
	}
	// A query that is the start of a longer word may also hold a typo.
	if r := []rune(w); len(r) > len([]rune(q)) {
		return editDistance(q, string(r[:len([]rune(q))])) <= allowed
	}
	return false
}

// editDistance returns the number of single-rune insertions, deletions, substitutions and swaps of neighbors that
// turn a into b.
func editDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev2 := make([]int, len(rb)+1)
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
			if i > 1 && j > 1 && ra[i-1] == rb[j-2] && ra[i-2] == rb[j-1] {
				cur[j] = min(cur[j], prev2[j-2]+1)
			}
		}
		prev2, prev, cur = prev, cur, prev2
	}
	return prev[len(rb)]
}

// grantItemCellFactory draws a library item of the grant picker as its name, points or difficulty, and library file.
type grantItemCellFactory struct{}

func (f *grantItemCellFactory) CellHeight() float32 {
	return xmath.Ceil(unison.FieldFont.LineHeight() + 2*unison.StdVSpacing)
}

func (f *grantItemCellFactory) CreateCell(_ unison.Paneler, element any, _ int, foreground, _ unison.Ink, _, _ bool) unison.Paneler {
	item, ok := element.(*grantLibraryItem)
	if !ok {
		return unison.NewPanel()
	}
	row := unison.NewPanel()
	row.SetBorder(unison.NewEmptyBorder(geom.Insets{
		Top: unison.StdVSpacing, Left: unison.StdHSpacing,
		Bottom: unison.StdVSpacing, Right: unison.StdHSpacing,
	}))
	row.SetLayout(&unison.FlexLayout{Columns: 3, HSpacing: unison.StdHSpacing * 2})
	// The second and third columns have fixed widths, so that they line up from row to row.
	widths := []float32{0, grantPickerPointsWidth, grantPickerFileWidth}
	for i, text := range []string{item.name, item.points, item.file} {
		label := unison.NewLabel()
		label.Font = unison.FieldFont
		label.OnBackgroundInk = foreground
		label.SetTitle(text)
		data := &unison.FlexLayoutData{VAlign: align.Middle}
		if i == 0 {
			data.HAlign = align.Fill
			data.HGrab = true
		} else {
			data.SizeHint = geom.NewSize(widths[i], 0)
			data.MinSize = data.SizeHint
			if i == 1 {
				label.HAlign = align.End
			}
		}
		label.SetLayoutData(data)
		row.AddChild(label)
	}
	return row
}
