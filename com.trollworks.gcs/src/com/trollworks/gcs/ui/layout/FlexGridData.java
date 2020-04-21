/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.layout;

import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Dimension;

class FlexGridData {
    FlexCell  mCell;
    int       mRow;
    int       mColumn;
    int       mRowSpan;
    int       mColumnSpan;
    Dimension mSize;
    Dimension mMinSize;
    Dimension mMaxSize;

    FlexGridData(FlexCell cell, int row, int column, int rowSpan, int columnSpan) {
        if (row < 0) {
            throw new IllegalArgumentException("row must be >= 0");
        }
        if (column < 0) {
            throw new IllegalArgumentException("column must be >= 0");
        }
        if (rowSpan < 1) {
            throw new IllegalArgumentException("rowSpan must be > 0");
        }
        if (columnSpan < 1) {
            throw new IllegalArgumentException("columnSpan must be > 0");
        }

        mCell = cell;
        mRow = row * 2;
        mColumn = column * 2;
        mRowSpan = 1 + (rowSpan - 1) * 2;
        mColumnSpan = 1 + (columnSpan - 1) * 2;
    }

    FlexGridData(Scale scale, int gapRow, int gapColumn, int rowGap, int columnGap) {
        boolean oddRow    = (gapRow & 1) == 1;
        boolean oddColumn = (gapColumn & 1) == 1;
        mCell = new FlexSpacer(columnGap, rowGap, oddRow && !oddColumn, !oddRow && oddColumn);
        mRow = gapRow;
        mColumn = gapColumn;
        mRowSpan = 1;
        mColumnSpan = 1;
        mSize = mCell.getSize(scale, LayoutSize.PREFERRED);
        mMinSize = mCell.getSize(scale, LayoutSize.MINIMUM);
        mMaxSize = mCell.getSize(scale, LayoutSize.MAXIMUM);
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

    @Override
    public String toString() {
        StringBuilder buffer = new StringBuilder();
        buffer.append(getClass().getSimpleName());
        buffer.append('[');
        if ((mRow & 1) == 1 || (mColumn & 1) == 1) {
            buffer.append('g');
        }
        buffer.append(formatRowColumn(mRow / 2, mColumn / 2, false));
        if (mRowSpan > 1 || mColumnSpan > 1) {
            buffer.append(", span ");
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
