# GCS Project Guidelines

## Build and Test

- Use Go 1.27 as declared in `go.mod`; `build.sh` sets `GOEXPERIMENT=simd` for the project.
- Run `./build.sh -a` for the full local verification: generation, build, formatting, linting, tests, race tests, and packaging. For focused checks, use `./build.sh -t`, `./build.sh -r`, `./build.sh -f`, or `./build.sh -l`.
- Run a changed package's tests with `go test ./path/to/package`; the full suite is `go test ./...`.
- Exercise user-facing changes in the running application after the build. See [README.md](README.md) for contribution and runtime prerequisites.

## Architecture

- `model/` owns domain data, persistence, file formats, and library management. Keep UI concerns out of it.
- `ux/` owns windows, panels, actions, localization, and UI-thread integration; it uses Unison for UI and OS integration.
- `updater/` owns application update lifecycle and platform-specific process/swap behavior.
- `runmode/` defines non-interactive modes; `cmd/` contains generators and packaging tools. `main.go` wires flags, run modes, and the interactive Unison application.
- Platform-specific behavior is selected with Go filename suffixes such as `_darwin.go`, `_windows.go`, and `_other.go`; preserve that split when changing OS behavior.

## Conventions

- Preserve the surrounding Go style and keep changes narrowly scoped. Formatting and lint rules are enforced by `.golangci.yml` and `build.sh`.
- Add or update focused tests beside the implementation. Tests commonly use `github.com/richardwilkes/toolbox/v2/check` and `t.TempDir()`; concurrency-sensitive code must be validated with the relevant race tests.
- Treat mutable model state as potentially shared with background goroutines. Follow existing accessor and locking patterns instead of exposing unsynchronized fields.
- Use the project's JSON I/O helpers in `model/jio` for application data and respect version checks when loading persisted files.
- Do not hand-edit generated `*_gen.go` files or generated packaging/icon assets. Update generator inputs, then run `go generate ./cmd/enumgen/main.go` or the appropriate `cmd` generator and verify the generated output.
- Avoid new dependencies unless they are necessary and discussed; do not reformat unrelated files.

## Documentation and CI

- The [README](README.md) is the source of truth for prerequisites, contribution expectations, and the generative AI use policy.
- [build.sh](build.sh) documents all local build modes and their flags.
- [Build workflow](.github/workflows/build.yml) and [release workflow](.github/workflows/release.yml) define supported platforms and CI packaging behavior.
