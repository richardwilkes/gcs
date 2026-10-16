// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

import (
	"slices"
	"strconv"
	"strings"
)

// SlotKind identifies the kind of step a Slot takes from a node. It is never persisted.
type SlotKind byte

// Possible SlotKind values.
const (
	// SlotChild steps into a group's child at Index.
	SlotChild SlotKind = iota
	// SlotOnlyWhen steps into a group's only_when guard.
	SlotOnlyWhen
	// SlotWhere steps into the where of a count or an aggregate operand.
	SlotWhere
	// SlotLeft steps into a compare's left operand.
	SlotLeft
	// SlotRight steps into a compare's right operand.
	SlotRight
	// SlotSum steps into the operand at Index of a sum.
	SlotSum
)

// Slot is one step of a Path.
type Slot struct {
	Kind  SlotKind
	Index int // child and sum only
}

// String implements fmt.Stringer.
func (s Slot) String() string {
	switch s.Kind {
	case SlotChild:
		return "children[" + strconv.Itoa(s.Index) + "]"
	case SlotOnlyWhen:
		return "only_when"
	case SlotWhere:
		return "where"
	case SlotLeft:
		return "left"
	case SlotRight:
		return "right"
	case SlotSum:
		return "sum[" + strconv.Itoa(s.Index) + "]"
	default:
		return "?"
	}
}

// Path addresses a node, or an operand, relative to the node a walk or validation started from. Nodes have no parent
// pointers, so the editor addresses them by path.
type Path []Slot

// String implements fmt.Stringer, joining the slots with dots.
func (p Path) String() string {
	parts := make([]string, len(p))
	for i, s := range p {
		parts[i] = s.String()
	}
	return strings.Join(parts, ".")
}

// Walk calls fn for n and then, in pre-order, for every node beneath it: only_when, children, where, and the wheres
// inside operands and sums. When fn returns false, the nodes beneath the one it was called for are skipped. Nil nodes,
// including nil children, are not visited. The path passed to fn is reused by the walk, so clone it to keep it.
func Walk(n Node, fn func(path Path, n Node) bool) {
	if n != nil {
		(&walker{fn: fn}).visit(n)
	}
}

type walker struct {
	fn   func(path Path, n Node) bool
	path Path
}

func (w *walker) visit(n Node) {
	if w.fn(slices.Clip(w.path), n) {
		n.walk(w)
	}
}

func (w *walker) child(s Slot, n Node) {
	if n != nil {
		w.path = append(w.path, s)
		w.visit(n)
		w.path = w.path[:len(w.path)-1]
	}
}

func (w *walker) ref(kind SlotKind, r Ref) {
	if !r.IsZero() {
		w.child(Slot{Kind: kind}, r.Node)
	}
}

func (w *walker) operand(s Slot, o *Operand) {
	w.path = append(w.path, s)
	o.walk(w)
	w.path = w.path[:len(w.path)-1]
}
