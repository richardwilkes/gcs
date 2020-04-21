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

/** Defines constants for each platform we support. */
public enum Platform {
    LINUX, MAC, WINDOWS, UNKNOWN;

    private static final Platform CURRENT;

    static {
        String osName = System.getProperty("os.name");
        if (osName.startsWith("Mac")) {
            CURRENT = MAC;
        } else if (osName.startsWith("Win")) {
            CURRENT = WINDOWS;
        } else if (osName.startsWith("Linux")) {
            CURRENT = LINUX;
        } else {
            CURRENT = UNKNOWN;
        }
    }

    /** @return The platform being run on. */
    public static final Platform getPlatform() {
        return CURRENT;
    }

    /** @return {@code true} if Macintosh is the platform being run on. */
    public static final boolean isMacintosh() {
        return CURRENT == MAC;
    }

    /** @return {@code true} if Windows is the platform being run on. */
    public static final boolean isWindows() {
        return CURRENT == WINDOWS;
    }

    /** @return {@code true} if Linux is the platform being run on. */
    public static final boolean isLinux() {
        return CURRENT == LINUX;
    }
}
