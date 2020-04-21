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

package com.trollworks.gcs.ui.widget.outline;

import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Color;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import javax.swing.SwingConstants;

/** Used to draw headers in a {@link Outline}. */
public class HeaderCell extends TextCell {
    /** The width of the sorter widget. */
    public static final int     SORTER_WIDTH = 12;
    private             int     mSortSequence;
    private             boolean mSortAscending;
    private             Color   mSorterColor;
    private             boolean mAllowSort;

    /** Create a new header cell. */
    public HeaderCell() {
        super(SwingConstants.CENTER);
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
     * @return {@code true} if the column should be sorted in ascending order.
     */
    public boolean isSortAscending() {
        return mSortAscending;
    }

    /**
     * Sets the sort criteria for this column.
     *
     * @param sequence  The column's sort sequence. Use {@code -1} if it has none.
     * @param ascending Pass in {@code true} for an ascending sort.
     */
    public void setSortCriteria(int sequence, boolean ascending) {
        if (mAllowSort) {
            mSortSequence = sequence;
            mSortAscending = ascending;
        }
    }

    /** @return The column's sort sequence, or {@code -1} if it has none. */
    public int getSortSequence() {
        return mSortSequence;
    }

    /**
     * Draws the cell.
     *
     * @param outline  The {@link Outline} being drawn.
     * @param gc       The graphics context to use.
     * @param bounds   The bounds of the cell.
     * @param row      The row to draw.
     * @param column   The column to draw.
     * @param selected Pass in {@code true} if the cell should be drawn in its selected state.
     * @param active   Pass in {@code true} if the cell should be drawn in its active state.
     */
    protected void drawCellSuper(Outline outline, Graphics gc, Rectangle bounds, Row row, Column column, boolean selected, boolean active) {
        super.drawCell(outline, gc, bounds, row, column, selected, active);
    }

    @Override
    public void drawCell(Outline outline, Graphics gc, Rectangle bounds, Row row, Column column, boolean selected, boolean active) {
        if (mAllowSort) {
            Scale scale       = Scale.get(outline);
            int   sorterWidth = scale.scale(SORTER_WIDTH);
            int   two         = scale.scale(2);
            bounds.x += two;
            bounds.width -= two + two;

            if (mSortSequence != -1) {
                bounds.width -= sorterWidth;
            }

            drawCellSuper(outline, gc, bounds, row, column, selected, active);

            if (mSortSequence != -1) {
                int three = scale.scale(3);
                int x;
                int y;
                int i;

                bounds.width += sorterWidth;
                x = bounds.x + bounds.width - (three + three);
                y = bounds.y + bounds.height / 2 - 1;

                gc.setColor(getSorterColor());
                int count = 0;
                for (Column one : outline.getModel().getColumns()) {
                    if (one.getSortSequence() >= 0) {
                        if (++count > 1) {
                            break;
                        }
                    }
                }
                int one = scale.scale(1);
                if (count > 1) {
                    for (i = 0; i <= mSortSequence; i++) {
                        gc.fillRect(bounds.x + one + three * i, bounds.y + bounds.height - three, two, two);
                    }
                }

                int four = scale.scale(4);
                int five = scale.scale(5);
                for (i = 0; i < five; i++) {
                    gc.fillRect(mSortAscending ? x - i : x + i - four, y + i, 1 + (mSortAscending ? i + i : (five - 1 - i) * 2), 1);
                }
            }

            bounds.x -= two;
            bounds.width += two + two;
        } else {
            drawCellSuper(outline, gc, bounds, row, column, selected, active);
        }
    }

    @Override
    public int getPreferredWidth(Outline outline, Row row, Column column) {
        int width = super.getPreferredWidth(outline, row, column);
        if (mAllowSort) {
            Scale scale  = Scale.get(outline);
            int   margin = scale.scale(2);
            width += margin + scale.scale(SORTER_WIDTH) + margin;
        }
        return width;
    }

    @Override
    public String getToolTipText(Outline outline, MouseEvent event, Rectangle bounds, Row row, Column column) {
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
