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

package com.trollworks.gcs.attribute;

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.VersionException;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.io.BufferedReader;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;

public class AttributeSet implements Comparable<AttributeSet> {
    private static final String                    KEY_NAME       = "name";
    private static final String                    JSON_TYPE_NAME = "attribute_settings";
    private              String                    mName;
    private              Map<String, AttributeDef> mAttributes;

    public AttributeSet(String name, Map<String, AttributeDef> attributes) {
        mName = name;
        mAttributes = attributes;
    }

    public AttributeSet(Path path) throws IOException {
        try (BufferedReader fileReader = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            JsonMap   m     = Json.asMap(Json.parse(fileReader));
            LoadState state = new LoadState();
            state.mDataFileVersion = m.getInt(DataFile.VERSION);
            if (state.mDataFileVersion > DataFile.CURRENT_VERSION) {
                throw VersionException.createTooNew();
            }
            if (!JSON_TYPE_NAME.equals(m.getString(DataFile.TYPE))) {
                throw new IOException("invalid data type");
            }
            mName = m.getString(KEY_NAME);
            if (mName.isBlank()) {
                // For legacy files that didn't have the name key
                mName = PathUtils.getLeafName(path, false);
            }
            mAttributes = AttributeDef.load(m.getArray(JSON_TYPE_NAME));
        }
    }

    public void toJSON(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(DataFile.TYPE, JSON_TYPE_NAME);
        w.keyValue(DataFile.VERSION, DataFile.CURRENT_VERSION);
        w.keyValue(KEY_NAME, mName);
        w.key(JSON_TYPE_NAME);
        AttributeDef.writeOrdered(w, mAttributes);
        w.endMap();
    }

    public Map<String, AttributeDef> getAttributes() {
        return mAttributes;
    }

    @Override
    public int compareTo(AttributeSet other) {
        return NumericComparator.caselessCompareStrings(mName, other.mName);
    }

    @Override
    public String toString() {
        return mName;
    }

    public static List<AttributeSet> get() {
        List<AttributeSet> list = new ArrayList<>();
        Preferences.getInstance(); // Just to ensure the libraries list is initialized
        for (Library lib : Library.LIBRARIES) {
            Path dir = lib.getPath().resolve("Attributes");
            if (Files.isDirectory(dir)) {
                // IMPORTANT: On Windows, calling any of the older methods to list the contents of a
                // directory results in leaving state around that prevents future move & delete
                // operations. Only use this style of access for directory listings to avoid that.
                try (DirectoryStream<Path> stream = Files.newDirectoryStream(dir)) {
                    for (Path path : stream) {
                        try {
                            list.add(new AttributeSet(path));
                        } catch (IOException ioe) {
                            Log.error("unable to load " + path, ioe);
                        }
                    }
                } catch (IOException exception) {
                    Log.error(exception);
                }
            }
        }
        Collections.sort(list);
        return list;
    }
}
