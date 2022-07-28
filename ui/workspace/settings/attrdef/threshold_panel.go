package attrdef

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/attribute"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

type thresholdPanel struct {
	unison.Panel
	pool         *poolPanel
	threshold    *gurps.PoolThreshold
	deleteButton *unison.Button
}

func newThresholdPanel(pool *poolPanel, threshold *gurps.PoolThreshold) *thresholdPanel {
	p := &thresholdPanel{
		pool:      pool,
		threshold: threshold,
	}
	p.Self = p
	p.SetBorder(unison.NewEmptyBorder(unison.Insets{
		Top:    unison.StdVSpacing,
		Left:   unison.StdHSpacing,
		Bottom: unison.StdVSpacing,
		Right:  unison.StdHSpacing,
	}))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		color := unison.ContentColor
		if p.Parent().IndexOfChild(p)%2 == 1 {
			color = unison.BandingColor
		}
		gc.DrawRect(rect, color.Paint(gc, rect, unison.Fill))
	}
	p.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
		HGrab:  true,
	})

	p.AddChild(widget.NewDragHandle(map[string]any{
		attributesDragDataKey: &attributesDragData{
			owner:     pool.dockable.Entity(),
			def:       pool.def,
			threshold: threshold,
		},
	}))
	p.AddChild(p.createButtons())
	p.AddChild(p.createContent())
	return p
}

func (p *thresholdPanel) createButtons() *unison.Panel {
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

	p.deleteButton = unison.NewSVGButton(res.TrashSVG)
	p.deleteButton.ClickCallback = func() { p.pool.deleteThreshold(p) }
	p.deleteButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Remove pool threshold"))
	p.deleteButton.SetEnabled(len(p.pool.def.Thresholds) > 1)
	buttons.AddChild(p.deleteButton)
	return buttons
}

func (p *thresholdPanel) createContent() *unison.Panel {
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
		HGrab:  true,
	})

	text := i18n.Text("State")
	content.AddChild(widget.NewFieldLeadingLabel(text))
	field := widget.NewStringField(p.pool.dockable.targetMgr, p.threshold.KeyPrefix+"state", text,
		func() string { return p.threshold.State },
		func(s string) { p.threshold.State = s })
	field.SetMinimumTextWidthUsing(prototypeMinIDWidth)
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.StartAlignment})
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("A short description of the threshold state"))
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.FillAlignment})
	content.AddChild(field)

	text = i18n.Text("Threshold")
	content.AddChild(widget.NewFieldLeadingLabel(text))
	field = widget.NewStringField(p.pool.dockable.targetMgr, p.threshold.KeyPrefix+"threshold", text,
		func() string { return p.threshold.Expression },
		func(s string) { p.threshold.Expression = s })
	field.SetMinimumTextWidthUsing("round($self*100/50+20)")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("An expression to calculate the threshold value"))
	content.AddChild(field)

	for _, op := range attribute.AllThresholdOp[1:] {
		content.AddChild(unison.NewPanel())
		content.AddChild(p.createOpCheckBox(op))
	}

	text = i18n.Text("Explanation")
	content.AddChild(widget.NewFieldLeadingLabel(text))
	field = widget.NewMultiLineStringField(p.pool.dockable.targetMgr, p.threshold.KeyPrefix+"explanation", text,
		func() string { return p.threshold.Explanation },
		func(s string) { p.threshold.Explanation = s })
	field.SetMinimumTextWidthUsing(prototypeMinNameWidth)
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.StartAlignment})
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("A explanation of the effects of the threshold state"))
	content.AddChild(field)

	return content
}

func (p *thresholdPanel) createOpCheckBox(op attribute.ThresholdOp) *widget.CheckBox {
	c := widget.NewCheckBox(p.pool.dockable.targetMgr, p.threshold.KeyPrefix+op.Key(), op.String(),
		func() unison.CheckState { return unison.CheckStateFromBool(p.threshold.ContainsOp(op)) },
		func(state unison.CheckState) {
			if state == unison.OnCheckState {
				p.threshold.AddOp(op)
			} else {
				p.threshold.RemoveOp(op)
			}
		})
	c.Tooltip = unison.NewTooltipWithText(op.AltString())
	return c
}
