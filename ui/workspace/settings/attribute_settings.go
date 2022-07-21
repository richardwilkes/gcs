package settings

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/attribute"
	"github.com/richardwilkes/gcs/v5/model/id"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/slices"
)

var _ widget.GroupedCloser = &attributesDockable{}

type attributesDockable struct {
	Dockable
	owner         widget.EntityPanel
	undoMgr       *unison.UndoManager
	defs          *gurps.AttributeDefs
	originalCRC   uint64
	content       *unison.Panel
	applyButton   *unison.Button
	cancelButton  *unison.Button
	promptForSave bool
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
		d := &attributesDockable{
			owner:         owner,
			promptForSave: true,
		}
		d.Self = d
		if owner != nil {
			d.defs = d.owner.Entity().SheetSettings.Attributes.Clone()
			d.TabTitle = i18n.Text("Attributes: " + owner.Entity().Profile.Name)
		} else {
			d.defs = settings.Global().Sheet.Attributes.Clone()
			d.TabTitle = i18n.Text("Default Attributes")
		}
		d.originalCRC = d.defs.CRC64()
		d.Extensions = []string{".attr", ".attributes", ".gas"}
		d.undoMgr = unison.NewUndoManager(100, func(err error) { jot.Error(err) })
		d.Loader = d.load
		d.Saver = d.save
		d.Resetter = d.reset
		d.ModifiedCallback = d.modified
		d.WillCloseCallback = d.willClose
		d.Setup(ws, dc, d.addToStartToolbar, nil, d.initContent)
	}
}

func (d *attributesDockable) UndoManager() *unison.UndoManager {
	return d.undoMgr
}

func (d *attributesDockable) modified() bool {
	modified := d.originalCRC != d.defs.CRC64()
	d.applyButton.SetEnabled(modified)
	d.cancelButton.SetEnabled(modified)
	return modified
}

func (d *attributesDockable) willClose() bool {
	if d.promptForSave && d.originalCRC != d.defs.CRC64() {
		switch unison.YesNoCancelDialog(fmt.Sprintf(i18n.Text("Apply changes made to\n%s?"), d.Title()), "") {
		case unison.ModalResponseDiscard:
		case unison.ModalResponseOK:
			d.apply()
		case unison.ModalResponseCancel:
			return false
		}
	}
	return true
}

func (d *attributesDockable) CloseWithGroup(other unison.Paneler) bool {
	return d.owner != nil && d.owner == other
}

func (d *attributesDockable) addToStartToolbar(toolbar *unison.Panel) {
	d.applyButton = unison.NewSVGButton(res.CheckmarkSVG)
	d.applyButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Apply Changes"))
	d.applyButton.SetEnabled(false)
	d.applyButton.ClickCallback = func() {
		d.apply()
		d.promptForSave = false
		d.AttemptClose()
	}
	toolbar.AddChild(d.applyButton)

	d.cancelButton = unison.NewSVGButton(res.NotSVG)
	d.cancelButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Discard Changes"))
	d.cancelButton.SetEnabled(false)
	d.cancelButton.ClickCallback = func() {
		d.promptForSave = false
		d.AttemptClose()
	}
	toolbar.AddChild(d.cancelButton)
}

func (d *attributesDockable) initContent(content *unison.Panel) {
	d.content = content
	content.SetBorder(nil)
	content.SetLayout(&unison.FlexLayout{Columns: 1})
	for _, def := range d.defs.List() {
		content.AddChild(d.createAttrDefPanel(def))
	}
}

func (d *attributesDockable) createAttrDefPanel(def *gurps.AttributeDef) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetBorder(unison.NewEmptyBorder(unison.Insets{
		Top:    unison.StdVSpacing,
		Left:   unison.StdHSpacing,
		Bottom: unison.StdVSpacing,
		Right:  unison.StdHSpacing * 2,
	}))
	panel.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		color := unison.ContentColor
		if panel.Parent().IndexOfChild(panel)%2 == 1 {
			color = unison.BandingColor
		}
		gc.DrawRect(rect, color.Paint(gc, rect, unison.Fill))
	}
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
		HGrab:  true,
	})

	panel.AddChild(d.createButtons(def))
	panel.AddChild(d.createContent(def))
	return panel
}

func (d *attributesDockable) createButtons(def *gurps.AttributeDef) *unison.Panel {
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

	deleteButton := unison.NewSVGButton(res.TrashSVG)
	deleteButton.ClickCallback = func() {
		attrsPanel := buttons.Parent()
		attrPanel := attrsPanel.Parent()
		attrsPanel.RemoveFromParent()
		children := attrPanel.Children()
		if len(children) == 1 {
			children[0].Children()[0].Children()[0].SetEnabled(false)
		}
		undo := &unison.UndoEdit[*gurps.AttributeDefs]{
			ID:         unison.NextUndoID(),
			EditName:   i18n.Text("Delete Attribute"),
			UndoFunc:   func(e *unison.UndoEdit[*gurps.AttributeDefs]) { d.applyAttrDefs(e.BeforeData) },
			RedoFunc:   func(e *unison.UndoEdit[*gurps.AttributeDefs]) { d.applyAttrDefs(e.AfterData) },
			AbsorbFunc: func(e *unison.UndoEdit[*gurps.AttributeDefs], other unison.Undoable) bool { return false },
		}
		undo.BeforeData = d.defs.Clone()
		delete(d.defs.Set, def.DefID)
		undo.AfterData = d.defs.Clone()
		d.UndoManager().Add(undo)
		d.MarkModified()
	}
	buttons.AddChild(deleteButton)

	addButton := unison.NewSVGButton(res.CircledAddSVG)
	addButton.ClickCallback = func() {
		undo := &unison.UndoEdit[[]*gurps.PoolThreshold]{
			ID:         unison.NextUndoID(),
			EditName:   i18n.Text("Add Pool Threshold"),
			UndoFunc:   func(e *unison.UndoEdit[[]*gurps.PoolThreshold]) { d.applyPoolThresholds(def, e.BeforeData) },
			RedoFunc:   func(e *unison.UndoEdit[[]*gurps.PoolThreshold]) { d.applyPoolThresholds(def, e.AfterData) },
			AbsorbFunc: func(e *unison.UndoEdit[[]*gurps.PoolThreshold], other unison.Undoable) bool { return false },
		}
		undo.BeforeData = clonePoolThresholds(def.Thresholds)
		threshold := &gurps.PoolThreshold{}
		def.Thresholds = append(def.Thresholds, threshold)
		poolPanel := buttons.Parent().Children()[1].Children()[2]
		thresholdPanel := d.createThresholdPanel(def, threshold)
		poolPanel.AddChild(thresholdPanel)
		children := poolPanel.Children()
		if len(children) == 2 {
			children[0].Children()[0].Children()[0].SetEnabled(true)
		}
		thresholdPanel.Children()[1].RequestFocus()
		undo.AfterData = clonePoolThresholds(def.Thresholds)
		d.UndoManager().Add(undo)
		d.MarkModified()
	}
	addButton.SetEnabled(def.Type == attribute.Pool)
	buttons.AddChild(addButton)
	return buttons
}

func (d *attributesDockable) applyAttrDefs(defs *gurps.AttributeDefs) {
	d.defs = defs.Clone()
	d.sync()
}

func (d *attributesDockable) createContent(def *gurps.AttributeDef) *unison.Panel {
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
	content.AddChild(d.createFirstLine(def))
	content.AddChild(d.createSecondLine(def))
	if def.Type == attribute.Pool {
		content.AddChild(d.createPool(def))
	}
	return content
}

func (d *attributesDockable) createFirstLine(def *gurps.AttributeDef) *unison.Panel {
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
	field := widget.NewStringField(text,
		func() string { return def.DefID },
		func(s string) {
			if d.validateAttrID(s, def) {
				delete(d.defs.Set, def.DefID)
				def.DefID = strings.TrimSpace(strings.ToLower(s))
				d.defs.Set[def.DefID] = def
			}
		})
	field.ValidateCallback = func(field *widget.StringField, def *gurps.AttributeDef) func() bool {
		return func() bool { return d.validateAttrID(field.Text(), def) }
	}(field, def)
	field.SetMinimumTextWidthUsing("basic_speed")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("A unique ID for the attribute"))
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.FillAlignment})
	panel.AddChild(field)

	text = i18n.Text("Short Name")
	panel.AddChild(widget.NewFieldLeadingLabel(text))
	field = widget.NewStringField(text,
		func() string { return def.Name },
		func(s string) { def.Name = s })
	field.SetMinimumTextWidthUsing("Taste & Smell")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("The name of this attribute, often an abbreviation"))
	panel.AddChild(field)

	text = i18n.Text("Full Name")
	panel.AddChild(widget.NewFieldLeadingLabel(text))
	field = widget.NewStringField(text,
		func() string { return def.FullName },
		func(s string) { def.FullName = s })
	field.SetMinimumTextWidthUsing("Fatigue Points")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("The full name of this attribute (may be omitted, in which case the Short Name will be used instead)"))
	panel.AddChild(field)
	return panel
}

func (d *attributesDockable) validateAttrID(attrID string, def *gurps.AttributeDef) bool {
	if key := strings.TrimSpace(strings.ToLower(attrID)); key != "" {
		if key != id.Sanitize(key, false, gurps.ReservedIDs...) {
			return false
		}
		if key == def.DefID {
			return true
		}
		_, exists := d.defs.Set[key]
		return !exists
	}
	return false
}

func (d *attributesDockable) createSecondLine(def *gurps.AttributeDef) *unison.Panel {
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

	popup := unison.NewPopupMenu[attribute.Type]()
	for _, one := range attribute.AllType {
		popup.AddItem(one)
	}
	popup.Select(def.Type)
	popup.SelectionCallback = func(_ int, typ attribute.Type) {
		undo := &unison.UndoEdit[attribute.Type]{
			ID:         unison.NextUndoID(),
			EditName:   i18n.Text("Attribute Type"),
			UndoFunc:   func(e *unison.UndoEdit[attribute.Type]) { d.applyAttributeType(def, e.BeforeData) },
			RedoFunc:   func(e *unison.UndoEdit[attribute.Type]) { d.applyAttributeType(def, e.AfterData) },
			AbsorbFunc: func(e *unison.UndoEdit[attribute.Type], other unison.Undoable) bool { return false },
		}
		undo.BeforeData = def.Type
		d.applyAttributeType(def, typ)
		undo.AfterData = def.Type
		d.UndoManager().Add(undo)
	}
	panel.AddChild(popup)

	text := i18n.Text("Base")
	panel.AddChild(widget.NewFieldLeadingLabel(text))
	field := widget.NewStringField(text,
		func() string { return def.AttributeBase },
		func(s string) { def.AttributeBase = s })
	field.SetMinimumTextWidthUsing("floor($basic_speed)")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("The base value, which may be a number or a formula"))
	panel.AddChild(field)

	text = i18n.Text("Cost")
	panel.AddChild(widget.NewFieldLeadingLabel(text))
	numField := widget.NewIntegerField(text,
		func() int { return fxp.As[int](def.CostPerPoint) },
		func(v int) { def.CostPerPoint = fxp.From(v) },
		0, 9999, false, false)
	numField.Tooltip = unison.NewTooltipWithText(i18n.Text("The cost per point difference from the base"))
	panel.AddChild(numField)

	text = i18n.Text("SM Reduction")
	panel.AddChild(widget.NewFieldLeadingLabel(text))
	numField = widget.NewPercentageField(text,
		func() int { return fxp.As[int](def.CostAdjPercentPerSM) },
		func(v int) { def.CostAdjPercentPerSM = fxp.From(v) },
		0, 80, false, false)
	numField.Tooltip = unison.NewTooltipWithText(i18n.Text("The reduction in cost for each SM greater than 0"))
	panel.AddChild(numField)

	return panel
}

func (d *attributesDockable) applyAttributeType(def *gurps.AttributeDef, attrType attribute.Type) {
	if def.Type = attrType; def.Type == attribute.Pool {
		if len(def.Thresholds) == 0 {
			def.Thresholds = append(def.Thresholds, &gurps.PoolThreshold{})
		}
	}
	d.sync()
}

func (d *attributesDockable) createPool(def *gurps.AttributeDef) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
		HGrab:  true,
	})
	for _, threshold := range def.Thresholds {
		panel.AddChild(d.createThresholdPanel(def, threshold))
	}
	return panel
}

func (d *attributesDockable) createThresholdPanel(def *gurps.AttributeDef, threshold *gurps.PoolThreshold) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetBorder(unison.NewEmptyBorder(unison.Insets{
		Top:    unison.StdVSpacing,
		Left:   unison.StdHSpacing,
		Bottom: unison.StdVSpacing,
		Right:  unison.StdHSpacing,
	}))
	panel.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		color := unison.ContentColor
		if panel.Parent().IndexOfChild(panel)%2 == 1 {
			color = unison.BandingColor
		}
		gc.DrawRect(rect, color.Paint(gc, rect, unison.Fill))
	}
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
		HGrab:  true,
	})

	panel.AddChild(d.createThresholdButtons(def))
	panel.AddChild(d.createThesholdContent(threshold))
	return panel
}

func (d *attributesDockable) createThresholdButtons(def *gurps.AttributeDef) *unison.Panel {
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

	deleteButton := unison.NewSVGButton(res.TrashSVG)
	deleteButton.ClickCallback = func() {
		thresholdPanel := buttons.Parent()
		poolPanel := thresholdPanel.Parent()
		i := poolPanel.IndexOfChild(thresholdPanel)
		thresholdPanel.RemoveFromParent()
		children := poolPanel.Children()
		if len(children) == 1 {
			children[0].Children()[0].Children()[0].SetEnabled(false)
		}
		undo := &unison.UndoEdit[[]*gurps.PoolThreshold]{
			ID:         unison.NextUndoID(),
			EditName:   i18n.Text("Delete Pool Threshold"),
			UndoFunc:   func(e *unison.UndoEdit[[]*gurps.PoolThreshold]) { d.applyPoolThresholds(def, e.BeforeData) },
			RedoFunc:   func(e *unison.UndoEdit[[]*gurps.PoolThreshold]) { d.applyPoolThresholds(def, e.AfterData) },
			AbsorbFunc: func(e *unison.UndoEdit[[]*gurps.PoolThreshold], other unison.Undoable) bool { return false },
		}
		undo.BeforeData = clonePoolThresholds(def.Thresholds)
		def.Thresholds = slices.Delete(def.Thresholds, i, i+1)
		undo.AfterData = clonePoolThresholds(def.Thresholds)
		d.UndoManager().Add(undo)
		d.MarkModified()
	}
	deleteButton.SetEnabled(len(def.Thresholds) > 1)
	buttons.AddChild(deleteButton)
	return buttons
}

func (d *attributesDockable) applyPoolThresholds(def *gurps.AttributeDef, thresholds []*gurps.PoolThreshold) {
	def.Thresholds = clonePoolThresholds(thresholds)
	d.sync()
}

func clonePoolThresholds(in []*gurps.PoolThreshold) []*gurps.PoolThreshold {
	thresholds := make([]*gurps.PoolThreshold, len(in))
	for i, one := range in {
		thresholds[i] = one.Clone()
	}
	return thresholds
}

func (d *attributesDockable) createThesholdContent(threshold *gurps.PoolThreshold) *unison.Panel {
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
	content.AddChild(d.createFirstThresholdLine(threshold))
	content.AddChild(d.createSecondThresholdLine(threshold))
	content.AddChild(d.createThirdThresholdLine(threshold))
	return content
}

func (d *attributesDockable) createFirstThresholdLine(threshold *gurps.PoolThreshold) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  4,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
		HGrab:  true,
	})

	text := i18n.Text("State")
	panel.AddChild(widget.NewFieldLeadingLabel(text))
	field := widget.NewStringField(text,
		func() string { return threshold.State },
		func(s string) { threshold.State = s })
	field.SetMinimumTextWidthUsing("Unconscious")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("A short description of the threshold state"))
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.FillAlignment})
	panel.AddChild(field)

	text = i18n.Text("Threshold")
	panel.AddChild(widget.NewFieldLeadingLabel(text))
	field = widget.NewStringField(text,
		func() string { return threshold.Expression },
		func(s string) { threshold.Expression = s })
	field.SetMinimumTextWidthUsing("round($self*100/50+20)")
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("An expression to calculate the threshold value"))
	panel.AddChild(field)
	return panel
}

func (d *attributesDockable) createSecondThresholdLine(threshold *gurps.PoolThreshold) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.StartAlignment,
		VAlign: unison.StartAlignment,
	})

	for _, op := range attribute.AllThresholdOp[1:] {
		panel.AddChild(createOpCheckBox(threshold, op))
	}
	return panel
}

func createOpCheckBox(threshold *gurps.PoolThreshold, op attribute.ThresholdOp) *unison.CheckBox {
	checkBox := widget.NewCheckBox(op.String(), threshold.ContainsOp(op), func(b bool) {
		if b {
			threshold.AddOp(op)
		} else {
			threshold.RemoveOp(op)
		}
	})
	checkBox.Tooltip = unison.NewTooltipWithText(op.AltString())
	return checkBox
}

func (d *attributesDockable) createThirdThresholdLine(threshold *gurps.PoolThreshold) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
		HGrab:  true,
	})

	text := i18n.Text("Explanation")
	panel.AddChild(widget.NewFieldLeadingLabel(text))
	field := widget.NewMultiLineStringField(text,
		func() string { return threshold.Explanation },
		func(s string) { threshold.Explanation = s })
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("A explanation of the effects of the threshold state"))
	panel.AddChild(field)
	return panel
}

func (d *attributesDockable) reset() {
	if d.owner != nil {
		d.defs = settings.Global().Sheet.Attributes.Clone()
	} else {
		d.defs = gurps.FactoryAttributeDefs()
	}
	d.sync()
}

func (d *attributesDockable) sync() {
	scrollRoot := d.content.ScrollRoot()
	h, v := scrollRoot.Position()
	d.content.RemoveAllChildren()
	for _, def := range d.defs.List() {
		d.content.AddChild(d.createAttrDefPanel(def))
	}
	scrollRoot.SetPosition(h, v)
	d.MarkForLayoutAndRedraw()
	d.MarkModified()
}

func (d *attributesDockable) load(fileSystem fs.FS, filePath string) error {
	defs, err := gurps.NewAttributeDefsFromFile(fileSystem, filePath)
	if err != nil {
		return err
	}
	d.defs = defs
	d.sync()
	return nil
}

func (d *attributesDockable) save(filePath string) error {
	return d.defs.Save(filePath)
}

func (d *attributesDockable) apply() {
	d.Window().FocusNext() // Intentionally move the focus to ensure any pending edits are flushed
	if d.owner == nil {
		settings.Global().Sheet.Attributes = d.defs.Clone()
		return
	}
	entity := d.owner.Entity()
	entity.SheetSettings.Attributes = d.defs.Clone()
	for attrID, def := range entity.SheetSettings.Attributes.Set {
		if attr, exists := entity.Attributes.Set[attrID]; exists {
			attr.Order = def.Order
		} else {
			entity.Attributes.Set[attrID] = gurps.NewAttribute(entity, attrID, def.Order)
		}
	}
	for attrID := range entity.Attributes.Set {
		if _, exists := d.defs.Set[attrID]; !exists {
			delete(entity.Attributes.Set, attrID)
		}
	}
	for _, wnd := range unison.Windows() {
		if ws := workspace.FromWindow(wnd); ws != nil {
			ws.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
				for _, one := range dc.Dockables() {
					if s, ok := one.(gurps.SheetSettingsResponder); ok {
						s.SheetSettingsUpdated(entity, true)
					}
				}
				return false
			})
		}
	}
}
