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

package com.trollworks.gcs.settings;

import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.EnumMap;
import java.util.Map;

public enum Dirs {
    ATTRIBUTES {
        @Override
        public Path getDefaultPath() {
            return Library.getDefaultUserLibraryPath().resolve("Attributes");
        }
    },
    GENERAL {
        @Override
        public Path getDefaultPath() {
            return Paths.get(System.getProperty("user.home", "."));
        }
    },
    HIT_LOCATIONS {
        @Override
        public Path getDefaultPath() {
            return Library.getDefaultUserLibraryPath().resolve("Hit Locations");
        }
    },
    PDF {
        @Override
        public Path getDefaultPath() {
            return Paths.get(System.getProperty("user.home", "."));
        }
    },
    THEME {
        @Override
        public Path getDefaultPath() {
            return Library.getDefaultUserLibraryPath().resolve("Theme");
        }
    };

    private static final Map<Dirs, Path> LAST_DIRS = new EnumMap<>(Dirs.class);

    public abstract Path getDefaultPath();

    public final Path get() {
        synchronized (LAST_DIRS) {
            Path path = LAST_DIRS.get(this);
            if (path == null || !Files.isDirectory(path)) {
                path = getDefaultPath().normalize().toAbsolutePath();
                try {
                    Files.createDirectories(path);
                    LAST_DIRS.put(this, path);
                } catch (IOException exception) {
                    Log.error(exception);
                }
            }
            return path;
        }
    }

    public final void set(Path path) {
        synchronized (LAST_DIRS) {
            LAST_DIRS.put(this, path.normalize().toAbsolutePath());
        }
    }

    public static final void load(JsonMap m) {
        synchronized (LAST_DIRS) {
            LAST_DIRS.clear();
            for (Dirs one : values()) {
                String p = m.getString(Enums.toId(one));
                if (p != null && !p.isBlank()) {
                    LAST_DIRS.put(one, Paths.get(p));
                }
            }
        }
    }

    public static final void save(String key, JsonWriter w) throws IOException {
        synchronized (LAST_DIRS) {
            if (!LAST_DIRS.isEmpty()) {
                w.key(key);
                w.startMap();
                for (Dirs one : values()) {
                    Path path = LAST_DIRS.get(one);
                    if (path != null) {
                        w.keyValue(Enums.toId(one), path.toString());
                    }
                }
                w.endMap();
            }
        }
    }
}
