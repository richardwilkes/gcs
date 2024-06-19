// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
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

	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
)

// ContainerKeyPostfix is the key postfix used to identify containers.
const ContainerKeyPostfix = "_container"

// ContainerBase holds the type and ID of the data.
type ContainerBase[T NodeTypes] struct {
	LocalID       tid.TID        `json:"local_id"`
	DataSource    string         `json:"data_source,omitempty"`
	DataSourceID  tid.TID        `json:"data_source_id,omitempty"`
	Type          string         `json:"type"`
	ThirdParty    map[string]any `json:"third_party,omitempty"`
	parent        T
	itemKind      byte
	containerKind byte
	ContainerBaseContainerOnly[T]
}

// ContainerBaseContainerOnly holds the ContainerBase data that is only applicable to containers.
type ContainerBaseContainerOnly[T NodeTypes] struct {
	Children []T  `json:"children,omitempty"`
	IsOpen   bool `json:"open,omitempty"`
}

func newContainerBase[T NodeTypes](parent T, itemKind, containerKind byte, container bool) ContainerBase[T] {
	var kind byte
	if container {
		kind = containerKind
	} else {
		kind = itemKind
	}
	return ContainerBase[T]{
		LocalID:       tid.MustNewTID(kind),
		parent:        parent,
		itemKind:      itemKind,
		containerKind: containerKind,
		ContainerBaseContainerOnly: ContainerBaseContainerOnly[T]{
			IsOpen: container,
		},
	}
}

// GetLocalID returns the local ID of this data.
func (c *ContainerBase[T]) GetLocalID() tid.TID {
	return c.LocalID
}

// Container returns true if this is a container.
func (c *ContainerBase[T]) Container() bool {
	return tid.IsKindAndValid(c.LocalID, c.containerKind)
}

func (c *ContainerBase[T]) kind(base string) string {
	if c.Container() {
		return fmt.Sprintf(i18n.Text("%s Container"), base)
	}
	return base
}

// GetType returns the type.
func (c *ContainerBase[T]) GetType() string {
	return c.Type
}

// SetType sets the type.
func (c *ContainerBase[T]) SetType(t string) {
	c.Type = t
}

// Open returns true if this node is currently open.
func (c *ContainerBase[T]) Open() bool {
	return c.IsOpen && c.Container()
}

// SetOpen sets the current open state for this node.
func (c *ContainerBase[T]) SetOpen(open bool) {
	c.IsOpen = open && c.Container()
}

// Parent returns the parent.
func (c *ContainerBase[T]) Parent() T {
	return c.parent
}

// SetParent sets the parent.
func (c *ContainerBase[T]) SetParent(parent T) {
	c.parent = parent
}

// HasChildren returns true if this node has children.
func (c *ContainerBase[T]) HasChildren() bool {
	return c.Container() && len(c.Children) > 0
}

// NodeChildren returns the children of this node, if any.
func (c *ContainerBase[T]) NodeChildren() []T {
	return c.Children
}

// SetChildren sets the children of this node.
func (c *ContainerBase[T]) SetChildren(children []T) {
	c.Children = children
}

func (c *ContainerBase[T]) clearUnusedFields() {
	if !c.Container() {
		c.Children = nil
		c.IsOpen = false
	}
}
