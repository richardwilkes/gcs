package body

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

const hitLocationDragDataKey = "drag.body"

type bodyPanel struct {
	unison.Panel
	dockable *bodyDockable
}

type hitLocationDragData struct {
	owner *gurps.Body
	loc   *gurps.HitLocation
}

func newBodyPanel(d *bodyDockable) *bodyPanel {
	p := &bodyPanel{
		dockable: d,
	}
	p.Self = p
	p.SetBorder(unison.NewEmptyBorder(unison.Insets{
		Top:    unison.StdVSpacing,
		Left:   unison.StdHSpacing,
		Bottom: unison.StdVSpacing,
		Right:  unison.StdHSpacing * 2,
	}))
	p.SetLayout(&unison.FlexLayout{
		Columns:  5,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
		HGrab:  true,
	})

	addButton := unison.NewSVGButton(res.CircledAddSVG)
	addButton.ClickCallback = p.addHitLocation
	addButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Add hit location"))
	p.AddChild(addButton)

	text := i18n.Text("Name")
	p.AddChild(widget.NewFieldLeadingLabel(text))
	field := widget.NewStringField(p.dockable.targetMgr, p.dockable.body.KeyPrefix+"name", text,
		func() string { return p.dockable.body.Name },
		func(s string) { p.dockable.body.Name = s })
	field.SetMinimumTextWidthUsing("humanoid")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("The name of this body type"))
	p.AddChild(field)

	text = i18n.Text("Roll")
	p.AddChild(widget.NewFieldLeadingLabel(text))
	field = widget.NewStringField(p.dockable.targetMgr, p.dockable.body.KeyPrefix+"roll", text,
		func() string { return p.dockable.body.Roll.String() },
		func(s string) { p.dockable.body.Roll = dice.New(s) })
	field.SetMinimumTextWidthUsing("2d100")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("The dice to roll on the table"))
	p.AddChild(field)

	wrapper := unison.NewPanel()
	wrapper.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  4,
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	wrapper.SetLayout(&unison.FlexLayout{Columns: 1})
	p.AddChild(unison.NewPanel())
	p.AddChild(wrapper)

	for _, loc := range p.dockable.body.Locations {
		wrapper.AddChild(newHitLocationPanel(p.dockable, loc))
	}

	return p
}

func (p *bodyPanel) addHitLocation() {
	undo := p.dockable.prepareUndo(i18n.Text("Add Hit Location"))
	location := gurps.NewHitLocation(p.dockable.Entity(), p.dockable.targetMgr.NextPrefix())
	p.dockable.body.AddLocation(location)
	p.dockable.finishAndPostUndo(undo)
	p.dockable.sync()
	if focus := p.dockable.targetMgr.Find(location.KeyPrefix + "id"); focus != nil {
		focus.RequestFocus()
	}
}
