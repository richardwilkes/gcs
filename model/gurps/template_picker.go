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
	"github.com/richardwilkes/toolbox/v2/xreflect"
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

// String returns a description of the picker. A weight is described in the default sheet settings' units, since only a
// template keeps a picker made by weight and a template has no entity; see StringWithUnits for describing it in
// another's.
func (t TemplatePicker) String() string {
	return t.StringWithUnits(SheetSettingsFor(nil).DefaultWeightUnits)
}

// StringWithUnits returns a description of the picker, describing a weight in the given units.
func (t TemplatePicker) StringWithUnits(units fxp.WeightUnit) string {
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
	case picker.Value:
		return fmt.Sprintf(i18n.Text("Pick %s worth"), t.Qualifier.Compare.AltDescribeWith("$"+t.Qualifier.Qualifier.Comma()))
	case picker.Weight:
		return fmt.Sprintf(i18n.Text("Pick %s in weight"),
			t.Qualifier.Compare.AltDescribeWith(units.Format(fxp.Weight(t.Qualifier.Qualifier))))
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

// TemplatePickerQualifierMinimum returns the least a picker of the given type may ask for. Only points may be less than
// nothing, since a disadvantage costs negative points; a count, a value or a weight never can be.
func TemplatePickerQualifierMinimum(pickerType picker.Type) fxp.Int {
	if pickerType == picker.Points {
		return fxp.Min
	}
	return 0
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

// ClearTemplatePickerData removes the template picker data from the nodes and their children, along with every flag to
// pick a group separately. A node that loses its picker data also loses its source, since only a template may hold
// picker data and a template is never a source.
func ClearTemplatePickerData[T Node[T]](nodes ...T) {
	Traverse(func(node T) bool {
		clearPickSeparately(node)
		if IsTemplateChoiceContainer(node) {
			_, data := any(node).(TemplatePickerProvider).TemplatePickerData() //nolint:errcheck // IsTemplateChoiceContainer checked this
			*data = TemplatePicker{}
			node.ClearSource()
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
	// clearTemplateChoiceContainerExclusions removes the data templateChoiceContainerExclusions describes, along with
	// anything else a choice container doesn't use.
	clearTemplateChoiceContainerExclusions()
}

// NormalizeTemplateChoiceContainers strips every template choice container among the nodes and their children of
// everything a choice container doesn't use, its source included. A choice container dissolves into the options
// chosen from it when the template is applied, and nothing on it reaches the sheet, so anything else it holds would
// only be hidden from its editor while still affecting the template. A template does this whenever rows arrive in it,
// whether loaded from its file or transferred from elsewhere.
func NormalizeTemplateChoiceContainers[T Node[T]](nodes ...T) {
	Traverse(func(node T) bool {
		if IsTemplateChoiceContainer(node) {
			normalizeTemplateChoiceContainer(node)
		}
		return false
	}, false, false, nodes...)
}

// normalizeTemplateChoiceContainer strips a single choice container of everything a choice container doesn't use.
func normalizeTemplateChoiceContainer[T Node[T]](node T) {
	if tc, ok := any(node).(templateChoiceConvertible); ok {
		tc.clearTemplateChoiceContainerExclusions()
	}
	clearPickSeparately(node)
	node.ClearSource()
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
	if !node.GetSource().IsZero() {
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
	_, data := any(node).(templateChoiceConvertible).TemplatePickerData() //nolint:errcheck // CanConvertToTemplateChoiceContainer checked this
	*data = newTemplateChoicePicker()
	normalizeTemplateChoiceContainer(node)
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

// groupConvertible is implemented by the node types whose containers come in a kind that holds things as an item in
// its own right, such as a backpack, as well as a group, which only organizes what it holds, so that the one may be
// converted into the other.
type groupConvertible interface {
	// CanConvertToGroup returns true if this container can be converted to a group.
	CanConvertToGroup() bool
	// GroupConversionLosses returns a description of each piece of data this container holds that a group can't.
	GroupConversionLosses() []string
	// ConvertToGroup converts this container to a group, discarding what GroupConversionLosses describes.
	ConvertToGroup()
}

// CanConvertToGroupContainer returns true if the node is a container that can be converted to a group: a template
// choice container or a modifier choice, which loses its choice, or a container of a kind that holds things in its own
// right.
func CanConvertToGroupContainer[T Node[T]](node T) bool {
	if IsTemplateChoiceContainer(node) || IsModifierChoice(node) {
		return true
	}
	gc, ok := any(node).(groupConvertible)
	return ok && gc.CanConvertToGroup()
}

// GroupConversionLosses returns a description of each piece of data the node would lose by being converted to a group,
// other than a template choice container's or modifier choice's choice.
func GroupConversionLosses[T Node[T]](node T) []string {
	if IsTemplateChoiceContainer(node) || IsModifierChoice(node) {
		return nil
	}
	if gc, ok := any(node).(groupConvertible); ok {
		return gc.GroupConversionLosses()
	}
	return nil
}

// ConvertToGroupContainer converts the node to a group, if it can be (see CanConvertToGroupContainer).
func ConvertToGroupContainer[T Node[T]](node T) {
	if IsTemplateChoiceContainer(node) {
		ConvertFromTemplateChoiceContainer(node)
		return
	}
	if IsModifierChoice(node) {
		ConvertFromModifierChoice(node)
		return
	}
	if gc, ok := any(node).(groupConvertible); ok && gc.CanConvertToGroup() {
		gc.ConvertToGroup()
	}
}

// PickerMeasureRange returns the span of what the node counts toward a template choice made by the given picker type.
// Everything counts as one toward a choice made by count. Toward one made by points, a skill or spell counts by its raw
// points, inside a container as much as on its own, since a choice counts what is being bought rather than what the
// destination sheet's bonuses make of it, and the rows are already owned by that sheet by the time the choice is made.
// A trait counts by its adjusted points, the only cost a trait has. Toward one made by value or weight, equipment counts
// by its extended value or weight, its quantity and contents included. Either way a container accounts for any choices
// it presents, including the exact ones, which are worth what they ask for rather than what their children add up to.
// When prompted, open modifier choices are costed as the modifier prompt will see them, even on a sheet. A node taken
// (which may be nil) reports counts as preconfigured, for the choices the nodes inside inherit from it when told so.
func PickerMeasureRange[T Node[T]](node T, pickerType picker.Type, prompted bool, taken func(T, bool) bool) NumericRange {
	if xreflect.IsNil(node) {
		return NumericRangeOf(0)
	}
	view := promptedView(taken)
	view.prompted = prompted
	switch pickerType {
	case picker.Count:
		return NumericRangeOf(fxp.One)
	case picker.Points:
		if rp, ok := any(node).(interface{ RawPointsRange() NumericRange }); ok {
			return rp.RawPointsRange()
		}
		if trait, ok := any(node).(*Trait); ok {
			return trait.pointsRange(nil, view)
		}
	case picker.Value:
		if eqp, ok := any(node).(*Equipment); ok {
			return equipmentValue().seenAs(view).rangeOf(eqp, eqp.Quantity)
		}
	case picker.Weight:
		if eqp, ok := any(node).(*Equipment); ok {
			units := SheetSettingsFor(EntityFromNode(node)).DefaultWeightUnits
			return equipmentWeight(false, units).seenAs(view).rangeOf(eqp, eqp.Quantity)
		}
	default:
	}
	return NumericRangeOf(0)
}
