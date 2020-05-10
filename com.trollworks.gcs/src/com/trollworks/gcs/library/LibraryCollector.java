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

import com.trollworks.gcs.collections.Stack;
import com.trollworks.gcs.io.Log;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.io.IOException;
import java.nio.file.FileVisitResult;
import java.nio.file.FileVisitor;
import java.nio.file.Path;
import java.nio.file.attribute.BasicFileAttributes;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

public class LibraryCollector implements FileVisitor<Path>, Comparator<Object> {
    private List<Object>        mCurrent;
    private Stack<List<Object>> mStack;
    private Set<Path>           mDirs;

    public LibraryCollector() {
        mDirs = new HashSet<>();
        mCurrent = new ArrayList<>();
        mStack = new Stack<>();
    }

    @SuppressWarnings("unchecked")
    public List<Object> getResult(String name) {
        if (mCurrent.isEmpty()) {
            mCurrent.add(name);
        } else {
            mCurrent = (List<Object>) mCurrent.get(0);
            mCurrent.set(0, name);
        }
        return mCurrent;
    }

    public Set<Path> getDirs() {
        return mDirs;
    }

    @Override
    public FileVisitResult preVisitDirectory(Path dir, BasicFileAttributes attrs) throws IOException {
        if (shouldSkip(dir)) {
            return FileVisitResult.SKIP_SUBTREE;
        }
        mDirs.add(dir);
        mStack.push(mCurrent);
        mCurrent = new ArrayList<>();
        mCurrent.add(dir.getFileName().toString());
        return FileVisitResult.CONTINUE;
    }

    @Override
    public FileVisitResult visitFile(Path file, BasicFileAttributes attrs) throws IOException {
        if (!shouldSkip(file)) {
            String ext = PathUtils.getExtension(file.getFileName());
            for (FileType one : FileType.OPENABLE) {
                if (one.matchExtension(ext)) {
                    mCurrent.add(file);
                    break;
                }
            }
        }
        return FileVisitResult.CONTINUE;
    }

    @Override
    public FileVisitResult visitFileFailed(Path file, IOException exception) throws IOException {
        Log.error(exception);
        return FileVisitResult.CONTINUE;
    }

    @Override
    public FileVisitResult postVisitDirectory(Path dir, IOException exception) throws IOException {
        if (exception != null) {
            Log.error(exception);
        }
        mCurrent.sort(this);
        List<Object> restoring = mStack.pop();
        if (mCurrent.size() > 1) {
            restoring.add(mCurrent);
        }
        mCurrent = restoring;
        return FileVisitResult.CONTINUE;
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
