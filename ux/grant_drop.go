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
	"slices"
	"strconv"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/promptstep"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/mod"
)

// grantItemsFrom returns the kind of grant that the rows of a table drag make, and the items to grant, one for each
// row, or for each item within a container among them, left out as the picker leaves them out.
func grantItemsFrom(data any) (feature.Type, []*grantLibraryItem) {
	switch d := data.(type) {
	case *unison.TableDragData[*Node[*gurps.Trait]]:
		return feature.GrantTrait, grantItemsFromRows(d.Table, d.Rows)
	case *unison.TableDragData[*Node[*gurps.Skill]]:
		return feature.GrantSkill, grantItemsFromRows(d.Table, d.Rows)
	case *unison.TableDragData[*Node[*gurps.Spell]]:
		return feature.GrantSpell, grantItemsFromRows(d.Table, d.Rows)
	default:
		return feature.Unknown, nil
	}
}

func grantItemsFromRows[T gurps.Node[T]](table *unison.Table[*Node[T]], rows []*Node[T]) []*grantLibraryItem {
	nodes := make([]T, 0, len(rows))
	for _, row := range rows {
		nodes = append(nodes, row.Data())
	}
	return collectGrantItems(libraryFileFromTable(table), nodes)
}

// initGrantDrop lets trait, skill and spell rows dragged from a table be dropped on the panel, each becoming a grant of
// it: onto an empty grant of the same kind, filling it, or between the rows, adding new ones.
func (p *featuresPanel) initGrantDrop() {
	canAccept := p.CanAcceptDropCallback
	p.CanAcceptDropCallback = func(di drag.Info) bool {
		if t, _ := p.grantDrag(di); t != feature.Unknown {
			return true
		}
		return canAccept(di)
	}
	entered, updated, dropped := p.DragEnteredCallback, p.DragUpdatedCallback, p.DropCallback
	over := func(di drag.Info, where geom.Point, mods mod.Modifiers, next func(drag.Info, geom.Point, mod.Modifiers) drag.Op) drag.Op {
		t, items := p.grantDrag(di)
		if t == feature.Unknown {
			return next(di, where, mods)
		}
		p.ScrollRectIntoView(geom.NewRect(where.X, where.Y-16, 1, 1))
		p.ScrollRectIntoView(geom.NewRect(where.X, where.Y+16, 1, 1))
		op := drag.None
		target, at := p.grantDropAt(where, t)
		if len(items) != 0 {
			op = drag.Copy
		} else {
			target = nil
		}
		if target != p.dropTarget || at != p.dropWhere {
			p.dropTarget, p.dropWhere = target, at
			p.MarkForRedraw()
		}
		flushDragFeedback(p.AsPanel())
		return op
	}
	p.DragEnteredCallback = func(di drag.Info, where geom.Point, mods mod.Modifiers) drag.Op {
		return over(di, where, mods, entered)
	}
	p.DragUpdatedCallback = func(di drag.Info, where geom.Point, mods mod.Modifiers) drag.Op {
		return over(di, where, mods, updated)
	}
	p.DropCallback = func(di drag.Info, where geom.Point, mods mod.Modifiers) bool {
		t, items := p.grantDrag(di)
		if t == feature.Unknown {
			return dropped(di, where, mods)
		}
		target, at := p.grantDropAt(where, t)
		p.dragExit()
		if len(items) == 0 {
			return false
		}
		p.dropGrants(target, at, t, items)
		return true
	}
}

// grantDrag returns the kind and items of the grants the rows being dragged from a table would make, or an unknown
// kind when the drag isn't of rows a grant can give or the panel can't take grants.
func (p *featuresPanel) grantDrag(di drag.Info) (feature.Type, []*grantLibraryItem) {
	if !p.takesGrants() || !slices.ContainsFunc([]string{traitDragKey.UTI, skillDragKey.UTI, spellDragKey.UTI},
		di.HasDataType) {
		return feature.Unknown, nil
	}
	t, items := grantItemsFrom(draggedTableData)
	return t, p.grantable(items)
}

// grantable returns the items the panel's owner can be given as grants: all but the owner itself, or a copy of it,
// which it would otherwise grant a copy of itself, and likewise the items whose editors the panel's editor was opened
// from, such as the trait a modifier belongs to, whose grants its modifiers' are.
func (p *featuresPanel) grantable(items []*grantLibraryItem) []*grantLibraryItem {
	owners := p.grantOwners()
	return slices.DeleteFunc(slices.Clone(items), func(item *grantLibraryItem) bool {
		return slices.ContainsFunc(owners, func(owner any) bool {
			return item.node == owner || sameGrantItem(item.node, owner)
		})
	})
}

// grantOwners returns the panel's owner along with the items edited by the editors the panel's editor was opened from,
// one from another.
func (p *featuresPanel) grantOwners() []any {
	owners := []any{p.owner}
	for owned := unison.AncestorOrSelf[Owned](p); owned != nil; {
		owner := owned.Owner()
		if xreflect.IsNil(owner) {
			break
		}
		if e, ok := owner.(interface{ editTarget() any }); ok {
			owners = append(owners, e.editTarget())
		}
		next, ok := owner.AsPanel().Self.(Owned)
		if !ok {
			break
		}
		owned = next
	}
	return owners
}

// sameGrantItem returns true if a and b are copies of the same trait, skill or spell.
func sameGrantItem(a, b any) bool {
	switch one := a.(type) {
	case *gurps.Trait:
		other, ok := b.(*gurps.Trait)
		return ok && gurps.Hash64(one) == gurps.Hash64(other) && gurps.SameItem(one, other)
	case *gurps.Skill:
		other, ok := b.(*gurps.Skill)
		return ok && gurps.Hash64(one) == gurps.Hash64(other) && gurps.SameItem(one, other)
	case *gurps.Spell:
		other, ok := b.(*gurps.Spell)
		return ok && gurps.Hash64(one) == gurps.Hash64(other) && gurps.SameItem(one, other)
	default:
		return false
	}
}

// takesGrants returns true if grants can be added to the panel: it isn't read-only, nor in the editor of a grant's
// item, which can't grant anything.
func (p *featuresPanel) takesGrants() bool {
	return !p.readOnly && !inGrantItemEditor(p)
}

// inGrantItemEditor returns true if the panel is within the editor of a grant's item, or within an editor opened from
// that one, such as for one of the item's modifiers, whose grants would go unheeded along with the item's.
func inGrantItemEditor(panel unison.Paneler) bool {
	for owned := unison.AncestorOrSelf[Owned](panel); owned != nil; {
		owner := owned.Owner()
		if _, nested := owner.(*grantItemOwner); nested {
			return true
		}
		if xreflect.IsNil(owner) {
			return false
		}
		next, ok := owner.AsPanel().Self.(Owned)
		if !ok {
			return false
		}
		owned = next
	}
	return false
}

// grantDropAt returns where grants of the given kind dropped at where would go: into the row of an empty grant of that
// kind, before or after another row, by which half of it where is in, or anywhere else, after the last row.
func (p *featuresPanel) grantDropAt(where geom.Point, t feature.Type) (target *unison.Panel, at int) {
	target, path, at := p.rowAt(where)
	if target != nil {
		if g, ok := (*p.features)[p.index(path)].(*gurps.Grant); ok && g.Type == t && g.Key() == "" {
			return target, dropInto
		}
		return target, at
	}
	for _, child := range p.Children() {
		if _, isRow := child.ClientData()[featureDropKey].(string); isRow {
			target = child
		}
	}
	if target != nil {
		return target, dropAfter
	}
	// With no rows, or with the panel collapsed, the grants are shown going over what stands in for the rows.
	return cmp.Or(p.FindRefKey(featureEmptyKey), p.FindRefKey(sectionSummaryKey)), dropInto
}

// dropGrants makes grants of the items where grantDropAt said they go.
func (p *featuresPanel) dropGrants(target *unison.Panel, at int, t feature.Type, items []*grantLibraryItem) {
	index := len(*p.features)
	if target != nil {
		if path, isRow := target.ClientData()[featureDropKey].(string); isRow {
			index = p.index(path)
			if at == dropAfter {
				index++
			}
			if at == dropInto {
				g, ok := (*p.features)[index].(*gurps.Grant)
				if !ok {
					return
				}
				rest := newGrants(t, items[1:])
				p.sentenceRows.restructure(i18n.Text("Choose Granted Item"), path+keyMore, "", func() string {
					g.SetItem(items[0].node, items[0].from)
					p.grantPicking = ""
					*p.features = slices.Insert(*p.features, index+1, rest...)
					p.open = path
					return grantFocusKey(path, g)
				})
				return
			}
		}
	}
	p.addGrants(t, items, index)
}

// addGrants adds a grant of each of the items at the index, expanding the panel and opening the first with its level
// field focused, as choosing an item in the picker does.
func (p *featuresPanel) addGrants(t feature.Type, items []*grantLibraryItem, index int) {
	if p.collapse.collapsed {
		p.collapse.toggle()
	}
	grants := newGrants(t, items)
	first, ok := grants[0].(*gurps.Grant)
	if !ok {
		return
	}
	path := strconv.Itoa(index)
	title := i18n.Text("Add Grant")
	if len(grants) > 1 {
		title = i18n.Text("Add Grants")
	}
	p.sentenceRows.restructure(title, featureAddKey, "", func() string {
		*p.features = slices.Insert(*p.features, index, grants...)
		p.open = path
		return grantFocusKey(path, first)
	})
}

// newGrants returns a grant of the given kind for each of the items.
func newGrants(t feature.Type, items []*grantLibraryItem) []gurps.Feature {
	grants := make([]gurps.Feature, 0, len(items))
	for _, item := range items {
		g := gurps.NewGrant(t)
		g.SetItem(item.node, item.from)
		grants = append(grants, g)
	}
	return grants
}

// grantDestination is an open editor whose features can take grants.
type grantDestination struct {
	dockable unison.Dockable
	panel    *featuresPanel
}

// String names the item being edited, along with where it is: the editor's title, which for an editor opened from
// another names only the item, so the items of the editors it was opened from follow, then the title of the outermost,
// which names the document.
func (d grantDestination) String() string {
	where := d.dockable.Title()
	for current := d.dockable; ; {
		owned, ok := current.(Owned)
		if !ok {
			break
		}
		outer, isEditor := owned.Owner().(nestedEditor)
		if !isEditor {
			break
		}
		where = i18n.Text("%s, within %s", where, outer.targetName())
		if _, deeper := outer.Owner().(nestedEditor); !deeper {
			where = i18n.Text("%s, %s", where, outer.Title())
		}
		current = outer
	}
	return i18n.Text("%s (%s)", d.panel.owner.String(), where)
}

// nestedEditor is what an editor opened from another editor offers for naming where it is.
type nestedEditor interface {
	unison.Dockable
	Owned
	targetName() string
}

// grantDestinations returns the open editors whose features can take grants of the items, by name, leaving out those
// that can take none of them (see grantable).
func grantDestinations(items []*grantLibraryItem) []grantDestination {
	var result []grantDestination
	for _, d := range AllDockables() {
		if _, isEditor := d.(interface{ editor() editorDockable }); !isEditor {
			continue
		}
		var walk func(panel *unison.Panel)
		walk = func(panel *unison.Panel) {
			if p, ok := panel.Self.(*featuresPanel); ok {
				if p.takesGrants() && len(p.grantable(items)) != 0 {
					result = append(result, grantDestination{dockable: d, panel: p})
				}
				return
			}
			for _, child := range panel.Children() {
				walk(child)
			}
		}
		walk(d.AsPanel())
	}
	slices.SortFunc(result, func(a, b grantDestination) int { return xstrings.NaturalCmp(a.String(), b.String(), true) })
	return result
}

// installAddAsGrantHandler installs the Add as Grant command on target for the table, which is enabled when the table
// holds traits, skills or spells. It is the keyboard counterpart of dropping the selected rows on an editor's features.
func installAddAsGrantHandler[T gurps.Node[T]](target unison.Paneler, table *unison.Table[*Node[T]]) {
	selection := func() (feature.Type, []*grantLibraryItem) {
		return grantItemsFrom(&unison.TableDragData[*Node[T]]{Table: table, Rows: table.SelectedRows(true)})
	}
	target.AsPanel().InstallCmdHandlers(AddAsGrantItemID,
		func(_ any) bool {
			_, items := selection()
			return len(items) != 0 && len(grantDestinations(items)) != 0
		},
		func(_ any) {
			t, items := selection()
			addAsGrant(t, items)
		})
}

// addAsGrant adds a grant of each of the items the editor can take (see grantable) to the features of the open editor
// the user picks, when there is more than one, and brings that editor to the front.
func addAsGrant(t feature.Type, items []*grantLibraryItem) {
	choices := grantDestinations(items)
	if len(choices) == 0 {
		return
	}
	dest := choices[0]
	if len(choices) > 1 {
		list := newChoiceList[grantDestination](false)
		list.Append(choices...)
		list.Select(false, 0)
		op := promptOperation{name: addAsGrantAction.Title, description: i18n.Text("Adding a grant to an editor")}
		if !showListQuestionDialog(op.at(promptstep.Destination), i18n.Text("Choose the editor to add to:"), list) {
			return
		}
		picked := pickFromList(list, choices)
		if len(picked) == 0 {
			return
		}
		dest = picked[0]
	}
	if wnd := dest.dockable.AsPanel().Window(); wnd != nil {
		ActivateDockable(dest.dockable)
		wnd.ToFront()
	}
	dest.panel.addGrants(t, dest.panel.grantable(items), len(*dest.panel.features))
}
