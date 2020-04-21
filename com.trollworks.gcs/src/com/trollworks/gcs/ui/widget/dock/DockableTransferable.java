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

package com.trollworks.gcs.ui.widget.dock;

import java.awt.datatransfer.DataFlavor;
import java.awt.datatransfer.Transferable;
import java.awt.datatransfer.UnsupportedFlavorException;

/** Allows {@link Dockable}s to be part of drag and drop operations internal to the JVM. */
public class DockableTransferable implements Transferable {
    /** The data flavor for this class. */
    public static final DataFlavor DATA_FLAVOR = new DataFlavor(DockableTransferable.class, "Dockable");
    private             Dockable   mDockable;

    public DockableTransferable(Dockable dockable) {
        mDockable = dockable;
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
            return mDockable;
        }
        throw new UnsupportedFlavorException(flavor);
    }
}
