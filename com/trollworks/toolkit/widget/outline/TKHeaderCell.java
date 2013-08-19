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

package com.trollworks.toolkit.widget.outline;

import com.trollworks.toolkit.utility.TKAlignment;

import java.awt.Color;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;

/** Used to draw headers in a {@link TKOutline}. */
public class TKHeaderCell extends TKTextCell {
	/** The width of the sorter widget. */
	public static final int	SORTER_WIDTH	= 12;
	private int				mSortSequence;
	private boolean			mSortAscending;
	private Color			mSorterColor;
	private boolean			mAllowSort;

	/** Create a new header cell. */
	public TKHeaderCell() {
		super(TKAlignment.CENTER);
		mSortSequence = -1;
		mSortAscending = true;
		mAllowSort = true;
	}

	/** @return Whether sorting is allowed and displayed. */
	public boolean isSortAllowed() {
		return mAllowSort;
	}

	/** @param allow Whether sorting is allowed and displayed. */
	public void allowSort(boolean allow) {
		if (allow != mAllowSort) {
			mAllowSort = allow;
			if (!mAllowSort) {
				mSortSequence = -1;
			}
		}
	}

	/**
	 * @return <code>true</code> if the column should be sorted in ascending order.
	 */
	public boolean isSortAscending() {
		return mSortAscending;
	}

	/**
	 * Sets the sort criteria for this column.
	 * 
	 * @param sequence The column's sort sequence. Use <code>-1</code> if it has none.
	 * @param ascending Pass in <code>true</code> for an ascending sort.
	 */
	public void setSortCriteria(int sequence, boolean ascending) {
		if (mAllowSort) {
			mSortSequence = sequence;
			mSortAscending = ascending;
		}
	}

	/** @return The column's sort sequence, or <code>-1</code> if it has none. */
	public int getSortSequence() {
		return mSortSequence;
	}

	/**
	 * Draws the cell.
	 * 
	 * @param g2d The graphics context to use.
	 * @param bounds The bounds of the cell.
	 * @param row The row to draw.
	 * @param column The column to draw.
	 * @param selected Pass in <code>true</code> if the cell should be drawn in its selected
	 *            state.
	 * @param active Pass in <code>true</code> if the cell should be drawn in its active state.
	 */
	protected void drawCellSuper(Graphics2D g2d, Rectangle bounds, TKRow row, TKColumn column, boolean selected, boolean active) {
		super.drawCell(g2d, bounds, row, column, selected, active);
	}

	@Override public void drawCell(Graphics2D g2d, Rectangle bounds, TKRow row, TKColumn column, boolean selected, boolean active) {
		if (mAllowSort) {
			bounds.x += 2;
			bounds.width -= 4;

			if (mSortSequence != -1) {
				bounds.width -= SORTER_WIDTH;
			}

			drawCellSuper(g2d, bounds, row, column, selected, active);

			if (mSortSequence != -1) {
				int x;
				int y;
				int i;

				bounds.width += SORTER_WIDTH;
				x = bounds.x + bounds.width - 6;
				y = bounds.y + bounds.height / 2 - 1;

				g2d.setColor(getSorterColor());
				for (i = 0; i <= mSortSequence; i++) {
					g2d.fillRect(bounds.x + 1 + 3 * i, bounds.y + bounds.height - 3, 2, 2);
				}

				for (i = 0; i < 5; i++) {
					int x1 = mSortAscending ? x - i : x + i - 4;
					int x2 = mSortAscending ? x + i : x + 4 - i;

					g2d.drawLine(x1, y + i, x2, y + i);
				}
			}

			bounds.x -= 2;
			bounds.width += 4;
		} else {
			drawCellSuper(g2d, bounds, row, column, selected, active);
		}
	}

	@Override public int getPreferredWidth(TKRow row, TKColumn column) {
		int width = super.getPreferredWidth(row, column);

		if (mAllowSort) {
			width += SORTER_WIDTH + 4;
		}
		return width;
	}

	@Override public String getToolTipText(MouseEvent event, Rectangle bounds, TKRow row, TKColumn column) {
		return column.getToolTipText(event, bounds);
	}

	/** @return The color used for the sorter markings. */
	public Color getSorterColor() {
		if (mSorterColor == null) {
			mSorterColor = Color.blue.darker();
		}
		return mSorterColor;
	}

	/**
	 * Sets the color used for the sorter markings.
	 * 
	 * @param sorterColor The new sorter markings color.
	 */
	public void setSorterColor(Color sorterColor) {
		mSorterColor = sorterColor;
	}
}
