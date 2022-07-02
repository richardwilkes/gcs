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

type equipmentModifierListProvider struct {
	modifiers []*gurps.EquipmentModifier
}

func (p *equipmentModifierListProvider) Entity() *gurps.Entity {
	return nil
}

func (p *equipmentModifierListProvider) EquipmentModifierList() []*gurps.EquipmentModifier {
	return p.modifiers
}

func (p *equipmentModifierListProvider) SetEquipmentModifierList(list []*gurps.EquipmentModifier) {
	p.modifiers = list
}

// NewEquipmentModifierTableDockableFromFile loads a list of equipment modifiers from a file and creates a new
// unison.Dockable for them.
func NewEquipmentModifierTableDockableFromFile(filePath string) (unison.Dockable, error) {
	modifiers, err := gurps.NewEquipmentModifiersFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewEquipmentModifierTableDockable(filePath, modifiers)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewEquipmentModifierTableDockable creates a new unison.Dockable for equipment modifier list files.
func NewEquipmentModifierTableDockable(filePath string, modifiers []*gurps.EquipmentModifier) *TableDockable[*gurps.EquipmentModifier] {
	provider := &equipmentModifierListProvider{modifiers: modifiers}
	return NewTableDockable(filePath, library.EquipmentModifiersExt,
		editors.NewEquipmentModifiersProvider(provider, false),
		func(path string) error { return gurps.SaveEquipmentModifiers(provider.EquipmentModifierList(), path) },
		constants.NewEquipmentModifierItemID, constants.NewEquipmentContainerModifierItemID)
}
