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

package com.trollworks.gcs.ui.sheet;

import com.trollworks.gcs.ui.common.CSDropPanel;
import com.trollworks.toolkit.widget.border.TKBoxedDropShadowBorder;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKOutlineHeader;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKProxyOutline;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;
import java.awt.Rectangle;
import java.util.ArrayList;

/** An outline panel. */
public class CSSingleOutlinePanel extends CSDropPanel implements LayoutManager2 {
	private TKOutlineHeader	mHeader;
	private TKOutline		mOutline;

	/**
	 * Creates a new outline panel.
	 * 
	 * @param outline The outline to display.
	 * @param title The localized title for the panel.
	 * @param useProxy <code>true</code> if a proxy of the outline should be used.
	 */
	public CSSingleOutlinePanel(TKOutline outline, String title, boolean useProxy) {
		super(null, title, false);
		mOutline = useProxy ? new TKProxyOutline(outline) : outline;
		mHeader = mOutline.getHeaderPanel();
		CSSheet.prepOutline(mOutline);
		add(mHeader);
		add(mOutline);
		setBorder(new TKBoxedDropShadowBorder());
		setLayout(this);
	}

	/**
	 * Sets the embedded outline's display range.
	 * 
	 * @param first The first row to display.
	 * @param last The last row to display.
	 */
	public void setOutlineRowRange(int first, int last) {
		mOutline.setFirstRowToDisplay(first);
		mOutline.setLastRowToDisplay(last);
	}

	public float getLayoutAlignmentX(Container target) {
		return CENTER_ALIGNMENT;
	}

	public float getLayoutAlignmentY(Container target) {
		return CENTER_ALIGNMENT;
	}

	public void invalidateLayout(Container target) {
		// Nothing to do...
	}

	public Dimension maximumLayoutSize(Container target) {
		return getLayoutSize(mOutline.getMaximumSize());
	}

	public void addLayoutComponent(Component comp, Object constraints) {
		// Nothing to do...
	}

	public void removeLayoutComponent(Component comp) {
		// Nothing to do...
	}

	/** @return The preferred width. */
	public int getPreferredWidth() {
		Insets insets = getInsets();
		int width = insets.left + insets.right;
		TKOutlineModel outlineModel = mOutline.getModel();
		int count = outlineModel.getColumnCount();

		if (mOutline.shouldDrawColumnDividers()) {
			width += count - 1;
		}

		for (int i = 0; i < count; i++) {
			TKColumn column = outlineModel.getColumnAtIndex(i);

			width += column.getPreferredWidth(mOutline);
		}
		return width;
	}

	public void layoutContainer(Container parent) {
		Rectangle bounds = getLocalInsetBounds();
		int width = bounds.width;
		int height = mHeader.getPreferredSize().height;
		TKOutlineModel outlineModel = mOutline.getModel();
		int count = outlineModel.getColumnCount();
		ArrayList<TKColumn> changed = new ArrayList<TKColumn>();
		int[] widths = new int[count];
		TKColumn column;

		mHeader.setBounds(bounds.x, bounds.y, width, height);
		bounds.y += height;
		bounds.height -= height;
		mOutline.setBounds(bounds.x, bounds.y, width, bounds.height);
		if (mOutline.shouldDrawColumnDividers()) {
			width -= count - 1;
		}

		for (int i = 0; i < count; i++) {
			column = outlineModel.getColumnAtIndex(i);
			widths[i] = column.getPreferredWidth(mOutline);
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

		mOutline.updateRowHeightsIfNeeded(changed);
		mOutline.revalidateView();
	}

	public void addLayoutComponent(String name, Component comp) {
		// Nothing to do...
	}

	public Dimension minimumLayoutSize(Container parent) {
		return getLayoutSize(mOutline.getMinimumSize());
	}

	public Dimension preferredLayoutSize(Container parent) {
		return getLayoutSize(mOutline.getPreferredSize());
	}

	private Dimension getLayoutSize(Dimension size) {
		Insets insets = getInsets();

		return new Dimension(1 + insets.left + insets.right + size.width, insets.top + insets.bottom + size.height + mHeader.getPreferredSize().height);
	}
}
