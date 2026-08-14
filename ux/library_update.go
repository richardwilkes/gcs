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

	frame := windowPlacementFrame()
	wnd, err := unison.NewWindow(i18n.Text("Updating…"), unison.FloatingWindowOption(),
		unison.NotResizableWindowOption(), unison.UndecoratedWindowOption(), unison.TransientWindowOption())
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to update"), err)
		return false
	}
	content := unison.NewPanel()
	content.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.NewUniformInsets(1), false), unison.NewEmptyBorder(geom.NewUniformInsets(2*unison.StdHSpacing))))
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
	})
	label := unison.NewLabel()
	label.SetTitle(fmt.Sprintf(i18n.Text("Updating %s to %s…"), libData.Title, filterVersion(rel.Version)))
	content.AddChild(label)
	progress := unison.NewProgressBar(0)
	progress.SetLayoutData(&unison.FlexLayoutData{
		MinSize: geom.Size{Width: 500},
		HAlign:  align.Fill,
		HGrab:   true,
	})
	content.AddChild(progress)
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
	resultChan := make(chan error, 1)
	go runLibraryUpdate(resultChan, func() error { return performLibraryUpdate(lib, rel) },
		func() { finishLibraryUpdate(wnd, lib) })
	wnd.RunModal()
	if err = <-resultChan; err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to update"), err)
		return false
	}
	return true
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

func performLibraryUpdate(lib *gurps.Library, rel *gurps.Release) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	lib.StopAllWatches()
	return lib.Download(ctx, &http.Client{}, rel)
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
