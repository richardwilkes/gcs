# Changes since the last release

## New & Improved

- Revised the prerequisites editor to improve usability.

## Bug Fixes

- On macOS, opening a file from the Finder while a copy of GCS other than the one the Finder launches was already
  running did nothing. The file is now opened by the running copy.
- A skill with no points spent on it no longer satisfies a skill prerequisite, even when it has a level from one of
  its defaults. Such skills are there to show what a character would have at default, not to stand in for the skill.
  This changes existing behavior: a skill with no points used to meet a skill prerequisite whenever its level from a
  default was high enough, and a prerequisite that doesn't ask for a level, as many in the libraries don't, was met by
  any skill with a matching name. Characters that relied on that may now show the prerequisite as unmet.
- Adding a skill or spell to a template no longer merges its points into a matching one already there, which made the
  row vanish, even when it was added inside a choice. Points are still combined once the template is applied to a
  sheet.
- An unset substitution written as "Label: text", such as Patron's "@Who: A deity@", was shown as just "@Who@", so
  modifiers that differ only in that text, like Patron's, couldn't be told apart. The full text is shown again.
  Substitutions that list examples ending in "etc." still show just their label.
- A prerequisite group whose tech level condition doesn't match the character's no longer counts as met, which made
  an "any of" group containing one always met. Such a group is now left out of the check, and so is an empty group or
  one with nothing left in it to check. Prerequisites with nothing left to check are met.
