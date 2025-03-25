// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

func initiateLibraryUpdate(lib *gurps.Library, rel gurps.Release) bool {
	if unison.QuestionDialog(fmt.Sprintf(i18n.Text("Update %s to v%s?"), lib.Title, filterVersion(rel.Version)),
		i18n.Text(`Existing content for this library will be removed and replaced.
Content in other libraries will not be modified`)) != unison.ModalResponseOK {
		return false
	}
	var list []unison.TabCloser
	p := lib.PathOnDisk + "/"
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

	var frame unison.Rect
	if focused := unison.ActiveWindow(); focused != nil {
		frame = focused.FrameRect()
	} else {
		frame = unison.PrimaryDisplay().Usable
	}
	wnd, err := unison.NewWindow(i18n.Text("Updating…"), unison.FloatingWindowOption(),
		unison.NotResizableWindowOption(), unison.UndecoratedWindowOption(), unison.TransientWindowOption())
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to update"), err)
		return false
	}
	content := unison.NewPanel()
	content.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0,
		unison.NewUniformInsets(1), false), unison.NewEmptyBorder(unison.NewUniformInsets(2*unison.StdHSpacing))))
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
	})
	label := unison.NewLabel()
	label.SetTitle(fmt.Sprintf(i18n.Text("Updating %s to v%s…"), lib.Title, filterVersion(rel.Version)))
	content.AddChild(label)
	progress := unison.NewProgressBar(0)
	progress.SetLayoutData(&unison.FlexLayoutData{
		MinSize: unison.Size{Width: 500},
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
	go performLibraryUpdate(wnd, lib, rel, &err)
	wnd.RunModal()
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to update"), err)
		return false
	}
	return true
}

//nolint:gocritic // We need to return the error, but can't return it directly thanks to using a goroutine
func performLibraryUpdate(wnd *unison.Window, lib *gurps.Library, rel gurps.Release, err *error) {
	defer finishLibraryUpdate(wnd, lib)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	lib.StopAllWatches()
	*err = lib.Download(ctx, &http.Client{}, rel)
}

func finishLibraryUpdate(wnd *unison.Window, lib *gurps.Library) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	lib.CheckForAvailableUpgrade(ctx, &http.Client{})
	Workspace.Navigator.EventuallyReload()
	wnd.StopModal(unison.ModalResponseOK)
}
