/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.widgets.outline.ColumnUtils;
import com.trollworks.toolkit.ui.border.BoxedDropShadowBorder;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.Row;

import java.awt.Insets;
import java.util.List;

/** Holds information about the outline relevant for page layout. */
public class OutlineInfo {
	private int		mRowIndex;
	private int[]	mHeights;
	private int		mOverheadHeight;
	private int		mMinimumHeight;

	/**
	 * Creates a new outline information holder.
	 * 
	 * @param outline The outline to collect information about.
	 * @param contentWidth The content width.
	 */
	public OutlineInfo(Outline outline, int contentWidth) {
		Insets insets = new BoxedDropShadowBorder().getBorderInsets(null);
		OutlineModel outlineModel = outline.getModel();
		int count = outlineModel.getRowCount();
		List<Column> columns = outlineModel.getColumns();
		boolean hasRowDividers = outline.shouldDrawRowDividers();

		ColumnUtils.pack(outline, contentWidth - (insets.left + insets.right));

		mRowIndex = -1;
		mHeights = new int[count];

		for (int i = 0; i < count; i++) {
			Row row = outlineModel.getRowAtIndex(i);
			mHeights[i] = row.getHeight();
			if (mHeights[i] == -1) {
				mHeights[i] = row.getPreferredHeight(columns);
			}
			if (hasRowDividers) {
				mHeights[i]++;
			}
		}

		mOverheadHeight = insets.top + insets.bottom + outline.getHeaderPanel().getPreferredSize().height;
		mMinimumHeight = mOverheadHeight + 15;
	}

	/**
	 * @param remaining The remaining vertical space on the page.
	 * @return The space the outline will consume before being complete or requiring another page.
	 */
	public int determineHeightForOutline(int remaining) {
		int total = mOverheadHeight;
		int start = mRowIndex++;

		while (mRowIndex < mHeights.length) {
			int tmp = total + mHeights[mRowIndex];
			if (tmp > remaining) {
				if (--mRowIndex == start) {
					mRowIndex++;
					return tmp;
				}
				return total;
			}
			total = tmp;
			mRowIndex++;
		}
		return total;
	}

	/** @return Whether more rows need to be placed on a page or not. */
	public boolean hasMore() {
		return mRowIndex < mHeights.length;
	}

	/** @return The minimum height. */
	public int getMinimumHeight() {
		return mMinimumHeight;
	}

	/** @return The current row index. */
	public int getRowIndex() {
		return mRowIndex;
	}
}
