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

import java.util.Collection;
import java.util.Set;

/** Objects that want to provide tracking of a selection must implement this interface. */
public interface SelectionModel {
    /** @return {@code true} if at least one object is currently selected. */
    boolean hasSelection();

    /** @return The set of objects currently selected. */
    Set<Object> getSelection();

    /**
     * @param obj The object to check.
     * @return {@code true} if the specified object is selected.
     */
    boolean isSelected(Object obj);

    /**
     * @param obj The object to select.
     * @param add {@code true} if this should be added to an existing selection. {@code false} if it
     *            should replace the existing selection.
     */
    void select(Object obj, boolean add);

    /**
     * @param objs The objects to select.
     * @param add  {@code true} if these should be added to an existing selection. {@code false} if
     *             these should replace the existing selection.
     */
    void select(Collection<?> objs, boolean add);

    /** @param obj The object to deselect. */
    void deselect(Object obj);

    /** @param objs The objects to deselect. */
    void deselect(Collection<?> objs);

    /** Removes all objects from the selection. */
    void clearSelection();
}
