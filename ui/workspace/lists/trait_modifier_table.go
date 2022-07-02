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

type traitModifierListProvider struct {
	modifiers []*gurps.TraitModifier
}

func (p *traitModifierListProvider) Entity() *gurps.Entity {
	return nil
}

func (p *traitModifierListProvider) TraitModifierList() []*gurps.TraitModifier {
	return p.modifiers
}

func (p *traitModifierListProvider) SetTraitModifierList(list []*gurps.TraitModifier) {
	p.modifiers = list
}

// NewTraitModifierTableDockableFromFile loads a list of trait modifiers from a file and creates a new
// unison.Dockable for them.
func NewTraitModifierTableDockableFromFile(filePath string) (unison.Dockable, error) {
	modifiers, err := gurps.NewTraitModifiersFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewTraitModifierTableDockable(filePath, modifiers)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewTraitModifierTableDockable creates a new unison.Dockable for trait modifier list files.
func NewTraitModifierTableDockable(filePath string, modifiers []*gurps.TraitModifier) *TableDockable[*gurps.TraitModifier] {
	provider := &traitModifierListProvider{modifiers: modifiers}
	return NewTableDockable(filePath, library.TraitModifiersExt,
		editors.NewTraitModifiersProvider(provider, false),
		func(path string) error { return gurps.SaveTraitModifiers(provider.TraitModifierList(), path) },
		constants.NewTraitModifierItemID, constants.NewTraitContainerModifierItemID)
}
