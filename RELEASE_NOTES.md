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
