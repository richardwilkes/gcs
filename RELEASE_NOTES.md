# Changes since the last release

## New & Improved

- (none yet)

## Bug Fixes

- Syncing with library sources no longer overwrites data whose kind differs from its source's, such as a trait whose
  source is a trait container or equipment whose source is an equipment group, which used to lose data. The library
  source column marks such data, including a note converted to or from a note container, with "≠" instead. Equipment
  and a physical container still sync with each other, and the editor's Sync with Source now allows that as well.
