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
	"bytes"
	"context"
	"encoding/json/v2"
	"io"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xhttp"
	"github.com/richardwilkes/toolbox/v2/xio"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

type rule struct {
	Rule     string
	Category []string
	Book     string
	Page     string
	Link     string
}

func downloadRulesLookupFile() {
	unableMsg := i18n.Text("Unable to download the GURPS Rules Lookup data.")
	frame := windowPlacementFrame()
	wnd, err := unison.NewWindow(i18n.Text("Downloading…"), unison.FloatingWindowOption(),
		unison.NotResizableWindowOption(), unison.UndecoratedWindowOption(), unison.TransientWindowOption())
	if err != nil {
		Workspace.ErrorHandler(unableMsg, err)
		return
	}
	content := unison.NewPanel()
	content.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.NewUniformInsets(1), false), unison.NewEmptyBorder(geom.NewUniformInsets(2*unison.StdHSpacing))))
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
	})
	label := unison.NewLabel()
	label.SetTitle(i18n.Text("Downloading the GURPS Rules Lookup data…"))
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
	rules := make(map[string][]*rule)
	go func() {
		defer wnd.StopModal(unison.ModalResponseOK)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		const gurpsRulesLookupURL = "https://raw.githubusercontent.com/StefanLeng/GURPSRulesLookup/refs/heads/master/src/assets/gurps_rules.json"
		var data []byte
		if data, err = xhttp.RetrieveData(ctx, nil, gurpsRulesLookupURL); err != nil {
			return
		}
		var r io.Reader
		if r, err = xio.NewBOMStripper(bytes.NewBuffer(data)); err != nil {
			return
		}
		var all []*rule
		if err = json.UnmarshalRead(r, &all); err != nil {
			err = errs.Wrap(err)
			return
		}
		for _, r := range all {
			rules[r.Book] = append(rules[r.Book], r)
		}
	}()
	wnd.RunModal()
	if err != nil {
		Workspace.ErrorHandler(unableMsg, err)
		return
	}
	dialog := unison.NewSaveDialog()
	settings := gurps.GlobalSettings()
	dialog.SetInitialDirectory(settings.LastDir(gurps.RulesLookupLastDirKey))
	dialog.SetAllowedExtensions(gurps.NotesExt)
	dialog.SetInitialFileName("GURPS Rules Lookup")
	if dialog.RunModal() {
		if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), gurps.NotesExt, false); ok {
			settings.SetLastDir(gurps.RulesLookupLastDirKey, filepath.Dir(filePath))
			filePath = filepath.Clean(filePath)
			for _, one := range AllDockables() {
				if tc, ok2 := one.(unison.TabCloser); ok2 {
					var fbd FileBackedDockable
					if fbd, ok2 = one.(FileBackedDockable); ok2 {
						if filepath.Clean(fbd.BackingFilePath()) == filePath {
							if !tc.MayAttemptClose() || !tc.AttemptClose() {
								unison.WarningDialogWithMessage(i18n.Text("Download canceled"),
									i18n.Text("Cannot update the file while it is open."))
								return
							}
							break
						}
					}
				}
			}
			idMap := make(map[string]tid.TID)
			var existing []*gurps.Note
			if existing, err = gurps.NewNotesFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath)); err == nil {
				for _, n := range existing {
					if n.Container() {
						idMap[n.MarkDown] = n.ID()
					}
				}
			}
			notes := make([]*gurps.Note, 0, len(rules))
			for book, list := range rules {
				parent := gurps.NewNote(nil, nil, true)
				if id, ok2 := idMap[book]; ok2 {
					parent.TID = id
				}
				parent.MarkDown = book
				children := make([]*gurps.Note, 0, len(list))
				for _, rule := range list {
					note := gurps.NewNote(nil, parent, false)
					note.MarkDown = rule.Rule
					note.Tags = rule.Category
					note.PageRef = rule.Link
					slices.Sort(note.Tags)
					children = append(children, note)
				}
				slices.SortFunc(children, func(a, b *gurps.Note) int {
					return xstrings.NaturalCmp(a.MarkDown, b.MarkDown, true)
				})
				parent.Children = children
				notes = append(notes, parent)
			}
			slices.SortFunc(notes, func(a, b *gurps.Note) int {
				return xstrings.NaturalCmp(a.MarkDown, b.MarkDown, true)
			})
			if err = gurps.SaveNotes(notes, filePath); err != nil {
				Workspace.ErrorHandler(unableMsg, err)
			}
			Workspace.Navigator.EventuallyReload()
		}
	}
}
