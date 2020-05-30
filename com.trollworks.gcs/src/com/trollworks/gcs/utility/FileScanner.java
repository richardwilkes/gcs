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
import java.nio.file.FileVisitOption;
import java.nio.file.FileVisitResult;
import java.nio.file.FileVisitor;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.attribute.BasicFileAttributes;
import java.util.EnumSet;

/** Walks a file tree, calling your {@link Handler} for each file found. */
public class FileScanner implements FileVisitor<Path> {
    private Path    mPath;
    private Handler mHandler;
    private boolean mSkipHidden;

    /**
     * Walks a file tree, calling the specified {@link Handler} for each file found. Hidden files
     * and directories (those whose names start with a period) are skipped.
     *
     * @param path    The starting point.
     * @param handler The {@link Handler} to call for each file.
     */
    public static final void walk(Path path, Handler handler) {
        walk(path, handler, true);
    }

    /**
     * Walks a file tree, calling the specified {@link Handler} for each file found. Hidden files
     * and directories (those that start with a period) are skipped.
     *
     * @param path       The starting point.
     * @param handler    The {@link Handler} to call for each file.
     * @param skipHidden Pass in {@code true} if files and directories whose names start with a
     *                   period should be skipped.
     */
    public static final void walk(Path path, Handler handler, boolean skipHidden) {
        try {
            Files.walkFileTree(path, EnumSet.of(FileVisitOption.FOLLOW_LINKS), Integer.MAX_VALUE, new FileScanner(path, handler, skipHidden));
        } catch (Exception exception) {
            Log.error(exception);
        }
    }

    private FileScanner(Path path, Handler handler, boolean skipHidden) {
        mPath = path;
        mHandler = handler;
        mSkipHidden = skipHidden;
    }

    private boolean shouldSkip(Path path) {
        return mSkipHidden && !mPath.equals(path) && path.getFileName().toString().startsWith(".");
    }

    @Override
    public FileVisitResult preVisitDirectory(Path path, BasicFileAttributes attrs) throws IOException {
        if (shouldSkip(path)) {
            return FileVisitResult.SKIP_SUBTREE;
        }
        return FileVisitResult.CONTINUE;
    }

    @Override
    public FileVisitResult visitFile(Path path, BasicFileAttributes attrs) throws IOException {
        if (!shouldSkip(path)) {
            mHandler.processFile(path);
        }
        return FileVisitResult.CONTINUE;
    }

    @Override
    public FileVisitResult visitFileFailed(Path path, IOException exception) throws IOException {
        Log.error(exception);
        return FileVisitResult.CONTINUE;
    }

    @Override
    public FileVisitResult postVisitDirectory(Path path, IOException exception) throws IOException {
        if (exception != null) {
            Log.error(exception);
        }
        return FileVisitResult.CONTINUE;
    }

    /** The callback used for {@link FileScanner}. */
    public interface Handler {
        /** @param path The {@link Path} to the file to be processed. */
        void processFile(Path path) throws IOException;
    }
}
