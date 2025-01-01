// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"fmt"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/txt"
)

// DataOwner defines the methods required of data owners.
type DataOwner interface {
	OwningEntity() *Entity
	SourceMatcher() *SrcMatcher
	WeightUnit() fxp.WeightUnit
}

// DataOwnerProvider provides a way to retrieve a (possibly nil) data owner.
type DataOwnerProvider interface {
	DataOwner() DataOwner
}

// NodeTypes is a constraint that defines the types that may be nodes.
type NodeTypes interface {
	*ConditionalModifier | *Equipment | *EquipmentModifier | *Note | *Skill | *Spell | *Trait | *TraitModifier | *Weapon
	nameable.Applier
	fmt.Stringer
}

// Node defines the methods required of nodes in our tables.
type Node[T NodeTypes] interface {
	fmt.Stringer
	Openable
	Hashable
	nameable.Applier
	Clone(from LibraryFile, owner DataOwner, newParent T, preserveID bool) T
	GetSource() Source
	ClearSource()
	SyncWithSource()
	DataOwner() DataOwner
	SetDataOwner(owner DataOwner)
	Kind() string
	Parent() T
	SetParent(parent T)
	HasChildren() bool
	NodeChildren() []T
	SetChildren(children []T)
	Enabled() bool
	CellData(columnID int, data *CellData)
}

// RawPointsAdjuster defines methods for nodes that can have their raw points adjusted must implement.
type RawPointsAdjuster[T NodeTypes] interface {
	Node[T]
	RawPoints() fxp.Int
	SetRawPoints(points fxp.Int) bool
}

// SkillAdjustmentProvider defines methods for nodes that can have their skill level adjusted must implement.
type SkillAdjustmentProvider[T NodeTypes] interface {
	RawPointsAdjuster[T]
	IncrementSkillLevel()
	DecrementSkillLevel()
}

// EditorData defines the methods required of editor data.
type EditorData[T NodeTypes] interface {
	// CopyFrom copies the corresponding data from the node into this editor data.
	CopyFrom(T)
	// ApplyTo copies the editor data into the provided node.
	ApplyTo(T)
}

// AsNode converts a T to a Node[T]. This shouldn't require these hoops, but Go generics (as of 1.19) fails to compile
// otherwise.
func AsNode[T NodeTypes](in T) Node[T] {
	if node, ok := any(in).(Node[T]); ok {
		return node
	}
	return nil
}

// EntityFromNode returns the owning entity of the node, or nil.
func EntityFromNode[T NodeTypes](node Node[T]) *Entity {
	if toolbox.IsNil(node) {
		return nil
	}
	owner := node.DataOwner()
	if toolbox.IsNil(owner) {
		return nil
	}
	return owner.OwningEntity()
}

func convertOldCategoriesToTags(tags, categories []string) []string {
	if categories == nil {
		return tags
	}
	for _, one := range categories {
		parts := strings.Split(one, "/")
		for _, part := range parts {
			if part = strings.TrimSpace(part); part != "" {
				if !txt.CaselessSliceContains(tags, part) {
					tags = append(tags, part)
				}
			}
		}
	}
	return tags
}
