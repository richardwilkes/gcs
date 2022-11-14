/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package model

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/richardwilkes/toolbox/i18n"
)

// ContainerKeyPostfix is the key postfix used to identify containers.
const ContainerKeyPostfix = "_container"

// ContainerBase holds the type and ID of the data.
type ContainerBase[T NodeTypes] struct {
	ID       uuid.UUID `json:"id"`
	Type     string    `json:"type"`
	IsOpen   bool      `json:"open,omitempty"`     // Container only
	Children []T       `json:"children,omitempty"` // Container only
	parent   T
}

func newContainerBase[T NodeTypes](typeKey string, isContainer bool) ContainerBase[T] {
	if isContainer {
		typeKey += ContainerKeyPostfix
	}
	return ContainerBase[T]{
		ID:     NewUUID(),
		Type:   typeKey,
		IsOpen: isContainer,
	}
}

// UUID returns the UUID of this data.
func (c *ContainerBase[T]) UUID() uuid.UUID {
	return c.ID
}

// Container returns true if this is a container.
func (c *ContainerBase[T]) Container() bool {
	return strings.HasSuffix(c.Type, ContainerKeyPostfix)
}

func (c *ContainerBase[T]) kind(base string) string {
	if c.Container() {
		return fmt.Sprintf(i18n.Text("%s Container"), base)
	}
	return base
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
