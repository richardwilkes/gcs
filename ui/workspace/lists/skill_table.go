package lists

import (
	"os"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ui/workspace/editors"
	"github.com/richardwilkes/unison"
)

type skillListProvider struct {
	skills []*gurps.Skill
}

func (p *skillListProvider) Entity() *gurps.Entity {
	return nil
}

func (p *skillListProvider) SkillList() []*gurps.Skill {
	return p.skills
}

func (p *skillListProvider) SetSkillList(list []*gurps.Skill) {
	p.skills = list
}

// NewSkillTableDockableFromFile loads a list of skills from a file and creates a new unison.Dockable for them.
func NewSkillTableDockableFromFile(filePath string) (unison.Dockable, error) {
	skills, err := gurps.NewSkillsFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewSkillTableDockable(filePath, skills)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewSkillTableDockable creates a new unison.Dockable for skill list files.
func NewSkillTableDockable(filePath string, skills []*gurps.Skill) *TableDockable[*gurps.Skill] {
	provider := &skillListProvider{skills: skills}
	return NewTableDockable(filePath, library.SkillsExt, editors.NewSkillsProvider(provider, false),
		func(path string) error { return gurps.SaveSkills(provider.SkillList(), path) },
		constants.NewSkillItemID, constants.NewSkillContainerItemID, constants.NewTechniqueItemID)
}
