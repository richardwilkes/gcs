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

package com.trollworks.gcs.pdfview;

import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;

/** Tracks data for opening and navigating PDFs. */
public class PdfRef implements Comparable<PdfRef> {
    private static final String ID     = "id";
    private static final String PATH   = "path";
    private static final String OFFSET = "offset";
    private              String mId;
    private              Path   mPath;
    private              int    mPageToIndexOffset;

    /**
     * Creates a new {@link PdfRef}.
     *
     * @param id     The id to use. Pass in {@code null} or an empty string to create a {@link
     *               PdfRef} that won't update preferences.
     * @param path   The path that the {@code id} refers to.
     * @param offset The amount to add to a symbolic page number to find the actual index.
     */
    public PdfRef(String id, Path path, int offset) {
        mId = id == null ? "" : id;
        mPath = path;
        mPageToIndexOffset = offset;
    }

    public PdfRef(JsonMap m) {
        mId = m.getString(ID);
        mPath = Paths.get(m.getStringWithDefault(PATH, ".")).normalize().toAbsolutePath();
        mPageToIndexOffset = m.getInt(OFFSET);
    }

    /** @return The id. */
    public String getId() {
        return mId;
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
    public int compareTo(PdfRef other) {
        return NumericComparator.caselessCompareStrings(mId, other.mId);
    }

    public void toJSON(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(ID, mId);
        w.keyValue(PATH, mPath.toString());
        w.keyValue(OFFSET, mPageToIndexOffset);
        w.endMap();
    }
}
