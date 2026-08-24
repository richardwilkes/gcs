# Changes since the last release

## New & Improved

- Text in PDF documents can now be selected by dragging and copied with Edit ▸ Copy. Hold Option/Alt while dragging to
  pan the page, as a plain drag used to do. You may also right-click-drag to pan the page.
- Search matches in PDF documents are now marked with an underline in the warning color, so that they can be told
  apart from selected text. Searching no longer re-renders the pages, either: the matches are found in the page's text
  and update as you type, so the pages stay on screen instead of being redrawn with every keystroke.

## Bug Fixes

- Opening a PDF, for example by clicking a page reference, no longer crashes on older computers whose processors lack
  AVX2 support, such as Intel 2nd/3rd-generation Core (Sandy Bridge, Ivy Bridge) and AMD FX / A-series CPUs.
