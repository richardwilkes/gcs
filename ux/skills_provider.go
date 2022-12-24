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

package ux

import (
	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/maps"
)

var (
	skillListColMap = map[int]int{
		0: model.SkillDescriptionColumn,
		1: model.SkillDifficultyColumn,
		2: model.SkillTagsColumn,
		3: model.SkillReferenceColumn,
	}
	entitySkillPageColMap = map[int]int{
		0: model.SkillDescriptionColumn,
		1: model.SkillLevelColumn,
		2: model.SkillRelativeLevelColumn,
		3: model.SkillPointsColumn,
		4: model.SkillReferenceColumn,
	}
	skillPageColMap = map[int]int{
		0: model.SkillDescriptionColumn,
		1: model.SkillPointsColumn,
		2: model.SkillReferenceColumn,
	}
	_ TableProvider[*model.Skill] = &skillsProvider{}
)

type skillsProvider struct {
	table    *unison.Table[*Node[*model.Skill]]
	colMap   map[int]int
	provider model.SkillListProvider
	forPage  bool
}

// NewSkillsProvider creates a new table provider for skills.
func NewSkillsProvider(provider model.SkillListProvider, forPage bool) TableProvider[*model.Skill] {
	p := &skillsProvider{
		provider: provider,
		forPage:  forPage,
	}
	if forPage {
		if _, ok := provider.(*model.Entity); ok {
			p.colMap = entitySkillPageColMap
		} else {
			p.colMap = skillPageColMap
		}
	} else {
		p.colMap = skillListColMap
	}
	return p
}

func (p *skillsProvider) RefKey() string {
	return model.BlockLayoutSkillsKey
}

func (p *skillsProvider) AllTags() []string {
	set := make(map[string]struct{})
	model.Traverse(func(modifier *model.Skill) bool {
		for _, tag := range modifier.Tags {
			set[tag] = struct{}{}
		}
		return false
	}, false, false, p.RootData()...)
	tags := maps.Keys(set)
	txt.SortStringsNaturalAscending(tags)
	return tags
}

func (p *skillsProvider) SetTable(table *unison.Table[*Node[*model.Skill]]) {
	p.table = table
}

func (p *skillsProvider) RootRowCount() int {
	return len(p.provider.SkillList())
}

func (p *skillsProvider) RootRows() []*Node[*model.Skill] {
	data := p.provider.SkillList()
	rows := make([]*Node[*model.Skill], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode[*model.Skill](p.table, nil, p.colMap, one, p.forPage))
	}
	return rows
}

func (p *skillsProvider) SetRootRows(rows []*Node[*model.Skill]) {
	p.provider.SetSkillList(ExtractNodeDataFromList(rows))
}

func (p *skillsProvider) RootData() []*model.Skill {
	return p.provider.SkillList()
}

func (p *skillsProvider) SetRootData(data []*model.Skill) {
	p.provider.SetSkillList(data)
}

func (p *skillsProvider) Entity() *model.Entity {
	return p.provider.Entity()
}

func (p *skillsProvider) DragKey() string {
	return model.SkillID
}

func (p *skillsProvider) DragSVG() *unison.SVG {
	return svg.GCSSkills
}

func (p *skillsProvider) DropShouldMoveData(from, to *unison.Table[*Node[*model.Skill]]) bool {
	return from == to
}

func (p *skillsProvider) ProcessDropData(_, to *unison.Table[*Node[*model.Skill]]) {
	entityProvider := unison.Ancestor[model.EntityProvider](to)
	if !toolbox.IsNil(entityProvider) {
		entity := entityProvider.Entity()
		if entity != nil {
			for _, row := range to.SelectedRows(true) {
				model.Traverse(func(skill *model.Skill) bool {
					if skill.TechLevel != nil && *skill.TechLevel == "" {
						tl := entity.Profile.TechLevel
						skill.TechLevel = &tl
					}
					return false
				}, false, true, row.Data())
			}
		}
	}
}

func (p *skillsProvider) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *skillsProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Skill"), i18n.Text("Skills")
}

func (p *skillsProvider) Headers() []unison.TableColumnHeader[*Node[*model.Skill]] {
	var headers []unison.TableColumnHeader[*Node[*model.Skill]]
	for i := 0; i < len(p.colMap); i++ {
		switch p.colMap[i] {
		case model.SkillDescriptionColumn:
			headers = append(headers, NewEditorListHeader[*model.Skill](i18n.Text("Skill / Technique"), "", p.forPage))
		case model.SkillDifficultyColumn:
			headers = append(headers, NewEditorListHeader[*model.Skill](i18n.Text("Diff"), i18n.Text("Difficulty"), p.forPage))
		case model.SkillTagsColumn:
			headers = append(headers, NewEditorListHeader[*model.Skill](i18n.Text("Tags"), "", p.forPage))
		case model.SkillReferenceColumn:
			headers = append(headers, NewEditorPageRefHeader[*model.Skill](p.forPage))
		case model.SkillLevelColumn:
			headers = append(headers, NewEditorListHeader[*model.Skill](i18n.Text("SL"), i18n.Text("Skill Level"), p.forPage))
		case model.SkillRelativeLevelColumn:
			headers = append(headers, NewEditorListHeader[*model.Skill](i18n.Text("RSL"), i18n.Text("Relative Skill Level"), p.forPage))
		case model.SkillPointsColumn:
			headers = append(headers, NewEditorListHeader[*model.Skill](i18n.Text("Pts"), i18n.Text("Points"), p.forPage))
		default:
			jot.Fatalf(1, "invalid skill column: %d", p.colMap[i])
		}
	}
	return headers
}

func (p *skillsProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*model.Skill]]) {
}

func (p *skillsProvider) HierarchyColumnIndex() int {
	for k, v := range p.colMap {
		if v == model.SkillDescriptionColumn {
			return k
		}
	}
	return 0
}

func (p *skillsProvider) ExcessWidthColumnIndex() int {
	return p.HierarchyColumnIndex()
}

func (p *skillsProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*model.Skill]]) {
	OpenEditor[*model.Skill](table, func(item *model.Skill) { EditSkill(owner, item) })
}

func (p *skillsProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*model.Skill]], variant ItemVariant) {
	var item *model.Skill
	switch variant {
	case NoItemVariant:
		item = model.NewSkill(p.Entity(), nil, false)
	case ContainerItemVariant:
		item = model.NewSkill(p.Entity(), nil, true)
	case AlternateItemVariant:
		item = model.NewTechnique(p.Entity(), nil, "")
	default:
		jot.Fatal(1, "unhandled variant")
	}
	InsertItems[*model.Skill](owner, table, p.provider.SkillList, p.provider.SetSkillList,
		func(_ *unison.Table[*Node[*model.Skill]]) []*Node[*model.Skill] { return p.RootRows() }, item)
	EditSkill(owner, item)
}

func (p *skillsProvider) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.provider.SkillList())
}

func (p *skillsProvider) Deserialize(data []byte) error {
	var rows []*model.Skill
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.provider.SetSkillList(rows)
	return nil
}

func (p *skillsProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	list = append(list,
		ContextMenuItem{i18n.Text("New Skill"), NewSkillItemID},
		ContextMenuItem{i18n.Text("New Skill Container"), NewSkillContainerItemID},
		ContextMenuItem{i18n.Text("New Technique"), NewTechniqueItemID},
	)
	return AppendDefaultContextMenuItems(list)
}
