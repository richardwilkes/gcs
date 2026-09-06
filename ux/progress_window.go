// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// newProgressWindow builds the small floating window a long-running operation reports on itself in: a label saying what
// is happening, a progress bar and, unless cancel is nil, a Cancel button that disables itself, rewrites the label to
// say the operation is being canceled and calls cancel. The window is packed and placed over the active window, ready
// for RunModal, which is what disposes of it.
func newProgressWindow(windowTitle, labelTitle string, bar *unison.ProgressBar, cancel func()) (wnd *unison.Window, label *unison.Label, err error) {
	frame := windowPlacementFrame()
	if wnd, err = unison.NewWindow(windowTitle, unison.FloatingWindowOption(), unison.NotResizableWindowOption(),
		unison.UndecoratedWindowOption(), unison.TransientWindowOption()); err != nil {
		return nil, nil, err
	}
	content := unison.NewPanel()
	content.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.NewUniformInsets(1), false), unison.NewEmptyBorder(geom.NewUniformInsets(2*unison.StdHSpacing))))
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
	})
	label = unison.NewLabel()
	label.SetTitle(labelTitle)
	content.AddChild(label)
	bar.SetLayoutData(&unison.FlexLayoutData{
		MinSize: geom.Size{Width: 500},
		HAlign:  align.Fill,
		HGrab:   true,
	})
	content.AddChild(bar)
	if cancel != nil {
		cancelButton := unison.NewButton()
		cancelButton.SetTitle(i18n.Text("Cancel"))
		cancelButton.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End})
		cancelButton.ClickCallback = func() {
			cancelButton.SetEnabled(false)
			label.SetTitle(i18n.Text("Canceling…"))
			cancel()
		}
		content.AddChild(cancelButton)
	}
	wnd.SetContent(content)
	wnd.Pack()
	wndFrame := wnd.FrameRect()
	frame.Y += (frame.Height - wndFrame.Height) / 3
	frame.Height = wndFrame.Height
	frame.X += (frame.Width - wndFrame.Width) / 2
	frame.Width = wndFrame.Width
	frame = frame.Align()
	wnd.SetFrameRect(unison.BestDisplayForRect(frame).FitRectOnto(frame))
	wnd.ToFront()
	return wnd, label, nil
}

// runInBackground starts a goroutine that performs work while the UI thread waits inside RunModal(), hands the result
// to that thread through resultChan and only then calls finish, which is what eventually stops the modal loop.
// resultChan must be buffered so that the send can never block, since nothing can receive from it until the modal loop
// has been stopped. The ordering is what makes the hand-off safe: the caller receives from resultChan only after
// RunModal() has returned, so the send is guaranteed to have completed first. Letting the goroutine write variables
// shared with the UI thread instead would allow a failed operation to be observed as a success.
func runInBackground[T any](resultChan chan<- T, work func() T, finish func()) {
	go func() {
		var result T
		defer func() {
			resultChan <- result
			finish()
		}()
		result = work()
	}()
}
