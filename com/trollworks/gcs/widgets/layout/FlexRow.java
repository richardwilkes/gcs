/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets.layout;

import java.awt.Dimension;
import java.awt.Rectangle;

/** A row within a {@link FlexLayout}. */
public class FlexRow extends FlexContainer {
	@Override protected void layoutSelf(Rectangle bounds) {
		int count = getChildCount();
		int hGap = getHorizontalGap();
		int[] gaps = new int[count > 0 ? count - 1 : 0];
		for (int i = 0; i < gaps.length; i++) {
			gaps[i] = hGap;
		}
		int width = hGap * (count > 0 ? count - 1 : 0);
		Dimension[] minSizes = getChildSizes(LayoutSize.MINIMUM);
		Dimension[] prefSizes = getChildSizes(LayoutSize.PREFERRED);
		Dimension[] maxSizes = getChildSizes(LayoutSize.MAXIMUM);
		for (int i = 0; i < count; i++) {
			width += prefSizes[i].width;
		}
		int extra = bounds.width - width;
		if (extra != 0) {
			int[] values = new int[count];
			int[] limits = new int[count];
			for (int i = 0; i < count; i++) {
				values[i] = prefSizes[i].width;
				limits[i] = extra > 0 ? maxSizes[i].width : minSizes[i].width;
			}
			extra = distribute(extra, values, limits);
			for (int i = 0; i < count; i++) {
				prefSizes[i].width = values[i];
			}
			if (extra > 0 && getFillHorizontal() && gaps.length > 0) {
				int amt = extra / gaps.length;
				for (int i = 0; i < gaps.length; i++) {
					gaps[i] += amt;
				}
				extra -= amt * gaps.length;
				for (int i = 0; i < extra; i++) {
					gaps[i]++;
				}
				extra = 0;
			}
		}
		Alignment vAlign = getVerticalAlignment();
		Rectangle[] childBounds = new Rectangle[count];
		for (int i = 0; i < count; i++) {
			childBounds[i] = new Rectangle(prefSizes[i]);
			if (getFillVertical()) {
				childBounds[i].height = Math.min(maxSizes[i].height, bounds.height);
			}
			switch (vAlign) {
				case LEFT_TOP:
					childBounds[i].y = bounds.y;
					break;
				case CENTER:
					childBounds[i].y = bounds.y + (bounds.height - prefSizes[i].height) / 2;
					break;
				case RIGHT_BOTTOM:
					childBounds[i].y = bounds.y + bounds.height - prefSizes[i].height;
					break;
			}
		}
		int x = bounds.x;
		Alignment hAlign = getHorizontalAlignment();
		if (hAlign == Alignment.CENTER) {
			x += extra / 2;
		} else if (hAlign == Alignment.RIGHT_BOTTOM) {
			x += extra;
		}
		for (int i = 0; i < count; i++) {
			childBounds[i].x = x;
			if (i < count - 1) {
				x += prefSizes[i].width;
				x += gaps[i];
			}
		}
		layoutChildren(childBounds);
	}

	@Override protected Dimension getSizeSelf(LayoutSize type) {
		Dimension[] sizes = getChildSizes(type);
		Dimension size = new Dimension(getHorizontalGap() * (sizes.length > 0 ? sizes.length - 1 : 0), 0);
		for (Dimension one : sizes) {
			size.width += one.width;
			if (one.height > size.height) {
				size.height = one.height;
			}
		}
		return size;
	}
}
