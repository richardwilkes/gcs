# GCS headless debug API

This is a headless GCS session: no real window is visible anywhere, but the full UI is running and can be driven and
inspected over this HTTP API. It only exists in builds compiled with the `headlessapi` build tag, via the hidden
`-headless-api` command line flag.

Coordinates everywhere in this API -- input points, inspected rects, screenshot rects -- are in the same logical,
screen-absolute space: the one `/inspect` and `/inspect/focus` report a widget's rect in is exactly the one a
screenshot rect or an input point should be given in. Windows other than the main one (such as a modal error dialog)
are positioned within this same space, not at their own private origin. The virtual screen's size is fixed for the
life of the session; there is no live resize.

## GET /

This document. Content-negotiated via the `Accept` header: Markdown (`text/markdown`) by default, or this API's
full [OpenAPI](https://www.openapis.org/) specification as YAML (`application/yaml` or `application/x-yaml`) or JSON
(`application/json`) if that's what `Accept` asks for.

## GET /input

The vocabulary `POST /input`'s `key` and `mods` fields recognize, straight from unison's own names (so this list
tracks whatever GCS is built against, rather than being maintained by hand):

```json
{"keys": ["A", "B", ..., "return", "escape", ...], "modifiers": ["ctrl", "alt", "shift", "caps", "num", "cmd"]}
```

`key` also tolerates other casings of these names (`"a"` and `"A"` both work; `"f1"` and `"F1"` both work). A `mods`
entry may also be a `"+"`-joined combination of these names, e.g. `"ctrl+shift"`.

## POST /input

Injects one input event. JSON body:

```json
{ "op": "click", "x": 0, "y": 0 }
```

Fields, all optional except `op`:

| Field              | Type        | Used by                                                                                                                     |
| ------------------ | ----------- | --------------------------------------------------------------------------------------------------------------------------- |
| `op`               | string      | always -- one of the ops below                                                                                              |
| `x`, `y`           | number      | `click`, `doubleclick`, `drag` (from), `wheel`                                                                              |
| `toX`, `toY`       | number      | `drag` (to)                                                                                                                 |
| `deltaX`, `deltaY` | number      | `wheel`                                                                                                                     |
| `steps`            | int         | `drag` -- number of intermediate move steps (default 1)                                                                     |
| `button`           | string      | `click` -- `"left"` (default), `"right"`, or `"middle"`                                                                     |
| `mods`             | string[]    | `click`, `wheel`, `key` -- see `GET /input` for the recognized names; entries may also be `"+"`-joined, e.g. `"ctrl+shift"` |
| `text`             | string      | `type` -- typed one rune at a time, the way a person at a keyboard would                                                    |
| `key`, `code`      | string, int | `key` -- see `GET /input`; `code` is a raw numeric key code and takes precedence over `key` if both are given               |

Ops:

- `click` -- press and release `button` at `(x, y)`
- `doubleclick` -- click twice at `(x, y)` with a click count of two
- `drag` -- press at `(x, y)`, move to `(toX, toY)` over `steps` steps, release
- `wheel` -- rotate the mouse wheel by `(deltaX, deltaY)` at `(x, y)`
- `type` -- type `text`
- `key` -- press and release `key` (or `code`) with `mods` held

Response: `{"ok": true}`, or a `4xx` with a plain-text error.

## GET /inspect?x=&y=

Reports the widget at a point: its type, absolute bounding rect, tooltip, enabled state and text, along with the same
for every ancestor up to the window's root. `404` if there is no window at that point.

```json
{
  "window": {"title": "GCS", "rect": {"x": 0, "y": 0, "w": 1400, "h": 900}},
  "panel": {"type": "*unison.Label", "rect": {"x": 8, "y": 28, "w": 115, "h": 17},
            "visible": {"x": 8, "y": 28, "w": 115, "h": 17}, "enabled": true, "text": "..."},
  "ancestors": [{"type": "...", "rect": {...}, "visible": {...}, "enabled": true}, ...]
}
```

`rect` is the widget's whole extent; `visible` is the part of it that is actually on screen, after every ancestor has
clipped it. Inside a scroll panel the two differ, and a widget scrolled out of sight has no `visible` at all. **Aim
input at `visible`**, not at `rect`: the center of a `rect` that is mostly clipped away lands on whatever is drawn
over that spot instead, or outside the window entirely.

## GET /inspect/focus

The same shape as `GET /inspect`, but for whatever panel currently holds keyboard focus -- in whichever window
currently has it, which may be a modal dialog (such as an error dialog a normal user would see and dismiss) rather
than the main window. `404` if no window is focused. Purely a read: a window that nothing in has been focused into
reports no `panel` rather than acquiring one.

## GET /screenshot

A PNG (`Content-Type: image/png`) of the whole virtual screen. With `x`, `y`, `w` and `h` all given, a PNG cropped to
that absolute rectangle instead.

## GET /console?since=

Log output and session errors recorded since sequence number `since` (default 0 -- everything), for polling
incrementally:

```json
{
    "entries": [
        {
            "seq": 1,
            "time": "2026-01-01T00:00:00Z",
            "source": "log",
            "level": "INFO",
            "message": "..."
        }
    ]
}
```

`source` is `"log"` for slog output (including anything logged via `errs.Log`, wherever its handler sends it) or
`"session"` for a recovered panic or a request the headless session itself refused. This does not include app-level
errors surfaced after startup through the normal error dialog (a failed file load, say) -- those still show up as a
real modal dialog, findable and dismissable through `/inspect/focus`, `/input` and `/screenshot` like any other UI
state, exactly as a user would see them.
