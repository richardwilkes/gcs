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
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
	})

	p.AddChild(p.createButtons())
	p.AddChild(p.createContent())

	return p
}

func (p *bodyPanel) createButtons() *unison.Panel {
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

	addButton := unison.NewSVGButton(res.CircledAddSVG)
	addButton.ClickCallback = p.addHitLocation
	addButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Add hit location"))
	buttons.AddChild(addButton)
	return buttons
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

func (p *bodyPanel) createContent() *unison.Panel {
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
	})

	text := i18n.Text("Name")
	content.AddChild(widget.NewFieldLeadingLabel(text))
	field := widget.NewStringField(p.dockable.targetMgr, p.dockable.body.KeyPrefix+"name", text,
		func() string { return p.dockable.body.Name },
		func(s string) { p.dockable.body.Name = s })
	field.SetMinimumTextWidthUsing(prototypeMinNameWidth)
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.StartAlignment})
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("The name of this body type"))
	content.AddChild(field)

	text = i18n.Text("Roll")
	content.AddChild(widget.NewFieldLeadingLabel(text))
	field = widget.NewStringField(p.dockable.targetMgr, p.dockable.body.KeyPrefix+"roll", text,
		func() string { return p.dockable.body.Roll.String() },
		func(s string) { p.dockable.body.Roll = dice.New(s) })
	field.SetMinimumTextWidthUsing("100d1000")
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.StartAlignment})
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("The dice to roll on the table"))
	content.AddChild(field)

	wrapper := unison.NewPanel()
	wrapper.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: unison.FillAlignment,
	})
	wrapper.SetLayout(&unison.FlexLayout{Columns: 1})
	content.AddChild(wrapper)

	for _, loc := range p.dockable.body.Locations {
		wrapper.AddChild(newHitLocationPanel(p.dockable, loc))
	}
	return content
}
