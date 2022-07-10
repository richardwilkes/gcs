package settings

import (
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

var _ widget.GroupedCloser = &attributesDockable{}

type attributesDockable struct {
	Dockable
	owner widget.EntityPanel
}

// ShowAttributeSettings the Attribute Settings. Pass in nil to edit the defaults or a sheet to edit the sheet's.
func ShowAttributeSettings(owner widget.EntityPanel) {
	ws, dc, found := workspace.Activate(func(d unison.Dockable) bool {
		if s, ok := d.(*attributesDockable); ok && owner == s.owner {
			return true
		}
		return false
	})
	if !found && ws != nil {
		d := &attributesDockable{owner: owner}
		d.Self = d
		if owner != nil {
			d.TabTitle = i18n.Text("Attributes: " + owner.Entity().Profile.Name)
		} else {
			d.TabTitle = i18n.Text("Default Attributes")
		}
		d.Extensions = []string{".attr", ".attributes", ".gas"}
		d.Loader = d.load
		d.Saver = d.save
		d.Resetter = d.reset
		d.Setup(ws, dc, nil, nil, d.initContent)
	}
}

func (d *attributesDockable) CloseWithGroup(other unison.Paneler) bool {
	return d.owner != nil && d.owner == other
}

func (d *attributesDockable) defs() *gurps.AttributeDefs {
	if d.owner != nil {
		return d.owner.Entity().SheetSettings.Attributes
	}
	return settings.Global().Sheet.Attributes
}

func (d *attributesDockable) initContent(content *unison.Panel) {
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.DefaultLabelTheme.Font.LineHeight(),
	})
	// TODO: build content here
}

func (d *attributesDockable) reset() {
	if d.owner != nil {
		entity := d.owner.Entity()
		entity.SheetSettings.Attributes = settings.Global().Sheet.Attributes.Clone()
	} else {
		settings.Global().Sheet.Attributes = gurps.FactoryAttributeDefs()
	}
	d.sync()
}

func (d *attributesDockable) sync() {
	// TODO: Sync content here
	d.MarkForRedraw()
	d.syncSheet(true) // TODO: Remove when actual sync logic is installed, above
}

func (d *attributesDockable) syncSheet(full bool) {
	if d.owner == nil {
		return
	}
	entity := d.owner.Entity()
	for _, wnd := range unison.Windows() {
		if ws := workspace.FromWindow(wnd); ws != nil {
			ws.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
				for _, one := range dc.Dockables() {
					if s, ok := one.(gurps.SheetSettingsResponder); ok {
						s.SheetSettingsUpdated(entity, full)
					}
				}
				return false
			})
		}
	}
}

func (d *attributesDockable) load(fileSystem fs.FS, filePath string) error {
	defs, err := gurps.NewAttributeDefsFromFile(fileSystem, filePath)
	if err != nil {
		return err
	}
	if d.owner != nil {
		entity := d.owner.Entity()
		entity.SheetSettings.Attributes = defs
		for attrID, def := range entity.SheetSettings.Attributes.Set {
			if attr, exists := entity.Attributes.Set[attrID]; exists {
				attr.Order = def.Order
			} else {
				entity.Attributes.Set[attrID] = gurps.NewAttribute(entity, attrID, def.Order)
			}
		}
		for attrID := range entity.Attributes.Set {
			if _, exists := defs.Set[attrID]; !exists {
				delete(entity.Attributes.Set, attrID)
			}
		}
	} else {
		settings.Global().Sheet.Attributes = defs
	}
	d.sync()
	return nil
}

func (d *attributesDockable) save(filePath string) error {
	return d.defs().Save(filePath)
}
