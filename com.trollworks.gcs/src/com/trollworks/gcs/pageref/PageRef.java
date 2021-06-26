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

package com.trollworks.gcs.pageref;

import com.trollworks.gcs.utility.text.NumericComparator;

import java.nio.file.Path;

/** Tracks data for opening and navigating page references. */
public class PageRef implements Comparable<PageRef> {
    private String mID;
    private Path   mPath;
    private int    mPageToIndexOffset;

    /**
     * Creates a new PageRef.
     *
     * @param id     The id to use. Pass in {@code null} or an empty string to create a {@link
     *               PageRef} that won't update preferences.
     * @param path   The path that the {@code id} refers to.
     * @param offset The amount to add to a symbolic page number to find the actual index.
     */
    public PageRef(String id, Path path, int offset) {
        mID = id == null ? "" : id;
        mPath = path;
        mPageToIndexOffset = offset;
    }

    public PageRef(PageRef other) {
        mID = other.mID;
        mPath = other.mPath;
        mPageToIndexOffset = other.mPageToIndexOffset;
    }

    /** @return The id. */
    public String getID() {
        return mID;
    }

    /** @return The path. */
    public Path getPath() {
        return mPath;
    }

    /** @return The amount to add to a symbolic page number to find the actual index. */
    public int getPageToIndexOffset() {
        return mPageToIndexOffset;
    }

    /** @param offset The amount to add to a symbolic page number to find the actual index. */
    public void setPageToIndexOffset(int offset) {
        mPageToIndexOffset = offset;
    }

    @Override
    public int compareTo(PageRef other) {
        return NumericComparator.caselessCompareStrings(mID, other.mID);
    }
}
