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

type spellListProvider struct {
	spells []*gurps.Spell
}

func (p *spellListProvider) Entity() *gurps.Entity {
	return nil
}

func (p *spellListProvider) SpellList() []*gurps.Spell {
	return p.spells
}

func (p *spellListProvider) SetSpellList(list []*gurps.Spell) {
	p.spells = list
}

// NewSpellTableDockableFromFile loads a list of spells from a file and creates a new unison.Dockable for them.
func NewSpellTableDockableFromFile(filePath string) (unison.Dockable, error) {
	spells, err := gurps.NewSpellsFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewSpellTableDockable(filePath, spells)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewSpellTableDockable creates a new unison.Dockable for spell list files.
func NewSpellTableDockable(filePath string, spells []*gurps.Spell) *TableDockable[*gurps.Spell] {
	provider := &spellListProvider{spells: spells}
	return NewTableDockable(filePath, library.SpellsExt, editors.NewSpellsProvider(provider, false),
		func(path string) error { return gurps.SaveSpells(provider.SpellList(), path) },
		constants.NewSpellItemID, constants.NewSpellContainerItemID, constants.NewRitualMagicSpellItemID)
}
