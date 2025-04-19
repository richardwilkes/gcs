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
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/imgutil"
	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/filtermode"
	"github.com/richardwilkes/unison/enums/imgfmt"
	"github.com/richardwilkes/unison/enums/mipmapmode"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// PortraitPanel holds the contents of the portrait block on the sheet.
type PortraitPanel struct {
	unison.Panel
	entity      *gurps.Entity
	mouseIsOver bool
}

// NewPortraitPanel creates a new portrait panel.
func NewPortraitPanel(entity *gurps.Entity) *PortraitPanel {
	p := &PortraitPanel{entity: entity}
	p.Self = p
	p.SetLayoutData(&unison.FlexLayoutData{VSpan: 2})
	p.SetBorder(&TitledBorder{Title: i18n.Text("Portrait")})
	p.DrawCallback = p.drawSelf
	p.FileDropCallback = p.fileDrop
	p.MouseDownCallback = p.mouseDown
	p.MouseEnterCallback = func(_ unison.Point, _ unison.Modifiers) bool {
		p.mouseIsOver = true
		p.MarkForRedraw()
		return false
	}
	p.MouseExitCallback = func() bool {
		p.mouseIsOver = false
		p.MarkForRedraw()
		return false
	}
	InstallTintFunc(p, colors.TintPortrait)
	return p
}

func (p *PortraitPanel) drawSelf(gc *unison.Canvas, _ unison.Rect) {
	r := p.ContentRect(false)
	paint := unison.ThemeBelowSurface.Paint(gc, r, paintstyle.Fill)
	gc.DrawRect(r, paint)
	if img := p.entity.Profile.Portrait(); img != nil {
		size := img.LogicalSize()
		pr := r
		if size != pr.Size {
			var scale float32
			if size.Width > size.Height {
				scale = pr.Width / size.Width
			} else {
				scale = pr.Height / size.Height
			}
			width := size.Width * scale
			pr.X += (pr.Width - width) / 2
			pr.Width = width
			height := size.Height * scale
			pr.Y += (pr.Height - height) / 2
			pr.Height = height
		}
		img.DrawInRect(gc, pr, &unison.SamplingOptions{
			UseCubic:       true,
			CubicResampler: unison.MitchellResampler(),
			FilterMode:     filtermode.Linear,
			MipMapMode:     mipmapmode.Linear,
		}, paint)
	}
	if p.mouseIsOver {
		gc.DrawRect(r, unison.Black.SetAlphaIntensity(0.3).Paint(gc, r, paintstyle.Fill))
		text := unison.NewTextWrappedLines(i18n.Text("Drop an image here or double-click to change the portrait"),
			&unison.TextDecoration{
				Font:            fonts.PageFieldPrimary,
				OnBackgroundInk: unison.White,
			}, r.Width-unison.StdHSpacing*2)
		var height float32
		for _, line := range text {
			height += line.Height()
		}
		y := r.Y + (r.Height-height)/2 + text[0].Baseline()
		for _, line := range text {
			size := line.Extents()
			line.Draw(gc, r.X+(r.Width-size.Width)/2, y)
			y += size.Height
		}
	}
}

// Sync the panel to the current data.
func (p *PortraitPanel) Sync() {
	// Nothing to do
}

func (p *PortraitPanel) mouseDown(_ unison.Point, button, clickCount int, _ unison.Modifiers) bool {
	if button == unison.ButtonLeft && clickCount == 2 {
		d := unison.NewOpenDialog()
		d.SetAllowsMultipleSelection(false)
		d.SetResolvesAliases(true)
		d.SetAllowedExtensions(imgfmt.AllReadableExtensions()...)
		d.SetCanChooseDirectories(false)
		d.SetCanChooseFiles(true)
		global := gurps.GlobalSettings()
		d.SetInitialDirectory(global.LastDir(gurps.ImagesLastDirKey))
		if d.RunModal() {
			file := d.Path()
			global.SetLastDir(gurps.ImagesLastDirKey, filepath.Dir(file))
			p.fileDrop([]string{file})
		}
	}
	return true
}

func (p *PortraitPanel) fileDrop(files []string) {
	for _, f := range files {
		data, err := xio.RetrieveData(f)
		if err != nil {
			errs.Log(errs.NewWithCause("unable to load", err), "file", f)
			continue
		}
		if data, err = imgutil.ConvertForPortraitUse(data); err != nil {
			errs.Log(err, "file", f)
			continue
		}
		sheet := unison.Ancestor[*Sheet](p)
		sheet.undoMgr.Add(&unison.UndoEdit[[]byte]{
			ID:         unison.NextUndoID(),
			EditName:   i18n.Text("Set Portrait"),
			UndoFunc:   func(edit *unison.UndoEdit[[]byte]) { sheet.updatePortrait(edit.BeforeData) },
			RedoFunc:   func(edit *unison.UndoEdit[[]byte]) { sheet.updatePortrait(edit.AfterData) },
			BeforeData: sheet.entity.Profile.PortraitData,
			AfterData:  data,
		})
		sheet.updatePortrait(data)
	}
}
