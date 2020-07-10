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

import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;

import java.net.URL;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

public class Release implements Comparable<Release> {
    private Version mVersion;
    private String  mNotes;
    private String  mZipFileURL;
    private boolean mUnableToAccessRepo;

    public interface ReleaseFilter {
        boolean isReleaseUsable(Version version, String notes);
    }

    public static List<Release> load(String githubAccountName, String repoName, Version currentVersion, ReleaseFilter filter) {
        List<Release> versions = new ArrayList<>();
        try {
            JsonArray list  = Json.asArray(Json.parse(new URL("https://api.github.com/repos/" + githubAccountName + "/" + repoName + "/releases")));
            int       count = list.size();
            for (int i = 0; i < count; i++) {
                JsonMap m   = list.getMap(i);
                String  tag = m.getString("tag_name");
                if (tag.startsWith("v")) {
                    Version version = new Version(tag.substring(1));
                    if (!version.isZero() && version.compareTo(currentVersion) >= 0) {
                        String notes = m.getString("body");
                        if (filter == null || filter.isReleaseUsable(version, notes)) {
                            versions.add(new Release(version, notes, m.getString("zipball_url")));
                        }
                    }
                }
            }
        } catch (Exception exception) {
            Log.error(exception);
            return null;
        }
        Collections.sort(versions);
        return versions;
    }

    public Release() {
        mVersion = new Version();
    }

    public Release(Version version, String notes, String zipFileURL) {
        mVersion = version;
        mNotes = notes;
        mZipFileURL = zipFileURL;
    }

    public Release(List<Release> releases) {
        if (releases == null) {
            mVersion = new Version();
            mUnableToAccessRepo = true;
            return;
        }
        switch (releases.size()) {
        case 0:
            mVersion = new Version();
            return;
        case 1:
            Release other = releases.get(0);
            mVersion = other.mVersion;
            mNotes = other.mNotes;
            mZipFileURL = other.mZipFileURL;
            break;
        default:
            Release other2 = releases.get(0);
            mVersion = other2.mVersion;
            mZipFileURL = other2.mZipFileURL;
            StringBuilder buffer = new StringBuilder();
            for (Release one : releases) {
                if (mVersion != one.mVersion) {
                    buffer.append("\n\n");
                }
                buffer.append("## Version ");
                buffer.append(one.mVersion);
                buffer.append("\n");
                buffer.append(one.mNotes);
            }
            mNotes = buffer.toString();
            break;
        }
    }

    public boolean hasUpdate() {
        return !mVersion.isZero();
    }

    public Version getVersion() {
        return mVersion;
    }

    public String getNotes() {
        return mNotes;
    }

    public String getZipFileURL() {
        return mZipFileURL;
    }

    public boolean unableToAccessRepo() {
        return mUnableToAccessRepo;
    }

    @Override
    public int compareTo(Release other) {
        return other.mVersion.compareTo(mVersion);
    }
}
