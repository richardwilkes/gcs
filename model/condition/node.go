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
	"hash"

	"github.com/richardwilkes/toolbox/v2/xhash"
)

// The wire type keys of the node types.
const (
	typeGroup   = "group"
	typeCheck   = "check"
	typeConst   = "const"
	typeCount   = "count"
	typeCompare = "compare"
	typeScript  = "script"
)

// The kind bytes that start each node's hash.
const (
	hashGroup uint8 = iota
	hashCheck
	hashConst
	hashCount
	hashCompare
	hashScript
	hashBroken
	hashNil = uint8(255)
)

// Node is one node of a condition tree. The set of node types is closed: Group, Check, Const, Count, Compare and
// Script.
type Node interface {
	// Clone returns a deep copy: new pointer targets, copied slices and cloned node fields.
	Clone() Node
	// Hash writes the canonical view of the node into the hasher.
	Hash(h hash.Hash)
	// FillWithNameableKeys adds the nameable keys of every qualifier, script name and script text in the subtree,
	// including those under where, only_when and operands.
	FillWithNameableKeys(m, existing map[string]string)
	validate(v *validator)
	walk(w *walker)
	eval(e *evaluator, f frame) Result
}

// CloneNode returns a deep copy of n, or nil if n is nil.
func CloneNode(n Node) Node {
	if n == nil {
		return nil
	}
	return n.Clone()
}

// HashNode writes n into the hasher, or 255 if n is nil.
func HashNode(h hash.Hash, n Node) {
	if n == nil {
		xhash.Num8(h, hashNil)
		return
	}
	n.Hash(h)
}

// Ref holds a single node slot: only_when, where, or the root of a tree. It is a named field, so a Ref never satisfies
// Node itself.
type Ref struct {
	Node Node
}

// IsZero returns true for a nil node and for an empty group, which evaluates exactly like nil.
func (r Ref) IsZero() bool {
	if r.Node == nil {
		return true
	}
	g, ok := r.Node.(*Group)
	return ok && len(g.Children) == 0 && g.OnlyWhen.IsZero() && g.N == 0
}

func (r Ref) clone() Ref {
	return Ref{Node: CloneNode(r.Node)}
}

func (r Ref) hash(h hash.Hash) {
	if r.IsZero() {
		xhash.Num8(h, hashNil)
		return
	}
	r.Node.Hash(h)
}

func (r Ref) fillWithNameableKeys(m, existing map[string]string) {
	if r.Node != nil {
		r.Node.FillWithNameableKeys(m, existing)
	}
}

// Nodes holds the children of a group. A nil element only comes from code: it is written as null, hashed as 255 and
// reported by Validate.
type Nodes []Node

func (n Nodes) clone() Nodes {
	if n == nil {
		return nil
	}
	list := make(Nodes, len(n))
	for i, one := range n {
		list[i] = CloneNode(one)
	}
	return list
}

func (n Nodes) hash(h hash.Hash) {
	xhash.Num64(h, len(n))
	for _, one := range n {
		HashNode(h, one)
	}
}

func (n Nodes) fillWithNameableKeys(m, existing map[string]string) {
	for _, one := range n {
		if one != nil {
			one.FillWithNameableKeys(m, existing)
		}
	}
}
