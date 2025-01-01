// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package imgutil

import (
	"image"

	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/unison"
	"golang.org/x/image/draw"
)

const maxPortraitDimension = 400

// ConvertForPortraitUse converts the image data for use as a portrait.
func ConvertForPortraitUse(imageData []byte) ([]byte, error) {
	img, err := unison.NewImageFromBytes(imageData, 0.5)
	if err != nil {
		return nil, errs.NewWithCause("does not appear to be a valid image", err)
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
	if size != imgSize || !isWEBP(imageData) {
		var src *image.NRGBA
		if src, err = img.ToNRGBA(); err != nil {
			return nil, errs.NewWithCause("unable to convert", err)
		}
		dst := image.NewNRGBA(image.Rect(0, 0, int(size.Width), int(size.Height)))
		x := int((size.Width - imgSize.Width*scale) / 2)
		y := int((size.Height - imgSize.Height*scale) / 2)
		draw.CatmullRom.Scale(dst, image.Rect(x, y, x+int(size.Width), y+int(size.Height)), src, src.Rect, draw.Over, nil)
		if img, err = unison.NewImageFromPixels(int(size.Width), int(size.Height), dst.Pix, 0.5); err != nil {
			return nil, errs.NewWithCause("unable to create scaled image", err)
		}
		if imageData, err = img.ToWebp(80, true); err != nil {
			return nil, errs.NewWithCause("unable to create webp image", err)
		}
	}
	return imageData, nil
}

func isWEBP(imageData []byte) bool {
	return len(imageData) > 12 && imageData[0] == 'R' && imageData[1] == 'I' && imageData[2] == 'F' && imageData[3] == 'F' && imageData[8] == 'W' && imageData[9] == 'E' && imageData[10] == 'B' && imageData[11] == 'P'
}
