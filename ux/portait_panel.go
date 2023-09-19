/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"image"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/unison"
	"golang.org/x/image/draw"
)

const maxPortraitDimension = 400

// PortraitPanel holds the contents of the portrait block on the sheet.
type PortraitPanel struct {
	unison.Panel
	entity *gurps.Entity
}

// NewPortraitPanel creates a new portrait panel.
func NewPortraitPanel(entity *gurps.Entity) *PortraitPanel {
	p := &PortraitPanel{entity: entity}
	p.Self = p
	p.SetLayoutData(&unison.FlexLayoutData{VSpan: 2})
	p.SetBorder(&TitledBorder{Title: i18n.Text("Portrait")})
	p.Tooltip = newWrappedTooltip(i18n.Text(`Double-click to set a character portrait, or drag an image onto this block.`))
	p.DrawCallback = p.drawSelf
	p.FileDropCallback = p.fileDrop
	p.MouseDownCallback = p.mouseDown
	return p
}

func (p *PortraitPanel) drawSelf(gc *unison.Canvas, _ unison.Rect) {
	r := p.ContentRect(false)
	paint := unison.ContentColor.Paint(gc, r, unison.Fill)
	gc.DrawRect(r, paint)
	if img := p.entity.Profile.Portrait(); img != nil {
		size := img.LogicalSize()
		if size != r.Size {
			var scale float32
			if size.Width > size.Height {
				scale = r.Width / size.Width
			} else {
				scale = r.Height / size.Height
			}
			width := size.Width * scale
			r.X += (r.Width - width) / 2
			r.Width = width
			height := size.Height * scale
			r.Y += (r.Height - height) / 2
			r.Height = height
		}
		img.DrawInRect(gc, r, &unison.SamplingOptions{
			UseCubic:       true,
			CubicResampler: unison.MitchellResampler(),
			FilterMode:     unison.FilterModeLinear,
			MipMapMode:     unison.MipMapModeLinear,
		}, paint)
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
		d.SetAllowedExtensions(unison.KnownImageFormatExtensions...)
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
		var img *unison.Image
		if img, err = unison.NewImageFromBytes(data, 0.5); err != nil {
			errs.Log(errs.NewWithCause("does not appear to be a valid image", err), "file", f)
			continue
		}
		scale := float32(1)
		imgSize := img.Size()
		size := imgSize
		if size.Width > maxPortraitDimension || size.Height > maxPortraitDimension {
			if size.Width > size.Height {
				scale = maxPortraitDimension / size.Width
			} else {
				scale = maxPortraitDimension / size.Height
			}
			size = size.Mul(scale).Ceil().Max(unison.NewSize(1, 1))
		}
		if size != imgSize || !strings.HasSuffix(strings.ToLower(f), ".webp") {
			var src *image.NRGBA
			if src, err = img.ToNRGBA(); err != nil {
				errs.Log(errs.NewWithCause("unable to convert", err), "file", f)
				continue
			}
			dst := image.NewNRGBA(image.Rect(0, 0, int(size.Width), int(size.Height)))
			x := int((size.Width - imgSize.Width*scale) / 2)
			y := int((size.Height - imgSize.Height*scale) / 2)
			draw.CatmullRom.Scale(dst, image.Rect(x, y, x+int(size.Width), y+int(size.Height)), src, src.Rect, draw.Over, nil)
			if img, err = unison.NewImageFromPixels(int(size.Width), int(size.Height), dst.Pix, 0.5); err != nil {
				errs.Log(errs.NewWithCause("unable to create scaled image", err), "file", f, "size", size)
				continue
			}
			if data, err = img.ToWebp(80); err != nil {
				errs.Log(errs.NewWithCause("unable to create webp image", err), "file", f)
				continue
			}
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
