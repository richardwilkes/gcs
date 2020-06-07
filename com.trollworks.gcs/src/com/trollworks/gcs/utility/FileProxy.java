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

import java.nio.file.Path;

/** Objects that provide a backing file should implement this interface. */
public interface FileProxy {
    /** @return The backing file path. May be {@code null}. */
    Path getBackingFile();

    /**
     * Called to request the UI that displays the file associated with this {@link
     * FileProxy} be brought to the foreground and given focus.
     */
    void toFrontAndFocus();

    /** @return A {@link PrintProxy} for this file, or {@code null}. */
    PrintProxy getPrintProxy();
}
