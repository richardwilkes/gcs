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
import java.util.List;
import java.util.NoSuchElementException;

public class ReverseListIterator<T> implements Iterator<T>, Iterable<T> {
    private List<T> mList;
    private int     mPos;

    public ReverseListIterator(List<T> list) {
        mList = list;
        mPos = mList.size() - 1;
    }

    @Override
    public boolean hasNext() {
        return mPos >= 0;
    }

    @Override
    public T next() throws NoSuchElementException {
        if (mPos < 0) {
            throw new NoSuchElementException();
        }
        return mList.get(mPos--);
    }

    @Override
    public Iterator<T> iterator() {
        return this;
    }
}
