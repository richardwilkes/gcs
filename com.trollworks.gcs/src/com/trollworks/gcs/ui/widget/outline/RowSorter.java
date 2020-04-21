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

import java.util.Comparator;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

/** Sorts rows by the sort sequence specified in the associated columns. */
public class RowSorter implements Comparator<Row> {
    private Column[] mSortingOrder;

    private RowSorter(List<Column> columns) {
        int      count = columns.size();
        Column[] orig  = new Column[count];
        int      pos   = -1;
        int      i;

        mSortingOrder = new Column[count];

        for (i = 0; i < count; i++) {
            Column column = columns.get(i);
            int    order  = column.getSortSequence();

            if (order >= 0 && order < count) {
                mSortingOrder[order] = column;
            } else {
                orig[i] = column;
                if (pos == -1) {
                    pos = i;
                }
            }
        }

        if (pos != -1) {
            for (i = 0; i < count; i++) {
                if (mSortingOrder[i] == null) {
                    mSortingOrder[i] = orig[pos++];
                    while (pos < count && orig[pos] == null) {
                        pos++;
                    }
                    if (pos >= count) {
                        break;
                    }
                }
            }
        }
    }

    /**
     * Used to sort an outline.
     *
     * @param columns The columns in the {@link Outline}.
     * @param rows    The rows in the {@link Outline}.
     */
    public static void sort(List<Column> columns, List<Row> rows) {
        sort(columns, rows, false);
    }

    /**
     * Used to sort an outline.
     *
     * @param columns  The columns in the {@link Outline}.
     * @param rows     The rows in the {@link Outline}.
     * @param internal Pass in {@code true} if the actual row child storage should also be sorted.
     */
    public static void sort(List<Column> columns, List<Row> rows, boolean internal) {
        for (Column column : columns) {
            if (column.getSortSequence() != -1) {
                RowSorter rowSorter = new RowSorter(columns);

                rows.sort(rowSorter);
                if (internal) {
                    for (Row row : collectContainerRows(rows, new HashSet<>())) {
                        if (row.hasChildren()) {
                            row.getChildList().sort(rowSorter);
                        }
                    }
                }
                return;
            }
        }
    }

    /**
     * Collects all container rows from the passed in rows and their children.
     *
     * @param rows       The rows to collect container rows from.
     * @param containers The set to add the container rows to.
     * @return The passed in set.
     */
    public static Set<Row> collectContainerRows(List<Row> rows, Set<Row> containers) {
        for (Row row : rows) {
            if (row.canHaveChildren()) {
                containers.add(row);
                if (row.hasChildren()) {
                    collectContainerRows(row.getChildren(), containers);
                }
            }
        }
        return containers;
    }

    @Override
    public int compare(Row rowOne, Row rowTwo) {
        if (rowOne.getParent() == rowTwo.getParent()) {
            for (Column column : mSortingOrder) {
                int result;

                if (column == null) {
                    return 0;
                }
                result = column.getRowCell(null).compare(column, rowOne, rowTwo);
                if (result != 0) {
                    return column.isSortAscending() ? result : -result;
                }
            }
        } else {
            if (rowOne.isDescendantOf(rowTwo)) {
                return 1;
            }
            if (rowTwo.isDescendantOf(rowOne)) {
                return -1;
            }

            // Find common parents and compare them...
            Row[] oneParents = rowOne.getPath();
            Row[] twoParents = rowTwo.getPath();
            int   max        = Math.min(oneParents.length, twoParents.length);
            int   i;

            for (i = 0; i < max; i++) {
                if (oneParents[i] != twoParents[i]) {
                    break;
                }
            }

            return compare(oneParents[i], twoParents[i]);
        }
        return 0;
    }
}
