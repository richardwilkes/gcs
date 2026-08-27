// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package nameable

import (
	"regexp"
	"strings"
)

// legacyExampleListPattern matches an old-format marker shaped like "Label: item, item[, item...], etc." -- a
// short leading label, then two or more comma-separated example items, ending in a trailing "etc."/"etc" that
// signals the list isn't exhaustive. The item list itself is captured without the trailing ", etc." part.
var legacyExampleListPattern = regexp.MustCompile(`^([^:\n]{1,40}): ((?:[^,\n]+, )+[^,\n]+), ?etc\.?$`)

// legacyLabeledPattern matches an old-format marker shaped like "Label: anything" -- just a short leading label
// before the first colon, with no other requirement placed on what follows.
var legacyLabeledPattern = regexp.MustCompile(`^([^:\n]{1,40}): (.+)$`)

// ParseMarker parses raw -- the full text between a nameable's two '@' delimiters -- into a Marker for display
// purposes. Unlike ParseSyntax, it always succeeds.
//
// If raw uses the pipe-delimited syntax, ParseSyntax's result is used directly. Otherwise (an old-format,
// no-pipe marker), a fallback cascade applies, purely for rendering -- raw's exact text remains what's used as the
// replacements-map key (see Extract/Apply) and as DefaultValue's seed regardless of which tier below matched;
// DefaultValue deliberately does not use this cascade, so an old-format marker's seeded value stays its own raw
// text exactly as it always has. Only the widget shown for editing it, and its compact, unresolved display (see
// Apply), are affected by this cascade:
//
//  1. If raw looks like "Label: item, item[, item...], etc." (two or more comma-separated items, ending in
//     "etc."/"etc"), it's parsed into that Label plus those Options, with FreeForm set -- the trailing "etc."
//     signals the list isn't exhaustive, so typing something else is also allowed. The Tooltip is set too, to the
//     full, unsplit remainder (everything but the label) -- a safety net in case the Options split missed
//     something, so the original phrasing is never lost, just also offered as clickable choices.
//  2. Else, if raw looks like "Label: anything", it's parsed into that Label plus a Tooltip carrying the
//     remainder, with FreeForm set but no concrete Options -- there's no safe way to treat the remainder as a
//     list of choices, but it still reads better as a hint than crammed into the visible label.
//  3. Else, raw is used verbatim as the Label, with FreeForm set and no Options or Tooltip -- today's behavior,
//     unchanged, for anything the first two tiers can't make sense of.
//
// All three fallback tiers also set AllowEmpty, matching old-format markers' traditional freedom to be left as
// unfilled.
func ParseMarker(raw string) Marker {
	if marker, err := ParseSyntax(raw); err == nil {
		return marker
	}
	if m := legacyExampleListPattern.FindStringSubmatch(raw); m != nil {
		var options []string
		for item := range strings.SplitSeq(m[2], ", ") {
			if item = strings.TrimSpace(item); item != "" {
				options = append(options, item)
			}
		}
		if len(options) >= 2 {
			tooltip := strings.TrimPrefix(raw, m[1]+": ")
			return Marker{Raw: raw, Label: m[1], Tooltip: tooltip, Options: options, AllowEmpty: true, FreeForm: true}
		}
	}
	if m := legacyLabeledPattern.FindStringSubmatch(raw); m != nil {
		return Marker{Raw: raw, Label: m[1], Tooltip: m[2], AllowEmpty: true, FreeForm: true}
	}
	return Marker{Raw: raw, Label: raw, AllowEmpty: true, FreeForm: true}
}
