// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
	uncheck "github.com/richardwilkes/unison/enums/check"
)

// putInChoice makes the parent a template choice holding the child.
func putInChoice[T gurps.Node[T]](parent, child T, tp *gurps.TemplatePicker) {
	tp.Type = picker.Count
	parent.SetChildren([]T{child})
	SetParents([]T{child}, parent)
}

// TestPickedAsAUnitCheckBox verifies that a group in a template choice offers Picked as a Unit, checked until it is
// set to be picked from separately, and that nothing else does.
func TestPickedAsAUnitCheckBox(t *testing.T) {
	c := check.New(t)
	const title = "Picked as a Unit"
	template := newTestTemplateDockable("Template", gurps.NewTemplate())
	traitChoice, group := gurps.NewTrait(nil, nil, true), gurps.NewTrait(nil, nil, true)
	putInChoice(traitChoice, group, &traitChoice.TemplatePicker)
	e, content := buildEditorContent(Rebuildable(template), group, initTraitEditor)
	boxes := checkBoxesTitled(content, title)
	c.Equal(1, len(boxes), "a trait group in a template choice offers it")
	if len(boxes) == 1 {
		c.Equal(uncheck.On, boxes[0].State, "checked by default")
		boxes[0].Click()
		e.editorData.ApplyTo(group)
		c.True(group.PickSeparately, "unchecking it picks from the group separately")
	}
	_, content = buildEditorContent(Rebuildable(template), group, initTraitEditor)
	if boxes = checkBoxesTitled(content, title); len(boxes) == 1 {
		c.Equal(uncheck.Off, boxes[0].State, "a group already picked from separately opens unchecked")
	}
	shown := func(owner Rebuildable, target *gurps.Trait) int {
		_, content = buildEditorContent(owner, target, initTraitEditor)
		return len(checkBoxesTitled(content, title))
	}
	inner := gurps.NewTrait(nil, nil, true)
	group.Children = []*gurps.Trait{inner}
	SetParents(group.Children, group)
	c.Equal(1, shown(template, inner), "a group in one picked from separately offers it")
	group.PickSeparately = false
	c.Equal(0, shown(template, inner), "a group in one picked as a unit must not offer it")
	c.Equal(0, shown(newTestSheetForTemplate(t), group), "a sheet must not offer it")
	c.Equal(0, shown(NewTraitTableDockable("traits"+gurps.TraitsExt, nil), group), "a library must not offer it")
	group.ContainerType = container.AlternativeAbilities
	c.Equal(0, shown(template, group), "alternative abilities must not offer it")
	outside := gurps.NewTrait(nil, nil, true)
	c.Equal(0, shown(template, outside), "a group outside a choice must not offer it")

	skillChoice, skills := gurps.NewSkill(nil, nil, true), gurps.NewSkill(nil, nil, true)
	putInChoice(skillChoice, skills, &skillChoice.TemplatePicker)
	_, content = buildEditorContent(Rebuildable(template), skills, initSkillEditor)
	c.Equal(1, len(checkBoxesTitled(content, title)), "a skill container in a choice offers it")
	spellChoice, spells := gurps.NewSpell(nil, nil, true), gurps.NewSpell(nil, nil, true)
	putInChoice(spellChoice, spells, &spellChoice.TemplatePicker)
	_, content = buildEditorContent(Rebuildable(template), spells, initSpellEditor)
	c.Equal(1, len(checkBoxesTitled(content, title)), "a spell container in a choice offers it")
	eqpChoice, eqp := gurps.NewEquipment(nil, nil, true), gurps.NewEquipment(nil, nil, true)
	eqpChoice.ContainerType, eqp.ContainerType = eqcontainer.Group, eqcontainer.Group
	putInChoice(eqpChoice, eqp, &eqpChoice.TemplatePicker)
	_, content = buildEditorContent(Rebuildable(template), eqp, initEquipmentEditor(false))
	c.Equal(1, len(checkBoxesTitled(content, title)), "an equipment group in a choice offers it")
	eqp.ContainerType = eqcontainer.Container
	_, content = buildEditorContent(Rebuildable(template), eqp, initEquipmentEditor(false))
	c.Equal(0, len(checkBoxesTitled(content, title)), "a physical container must not offer it")
}
