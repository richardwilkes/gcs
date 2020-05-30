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

package com.trollworks.gcs.utility;

import java.util.Iterator;
import java.util.NoSuchElementException;

/**
 * An {@link Iterator} that filters an {@link Iterable} to return only objects of a specific type.
 *
 * @param <T> The type of object the {@link Iterator} should return.
 */
public class FilteredIterator<T> implements Iterator<T>, Iterable<T> {
    private Iterator<?> mIterator;
    private T           mNext;
    private Class<T>    mContentClass;
    private boolean     mNextValid;
    private boolean     mOmitNulls;

    /**
     * Creates a new {@link FilteredIterator}. Will not include {@code null} values.
     *
     * @param iterable     The {@link Iterable} to filter by type.
     * @param contentClass The class of objects to extract from the collection.
     */
    public FilteredIterator(Iterable<?> iterable, Class<T> contentClass) {
        this(iterable, contentClass, true);
    }

    /**
     * Creates a new {@link FilteredIterator}.
     *
     * @param iterable     The {@link Iterable} to filter by type.
     * @param contentClass The class of objects to extract from the collection.
     * @param omitNulls    Whether to omit {@code null} values or not.
     */
    public FilteredIterator(Iterable<?> iterable, Class<T> contentClass, boolean omitNulls) {
        mIterator = iterable.iterator();
        mContentClass = contentClass;
        mOmitNulls = omitNulls;
    }

    @Override
    public boolean hasNext() {
        if (mNextValid) {
            return true;
        }
        while (mIterator.hasNext()) {
            Object obj = mIterator.next();

            if (obj == null) {
                if (!mOmitNulls) {
                    mNext = null;
                    mNextValid = true;
                    return true;
                }
            } else if (mContentClass.isInstance(obj)) {
                mNext = mContentClass.cast(obj);
                mNextValid = true;
                return true;
            }
        }
        return false;
    }

    @Override
    public T next() {
        if (!mNextValid) {
            if (!hasNext()) {
                throw new NoSuchElementException();
            }
        }
        mNextValid = false;
        return mNext;
    }

    /** Not supported. */
    @Override
    public void remove() {
        throw new UnsupportedOperationException();
    }

    @Override
    public Iterator<T> iterator() {
        return this;
    }
}
