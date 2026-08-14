# Changes since the last release

## New & Improved

- GCS can now install its own updates. The dialog that appears when a new version is available has an **Install &
  Restart** button that downloads the release built for your system, verifies it, installs it over the copy you are
  running, and starts the new version -- no visit to the web site and no dragging anything anywhere. The **Download
  Page** button is still there for anyone who would rather do it by hand.

  Nothing is touched until the download has been checked and the new version has been shown to actually start on your
  machine, so a bad download or a release that cannot run leaves your installation exactly as it was. The previous
  version is kept until the new one has started successfully. If GCS cannot install the update itself -- because it was
  installed by Homebrew or a Linux package manager, because it is running from a folder it cannot write to, or because
  it is running straight from the disk image -- the dialog says so and why, and offers the download page instead.

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
