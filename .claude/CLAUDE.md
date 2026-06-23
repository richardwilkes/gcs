# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

GCS (GURPS Character Sheet) is a stand-alone, interactive desktop GUI application for building characters for the
GURPS Fourth Edition tabletop RPG. It is written in Go and uses [Unison](https://github.com/richardwilkes/unison)
(the author's own GLFW/OpenGL-based toolkit) for the UI and OS integration. Module path: `github.com/richardwilkes/gcs/v5`.

## Critical: GOEXPERIMENT=jsonv2

The code imports `encoding/json/v2` and `encoding/json/jsontext` (~66 files). **Every Go invocation must run with
`GOEXPERIMENT=jsonv2` in the environment** or it will not compile. `build.sh` exports it and `.vscode/settings.json`
sets it for the Go tooling, but any manual command must set it explicitly:

```bash
GOEXPERIMENT=jsonv2 go test ./model/gurps/ -run TestName
GOEXPERIMENT=jsonv2 go build ./...
GOEXPERIMENT=jsonv2 go vet ./...
```

## Common commands

All builds go through `./build.sh` (run `./build.sh -h` for the full list):

| Command | Purpose |
| --- | --- |
| `./build.sh` | Default: generate source, then build the Go binary |
| `./build.sh -G` | Regenerate generated source (enums) |
| `./build.sh -t` | Run all tests |
| `./build.sh -r` | Run tests with the race detector |
| `./build.sh -l` | Run golangci-lint (auto-installs the matching version into `$GOPATH/bin`) |
| `./build.sh -a` | Everything: gen + build + lint + race tests + package |
| `./build.sh -d` | Build a distribution (sets the release version) |
| `./build.sh -i` | Extract the i18n translation template |
| `./build.sh -p` | Regenerate icons and `packaging.yml` |

Run a single test directly: `GOEXPERIMENT=jsonv2 go test ./model/fxp/ -run TestWeight`.

Linting uses **golangci-lint v2** (config in `.golangci.yml`) with a broad linter set (gocritic all-tags, gosec,
revive, staticcheck, govet enable-all). Formatting is enforced by **gofumpt (extra-rules) + goimports** — plain
`gofmt` output will not pass.

## Architecture

The codebase has a strict **model / UI split**:

- **`model/`** — all domain logic and business rules, no UI dependencies.
  - **`model/gurps/`** — the heart of the application (~140 files). `Entity` (the character) is the root aggregate;
    its children are the `Node[T]` types: `Trait`, `TraitModifier`, `Skill`, `Spell`, `Equipment`,
    `EquipmentModifier`, `Note`, `Weapon`, `ConditionalModifier`. `node.go` defines the generic `Node[T NodeTypes]`
    interface these all implement — these are the tree/table rows shown in the UI. Features, prerequisites, and
    bonuses are also modeled here.
  - **`model/fxp/`** — fixed-point decimal math (`fxp.Int`, backed by toolbox `fixed64`), plus `Length` and `Weight`
    with units. **All game numbers use `fxp.Int`, never `float64`.**
  - **`model/jio/`** — JSON load/save and data versioning. `CurrentDataVersion = 5`, `MinimumDataVersion = 2`; all GCS
    data files share one version number.
  - Other support packages: `criteria`, `nameable`, `colors`, `fonts`, `paper`, `kinds`.
- **`ux/`** — the UI layer (~150 files) built on Unison. Heavily depends on `model/gurps`. The workspace is a
  dock-based UI of `*Dockable` panels (character sheets, templates, settings, PDF/markdown viewers). `ux.Start()` is
  the GUI entry point and never returns.
- **`main.go`** — entry point. Configures app identity via `early.Configure()`, then either runs a headless CLI mode
  (`--convert`, `--sync`, `--text` export) or launches the GUI through `ux.Start()`.
- **`cmd/`** — auxiliary tools, not part of the app binary: `enumgen` (enum code generation), `genpkg` (icons +
  packaging), `pack` (distribution packaging), `prereq-counts`.

## Code generation (enums)

Enums are code-generated. The canonical list of all enums lives in the `allEnums` variable in
[cmd/enumgen/main.go](cmd/enumgen/main.go). For each enum there is a hand-written file (e.g.
`model/gurps/enums/difficulty/level.go`, holding custom methods) and a generated companion
`*_gen.go` file (constants, `String()`, marshaling — marked `// Code generated ... DO NOT EDIT`). Most enums live in
their own package under `model/gurps/enums/<name>/`.

To add or change an enum: edit `allEnums` in `cmd/enumgen/main.go` (and the hand-written method file if needed), then
run `./build.sh -G`. **Regeneration deletes every `*_gen.go` file in the repo and rewrites it**, so never hand-edit a
`_gen.go` file.

## Libraries & data files

Character/library content lives in **libraries** — Git repositories (a master library, the user library, and add-on
libraries) that can be browsed and updated from within the app via go-git ([library.go](model/gurps/library.go),
[git_latest.go](model/gurps/git_latest.go)). Individual data nodes can be linked to a library "source" and synced
(`GetSource`/`SyncWithSource`, `srcstate` enum). File extensions (`.gcs` sheet, `.gct` template, `.adq` traits, `.skl`
skills, `.spl` spells, `.eqp` equipment, etc.) are defined in [file_type.go](model/gurps/file_type.go).

## Scripting

Calculated values can be driven by embedded JavaScript via [goja](https://github.com/dop251/goja). The bridge lives
in `model/gurps/scripting*.go`, which exposes the entity, attributes, skills, spells, traits, equipment, weapons, and
dice to scripts.

## Conventions

- Every Go source file begins with the MPL 2.0 copyright header (see any existing `.go` file); new files must include it.
- User-facing strings are wrapped in `i18n.Text(...)`; `./build.sh -i` extracts the translation template.
- The toolbox v2 library (`xos`, `errs`, `i18n`, `tid`, `fixed64`, …) is used pervasively for cross-platform and
  utility needs — prefer it over rolling your own.
