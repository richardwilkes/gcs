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
		VAlign: unison.StartAlignment,
		HGrab:  true,
	})

	p.AddChild(widget.NewDragHandle(map[string]any{attributesDragDataKey: &attributesDragData{
		owner: dockable.Entity(),
		def:   def,
	}}))
	p.AddChild(p.createButtons())
	p.AddChild(p.createContent(def))
	return p
}

func (p *attrDefPanel) createButtons() *unison.Panel {
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
	p.deleteButton.ClickCallback = p.deleteAttrDef
	buttons.AddChild(p.deleteButton)

	p.addThresholdButton = unison.NewSVGButton(res.CircledAddSVG)
	p.addThresholdButton.ClickCallback = func() { p.poolPanel.addThreshold() }
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

func (p *attrDefPanel) createContent(def *gurps.AttributeDef) *unison.Panel {
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
		HGrab:  true,
	})
	content.AddChild(p.createFirstLine(def))
	content.AddChild(p.createSecondLine(def))
	if def.Type == attribute.Pool {
		p.poolPanel = newPoolPanel(p.dockable, def)
		content.AddChild(p.poolPanel)
	} else {
		p.poolPanel = nil
	}
	return content
}

func (p *attrDefPanel) createFirstLine(def *gurps.AttributeDef) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  6,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
		HGrab:  true,
	})

	text := i18n.Text("ID")
	panel.AddChild(widget.NewFieldLeadingLabel(text))
	field := widget.NewStringField(p.dockable.targetMgr, def.KeyPrefix+"id", text,
		func() string { return def.DefID },
		func(s string) {
			if p.validateAttrID(s) {
				delete(p.dockable.defs.Set, def.DefID)
				def.DefID = strings.TrimSpace(strings.ToLower(s))
				p.dockable.defs.Set[def.DefID] = def
			}
		})
	field.ValidateCallback = func(field *widget.StringField, def *gurps.AttributeDef) func() bool {
		return func() bool { return p.validateAttrID(field.Text()) }
	}(field, def)
	field.SetMinimumTextWidthUsing("basic_speed")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("A unique ID for the attribute"))
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.FillAlignment})
	panel.AddChild(field)

	text = i18n.Text("Short Name")
	panel.AddChild(widget.NewFieldLeadingLabel(text))
	field = widget.NewStringField(p.dockable.targetMgr, def.KeyPrefix+"name", text,
		func() string { return def.Name },
		func(s string) { def.Name = s })
	field.SetMinimumTextWidthUsing("Taste & Smell")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("The name of this attribute, often an abbreviation"))
	panel.AddChild(field)

	text = i18n.Text("Full Name")
	panel.AddChild(widget.NewFieldLeadingLabel(text))
	field = widget.NewStringField(p.dockable.targetMgr, def.KeyPrefix+"fullname", text,
		func() string { return def.FullName },
		func(s string) { def.FullName = s })
	field.SetMinimumTextWidthUsing("Fatigue Points")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("The full name of this attribute (may be omitted, in which case the Short Name will be used instead)"))
	panel.AddChild(field)
	return panel
}

func (p *attrDefPanel) createSecondLine(def *gurps.AttributeDef) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  7,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
		HGrab:  true,
	})

	panel.AddChild(widget.NewPopup[attribute.Type](p.dockable.targetMgr, def.KeyPrefix+"type", i18n.Text("Attribute Type"),
		func() attribute.Type { return def.Type },
		func(typ attribute.Type) { p.applyAttributeType(typ) },
		attribute.AllType...))

	text := i18n.Text("Base")
	panel.AddChild(widget.NewFieldLeadingLabel(text))
	field := widget.NewStringField(p.dockable.targetMgr, def.KeyPrefix+"base", text,
		func() string { return def.AttributeBase },
		func(s string) { def.AttributeBase = s })
	field.SetMinimumTextWidthUsing("floor($basic_speed)")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("The base value, which may be a number or a formula"))
	panel.AddChild(field)

	text = i18n.Text("Cost")
	panel.AddChild(widget.NewFieldLeadingLabel(text))
	numField := widget.NewIntegerField(p.dockable.targetMgr, def.KeyPrefix+"cost", text,
		func() int { return fxp.As[int](def.CostPerPoint) },
		func(v int) { def.CostPerPoint = fxp.From(v) },
		0, 9999, false, false)
	numField.Tooltip = unison.NewTooltipWithText(i18n.Text("The cost per point difference from the base"))
	panel.AddChild(numField)

	text = i18n.Text("SM Reduction")
	panel.AddChild(widget.NewFieldLeadingLabel(text))
	numField = widget.NewPercentageField(p.dockable.targetMgr, def.KeyPrefix+"sm", text,
		func() int { return fxp.As[int](def.CostAdjPercentPerSM) },
		func(v int) { def.CostAdjPercentPerSM = fxp.From(v) },
		0, 80, false, false)
	numField.Tooltip = unison.NewTooltipWithText(i18n.Text("The reduction in cost for each SM greater than 0"))
	panel.AddChild(numField)

	return panel
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
	if p.def.Type = attrType; p.def.Type == attribute.Pool && len(p.def.Thresholds) == 0 {
		p.def.Thresholds = append(p.def.Thresholds, &gurps.PoolThreshold{KeyPrefix: p.dockable.targetMgr.NextPrefix()})
	}
	p.dockable.sync()
}
