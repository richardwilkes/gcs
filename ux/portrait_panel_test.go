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
	"image"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/canvas/codecs"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// Space and a press open a file chooser, so only the keys and requests the portrait ignores are exercised.
func TestPortraitPanelIsAKeyboardControl(t *testing.T) {
	c := check.New(t)
	p := NewPortraitPanel(gurps.NewEntity())
	c.True(p.Focusable(), "the portrait takes the keyboard focus")
	c.Equal(portraitPanelRefKey, p.RefKey, "the focus is put back on the portrait by its reference key")
	c.Equal(role.Image, p.Accessibility.Role)
	c.Equal("Portrait", p.Accessibility.Name)
	c.NotEqual("", p.Accessibility.Description, "how to change the portrait is described")
	node := &accessibility.Node{}
	p.Accessibility.Callback(node)
	c.True(node.Actions.Has(accessibility.Press), "a screen reader is offered the press")
	c.False(p.KeyDownCallback(unison.KeyA, mod.None, false), "keys other than the control action are left to others")
	c.False(p.Accessibility.ActionCallback(accessibility.ActionRequest{Action: accessibility.ScrollIntoView}),
		"requests other than the press are left to unison")
}

// fileDragInfo is a drag.Info carrying file paths and nothing else, as a drag of files from the desktop does.
type fileDragInfo struct {
	paths []string
}

func (f *fileDragInfo) SourceDragOpMask() drag.Op { return drag.Copy }
func (f *fileDragInfo) DataTypes() []string       { return nil }
func (f *fileDragInfo) HasString() bool           { return false }
func (f *fileDragInfo) HasFilePaths() bool        { return len(f.paths) != 0 }
func (f *fileDragInfo) HasURLs() bool             { return false }
func (f *fileDragInfo) HasDataType(_ string) bool { return false }
func (f *fileDragInfo) Text() string              { return "" }
func (f *fileDragInfo) FilePaths() []string       { return f.paths }
func (f *fileDragInfo) URLs() []*url.URL          { return nil }
func (f *fileDragInfo) Data(_ string) []byte      { return nil }

// writePNGFile writes a PNG image of the given size to a file in dir and returns the file's path.
func writePNGFile(t *testing.T, dir, name string, width, height int) string {
	t.Helper()
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image.NewNRGBA(image.Rect(0, 0, width, height))); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, buffer.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// An image file dropped on the portrait becomes the portrait, kept as WEBP and scaled down, if need be, to fit
// maxPortraitDimension with its proportions kept, as an edit that can be undone. A drag of files none of which is an
// image is declined, and the files in it that are not images are passed over.
func TestPortraitPanelTakesADroppedImage(t *testing.T) {
	c := check.New(t)
	codecs.Register() // The app registers the image codecs as it starts, which a test without a workspace skips.
	sheet := newTestSheetForTemplate(t)
	profile := &sheet.Entity().Profile
	p, ok := sheet.blocks[gurps.BlockPortraitKey].(*PortraitPanel)
	if !ok {
		t.Fatal("the sheet must have a portrait block")
	}
	dir := t.TempDir()
	text := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(text, []byte("not an image"), 0o600); err != nil {
		t.Fatal(err)
	}
	notImages := &fileDragInfo{paths: []string{text}}
	c.False(p.acceptableDrag(notImages), "a drag without an image is declined")
	c.False(p.drop(notImages, geom.Point{}, mod.None), "and dropping it does nothing")
	c.Equal(0, len(profile.PortraitData))

	var small []byte
	for _, one := range []struct {
		name          string
		width, height int
		want          geom.Size
	}{
		{name: "small.png", width: 100, height: 50, want: geom.NewSize(100, 50)},
		{
			name: "large.png", width: 2 * maxPortraitDimension, height: maxPortraitDimension,
			want: geom.NewSize(maxPortraitDimension, maxPortraitDimension/2),
		},
	} {
		drop := &fileDragInfo{paths: []string{text, writePNGFile(t, dir, one.name, one.width, one.height)}}
		c.True(p.acceptableDrag(drop), "%s: a drag holding an image is taken", one.name)
		c.True(p.drop(drop, geom.Point{}, mod.None), "%s: and dropping it is handled", one.name)
		c.Equal(".webp", profile.PortraitExtension(), "%s: the portrait is kept as WEBP", one.name)
		portrait := profile.Portrait()
		if portrait == nil {
			t.Fatalf("%s: the portrait must be readable", one.name)
		}
		c.Equal(one.want, portrait.Size(), "%s: the size of the portrait", one.name)
		if small == nil {
			small = profile.PortraitData
		}
	}

	sheet.undoMgr.Undo()
	c.True(bytes.Equal(small, profile.PortraitData), "undo puts back the portrait before")
	sheet.undoMgr.Undo()
	c.Equal(0, len(profile.PortraitData), "and then none")
}
