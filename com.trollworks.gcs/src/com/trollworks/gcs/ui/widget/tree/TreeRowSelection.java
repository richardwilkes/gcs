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

import java.awt.datatransfer.DataFlavor;
import java.awt.datatransfer.Transferable;
import java.awt.datatransfer.UnsupportedFlavorException;
import java.util.ArrayList;
import java.util.Collection;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

/** Allows rows to be part of drag and drop operations internal to the JVM. */
public class TreeRowSelection implements Transferable {
    /** The data flavor for this class. */
    public static final DataFlavor            DATA_FLAVOR = new DataFlavor(TreeRowSelection.class, "Tree Rows");
    private             List<TreeRow>         mRows;
    private             Set<TreeContainerRow> mOpenRows;

    /**
     * Creates a new transferable row object.
     *
     * @param rows     The {@link TreeRow}s to transfer.
     * @param openRows The {@link TreeContainerRow}s within rows which are 'open'.
     */
    public TreeRowSelection(Collection<TreeRow> rows, Collection<TreeContainerRow> openRows) {
        mRows = new ArrayList<>(rows);
        mOpenRows = new HashSet<>(openRows);
    }

    @Override
    public DataFlavor[] getTransferDataFlavors() {
        return new DataFlavor[]{DATA_FLAVOR};
    }

    @Override
    public boolean isDataFlavorSupported(DataFlavor flavor) {
        return DATA_FLAVOR.equals(flavor);
    }

    @Override
    public Object getTransferData(DataFlavor flavor) throws UnsupportedFlavorException {
        if (DATA_FLAVOR.equals(flavor)) {
            return this;
        }
        throw new UnsupportedFlavorException(flavor);
    }

    /** @return The {@link TreeRow}s being transferred. */
    public List<TreeRow> getRows() {
        return mRows;
    }

    /** @return The {@link TreeContainerRow}s within rows which are 'open'. */
    public Set<TreeContainerRow> getOpenRows() {
        return mOpenRows;
    }
}
