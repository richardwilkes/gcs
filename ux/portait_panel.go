/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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
	"fmt"
	"image"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/unison"
	"golang.org/x/image/draw"
)

// PortraitPanel holds the contents of the portrait block on the sheet.
type PortraitPanel struct {
	unison.Panel
	entity *model.Entity
}

// NewPortraitPanel creates a new portrait panel.
func NewPortraitPanel(entity *model.Entity) *PortraitPanel {
	p := &PortraitPanel{entity: entity}
	p.Self = p
	p.SetLayoutData(&unison.FlexLayoutData{VSpan: 2})
	p.SetBorder(&TitledBorder{Title: i18n.Text("Portrait")})
	p.Tooltip = unison.NewTooltipWithText(fmt.Sprintf(i18n.Text(`Double-click to set a character portrait, or drag an image onto this block.

The dimensions of the chosen picture should be in a ratio of 3 pixels wide
for every 4 pixels tall to fill the block (recommended %dx%d pixels).`), model.PortraitWidth*2, model.PortraitHeight*2))
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
		global := model.Global()
		d.SetInitialDirectory(global.LastDir(model.ImagesDirKey))
		if d.RunModal() {
			file := d.Path()
			global.SetLastDir(model.DefaultLastDirKey, filepath.Dir(file))
			p.fileDrop([]string{file})
		}
	}
	return true
}

func (p *PortraitPanel) fileDrop(files []string) {
	for _, f := range files {
		data, err := xio.RetrieveData(f)
		if err != nil {
			jot.Error(errs.NewWithCause("unable to load: "+f, err))
			continue
		}
		var img *unison.Image
		if img, err = unison.NewImageFromBytes(data, 0.5); err != nil {
			jot.Error(errs.NewWithCause("does not appear to be a valid image: "+f, err))
			continue
		}
		size := img.Size()
		if size.Width != model.PortraitWidth*2 || size.Height != model.PortraitHeight*2 {
			var src *image.NRGBA
			if src, err = img.ToNRGBA(); err != nil {
				jot.Error(errs.NewWithCause("unable to convert: "+f, err))
				continue
			}
			dst := image.NewNRGBA(image.Rect(0, 0, model.PortraitWidth*2, model.PortraitHeight*2))
			if size.Width > model.PortraitWidth*2 || size.Height > model.PortraitHeight*2 {
				if size.Width > model.PortraitWidth*2 {
					factor := model.PortraitWidth * 2 / size.Width
					size.Width = model.PortraitWidth * 2
					size.Height = xmath.Max(xmath.Floor(size.Height*factor), 1)
				}
				if size.Height > model.PortraitHeight*2 {
					factor := model.PortraitHeight * 2 / size.Height
					size.Height = model.PortraitHeight * 2
					size.Width = xmath.Max(xmath.Floor(size.Width*factor), 1)
				}
				x := int((model.PortraitWidth*2 - size.Width) / 2)
				y := int((model.PortraitHeight*2 - size.Height) / 2)
				draw.CatmullRom.Scale(dst, image.Rect(x, y, x+int(size.Width), y+int(size.Height)), src, src.Rect,
					draw.Over, nil)
				src = dst
			} else {
				x := int((model.PortraitWidth*2 - size.Width) / 2)
				y := int((model.PortraitHeight*2 - size.Height) / 2)
				draw.Draw(dst, image.Rect(x, y, x+int(size.Width), y+int(size.Height)), src, image.Pt(0, 0), draw.Over)
			}
			if img, err = unison.NewImageFromPixels(model.PortraitWidth*2, model.PortraitHeight*2, dst.Pix, 0.5); err != nil {
				jot.Error(errs.NewWithCause("unable to create scaled image from: "+f, err))
				continue
			}
			if data, err = img.ToWebp(80); err != nil {
				jot.Error(errs.NewWithCause("unable to create webp image from: "+f, err))
				continue
			}
		}
		p.entity.Profile.PortraitData = data
		p.entity.Profile.PortraitImage = img
		p.MarkForRedraw()
		MarkModified(p)
	}
}
