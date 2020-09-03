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

package com.trollworks.gcs.character;

import com.trollworks.gcs.ui.border.TitledBorder;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.widget.outline.ColumnUtils;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.Row;

import java.awt.Insets;

/** Holds information about the outline relevant for page layout. */
public class OutlineInfo {
    private int   mRowIndex;
    private int[] mHeights;
    private int   mOverheadHeight;
    private int   mMinimumHeight;

    /**
     * Creates a new outline information holder.
     *
     * @param outline      The outline to collect information about.
     * @param contentWidth The content width.
     */
    public OutlineInfo(Outline outline, int contentWidth) {
        int          one            = Scale.get(outline).scale(1);
        Insets       insets         = new TitledBorder().getBorderInsets(outline);
        OutlineModel outlineModel   = outline.getModel();
        int          count          = outlineModel.getRowCount();
        boolean      hasRowDividers = outline.shouldDrawRowDividers();

        ColumnUtils.pack(outline, contentWidth - (insets.left + insets.right));
        outline.updateRowHeights();

        mRowIndex = -1;
        mHeights = new int[count];

        for (int i = 0; i < count; i++) {
            Row row = outlineModel.getRowAtIndex(i);
            mHeights[i] = row.getHeight();
            if (hasRowDividers) {
                mHeights[i] += one;
            }
        }

        mOverheadHeight = insets.top + insets.bottom + outline.getHeaderPanel().getPreferredSize().height;
        mMinimumHeight = mOverheadHeight + (count > 0 ? mHeights[0] : 0);
    }

    /**
     * @param remaining The remaining vertical space on the page.
     * @return The space the outline will consume before being complete or requiring another page.
     */
    public int determineHeightForOutline(int remaining) {
        int total  = mOverheadHeight;
        int start  = ++mRowIndex;
        int length = mHeights.length;
        while (true) {
            if (mRowIndex >= length) {
                return total;
            }
            total += mHeights[mRowIndex];
            if (total > remaining) {
                if (mRowIndex == start) {
                    return total;
                }
                return total - mHeights[mRowIndex--];
            }
            mRowIndex++;
        }
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
