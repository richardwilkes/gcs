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

type traitListProvider struct {
	traits []*gurps.Trait
}

func (p *traitListProvider) Entity() *gurps.Entity {
	return nil
}

func (p *traitListProvider) TraitList() []*gurps.Trait {
	return p.traits
}

func (p *traitListProvider) SetTraitList(list []*gurps.Trait) {
	p.traits = list
}

// NewTraitTableDockableFromFile loads a list of traits from a file and creates a new unison.Dockable for them.
func NewTraitTableDockableFromFile(filePath string) (unison.Dockable, error) {
	traits, err := gurps.NewTraitsFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewTraitTableDockable(filePath, traits)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewTraitTableDockable creates a new unison.Dockable for trait list files.
func NewTraitTableDockable(filePath string, traits []*gurps.Trait) *TableDockable[*gurps.Trait] {
	provider := &traitListProvider{traits: traits}
	return NewTableDockable(filePath, library.TraitsExt, editors.NewTraitsProvider(provider, false),
		func(path string) error { return gurps.SaveTraits(provider.TraitList(), path) },
		constants.NewTraitItemID, constants.NewTraitContainerItemID)
}
