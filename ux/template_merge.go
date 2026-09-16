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
	"maps"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// entityTechLevel returns the tech level of the entity that owns the table, or an empty string if there is none. This
// is the value substituted for an empty tech level when merging rows, matching what the table providers do on drop.
func entityTechLevel[T gurps.Node[T]](table *unison.Table[*Node[T]]) string {
	if dataOwnerProvider := table.Ancestor[gurps.DataOwnerProvider](); !xreflect.IsNil(dataOwnerProvider) {
		if dataOwner := dataOwnerProvider.DataOwner(); !xreflect.IsNil(dataOwner) {
			if entity := dataOwner.OwningEntity(); entity != nil {
				return entity.Profile.TechLevel
			}
		}
	}
	return ""
}

// resolveEmptyTechLevel replaces an empty (but present) tech level with defaultTechLevel. This is the substitution
// performed when a row is dropped onto a sheet and when a template's rows are merged into one, so both must agree.
func resolveEmptyTechLevel(item gurps.TechLevelProvider, defaultTechLevel string) {
	if item.RequiresTL() && item.TL() == "" {
		item.SetTL(defaultTechLevel)
	}
}

// sameTechLevel reports whether two rows' tech levels match for merge purposes: both absent, or both present and equal.
func sameTechLevel(a, b gurps.TechLevelProvider) bool {
	if a.RequiresTL() != b.RequiresTL() {
		return false
	}
	return !a.RequiresTL() || a.TL() == b.TL()
}

// MergeableNode is the set of row types whose identical rows are merged by folding their points together: skills and
// spells.
type MergeableNode[T gurps.Node[T]] interface {
	gurps.Node[T]
	gurps.TechLevelProvider
	RawPoints() fxp.Int
	SetRawPoints(points fxp.Int) bool
	NameableReplacements() map[string]string
}

// MergeAddedRows folds the points of the newly-added, currently-selected top-level rows (and any rows nested within
// them, such as the contents of an added container) into identical skill or spell rows already present in the sheet,
// removing the now-redundant new rows, so that dragging or copying a skill or spell that already exists on the sheet
// adds to its points rather than creating a duplicate. It must be called only after tech levels and nameables have
// been resolved on the new rows, since the match includes both. Only skills and spells are affected.
func MergeAddedRows[T gurps.Node[T]](table *unison.Table[*Node[T]]) {
	switch t := any(table).(type) {
	case *unison.Table[*Node[*gurps.Skill]]:
		mergeNewlySelectedRows(t)
	case *unison.Table[*Node[*gurps.Spell]]:
		mergeNewlySelectedRows(t)
	}
}

func mergeNewlySelectedRows[T MergeableNode[T]](table *unison.Table[*Node[T]]) {
	sel := table.CopySelectionMap()
	if len(sel) == 0 {
		return
	}
	roots := table.RootRows()
	existing := make([]T, 0, len(roots))
	incoming := make([]T, 0, len(roots))
	for _, node := range roots {
		if sel[node.ID()] {
			incoming = append(incoming, node.Data())
		} else {
			existing = append(existing, node.Data())
		}
	}
	if len(existing) == 0 || len(incoming) == 0 {
		return
	}
	newSel := make(map[tid.TID]bool)
	surviving := mergePoints(existing, incoming, entityTechLevel(table), newSel)
	if len(newSel) == 0 {
		return // Nothing merged, so leave the table untouched.
	}
	// Comparing len(surviving) to len(incoming) is not a valid way to detect that nothing merged: a merged row nested
	// inside an added container is pruned from the container's data without changing the top-level count, and the view
	// must still be refreshed or the pruned row remains visible until the next rebuild.
	survivingSet := make(map[T]bool, len(surviving))
	for _, data := range surviving {
		survivingSet[data] = true
	}
	newRoots := make([]*Node[T], 0, len(roots))
	for _, node := range roots {
		if sel[node.ID()] {
			if !survivingSet[node.Data()] {
				continue // This newly-added row was merged into an existing one, so drop it.
			}
			newSel[node.ID()] = true // Keep the surviving new rows selected alongside the rows they merged into.
			// Rows nested inside this row may have been pruned by the merge, so discard any cached child nodes to
			// force them to be rebuilt from the updated data.
			node.RefreshChildren()
		}
		newRoots = append(newRoots, node)
	}
	table.SetRootRows(newRoots)
	table.SetSelectionMap(newSel)
	rebuildAsModified(table.AncestorOrSelf[Rebuildable](), true)
}

// mergePoints folds the points of each incoming row into a matching row, returning the incoming rows that had no match
// (and should therefore be added as new rows). A match requires an identical hash, the same nameable replacements, and
// the same tech level. Since neither the tech level nor the replacements are part of the hash, several rows can share
// a hash, so all candidates for a hash are considered. An incoming row can match either an existing row or an earlier
// incoming row, so a template containing two identical entries collapses them into one.
//
// An incoming row with an empty (but non-nil) tech level has it resolved to defaultTechLevel first, mirroring the
// substitution performed on drop by the skills and spells providers. Without this, a template applied a second time
// would compare the incoming empty tech level against the already-resolved tech level of the existing row, fail to
// match, and add a duplicate row instead of merging.
func mergePoints[T MergeableNode[T]](existing, incoming []T, defaultTechLevel string, selMap map[tid.TID]bool) []T {
	byHash := make(map[uint64][]T)
	gurps.Traverse(func(item T) bool {
		hash := gurps.Hash64(item)
		byHash[hash] = append(byHash[hash], item)
		return false
	}, true, true, existing...)
	pruneMap := make(map[T]bool)
	gurps.Traverse(func(item T) bool {
		resolveEmptyTechLevel(item, defaultTechLevel)
		hash := gurps.Hash64(item)
		matched := false
		for _, candidate := range byHash[hash] {
			if !maps.Equal(candidate.NameableReplacements(), item.NameableReplacements()) ||
				!sameTechLevel(candidate, item) {
				continue
			}
			pruneMap[item] = true
			candidate.SetRawPoints(candidate.RawPoints() + item.RawPoints())
			selMap[candidate.ID()] = true
			matched = true
			break
		}
		if !matched {
			// Register this surviving incoming row so that any later identical incoming row merges into it.
			byHash[hash] = append(byHash[hash], item)
		}
		return false
	}, true, true, incoming...)
	for item := range pruneMap {
		isItem := func(other T) bool { return other == item }
		if parent := item.Parent(); xreflect.IsNil(parent) {
			incoming = slices.DeleteFunc(incoming, isItem)
		} else {
			parent.SetChildren(slices.DeleteFunc(parent.NodeChildren(), isItem))
		}
	}
	return incoming
}

// mergeRowsFor bridges from a caller generic over any node type, which only knows the row type once it has switched on
// the table's concrete type, to mergeRows, which needs that concrete type. The conversions cannot fail once the switch
// has matched, but rows are returned untouched should one somehow not hold up.
func mergeRowsFor[T MergeableNode[T], U gurps.Node[U]](table *unison.Table[*Node[T]], existing, rows []*Node[U], selMap map[tid.TID]bool) []*Node[U] {
	if existingNodes, ok := any(existing).([]*Node[T]); ok {
		if rowNodes, ok2 := any(rows).([]*Node[T]); ok2 {
			if merged, ok3 := any(mergeRows(table, existingNodes, rowNodes, selMap)).([]*Node[U]); ok3 {
				return merged
			}
		}
	}
	return rows
}

// mergeRows folds the points of the incoming rows into matching existing rows (see mergePoints) and returns fresh
// nodes for the incoming rows that survived, ready to be added to the table.
func mergeRows[T MergeableNode[T]](table *unison.Table[*Node[T]], existing, rows []*Node[T], selMap map[tid.TID]bool) []*Node[T] {
	surviving := mergePoints(
		ExtractNodeDataFromList(existing),
		ExtractNodeDataFromList(rows),
		entityTechLevel(table),
		selMap,
	)
	replacements := make([]*Node[T], 0, len(surviving))
	for _, item := range surviving {
		replacements = append(replacements, NewNode(table, nil, item, true))
	}
	return replacements
}
