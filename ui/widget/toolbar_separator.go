package widget

import "github.com/richardwilkes/unison"

// NewToolbarSeparator creates a new separator for the toolbar.
func NewToolbarSeparator(margin float32) *unison.Separator {
	spacer := unison.NewSeparator()
	spacer.Vertical = true
	spacer.SetBorder(unison.NewEmptyBorder(unison.NewHorizontalInsets(margin)))
	spacer.SetLayoutData(&unison.FlexLayoutData{
		VAlign: unison.FillAlignment,
	})
	return spacer
}
