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
import java.nio.file.FileVisitResult;
import java.nio.file.FileVisitor;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.attribute.BasicFileAttributes;

/** Provides a simple way to remove all files in a directory tree. */
public class RecursiveDirectoryRemover implements FileVisitor<Path> {
    private Path mPath;

    /**
     * @param path           The starting point.
     * @param includeRootDir Pass in {@code true} to remove the specified path as well as its
     *                       contents.
     */
    public static final void remove(Path path, boolean includeRootDir) {
        try {
            Files.walkFileTree(path, new RecursiveDirectoryRemover(includeRootDir ? null : path));
        } catch (IOException exception) {
            Log.error(exception);
        }
    }

    private RecursiveDirectoryRemover(Path exclude) {
        mPath = exclude;
    }

    @Override
    public FileVisitResult preVisitDirectory(Path dir, BasicFileAttributes attrs) throws IOException {
        return FileVisitResult.CONTINUE;
    }

    @Override
    public FileVisitResult visitFile(Path file, BasicFileAttributes attrs) throws IOException {
        Files.delete(file);
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
        if (!dir.equals(mPath)) {
            Files.delete(dir);
        }
        return FileVisitResult.CONTINUE;
    }
}
