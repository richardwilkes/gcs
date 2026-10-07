# GCS Agent Instructions

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

- Every `.go` file carries the MPL 2.0 copyright header found at the top of any existing source file (generated files
  put their "Code generated" line ahead of it). A root test fails when a file lacks it, so copy the header into every
  new file.
- The `*_gen.go` files are generated from the `allEnums` table in `cmd/enumgen/main.go`. They are deleted and
  rewritten by every regeneration, so never edit them by hand: change the table and run `./build.sh -G`.
- A new test that puts two or more goroutines over shared state must also be listed in its package's `TestRace`
  wrapper (see the `race_coverage_test.go` files), or the race pass never runs it.
- Tests that start a headless workspace with `startHeadlessWorkspace` must not call `t.Parallel`: a session owns most
  of unison's mutable globals while it runs.

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
