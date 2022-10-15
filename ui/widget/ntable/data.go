package ntable

import (
	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/unison"
)

// PreservedTableData holds the data and selection state of a table in a serialized form.
type PreservedTableData[T gurps.NodeTypes] struct {
	data   []byte
	selMap map[uuid.UUID]bool
}

// Collect the data and selection state from a table.
func (d *PreservedTableData[T]) Collect(table *unison.Table[*Node[T]]) error {
	provider, ok := table.ClientData()[TableProviderClientKey].(TableProvider[T])
	if !ok {
		return errs.New("unable to locate provider")
	}
	data, err := provider.Serialize()
	if err != nil {
		return err
	}
	d.data = data
	d.selMap = table.CopySelectionMap()
	return nil
}

// Apply the data and selection state to a table.
func (d *PreservedTableData[T]) Apply(table *unison.Table[*Node[T]]) error {
	provider, ok := table.ClientData()[TableProviderClientKey].(TableProvider[T])
	if !ok {
		return errs.New("unable to locate provider")
	}
	if err := provider.Deserialize(d.data); err != nil {
		return err
	}
	table.SyncToModel()
	widget.MarkModified(table)
	table.SetSelectionMap(d.selMap)
	return nil
}
