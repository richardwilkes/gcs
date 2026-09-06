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
	"hash"
	"io/fs"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xhash"
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
// The type union in the body makes this a constraint, not an ordinary interface: it may only be used as a type
// parameter's constraint or embedded in another constraint, never as a type. Generic code takes a T Node[T] type
// parameter and works with T directly.
type Node[T Node[T]] interface {
	// These are the limited set of types that can be nodes. New node types *must* be added here
	*ConditionalModifier | *Equipment | *EquipmentModifier | *Note | *Skill | *Spell | *Trait | *TraitModifier | *Weapon

	// These are the interface methods each of the above types must implement as a minimum
	fmt.Stringer
	Openable
	Hashable
	nameable.Applier
	Clone(from LibraryFile, owner DataOwner, newParent T, mode CloneMode) T
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

// NodeSyncData holds the sync data that every named node shares: the fields a library copy is expected to keep in step
// with its source. The per-type SyncData types are either aliases of it or structs that embed it, which json/v2
// inlines, so the on-disk format is the same as if each declared the fields itself.
type NodeSyncData struct {
	Name             string   `json:"name,omitzero"`
	PageRef          string   `json:"reference,omitzero"`
	PageRefHighlight string   `json:"reference_highlight,omitzero"`
	LocalNotes       string   `json:"local_notes,omitzero"`
	Tags             []string `json:"tags,omitempty"`
}

func (n *NodeSyncData) hash(h hash.Hash) {
	xhash.StringWithLen(h, n.Name)
	xhash.StringWithLen(h, n.PageRef)
	xhash.StringWithLen(h, n.PageRefHighlight)
	xhash.StringWithLen(h, n.LocalNotes)
	hashStrings(h, n.Tags)
}

// RawPointsAdjuster interface for objects that can have their raw points adjusted.
type RawPointsAdjuster interface {
	Container() bool
	RawPoints() fxp.Int
	SetRawPoints(points fxp.Int) bool
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

// listData is the on-disk form of a standalone list file: the data version and the top-level rows.
type listData[T any] struct {
	Version int `json:"version"`
	Rows    []T `json:"rows"`
}

// loadRows loads the rows of a standalone list file. Each top-level row is given a nil data owner, which is what
// attaches the weapons and modifiers throughout the tree, since SetDataOwner recurses into the children on its own.
// Containers must not be skipped along the way: they carry their own weapons and modifiers, which would otherwise never
// be attached.
func loadRows[T Node[T]](fileSystem fs.FS, filePath string) ([]T, error) {
	var data listData[T]
	if err := jio.LoadVersionedFile(fileSystem, filePath, &data, &data.Version); err != nil {
		return nil, err
	}
	SetDataOwnerAll(nil, data.Rows)
	return data.Rows, nil
}

// saveRows writes the rows to the file as a standalone list stamped with the current data version.
func saveRows[T any](filePath string, rows []T) error {
	return jio.SaveToFile(filePath, &listData[T]{Version: jio.CurrentDataVersion, Rows: rows})
}

// fixupLegacyTID gives a node written before TIDs existed -- its ID is a UUID, or missing altogether -- a fresh TID
// whose kind is derived from the legacy "type" string, which was the kind's name with containerKeyPostfix appended for
// a container. It returns true when a fixup was made, which is also the signal to carry over the legacy "open" flag:
// the open state of a container is tracked by TID nowadays, so it could not have been stored for a node without one.
func fixupLegacyTID(id *tid.TID, legacyType string, kindFor func(container bool) byte) bool {
	if tid.IsValid(*id) {
		return false
	}
	*id = tid.MustNewTID(kindFor(strings.HasSuffix(legacyType, containerKeyPostfix)))
	return true
}

// migrateLegacyText fills in text from its legacy counterpart, which held embedded expressions rather than scripts, when
// the node was written before the field text belongs to existed.
func migrateLegacyText(text *string, legacy string) {
	if *text == "" && legacy != "" {
		*text = EmbeddedExprToScript(legacy)
	}
}

// finishNodeUnmarshal applies the fix-ups every node type needs at the end of its UnmarshalJSONFrom: folding the legacy
// categories into the tags and sorting them, pointing the children back at their parent, and opening the node when the
// legacy "open" flag asked for it (see fixupLegacyTID).
func finishNodeUnmarshal[T Node[T]](node T, tags *[]string, legacyCategories []string, open bool) {
	*tags = convertOldCategoriesToTags(*tags, legacyCategories)
	slices.Sort(*tags)
	if node.Container() {
		for _, child := range node.NodeChildren() {
			child.SetParent(node)
		}
	}
	if open {
		SetNodeOpen(node, true)
	}
}

func convertOldCategoriesToTags(tags, categories []string) []string {
	if categories == nil {
		return tags
	}
	for _, one := range categories {
		for part := range strings.SplitSeq(one, "/") {
			if part = strings.TrimSpace(part); part != "" {
				if !slices.ContainsFunc(tags, func(s string) bool { return strings.EqualFold(s, part) }) {
					tags = append(tags, part)
				}
			}
		}
	}
	return tags
}

// modifierHolder is what cloneModifiers needs from the trait or piece of equipment whose modifiers are being cloned.
type modifierHolder interface {
	DataOwnerProvider
	GetSource() Source
}

// cloneModifiers clones the modifiers held by a trait or piece of equipment for a copy of that holder -- a clone of it,
// or the editor data staged from or committed back to it -- and hands each copy to attach, which points it at the
// holder. It returns nil when there is nothing to clone.
//
// The LibraryFile for each clone must come from the holder rather than from the modifier being cloned. This covers the
// case where the source data *is* the authoritative source and therefore carries no source information of its own: the
// modifier's Source.LibraryFile is empty and AdjustSource won't set source data on the copy, whereas the holder's
// Source.LibraryFile holds the already-adjusted source for the holder's copy, so it always has the correct library path.
//
// Background: when GCS clones an item from one library into another location (as opposed to duplicating in place), it
// passes the *source* library as the first argument to Clone. That path, combined with the IDs from the source nodes,
// is what builds the `source` values for the clone.
func cloneModifiers[M Node[M]](modifiers []M, holder modifierHolder, mode CloneMode, attach func(M)) []M {
	if len(modifiers) == 0 {
		return nil
	}
	from := holder.GetSource().LibraryFile
	owner := holder.DataOwner()
	var noParent M
	result := make([]M, 0, len(modifiers))
	for _, one := range modifiers {
		cloned := one.Clone(from, owner, noParent, mode)
		attach(cloned)
		result = append(result, cloned)
	}
	return result
}

// PropagateNodeNoteClosedState propagates the note closed state from one node to another.
func PropagateNodeNoteClosedState[T Node[T]](from, to T) {
	SetClosedState("N:"+string(to.ID()), IsClosed("N:"+string(from.ID())))
}

// CloneMode controls how Clone treats the resulting node's ID and Source relative to the node being cloned.
type CloneMode int

const (
	// Reference creates a fresh ID and, if the node being cloned has no Source of its own, anchors the
	// result's Source to reference it. Used when copying a node into a different list: "Copy to Character
	// Sheet"/"Copy to Template", drag-and-drop from a library into a sheet or template, dropping a
	// modifier onto a row.
	Reference CloneMode = iota
	// Duplicate create a fresh ID but copies the node being cloned's Source verbatim, even when empty. Used
	// for "Duplicate": the result is a sibling within the same list, not a new reference to the original.
	Duplicate
	// Copy preserves the ID of the node being cloned and copies its Source verbatim, even when empty.
	// Used for in-memory working copies never inserted into any list: the editor's CopyFrom/ApplyTo
	// staging round trip, and scratch clones for live preview/calculation.
	Copy
)
