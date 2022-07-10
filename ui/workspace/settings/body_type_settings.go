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

var _ widget.GroupedCloser = &bodyTypesDockable{}

type bodyTypesDockable struct {
	Dockable
	owner widget.EntityPanel
}

// ShowBodyTypeSettings the Body Type Settings. Pass in nil to edit the defaults or a sheet to edit the sheet's.
func ShowBodyTypeSettings(owner widget.EntityPanel) {
	ws, dc, found := workspace.Activate(func(d unison.Dockable) bool {
		if s, ok := d.(*bodyTypesDockable); ok && owner == s.owner {
			return true
		}
		return false
	})
	if !found && ws != nil {
		d := &bodyTypesDockable{owner: owner}
		d.Self = d
		if owner != nil {
			d.TabTitle = i18n.Text("Body Type: " + owner.Entity().Profile.Name)
		} else {
			d.TabTitle = i18n.Text("Default Body Type")
		}
		d.Extensions = []string{".body", ".ghl"}
		d.Loader = d.load
		d.Saver = d.save
		d.Resetter = d.reset
		d.Setup(ws, dc, nil, nil, d.initContent)
	}
}

func (d *bodyTypesDockable) CloseWithGroup(other unison.Paneler) bool {
	return d.owner != nil && d.owner == other
}

func (d *bodyTypesDockable) bodyType() *gurps.BodyType {
	if d.owner != nil {
		return d.owner.Entity().SheetSettings.BodyType
	}
	return settings.Global().Sheet.BodyType
}

func (d *bodyTypesDockable) initContent(content *unison.Panel) {
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.DefaultLabelTheme.Font.LineHeight(),
	})
	// TODO: build content here
}

func (d *bodyTypesDockable) reset() {
	if d.owner != nil {
		entity := d.owner.Entity()
		entity.SheetSettings.BodyType = settings.Global().Sheet.BodyType.Clone(entity, nil)
	} else {
		settings.Global().Sheet.BodyType = gurps.FactoryBodyType()
	}
	d.sync()
}

func (d *bodyTypesDockable) sync() {
	// TODO: Sync content here
	d.MarkForRedraw()
	d.syncSheet(true) // TODO: Remove when actual sync logic is installed, above
}

func (d *bodyTypesDockable) syncSheet(full bool) {
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

func (d *bodyTypesDockable) load(fileSystem fs.FS, filePath string) error {
	bodyType, err := gurps.NewBodyTypeFromFile(fileSystem, filePath)
	if err != nil {
		return err
	}
	if d.owner != nil {
		d.owner.Entity().SheetSettings.BodyType = bodyType
	} else {
		settings.Global().Sheet.BodyType = bodyType
	}
	d.sync()
	return nil
}

func (d *bodyTypesDockable) save(filePath string) error {
	return d.bodyType().Save(filePath)
}
