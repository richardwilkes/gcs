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

import java.awt.Component;
import java.awt.Dimension;
import java.awt.Rectangle;
import java.security.InvalidParameterException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;
import java.util.TreeMap;

/** A grid within a {@link FlexLayout}. */
public class FlexGrid extends FlexContainer {
	private HashMap<FlexCell, FlexGridData>				mData		= new HashMap<FlexCell, FlexGridData>();
	private int											mColumns	= 1;
	private int											mRows		= 1;
	private FlexGridData[][]							mGrid;
	private int[]										mColumnWidths;
	private int[]										mMinColumnWidths;
	private int[]										mMaxColumnWidths;
	private int[]										mRowHeights;
	private int[]										mMinRowHeights;
	private int[]										mMaxRowHeights;
	private TreeMap<Integer, ArrayList<FlexGridData>>	mRowSpanMap;
	private TreeMap<Integer, ArrayList<FlexGridData>>	mColumnSpanMap;

	/** Do not call this method. */
	@Override public void add(Component comp) {
		throw new UnsupportedOperationException();
	}

	/** Do not call this method. */
	@Override public void add(FlexCell cell) {
		throw new UnsupportedOperationException();
	}

	/**
	 * @param comp The {@link Component} to add as a child.
	 * @param row The row index within the grid to set the upper-left corner of the cell to.
	 * @param column The column index within the grid to set the upper-left corner of the cell to.
	 */
	public void add(Component comp, int row, int column) {
		add(comp, row, column, 1, 1);
	}

	/**
	 * @param comp The {@link Component} to add as a child.
	 * @param row The row index within the grid to set the upper-left corner of the cell to.
	 * @param column The column index within the grid to set the upper-left corner of the cell to.
	 * @param rowSpan The number of rows the child should span.
	 * @param columnSpan The number of columns the child should span.
	 */
	public void add(Component comp, int row, int column, int rowSpan, int columnSpan) {
		add(new FlexComponent(comp), row, column, rowSpan, columnSpan);
	}

	/**
	 * @param cell The {@link FlexCell} to add as a child.
	 * @param row The row index within the grid to set the upper-left corner of the cell to.
	 * @param column The column index within the grid to set the upper-left corner of the cell to.
	 */
	public void add(FlexCell cell, int row, int column) {
		add(cell, row, column, 1, 1);
	}

	/**
	 * @param cell The {@link FlexCell} to add as a child.
	 * @param row The row index within the grid to set the upper-left corner of the cell to.
	 * @param column The column index within the grid to set the upper-left corner of the cell to.
	 * @param rowSpan The number of rows the child should span.
	 * @param columnSpan The number of columns the child should span.
	 */
	public void add(FlexCell cell, int row, int column, int rowSpan, int columnSpan) {
		super.add(cell);
		FlexGridData data = new FlexGridData(cell, row, column, rowSpan, columnSpan);
		mData.put(cell, data);
		row = data.getLastRow() + 1;
		if (row > mRows) {
			mRows = row;
		}
		column = data.getLastColumn() + 1;
		if (column > mColumns) {
			mColumns = column;
		}
	}

	private int updateStarts(int base, int[] values, int[] starts) {
		int start = base;
		for (int i = 0; i < values.length; i++) {
			starts[i] = start;
			start += values[i];
		}
		return start - base;
	}

	private void grow(int extra, int[] values, int[] max, boolean fill) {
		int count = 1 + (values.length - 1) / 2;
		int[] tv = new int[count];
		int[] tm = new int[count];
		for (int i = 0; i < count; i++) {
			tv[i] = values[i * 2];
			tm[i] = max[i * 2];
		}
		extra = distribute(extra, tv, tm);
		for (int i = 0; i < count; i++) {
			values[i * 2] = tv[i];
		}

		if (extra > 0 && fill) {
			// None are accepting more space, so force them
			count = 1 + (values.length - 1) / 2;
			while (extra > 0) {
				int amt = extra / count;
				if (amt < 1) {
					amt = 1;
				}
				for (int i = 0; i <= values.length && extra > 0; i += 2) {
					values[i] += amt;
					max[i] += amt;
					extra -= amt;
				}
			}
		}
	}

	private void shrink(int extra, int[] values, int[] min) {
		int count = 1 + (values.length - 1) / 2;
		int[] tv = new int[count];
		int[] tm = new int[count];
		for (int i = 0; i < count; i++) {
			tv[i] = values[i * 2];
			tm[i] = min[i * 2];
		}
		extra = distribute(-extra, tv, tm);
		for (int i = 0; i < count; i++) {
			values[i * 2] = tv[i];
		}

		if (extra < 0) {
			// None are accepting less space, so force them
			count = 1 + (values.length - 1) / 2;
			while (extra > 0 && count > 0) {
				int amt = extra / count;
				if (amt > -1) {
					amt = -1;
				}
				count = 0;
				for (int i = 0; i <= values.length && extra < 0; i += 2) {
					if (values[i] > 0) {
						values[i] += amt;
						min[i] += amt;
						extra += amt;
						count++;
					}
				}
			}
		}
	}

	@Override protected void layoutSelf(Rectangle bounds) {
		createGrid(LayoutSize.PREFERRED);

		// Size widths appropriately
		int[] x = new int[mColumns];
		int width = updateStarts(bounds.x, mColumnWidths, x);
		int extra = bounds.width - width;
		if (extra != 0) {
			if (extra > 0) {
				grow(extra, mColumnWidths, mMaxColumnWidths, getFillHorizontal());
			} else {
				shrink(-extra, mColumnWidths, mMinColumnWidths);
			}
			width = updateStarts(bounds.x, mColumnWidths, x);
		}

		// Size heights appropriately
		int[] y = new int[mRows];
		int height = updateStarts(bounds.y, mRowHeights, y);
		extra = bounds.height - height;
		if (extra != 0) {
			if (extra > 0) {
				grow(extra, mRowHeights, mMaxRowHeights, getFillVertical());
			} else {
				shrink(-extra, mRowHeights, mMinRowHeights);
			}
			height = updateStarts(bounds.y, mRowHeights, y);
		}

		// Set the child bounds
		ArrayList<Rectangle> cellBounds = new ArrayList<Rectangle>(getChildCount());
		for (FlexCell cell : getChildren()) {
			FlexGridData data = mData.get(cell);
			Rectangle rect = new Rectangle();
			rect.x = x[data.mColumn];
			rect.y = y[data.mRow];
			int last = data.getLastColumn();
			width = 0;
			for (int column = data.mColumn; column <= last; column++) {
				width += mColumnWidths[column];
			}
			rect.width = width;
			last = data.getLastRow();
			height = 0;
			for (int row = data.mRow; row <= last; row++) {
				height += mRowHeights[row];
			}
			rect.height = height;
			cellBounds.add(rect);
		}
		flush();
		layoutChildren(cellBounds.toArray(new Rectangle[cellBounds.size()]));
	}

	@Override protected Dimension getSizeSelf(LayoutSize type) {
		createGrid(type);
		int width = 0;
		for (int i = 0; i < mColumns; i++) {
			width += mColumnWidths[i];
		}
		int height = 0;
		for (int i = 0; i < mRows; i++) {
			height += mRowHeights[i];
		}
		flush();
		return new Dimension(width, height);
	}

	private void flush() {
		mGrid = null;
		mColumnWidths = null;
		mMinColumnWidths = null;
		mMaxColumnWidths = null;
		mRowHeights = null;
		mMinRowHeights = null;
		mMaxRowHeights = null;
		mRowSpanMap = null;
		mColumnSpanMap = null;
	}

	private void createGrid(LayoutSize type) {
		mGrid = new FlexGridData[mRows][mColumns];
		mColumnWidths = new int[mColumns];
		mMinColumnWidths = new int[mColumns];
		mMaxColumnWidths = new int[mColumns];
		mRowHeights = new int[mRows];
		mMinRowHeights = new int[mRows];
		mMaxRowHeights = new int[mRows];
		mRowSpanMap = new TreeMap<Integer, ArrayList<FlexGridData>>();
		mColumnSpanMap = new TreeMap<Integer, ArrayList<FlexGridData>>();

		for (int i = 0; i < mColumns; i++) {
			mMaxColumnWidths[i] = LayoutSize.MAXIMUM_SIZE;
		}
		for (int i = 0; i < mRows; i++) {
			mMaxRowHeights[i] = LayoutSize.MAXIMUM_SIZE;
		}

		// Fill in actual cells
		for (Map.Entry<FlexCell, FlexGridData> entry : mData.entrySet()) {
			FlexCell cell = entry.getKey();
			FlexGridData data = entry.getValue();
			for (int row = 0; row < data.mRowSpan; row++) {
				for (int column = 0; column < data.mColumnSpan; column++) {
					int rp = data.mRow + row;
					int cp = data.mColumn + column;
					if (mGrid[rp][cp] != null) {
						throw new InvalidParameterException(data + ": " + FlexGridData.formatRowColumn(rp / 2, cp / 2, false) + " already occupied by " + mData.get(mGrid[rp][cp])); //$NON-NLS-1$ //$NON-NLS-2$
					}
					mGrid[rp][cp] = data;
				}
			}
			data.mSize = cell.getSize(type);
			data.mMinSize = cell.getSize(LayoutSize.MINIMUM);
			data.mMaxSize = cell.getSize(LayoutSize.MAXIMUM);
			updateRowColumnSizes(data);
			addToSpanMap(data.mRowSpan, mRowSpanMap, data);
			addToSpanMap(data.mColumnSpan, mColumnSpanMap, data);
		}

		// Fill in gaps
		for (int row = 0; row < mRows; row++) {
			for (int column = 0; column < mColumns; column++) {
				if (mGrid[row][column] == null) {
					mGrid[row][column] = new FlexGridData(row, column, row % 2 == 1 ? getVerticalGap() : 0, column % 2 == 1 ? getHorizontalGap() : 0);
					updateRowColumnSizes(mGrid[row][column]);
				}
			}
		}

		processRowSpanMap();
		processColumnSpanMap();
	}

	private void processRowSpanMap() {
		for (ArrayList<FlexGridData> set : mRowSpanMap.values()) {
			for (FlexGridData data : set) {
				int last = data.getLastRow();
				int height = 0;
				for (int row = data.mRow; row <= last; row++) {
					height += mRowHeights[row];
				}
				if (height < data.mSize.height) {
					int count = data.getNonGapRowSpan();
					height = data.mSize.height - height;
					while (height > 0 && count > 0) {
						int extra = height / count;
						count = 0;
						for (int row = data.mRow; row <= last && height > 0; row += 2) {
							int rHeight = mRowHeights[row];
							int maxHeight = mMaxRowHeights[row];
							if (rHeight < maxHeight) {
								maxHeight -= rHeight;
								int amt = extra <= maxHeight ? extra : maxHeight;
								mRowHeights[row] += amt;
								height -= amt;
								if (amt != maxHeight) {
									count++;
								}
							}
						}
					}

					// No rows are accepting more space, so force them taller
					count = data.getNonGapRowSpan();
					while (height > 0) {
						int extra = height / count;
						if (extra < 1) {
							extra = 1;
						}
						for (int row = data.mRow; row <= last && height > 0; row += 2) {
							mRowHeights[row] += extra;
							mMaxRowHeights[row] += extra;
							height -= extra;
						}
					}
				}
			}
		}
	}

	private void processColumnSpanMap() {
		for (ArrayList<FlexGridData> set : mColumnSpanMap.values()) {
			for (FlexGridData data : set) {
				int last = data.getLastColumn();
				int width = 0;
				for (int column = data.mColumn; column <= last; column++) {
					width += mColumnWidths[column];
				}
				if (width < data.mSize.width) {
					int count = data.getNonGapColumnSpan();
					width = data.mSize.width - width;
					while (width > 0 && count > 0) {
						int extra = width / count;
						count = 0;
						for (int column = data.mColumn; column <= last && width > 0; column += 2) {
							int cWidth = mColumnWidths[column];
							int maxWidth = mMaxColumnWidths[column];
							if (cWidth < maxWidth) {
								maxWidth -= cWidth;
								int amt = extra <= maxWidth ? extra : maxWidth;
								mColumnWidths[column] += amt;
								width -= amt;
								if (amt != maxWidth) {
									count++;
								}
							}
						}
					}

					// No columns are accepting more space, so force them wider
					count = data.getNonGapColumnSpan();
					while (width > 0) {
						int extra = width / count;
						if (extra < 1) {
							extra = 1;
						}
						for (int column = data.mColumn; column <= last && width > 0; column += 2) {
							mColumnWidths[column] += extra;
							mMaxColumnWidths[column] += extra;
							width -= extra;
						}
					}
				}
			}
		}
	}

	private void addToSpanMap(int span, TreeMap<Integer, ArrayList<FlexGridData>> map, FlexGridData data) {
		if (span > 1) {
			Integer key = new Integer(span);
			ArrayList<FlexGridData> set = map.get(key);
			if (set == null) {
				set = new ArrayList<FlexGridData>();
				map.put(key, set);
			}
			set.add(data);
		}
	}

	private void updateRowColumnSizes(FlexGridData data) {
		if (data.mRowSpan == 1) {
			updateSizes(data.mRow, data.mSize.height, mRowHeights, data.mMinSize.height, mMinRowHeights, data.mMaxSize.height, mMaxRowHeights);
		}
		if (data.mColumnSpan == 1) {
			updateSizes(data.mColumn, data.mSize.width, mColumnWidths, data.mMinSize.width, mMinColumnWidths, data.mMaxSize.width, mMaxColumnWidths);
		}
	}

	private void updateSizes(int index, int cur, int[] curArray, int min, int[] minArray, int max, int[] maxArray) {
		if (min > minArray[index]) {
			minArray[index] = min;
		}
		if (max < maxArray[index]) {
			maxArray[index] = max;
		}
		if (cur > curArray[index]) {
			curArray[index] = cur;
		}
		if (curArray[index] < minArray[index]) {
			curArray[index] = minArray[index];
		}
		if (curArray[index] > maxArray[index]) {
			maxArray[index] = curArray[index];
		}
	}
}
