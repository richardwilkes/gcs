package attrdef

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/attribute"
	"github.com/richardwilkes/gcs/v5/model/id"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

type attrDefPanel struct {
	unison.Panel
	dockable           *attributesDockable
	def                *gurps.AttributeDef
	deleteButton       *unison.Button
	addThresholdButton *unison.Button
	poolPanel          *poolPanel
}

func newAttrDefPanel(dockable *attributesDockable, def *gurps.AttributeDef) *attrDefPanel {
	p := &attrDefPanel{
		dockable: dockable,
		def:      def,
	}
	p.Self = p
	p.SetBorder(unison.NewEmptyBorder(unison.Insets{
		Top:    unison.StdVSpacing,
		Left:   unison.StdHSpacing,
		Bottom: unison.StdVSpacing,
		Right:  unison.StdHSpacing * 2,
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
		HGrab:  true,
	})

	p.AddChild(widget.NewDragHandle(map[string]any{attributesDragDataKey: &attributesDragData{
		owner: dockable.Entity(),
		def:   def,
	}}))
	p.AddChild(p.createButtons())
	p.AddChild(p.createContent())
	return p
}

func (p *attrDefPanel) createButtons() *unison.Panel {
	buttons := unison.NewPanel()
	buttons.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	buttons.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.MiddleAlignment})

	p.deleteButton = unison.NewSVGButton(res.TrashSVG)
	p.deleteButton.ClickCallback = p.deleteAttrDef
	p.deleteButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Remove attribute"))
	buttons.AddChild(p.deleteButton)

	p.addThresholdButton = unison.NewSVGButton(res.CircledAddSVG)
	p.addThresholdButton.ClickCallback = func() { p.poolPanel.addThreshold() }
	p.addThresholdButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Add pool threshold"))
	p.addThresholdButton.SetEnabled(p.def.Type == attribute.Pool)
	buttons.AddChild(p.addThresholdButton)
	return buttons
}

func (p *attrDefPanel) deleteAttrDef() {
	attrPanel := p.Parent()
	p.RemoveFromParent()
	children := attrPanel.Children()
	if len(children) == 1 {
		children[0].Self.(*attrDefPanel).deleteButton.SetEnabled(false)
	}
	undo := &unison.UndoEdit[*gurps.AttributeDefs]{
		ID:         unison.NextUndoID(),
		EditName:   i18n.Text("Delete Attribute"),
		UndoFunc:   func(e *unison.UndoEdit[*gurps.AttributeDefs]) { p.dockable.applyAttrDefs(e.BeforeData) },
		RedoFunc:   func(e *unison.UndoEdit[*gurps.AttributeDefs]) { p.dockable.applyAttrDefs(e.AfterData) },
		AbsorbFunc: func(e *unison.UndoEdit[*gurps.AttributeDefs], other unison.Undoable) bool { return false },
	}
	undo.BeforeData = p.dockable.defs.Clone()
	delete(p.dockable.defs.Set, p.def.DefID)
	undo.AfterData = p.dockable.defs.Clone()
	p.dockable.UndoManager().Add(undo)
	p.dockable.MarkModified()
}

func (p *attrDefPanel) createContent() *unison.Panel {
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})

	text := i18n.Text("ID")
	content.AddChild(widget.NewFieldLeadingLabel(text))
	field := widget.NewStringField(p.dockable.targetMgr, p.def.KeyPrefix+"id", text,
		func() string { return p.def.DefID },
		func(s string) {
			if p.validateAttrID(s) {
				delete(p.dockable.defs.Set, p.def.DefID)
				p.def.DefID = strings.TrimSpace(strings.ToLower(s))
				p.dockable.defs.Set[p.def.DefID] = p.def
			}
		})
	field.ValidateCallback = func(field *widget.StringField, def *gurps.AttributeDef) func() bool {
		return func() bool { return p.validateAttrID(field.Text()) }
	}(field, p.def)
	field.SetMinimumTextWidthUsing(prototypeMinIDWidth)
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("A unique ID for the attribute"))
	content.AddChild(field)

	if p.def.IsSeparator() {
		text = i18n.Text("Name")
		content.AddChild(widget.NewFieldLeadingLabel(text))
		field = widget.NewStringField(p.dockable.targetMgr, p.def.KeyPrefix+"name", text,
			func() string { return p.def.Name },
			func(s string) { p.def.Name = s })
		field.SetMinimumTextWidthUsing(prototypeMinIDWidth)
		field.Tooltip = unison.NewTooltipWithText(i18n.Text("A title to use with the separator"))
		content.AddChild(field)
	} else {
		text = i18n.Text("Short Name")
		content.AddChild(widget.NewFieldLeadingLabel(text))
		field = widget.NewStringField(p.dockable.targetMgr, p.def.KeyPrefix+"name", text,
			func() string { return p.def.Name },
			func(s string) { p.def.Name = s })
		field.SetMinimumTextWidthUsing(prototypeMinIDWidth)
		field.Tooltip = unison.NewTooltipWithText(i18n.Text("The name of this attribute, often an abbreviation"))
		content.AddChild(field)

		text = i18n.Text("Full Name")
		content.AddChild(widget.NewFieldLeadingLabel(text))
		field = widget.NewStringField(p.dockable.targetMgr, p.def.KeyPrefix+"fullname", text,
			func() string { return p.def.FullName },
			func(s string) { p.def.FullName = s })
		field.SetMinimumTextWidthUsing(prototypeMinNameWidth)
		field.Tooltip = unison.NewTooltipWithText(i18n.Text("The full name of this attribute (may be omitted, in which case the Short Name will be used instead)"))
		content.AddChild(field)

		text = i18n.Text("Base Value")
		content.AddChild(widget.NewFieldLeadingLabel(text))
		field = widget.NewStringField(p.dockable.targetMgr, p.def.KeyPrefix+"base", text,
			func() string { return p.def.AttributeBase },
			func(s string) { p.def.AttributeBase = s })
		field.SetMinimumTextWidthUsing("floor($basic_speed)")
		field.Tooltip = unison.NewTooltipWithText(i18n.Text("The base value, which may be a number or a formula"))
		content.AddChild(field)

		text = i18n.Text("Cost per Point")
		content.AddChild(widget.NewFieldLeadingLabel(text))
		numField := widget.NewIntegerField(p.dockable.targetMgr, p.def.KeyPrefix+"cost", text,
			func() int { return fxp.As[int](p.def.CostPerPoint) },
			func(v int) { p.def.CostPerPoint = fxp.From(v) },
			0, 9999, false, false)
		numField.Tooltip = unison.NewTooltipWithText(i18n.Text("The cost per point difference from the base"))
		content.AddChild(numField)

		text = i18n.Text("SM Reduction")
		content.AddChild(widget.NewFieldLeadingLabel(text))
		numField = widget.NewPercentageField(p.dockable.targetMgr, p.def.KeyPrefix+"sm", text,
			func() int { return fxp.As[int](p.def.CostAdjPercentPerSM) },
			func(v int) { p.def.CostAdjPercentPerSM = fxp.From(v) },
			0, 80, false, false)
		numField.Tooltip = unison.NewTooltipWithText(i18n.Text("The reduction in cost for each SM greater than 0"))
		content.AddChild(numField)
	}

	text = i18n.Text("Attribute Type")
	content.AddChild(widget.NewFieldLeadingLabel(text))
	content.AddChild(widget.NewPopup[attribute.Type](p.dockable.targetMgr, p.def.KeyPrefix+"type", text,
		func() attribute.Type { return p.def.Type },
		func(typ attribute.Type) { p.applyAttributeType(typ) },
		attribute.AllType...))

	if p.def.Type == attribute.Pool {
		p.poolPanel = newPoolPanel(p.dockable, p.def)
		content.AddChild(p.poolPanel)
	} else {
		p.poolPanel = nil
	}
	return content
}

func (p *attrDefPanel) validateAttrID(attrID string) bool {
	if key := strings.TrimSpace(strings.ToLower(attrID)); key != "" {
		if key != id.Sanitize(key, false, gurps.ReservedIDs...) {
			return false
		}
		if key == p.def.DefID {
			return true
		}
		_, exists := p.dockable.defs.Set[key]
		return !exists
	}
	return false
}

func (p *attrDefPanel) applyAttributeType(attrType attribute.Type) {
	p.def.Type = attrType
	if p.def.Type == attribute.Pool && len(p.def.Thresholds) == 0 {
		p.def.Thresholds = append(p.def.Thresholds, &gurps.PoolThreshold{KeyPrefix: p.dockable.targetMgr.NextPrefix()})
	} else if p.def.IsSeparator() {
		p.def.FullName = ""
		p.def.AttributeBase = ""
		p.def.CostPerPoint = 0
		p.def.CostAdjPercentPerSM = 0
		p.def.Thresholds = nil
	}
	p.dockable.sync()
}
