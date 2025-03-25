package ux

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/side"
)

func processPickerRows[T gurps.NodeTypes](rows []*Node[T]) (revised []*Node[T], abort bool) {
	for _, one := range ExtractNodeDataFromList(rows) {
		result, cancel := processPickerRow(one)
		if cancel {
			return nil, true
		}
		for _, replacement := range result {
			revised = append(revised, NewNodeLike(rows[0], replacement))
		}
	}
	return revised, false
}

func processPickerRow[T gurps.NodeTypes](row T) (revised []T, abort bool) {
	n := gurps.AsNode(row)
	if !n.Container() {
		return []T{row}, false
	}
	children := n.NodeChildren()
	tpp, ok := n.(gurps.TemplatePickerProvider)
	if !ok || tpp.TemplatePickerData().ShouldOmit() {
		rowChildren := make([]T, 0, len(children))
		for _, child := range children {
			var result []T
			result, abort = processPickerRow(child)
			if abort {
				return nil, true
			}
			rowChildren = append(rowChildren, result...)
		}
		n.SetChildren(rowChildren)
		SetParents(rowChildren, row)
		return []T{row}, false
	}
	tp := tpp.TemplatePickerData()

	list := unison.NewPanel()
	list.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(unison.StdHSpacing)))
	list.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})

	progress := unison.NewLabel()
	progressBackground := pickerMatchStateColor(tp.Qualifier.Matches(0))
	progress.SetBorder(unison.NewCompoundBorder(
		unison.NewEmptyBorder(unison.Insets{Top: unison.StdVSpacing * 2}),
		unison.NewEmptyBorder(unison.NewHorizontalInsets(unison.StdHSpacing)),
	))
	progress.Side = side.Right
	progress.OnBackgroundInk = progressBackground.On()
	progress.DrawCallback = func(gc *unison.Canvas, _ unison.Rect) {
		if tp.Type == picker.NotApplicable {
			return
		}
		r := progress.ContentRect(true)
		r.Y += unison.StdVSpacing * 2
		r.Height -= unison.StdVSpacing * 2
		gc.DrawRoundedRect(r, 8, 8, progressBackground.Paint(gc, r, paintstyle.Fill))
		progress.DefaultDraw(gc, r)
	}
	boxes := make([]*unison.CheckBox, 0, len(children))
	var dialog *unison.Dialog
	callback := func() {
		var total fxp.Int
		for i, box := range boxes {
			if box.State == check.On {
				switch tp.Type {
				case picker.NotApplicable:
				case picker.Count:
					total += fxp.One
				case picker.Points:
					total += rawPoints(children[i])
				}
			}
		}
		matches := tp.Qualifier.Matches(total)
		dialog.Button(unison.ModalResponseOK).SetEnabled(matches)
		if tp.Type != picker.NotApplicable {
			var img *unison.SVG
			if matches {
				img = unison.CheckmarkSVG
			} else {
				img = svg.Not
			}
			size := max(progress.Font.Baseline()-2, 6)
			progress.Drawable = &unison.DrawableSVG{
				SVG:  img,
				Size: unison.NewSize(size, size),
			}
			progressBackground = pickerMatchStateColor(matches)
			progress.OnBackgroundInk = progressBackground.On()
			progress.SetTitle(total.Comma())
			progress.MarkForLayoutRecursivelyUpward()
			progress.MarkForRedraw()
		}
	}
	for _, child := range children {
		boxes = addPickerRow(list, child, callback, boxes)
	}

	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.NewUniformInsets(1), false))
	scroll.SetContent(list, behavior.Fill, behavior.Fill)
	scroll.BackgroundInk = unison.ThemeSurface
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HSpan:  2,
		HGrab:  true,
		VGrab:  true,
	})

	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	label := unison.NewLabel()
	label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	label.SetTitle(row.String())
	panel.AddChild(label)
	if notesCapable, hasNotes := any(row).(interface{ Notes() string }); hasNotes {
		if notes := notesCapable.Notes(); notes != "" {
			label = unison.NewLabel()
			label.Font = fonts.FieldSecondary
			label.SetTitle(notes)
			label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
			panel.AddChild(label)
		}
	}
	label = unison.NewLabel()
	label.SetTitle(tp.Description())
	label.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: unison.StdVSpacing * 2}))
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Start,
		VAlign: align.Middle,
	})
	panel.AddChild(label)
	progress.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.End,
		VAlign: align.Middle,
	})
	panel.AddChild(progress)
	panel.AddChild(scroll)

	var err error
	dialog, err = unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon,
		unison.DefaultDialogTheme.QuestionIconInk, panel,
		[]*unison.DialogButtonInfo{
			unison.NewCancelButtonInfo(),
			{
				Title:        i18n.Text("Override"),
				ResponseCode: unison.ModalResponseUserBase,
			},
			unison.NewOKButtonInfo(),
		})
	if err != nil {
		errs.Log(err)
		return nil, true
	}
	callback()
	if dialog.RunModal() == unison.ModalResponseCancel {
		return nil, true
	}

	rowChildren := make([]T, 0, len(children))
	for i, box := range boxes {
		if box.State == check.On {
			var result []T
			result, abort = processPickerRow(children[i])
			if abort {
				return nil, true
			}
			rowChildren = append(rowChildren, result...)
		}
	}
	SetParents(rowChildren, n.Parent())
	return rowChildren, false
}

func pickerMatchStateColor(matches bool) unison.Color {
	if matches {
		return unison.Green
	}
	return unison.ThemeError.GetColor()
}

func addPickerRow[T gurps.NodeTypes](parent *unison.Panel, row T, callback func(), boxes []*unison.CheckBox) []*unison.CheckBox {
	checkBox := unison.NewCheckBox()
	updatePickerCheckBoxTitle(checkBox, row)
	checkBox.ClickCallback = callback
	parent.AddChild(checkBox)
	boxes = append(boxes, checkBox)
	var onClick func()
	switch actual := interface{}(row).(type) {
	case *gurps.Trait:
		if actual.IsLeveled() {
			onClick = func() { pickerRowLevelEditor(actual, checkBox, callback) }
		}
	case *gurps.Skill:
		if !actual.Container() {
			onClick = func() { pickerRowPointEditor(actual, checkBox, callback) }
		}
	case *gurps.Spell:
		if !actual.Container() {
			onClick = func() { pickerRowPointEditor(actual, checkBox, callback) }
		}
	}
	if onClick == nil {
		label := unison.NewLabel()
		label.SetTitle(" ")
		parent.AddChild(label)
	} else {
		button := NewSVGButtonForFont(svg.Edit, checkBox.Font, -2)
		button.ClickCallback = onClick
		parent.AddChild(button)
	}
	return boxes
}

func updatePickerCheckBoxTitle[T gurps.NodeTypes](checkBox *unison.CheckBox, row T) {
	title := row.String()
	points := rawPoints(row)
	if points != 0 {
		pointsLabel := i18n.Text("points")
		if points == fxp.One {
			pointsLabel = i18n.Text("point")
		}
		title += fmt.Sprintf(" [%s %s]", points.Comma(), pointsLabel)
	}
	checkBox.SetTitle(title)
}

func pickerRowLevelEditor(trait *gurps.Trait, checkBox *unison.CheckBox, callback func()) {
	levels := trait.Levels
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VAlign:   align.Middle,
	})
	label := unison.NewLabel()
	label.SetTitle(fmt.Sprintf(i18n.Text("%s Level"), trait.Kind()))
	panel.AddChild(label)
	panel.AddChild(NewDecimalField(nil, "", "", func() fxp.Int { return levels },
		func(value fxp.Int) { levels = value }, 0, fxp.MaxBasePoints, false, false))
	dialog, err := unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon,
		unison.DefaultDialogTheme.QuestionIconInk, panel,
		[]*unison.DialogButtonInfo{
			unison.NewCancelButtonInfo(),
			unison.NewOKButtonInfo(),
		})
	if err != nil {
		errs.Log(err)
		return
	}
	if dialog.RunModal() != unison.ModalResponseOK {
		return
	}
	trait.Levels = levels
	updatePickerCheckBoxTitle(checkBox, trait)
	callback()
	checkBox.MarkForLayoutRecursivelyUpward()
	checkBox.MarkForRedraw()
}

type pickerRowPointEditorTypes interface {
	*gurps.Skill | *gurps.Spell
	nameable.Applier
	fmt.Stringer
	Kind() string
	RawPoints() fxp.Int
	SetRawPoints(fxp.Int) bool
}

func pickerRowPointEditor[T pickerRowPointEditorTypes](node T, checkBox *unison.CheckBox, callback func()) {
	points := node.RawPoints()
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VAlign:   align.Middle,
	})
	label := unison.NewLabel()
	label.SetTitle(fmt.Sprintf(i18n.Text("%s Points"), node.Kind()))
	panel.AddChild(label)
	panel.AddChild(NewDecimalField(nil, "", "", func() fxp.Int { return points },
		func(value fxp.Int) { points = value }, 0, fxp.MaxBasePoints, false, false))
	dialog, err := unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon,
		unison.DefaultDialogTheme.QuestionIconInk, panel,
		[]*unison.DialogButtonInfo{
			unison.NewCancelButtonInfo(),
			unison.NewOKButtonInfo(),
		})
	if err != nil {
		errs.Log(err)
		return
	}
	if dialog.RunModal() != unison.ModalResponseOK {
		return
	}
	node.SetRawPoints(points)
	updatePickerCheckBoxTitle(checkBox, node)
	callback()
	checkBox.MarkForLayoutRecursivelyUpward()
	checkBox.MarkForRedraw()
}
