/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.library;

import com.trollworks.gcs.io.Log;
import com.trollworks.gcs.io.RecursiveDirectoryRemover;
import com.trollworks.gcs.io.UrlUtils;
import com.trollworks.gcs.io.json.Json;
import com.trollworks.gcs.io.json.JsonArray;
import com.trollworks.gcs.io.json.JsonMap;
import com.trollworks.gcs.utility.Version;

import java.io.BufferedInputStream;
import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.zip.ZipEntry;
import java.util.zip.ZipInputStream;

public class Library {
    private static final String SHA_PREFIX   = "\"sha\": \"";
    private static final String SHA_SUFFIX   = "\",";
    private static final String ROOT_PREFIX  = "richardwilkes-gcs_library-";
    private static final String VERSION_FILE = "version.txt";

    /** @return The path to the system GCS library files. */
    public static Path getSystemRootPath() {
        Path   path;
        String library = System.getenv("GCS_LIBRARY");
        if (library != null) {
            path = Paths.get(library);
        } else {
            path = Paths.get(System.getProperty("user.home", "."), "GCS", "Library");
        }
        path = path.normalize();
        if (!Files.exists(path)) {
            try {
                Files.createDirectories(path);
            } catch (IOException exception) {
                Log.error(exception);
            }
        }
        return path;
    }

    /** @return The path to the user GCS library files. */
    public static Path getUserRootPath() {
        Path   path;
        String library = System.getenv("GCS_USER_LIBRARY");
        if (library != null) {
            path = Paths.get(library);
        } else {
            path = Paths.get(System.getProperty("user.home", "."), "GCS", "User Library");
        }
        path = path.normalize();
        if (!Files.exists(path)) {
            try {
                Files.createDirectories(path);
            } catch (IOException exception) {
                Log.error(exception);
            }
        }
        return path;
    }

    public static final String getRecordedCommit() {
        try (BufferedReader in = Files.newBufferedReader(getSystemRootPath().resolve(VERSION_FILE))) {
            String line = in.readLine();
            while (line != null) {
                line = line.trim();
                if (!line.isEmpty()) {
                    return line;
                }
                line = in.readLine();
            }
        } catch (IOException exception) {
            // Ignore
        }
        return "";
    }

    public static final String getLatestCommit() {
        String sha = "";
        try {
            JsonArray array = Json.asArray(Json.parse(new URL("https://api.github.com/repos/richardwilkes/gcs_library/commits?per_page=1")), false);
            JsonMap   map   = array.getMap(0, false);
            sha = map.getString("sha", false);
            if (sha.length() > 7) {
                sha = sha.substring(0, 7);
            }
        } catch (IOException exception) {
            Log.error(exception);
        }
        return sha;
    }

    public static final long getMinimumGCSVersion() {
        String version = "";
        try (BufferedReader in = new BufferedReader(new InputStreamReader(UrlUtils.setupConnection("https://raw.githubusercontent.com/richardwilkes/gcs_library/master/minimum_version.txt").getInputStream(), StandardCharsets.UTF_8))) {
            String line;
            while ((line = in.readLine()) != null) {
                if (version.isBlank()) {
                    line = line.trim();
                    if (!line.isBlank()) {
                        version = line;
                    }
                }
            }
        } catch (IOException exception) {
            Log.error(exception);
        }
        return Version.extract(version, 0);
    }

    public static final boolean download() {
        Path root = getSystemRootPath();
        try (ZipInputStream in = new ZipInputStream(new BufferedInputStream(UrlUtils.setupConnection("https://api.github.com/repos/richardwilkes/gcs_library/zipball/master").getInputStream()))) {
            RecursiveDirectoryRemover.remove(root, false);
            byte[]   buffer = new byte[8192];
            ZipEntry entry;
            String   sha    = "unknown";
            while ((entry = in.getNextEntry()) != null) {
                if (entry.isDirectory()) {
                    continue;
                }
                Path entryPath = Paths.get(entry.getName());
                int  nameCount = entryPath.getNameCount();
                if (nameCount < 3 || !entryPath.getName(0).toString().startsWith(ROOT_PREFIX) || !"Library".equals(entryPath.getName(1).toString())) {
                    continue;
                }
                long size = entry.getSize();
                if (size < 1) {
                    continue;
                }
                sha = entryPath.getName(0).toString().substring(ROOT_PREFIX.length());
                entryPath = entryPath.subpath(2, nameCount);
                Path path = root.resolve(entryPath);
                Files.createDirectories(path.getParent());
                try (OutputStream out = Files.newOutputStream(path)) {
                    while (size > 0) {
                        int amt = in.read(buffer);
                        if (amt < 0) {
                            break;
                        }
                        if (amt > 0) {
                            size -= amt;
                            out.write(buffer, 0, amt);
                        }
                    }
                }
            }
            if (sha.length() > 7) {
                sha = sha.substring(0, 7);
            }
            Files.writeString(root.resolve(VERSION_FILE), sha + "\n");
            return true;
        } catch (IOException exception) {
            return false;
        }
    }
}
