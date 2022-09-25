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

package editors

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

var (
	skillListColMap = map[int]int{
		0: gurps.SkillDescriptionColumn,
		1: gurps.SkillDifficultyColumn,
		2: gurps.SkillTagsColumn,
		3: gurps.SkillReferenceColumn,
	}
	entitySkillPageColMap = map[int]int{
		0: gurps.SkillDescriptionColumn,
		1: gurps.SkillLevelColumn,
		2: gurps.SkillRelativeLevelColumn,
		3: gurps.SkillPointsColumn,
		4: gurps.SkillReferenceColumn,
	}
	skillPageColMap = map[int]int{
		0: gurps.SkillDescriptionColumn,
		1: gurps.SkillPointsColumn,
		2: gurps.SkillReferenceColumn,
	}
	_ ntable.TableProvider[*gurps.Skill] = &skillsProvider{}
)

type skillsProvider struct {
	table    *unison.Table[*ntable.Node[*gurps.Skill]]
	colMap   map[int]int
	provider gurps.SkillListProvider
	forPage  bool
}

// NewSkillsProvider creates a new table provider for skills.
func NewSkillsProvider(provider gurps.SkillListProvider, forPage bool) ntable.TableProvider[*gurps.Skill] {
	p := &skillsProvider{
		provider: provider,
		forPage:  forPage,
	}
	if forPage {
		if _, ok := provider.(*gurps.Entity); ok {
			p.colMap = entitySkillPageColMap
		} else {
			p.colMap = skillPageColMap
		}
	} else {
		p.colMap = skillListColMap
	}
	return p
}

func (p *skillsProvider) SetTable(table *unison.Table[*ntable.Node[*gurps.Skill]]) {
	p.table = table
}

func (p *skillsProvider) RootRowCount() int {
	return len(p.provider.SkillList())
}

func (p *skillsProvider) RootRows() []*ntable.Node[*gurps.Skill] {
	data := p.provider.SkillList()
	rows := make([]*ntable.Node[*gurps.Skill], 0, len(data))
	for _, one := range data {
		rows = append(rows, ntable.NewNode[*gurps.Skill](p.table, nil, p.colMap, one, p.forPage))
	}
	return rows
}

func (p *skillsProvider) SetRootRows(rows []*ntable.Node[*gurps.Skill]) {
	p.provider.SetSkillList(ntable.ExtractNodeDataFromList(rows))
}

func (p *skillsProvider) RootData() []*gurps.Skill {
	return p.provider.SkillList()
}

func (p *skillsProvider) SetRootData(data []*gurps.Skill) {
	p.provider.SetSkillList(data)
}

func (p *skillsProvider) Entity() *gurps.Entity {
	return p.provider.Entity()
}

func (p *skillsProvider) DragKey() string {
	return gid.Skill
}

func (p *skillsProvider) DragSVG() *unison.SVG {
	return res.GCSSkillsSVG
}

func (p *skillsProvider) DropShouldMoveData(from, to *unison.Table[*ntable.Node[*gurps.Skill]]) bool {
	return from == to
}

func (p *skillsProvider) ProcessDropData(_, to *unison.Table[*ntable.Node[*gurps.Skill]]) {
	entityProvider := unison.Ancestor[gurps.EntityProvider](to)
	if !toolbox.IsNil(entityProvider) {
		entity := entityProvider.Entity()
		if entity != nil {
			for _, row := range to.SelectedRows(true) {
				gurps.Traverse(func(skill *gurps.Skill) bool {
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

func (p *skillsProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Skill"), i18n.Text("Skills")
}

func (p *skillsProvider) Headers() []unison.TableColumnHeader[*ntable.Node[*gurps.Skill]] {
	var headers []unison.TableColumnHeader[*ntable.Node[*gurps.Skill]]
	for i := 0; i < len(p.colMap); i++ {
		switch p.colMap[i] {
		case gurps.SkillDescriptionColumn:
			headers = append(headers, NewHeader[*gurps.Skill](i18n.Text("Skill / Technique"), "", p.forPage))
		case gurps.SkillDifficultyColumn:
			headers = append(headers, NewHeader[*gurps.Skill](i18n.Text("Diff"), i18n.Text("Difficulty"), p.forPage))
		case gurps.SkillTagsColumn:
			headers = append(headers, NewHeader[*gurps.Skill](i18n.Text("Tags"), "", p.forPage))
		case gurps.SkillReferenceColumn:
			headers = append(headers, NewPageRefHeader[*gurps.Skill](p.forPage))
		case gurps.SkillLevelColumn:
			headers = append(headers, NewHeader[*gurps.Skill](i18n.Text("SL"), i18n.Text("Skill Level"), p.forPage))
		case gurps.SkillRelativeLevelColumn:
			headers = append(headers, NewHeader[*gurps.Skill](i18n.Text("RSL"), i18n.Text("Relative Skill Level"), p.forPage))
		case gurps.SkillPointsColumn:
			headers = append(headers, NewHeader[*gurps.Skill](i18n.Text("Pts"), i18n.Text("Points"), p.forPage))
		default:
			jot.Fatalf(1, "invalid skill column: %d", p.colMap[i])
		}
	}
	return headers
}

func (p *skillsProvider) SyncHeader(_ []unison.TableColumnHeader[*ntable.Node[*gurps.Skill]]) {
}

func (p *skillsProvider) HierarchyColumnIndex() int {
	for k, v := range p.colMap {
		if v == gurps.SkillDescriptionColumn {
			return k
		}
	}
	return 0
}

func (p *skillsProvider) ExcessWidthColumnIndex() int {
	return p.HierarchyColumnIndex()
}

func (p *skillsProvider) OpenEditor(owner widget.Rebuildable, table *unison.Table[*ntable.Node[*gurps.Skill]]) {
	ntable.OpenEditor[*gurps.Skill](table, func(item *gurps.Skill) { EditSkill(owner, item) })
}

func (p *skillsProvider) CreateItem(owner widget.Rebuildable, table *unison.Table[*ntable.Node[*gurps.Skill]], variant ntable.ItemVariant) {
	var item *gurps.Skill
	switch variant {
	case ntable.NoItemVariant:
		item = gurps.NewSkill(p.Entity(), nil, false)
	case ntable.ContainerItemVariant:
		item = gurps.NewSkill(p.Entity(), nil, true)
	case ntable.AlternateItemVariant:
		item = gurps.NewTechnique(p.Entity(), nil, "")
	default:
		jot.Fatal(1, "unhandled variant")
	}
	ntable.InsertItems[*gurps.Skill](owner, table, p.provider.SkillList, p.provider.SetSkillList,
		func(_ *unison.Table[*ntable.Node[*gurps.Skill]]) []*ntable.Node[*gurps.Skill] { return p.RootRows() }, item)
	EditSkill(owner, item)
}

func (p *skillsProvider) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.provider.SkillList())
}

func (p *skillsProvider) Deserialize(data []byte) error {
	var rows []*gurps.Skill
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.provider.SetSkillList(rows)
	return nil
}

func (p *skillsProvider) ContextMenuItems() []ntable.ContextMenuItem {
	var list []ntable.ContextMenuItem
	list = append(list, ntable.SkillExtraContextMenuItems...)
	return append(list, ntable.DefaultContextMenuItems...)
}
