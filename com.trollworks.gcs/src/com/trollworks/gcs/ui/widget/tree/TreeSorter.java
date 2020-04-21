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

package com.trollworks.gcs.ui.widget.tree;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;

/** Provides sorting of tree data. */
public class TreeSorter implements Comparator<TreeRow> {
    private List<SortData> mSortData = new ArrayList<>();

    /** @return Whether or not multiple sort criteria exists. */
    public boolean hasMultipleCriteria() {
        return mSortData.size() > 1;
    }

    /**
     * @param column The {@link TreeColumn} to look for.
     * @return The {@link TreeColumn}'s sort sequence, or {@code -1}.
     */
    public int getSortSequence(TreeColumn column) {
        int size = mSortData.size();
        for (int i = 0; i < size; i++) {
            if (mSortData.get(i).mColumn == column) {
                return i;
            }
        }
        return -1;
    }

    /**
     * @param column The {@link TreeColumn} to look for.
     * @return Whether or not the {@link TreeColumn} is sorted in ascending order.
     */
    public boolean isSortAscending(TreeColumn column) {
        for (SortData sortData : mSortData) {
            if (sortData.mColumn == column) {
                return sortData.mAscending;
            }
        }
        return true;
    }

    /** Removes all sort criteria. */
    public void clearSort() {
        mSortData.clear();
    }

    /**
     * @param column    The {@link TreeColumn} to sort on.
     * @param ascending Whether the sort should be ascending or descending.
     * @param replace   Whether to replace the current sort or not.
     */
    public void setSort(TreeColumn column, boolean ascending, boolean replace) {
        if (replace) {
            mSortData.clear();
            mSortData.add(new SortData(column, ascending));
        } else {
            boolean found = false;
            for (SortData data : mSortData) {
                if (data.mColumn == column) {
                    data.mAscending = ascending;
                    found = true;
                    break;
                }
            }
            if (!found) {
                mSortData.add(new SortData(column, ascending));
            }
        }
    }

    /**
     * Sets the sort criteria. If the {@link TreeColumn} is already being sorted, its sort direction
     * will be inverted, otherwise it will be set to ascending.
     *
     * @param column  The {@link TreeColumn} to sort on.
     * @param replace Whether to replace the current sort or not.
     */
    public void setSort(TreeColumn column, boolean replace) {
        SortData found     = null;
        boolean  ascending = true;
        for (SortData data : mSortData) {
            if (data.mColumn == column) {
                ascending = !data.mAscending;
                found = data;
                break;
            }
        }
        if (replace) {
            mSortData.clear();
            mSortData.add(new SortData(column, ascending));
        } else if (found != null) {
            found.mAscending = ascending;
        } else {
            mSortData.add(new SortData(column, true));
        }
    }

    /** @param container The {@link TreeContainerRow} to sort. */
    public void sort(TreeContainerRow container) {
        if (!mSortData.isEmpty()) {
            container.sort(this);
        }
    }

    @Override
    public int compare(TreeRow r1, TreeRow r2) {
        for (SortData sortData : mSortData) {
            int result = sortData.mColumn.compare(r1, r2);
            if (result != 0) {
                return sortData.mAscending ? result : -result;
            }
        }
        return 0;
    }

    static class SortData {
        TreeColumn mColumn;
        boolean    mAscending;

        SortData(TreeColumn column, boolean ascending) {
            mColumn = column;
            mAscending = ascending;
        }
    }
}
