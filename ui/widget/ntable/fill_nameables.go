package ntable

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
)

// ProcessNameablesForSelection processes the selected rows and their children for any nameables.
func ProcessNameablesForSelection[T gurps.NodeConstraint[T]](table *unison.Table[*Node[T]], sel []*Node[T]) {
	var rows []T
	var nameables []map[string]string
	for _, row := range sel {
		gurps.Traverse[T](func(row T) bool {
			m := make(map[string]string)
			row.FillWithNameableKeys(m)
			if len(m) > 0 {
				rows = append(rows, row)
				nameables = append(nameables, m)
			}
			return false
		}, false, false, row.Data())
	}
	if len(rows) > 0 {
		list := unison.NewPanel()
		list.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(unison.StdHSpacing)))
		list.SetLayout(&unison.FlexLayout{
			Columns:  2,
			HSpacing: unison.StdHSpacing,
			VSpacing: unison.StdVSpacing,
		})
		for i := range rows {
			keys := make([]string, 0, len(nameables[i]))
			for k := range nameables[i] {
				keys = append(keys, k)
			}
			txt.SortStringsNaturalAscending(keys)
			if i != 0 {
				sep := unison.NewSeparator()
				sep.SetLayoutData(&unison.FlexLayoutData{
					HSpan:  2,
					HAlign: unison.FillAlignment,
					VAlign: unison.MiddleAlignment,
					HGrab:  true,
				})
				list.AddChild(sep)
			}
			for _, k := range keys {
				label := unison.NewLabel()
				label.Text = k
				label.SetLayoutData(&unison.FlexLayoutData{
					HAlign: unison.EndAlignment,
					VAlign: unison.MiddleAlignment,
				})
				list.AddChild(label)
				list.AddChild(createNameableField(k, nameables[i]))
			}
		}
		scroll := unison.NewScrollPanel()
		scroll.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
		scroll.SetContent(list, unison.FillBehavior, unison.FillBehavior)
		scroll.BackgroundInk = unison.ContentColor
		scroll.SetLayoutData(&unison.FlexLayoutData{
			HAlign: unison.FillAlignment,
			VAlign: unison.FillAlignment,
			HGrab:  true,
			VGrab:  true,
		})
		panel := unison.NewPanel()
		panel.SetLayout(&unison.FlexLayout{
			Columns:  1,
			HSpacing: unison.StdHSpacing,
			VSpacing: unison.StdVSpacing,
			HAlign:   unison.FillAlignment,
			VAlign:   unison.FillAlignment,
		})
		label := unison.NewLabel()
		label.Text = i18n.Text("Provide substitutions:")
		panel.AddChild(label)
		panel.AddChild(scroll)
		if unison.QuestionDialogWithPanel(panel) == unison.ModalResponseOK {
			var undo *unison.UndoEdit[*TableUndoEditData[T]]
			mgr := unison.UndoManagerFor(table)
			if mgr != nil {
				undo = &unison.UndoEdit[*TableUndoEditData[T]]{
					ID:         unison.NextUndoID(),
					EditName:   i18n.Text("Substitutions"),
					UndoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.BeforeData.Apply() },
					RedoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.AfterData.Apply() },
					AbsorbFunc: func(e *unison.UndoEdit[*TableUndoEditData[T]], other unison.Undoable) bool { return false },
					BeforeData: NewTableUndoEditData(table),
				}
			}
			for i, row := range rows {
				row.ApplyNameableKeys(nameables[i])
			}
			if mgr != nil && undo != nil {
				undo.AfterData = NewTableUndoEditData(table)
				mgr.Add(undo)
			}
			unison.Ancestor[widget.Rebuildable](table).Rebuild(true)
		}
	}
}

func createNameableField(key string, m map[string]string) *unison.Field {
	field := unison.NewField()
	field.SetMinimumTextWidthUsing("Something reasonable")
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
		HGrab:  true,
	})
	m[key] = ""
	field.ModifiedCallback = func() {
		m[key] = field.Text()
	}
	return field
}
