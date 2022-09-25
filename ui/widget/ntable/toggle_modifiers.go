package ntable

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// ProcessModifiersForSelection processes the selected rows for modifiers that can be toggled on or off.
func ProcessModifiersForSelection[T gurps.NodeTypes](table *unison.Table[*Node[T]]) {
	rows := table.SelectedRows(true)
	data := make([]T, 0, len(rows))
	for _, row := range rows {
		data = append(data, row.Data())
	}
	ProcessModifiers(table, data)
}

// ProcessModifiers processes the rows for modifiers that can be toggled on or off.
func ProcessModifiers[T gurps.NodeTypes](owner unison.Paneler, rows []T) {
	for _, row := range rows {
		gurps.Traverse(func(row T) bool {
			switch t := (any(row)).(type) {
			case *gurps.Trait:
				if processModifiers(t.Modifiers) {
					unison.Ancestor[widget.Rebuildable](owner).Rebuild(true)
				}
			case *gurps.Equipment:
				if processModifiers(t.Modifiers) {
					unison.Ancestor[widget.Rebuildable](owner).Rebuild(true)
				}
			}
			return false
		}, false, false, row)
	}
}

func processModifiers[T *gurps.TraitModifier | *gurps.EquipmentModifier](modifiers []T) bool {
	if len(modifiers) == 0 {
		return false
	}
	list := unison.NewPanel()
	list.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(unison.StdHSpacing)))
	list.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	tracker := make(map[*unison.CheckBox]gurps.GeneralModifier)
	gurps.Traverse[T](func(m T) bool {
		var p *unison.Panel
		if mod, ok := any(m).(gurps.GeneralModifier); ok {
			if mod.Container() {
				label := unison.NewLabel()
				label.Text = mod.FullDescription()
				p = label.AsPanel()
			} else {
				cb := unison.NewCheckBox()
				cb.Text = mod.FullDescription()
				cb.State = unison.CheckStateFromBool(mod.Enabled())
				tracker[cb] = mod
				p = cb.AsPanel()
			}
			p.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: float32(mod.Depth()) * 16}))
			list.AddChild(p)
		}
		return false
	}, false, false, modifiers...)
	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
	scroll.SetContent(list, unison.FillBehavior, unison.FillBehavior)
	scroll.BackgroundInk = unison.ContentColor
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   unison.FillAlignment,
		VAlign:   unison.FillAlignment,
	})
	label := unison.NewLabel()
	label.Text = i18n.Text("Select Modifiers:")
	panel.AddChild(label)
	panel.AddChild(scroll)
	if unison.QuestionDialogWithPanel(panel) == unison.ModalResponseOK {
		changed := false
		for _, row := range list.Children() {
			if cb, ok := row.Self.(*unison.CheckBox); ok {
				var mod gurps.GeneralModifier
				if mod, ok = tracker[cb]; ok {
					if on := cb.State == unison.OnCheckState; mod.Enabled() != on {
						mod.SetEnabled(on)
						changed = true
					}
				}
			}
		}
		return changed
	}
	return false
}
