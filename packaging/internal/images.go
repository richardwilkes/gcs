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

package internal

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

//go:embed embedded/app-1024.png
var appImgBytes []byte
var appImg image.Image

//go:embed embedded/doc-1024.png
var docImgBytes []byte
var docImg image.Image

func loadBaseImages() error {
	var err error
	if appImg, _, err = image.Decode(bytes.NewBuffer(appImgBytes)); err != nil {
		return errs.Wrap(err)
	}
	if docImg, _, err = image.Decode(bytes.NewBuffer(docImgBytes)); err != nil {
		return errs.Wrap(err)
	}
	return nil
}

// createImageFromSVG turns one of our svg-as-a-path objects into an actual SVG document, then renders it into an image
// at the specified square size. Note that this is not currently GPU accelerated, as I haven't added the necessary bits
// to unison to support scribbling into arbitrary offscreen images yet.
func createImageFromSVG(fi *model.FileInfo, size int) (image.Image, error) {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %f %f"><path d="%s"/></svg>`,
		fi.SVG.Size().Width, fi.SVG.Size().Height, fi.SVG.PathScaledTo(1).ToSVGString())
	svg, err := oksvg.ReadIconStream(&buffer)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	svg.SetTarget(0, 0, float64(size), float64(size))
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	svg.Draw(rasterx.NewDasher(size, size, rasterx.NewScannerGV(size, size, img, img.Bounds())), 1)
	return img, nil
}
