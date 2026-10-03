# Changes since the last release

## New & Improved

- (none yet)

## Bug Fixes

- On macOS, opening a file from the Finder while a copy of GCS other than the one the Finder launches was already
  running did nothing. The file is now opened by the running copy.
- A skill with no points spent on it no longer satisfies a skill prerequisite, even when it has a level from one of
  its defaults. Such skills are there to show what a character would have at default, not to stand in for the skill.
  This changes existing behavior: a skill with no points used to meet a skill prerequisite whenever its level from a
  default was high enough, and a prerequisite that doesn't ask for a level, as many in the libraries don't, was met by
  any skill with a matching name. Characters that relied on that may now show the prerequisite as unmet.
