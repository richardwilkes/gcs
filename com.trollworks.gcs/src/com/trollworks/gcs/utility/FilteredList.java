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

import java.util.ArrayList;

/**
 * A list that filters an {@link Iterable} to only contain objects of a specific type.
 *
 * @param <T> The type of object the list should contain.
 */
public class FilteredList<T> extends ArrayList<T> {
    /**
     * Creates a new {@link FilteredList}. Will not include {@code null} values.
     *
     * @param iterable     The {@link Iterable} to filter by type.
     * @param contentClass The class of objects to extract from the collection.
     */
    public FilteredList(Iterable<?> iterable, Class<T> contentClass) {
        this(iterable, contentClass, true);
    }

    /**
     * Creates a new {@link FilteredList}.
     *
     * @param iterable     The {@link Iterable} to filter by type.
     * @param contentClass The class of objects to extract from the collection.
     * @param omitNulls    Whether to omit {@code null} values or not.
     */
    public FilteredList(Iterable<?> iterable, Class<T> contentClass, boolean omitNulls) {
        for (T item : new FilteredIterator<>(iterable, contentClass, omitNulls)) {
            add(item);
        }
    }
}
