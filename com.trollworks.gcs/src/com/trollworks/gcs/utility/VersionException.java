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

import java.io.IOException;

/** An exception for data files that are too old or new to be loaded. */
public class VersionException extends IOException {
    /** @return An {@link VersionException} for files that are too old. */
    public static final VersionException createTooOld() {
        return new VersionException(I18n.Text("The file is from a older version and cannot be loaded."));
    }

    /** @return An {@link VersionException} for files that are too new. */
    public static final VersionException createTooNew() {
        return new VersionException(I18n.Text("The file is from an newer version and cannot be loaded."));
    }

    private VersionException(String msg) {
        super(msg);
    }
}
