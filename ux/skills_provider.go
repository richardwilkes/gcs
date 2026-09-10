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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

var _ TableProvider[*gurps.Skill] = &skillsProvider{}

type skillsProvider struct {
	listProvider[*gurps.Skill]
	provider gurps.SkillListProvider
}

// NewSkillsProvider creates a new table provider for skills.
func NewSkillsProvider(provider gurps.SkillListProvider, forPage bool) TableProvider[*gurps.Skill] {
	p := &skillsProvider{provider: provider}
	p.listProvider = listProvider[*gurps.Skill]{
		dataOwner:    provider,
		list:         provider.SkillList,
		setList:      provider.SetSkillList,
		columnIDs:    p.ColumnIDs,
		headerData:   gurps.SkillsHeaderData,
		newItem:      gurps.NewSkill,
		edit:         func(owner Rebuildable, item *gurps.Skill) { EditSkill(owner, item) },
		forPage:      forPage,
		filterKey:    gurps.ListFilterKeyForExtension(gurps.SkillsExt),
		filterFields: gurps.SkillFilterFields,
	}
	return p
}

func (p *skillsProvider) RefKey() string {
	return gurps.BlockSkillsKey
}

func (p *skillsProvider) DragKey() *uti.DataType {
	return skillDragKey
}

func (p *skillsProvider) DragSVG() *unison.SVG {
	return svg.GCSSkills
}

func (p *skillsProvider) ProcessDropData(_, to *unison.Table[*Node[*gurps.Skill]]) {
	if dataOwnerProvider := to.Ancestor[gurps.DataOwnerProvider](); !xreflect.IsNil(dataOwnerProvider) {
		if dataOwner := dataOwnerProvider.DataOwner(); !xreflect.IsNil(dataOwner) {
			if entity := dataOwner.OwningEntity(); entity != nil {
				for _, row := range to.SelectedRows(true) {
					gurps.Traverse(func(skill *gurps.Skill) bool {
						resolveEmptyTechLevel(skill, entity.Profile.TechLevel)
						skill.UpdateLevel()
						return false
					}, false, true, row.Data())
				}
			}
		}
	}
}

func (p *skillsProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Skill"), i18n.Text("Skills")
}

func (p *skillsProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 7)
	if showSwitchColumn(p.forPage, p.provider, p.RootData()) {
		columnIDs = append(columnIDs, gurps.SkillSwitchColumn)
	}
	columnIDs = append(columnIDs, gurps.SkillDescriptionColumn)
	if p.forPage {
		if _, ok := p.provider.(*gurps.Entity); ok {
			columnIDs = append(
				columnIDs,
				gurps.SkillLevelColumn,
				gurps.SkillRelativeLevelColumn,
			)
		}
		columnIDs = append(columnIDs, gurps.SkillPointsColumn)
	} else {
		columnIDs = append(
			columnIDs,
			gurps.SkillDifficultyColumn,
			gurps.SkillTagsColumn,
		)
	}
	return p.appendReferenceColumns(columnIDs, gurps.SkillReferenceColumn, gurps.SkillLibSrcColumn)
}

func (p *skillsProvider) HierarchyColumnID() int {
	return gurps.SkillDescriptionColumn
}

func (p *skillsProvider) ExcessWidthColumnID() int {
	return gurps.SkillDescriptionColumn
}

// CreateItem adds the alternate variant, a technique, to the skill and skill container the shared implementation
// creates.
func (p *skillsProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.Skill]], variant ItemVariant) {
	switch variant {
	case NoItemVariant, ContainerItemVariant:
		p.listProvider.CreateItem(owner, table, variant)
	case AlternateItemVariant:
		p.createItem(owner, table, gurps.NewTechnique(p.DataOwner(), nil, ""))
	default:
		errs.Log(errs.New("unhandled variant"), "variant", int(variant))
	}
}

func (p *skillsProvider) ContextMenuItems() []ContextMenuItem {
	return AppendDefaultContextMenuItems([]ContextMenuItem{
		contextMenuItemFor(newSkillAction),
		contextMenuItemFor(newSkillContainerAction),
		contextMenuItemFor(newTechniqueAction),
	})
}
