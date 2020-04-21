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

/** Objects that want to be closeable by the {@link CloseCommand} must implement this interface. */
public interface CloseHandler {
    /** @return {@code true} if {@link #attemptClose()} may be called. */
    boolean mayAttemptClose();

    /**
     * Called to try and close the specified object.
     *
     * @return {@code true} if the object has been closed.
     */
    boolean attemptClose();
}
