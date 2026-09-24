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

// mergeableNode is the set of row types whose identical rows are merged by folding their points together: skills and
// spells.
type mergeableNode[T gurps.Node[T]] interface {
	gurps.Node[T]
	gurps.TechLevelProvider
	RawPoints() fxp.Int
	SetRawPoints(points fxp.Int) bool
	NameableReplacements() map[string]string
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
func mergePoints[T mergeableNode[T]](existing, incoming []T, defaultTechLevel string, selMap map[tid.TID]bool) []T {
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

// mergeIncoming folds the points of the incoming rows into identical skill or spell rows already present in the table
// (see mergePoints), returning the incoming rows that survived, and adding the rows that absorbed points to selMap. Rows
// of any other type are returned untouched.
func mergeIncoming[T gurps.Node[T]](table *unison.Table[*Node[T]], incoming []T, selMap map[tid.TID]bool) []T {
	switch t := any(table).(type) {
	case *unison.Table[*Node[*gurps.Skill]]:
		return mergeIncomingFor(t, incoming, selMap)
	case *unison.Table[*Node[*gurps.Spell]]:
		return mergeIncomingFor(t, incoming, selMap)
	default:
		return incoming
	}
}

// mergeIncomingFor bridges from mergeIncoming, which is generic over any node type and only knows the row type once it
// has switched on the table's concrete type, to mergePoints, which needs that concrete type. The conversions cannot
// fail once the switch has matched, but the rows are returned untouched should one somehow not hold up.
func mergeIncomingFor[M mergeableNode[M], T gurps.Node[T]](table *unison.Table[*Node[M]], incoming []T, selMap map[tid.TID]bool) []T {
	rows, ok := any(incoming).([]M)
	if !ok {
		return incoming
	}
	existing := ExtractNodeDataFromList(table.RootRows())
	if merged, ok2 := any(mergePoints(existing, rows, entityTechLevel(table), selMap)).([]T); ok2 {
		return merged
	}
	return incoming
}
