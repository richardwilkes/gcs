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

import java.util.Objects;

public class Version implements Comparable<Version> {
    public int major;
    public int minor;
    public int bugfix;

    public Version() {
    }

    public Version(int majorVersion, int minorVersion, int bugfixVersion) {
        major = majorVersion;
        minor = minorVersion;
        bugfix = bugfixVersion;
    }

    public Version(Version other) {
        major = other.major;
        minor = other.minor;
        bugfix = other.bugfix;
    }

    public Version(String buffer) {
        extract(buffer);
    }

    public void extract(String buffer) {
        try {
            String[] parts = buffer.split("\\.", 3);
            switch (parts.length) {
            case 3:
                bugfix = Integer.parseInt(parts[2]);
                //noinspection fallthrough
            case 2:
                minor = Integer.parseInt(parts[1]);
                //noinspection fallthrough
            default:
                major = Integer.parseInt(parts[0]);
            }
        } catch (NumberFormatException nfe) {
            major = 0;
            minor = 0;
            bugfix = 0;
        }
    }

    public boolean isZero() {
        return major == 0 && minor == 0 && bugfix == 0;
    }

    public String toString() {
        StringBuilder buffer = new StringBuilder();
        buffer.append(major);
        buffer.append('.');
        buffer.append(minor);
        if (bugfix != 0) {
            buffer.append('.');
            buffer.append(bugfix);
        }
        return buffer.toString();
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }
        Version version = (Version) other;
        return major == version.major && minor == version.minor && bugfix == version.bugfix;
    }

    @Override
    public int hashCode() {
        return Objects.hash(Integer.valueOf(major), Integer.valueOf(minor), Integer.valueOf(bugfix));
    }

    @Override
    public int compareTo(Version other) {
        int result = Integer.compare(major, other.major);
        if (result == 0) {
            result = Integer.compare(minor, other.minor);
            if (result == 0) {
                result = Integer.compare(bugfix, other.bugfix);
            }
        }
        return result;
    }
}
