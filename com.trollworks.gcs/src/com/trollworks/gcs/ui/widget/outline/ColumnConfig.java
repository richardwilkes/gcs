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

/** Stores the configuration of an {@link Column}. */
public class ColumnConfig {
    /** The id of the column */
    public int     mID;
    /** The visibility of the column */
    public boolean mVisible;
    /** The width of the column */
    public int     mWidth;
    /** the sort sequence of the column */
    public int     mSortSequence;
    /** {@code true} if the sort is ascending */
    public boolean mSortAscending;

    /**
     * Creates a new {@link ColumnConfig} with the given id.
     *
     * @param id The id of the column.
     */
    public ColumnConfig(int id) {
        this(id, true);
    }

    /**
     * Creates a new {@link ColumnConfig} with the given id and visibility.
     *
     * @param id      The id of the column.
     * @param visible The visiblity of the column.
     */
    public ColumnConfig(int id, boolean visible) {
        this(id, visible, -1, false);
    }

    /**
     * Creates a new {@link ColumnConfig} with the given id, visibility, sort sequence and if the
     * sort is ascending.
     *
     * @param id            The id of the column.
     * @param visible       The visiblity of the column.
     * @param sortSequence  The sort sequence of the column.
     * @param sortAscending {@code true} if the sort is ascending.
     */
    public ColumnConfig(int id, boolean visible, int sortSequence, boolean sortAscending) {
        this(id, visible, -1, sortSequence, sortAscending);
    }

    /**
     * Creates a new {@link ColumnConfig} with the given id, visibility, width, sort sequence and if
     * the sort is ascending.
     *
     * @param id            The id of the column.
     * @param visible       The visiblity of the column.
     * @param width         The width of the column.
     * @param sortSequence  The sort sequence of the column.
     * @param sortAscending {@code true} if the sort is ascending.
     */
    public ColumnConfig(int id, boolean visible, int width, int sortSequence, boolean sortAscending) {
        mID = id;
        mVisible = visible;
        mWidth = width;
        mSortSequence = sortSequence;
        mSortAscending = sortAscending;
    }
}
