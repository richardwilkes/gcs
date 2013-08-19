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
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets.outline;

import com.trollworks.ttk.widgets.outline.Column;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.OutlineModel;

import java.util.ArrayList;

/** Utilities for columns. */
public class ColumnUtils {
	/**
	 * Packs the columns to their preferred sizes.
	 * 
	 * @param outline The {@link Outline} to pack.
	 * @param width The width available for all columns.
	 */
	public static void pack(Outline outline, int width) {
		OutlineModel outlineModel = outline.getModel();
		int count = outlineModel.getColumnCount();
		ArrayList<Column> changed = new ArrayList<Column>();
		int[] widths = new int[count];
		Column column;
		if (outline.shouldDrawColumnDividers()) {
			width -= count - 1;
		}
		for (int i = 0; i < count; i++) {
			column = outlineModel.getColumnAtIndex(i);
			widths[i] = column.getPreferredWidth(outline);
			width -= widths[i];
		}
		if (width >= 0) {
			widths[0] += width;
		} else {
			int pos = 0;
			int[] list = new int[count];
			int[] minList = new int[count];
			for (int i = 0; i < count; i++) {
				column = outlineModel.getColumnAtIndex(i);
				if (column.getRowCell(null).participatesInDynamicRowLayout()) {
					int min = column.getPreferredHeaderWidth();

					if (min < widths[i]) {
						list[pos] = i;
						minList[pos++] = min;
					}
				}
			}
			int[] list2 = new int[count];
			int[] minList2 = new int[count];
			int pos2 = 0;
			while (width < 0 && pos > 0) {
				int amt;
				if (-width > pos) {
					amt = width / pos;
				} else {
					amt = -1;
				}
				for (int i = 0; i < pos && width < 0; i++) {
					int which = list[i];
					int minWidth = minList[i];

					widths[which] += amt;
					width -= amt;
					if (widths[which] < minWidth) {
						width -= minWidth - widths[which];
						widths[which] = minWidth;
					} else if (widths[which] > minWidth) {
						list2[pos2] = which;
						minList2[pos2++] = minWidth;
					}
				}
				int[] swap = list;
				list = list2;
				list2 = swap;
				swap = minList;
				minList = minList2;
				minList2 = swap;
				pos = pos2;
				pos2 = 0;
			}
		}
		for (int i = 0; i < count; i++) {
			column = outlineModel.getColumnAtIndex(i);
			if (widths[i] != column.getWidth()) {
				column.setWidth(widths[i]);
				changed.add(column);
			}
		}
		outline.updateRowHeightsIfNeeded(changed);
	}
}
