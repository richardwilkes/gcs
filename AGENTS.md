# GCS Agent Instructions

GCS is a desktop app for GURPS character sheets, templates and libraries, built on the `unison` UI toolkit.

## Repository layout

- `model/gurps` holds the data model, `model/jio` the JSON load and save, and the other `model` packages shared
  support such as criteria and fixed-point math.
- `ux` is the user interface: editors, tables and the sheet and template dockables. `ux/uxtest` drives it headlessly
  for tests.
- `cmd` holds the code generators, such as `cmd/enumgen`.

## Related repositories

GCS depends on these repositories, all maintained alongside it:

- [`unison`](https://github.com/richardwilkes/unison): the cross-platform UI toolkit, including the headless screen
  the `ux` tests run on.
- [`toolbox`](https://github.com/richardwilkes/toolbox): general utilities used throughout, such as `check`, `errs`,
  `i18n` and the ID scheme.
- [`canvas`](https://github.com/richardwilkes/canvas): the pure-Go 2D graphics library unison draws with.
- [`pdfview`](https://github.com/richardwilkes/pdfview): the pure-Go PDF engine behind the PDF viewer.
- [`rpgtools`](https://github.com/richardwilkes/rpgtools): calendar, dice and name generation.
- [`gcs_master_library`](https://github.com/richardwilkes/gcs_master_library): the content library GCS downloads at
  run time (traits, skills, equipment, templates and more). Use it as real-world data when testing a model change.

Run `go list -m -f '{{.Dir}}' <module>` to find a dependency's source in the module cache. When a bug or a missing
feature belongs in one of these repositories, fix it there rather than working around it in GCS.

## Data model

- `Node[T]` in `model/gurps/node.go` is a constraint, not an interface type. Its type union lists every node type,
  and a new node type must be added there. Generic code takes a `T Node[T]` type parameter and works with `T`
  directly. `Node[T]` can't be used as a variable, field or slice element type.
- Traits, skills, spells, equipment, their modifiers and notes split their fields in two. `XxxData` is what goes to
  disk: it embeds `SourcedID` and `XxxEditData`. `XxxEditData` holds the fields the editor changes. An editor stages
  edits by calling `CopyFrom` on an `XxxEditData` and writes them back with `ApplyTo`.
- `Clone` takes a `CloneMode` that decides what happens to the clone's ID and Source. `Reference` is for copying
  into another list, `Duplicate` for a sibling in the same list, and `Copy` for an in-memory working copy. Choose the
  mode deliberately, since using the wrong one is a common source of bugs.

## Formatting

- Wrap comments at a 120-character column limit, not a shorter width.
- Indent continuation lines inside a comment with spaces placed after the comment marker. The marker itself keeps
  whatever tab indentation the formatter gives it.
- Format Go code with `golangci-lint fmt`, which runs gofumpt with this repo's extra rules, never with `gofmt`.
  `./build.sh -f` checks the formatting without changing anything.

## Comments

Any comments you add or edit:

- Should be checked for correctness.
- Should be as concise as possible.
- Exported identifiers always get a doc comment; otherwise omit comments when the code is obvious.

## Source conventions

- Make changes in place. Don't move or reorganize code a change doesn't require, and don't add helpers or API that
  nothing uses.
- Every `.go` file carries the MPL 2.0 copyright header found at the top of any existing source file (generated files
  put their "Code generated" line ahead of it), with a blank line after it so that a package comment does not merge
  into it. The `goheader` linter checks the exact text, including a year range that ends in the current year, and a
  root test backstops it for the files lint never loads (generated files and those behind build tags), so copy the
  header into every new file.
- The `*_gen.go` files are generated from the `allEnums` table in `cmd/enumgen/main.go`. They are deleted and
  rewritten by every regeneration, so never edit them by hand: change the table and run `./build.sh -G`.
- An enum belongs with the code that owns it: model enums go under `model/gurps/enums/<pkg>/`, and an enum used only
  by the UI goes under `ux`. List it in `allEnums` rather than writing the type by hand. When an enum needs a
  "no value" member, make it the first one, so the zero value means unset.
- Keep each user-facing sentence in a single `i18n.Text` literal, even when other strings repeat part of it.
  Translators need whole sentences, so never assemble one from translated fragments.
- Methods on generic types can declare their own type parameters, so prefer methods to free functions unless an
  interface has to express the behavior.

## Tests

- A new test that puts two or more goroutines over shared state must also be listed in its package's `TestRace`
  wrapper (see the `race_coverage_test.go` files), or the race pass never runs it.
- Tests that start a headless workspace with `uxtest.StartHeadlessWorkspace` must not call `t.Parallel`: a session
  owns most of unison's mutable globals while it runs.
- In a headless UI test, do all UI work inside `HeadlessScreen.Do`, which runs it on the UI thread and waits for it to
  settle. Use `Post` for anything that runs a nested event loop, such as a modal dialog.
- `check.Equal` and `check.NotEqual` take the expected value first.
- Assert against values the test states independently, never against output produced by the code under test.

## Building, testing and linting

Prefer `./build.sh` over calling `go` directly. It exports `GOEXPERIMENT=simd` to match CI and, on macOS, the link
flags that keep the binary's deployment target at 11.0. `./build.sh -h` lists every option; the ones that matter day to
day are:

- `./build.sh` with no options regenerates the enum sources and builds the application without packaging it. `-G` only
  regenerates. `-g` builds without regenerating and then runs the packager, which on macOS wraps the binary in
  `GCS.app`.
- `./build.sh -t` runs `go test ./...`. `-r` does the same and then race-checks the `TestRace` wrappers, which cover the
  tests that put several goroutines over shared state.
- `./build.sh -f` checks the formatting, printing the diff of anything `golangci-lint fmt` would change.
- `./build.sh -l` runs `golangci-lint run` for this machine's GOOS only. CI lints darwin, linux and windows separately
  because `ux` has per-platform sources, so before calling the lint clean also run it for the other two, for example
  `GOEXPERIMENT=simd GOOS=linux golangci-lint run`. Run only one `golangci-lint` process at a time.
- `./build.sh -a` is the closest local equivalent of CI: generate, build, format check, lint for this GOOS only, tests
  with the race pass, and packaging. Follow it with the lint runs for the other two GOOSes described above.
- `./build.sh -H` builds and packages like `-g`, but with the headless debug API compiled in. The binary still runs the
  normal UI unless started with `-headless-api <address>`, which instead serves an HTTP API that drives and inspects
  the full UI without a display. `ux/headless_api.md` documents it.

Every `-f` or `-l` run asks GitHub for the latest golangci-lint release and installs it into `$(go env GOPATH)/bin`
when the local copy differs, so those runs need network access and may upgrade the tool. Set `GOLANGCI_LINT` to the
path of a binary to use it as is. Running `go test` directly on a single package, such as `go test ./model/gurps/...`,
is fine for a quick check. The tests under `ux` drive the UI headlessly and need no display.

## Pull requests

- Write the PR description from `.github/pull_request_template.md`. Fill in the sections that apply, delete the ones
  that don't, and remove the template's comments.
- Include the Release Notes section whenever a change is visible to users, since the release process collects the
  notes from it. Write each bullet for users, describing what changed rather than how it was done. Leave the section
  out for strictly internal changes, such as refactors, tests and tooling.
- Don't hard-wrap the description. GitHub renders single newlines as line breaks.
