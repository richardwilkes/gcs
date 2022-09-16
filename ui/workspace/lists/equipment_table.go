package lists

import (
	"os"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ui/workspace/editors"
	"github.com/richardwilkes/gcs/v5/ui/workspace/sheet"
	"github.com/richardwilkes/unison"
)

type equipmentListProvider struct {
	carried []*gurps.Equipment
	other   []*gurps.Equipment
}

func (p *equipmentListProvider) Entity() *gurps.Entity {
	return nil
}

func (p *equipmentListProvider) CarriedEquipmentList() []*gurps.Equipment {
	return p.carried
}

func (p *equipmentListProvider) SetCarriedEquipmentList(list []*gurps.Equipment) {
	p.carried = list
}

func (p *equipmentListProvider) OtherEquipmentList() []*gurps.Equipment {
	return p.other
}

func (p *equipmentListProvider) SetOtherEquipmentList(list []*gurps.Equipment) {
	p.other = list
}

// NewEquipmentTableDockableFromFile loads a list of equipment from a file and creates a new unison.Dockable for them.
func NewEquipmentTableDockableFromFile(filePath string) (unison.Dockable, error) {
	equipment, err := gurps.NewEquipmentFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewEquipmentTableDockable(filePath, equipment)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewEquipmentTableDockable creates a new unison.Dockable for equipment list files.
func NewEquipmentTableDockable(filePath string, equipment []*gurps.Equipment) *TableDockable[*gurps.Equipment] {
	provider := &equipmentListProvider{other: equipment}
	d := NewTableDockable(filePath, library.EquipmentExt, editors.NewEquipmentProvider(provider, false, false),
		func(path string) error { return gurps.SaveEquipment(provider.OtherEquipmentList(), path) },
		constants.NewCarriedEquipmentItemID, constants.NewCarriedEquipmentContainerItemID)
	d.InstallCmdHandlers(constants.ConvertToContainerItemID,
		func(_ any) bool { return sheet.CanConvertToContainer(d.table) },
		func(_ any) { sheet.ConvertToContainer(d, d.table) })
	return d
}
