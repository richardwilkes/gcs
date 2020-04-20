/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.utility;

import java.io.File;

/** Provides a way to find a {@link FileProxy} for a {@link File}. */
public interface FileProxyProvider {
    /**
     * @param file The {@link File} to locate a {@link FileProxy} for.
     * @return The {@link FileProxy}. May be {@code null} if this provider doesn't have one that
     *         represents the specified {@link File}.
     */
    FileProxy getFileProxy(File file);
}
