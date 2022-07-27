package body

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

type subTablePanel struct {
	unison.Panel
	dockable     *bodyDockable
	body         *gurps.Body
	addButton    *unison.Button
	deleteButton *unison.Button
}

func newSubTablePanel(d *bodyDockable, body *gurps.Body) *subTablePanel {
	p := &subTablePanel{
		dockable: d,
		body:     body,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})

	p.AddChild(p.createButtons())

	contentWrapper := unison.NewPanel()
	contentWrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	contentWrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	p.AddChild(contentWrapper)

	text := i18n.Text("Sub-Table Roll")
	contentWrapper.AddChild(widget.NewFieldLeadingLabel(text))
	field := widget.NewStringField(p.dockable.targetMgr, p.body.KeyPrefix+"sub-table_roll", text,
		func() string { return p.body.Roll.String() },
		func(s string) { p.body.Roll = dice.New(s) })
	field.SetMinimumTextWidthUsing("2d100")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("The dice to roll on the table"))
	contentWrapper.AddChild(field)

	wrapper := unison.NewPanel()
	wrapper.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	wrapper.SetLayout(&unison.FlexLayout{Columns: 1})
	contentWrapper.AddChild(wrapper)

	for _, loc := range p.body.Locations {
		wrapper.AddChild(newHitLocationPanel(p.dockable, loc))
	}

	return p
}

func (p *subTablePanel) createButtons() *unison.Panel {
	buttons := unison.NewPanel()
	buttons.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	buttons.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.MiddleAlignment,
		VAlign: unison.StartAlignment,
	})

	p.addButton = unison.NewSVGButton(res.CircledAddSVG)
	p.addButton.ClickCallback = p.addHitLocation
	p.addButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Add hit location"))
	buttons.AddChild(p.addButton)

	p.deleteButton = unison.NewSVGButton(res.TrashSVG)
	p.deleteButton.ClickCallback = p.removeSubTable
	p.deleteButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Remove sub-table"))
	buttons.AddChild(p.deleteButton)
	return buttons
}

func (p *subTablePanel) addHitLocation() {
	undo := p.dockable.prepareUndo(i18n.Text("Add Hit Location"))
	location := gurps.NewHitLocation(p.dockable.Entity(), p.dockable.targetMgr.NextPrefix())
	p.body.AddLocation(location)
	p.dockable.finishAndPostUndo(undo)
	p.dockable.sync()
	if focus := p.dockable.targetMgr.Find(location.KeyPrefix + "id"); focus != nil {
		focus.RequestFocus()
	}
}

func (p *subTablePanel) removeSubTable() {
	undo := p.dockable.prepareUndo(i18n.Text("Remove Sub-Table"))
	p.body.OwningLocation().SubTable = nil
	p.dockable.finishAndPostUndo(undo)
	p.dockable.sync()
}
