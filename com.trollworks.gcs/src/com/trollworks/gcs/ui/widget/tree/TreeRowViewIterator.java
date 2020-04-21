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

import java.util.Arrays;
import java.util.Iterator;
import java.util.List;
import java.util.NoSuchElementException;

/** Iterates through all {@link TreeRow}s, ignoring children of closed ones. */
public class TreeRowViewIterator implements Iterator<TreeRow>, Iterable<TreeRow> {
    private TreePanel           mPanel;
    private List<TreeRow>       mRows;
    private TreeRowViewIterator mIterator;
    private int                 mIndex;

    /**
     * Creates a new {@link TreeRowViewIterator}.
     *
     * @param panel The owning {@link TreePanel}.
     * @param rows  The {@link TreeRow}s to iterator over.
     */
    public TreeRowViewIterator(TreePanel panel, TreeRow... rows) {
        mPanel = panel;
        mRows = Arrays.asList(rows);
    }

    /**
     * Creates a new {@link TreeRowViewIterator}.
     *
     * @param panel The owning {@link TreePanel}.
     * @param rows  The {@link TreeRow}s to iterator over.
     */
    public TreeRowViewIterator(TreePanel panel, List<TreeRow> rows) {
        mPanel = panel;
        mRows = rows;
    }

    @Override
    public Iterator<TreeRow> iterator() {
        return this;
    }

    @Override
    public boolean hasNext() {
        boolean hasNext = mIterator != null && mIterator.hasNext();
        if (!hasNext) {
            mIterator = null;
            hasNext = mIndex < mRows.size();
        }
        return hasNext;
    }

    @Override
    public TreeRow next() {
        if (hasNext()) {
            if (mIterator == null) {
                TreeRow row = mRows.get(mIndex++);
                if (row instanceof TreeContainerRow) {
                    TreeContainerRow containerRow = (TreeContainerRow) row;
                    if (containerRow.getChildCount() > 0 && mPanel.isOpen(containerRow)) {
                        mIterator = new TreeRowViewIterator(mPanel, containerRow.getChildren());
                    }
                }
                return row;
            }
            return mIterator.next();
        }
        throw new NoSuchElementException();
    }

    @Override
    public void remove() {
        throw new UnsupportedOperationException();
    }
}
