// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//go:build headlessapi

package ux

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/runmode"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// init registers the headless debug API as a runmode.Mode, which is how main learns about -headless-api at all: a
// build compiled without the headlessapi tag never runs this file, so runmode.Factories stays empty and main has no
// trace of any of it -- no flags, no help text, no mode name, nothing.
func init() {
	runmode.Factories = append(runmode.Factories, newHeadlessAPIRunMode)
}

// newHeadlessAPIRunMode registers -headless-api, -headless-api-width and -headless-api-height on flag.CommandLine as
// a side effect of being called, and returns the runmode.Mode that starts the headless debug API those flags
// configure.
func newHeadlessAPIRunMode(flagSet *flag.FlagSet) runmode.Mode {
	if flagSet == nil {
		flagSet = flag.CommandLine
	}
	addr := flagSet.String("headless-api", "", i18n.Text("Internal use only. Start the headless debug API on the given `address` (e.g. \"127.0.0.1:8642\") instead of the normal UI"))
	width := flagSet.Float64("headless-api-width", 1400, i18n.Text("Internal use only. Logical width of the virtual screen the headless debug API renders to"))
	height := flagSet.Float64("headless-api-height", 900, i18n.Text("Internal use only. Logical height of the virtual screen the headless debug API renders to"))
	return runmode.Mode{
		Name:            "headless-api",
		HiddenFlagNames: []string{"headless-api", "headless-api-width", "headless-api-height"},
		Requested:       func() bool { return *addr != "" },
		Start: func(files []string) {
			StartHeadlessAPI(*addr, float32(*width), float32(*height), files)
		},
	}
}

// StartHeadlessAPI starts GCS headless -- no real window ever appears -- with a debug HTTP server listening on addr
// that lets automated tooling drive and inspect the running app:
//
//   - POST /input       inject a click, double-click, drag, wheel, key press or typed text
//   - GET  /input       the canonical key and modifier names POST /input recognizes
//   - GET  /inspect     the widget at a point: type, absolute bounding rect, tooltip, enabled state, text
//   - GET  /inspect/focus  the same, for whatever currently holds keyboard focus (e.g. an open error dialog)
//   - GET  /screenshot  a PNG of the whole virtual screen, or of an absolute rectangle within it
//   - GET  /console     log output and session errors recorded since a given sequence number
//
// Coordinates everywhere in the protocol -- input points, inspected rects, screenshot rects -- are in the same
// logical, screen-absolute space: the one /inspect and /inspect/focus report a widget's rect in is exactly the one a
// screenshot rect or an input point should be given in, and windows other than the main one (such as a modal error
// dialog) are positioned within it, not at their own private origin.
//
// The virtual screen's size is fixed for the life of the session; there is no live resize.
//
// It never returns.
func StartHeadlessAPI(addr string, width, height float32, files []string) {
	console := &consoleBuffer{}
	slog.SetDefault(slog.New(&teeLogHandler{orig: slog.Default().Handler(), buf: console}))

	if width <= 0 {
		width = 1400
	}
	if height <= 0 {
		height = 900
	}

	server := &headlessAPIServer{console: console}
	screen, err := unison.StartHeadless(unison.HeadlessConfig{Width: width, Height: height},
		unison.StartupFinishedCallback(func() {
			unison.DefaultTableColumnHeaderTheme.OnBackgroundInk = colors.OnHeader
			unison.DefaultMarkdownTheme.LinkHandler = HandleLink
			unison.DefaultMarkdownTheme.WorkingDirProvider = WorkingDirProvider
			unison.DefaultMarkdownTheme.AltLinkPrefixes = []string{"md:"}
			wnd, wndErr := unison.NewWindow(xos.AppName)
			if wndErr != nil {
				errs.Log(wndErr)
				xos.Exit(1)
			}
			registerWindowDragTypes(wnd)
			SetupMenuBar(wnd)
			InitWorkspace(wnd)
			OpenFiles(files)
		}),
	)
	if err != nil {
		xos.ExitWithMsg(fmt.Sprintf("unable to start the headless session: %v", err))
	}
	server.screen = screen
	xos.RunAtExit(screen.Stop)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /input", server.handleInputVocabulary)
	mux.HandleFunc("POST /input", server.handleInput)
	mux.HandleFunc("GET /inspect", server.handleInspect)
	mux.HandleFunc("GET /inspect/focus", server.handleInspectFocus)
	mux.HandleFunc("GET /screenshot", server.handleScreenshot)
	mux.HandleFunc("GET /console", server.handleConsole)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		xos.ExitWithMsg(fmt.Sprintf("unable to listen on %s: %v", addr, err))
	}
	slog.Info("headless debug API listening", "addr", listener.Addr().String())
	httpServer := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	// The workspace's own settings-saving close handler (workspaceWillClose, in workspace.go) only ever runs when the
	// main window is closed the normal way, and nothing here ever closes it -- the process just runs until killed.
	// Without this, everything settings-backed (page reference mappings, recent files, window layout, and so on) that
	// a session here changes is silently lost the moment the container stops, however it stops. A periodic autosave
	// covers an unclean stop (SIGKILL, OOM, a crash); the signal handler below covers a clean one (SIGINT/SIGTERM,
	// which is what "podman stop"/"docker stop" send, and Ctrl-C) by saving once more and only then letting the
	// process exit, so the common case loses nothing at all.
	stopAutosave := make(chan struct{})
	go autosaveSettings(screen, stopAutosave)
	xos.RunAtExit(func() {
		close(stopAutosave)
		saveSettings(screen)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shutdownErr := httpServer.Shutdown(shutdownCtx); shutdownErr != nil {
			errs.Log(shutdownErr)
		}
	})

	if err = httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errs.Log(err)
		xos.Exit(1)
	}
	xos.Exit(0)
}

// autosaveSettings saves the global settings every minute until stop is closed, as a safety net for a container stop
// that never reaches the signal handler in StartHeadlessAPI (SIGKILL, an OOM kill, a crash).
func autosaveSettings(screen *unison.HeadlessScreen, stop <-chan struct{}) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			saveSettings(screen)
		case <-stop:
			return
		}
	}
}

// saveSettings saves the global settings on the UI thread, the same way workspaceWillClose (workspace.go) does on a
// normal window close, logging rather than reporting through Workspace.ErrorHandler: nobody is here to see a dialog,
// and by the time this runs from the shutdown signal handler, the caller wants the process to exit either way.
func saveSettings(screen *unison.HeadlessScreen) {
	screen.Do(func() {
		if err := gurps.GlobalSettings().Save(); err != nil {
			errs.Log(err)
		}
	})
}

type headlessAPIServer struct {
	screen   *unison.HeadlessScreen
	console  *consoleBuffer
	errMu    sync.Mutex
	errCount int
}

// drainSessionErrors copies whatever HeadlessScreen.Errors() has recorded since the last drain -- recovered panics and
// requests the session itself refused -- into the console buffer, tagged distinctly from ordinary log output.
func (s *headlessAPIServer) drainSessionErrors() {
	s.errMu.Lock()
	defer s.errMu.Unlock()
	all := s.screen.Errors()
	for _, e := range all[s.errCount:] {
		s.console.add("session", "error", e.Error())
	}
	s.errCount = len(all)
}

// --- /input ---------------------------------------------------------------

type inputRequest struct {
	Op     string   `json:"op"`
	X      float32  `json:"x"`
	Y      float32  `json:"y"`
	ToX    float32  `json:"toX"`
	ToY    float32  `json:"toY"`
	DeltaX float32  `json:"deltaX"`
	DeltaY float32  `json:"deltaY"`
	Steps  int      `json:"steps"`
	Button string   `json:"button"`
	Mods   []string `json:"mods"`
	Text   string   `json:"text"`
	Key    string   `json:"key"`
	Code   int      `json:"code"`
}

func (s *headlessAPIServer) handleInput(w http.ResponseWriter, r *http.Request) {
	var req inputRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}
	mods := parseMods(req.Mods)
	switch req.Op {
	case "click":
		s.screen.ClickWith(geom.NewPoint(req.X, req.Y), parseButton(req.Button), mods)
	case "doubleclick":
		s.screen.DoubleClick(geom.NewPoint(req.X, req.Y))
	case "drag":
		s.screen.Drag(geom.NewPoint(req.X, req.Y), geom.NewPoint(req.ToX, req.ToY), max(req.Steps, 1))
	case "wheel":
		s.screen.Wheel(geom.NewPoint(req.X, req.Y), geom.NewPoint(req.DeltaX, req.DeltaY), mods)
	case "type":
		s.screen.Type(req.Text)
	case "key":
		code, ok := resolveKeyCode(req.Key, req.Code)
		if !ok {
			http.Error(w, "unknown key: give a recognized 'key' name or a numeric 'code'", http.StatusBadRequest)
			return
		}
		s.screen.KeyPress(code, mods)
	default:
		http.Error(w, "unknown op: "+req.Op, http.StatusBadRequest)
		return
	}
	s.drainSessionErrors()
	writeJSON(w, struct {
		OK bool `json:"ok"`
	}{OK: true})
}

func parseMods(names []string) mod.Modifiers {
	var m mod.Modifiers
	for _, n := range names {
		m |= mod.FromKey(n)
	}
	return m
}

func parseButton(name string) int {
	switch strings.ToLower(name) {
	case "right":
		return unison.ButtonRight
	case "middle":
		return unison.ButtonMiddle
	default:
		return unison.ButtonLeft
	}
}

// resolveKeyCode resolves a "key" op's target key: an explicit numeric code takes precedence, then unison.KeyCodeFromKey
func resolveKeyCode(name string, code int) (unison.KeyCode, bool) {
	if code != 0 {
		return unison.KeyCode(code), true
	}
	keyCode := unison.KeyCodeFromKey(name)
	return keyCode, keyCode != unison.KeyNone
}

// handleInputVocabulary reports the canonical name -- unison.KeyCode.Key() and mod.Modifiers.Key(), respectively --
// of every key and modifier POST /input recognizes, so a caller can discover the exact vocabulary rather than
// guessing at it. resolveKeyCode and parseMods also tolerate other forms of these names -- any casing for keys, and
// a handful of longer aliases plus "+"-joined combinations for modifiers.
func (s *headlessAPIServer) handleInputVocabulary(w http.ResponseWriter, _ *http.Request) {
	keyCodeList := unison.KeyCodeList()
	keys := make([]string, len(keyCodeList))
	for i, k := range keyCodeList {
		keys[i] = k.Key()
	}

	modList := mod.List()
	mods := make([]string, len(modList))
	for i, m := range modList {
		mods[i] = m.Key()
	}

	writeJSON(w, struct {
		Keys      []string `json:"keys"`
		Modifiers []string `json:"modifiers"`
	}{Keys: keys, Modifiers: mods})
}

// --- /inspect and /inspect/focus -------------------------------------------

type rectDTO struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	W float32 `json:"w"`
	H float32 `json:"h"`
}

func rectFromGeom(r geom.Rect) rectDTO {
	return rectDTO{X: r.X, Y: r.Y, W: r.Width, H: r.Height}
}

type panelDTO struct {
	Type    string  `json:"type"`
	Rect    rectDTO `json:"rect"`
	Tooltip string  `json:"tooltip,omitempty"`
	Enabled bool    `json:"enabled"`
	Text    string  `json:"text,omitempty"`
}

type windowDTO struct {
	Title string  `json:"title"`
	Rect  rectDTO `json:"rect"`
}

type inspectResult struct {
	Window    windowDTO  `json:"window"`
	Panel     *panelDTO  `json:"panel,omitempty"`
	Ancestors []panelDTO `json:"ancestors,omitempty"`
}

func (s *headlessAPIServer) handleInspect(w http.ResponseWriter, r *http.Request) {
	x, errX := strconv.ParseFloat(r.URL.Query().Get("x"), 32)
	y, errY := strconv.ParseFloat(r.URL.Query().Get("y"), 32)
	if errX != nil || errY != nil {
		http.Error(w, "x and y are required numbers", http.StatusBadRequest)
		return
	}
	pt := geom.NewPoint(float32(x), float32(y))
	var result inspectResult
	var ok bool
	s.screen.Do(func() {
		wnd := s.screen.WindowAt(pt)
		if wnd == nil {
			return
		}
		ok = true
		local := pt.Sub(wnd.ContentRect().Point)
		result = describeWindowAndPanel(wnd, wnd.Content().Parent().PanelAt(local))
	})
	if !ok {
		http.Error(w, "no window at that point", http.StatusNotFound)
		return
	}
	writeJSON(w, result)
}

func (s *headlessAPIServer) handleInspectFocus(w http.ResponseWriter, _ *http.Request) {
	var result inspectResult
	var ok bool
	s.screen.Do(func() {
		wnd := s.screen.FocusedWindow()
		if wnd == nil {
			return
		}
		ok = true
		result = describeWindowAndPanel(wnd, wnd.Focus())
	})
	if !ok {
		http.Error(w, "nothing is focused", http.StatusNotFound)
		return
	}
	writeJSON(w, result)
}

// describeWindowAndPanel must be called on the UI thread. It reports leaf's rect, and every ancestor's, translated by
// wnd's position on the shared virtual screen -- the same absolute space /screenshot and /input use -- rather than in
// wnd's own window-local space.
func describeWindowAndPanel(wnd *unison.Window, leaf *unison.Panel) inspectResult {
	origin := wnd.ContentRect().Point
	result := inspectResult{Window: windowDTO{Title: wnd.Title(), Rect: rectFromGeom(wnd.ContentRect())}}
	if leaf == nil {
		return result
	}
	d := describePanel(leaf, origin)
	result.Panel = &d
	for anc := leaf.Parent(); anc != nil; anc = anc.Parent() {
		result.Ancestors = append(result.Ancestors, describePanel(anc, origin))
	}
	return result
}

func describePanel(p *unison.Panel, origin geom.Point) panelDTO {
	rect := p.RectToRoot(p.ContentRect(false))
	rect.Point = rect.Point.Add(origin)
	d := panelDTO{
		Type:    fmt.Sprintf("%T", p.Self),
		Rect:    rectFromGeom(rect),
		Enabled: p.Enabled(),
	}
	if p.Tooltip != nil {
		d.Tooltip = tooltipOf(p.Tooltip)
	}
	switch v := p.Self.(type) {
	case interface{ Text() string }:
		d.Text = v.Text()
	case fmt.Stringer:
		d.Text = v.String()
	}
	return d
}

// tooltipOf returns the text of a tooltip panel built the way this app's tooltips are: one label per line.
func tooltipOf(tip *unison.Panel) string {
	var lines []string
	for _, child := range tip.Children() {
		if label, ok := child.Self.(*unison.Label); ok {
			lines = append(lines, label.String())
		}
	}
	return strings.Join(lines, "\n")
}

// --- /screenshot -------------------------------------------------------------

func (s *headlessAPIServer) handleScreenshot(w http.ResponseWriter, r *http.Request) {
	s.screen.Sync()
	img := s.screen.Capture()
	if img == nil {
		http.Error(w, "the screen has never been drawn", http.StatusInternalServerError)
		return
	}
	q := r.URL.Query()
	if q.Has("x") || q.Has("y") || q.Has("w") || q.Has("h") {
		x, errX := strconv.ParseFloat(q.Get("x"), 32)
		y, errY := strconv.ParseFloat(q.Get("y"), 32)
		width, errW := strconv.ParseFloat(q.Get("w"), 32)
		height, errH := strconv.ParseFloat(q.Get("h"), 32)
		if errX != nil || errY != nil || errW != nil || errH != nil {
			http.Error(w, "x, y, w and h must all be given together, as numbers", http.StatusBadRequest)
			return
		}
		// The query rect is in the same logical, screen-absolute space as every other coordinate in the protocol;
		// Capture() is in device pixels, so it is scaled up to match.
		scale := float64(s.screen.Scale())
		rect := image.Rect(
			int(scale*x), int(scale*y),
			int(scale*(x+width)), int(scale*(y+height)),
		).Intersect(img.Bounds())
		if rect.Empty() {
			http.Error(w, "the requested rectangle does not intersect the screen", http.StatusBadRequest)
			return
		}
		writeScreenshotPNG(w, img.SubImage(rect))
		return
	}
	writeScreenshotPNG(w, img)
}

func writeScreenshotPNG(w http.ResponseWriter, img image.Image) {
	w.Header().Set("Content-Type", "image/png")
	if err := png.Encode(w, img); err != nil {
		errs.Log(err)
	}
}

// --- /console ----------------------------------------------------------------

type consoleEntry struct {
	Seq     int64  `json:"seq"`
	Time    string `json:"time"`
	Source  string `json:"source"` // "log" for slog/errs.Log output, "session" for HeadlessScreen.Errors()
	Level   string `json:"level,omitempty"`
	Message string `json:"message"`
}

// consoleBuffer is a bounded, timestamped log of everything StartHeadlessAPI wants an agent driving the app to be
// able to see without a real console attached: slog output (including errs.Log, wherever it lands) and whatever
// HeadlessScreen.Errors() has recorded. It deliberately does not intercept Workspace.ErrorHandler: an app-level error
// after startup still pops the same modal dialog a real user would see, findable through /inspect/focus and
// /screenshot like any other UI state, rather than being diverted here.
type consoleBuffer struct {
	mu      sync.Mutex
	entries []consoleEntry
	nextSeq int64
}

const consoleBufferLimit = 5000

func (b *consoleBuffer) add(source, level, message string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nextSeq++
	b.entries = append(b.entries, consoleEntry{
		Seq:     b.nextSeq,
		Time:    time.Now().UTC().Format(time.RFC3339Nano),
		Source:  source,
		Level:   level,
		Message: message,
	})
	if len(b.entries) > consoleBufferLimit {
		b.entries = b.entries[len(b.entries)-consoleBufferLimit:]
	}
}

func (b *consoleBuffer) since(seq int64) []consoleEntry {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []consoleEntry
	for _, e := range b.entries {
		if e.Seq > seq {
			out = append(out, e)
		}
	}
	return out
}

func (s *headlessAPIServer) handleConsole(w http.ResponseWriter, r *http.Request) {
	s.drainSessionErrors()
	since := int64(0)
	if v := r.URL.Query().Get("since"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			http.Error(w, "since must be an integer", http.StatusBadRequest)
			return
		}
		since = n
	}
	writeJSON(w, struct {
		Entries []consoleEntry `json:"entries"`
	}{Entries: s.console.since(since)})
}

// teeLogHandler wraps whatever slog.Handler main.go's xslog.Config installed -- normally a rotating log file, plus
// stdout only when -console was passed -- so that everything logged through it, including errs.Log, also lands in the
// console buffer regardless of those flags. It changes nothing about where the original handler sends its own output.
type teeLogHandler struct {
	orig slog.Handler
	buf  *consoleBuffer
}

func (h *teeLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.orig.Enabled(ctx, level)
}

func (h *teeLogHandler) Handle(ctx context.Context, record slog.Record) error { //nolint:gocritic // signature fixed by slog.Handler
	msg := record.Message
	record.Attrs(func(a slog.Attr) bool {
		msg += " " + a.String()
		return true
	})
	h.buf.add("log", record.Level.String(), msg)
	return h.orig.Handle(ctx, record)
}

func (h *teeLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &teeLogHandler{orig: h.orig.WithAttrs(attrs), buf: h.buf}
}

func (h *teeLogHandler) WithGroup(name string) slog.Handler {
	return &teeLogHandler{orig: h.orig.WithGroup(name), buf: h.buf}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		errs.Log(err)
	}
}
