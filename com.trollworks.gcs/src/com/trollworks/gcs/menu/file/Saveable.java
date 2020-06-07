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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.ui.widget.DataModifiedListener;
import com.trollworks.gcs.utility.FileProxy;
import com.trollworks.gcs.utility.FileType;

import java.nio.file.Path;

/**
 * Windows that want to participate in the standard {@link SaveCommand} and {@link SaveAsCommand}
 * processing must implement this interface.
 */
public interface Saveable extends FileProxy {
    /** @return Whether the changes have been made that could be saved. */
    boolean isModified();

    /** @param listener The listener to add. */
    void addDataModifiedListener(DataModifiedListener listener);

    /** @param listener The listener to remove. */
    void removeDataModifiedListener(DataModifiedListener listener);

    /** @return The {@link FileType}. */
    FileType getFileType();

    /** @return The name the user will recognize as the name of the object being saved. */
    String getSaveTitle();

    /**
     * Called to actually save the contents to a path.
     *
     * @param path The path to save to.
     * @return The path(s) actually written to.
     */
    Path[] saveTo(Path path);
}
