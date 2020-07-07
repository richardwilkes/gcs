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

import com.trollworks.gcs.GCS;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.Release;
import com.trollworks.gcs.utility.UrlUtils;
import com.trollworks.gcs.utility.Version;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.io.BufferedInputStream;
import java.io.BufferedReader;
import java.io.IOException;
import java.io.OutputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;
import java.util.zip.ZipEntry;
import java.util.zip.ZipInputStream;

public class Library implements Comparable<Library> {
    private static final String        RELEASE_FILE  = "release.txt";
    private static final String        KEY_TITLE     = "title";
    private static final String        KEY_PATH      = "path";
    private static final String        KEY_LAST_SEEN = "last_seen";
    public static final  Library       MASTER        = new Library(I18n.Text("Master Library"), "richardwilkes", "gcs_master_library", getDefaultMasterLibraryPath());
    public static final  Library       USER          = new Library(I18n.Text("User Library"), "*", "gcs_user_library", getDefaultUserLibraryPath());
    public static final  List<Library> LIBRARIES     = new ArrayList<>();
    private              String        mTitle;
    private              String        mGitHubAccountName;
    private              String        mRepoName;
    private              Path          mPath;
    private              Version       mLastSeen;
    private              Release       mAvailableUpgrade;

    public static Path getDefaultMasterLibraryPath() {
        return Paths.get(System.getProperty("user.home", "."), "GCS", "Master Library").normalize().toAbsolutePath();
    }

    public static Path getDefaultUserLibraryPath() {
        return Paths.get(System.getProperty("user.home", "."), "GCS", "User Library").normalize().toAbsolutePath();
    }

    public Library(String title, String githubAccountName, String repoName, Path path) {
        mTitle = title;
        mGitHubAccountName = githubAccountName;
        mRepoName = repoName;
        mPath = path;
        mLastSeen = new Version();
    }

    public static Library fromJSON(String key, JsonMap m) throws IOException {
        String pathStr = m.getString(KEY_PATH);
        if (pathStr.isBlank()) {
            throw new IOException("invalid library path");
        }
        Path path = Paths.get(pathStr);
        if (USER.getKey().equals(key)) {
            USER.mPath = path;
            return USER;
        }
        Version lastSeen = new Version(m.getString(KEY_LAST_SEEN));
        if (MASTER.getKey().equals(key)) {
            MASTER.mPath = path;
            MASTER.mLastSeen = lastSeen;
            return MASTER;
        }
        String[] parts = key.split("/", 2);
        if (parts.length != 2) {
            throw new IOException("invalid library key");
        }
        String title = m.getString(KEY_TITLE);
        if (title.isBlank() || MASTER.getTitle().equalsIgnoreCase(title) || USER.getTitle().equalsIgnoreCase(title)) {
            throw new IOException("invalid library title");
        }
        Library library = new Library(title, parts[0], parts[1], path);
        library.mLastSeen = lastSeen;
        return library;
    }

    public void toJSON(JsonWriter w) throws IOException {
        w.key(getKey());
        w.startMap();
        if (this != MASTER && this != USER) {
            w.keyValue(KEY_TITLE, mTitle);
        }
        w.keyValue(KEY_PATH, mPath.toString());
        if (this != USER) {
            w.keyValue(KEY_LAST_SEEN, mLastSeen.toString());
        }
        w.endMap();
    }

    public String getKey() {
        return mGitHubAccountName + "/" + mRepoName;
    }

    public String getGitHubAccountName() {
        return mGitHubAccountName;
    }

    public String getRepoName() {
        return mRepoName;
    }

    public String getTitle() {
        return mTitle;
    }

    public Path getPathNoCreate() {
        return mPath;
    }

    public Path getPath() {
        if (!Files.exists(mPath)) {
            try {
                Files.createDirectories(mPath);
            } catch (IOException exception) {
                Log.error(exception);
            }
        }
        return mPath;
    }

    public void setPath(Path path) {
        path = path.normalize().toAbsolutePath();
        if (!mPath.equals(path)) {
            mPath = path;
            mLastSeen = getVersionOnDisk();
        }
    }

    public Version getLastSeen() {
        return mLastSeen;
    }

    public List<Release> checkForAvailableUpgrade() {
        Version nextMajorVersion = new Version(GCS.LIBRARY_VERSION);
        if (nextMajorVersion.isZero()) {
            nextMajorVersion.major = 2;
        } else {
            nextMajorVersion.major++;
        }
        return Release.load(mGitHubAccountName, mRepoName, getVersionOnDisk(), (version, notes) -> version.compareTo(GCS.LIBRARY_VERSION) > 0 && version.compareTo(nextMajorVersion) < 0);
    }

    public synchronized Release getAvailableUpgrade() {
        return mAvailableUpgrade;
    }

    public synchronized void setAvailableUpgrade(List<Release> availableUpgrades) {
        if (availableUpgrades == null) {
            mAvailableUpgrade = new Release(null);
            return;
        }
        int count = availableUpgrades.size();
        if (count > 1) {
            if (availableUpgrades.get(count - 1).getVersion().equals(getVersionOnDisk())) {
                availableUpgrades.remove(count - 1);
            }
        }
        mAvailableUpgrade = new Release(availableUpgrades);
        Version version = mAvailableUpgrade.getVersion();
        if (!version.isZero()) {
            mLastSeen = new Version(mAvailableUpgrade.getVersion());
        }
    }

    public Version getVersionOnDisk() {
        Path path = getPath().resolve(RELEASE_FILE);
        if (Files.exists(path)) {
            try (BufferedReader in = Files.newBufferedReader(path)) {
                String line = in.readLine();
                while (line != null) {
                    line = line.trim();
                    if (!line.isEmpty()) {
                        return new Version(line);
                    }
                    line = in.readLine();
                }
            } catch (IOException exception) {
                Log.warn(exception);
            }
        }
        return new Version();
    }

    public void download(Release release) throws IOException {
        Path root = getPath(); // will recreate the dir
        try (ZipInputStream in = new ZipInputStream(new BufferedInputStream(UrlUtils.setupConnection(release.getZipFileURL()).getInputStream()))) {
            byte[]   buffer = new byte[8192];
            ZipEntry entry;
            while ((entry = in.getNextEntry()) != null) {
                if (entry.isDirectory()) {
                    continue;
                }
                Path entryPath = Paths.get(entry.getName());
                int  nameCount = entryPath.getNameCount();
                if (nameCount < 3 || !"Library".equals(entryPath.getName(1).toString())) {
                    continue;
                }
                long size = entry.getSize();
                if (size < 1) {
                    continue;
                }
                Path path = root.resolve(entryPath.subpath(2, nameCount));
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
            Files.writeString(root.resolve(RELEASE_FILE), release.getVersion().toString() + "\n");
        }
    }

    private String getSortKey() {
        if (this == MASTER) {
            return "0." + mTitle;
        }
        if (this == USER) {
            return "1." + mTitle;
        }
        return "2." + mTitle;
    }

    @Override
    public int compareTo(Library other) {
        return NumericComparator.caselessCompareStrings(getSortKey(), other.getSortKey());
    }
}
