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

import java.awt.datatransfer.DataFlavor;
import java.awt.datatransfer.Transferable;
import java.awt.datatransfer.UnsupportedFlavorException;

/** Allows rows to be part of drag and drop operations internal to the JVM. */
public class RowSelection implements Transferable {
    /** The data flavor for this class. */
    public static final DataFlavor   DATA_FLAVOR = new DataFlavor(RowSelection.class, "Outline Rows");
    private             OutlineModel mModel;
    private             Row[]        mRows;
    private             String       mCache;

    /**
     * Creates a new transferable row object.
     *
     * @param model The owning outline model.
     * @param rows  The rows to transfer.
     */
    public RowSelection(OutlineModel model, Row[] rows) {
        mModel = model;
        mRows = new Row[rows.length];
        System.arraycopy(rows, 0, mRows, 0, rows.length);
    }

    @Override
    public DataFlavor[] getTransferDataFlavors() {
        return new DataFlavor[]{DATA_FLAVOR, DataFlavor.stringFlavor};
    }

    @Override
    public boolean isDataFlavorSupported(DataFlavor flavor) {
        return DATA_FLAVOR.equals(flavor) || DataFlavor.stringFlavor.equals(flavor);
    }

    @Override
    public Object getTransferData(DataFlavor flavor) throws UnsupportedFlavorException {
        if (DATA_FLAVOR.equals(flavor)) {
            return mRows;
        }
        if (DataFlavor.stringFlavor.equals(flavor)) {
            if (mCache == null) {
                StringBuilder buffer = new StringBuilder();

                if (mRows.length > 0) {
                    int count = mModel.getColumnCount();

                    for (Row element : mRows) {
                        boolean first = true;

                        for (int j = 0; j < count; j++) {
                            Column column = mModel.getColumnAtIndex(j);

                            if (column.isVisible()) {
                                if (first) {
                                    first = false;
                                } else {
                                    buffer.append('\t');
                                }
                                buffer.append(element.getDataAsText(column));
                            }
                        }
                        buffer.append('\n');
                    }
                }
                mCache = buffer.toString();
            }
            return mCache;
        }
        throw new UnsupportedFlavorException(flavor);
    }
}
