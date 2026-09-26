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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// TemplatePickerProvider provides access to the valid picker types and the picker data.
type TemplatePickerProvider interface {
	// TemplatePickerData returns the valid picker types and a non-nil pointer to the TemplatePicker.
	TemplatePickerData() ([]picker.Type, *TemplatePicker)
}

// TemplatePickerNode is a constraint for a Node that implements the TemplatePickerProvider interface
type TemplatePickerNode[T TemplatePickerNode[T]] interface {
	Node[T]
	TemplatePickerProvider
}

// assertTemplatePickerNode causes a compile-time constraint validation
func assertTemplatePickerNode[T TemplatePickerNode[T]]() {}

// TemplatePicker holds the data necessary to allow a template choice to be made.
type TemplatePicker struct {
	Type      picker.Type     `json:"type"`
	Qualifier criteria.Number `json:"qualifier,omitzero"`
}

// IsZero implements json.isZero.
func (t TemplatePicker) IsZero() bool {
	return t.Type == picker.NotApplicable
}

func (t TemplatePicker) String() string {
	if t.IsZero() {
		return ""
	}
	switch t.Type {
	case picker.Count:
		return fmt.Sprintf(i18n.Text("Pick %s"), t.Qualifier.AltString())
	case picker.Points:
		points := i18n.Text("points")
		if t.Qualifier.Qualifier == fxp.One {
			points = i18n.Text("point")
		}
		return fmt.Sprintf(i18n.Text("Pick %s %s worth"), t.Qualifier.AltString(), points)
	default:
		return ""
	}
}

// Hash writes this object's contents into the hasher.
func (t TemplatePicker) Hash(h hash.Hash) {
	xhash.Num8(h, t.Type)
	if t.Type != picker.NotApplicable {
		t.Qualifier.Hash(h)
	}
}

// newTemplateChoicePicker returns the picker data a new template choice container starts with: pick exactly one of its
// children.
func newTemplateChoicePicker() TemplatePicker {
	return TemplatePicker{
		Type: picker.Count,
		Qualifier: criteria.Number{NumberData: criteria.NumberData{
			Compare:   criteria.EqualsNumber,
			Qualifier: fxp.One,
		}},
	}
}

// IsTemplateChoiceContainer returns true if the node is a template choice container, that is, a container with non-zero
// template picker data. Such a container only declares the choice and holds the options for it, dissolving into the
// options chosen when the template is applied.
func IsTemplateChoiceContainer[T Node[T]](node T) bool {
	if !node.Container() {
		return false
	}
	if tpp, ok := any(node).(TemplatePickerProvider); ok {
		_, data := tpp.TemplatePickerData()
		return !data.IsZero()
	}
	return false
}

// HasTemplatePickerData returns true if any node or their child has non-zero template picker data
func HasTemplatePickerData[T Node[T]](nodes ...T) bool {
	var hasPickerData bool
	Traverse(func(node T) bool {
		hasPickerData = IsTemplateChoiceContainer(node)
		return hasPickerData
	}, false, false, nodes...)
	return hasPickerData
}

// ClearTemplatePickerData removes the template picker data from the nodes and their children. A node that loses its
// picker data also loses its source, since only a template may hold picker data and a template is never a source.
func ClearTemplatePickerData[T Node[T]](nodes ...T) {
	Traverse(func(node T) bool {
		if node.Container() {
			if tpp, ok := any(node).(TemplatePickerProvider); ok {
				if _, data := tpp.TemplatePickerData(); !data.IsZero() {
					*data = TemplatePicker{}
					node.ClearSource()
				}
			}
		}
		return false
	}, false, false, nodes...)
}

// templateChoiceConvertible is implemented by the node types whose containers can be converted to and from template
// choice containers.
type templateChoiceConvertible interface {
	TemplatePickerProvider
	// canBecomeTemplateChoiceContainer returns true if this container is of a kind that may become a choice container.
	canBecomeTemplateChoiceContainer() bool
	// templateChoiceContainerExclusions returns a description of each piece of data this container holds that a choice
	// container can't.
	templateChoiceContainerExclusions() []string
	// clearTemplateChoiceContainerExclusions removes the data templateChoiceContainerExclusions describes.
	clearTemplateChoiceContainerExclusions()
	// normalizeTemplateChoiceContainer brings this choice container into line with what a choice container may be.
	normalizeTemplateChoiceContainer()
}

// normalizeTemplateChoiceContainers brings every template choice container among the nodes and their children into
// line with what a choice container may be.
func normalizeTemplateChoiceContainers[T Node[T]](nodes ...T) {
	Traverse(func(node T) bool {
		if IsTemplateChoiceContainer(node) {
			if tc, ok := any(node).(templateChoiceConvertible); ok {
				tc.normalizeTemplateChoiceContainer()
			}
		}
		return false
	}, false, false, nodes...)
}

// CanConvertToTemplateChoiceContainer returns true if the node is a container that can be converted to a template
// choice container.
func CanConvertToTemplateChoiceContainer[T Node[T]](node T) bool {
	if !node.Container() || IsTemplateChoiceContainer(node) {
		return false
	}
	tc, ok := any(node).(templateChoiceConvertible)
	return ok && tc.canBecomeTemplateChoiceContainer()
}

// TemplateChoiceConversionLosses returns a description of each piece of data the node would lose by being converted to
// a template choice container. A choice container never has a source, so a source is among them.
func TemplateChoiceConversionLosses[T Node[T]](node T) []string {
	tc, ok := any(node).(templateChoiceConvertible)
	if !ok {
		return nil
	}
	losses := tc.templateChoiceContainerExclusions()
	if node.GetSource() != (Source{}) {
		losses = append(losses, i18n.Text("library source"))
	}
	return losses
}

// ConvertToTemplateChoiceContainer converts the node to a template choice container, if it can be, discarding the data
// TemplateChoiceConversionLosses describes. The new choice asks for exactly one of the container's children.
func ConvertToTemplateChoiceContainer[T Node[T]](node T) {
	if !CanConvertToTemplateChoiceContainer(node) {
		return
	}
	tc := any(node).(templateChoiceConvertible) //nolint:errcheck // CanConvertToTemplateChoiceContainer checked this
	tc.clearTemplateChoiceContainerExclusions()
	_, data := tc.TemplatePickerData()
	*data = newTemplateChoicePicker()
	node.ClearSource()
}

// ConvertFromTemplateChoiceContainer converts a template choice container back into a plain container, discarding its
// choice.
func ConvertFromTemplateChoiceContainer[T Node[T]](node T) {
	if !IsTemplateChoiceContainer(node) {
		return
	}
	_, data := any(node).(TemplatePickerProvider).TemplatePickerData() //nolint:errcheck // IsTemplateChoiceContainer checked this
	*data = TemplatePicker{}
}
