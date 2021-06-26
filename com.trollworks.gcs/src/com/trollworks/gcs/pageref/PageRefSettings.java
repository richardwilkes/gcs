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

package com.trollworks.gcs.pageref;

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.utility.SafeFileUpdater;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class PageRefSettings {
    private static final String KEY_PATH   = "path";
    private static final String KEY_OFFSET = "offset";

    private Map<String, PageRef> mRefs;

    public PageRefSettings() {
        mRefs = new HashMap<>();
    }

    public PageRefSettings(Path path) throws IOException {
        this();
        try (BufferedReader in = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            JsonMap m = Json.asMap(Json.parse(in));
            if (!m.isEmpty()) {
                int version = m.getInt(Settings.VERSION);
                if (version >= Settings.MINIMUM_VERSION && version <= DataFile.CURRENT_VERSION) {
                    load(m.getMap(Settings.PAGE_REFS));
                }
            }
        }
    }

    public PageRefSettings(JsonMap m) {
        this();
        load(m);
    }

    public void copyFrom(PageRefSettings other) {
        mRefs.clear();
        for (PageRef ref : other.mRefs.values()) {
            mRefs.put(ref.getID(), new PageRef(ref));
        }
    }

    private void load(JsonMap m) {
        for (String key : m.keySet()) {
            JsonMap m2 = m.getMap(key);
            mRefs.put(key, new PageRef(key, Path.of(m2.getString(KEY_PATH)), m2.getInt(KEY_OFFSET)));
        }
    }

    public void save(Path path) throws IOException {
        SafeFileUpdater trans = new SafeFileUpdater();
        trans.begin();
        try {
            Files.createDirectories(path.getParent());
            File file = trans.getTransactionFile(path.toFile());
            try (JsonWriter w = new JsonWriter(new BufferedWriter(new FileWriter(file, StandardCharsets.UTF_8)), "\t")) {
                w.startMap();
                w.keyValue(Settings.VERSION, DataFile.CURRENT_VERSION);
                w.key(Settings.PAGE_REFS);
                save(w);
                w.endMap();
            }
        } catch (IOException ioe) {
            trans.abort();
            throw ioe;
        }
        trans.commit();
    }

    public void save(JsonWriter w) throws IOException {
        w.startMap();
        for (Map.Entry<String, PageRef> entry : mRefs.entrySet()) {
            w.key(entry.getKey());
            PageRef ref = entry.getValue();
            w.startMap();
            w.keyValue(KEY_PATH, ref.getPath().toString());
            w.keyValueNot(KEY_OFFSET, ref.getPageToIndexOffset(), 0);
            w.endMap();
        }
        w.endMap();
    }

    public boolean isEmpty() {
        return mRefs.isEmpty();
    }

    public List<PageRef> list() {
        List<PageRef> list = new ArrayList<>(mRefs.values());
        Collections.sort(list);
        return list;
    }

    public void put(PageRef ref) {
        mRefs.put(ref.getID(), ref);
    }

    public void remove(PageRef ref) {
        mRefs.remove(ref.getID());
    }

    public PageRef lookup(String id, boolean requireFileExists) {
        PageRef ref = mRefs.get(id);
        if (ref == null) {
            return null;
        }
        if (requireFileExists) {
            Path path = ref.getPath();
            if (!Files.isReadable(path) || !Files.isRegularFile(path)) {
                return null;
            }
        }
        return ref;
    }
}
