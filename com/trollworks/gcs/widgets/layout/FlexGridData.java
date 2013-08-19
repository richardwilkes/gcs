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

package com.trollworks.gcs.widgets.layout;

import java.awt.Dimension;

class FlexGridData {
	FlexCell	mCell;
	int			mRow;
	int			mColumn;
	int			mRowSpan;
	int			mColumnSpan;
	Dimension	mSize;
	Dimension	mMinSize;
	Dimension	mMaxSize;

	FlexGridData(FlexCell cell, int row, int column, int rowSpan, int columnSpan) {
		if (row < 0) {
			throw new IllegalArgumentException("row must be >= 0"); //$NON-NLS-1$
		}
		if (column < 0) {
			throw new IllegalArgumentException("column must be >= 0"); //$NON-NLS-1$
		}
		if (rowSpan < 1) {
			throw new IllegalArgumentException("rowSpan must be > 0"); //$NON-NLS-1$
		}
		if (columnSpan < 1) {
			throw new IllegalArgumentException("columnSpan must be > 0"); //$NON-NLS-1$
		}

		mCell = cell;
		mRow = row * 2;
		mColumn = column * 2;
		mRowSpan = 1 + (rowSpan - 1) * 2;
		mColumnSpan = 1 + (columnSpan - 1) * 2;
	}

	FlexGridData(int gapRow, int gapColumn, int rowGap, int columnGap) {
		boolean oddRow = (gapRow & 1) == 1;
		boolean oddColumn = (gapColumn & 1) == 1;
		mCell = new FlexSpacer(columnGap, rowGap, oddRow && !oddColumn, !oddRow && oddColumn);
		mRow = gapRow;
		mColumn = gapColumn;
		mRowSpan = 1;
		mColumnSpan = 1;
		mSize = mCell.getSize(LayoutSize.PREFERRED);
		mMinSize = mCell.getSize(LayoutSize.MINIMUM);
		mMaxSize = mCell.getSize(LayoutSize.MAXIMUM);
	}

	int getLastRow() {
		return mRow + mRowSpan - 1;
	}

	int getLastColumn() {
		return mColumn + mColumnSpan - 1;
	}

	int getNonGapRowSpan() {
		return 1 + (mRowSpan - 1) / 2;
	}

	int getNonGapColumnSpan() {
		return 1 + (mColumnSpan - 1) / 2;
	}

	@Override public String toString() {
		StringBuilder buffer = new StringBuilder();
		buffer.append(getClass().getSimpleName());
		buffer.append('[');
		if (mRow % 2 == 1 || mColumn % 2 == 1) {
			buffer.append('g');
		}
		buffer.append(formatRowColumn(mRow / 2, mColumn / 2, false));
		if (mRowSpan > 1 || mColumnSpan > 1) {
			buffer.append(", span "); //$NON-NLS-1$
			buffer.append(formatRowColumn(getNonGapRowSpan(), getNonGapColumnSpan(), true));
		}
		buffer.append(']');
		return buffer.toString();
	}

	static String formatRowColumn(int row, int column, boolean omitOnes) {
		StringBuilder buffer = new StringBuilder();
		if (!omitOnes || row != 1) {
			buffer.append('r');
			buffer.append(row);
		}
		if (!omitOnes || column != 1) {
			buffer.append('c');
			buffer.append(column);
		}
		return buffer.toString();
	}
}
