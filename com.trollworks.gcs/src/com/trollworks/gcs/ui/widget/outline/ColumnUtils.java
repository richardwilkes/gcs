/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
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

import java.util.ArrayList;
import java.util.List;

/** Utilities for columns. */
public final class ColumnUtils {
    private ColumnUtils() {
    }

    /**
     * Packs the columns to their preferred sizes.
     *
     * @param outline The {@link Outline} to pack.
     * @param width   The width available for all columns.
     */
    public static void pack(Outline outline, int width) {
        OutlineModel outlineModel = outline.getModel();
        int          count        = outlineModel.getColumnCount();
        List<Column> changed      = new ArrayList<>();
        int[]        widths       = new int[count];
        Column       column;
        if (outline.shouldDrawColumnDividers()) {
            width -= Scale.get(outline).scale(count - 1);
        }
        for (int i = 0; i < count; i++) {
            column = outlineModel.getColumnAtIndex(i);
            widths[i] = column.getPreferredWidth(outline);
            width -= widths[i];
        }
        if (width >= 0) {
            if (width > 0) {
                for (int i = 0; i < count; i++) {
                    column = outlineModel.getColumnAtIndex(i);
                    if (outlineModel.isHierarchyColumn(column)) {
                        widths[i] += width;
                        width = 0;
                        break;
                    }
                }
                if (width > 0) {
                    widths[0] += width;
                }
            }
        } else {
            int   pos     = 0;
            int[] list    = new int[count];
            int[] minList = new int[count];
            for (int i = 0; i < count; i++) {
                column = outlineModel.getColumnAtIndex(i);
                if (column.getRowCell(null).participatesInDynamicRowLayout()) {
                    int min = column.getPreferredHeaderWidth(outline);
                    if (min < widths[i]) {
                        list[pos] = i;
                        minList[pos++] = min;
                    }
                }
            }
            int[] list2    = new int[count];
            int[] minList2 = new int[count];
            int   pos2     = 0;
            while (width < 0 && pos > 0) {
                int amt = -width > pos ? width / pos : -1;
                for (int i = 0; i < pos && width < 0; i++) {
                    int which    = list[i];
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
                column.setWidth(outline, widths[i]);
                changed.add(column);
            }
        }
        outline.updateRowHeightsIfNeeded(changed);
    }
}
