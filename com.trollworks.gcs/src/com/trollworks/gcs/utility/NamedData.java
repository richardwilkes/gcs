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

package com.trollworks.gcs.utility;

import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.io.IOException;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

public class NamedData<T> implements Comparable<NamedData<T>> {
    private String mName;
    private T      mData;

    public interface DataLoader<T> {
        T loadData(Path path) throws IOException;
    }

    public static <T> List<NamedData<List<NamedData<T>>>> scanLibraries(FileType fileType, Dirs lastDir, DataLoader<T> loader) {
        List<NamedData<List<NamedData<T>>>> all = new ArrayList<>();
        Settings.getInstance(); // Just to ensure the libraries list is initialized
        for (Library lib : Library.LIBRARIES) {
            Path libPath = lib.getPath();
            collect(libPath.resolve(lastDir.getDefaultPath().getFileName()), fileType, loader, all);
            Path legacy = lastDir.getLegacyDefaultPath(fileType);
            if (legacy != null) {
                collect(libPath.resolve(legacy.getFileName()), fileType, loader, all);
            }
        }
        return all;
    }

    private static <T> void collect(Path dir, FileType fileType, DataLoader<T> loader, List<NamedData<List<NamedData<T>>>> all) {
        if (Files.isDirectory(dir)) {
            List<NamedData<T>> list = new ArrayList<>();
            // IMPORTANT: On Windows, calling any of the older methods to list the contents of a
            // directory results in leaving state around that prevents future move & delete
            // operations. Only use this style of access for directory listings to avoid that.
            try (DirectoryStream<Path> stream = Files.newDirectoryStream(dir)) {
                for (Path path : stream) {
                    if (fileType.matchExtension(PathUtils.getExtension(path))) {
                        try {
                            list.add(new NamedData<>(PathUtils.getLeafName(path, false),
                                    loader.loadData(path)));
                        } catch (IOException ioe) {
                            Log.error("unable to load " + path, ioe);
                        }
                    }
                }
            } catch (IOException ioe) {
                Log.error("failed directory scan of " + dir, ioe);
            }
            if (!list.isEmpty()) {
                Collections.sort(list);
                all.add(new NamedData<>(dir.getParent().getFileName().resolve(dir.getFileName()).toString(), list));
            }
        }
    }

    public NamedData(String name, T data) {
        mName = name;
        mData = data;
    }

    public final String getName() {
        return mName;
    }

    public final T getData() {
        return mData;
    }

    @Override
    public final int compareTo(NamedData other) {
        return NumericComparator.CASELESS_COMPARATOR.compare(mName, other.mName);
    }

    @Override
    public final String toString() {
        return mName;
    }
}
