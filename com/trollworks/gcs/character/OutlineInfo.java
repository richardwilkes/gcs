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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.character;

import com.trollworks.gcs.widgets.outline.ColumnUtils;
import com.trollworks.ttk.border.BoxedDropShadowBorder;
import com.trollworks.ttk.widgets.outline.Column;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.OutlineModel;
import com.trollworks.ttk.widgets.outline.Row;

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
