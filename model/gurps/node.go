// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
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
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xreflect"
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

// Node defines the methods required of nodes in our tables.
//
// This interface is a "constraint" and cannot be used as a type param, just as a constraint.
// In practice this just means that it can only be used inside the square brackets of generic methods/types.
type Node[T Node[T]] interface {
	// These are the limited set of types that can be nodes. New node types *must* be added here
	*ConditionalModifier | *Equipment | *EquipmentModifier | *Note | *Skill | *Spell | *Trait | *TraitModifier | *Weapon

	// These are the interface mathods each of the above types must implement as a minimum
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

func assertNode[T Node[T]]() {}

// RawPointsAdjuster defines methods for nodes that can have their raw points adjusted must implement.
type RawPointsAdjuster interface {
	Container() bool
	RawPoints() fxp.Int
	SetRawPoints(points fxp.Int) bool
}

// SkillAdjustmentProvider defines methods for nodes that can have their skill level adjusted must implement.
type SkillAdjustmentProvider interface {
	RawPointsAdjuster
	IncrementSkillLevel()
	DecrementSkillLevel()
}

// EditorData defines the methods required of editor data.
type EditorData[T Node[T]] interface {
	// CopyFrom copies the corresponding data from the node into this editor data.
	CopyFrom(T)
	// ApplyTo copies the editor data into the provided node.
	ApplyTo(T)
}

func assertEditorData[T EditorData[N], N Node[N]]() {}

// EntityFromNode returns the owning entity of the node, or nil.
func EntityFromNode[T Node[T]](node T) *Entity {
	if xreflect.IsNil(node) {
		return nil
	}
	owner := node.DataOwner()
	if xreflect.IsNil(owner) {
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
				if !slices.ContainsFunc(tags, func(s string) bool { return strings.EqualFold(s, part) }) {
					tags = append(tags, part)
				}
			}
		}
	}
	return tags
}

// PropagateNodeNoteClosedState propagates the note closed state from one node to another.
func PropagateNodeNoteClosedState[T Node[T]](from, to T) {
	SetClosedState("N:"+string(to.ID()), IsClosed("N:"+string(from.ID())))
}
