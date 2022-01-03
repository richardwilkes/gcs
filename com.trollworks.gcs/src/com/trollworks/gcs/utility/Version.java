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

package com.trollworks.gcs.utility;

public class Version implements Comparable<Version> {
    public int mMajor;
    public int mMinor;
    public int mBugfix;

    public Version() {
    }

    public Version(int majorVersion, int minorVersion, int bugfixVersion) {
        mMajor = majorVersion;
        mMinor = minorVersion;
        mBugfix = bugfixVersion;
    }

    public Version(Version other) {
        mMajor = other.mMajor;
        mMinor = other.mMinor;
        mBugfix = other.mBugfix;
    }

    public Version(String buffer) {
        extract(buffer);
    }

    public void extract(String buffer) {
        try {
            String[] parts = buffer.split("\\.", 3);
            switch (parts.length) {
                case 3:
                    mBugfix = Integer.parseInt(parts[2]);
                    //noinspection fallthrough
                case 2:
                    mMinor = Integer.parseInt(parts[1]);
                    //noinspection fallthrough
                default:
                    mMajor = Integer.parseInt(parts[0]);
            }
        } catch (NumberFormatException nfe) {
            mMajor = 0;
            mMinor = 0;
            mBugfix = 0;
        }
    }

    public boolean isZero() {
        return mMajor == 0 && mMinor == 0 && mBugfix == 0;
    }

    public String toString() {
        StringBuilder buffer = new StringBuilder();
        buffer.append(mMajor);
        buffer.append('.');
        buffer.append(mMinor);
        if (mBugfix != 0) {
            buffer.append('.');
            buffer.append(mBugfix);
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
        return mMajor == version.mMajor && mMinor == version.mMinor && mBugfix == version.mBugfix;
    }

    @Override
    public int hashCode() {
        int result = 31 + mMajor;
        result *= 31;
        result += mMinor;
        result *= 31;
        result += mBugfix;
        return result;
    }

    @Override
    public int compareTo(Version other) {
        int result = Integer.compare(mMajor, other.mMajor);
        if (result == 0) {
            result = Integer.compare(mMinor, other.mMinor);
            if (result == 0) {
                result = Integer.compare(mBugfix, other.mBugfix);
            }
        }
        return result;
    }
}
