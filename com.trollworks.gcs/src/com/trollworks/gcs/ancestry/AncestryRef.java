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

package com.trollworks.gcs.ancestry;

import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.NamedData;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.io.IOException;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class AncestryRef implements Comparable<AncestryRef> {
    private static final Map<String, AncestryRef> REGISTERED_ANCESTRIES = new HashMap<>();
    public static final  AncestryRef              DEFAULT               = new AncestryRef();
    private              String                   mName;
    private              Ancestry                 mAncestry;

    public static final AncestryRef get(String name) {
        load();
        AncestryRef ref = REGISTERED_ANCESTRIES.get(name);
        return ref != null ? ref : DEFAULT;
    }

    public static final List<AncestryRef> choices() {
        load();
        List<AncestryRef> list = new ArrayList<>(REGISTERED_ANCESTRIES.values());
        Collections.sort(list);
        return list;
    }

    private static void load() {
        if (REGISTERED_ANCESTRIES.isEmpty()) {
            for (NamedData<List<NamedData<AncestryRef>>> list : NamedData.scanLibraries(FileType.ANCESTRY_SETTINGS, Dirs.SETTINGS, AncestryRef::new)) {
                for (NamedData<AncestryRef> data : list.getData()) {
                    REGISTERED_ANCESTRIES.putIfAbsent(data.getName(), data.getData());
                }
            }
            REGISTERED_ANCESTRIES.putIfAbsent(DEFAULT.mName, DEFAULT);
        }
    }

    public AncestryRef() {
        mName = "Human";
        mAncestry = new Ancestry();
    }

    public AncestryRef(Path path) throws IOException {
        mName = PathUtils.getLeafName(path, false);
        mAncestry = new Ancestry(path);
    }

    public AncestryRef(String name) {
        mName = name;
    }

    public String name() {
        return mName;
    }

    public Ancestry ancestry() {
        return mAncestry;
    }

    @Override
    public int compareTo(AncestryRef other) {
        return NumericComparator.caselessCompareStrings(mName, other.mName);
    }

    @Override
    public String toString() {
        return mName;
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }
        AncestryRef that = (AncestryRef) other;
        if (!mName.equals(that.mName)) {
            return false;
        }
        return mAncestry.equals(that.mAncestry);
    }

    @Override
    public int hashCode() {
        int result = mName.hashCode();
        result = 31 * result + mAncestry.hashCode();
        return result;
    }
}
