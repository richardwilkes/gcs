# Changes since the last release

## New & Improved

- When GCS crashes, the details are now written to a `gcs-crash.log` file kept alongside the usual log file. Those
  details previously went only to the standard error stream, which is discarded for an application started from the
  Finder or its equivalent, and the regular log received nothing, since the process stopped before anything could be
  written to it. That left nothing to work from when a crash was reported. New reports are appended, so an earlier
  crash is still recoverable if GCS is restarted before the file is collected, and nothing is written to the file
  unless a crash actually occurs. The file is rotated on the same terms as the log, and honors the same size and
  retention settings, so it cannot grow without bound.

## Bug Fixes

- macOS: Fixed a crash that could occur when dragging, which most often showed up as GCS disappearing while working on
  a character sheet -- collapsing a container, deleting rows, or seemingly just clicking, since a click with the
  slightest movement is the start of a drag. GCS quit without any warning and left nothing behind in the log.
