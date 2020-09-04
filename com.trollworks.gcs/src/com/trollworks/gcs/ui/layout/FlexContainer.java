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

import java.awt.Component;
import java.awt.Dimension;
import java.awt.Rectangle;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

/** A {@link FlexCell} that contains other {@link FlexCell}s. */
public abstract class FlexContainer extends FlexCell {
    private List<FlexCell> mChildren      = new ArrayList<>();
    private int            mHorizontalGap = 5;
    private int            mVerticalGap   = 2;
    private boolean        mFillHorizontal;
    private boolean        mFillVertical;

    /** @param cell The {@link FlexCell} to add as a child. */
    public void add(FlexCell cell) {
        mChildren.add(cell);
    }

    /** @param comp The {@link Component} to add as a child. */
    public void add(Component comp) {
        mChildren.add(new FlexComponent(comp));
    }

    /** @return The number of children of this {@link FlexContainer}. */
    protected int getChildCount() {
        return mChildren.size();
    }

    /** @return The children of this {@link FlexContainer}. */
    protected List<FlexCell> getChildren() {
        return mChildren;
    }

    /**
     * @param scale The {@link Scale} to use.
     * @param type  The type of size to return.
     * @return The sizes for each child.
     */
    protected Dimension[] getChildSizes(Scale scale, LayoutSize type) {
        int         count = getChildCount();
        Dimension[] sizes = new Dimension[count];
        for (int i = 0; i < count; i++) {
            sizes[i] = mChildren.get(i).getSize(scale, type);
        }
        return sizes;
    }

    /**
     * @param scale  The {@link Scale} to use.
     * @param bounds The bounds to use for each child.
     */
    protected void layoutChildren(Scale scale, Rectangle[] bounds) {
        int length = bounds.length;
        for (int i = 0; i < length; i++) {
            mChildren.get(i).layout(scale, bounds[i]);
        }
    }

    /** @return The horizontal gap between cells. */
    public int getHorizontalGap() {
        return mHorizontalGap;
    }

    /** @param horizontalGap The value to set for horizontal gap between cells. */
    public void setHorizontalGap(int horizontalGap) {
        mHorizontalGap = horizontalGap;
    }

    /** @return The vertical gap between components. */
    public int getVerticalGap() {
        return mVerticalGap;
    }

    /** @param verticalGap The value to set for vertical gap between cells. */
    public void setVerticalGap(int verticalGap) {
        mVerticalGap = verticalGap;
    }

    /** @param fill Whether all space will be taken up by expanding the gaps, if necessary. */
    public void setFill(boolean fill) {
        mFillHorizontal = fill;
        mFillVertical = fill;
    }

    /**
     * @return Whether all horizontal space will be taken up by expanding the gaps, if necessary.
     */
    public boolean getFillHorizontal() {
        return mFillHorizontal;
    }

    /**
     * @param fill Whether all horizontal space will be taken up by expanding the gaps, if
     *             necessary.
     */
    public void setFillHorizontal(boolean fill) {
        mFillHorizontal = fill;
    }

    /** @return Whether all vertical space will be taken up by expanding the gaps, if necessary. */
    public boolean getFillVertical() {
        return mFillVertical;
    }

    /**
     * @param fill Whether all vertical space will be taken up by expanding the gaps, if necessary.
     */
    public void setFillVertical(boolean fill) {
        mFillVertical = fill;
    }

    /**
     * Distribute an amount.
     *
     * @param amt    The amount to distribute.
     * @param values The initial values. On return, these will have been adjusted.
     * @param limits The limits for the values.
     * @return Any leftover amount.
     */
    @SuppressWarnings("static-method")
    protected int distribute(int amt, int[] values, int[] limits) {
        if (amt < 0) {
            return distributeShrink(-amt, values, limits);
        }
        return distributeGrow(amt, values, limits);
    }

    private static int distributeShrink(int amt, int[] values, int[] min) {
        int   valuesLength = values.length;
        int[] orig         = new int[valuesLength];
        System.arraycopy(values, 0, orig, 0, valuesLength);
        int   minLength = min.length;
        int[] max       = new int[minLength];
        for (int i = 0; i < minLength; i++) {
            max[i] = values[i] * 2 - min[i];
        }
        amt = distributeGrow(amt, values, max);
        for (int i = 0; i < valuesLength; i++) {
            values[i] = orig[i] * 2 - values[i];
        }
        return -amt;
    }

    private static int distributeGrow(int amt, int[] values, int[] max) {
        // Copy the values and sort them from smallest to largest
        int   length = values.length;
        int[] order  = new int[length];
        System.arraycopy(values, 0, order, 0, length);
        Arrays.sort(order);

        // Find the next-to-smallest
        int pos = 1;
        while (pos < length && order[pos] == order[pos - 1]) {
            pos++;
        }

        // Go through each position and try to expand it
        for (; pos < length && amt > 0; pos++) {
            amt = fill(amt, order[pos], values, max);
        }
        if (amt > 0) {
            amt = fill(amt, LayoutSize.MAXIMUM_SIZE, values, max);
        }

        return amt;
    }

    private static int fill(int amt, int upTo, int[] values, int[] max) {
        int count  = 0;
        int total  = 0;
        int length = values.length;
        for (int i = 0; i < length; i++) {
            if (values[i] < upTo && values[i] < max[i]) {
                total += Math.min(upTo, max[i]) - values[i];
                count++;
            }
        }
        if (count > 0) {
            if (total <= amt) {
                for (int i = 0; i < length; i++) {
                    if (values[i] < upTo && values[i] < max[i]) {
                        values[i] = Math.min(upTo, max[i]);
                    }
                }
                amt -= total;
            } else {
                while (count > 0 && amt > 0) {
                    int portion = Math.max(amt / count, 1);
                    count = 0;
                    for (int i = 0; i < length && amt > 0; i++) {
                        if (values[i] < upTo && values[i] < max[i]) {
                            if (values[i] + portion <= max[i]) {
                                values[i] += portion;
                                amt -= portion;
                                if (values[i] != upTo) {
                                    count++;
                                }
                            } else {
                                amt -= max[i] - values[i];
                                values[i] = max[i];
                            }
                        }
                    }
                }
            }
        }
        return amt;
    }

    @Override
    public String toString() {
        StringBuilder buffer = new StringBuilder();
        buffer.append(super.toString());
        buffer.append('[');
        boolean needComma = false;
        for (FlexCell cell : mChildren) {
            if (needComma) {
                buffer.append(", ");
            } else {
                needComma = true;
            }
            buffer.append(cell);
        }
        buffer.append(']');
        return buffer.toString();
    }
}
