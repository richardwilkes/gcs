/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.library;

public class LibraryExplorerSearchResult {
    private LibraryExplorerRow mRow;
    private boolean            mUseFullPath;

    public LibraryExplorerSearchResult(LibraryExplorerRow row) {
        mRow = row;
    }

    public String getTitle() {
        return mUseFullPath ? mRow.getName() + " : " + mRow.getSelectionKey() : mRow.getName();
    }

    public void useFullPath() {
        mUseFullPath = true;
    }

    public LibraryExplorerRow getRow() {
        return mRow;
    }
}
