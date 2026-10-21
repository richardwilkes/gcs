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
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// The benchmarks in this file measure what a character sheet with large lists costs to update, rebuild, draw and
// scroll, in a headless session. Run them with:
//
//	GOEXPERIMENT=simd go test ./ux -run '^$' -bench '^Benchmark(Sheet|PageList|NameablesPanel)' -benchmem -benchtime 2s \
//	    -count 6 -timeout 90m | tee baseline.txt
//
// For a quick look, add GCS_BENCH_SIZE=0.25 and -benchtime 5x. GCS_BENCH_FILE names a real sheet to measure instead
// of the generated one (see benchEntity); a benchmark of a list that sheet leaves off the page is skipped. Compare two
// runs with:
//
//	go run golang.org/x/perf/cmd/benchstat@latest before.txt after.txt
//
// Besides ns/op, which is the end-to-end time including the headless session's round trips, each benchmark reports
// ui-ns/op, the time the operation itself took on the UI thread, and draw-ns/op, the time the redraw it provoked
// took. Each benchmark starts one headless session and must therefore never run in parallel.

// benchFileEnvVar names the environment variable that, when set to the path of a character sheet, has the benchmarks
// measure that sheet instead of a generated one.
const benchFileEnvVar = "GCS_BENCH_FILE"

// sheetBench is a sheet open in a headless session, ready to be measured.
type sheetBench struct {
	screen *unison.HeadlessScreen
	wnd    *unison.Window
	sheet  *Sheet
	heapMB float64 // The heap in use once the sheet was open, which each measurement reports as heap-MB.
}

// benchEntity returns the character to measure: the sheet named by benchFileEnvVar, or a generated one of the size
// gurpstest.LargeEntitySizeFromEnv reports.
func benchEntity(b *testing.B) (entity *gurps.Entity, generated bool) {
	b.Helper()
	if path := os.Getenv(benchFileEnvVar); path != "" {
		e, err := gurps.NewEntityFromFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
		if err != nil {
			b.Fatalf("unable to load %s: %v", path, err)
		}
		return e, false
	}
	size, err := gurpstest.LargeEntitySizeFromEnv()
	if err != nil {
		b.Fatal(err)
	}
	return gurpstest.NewLargeEntity(&size), true
}

// newSheetBench starts a headless session, opens the character to measure in it as a sheet, checks that every list
// is on the page when the character was generated for that purpose, and performs one untimed update so that the
// first measurement doesn't pay for anything lazily built. It also records the heap in use once the sheet is open.
func newSheetBench(b *testing.B) *sheetBench {
	b.Helper()
	c := check.New(b)
	screen, wnd := uxtest.StartHeadlessWorkspace(b, c)
	entity, generated := benchEntity(b)
	var sheet *Sheet
	screen.Do(func() {
		sheet = NewSheet("bench"+gurps.SheetExt, entity)
		DisplayNewDockable(sheet)
	})
	if generated {
		for _, key := range gurps.AllBlockKeys {
			if !gurps.IsListBlockKey(key) {
				continue
			}
			list := sheet.list(key)
			if xreflect.IsNil(list) || !listAttachedToSheet(sheet, list) {
				b.Fatalf("the %s list is not on the page", key)
			}
		}
	}
	screen.Do(func() { sheet.MarkModified(sheet) })
	// Registered after the session's own cleanup, so that it runs first: the sheet is closed without a save prompt,
	// which a modified sheet would otherwise put up.
	b.Cleanup(func() {
		screen.Do(func() { sheet.markUnmodified() })
		uxtest.CloseEditorWithoutPrompt(b, screen, sheet)
	})
	runtime.GC()
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return &sheetBench{screen: screen, wnd: wnd, sheet: sheet, heapMB: float64(stats.HeapAlloc) / (1 << 20)}
}

// reportHeap reports the heap in use once the sheet was open. It has to be reported after the timed loop, since
// resetting the timer discards the metrics reported before it.
func (sb *sheetBench) reportHeap(b *testing.B) {
	b.ReportMetric(sb.heapMB, "heap-MB")
}

// run measures fn, which runs on the UI thread once per iteration, reporting the time it took there as ui-ns/op and
// the time of the redraw it provoked as draw-ns/op alongside the usual end-to-end ns/op.
func (sb *sheetBench) run(b *testing.B, fn func()) {
	b.Helper()
	var ui, draw, before time.Duration
	n := 0
	b.ResetTimer()
	for b.Loop() {
		sb.screen.Do(func() {
			// Do waits for the session to go quiet, so this is the redraw the previous iteration provoked, or, the
			// first time through, one that happened before the loop, which is taken back out below.
			last := sb.wnd.LastDrawDuration()
			if n == 0 {
				before = last
			}
			draw += last
			start := time.Now()
			fn()
			ui += time.Since(start)
		})
		n++
	}
	b.StopTimer()
	sb.screen.Do(func() { draw += sb.wnd.LastDrawDuration() })
	draw -= before
	b.ReportMetric(float64(ui.Nanoseconds())/float64(n), "ui-ns/op")
	b.ReportMetric(float64(draw.Nanoseconds())/float64(n), "draw-ns/op")
	sb.reportHeap(b)
}

// scrollTo brings the panel into view, so that drawing and scrolling measure the part of the page that holds it. It
// skips the benchmark when the panel isn't on the page, as is the case on a sheet loaded through benchFileEnvVar for an
// empty weapons list or a list its layout omits.
func (sb *sheetBench) scrollTo(b *testing.B, p unison.Paneler) {
	b.Helper()
	var onPage bool
	sb.screen.Do(func() {
		if onPage = listAttachedToSheet(sb.sheet, p); onPage {
			p.AsPanel().ScrollIntoView()
		}
	})
	if !onPage {
		b.Skip("not on the page")
	}
}

func BenchmarkSheetMarkModified(b *testing.B) {
	sb := newSheetBench(b)
	sb.run(b, func() { sb.sheet.MarkModified(sb.sheet) })
}

func BenchmarkSheetRebuild(b *testing.B) {
	sb := newSheetBench(b)
	for _, one := range []struct {
		name string
		full bool
	}{{"full", true}, {"partial", false}} {
		b.Run(one.name, func(b *testing.B) {
			sb.run(b, func() { sb.sheet.Rebuild(one.full) })
		})
	}
}

func BenchmarkSheetRedraw(b *testing.B) {
	sb := newSheetBench(b)
	for _, one := range []struct {
		name  string
		panel unison.Paneler
	}{{"top", sb.sheet.page}, {"melee", sb.sheet.MeleeWeapons}, {"spells", sb.sheet.Spells}} {
		b.Run(one.name, func(b *testing.B) {
			sb.scrollTo(b, one.panel)
			sb.run(b, func() {
				sb.wnd.MarkForRedraw()
				sb.wnd.FlushDrawing()
			})
		})
	}
}

// BenchmarkSheetScroll measures wheel scrolling over each of the lists that are costly to draw: twenty steps down
// and twenty back up per iteration, reporting the time per step.
func BenchmarkSheetScroll(b *testing.B) {
	sb := newSheetBench(b)
	const steps = 20
	for _, one := range []struct {
		name  string
		panel unison.Paneler
	}{
		{"melee", sb.sheet.MeleeWeapons},
		{"ranged", sb.sheet.RangedWeapons},
		{"skills", sb.sheet.Skills},
		{"spells", sb.sheet.Spells},
	} {
		b.Run(one.name, func(b *testing.B) {
			sb.scrollTo(b, one.panel)
			pt := sb.screen.PanelCenter(sb.sheet.scroll)
			var draw time.Duration
			n := 0
			b.ResetTimer()
			for b.Loop() {
				for range steps {
					sb.screen.Wheel(pt, geom.Point{Y: -3}, 0)
					draw += sb.wnd.LastDrawDuration()
				}
				for range steps {
					sb.screen.Wheel(pt, geom.Point{Y: 3}, 0)
					draw += sb.wnd.LastDrawDuration()
				}
				n++
			}
			b.StopTimer()
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(n*2*steps), "ns/step")
			b.ReportMetric(float64(draw.Nanoseconds())/float64(n*2*steps), "draw-ns/step")
			sb.reportHeap(b)
		})
	}
}

// BenchmarkPageListSync measures syncing each of the sheet's lists to its model, which is the bulk of an update.
func BenchmarkPageListSync(b *testing.B) {
	sb := newSheetBench(b)
	for _, key := range gurps.AllBlockKeys {
		if !gurps.IsListBlockKey(key) {
			continue
		}
		list := sb.sheet.list(key)
		if xreflect.IsNil(list) {
			continue
		}
		syncer, ok := list.(Syncer)
		if !ok {
			b.Fatalf("the %s list is not a Syncer", key)
		}
		b.Run(key, func(b *testing.B) {
			sb.run(b, syncer.Sync)
		})
	}
}

// BenchmarkSheetSearch measures the search refresh every update performs, with the field empty and with text in it.
func BenchmarkSheetSearch(b *testing.B) {
	sb := newSheetBench(b)
	for _, one := range []struct {
		name string
		text string
	}{{"empty", ""}, {"text", "spell"}} {
		b.Run(one.name, func(b *testing.B) {
			sb.screen.Do(func() { sb.sheet.searchTracker.searchField.SetText(one.text) })
			sb.run(b, sb.sheet.searchTracker.Refresh)
		})
	}
}

// BenchmarkNameablesPanel measures the Nameables section of the editor of a trait on the sheet whose modifiers share
// and are linked to its nameables: "sync" is the check it makes after every change in the editor, and "rebuild" builds
// its rows again, as each change to them does. The section isn't in a window, so both are timed on the UI thread
// alone, with nothing drawn.
func BenchmarkNameablesPanel(b *testing.B) {
	sb := newSheetBench(b)
	var p *nameablesPanel
	sb.screen.Do(func() {
		gurps.Traverse(func(t *gurps.Trait) bool {
			if len(t.NameableLinks) == 0 {
				return false
			}
			_, content := buildEditorContent(sb.sheet, t, initTraitEditor)
			p, _ = uxtest.FirstPanelOfType[*nameablesPanel](content)
			return true
		}, false, true, sb.sheet.entity.Traits...)
	})
	if p == nil {
		b.Skip("no trait on the sheet links its modifiers' nameables")
	}
	sb.screen.Do(func() {
		if p.collapse.collapsed {
			p.collapse.toggle()
		}
	})
	b.Run("sync", func(b *testing.B) {
		sb.screen.Do(func() {
			for b.Loop() {
				p.Sync()
			}
		})
	})
	b.Run("rebuild", func(b *testing.B) {
		sb.screen.Do(func() {
			for b.Loop() {
				p.RemoveAllChildren()
				p.build()
			}
		})
	})
}

// TestSheetBenchSkipsAListNotOnThePage verifies that a benchmark of a list that a loaded sheet leaves off the page, such
// as the ranged weapons of a character without any, is skipped rather than measuring whatever part of the page is in
// view.
func TestSheetBenchSkipsAListNotOnThePage(t *testing.T) {
	c := check.New(t)
	path := filepath.Join(t.TempDir(), "empty"+gurps.SheetExt)
	c.NoError(gurps.NewEntity().Save(path))
	t.Setenv(benchFileEnvVar, path)
	var opened, skipped bool
	testing.Benchmark(func(b *testing.B) {
		sb := newSheetBench(b)
		opened = true
		b.Run("ranged", func(b *testing.B) {
			defer func() { skipped = b.Skipped() }()
			sb.scrollTo(b, sb.sheet.RangedWeapons)
		})
	})
	c.True(opened, "the sheet must open")
	c.True(skipped, "a list that is not on the page must not be measured")
}
