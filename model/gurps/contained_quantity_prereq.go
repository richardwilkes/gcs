/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Prereq = &ContainedQuantityPrereq{}

// ContainedQuantityPrereq holds a prerequisite for an equipment contained quantity.
type ContainedQuantityPrereq struct {
	Parent            *PrereqList     `json:"-"`
	Type              PrereqType      `json:"type"`
	Has               bool            `json:"has"`
	QualifierCriteria NumericCriteria `json:"qualifier,omitempty"`
}

// NewContainedQuantityPrereq creates a new ContainedQuantityPrereq.
func NewContainedQuantityPrereq() *ContainedQuantityPrereq {
	return &ContainedQuantityPrereq{
		Type: ContainedQuantityPrereqType,
		QualifierCriteria: NumericCriteria{
			NumericCriteriaData: NumericCriteriaData{
				Compare:   AtMostNumber,
				Qualifier: fxp.One,
			},
		},
		Has: true,
	}
}

// PrereqType implements Prereq.
func (c *ContainedQuantityPrereq) PrereqType() PrereqType {
	return c.Type
}

// ParentList implements Prereq.
func (c *ContainedQuantityPrereq) ParentList() *PrereqList {
	return c.Parent
}

// Clone implements Prereq.
func (c *ContainedQuantityPrereq) Clone(parent *PrereqList) Prereq {
	clone := *c
	clone.Parent = parent
	return &clone
}

// FillWithNameableKeys implements Prereq.
func (c *ContainedQuantityPrereq) FillWithNameableKeys(_ map[string]string) {
}

// ApplyNameableKeys implements Prereq.
func (c *ContainedQuantityPrereq) ApplyNameableKeys(_ map[string]string) {
}

// Satisfied implements Prereq.
func (c *ContainedQuantityPrereq) Satisfied(_ *Entity, exclude any, tooltip *xio.ByteBuffer, prefix string, _ *bool) bool {
	satisfied := false
	if eqp, ok := exclude.(*Equipment); ok {
		if satisfied = !eqp.Container(); !satisfied {
			var qty fxp.Int
			for _, child := range eqp.Children {
				qty += child.Quantity
			}
			satisfied = c.QualifierCriteria.Matches(qty)
		}
	}
	if !c.Has {
		satisfied = !satisfied
	}
	if !satisfied && tooltip != nil {
		tooltip.WriteString(prefix)
		tooltip.WriteString(HasText(c.Has))
		tooltip.WriteString(i18n.Text(" a contained quantity which "))
		tooltip.WriteString(c.QualifierCriteria.String())
	}
	return satisfied
}
