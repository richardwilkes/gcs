// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.
package uxtest

import (
	"github.com/richardwilkes/unison"
)

// PanelsOfType returns every panel of the given type within the subtree rooted at root, in pre-order.
func PanelsOfType[T any](root *unison.Panel) []T {
	var found []T
	root.HasInSelfOrDescendants(func(p *unison.Panel) bool {
		if match, ok := p.Self.(T); ok {
			found = append(found, match)
		}
		return false
	})
	return found
}

// PanelsMatching returns every panel within the subtree rooted at root that keep accepts, in pre-order.
func PanelsMatching(root *unison.Panel, keep func(*unison.Panel) bool) []*unison.Panel {
	var found []*unison.Panel
	root.HasInSelfOrDescendants(func(p *unison.Panel) bool {
		if keep(p) {
			found = append(found, p)
		}
		return false
	})
	return found
}

// FirstPanelOfType returns the first panel of the given type within the subtree rooted at root, in pre-order. Editors
// build their sub-panels several levels down and hand back no references to most of them, so a test that wants to drive
// one has to go looking for it.
func FirstPanelOfType[T any](root *unison.Panel) (T, bool) {
	var found T
	ok := root.HasInSelfOrDescendants(func(p *unison.Panel) bool {
		match, matched := p.Self.(T)
		if matched {
			found = match
		}
		return matched
	})
	return found, ok
}

// linkLabel is a label whose text may hold links the keyboard can be on, such as ux.TextLabel, which this package
// cannot name; the CurrentLink method is what tells one from the rest of the panels that have a String method.
type linkLabel interface {
	CurrentLink() int
	String() string
}

// LabelTexts returns the text of every label within the subtree rooted at root, in pre-order, whether it is a
// unison.Label or a link-aware label such as ux.TextLabel.
func LabelTexts(root *unison.Panel) []string {
	var texts []string
	root.HasInSelfOrDescendants(func(p *unison.Panel) bool {
		switch label := p.Self.(type) {
		case *unison.Label:
			texts = append(texts, label.String())
		case linkLabel:
			texts = append(texts, label.String())
		}
		return false
	})
	return texts
}
