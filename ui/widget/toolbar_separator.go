package widget

import "github.com/richardwilkes/unison"

// NewToolbarSeparator creates a new separator for the toolbar.
func NewToolbarSeparator() *unison.Separator {
	spacer := unison.NewSeparator()
	spacer.Vertical = true
	spacer.SetBorder(unison.NewEmptyBorder(unison.NewHorizontalInsets(unison.StdHSpacing)))
	spacer.SetLayoutData(&unison.FlexLayoutData{
		VAlign: unison.FillAlignment,
	})
	return spacer
}
