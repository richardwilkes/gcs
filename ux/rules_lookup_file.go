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

// rulesLookupResult is the outcome of the download, handed from the download goroutine to the UI thread as a unit.
type rulesLookupResult struct {
	rules map[string][]*rule
	err   error
}

// runRulesLookupDownload performs the download on a background goroutine while the UI thread waits inside RunModal(),
// hands the result to that thread through resultChan and only then calls finish, which is what eventually stops the
// modal loop. resultChan must be buffered so that the send can never block, since nothing can receive from it until
// the modal loop has been stopped. The ordering is what makes the hand-off safe: downloadRulesLookupFile receives from
// resultChan only after RunModal() has returned, so the send is guaranteed to have completed first. Letting the
// goroutine write variables shared with the UI thread instead would allow a failed download to be observed as a
// success, silently writing an empty notes file rather than reporting the error.
func runRulesLookupDownload(resultChan chan<- rulesLookupResult, download func() (map[string][]*rule, error), finish func()) {
	var result rulesLookupResult
	defer func() {
		resultChan <- result
		finish()
	}()
	result.rules, result.err = download()
}

// retrieveRulesLookupData downloads the GURPS Rules Lookup data and returns its rules grouped by book.
func retrieveRulesLookupData() (map[string][]*rule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	data, err := xhttp.RetrieveData(ctx, nil,
		"https://raw.githubusercontent.com/StefanLeng/GURPSRulesLookup/refs/heads/master/src/assets/gurps_rules.json")
	if err != nil {
		return nil, err
	}
	return parseRulesLookupData(data)
}

// parseRulesLookupData parses the downloaded GURPS Rules Lookup data and returns its rules grouped by book.
func parseRulesLookupData(data []byte) (map[string][]*rule, error) {
	r, err := xio.NewBOMStripper(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	var all []*rule
	if err = json.UnmarshalRead(r, &all); err != nil {
		return nil, errs.Wrap(err)
	}
	rules := make(map[string][]*rule)
	for _, one := range all {
		rules[one.Book] = append(rules[one.Book], one)
	}
	return rules, nil
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
	resultChan := make(chan rulesLookupResult, 1)
	go runRulesLookupDownload(resultChan, retrieveRulesLookupData, func() {
		unison.InvokeTask(func() { wnd.StopModal(unison.ModalResponseOK) })
	})
	wnd.RunModal()
	result := <-resultChan
	if result.err != nil {
		Workspace.ErrorHandler(unableMsg, result.err)
		return
	}
	rules := result.rules
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
