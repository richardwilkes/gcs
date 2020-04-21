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

import java.util.Iterator;
import java.util.List;
import java.util.NoSuchElementException;

/**
 * Provides an iterator that will iterate over all rows (disclosed or not) in an outline model.
 *
 * @param <T> The type of row being iterated over.
 */
public class RowIterator<T extends Row> implements Iterator<T>, Iterable<T> {
    private List<T>        mList;
    private int            mIndex;
    private RowIterator<T> mIterator;
    private Filter<T>      mFilter;

    /**
     * Creates an iterator that will iterate over all rows (disclosed or not) in the specified
     * outline model.
     *
     * @param model The model to iterator over.
     */
    public RowIterator(OutlineModel model) {
        this(model, null);
    }

    /**
     * Creates an iterator that will iterate over all rows (disclosed or not) in the specified
     * outline model.
     *
     * @param model  The model to iterator over.
     * @param filter The filter to use.
     */
    @SuppressWarnings("unchecked")
    public RowIterator(OutlineModel model, Filter<T> filter) {
        this((List<T>) model.getTopLevelRows(), filter);
    }

    private RowIterator(List<T> rows, Filter<T> filter) {
        mList = rows;
        mFilter = filter;
    }

    @Override
    public void remove() {
        throw new UnsupportedOperationException();
    }

    @Override
    public boolean hasNext() {
        boolean hasNext = mIterator != null && mIterator.hasNext();
        if (!hasNext) {
            mIterator = null;
            int size = mList.size();
            if (mFilter != null) {
                while (mIndex < size && !mFilter.include(mList.get(mIndex))) {
                    mIndex++;
                }
            }
            hasNext = mIndex < size;
        }
        return hasNext;
    }

    @Override
    @SuppressWarnings("unchecked")
    public T next() {
        if (hasNext()) {
            if (mIterator == null) {
                T row = mList.get(mIndex++);
                if (row.hasChildren()) {
                    mIterator = new RowIterator<>((List<T>) row.getChildren(), mFilter);
                }
                return row;
            }
            return mIterator.next();
        }
        throw new NoSuchElementException();
    }

    @Override
    public Iterator<T> iterator() {
        return this;
    }

    public interface Filter<R extends Row> {
        boolean include(R row);
    }
}
