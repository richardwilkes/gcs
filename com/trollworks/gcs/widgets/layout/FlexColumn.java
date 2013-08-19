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
import java.util.ArrayList;

/** A column within a {@link FlexLayout}. */
public class FlexColumn extends FlexContainer {
	/** Creates a new {@link FlexColumn}. */
	public FlexColumn() {
		setFill(true);
	}

	@Override protected void layoutSelf(Rectangle bounds) {
		int count = getChildCount();
		int vGap = getVerticalGap();
		int[] gaps = new int[count > 0 ? count - 1 : 0];
		for (int i = 0; i < gaps.length; i++) {
			gaps[i] = vGap;
		}
		int height = vGap * (count > 0 ? count - 1 : 0);
		Dimension[] minSizes = getChildSizes(LayoutSize.MINIMUM);
		Dimension[] prefSizes = getChildSizes(LayoutSize.PREFERRED);
		Dimension[] maxSizes = getChildSizes(LayoutSize.MAXIMUM);
		for (int i = 0; i < count; i++) {
			height += prefSizes[i].height;
		}
		int extra = bounds.height - height;
		if (extra != 0) {
			int[] values = new int[count];
			int[] limits = new int[count];
			for (int i = 0; i < count; i++) {
				values[i] = prefSizes[i].height;
				limits[i] = extra > 0 ? maxSizes[i].height : minSizes[i].height;
			}
			extra = distribute(extra, values, limits);
			for (int i = 0; i < count; i++) {
				prefSizes[i].height = values[i];
			}
			if (extra > 0 && getFillVertical() && gaps.length > 0) {
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
		ArrayList<FlexCell> children = getChildren();
		Rectangle[] childBounds = new Rectangle[count];
		for (int i = 0; i < count; i++) {
			childBounds[i] = new Rectangle(prefSizes[i]);
			if (getFillHorizontal()) {
				childBounds[i].width = Math.min(maxSizes[i].width, bounds.width);
			}
			switch (children.get(i).getHorizontalAlignment()) {
				case LEFT_TOP:
					childBounds[i].x = bounds.x;
					break;
				case CENTER:
					childBounds[i].x = bounds.x + (bounds.width - childBounds[i].width) / 2;
					break;
				case RIGHT_BOTTOM:
					childBounds[i].x = bounds.x + bounds.width - childBounds[i].width;
					break;
			}
		}
		int y = bounds.y;
		Alignment vAlign = getVerticalAlignment();
		if (vAlign == Alignment.CENTER) {
			y += extra / 2;
		} else if (vAlign == Alignment.RIGHT_BOTTOM) {
			y += extra;
		}
		for (int i = 0; i < count; i++) {
			childBounds[i].y = y;
			if (i < count - 1) {
				y += prefSizes[i].height;
				y += gaps[i];
			}
		}
		layoutChildren(childBounds);
	}

	@Override protected Dimension getSizeSelf(LayoutSize type) {
		Dimension[] sizes = getChildSizes(type);
		Dimension size = new Dimension(0, getVerticalGap() * (sizes.length > 0 ? sizes.length - 1 : 0));
		for (Dimension one : sizes) {
			size.height += one.height;
			if (one.width > size.width) {
				size.width = one.width;
			}
		}
		return size;
	}
}
