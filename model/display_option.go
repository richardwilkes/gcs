/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package model

// Inline returns true if inline notes should be shown.
func (enum DisplayOption) Inline() bool {
	return enum == InlineDisplayOption || enum == InlineAndTooltipDisplayOption
}

// Tooltip returns true if tooltips should be shown.
func (enum DisplayOption) Tooltip() bool {
	return enum == TooltipDisplayOption || enum == InlineAndTooltipDisplayOption
}
