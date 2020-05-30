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

package com.trollworks.gcs.library;

import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.io.IOException;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

public class LibraryCollector implements Comparator<Object> {
    private List<Object>       mCurrent;
    private List<List<Object>> mStack;
    private Set<Path>          mDirs;

    @SuppressWarnings("unchecked")
    public static List<Object> list(String name, Path root, Set<Path> dirs) {
        LibraryCollector collector = new LibraryCollector();
        try {
            collector.traverse(root);
        } catch (Exception exception) {
            Log.error(exception);
        }
        dirs.addAll(collector.mDirs);
        List<Object> current = collector.mCurrent;
        if (current.isEmpty()) {
            current.add(name);
        } else {
            current = (List<Object>) current.get(0);
            current.set(0, name);
        }
        return current;
    }

    private LibraryCollector() {
        mDirs = new HashSet<>();
        mCurrent = new ArrayList<>();
        mStack = new ArrayList<>();
    }

    private void traverse(Path dir) throws IOException {
        if (!shouldSkip(dir)) {
            mDirs.add(dir.normalize().toAbsolutePath());
            mStack.add(mCurrent);
            mCurrent = new ArrayList<>();
            mCurrent.add(dir.getFileName().toString());
            // IMPORTANT: On Windows, calling any of the older methods to list the contents of a
            // directory results in leaving state around that prevents future move & delete
            // operations. Only use this style of access for directory listings to avoid that.
            try (DirectoryStream<Path> stream = Files.newDirectoryStream(dir)) {
                for (Path path : stream) {
                    if (Files.isDirectory(path)) {
                        traverse(path);
                    } else if (!shouldSkip(path)) {
                        String ext = PathUtils.getExtension(path.getFileName());
                        for (FileType one : FileType.OPENABLE) {
                            if (one.matchExtension(ext)) {
                                mCurrent.add(path);
                                break;
                            }
                        }
                    }
                }
            }
            mCurrent.sort(this);
            List<Object> restoring = mStack.remove(mStack.size() - 1);
            if (mCurrent.size() > 1) {
                restoring.add(mCurrent);
            }
            mCurrent = restoring;
        }
    }

    @Override
    public int compare(Object o1, Object o2) {
        return NumericComparator.compareStrings(getName(o1), getName(o2));
    }

    private static String getName(Object obj) {
        if (obj instanceof Path) {
            return ((Path) obj).getFileName().toString();
        }
        if (obj instanceof List) {
            return ((List<?>) obj).get(0).toString();
        }
        return "";
    }

    private static boolean shouldSkip(Path path) {
        return path.getFileName().toString().startsWith(".");
    }
}
