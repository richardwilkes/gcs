// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/atexit"
	"github.com/richardwilkes/toolbox/collection/dict"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
)

var _ TableProvider[*gurps.Skill] = &skillsProvider{}

type skillsProvider struct {
	table    *unison.Table[*Node[*gurps.Skill]]
	provider gurps.SkillListProvider
	forPage  bool
}

// NewSkillsProvider creates a new table provider for skills.
func NewSkillsProvider(provider gurps.SkillListProvider, forPage bool) TableProvider[*gurps.Skill] {
	return &skillsProvider{
		provider: provider,
		forPage:  forPage,
	}
}

func (p *skillsProvider) RefKey() string {
	return gurps.BlockLayoutSkillsKey
}

func (p *skillsProvider) AllTags() []string {
	set := make(map[string]struct{})
	gurps.Traverse(func(modifier *gurps.Skill) bool {
		for _, tag := range modifier.Tags {
			set[tag] = struct{}{}
		}
		return false
	}, false, false, p.RootData()...)
	tags := dict.Keys(set)
	txt.SortStringsNaturalAscending(tags)
	return tags
}

func (p *skillsProvider) SetTable(table *unison.Table[*Node[*gurps.Skill]]) {
	p.table = table
}

func (p *skillsProvider) RootRowCount() int {
	return len(p.provider.SkillList())
}

func (p *skillsProvider) RootRows() []*Node[*gurps.Skill] {
	data := p.provider.SkillList()
	rows := make([]*Node[*gurps.Skill], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode(p.table, nil, one, p.forPage))
	}
	return rows
}

func (p *skillsProvider) SetRootRows(rows []*Node[*gurps.Skill]) {
	p.provider.SetSkillList(ExtractNodeDataFromList(rows))
}

func (p *skillsProvider) RootData() []*gurps.Skill {
	return p.provider.SkillList()
}

func (p *skillsProvider) SetRootData(data []*gurps.Skill) {
	p.provider.SetSkillList(data)
}

func (p *skillsProvider) DataOwner() gurps.DataOwner {
	return p.provider.DataOwner()
}

func (p *skillsProvider) DragKey() string {
	return gurps.SkillID
}

func (p *skillsProvider) DragSVG() *unison.SVG {
	return svg.GCSSkills
}

func (p *skillsProvider) DropShouldMoveData(from, to *unison.Table[*Node[*gurps.Skill]]) bool {
	return from == to
}

func (p *skillsProvider) ProcessDropData(_, to *unison.Table[*Node[*gurps.Skill]]) {
	if dataOwnerProvider := unison.Ancestor[gurps.DataOwnerProvider](to); !toolbox.IsNil(dataOwnerProvider) {
		if dataOwner := dataOwnerProvider.DataOwner(); !toolbox.IsNil(dataOwner) {
			if entity := dataOwner.OwningEntity(); entity != nil {
				for _, row := range to.SelectedRows(true) {
					gurps.Traverse(func(skill *gurps.Skill) bool {
						if skill.TechLevel != nil && *skill.TechLevel == "" {
							tl := entity.Profile.TechLevel
							skill.TechLevel = &tl
						}
						skill.UpdateLevel()
						return false
					}, false, true, row.Data())
				}
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

func (p *skillsProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.Skill]] {
	ids := p.ColumnIDs()
	headers := make([]unison.TableColumnHeader[*Node[*gurps.Skill]], 0, len(ids))
	for _, id := range ids {
		headers = append(headers, headerFromData[*gurps.Skill](gurps.SkillsHeaderData(id), p.forPage))
	}
	return headers
}

func (p *skillsProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.Skill]]) {
}

func (p *skillsProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 5)
	columnIDs = append(columnIDs, gurps.SkillDescriptionColumn)
	if p.forPage {
		if _, ok := p.provider.(*gurps.Entity); ok {
			columnIDs = append(columnIDs,
				gurps.SkillLevelColumn,
				gurps.SkillRelativeLevelColumn,
			)
		}
		columnIDs = append(columnIDs, gurps.SkillPointsColumn)
	} else {
		columnIDs = append(columnIDs,
			gurps.SkillDifficultyColumn,
			gurps.SkillTagsColumn,
		)
	}
	columnIDs = append(columnIDs, gurps.SkillReferenceColumn)
	if p.forPage {
		if entity := p.DataOwner().OwningEntity(); entity == nil || !entity.SheetSettings.HideSourceMismatch {
			columnIDs = append(columnIDs, gurps.SkillLibSrcColumn)
		}
	}
	return columnIDs
}

func (p *skillsProvider) HierarchyColumnID() int {
	return gurps.SkillDescriptionColumn
}

func (p *skillsProvider) ExcessWidthColumnID() int {
	return gurps.SkillDescriptionColumn
}

func (p *skillsProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.Skill]]) {
	OpenEditor(table, func(item *gurps.Skill) { EditSkill(owner, item) })
}

func (p *skillsProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.Skill]], variant ItemVariant) {
	var item *gurps.Skill
	switch variant {
	case NoItemVariant:
		item = gurps.NewSkill(p.DataOwner(), nil, false)
	case ContainerItemVariant:
		item = gurps.NewSkill(p.DataOwner(), nil, true)
	case AlternateItemVariant:
		item = gurps.NewTechnique(p.DataOwner(), nil, "")
	default:
		errs.Log(errs.New("unhandled variant"), "variant", int(variant))
		atexit.Exit(1)
	}
	InsertItems(owner, table, p.provider.SkillList, p.provider.SetSkillList,
		func(_ *unison.Table[*Node[*gurps.Skill]]) []*Node[*gurps.Skill] { return p.RootRows() }, item)
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

func (p *skillsProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	list = append(list,
		ContextMenuItem{i18n.Text("New Skill"), NewSkillItemID},
		ContextMenuItem{i18n.Text("New Skill Container"), NewSkillContainerItemID},
		ContextMenuItem{i18n.Text("New Technique"), NewTechniqueItemID},
	)
	return AppendDefaultContextMenuItems(list)
}
