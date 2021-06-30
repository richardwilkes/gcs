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

import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.PathUtils;

import java.nio.file.Path;
import javax.swing.Icon;

/** A {@link Row} that represents a file in the library explorer. */
public class LibraryFileRow extends Row implements LibraryExplorerRow {
    private Path mFilePath;

    /** @param filePath A {@link Path} to library file. */
    public LibraryFileRow(Path filePath) {
        mFilePath = filePath;
    }

    /** @return The {@link Path}. */
    public Path getFilePath() {
        return mFilePath;
    }

    @Override
    public String getSelectionKey() {
        return mFilePath.toString();
    }

    @Override
    public Icon getIcon() {
        return FileType.getIconForFileName(mFilePath.getFileName().toString());
    }

    @Override
    public Icon getIcon(Column column) {
        return getIcon();
    }

    @Override
    public String getName() {
        return PathUtils.getLeafName(mFilePath.getFileName(), false);
    }

    @Override
    public Object getData(Column column) {
        return mFilePath;
    }

    @Override
    public String getDataAsText(Column column) {
        return getName();
    }

    @Override
    public void setData(Column column, Object data) {
        // Unused
    }
}
