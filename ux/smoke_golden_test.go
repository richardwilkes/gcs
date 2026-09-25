// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//go:build smoke

package ux

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
)

// A smoke test checks what the application shows at each step in two ways, each against a golden file kept in
// testdata/smoke/golden/<test name>:
//
//   - expectGUI compares a text description of what is on screen, taken from the accessibility tree, against
//     <name>.a11y.txt. It says what the controls are, what they hold and what state they are in, which is what most
//     steps want to know, and it diffs readably.
//   - expectScreenshot compares the pixels, of the whole screen or of one panel, against <name>.png.
//
// Run with -update to write the golden files from what the tests see instead of comparing against them. When a
// comparison fails, what the test saw is written to the directory named by GCS_SMOKE_ARTIFACTS (or gcs-smoke in the
// system's temporary directory), along with, for a screenshot, an image that marks the pixels that differ in red.

var updateSmokeGoldens = flag.Bool("update", false, "rewrite the smoke tests' golden files from what the tests see")

const smokeGoldenDir = "testdata/smoke/golden"

// goldenPath returns the path of this test's golden file with the given name.
func (s *smokeSession) goldenPath(name string) string {
	return filepath.Join(smokeGoldenDir, s.t.Name(), name)
}

// artifactPath returns the path to write what this test saw, for the golden file with the given name, creating the
// directory it goes in.
func (s *smokeSession) artifactPath(name string) string {
	dir := os.Getenv("GCS_SMOKE_ARTIFACTS")
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "gcs-smoke")
	}
	dir = filepath.Join(dir, s.t.Name())
	s.c.NoError(os.MkdirAll(dir, 0o750)) //nolint:gosec // G703: writing into the directory named in the environment is the point
	return filepath.Join(dir, name)
}

// writeGolden writes data to the golden file with the given name, creating the directory it goes in.
func (s *smokeSession) writeGolden(name string, data []byte) {
	p := s.goldenPath(name)
	s.c.NoError(os.MkdirAll(filepath.Dir(p), 0o750))
	s.c.NoError(os.WriteFile(p, data, 0o640))
}

// readGolden returns the content of the golden file with the given name, failing the test if there is none.
func (s *smokeSession) readGolden(name string) ([]byte, bool) {
	data, err := os.ReadFile(s.goldenPath(name))
	if err != nil {
		s.t.Errorf("unable to read golden file %s (run with -update to create it): %v", s.goldenPath(name), err)
		return nil, false
	}
	return data, true
}

// expectGUI compares the accessibility description of p, or of the whole workspace window when p is nil, against the
// golden file <name>.a11y.txt.
func (s *smokeSession) expectGUI(name string, p unison.Paneler) {
	s.t.Helper()
	golden := name + ".a11y.txt"
	actual := s.describeGUI(p)
	wnd := s.wnd
	if p != nil {
		s.screen.Do(func() { wnd = p.AsPanel().Window() })
	}
	s.checkInvariants(fmt.Sprintf("the %q snapshot", name), wnd)
	if *updateSmokeGoldens {
		s.writeGolden(golden, []byte(actual))
		return
	}
	expected, ok := s.readGolden(golden)
	if !ok {
		return
	}
	if string(expected) == actual {
		return
	}
	artifact := s.artifactPath(golden)
	s.c.NoError(os.WriteFile(artifact, []byte(actual), 0o640))
	s.t.Errorf("the GUI does not match %s; what the test saw is in %s\n%s", s.goldenPath(golden), artifact,
		firstDifference(string(expected), actual))
}

// describeGUI returns the accessibility description of p, or of the whole workspace window when p is nil: one line per
// node, indented by its depth, giving its role, name and value and whichever of its states are set. Nodes that carry
// nothing of their own are left out and their children take their place. Shortcuts are left out as well, since menu
// items show them in each platform's own style.
func (s *smokeSession) describeGUI(p unison.Paneler) string {
	s.t.Helper()
	wnd := s.wnd
	if p != nil {
		s.screen.Do(func() { wnd = p.AsPanel().Window() })
	}
	tree := s.screen.AccessibilityTree(wnd)
	if tree == nil {
		s.t.Fatal("no accessibility tree is available")
	}
	root := tree.Root
	if p != nil {
		node := s.screen.AccessibilityNodeFor(p)
		if node == nil {
			s.t.Fatal("the panel is not in the accessibility tree")
		}
		root = node.ID
	}
	var buf strings.Builder
	var walk func(id accessibility.NodeID, depth int)
	walk = func(id accessibility.NodeID, depth int) {
		n := tree.Node(id)
		if n == nil {
			return
		}
		if !n.Ignored {
			buf.WriteString(strings.Repeat("  ", depth))
			describeNode(&buf, n)
			buf.WriteByte('\n')
			depth++
		}
		for _, child := range n.Children {
			walk(child, depth)
		}
	}
	walk(root, 0)
	return buf.String()
}

// describeNode writes the one line describing n.
func describeNode(buf *strings.Builder, n *accessibility.Node) {
	buf.WriteString(n.Role.Key())
	if n.Name != "" {
		fmt.Fprintf(buf, " %q", n.Name)
	}
	if n.Value != "" {
		fmt.Fprintf(buf, " value=%q", n.Value)
	} else if n.Text != nil && n.Text.Text != "" {
		fmt.Fprintf(buf, " text=%q", n.Text.Text)
	}
	if n.Description != "" && n.Description != n.Name {
		fmt.Fprintf(buf, " desc=%q", n.Description)
	}
	if n.Placeholder != "" {
		fmt.Fprintf(buf, " placeholder=%q", n.Placeholder)
	}
	if n.HasCheck {
		fmt.Fprintf(buf, " checked=%s", n.Checked.Key())
	}
	for _, flag := range []struct {
		name string
		set  bool
	}{
		{"disabled", n.Disabled},
		{"focused", n.Focused},
		{"selected", n.Selected},
		{"pressed", n.Pressed},
		{"readonly", n.ReadOnly},
		{"modal", n.Modal},
		{"invalid", n.Invalid},
		{"offscreen", n.Offscreen},
		{"expanded", n.Expandable && n.Expanded},
		{"collapsed", n.Expandable && !n.Expanded},
	} {
		if flag.set {
			buf.WriteString(" [")
			buf.WriteString(flag.name)
			buf.WriteByte(']')
		}
	}
}

// firstDifference describes where two texts first part ways, by line, with a little of what follows.
func firstDifference(expected, actual string) string {
	want := strings.Split(expected, "\n")
	got := strings.Split(actual, "\n")
	for i := 0; i < len(want) || i < len(got); i++ {
		if i < len(want) && i < len(got) && want[i] == got[i] {
			continue
		}
		var buf strings.Builder
		fmt.Fprintf(&buf, "first difference at line %d:\n", i+1)
		for j := i; j < i+5; j++ {
			if j < len(want) {
				fmt.Fprintf(&buf, "  want: %s\n", want[j])
			}
			if j < len(got) {
				fmt.Fprintf(&buf, "  got:  %s\n", got[j])
			}
		}
		return buf.String()
	}
	return ""
}

// expectScreenshot compares the pixels of p, or of the whole screen when p is nil, against the golden file <name>.png.
func (s *smokeSession) expectScreenshot(name string, p unison.Paneler) {
	s.t.Helper()
	golden := name + ".png"
	actual := s.capture(p)
	if *updateSmokeGoldens {
		var buf bytes.Buffer
		s.c.NoError(png.Encode(&buf, actual))
		s.writeGolden(golden, buf.Bytes())
		return
	}
	data, ok := s.readGolden(golden)
	if !ok {
		return
	}
	expected, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		s.t.Errorf("unable to decode %s: %v", s.goldenPath(golden), err)
		return
	}
	diff, count, bounds := diffImages(expected, actual)
	if count == 0 {
		return
	}
	artifact := s.artifactPath(golden)
	s.writePNG(artifact, actual)
	diffPath := s.artifactPath(name + ".diff.png")
	if diff != nil {
		s.writePNG(diffPath, diff)
	}
	if count < 0 {
		s.t.Errorf("the screenshot is %v, but %s is %v; what the test saw is in %s", actual.Bounds().Size(),
			s.goldenPath(golden), expected.Bounds().Size(), artifact)
		return
	}
	s.t.Errorf("%d pixels differ from %s, within %v; what the test saw is in %s and the differences are marked in %s",
		count, s.goldenPath(golden), bounds, artifact, diffPath)
}

// capture returns the pixels of p, or of the whole screen when p is nil, once the application has gone quiet. The
// pointer is moved first to the empty end of the menu bar, where it highlights nothing, so that what is captured does
// not depend on where the last click was.
func (s *smokeSession) capture(p unison.Paneler) *image.NRGBA {
	s.t.Helper()
	size := s.screen.Size()
	s.screen.MouseMove(geom.NewPoint(size.Width-10, 10), 0)
	s.screen.Sync()
	img := s.screen.Capture()
	if img == nil {
		s.t.Fatal("the screen could not be captured")
	}
	if p == nil {
		return img
	}
	var r geom.Rect
	s.screen.Do(func() {
		panel := p.AsPanel()
		r = panel.RectToRoot(panel.ContentRect(true))
		r.Point = screenPoint(panel.Window(), r.Point)
	})
	scale := s.screen.Scale()
	crop := image.Rect(int(r.X*scale), int(r.Y*scale), int((r.X+r.Width)*scale), int((r.Y+r.Height)*scale))
	sub, ok := img.SubImage(crop).(*image.NRGBA)
	if !ok || sub.Bounds().Empty() {
		s.t.Fatalf("the panel's rect %v is not on the screen", r)
	}
	// Copied so that the result starts at the origin, as a decoded golden file does.
	out := image.NewNRGBA(image.Rect(0, 0, crop.Dx(), crop.Dy()))
	for y := range crop.Dy() {
		copy(out.Pix[y*out.Stride:y*out.Stride+crop.Dx()*4], sub.Pix[y*sub.Stride:y*sub.Stride+crop.Dx()*4])
	}
	return out
}

// writePNG writes img to the given path.
func (s *smokeSession) writePNG(path string, img image.Image) {
	var buf bytes.Buffer
	s.c.NoError(png.Encode(&buf, img))
	s.c.NoError(os.WriteFile(path, buf.Bytes(), 0o640))
}

// diffImages compares two images pixel by pixel. It returns an image of the expected one, faded, with the pixels that
// differ marked in red, along with how many differ and the bounds of those that do. A count of -1 means the images are
// not the same size, and no image is returned.
func diffImages(expected image.Image, actual *image.NRGBA) (diff *image.NRGBA, count int, bounds image.Rectangle) {
	if expected.Bounds().Size() != actual.Bounds().Size() {
		return nil, -1, image.Rectangle{}
	}
	eb := expected.Bounds()
	ab := actual.Bounds()
	diff = image.NewNRGBA(image.Rect(0, 0, eb.Dx(), eb.Dy()))
	for y := range eb.Dy() {
		for x := range eb.Dx() {
			want := color.NRGBAModel.Convert(expected.At(eb.Min.X+x, eb.Min.Y+y)).(color.NRGBA) //nolint:errcheck // Always NRGBA
			got := actual.NRGBAAt(ab.Min.X+x, ab.Min.Y+y)
			if want == got {
				want.A /= 4
				diff.SetNRGBA(x, y, want)
				continue
			}
			diff.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
			bounds = bounds.Union(image.Rect(x, y, x+1, y+1))
			count++
		}
	}
	return diff, count, bounds
}
