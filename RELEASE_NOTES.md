# Changes since v5.42.0

## ⚠️ Minimum System Requirements Have Changed

This release raises the minimum operating system requirements. Please check that your system meets these before
updating:

- **macOS:** macOS 11 (Big Sur) or newer. On Intel Macs this was previously macOS 10.15 (Catalina); Apple Silicon was
  already macOS 11.
- **Windows:** Windows 10 or newer (unchanged).
- **Linux:** glibc 2.35 or newer — this corresponds to Ubuntu 22.04, Debian 12, Fedora 36, or equivalent, and newer
  (previously glibc 2.29: Ubuntu 19.04, Debian 11, Fedora 30, or equivalent). The Linux kernel must also be 3.2 or newer
  on x86-64 and 3.7 or newer on ARM64, but glibc is the binding requirement in practice.

## New & Improved

- **Drag & drop now works between windows and other applications.** You can drag items (equipment, traits, skills,
  spells, notes, and more) from one open sheet, template, or loot window and drop them into another. Previously, drag &
  drop was limited to within a single window. You can also keep scrolling with the mouse wheel while a drag is in
  progress.
- **More ways to set a character portrait.** You can now drop an image directly onto the portrait — including image data
  dragged from another application or a web link — instead of only dropping an image file.
- **Easier reordering in the body (hit location) settings.** Dragging an item near the top or bottom edge of the body
  settings now scrolls the list automatically, so you can move things past the visible area without letting go.
- **Improved PDF viewing.** PDFs render more accurately (cleaner anti-aliased edges and colors), and the viewer is more
  stable when opening unusual or damaged PDFs. Internal links within a PDF now navigate to the correct named
  destination.
- **Dark mode on Linux.** GCS now follows the system light/dark appearance preference on Linux, as it already did on
  macOS and Windows.
- **New platform builds: Windows and Linux on ARM64.** GCS is now built for ARM64 (aarch64) on both Windows and Linux,
  in addition to the existing x86-64 builds. Note that I have not been able to test these builds due to a lack of ARM64
  hardware, so please report any issues you encounter.
- **Optional specialties for skills.** A skill's specialty is now split into a required part and an optional part,
  matching the GURPS distinction between the two. (#1038)
- **External PDF viewers now jump to the searched text.** When you open a page reference in an external PDF viewer, GCS
  passes the highlighted phrase along (via the `$TEXT` placeholder) so a viewer that supports it can navigate directly
  to the term. (#1023)
- **Metric length now converts at 25 mm per inch.** Lengths entered in metric units convert using the GURPS simplified
  value of 25 mm per inch, replacing the previous conversion of 1 m per yard. (#1032)
- **Improved hiking calculations in the Calculator.** Hiking distance now follows the clarification in HT55, the daily
  hiking duration can be chosen from a drop-down (in 4-hour increments, or a custom value) with a more typical default
  of 8 hours, and a new field computes how many days a given total distance would take with the current parameters.
  (#1035)
- **Added page reference mapping for *Loadouts: Starship Crew*.** Page references to this book now resolve correctly.
  (#1021)

## Bug Fixes

- Traits now contribute their weapons to the scripting context even when they aren't leveled. Previously only leveled
  traits exposed their weapons to scripts. (#1044)
- Fixed skill levels showing an enormous bogus value (e.g. 922337203685477) when a skill defaulted from another skill
  whose qualifier — such as a minimum Tech Level — wasn't met. (#1031)
- A skill default's Tech Level constraint is now evaluated against the Tech Level of the skill being defaulted from,
  rather than the character's Tech Level. Previously, changing the sheet's Tech Level could cause a skill to default from
  another skill whose Tech Level didn't actually satisfy the constraint. (#1040)
- Fixed a crash that could occur when an adjustment was based on the level of its owner and the owner, or any of the
  owner's ancestors, was disabled. (#1029)
- Fixed a crash that could occur after undoing a change in a standalone equipment, trait, skill, or spell list and then
  saving the file. (#1015)
