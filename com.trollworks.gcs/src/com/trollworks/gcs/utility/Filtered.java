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

package com.trollworks.gcs.utility;

import java.util.ArrayList;
import java.util.List;

public final class Filtered {
    private Filtered() {
    }

    /**
     * @param iterable     The {@link Iterable} to filter by type.
     * @param contentClass The class of objects to extract from the collection.
     * @return A new {@link List} with just the objects that are of the specified type. Will not
     *         include {@code null} values.
     */
    public static <T> List<T> list(Iterable<?> iterable, Class<T> contentClass) {
        List<T> list = new ArrayList<>();
        for (T item : new FilteredIterator<>(iterable, contentClass)) {
            list.add(item);
        }
        return list;
    }
}
