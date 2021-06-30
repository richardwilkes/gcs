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

package com.trollworks.gcs.library;

import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.FontIcon;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.Row;

import javax.swing.Icon;

/** A {@link Row} that represents a directory in the library explorer. */
public class LibraryDirectoryRow extends Row implements LibraryExplorerRow {
    private String mName;

    /** @param name The name of the directory. */
    public LibraryDirectoryRow(String name) {
        mName = name;
        setCanHaveChildren(true);
    }

    @Override
    public String getSelectionKey() {
        Row parent = getParent();
        return parent instanceof LibraryDirectoryRow ?
                ((LibraryDirectoryRow) parent).getSelectionKey() + "/" + mName : mName;
    }

    @Override
    public Icon getIcon() {
        return new FontIcon(FontAwesome.FOLDER, Fonts.FONT_ICON_LABEL_PRIMARY);
    }

    @Override
    public Icon getIcon(Column column) {
        return getIcon();
    }

    @Override
    public String getName() {
        return mName;
    }

    @Override
    public Object getData(Column column) {
        return mName;
    }

    @Override
    public String getDataAsText(Column column) {
        return mName;
    }

    @Override
    public void setData(Column column, Object data) {
        // Unused
    }
}
