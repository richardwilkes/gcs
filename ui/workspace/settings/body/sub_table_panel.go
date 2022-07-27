package body

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/res"
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
		HSpan:  2,
		HAlign: unison.FillAlignment,
	})

	p.AddChild(p.createButtons())
	p.AddChild(p.createContent())

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

func (p *subTablePanel) createContent() *unison.Panel {
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
	})
	content.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))

	for _, loc := range p.body.Locations {
		content.AddChild(newHitLocationPanel(p.dockable, loc))
	}

	return content
}
