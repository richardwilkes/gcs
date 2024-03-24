/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/unison"
)

func headerFromData[T gurps.NodeTypes](data gurps.HeaderData, forPage bool) unison.TableColumnHeader[*Node[T]] {
	if data.TitleIsImageKey {
		var img1, img2 *unison.SVG
		switch data.Title {
		case gurps.HeaderCheckmark:
			img1 = unison.CheckmarkSVG
		case gurps.HeaderCoins:
			img1 = svg.Coins
		case gurps.HeaderWeight:
			img1 = svg.Weight
		case gurps.HeaderBookmark:
			img1 = svg.Bookmark
		case gurps.HeaderStackedCoins:
			img1 = svg.Stack
			img2 = svg.Coins
		case gurps.HeaderStackedWeight:
			img1 = svg.Stack
			img2 = svg.Weight
		}
		if img2 != nil {
			return NewEditorListSVGPairHeader[T](img1, img2, data.Detail, forPage)
		}
		if img1 != nil {
			return NewEditorListSVGHeader[T](img1, data.Detail, forPage)
		}
	}
	return NewEditorListHeader[T](data.Title, data.Detail, forPage)
}
