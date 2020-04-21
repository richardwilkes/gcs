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

/** Objects interested in {@link OutlineModel} notifications must implement this interface. */
public interface OutlineModelListener {
    /**
     * Called after rows are added.
     *
     * @param model The model that was altered.
     * @param rows  The rows being added.
     */
    void rowsAdded(OutlineModel model, Row[] rows);

    /**
     * Called prior to rows being removed.
     *
     * @param model The affected model.
     * @param rows  The rows being removed.
     */
    void rowsWillBeRemoved(OutlineModel model, Row[] rows);

    /**
     * Called after rows are removed.
     *
     * @param model The affected model.
     * @param rows  The rows that were removed.
     */
    void rowsWereRemoved(OutlineModel model, Row[] rows);

    /**
     * Called after a row is modified by a call to {@link Row#setData(Column, Object)}.
     *
     * @param model  The affected model.
     * @param row    The affected row.
     * @param column The affected column.
     */
    void rowWasModified(OutlineModel model, Row row, Column column);

    /**
     * Called whenever the sort settings are cleared.
     *
     * @param model The model whose sort settings were cleared.
     */
    void sortCleared(OutlineModel model);

    /**
     * Called whenever the model is sorted.
     *
     * @param model     The model that was sorted.
     * @param restoring {@code true} when the sort is being restored (usually due to row
     *                  disclosure).
     */
    void sorted(OutlineModel model, boolean restoring);

    /**
     * Called whenever the "locked" state is about to change.
     *
     * @param model The model whose state will change.
     */
    void lockedStateWillChange(OutlineModel model);

    /**
     * Called whenever the "locked" state changes.
     *
     * @param model The model whose state changed.
     */
    void lockedStateDidChange(OutlineModel model);

    /**
     * Called whenever the selection is about to change.
     *
     * @param model The model whose selection will change.
     */
    void selectionWillChange(OutlineModel model);

    /**
     * Called whenever the selection changes.
     *
     * @param model The model whose selection changed.
     */
    void selectionDidChange(OutlineModel model);

    /**
     * Called whenever an undo/redo is about to be applied.
     *
     * @param model The model which will be affected.
     */
    void undoWillHappen(OutlineModel model);

    /**
     * Called whenever an undo/redo has been applied.
     *
     * @param model The model which was affected.
     */
    void undoDidHappen(OutlineModel model);
}
