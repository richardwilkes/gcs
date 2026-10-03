# Changes since the last release

## New & Improved

- (none yet)

## Bug Fixes

- On macOS, opening a file from the Finder while a copy of GCS other than the one the Finder launches was already
  running did nothing. The file is now opened by the running copy.
- A skill with no level, shown as "-", no longer satisfies a skill prerequisite. A skill with no points that gets a
  level from one of its defaults still does. This changes existing behavior: a skill prerequisite that doesn't ask for
  a level, as many in the libraries don't, used to be met by any skill with a matching name, even one with no level.
  Characters that relied on that may now show the prerequisite as unmet.
