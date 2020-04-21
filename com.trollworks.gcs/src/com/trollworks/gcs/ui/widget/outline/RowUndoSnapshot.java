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

import java.util.ArrayList;
import java.util.List;

/** The information an undo for the row needs to operate. */
public class RowUndoSnapshot {
    private Row       mParent;
    private boolean   mOpen;
    private List<Row> mChildren;

    /**
     * Creates a snapshot of the information needed to undo any changes to the row.
     *
     * @param row The row to create a snapshot for.
     */
    public RowUndoSnapshot(Row row) {
        mParent = row.getParent();
        mOpen = row.isOpen();
        mChildren = row.canHaveChildren() ? new ArrayList<>(row.getChildren()) : null;
    }

    /** @return The children. */
    public List<Row> getChildren() {
        return mChildren;
    }

    /** @return Whether the row should be open. */
    public boolean isOpen() {
        return mOpen;
    }

    /** @return The parent. */
    public Row getParent() {
        return mParent;
    }
}
