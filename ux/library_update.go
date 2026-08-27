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
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// libraryUpdateTimeout bounds the whole update. It is generous because a library is tens of megabytes of archive and a
// great many files to write, and some connections are slow; the Cancel button, not this, is how an impatient user stops
// it.
const libraryUpdateTimeout = 30 * time.Minute

func initiateLibraryUpdate(lib *gurps.Library, rel *gurps.Release) bool {
	if unison.QuestionDialog(fmt.Sprintf(i18n.Text("Update %s to %s?"), lib.Data().Title, filterVersion(rel.Version)),
		i18n.Text(`Existing content for this library will be removed and replaced.
Content in other libraries will not be modified`)) != unison.ModalResponseOK {
		return false
	}
	var list []unison.TabCloser
	libData := lib.Data()
	p := libData.PathOnDisk + "/"
	for _, one := range AllDockables() {
		if tc, ok := one.(unison.TabCloser); ok {
			var fbd FileBackedDockable
			if fbd, ok = one.(FileBackedDockable); ok {
				if strings.HasPrefix(fbd.BackingFilePath(), p) {
					list = append(list, tc)
				}
			}
		}
	}
	for _, one := range list {
		if !one.MayAttemptClose() || !one.AttemptClose() {
			unison.WarningDialogWithMessage(i18n.Text("Update canceled"),
				i18n.Text(`The library cannot be updated while
documents from the library are open.`))
			return false
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), libraryUpdateTimeout)
	defer cancel()
	// canceling is written and read only on the UI thread: the Cancel button's callback, and the task the progress
	// reporter posts. That is what keeps the label from being rewritten with the phase that was already underway when
	// the user asked to stop.
	canceling := false
	progress := unison.NewProgressBar(progressResolution)
	wnd, label, err := newLibraryProgressWindow(i18n.Text("Updating…"),
		libraryPhaseTitle(gurps.LibraryUpdateDownloading, libData.Title, rel.Version), progress, func() {
			canceling = true
			cancel()
		})
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to update"), err)
		return false
	}
	resultChan := make(chan error, 1)
	reportProgress := libraryUpdateProgress(label, progress, libData.Title, rel.Version, func() bool { return canceling })
	go runLibraryUpdate(resultChan, func() error { return performLibraryUpdate(ctx, lib, rel, reportProgress) },
		func() { finishLibraryUpdate(wnd, lib) })
	wnd.RunModal()
	if err = <-resultChan; err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			// The user asked for this, so there is nothing to report. The library itself is untouched: a failed update
			// puts the previous content back before it returns. A deadline that expired is deliberately not treated the
			// same way, since nobody asked for that and it needs saying.
			return false
		}
		Workspace.ErrorHandler(i18n.Text("Unable to update"), err)
		return false
	}
	return true
}

// newLibraryProgressWindow builds the small floating window the library operations report on themselves in: a label
// saying what is happening, a progress bar, and a Cancel button that disables itself, rewrites the label to say the
// operation is being canceled and calls cancel. The window is packed and placed over the active window, ready for
// RunModal, which is what disposes of it.
func newLibraryProgressWindow(windowTitle, labelTitle string, bar *unison.ProgressBar, cancel func()) (wnd *unison.Window, label *unison.Label, err error) {
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
	cancelButton := unison.NewButton()
	cancelButton.SetTitle(i18n.Text("Cancel"))
	cancelButton.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End})
	cancelButton.ClickCallback = func() {
		cancelButton.SetEnabled(false)
		label.SetTitle(i18n.Text("Canceling…"))
		cancel()
	}
	content.AddChild(cancelButton)
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

// runLibraryUpdate performs the download on a background goroutine while the UI thread waits inside RunModal(), hands
// the result to that thread through resultChan and only then calls finish, which is what eventually stops the modal
// loop. resultChan must be buffered so that the send can never block. The ordering is what makes the hand-off safe:
// initiateLibraryUpdate receives from resultChan only after RunModal() has returned, so the send is guaranteed to have
// completed first. Reading a plain variable instead would let a failed download be observed as a success.
func runLibraryUpdate(resultChan chan<- error, download func() error, finish func()) {
	var err error
	defer func() {
		resultChan <- err
		finish()
	}()
	err = download()
}

func performLibraryUpdate(ctx context.Context, lib *gurps.Library, rel *gurps.Release, progress gurps.LibraryUpdateProgress) error {
	lib.StopAllWatches()
	return lib.Download(ctx, &http.Client{}, rel, progress)
}

// libraryUpdateProgress returns a progress reporter that keeps the window's label and bar in step with the update. The
// fraction it receives is relative to the phase it comes with, so the bar starts over each time the phase changes,
// which is also when the label is rewritten to say what is now happening -- unless the user has asked to stop, in which
// case the label already says so and must be left alone. canceling is only ever consulted on the UI thread.
func libraryUpdateProgress(label *unison.Label, bar *unison.ProgressBar, title, version string, canceling func() bool) gurps.LibraryUpdateProgress {
	post := throttledProgress(bar)
	phase := gurps.LibraryUpdateDownloading
	return func(current gurps.LibraryUpdatePhase, fraction float64) {
		if current != phase {
			phase = current
			unison.InvokeTask(func() {
				if !canceling() {
					label.SetTitle(libraryPhaseTitle(current, title, version))
				}
			})
		}
		post(fraction)
	}
}

// libraryPhaseTitle describes what the update is doing at the moment.
func libraryPhaseTitle(phase gurps.LibraryUpdatePhase, title, version string) string {
	if phase == gurps.LibraryUpdateInstalling {
		return fmt.Sprintf(i18n.Text("Installing %s %s…"), title, filterVersion(version))
	}
	return fmt.Sprintf(i18n.Text("Downloading %s %s…"), title, filterVersion(version))
}

// finishLibraryUpdate refreshes the library's list of available releases and then posts the teardown of the progress
// window onto the UI thread. It runs on the background goroutine, so it must not touch the modal state itself:
// StopModal() mutates the window's modal fields and unison's package-global modal stack without any locking, all while
// the UI thread is spinning on them inside RunModal(). The release check is left here rather than being posted along
// with the teardown because it makes network calls that would otherwise stall the UI thread for up to a minute.
func finishLibraryUpdate(wnd *unison.Window, lib *gurps.Library) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	lib.CheckForAvailableUpgrade(ctx, &http.Client{})
	unison.InvokeTask(func() {
		Workspace.Navigator.EventuallyReload()
		wnd.StopModal(unison.ModalResponseOK)
	})
}
