package ux

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
)

// SelectAttributeDefs presents a dialog for selecting attribute definitions for import.
func SelectAttributeDefs(defs *gurps.AttributeDefs) (selected *gurps.AttributeDefs, replace, canceled bool) {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})

	list := unison.NewPanel()
	list.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing)))
	list.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	for _, def := range defs.List(true) {
		if def.IsSeparator() {
			continue
		}
		box := unison.NewCheckBox()
		box.SetTitle(def.CombinedName())
		box.State = check.On
		box.SetEnabled(false)
		box.ClientData()["def"] = def
		list.AddChild(box)
	}

	rb1 := unison.NewRadioButton()
	rb1.SetTitle(i18n.Text("Remove existing attribute definitions, replacing them with all of these"))
	rb1.ClickCallback = func() { setChildrenEnablementState(list, false) }
	panel.AddChild(rb1)
	rb2 := unison.NewRadioButton()
	rb2.SetTitle(i18n.Text("Merge the checked attribute definitions into the existing ones"))
	rb2.ClickCallback = func() { setChildrenEnablementState(list, true) }
	panel.AddChild(rb2)
	group := unison.NewGroup()
	group.Add(rb1)
	group.Add(rb2)
	group.Select(rb1)

	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	scroll.SetContent(list, behavior.Fill, behavior.Fill)
	scroll.BackgroundInk = unison.ThemeSurface
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HSpan:  2,
		HGrab:  true,
		VGrab:  true,
	})
	panel.AddChild(scroll)

	dialog, err := unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon,
		unison.DefaultDialogTheme.QuestionIconInk, panel,
		[]*unison.DialogButtonInfo{
			unison.NewCancelButtonInfo(),
			unison.NewOKButtonInfoWithTitle(i18n.Text("Import")),
		})
	if err != nil {
		errs.Log(err)
		return nil, false, true
	}
	if dialog.RunModal() == unison.ModalResponseCancel {
		return nil, false, true
	}
	if group.Selected(rb1) {
		return defs, true, false
	}
	replacements := &gurps.AttributeDefs{Set: make(map[string]*gurps.AttributeDef)}
	for _, child := range list.Children() {
		if box, ok := child.Self.(*unison.CheckBox); ok && box.State == check.On {
			if rawDef, exists := box.ClientData()["def"]; exists {
				if def, ok2 := rawDef.(*gurps.AttributeDef); ok2 {
					replacements.Set[def.ID()] = def
				}
			}
		}
	}
	return replacements, false, false
}

func setChildrenEnablementState(panel unison.Paneler, enabled bool) {
	for _, child := range panel.AsPanel().Children() {
		child.SetEnabled(enabled)
	}
}
