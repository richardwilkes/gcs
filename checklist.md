# Testing checklist: self-updating GCS

## Already covered by automated tests — don't redo these by hand

`./build.sh -a` exercises all of the following. They're listed so you can skip them, not so you can repeat them.

- Asset naming and selection for all six OS/arch combinations, including a release that's missing your platform's
  build, and the trap where a display version (`v5.46`) is used instead of the raw one (`5.46.0`).
- Download: correct SHA-256, wrong SHA-256, body shorter than advertised, body longer, HTTP 404/403/500, cancellation
  mid-stream. Every failure is asserted to leave no file behind.
- Archive extraction: correct `.tgz`/`.zip`, plus multi-entry, wrong name, `../` traversal, absolute paths, directory
  entries, symlink entries, empty files, and a decompression bomb.
- Preflight refusals: dev build, renamed executable, `/usr` `/opt` `/snap` `/nix/store` `/var/lib/flatpak` `/app`,
  the sandbox env vars, App Translocation, a bare executable on macOS, missing asset, missing digest, unwritable
  directory.
- The swap, including rollback when the second rename fails, refusal when the first fails, bundle vs. file mismatch,
  and refusal to overwrite an existing backup. Both the atomic-exchange and two-rename paths.
- The wait on the handoff port: free port, port released while waiting, port held by a different process.
- Smoke test: working build, `~`-suffixed version, a build that fails to start, wrong version, silent exit,
  a file that isn't executable at all.
- State file round-trip, corrupt/truncated/unknown-schema rejection, unknown-field tolerance.
- The startup report and sweep for every status, including repairing an installation left half-swapped.
- **macOS only:** mounting a real disk image, `ditto`ing the bundle out, unmounting — including the case where another
  image with the same `GCS v<version>` volume name is already mounted.

What's left below is what a machine can't check: real signed artifacts, the actual quit-and-relaunch, and the
platform behaviors that only appear on real hardware.

---

## Setup: getting a real "update available" without publishing anything

The update check compares the linked-in version against what's on GitHub. So **build with a version lower than the
latest published release** and GCS will offer to update itself to the real, signed, notarized one.

Latest published release is **5.45.2** (2026-08-13), so build as 5.45.0:

```sh
GCS_RELEASE=5.45.0 ./build.sh -d
```

- [ ] Confirm the built app reports 5.45.0 (`./gcs -v`, or the About box).

### macOS: sign the local build first

This matters. The update refuses to install unless the downloaded bundle is signed by the **same developer** as the
copy you're running. A locally built, unsigned app has no team identifier, so the check will (correctly) refuse.
`./build.sh -d` signs as part of packaging, but if you build the bundle any other way, sign it by hand:

```sh
codesign -s "Richard Wilkes" -f --timestamp --options runtime GCS.app
codesign -dv --verbose=2 GCS.app 2>&1 | grep TeamIdentifier   # expect PDL2XXWAW8
```

Then install it:

```sh
rm -rf /Applications/GCS.app && cp -R GCS.app /Applications/
open /Applications/GCS.app
```

### The one thing this setup can't test

Release 5.45.2 predates this feature, so after the update lands, the new version has no startup report or sweep. The
`update-state.json` file and the `.GCS.app.old-*` backup will be left behind. **That's expected here, not a bug** —
see "After the first release ships" at the end for how to test that half.

---

## The main path (do this on each of macOS, Linux, Windows)

- [ ] Launch. The update dialog appears with three buttons: `Cancel`, `Download Page`, `Install & Restart`.
- [ ] Click `Cancel` — nothing happens, no files created.
- [ ] Reopen via **Help ▸ Check for GCS updates**, click `Download Page` — the browser opens the GCS web site.
- [ ] Reopen and click `Install & Restart`.
- [ ] The progress window appears, shows a **moving** determinate bar, and the label changes through
      "Downloading…" → "Unpacking…" → "Verifying…" → "Preparing to install…".
- [ ] GCS quits and the new version starts on its own.
- [ ] The About box reports **5.45.2**.
- [ ] The relaunched instance is actually running (it didn't hand off to a stale listener and exit silently) — open a
      character sheet to confirm it's a live app, not a zombie.
- [ ] On macOS: the Dock icon and app switcher show GCS properly (this is what `open` buys over spawning the binary).

### Cancel mid-download

- [ ] Start an update, click `Cancel` while the bar is moving.
- [ ] The window closes with no error dialog.
- [ ] No `.gcs-update-*` directory is left in the install directory.
- [ ] GCS keeps running normally.

### Quit refused

- [ ] Open a character sheet and make an unsaved change.
- [ ] Start an update, let it finish preparing.
- [ ] At the "save changes?" prompt, click **Cancel**.
- [ ] A dialog says the update was discarded and will be offered again.
- [ ] The `.gcs-update-*` directory is gone and GCS is still running with your unsaved work intact.

---

## Preflight refusals (the dialog should drop `Install & Restart` and explain why)

Each of these should be refused **before** anything downloads — watch that no 30 MB transfer starts.

- [ ] **Read-only install directory.** Copy the app somewhere private, make the parent unwritable, launch from there:
      ```sh
      mkdir -p ~/gcs-test && cp -R GCS.app ~/gcs-test/ && chmod 500 ~/gcs-test
      open ~/gcs-test/GCS.app          # macOS
      ```
      Expect the "can't write to the folder it's installed in" message. Then `chmod 755 ~/gcs-test` to clean up.
- [ ] **Homebrew.** Safe to fake without installing the cask:
      ```sh
      sudo mkdir -p /opt/homebrew/Caskroom/gcs
      ```
      With the app in `/Applications`, expect the `brew upgrade --cask gcs` message. Then
      `sudo rmdir /opt/homebrew/Caskroom/gcs`. Confirm a copy **outside** `/Applications` is still offered the
      update while that directory exists — that's the distinction between "the cask is installed" and "this copy is
      the one it manages".
- [ ] **Renamed executable.** Easiest on Linux or Windows — just rename the file to `gcs2` and run it. (Don't bother
      on macOS: renaming inside the bundle breaks the code signature, so the app won't launch at all and you'd be
      testing the wrong thing.)
- [ ] **macOS translocation.** A locally built disk image carries no quarantine attribute, so it won't translocate on
      its own. Force it:
      ```sh
      xattr -w com.apple.quarantine "0083;00000000;Safari;" gcs-5.45.0-macos-arm64.dmg
      open gcs-5.45.0-macos-arm64.dmg
      ```
      Then launch GCS from the mounted volume without dragging it out. Expect the "running from a temporary copy"
      message. Confirm the path really is translocated — the About box or `ps` should show
      `/private/var/folders/…/AppTranslocation/…`.
- [ ] **Linux package-managed.** Copy the binary to `/opt/gcs/gcs` and run it from there. Expect the "installed by a
      package manager" message.
- [ ] **Dev build.** Run a build with no `GCS_RELEASE` at all. The check should say development versions don't look
      for updates, and never reach the dialog.

---

## Platform-specific

### macOS

- [ ] **Team ID mismatch is refused.** Re-sign your installed copy with an ad-hoc signature
      (`codesign -s - -f GCS.app`) and try to update. It must refuse with "signed by a different developer" — the
      downloaded release is signed by `PDL2XXWAW8` and your copy no longer is.
- [ ] **The Gatekeeper assessment is advisory, not a gate.** Timing a network cut mid-update isn't practical, so
      check it after the fact instead: after a successful update, grep the log for
      `did not pass a Gatekeeper assessment`. If that warning appears *and* the update still went through, the
      advisory behavior is confirmed. If it never appears, run `spctl --assess --type execute -vv /Applications/GCS.app`
      by hand and confirm what it says — the notarization ticket is stapled to the DMG rather than to the app, so a
      bundle copied out of one needs an online check and can legitimately fail it.
- [ ] **TCC permissions survive.** Before updating, grant GCS access to something protected (e.g. put a library in
      `~/Documents` and let it prompt). After the update, confirm it still has that access and doesn't re-prompt.
- [ ] **Launch Services isn't confused.** After the update, double-click a `.gcs` file in the Finder and confirm it
      opens in the new version, not a stale registration.
- [ ] **No leftover mounts.** `hdiutil info | grep image-path` — no GCS disk image should still be attached.
      `ls /Volumes` shows no `GCS v…` volume.
- [ ] **Rosetta.** On Apple silicon, install the **amd64** build and update. It should fetch the **arm64** release,
      not amd64 — otherwise a translated user stays translated forever.

### Windows

- [ ] The whole main path, from a directory you can write to (not `Program Files`).
- [ ] **`Program Files` is refused** cleanly before downloading, rather than failing at the swap.
- [ ] **The leftover is cleaned up.** After the update, a `.gcs.exe.old-*` file may briefly exist. Quit and relaunch;
      it should be gone.
- [ ] **Defender doesn't object.** Watch for any warning about GCS writing and launching an executable. The binaries
      are unsigned, so this is the most likely place for trouble.
- [ ] No console window flashes during the update or relaunch.
- [ ] File associations still work after the update (the registry entries point at the exe path, which is preserved).

### Linux

- [ ] The whole main path from `~/bin` or similar.
- [ ] **Desktop integration survives.** After the update, the GCS entry in your application menu still works and the
      icon is intact (`~/.local/share/applications/com.trollworks.gcs.desktop` should still point at the same path).
- [ ] **Sandboxed runtimes are refused.** If you have any of Snap/Flatpak handy, confirm the refusal; otherwise fake
      it with `SNAP=1 ./gcs`.
- [ ] Update from a directory on a **different filesystem** than `/tmp` (e.g. a separate mount), to confirm the
      staging-beside-the-target design holds.

---

## Interrupted updates

These are the ones that decide whether a bad day costs the user their application.

- [ ] **Kill the helper mid-swap.** Start an update, and as GCS quits, kill the helper process
      (`pkill -9 -f finish-update`) as fast as you can. Then launch GCS by hand. Either the old or the new version
      must start — never "no application found". Check the log for "restored the previous version".
- [ ] **Pull the network** partway through the download. Expect an error dialog, no staging directory left, and GCS
      still running.
- [ ] **Abandoned staging is cleared.** Manually create a stale directory and confirm the next update removes it:
      ```sh
      mkdir /Applications/.gcs-update-stale && touch -t 202001010000 /Applications/.gcs-update-stale
      ```
      Start an update; it should be gone. A *fresh* one (no `touch`) must be left alone.
- [ ] **Two updates at once.** Click `Install & Restart`, then try to trigger another update from the Help menu while
      the first is running. The second must do nothing.

---

## Where to look when something goes wrong

| What | macOS | Linux | Windows |
| --- | --- | --- | --- |
| Log | `~/Library/Logs/com.trollworks.gcs/gcs.log` | `~/.local/share/com.trollworks.gcs/Logs/` | `%LOCALAPPDATA%\com.trollworks.gcs\Logs\` |
| Update state | `~/Library/Application Support/com.trollworks.gcs/update-state.json` | `~/.local/share/com.trollworks.gcs/update-state.json` | `%LOCALAPPDATA%\com.trollworks.gcs\update-state.json` |
| Helper log | `<install dir>/.gcs-update-*/update.log` | same | same |

The helper has no user interface, so the log is the only place its failures appear. `update-state.json` tells you
exactly how far an update got — `status` will be `staged`, `applied` or `failed`, with a `reason`.

You can also drive the helper by hand:

```sh
/Applications/GCS.app/Contents/MacOS/gcs --finish-update /path/to/update-state.json
```

---

## Testing the startup report and sweep (without waiting for a release)

Because 5.45.2 has no updater in it, the "what happened last time" half of the feature can't be reached by the setup
above. Test it directly by writing a state file and launching GCS.

- [ ] **Failure is reported.** With GCS not running, write to the update-state path:
      ```json
      {"schema":1,"status":"failed","reason":"swap-failed","target":"/Applications/GCS.app","to_version":"5.46.0"}
      ```
      Launch GCS. Expect a warning dialog saying it wasn't updated, and the state file to be gone afterwards.
- [ ] **The backup is cleaned up on success.** Create a dummy backup and point a state file at it:
      ```json
      {"schema":1,"status":"applied","target":"/Applications/GCS.app",
       "backup":"/Applications/.GCS.app.old-test","to_version":"<the version you're running>"}
      ```
      Launch GCS. No dialog should appear, and `.GCS.app.old-test` should be removed.
- [ ] **A version mismatch keeps the backup.** Same as above but with `to_version` set to something else. Expect a
      warning, and the backup left in place.
- [ ] **A corrupt state file is survivable.** Write `{"schema":1,"status":` and launch. GCS must start normally and
      discard the file.

---

## After the first release ships

Once a release containing the updater is public, redo the **main path** with both sides carrying the feature. That's
the only way to see the complete round trip:

- [ ] The backup and staging directory are removed automatically after the new version starts.
- [ ] `update-state.json` is cleaned up rather than lingering.
- [ ] A second update in a row works (nothing left over from the first interferes).
