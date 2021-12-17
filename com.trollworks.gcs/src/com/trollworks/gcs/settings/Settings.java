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

import com.trollworks.gcs.GCS;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.ChangeableData;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.pageref.PageRefSettings;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.SafeFileUpdater;
import com.trollworks.gcs.utility.Version;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonArray;
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

public final class Settings extends ChangeableData {
    public static final int MINIMUM_VERSION = 1;

    private static final String DEPRECATED_PDF_REFS = "pdf_refs"; // June 25, 2021

    public static final  String COLORS                = "colors";
    private static final String DIVIDER_POSITION      = "divider_position";
    public static final  String FONTS                 = "fonts";
    public static final  String GENERAL               = "general";
    public static final  String KEY_BINDINGS          = "key_bindings";
    private static final String LAST_DIRS             = "last_dirs";
    private static final String LAST_SEEN_GCS_VERSION = "last_seen_gcs_version";
    private static final String LIBRARIES             = "libraries";
    private static final String LIBRARY_EXPLORER      = "library_explorer";
    private static final String OPEN_ROW_KEYS         = "open_row_keys";
    public static final  String PAGE_REFS             = "page_refs";
    private static final String QUICK_EXPORTS         = "quick_exports";
    private static final String RECENT_FILES          = "recent_files";
    public static final  String SHEET_SETTINGS        = "sheet_settings";
    private static final String THEME                 = "theme";
    public static final  String VERSION               = "version";
    private static final String WINDOW_POSITIONS      = "window_positions";

    public static final int DEFAULT_LIBRARY_EXPLORER_DIVIDER_POSITION = 300;

    public static final int MAX_QUICK_EXPORTS = 100;
    public static final int MAX_RECENT_FILES  = 20;

    private static Settings INSTANCE;

    private Version                          mLastSeenGCSVersion;
    private GeneralSettings                  mGeneralSettings;
    private int                              mLibraryExplorerDividerPosition;
    private List<String>                     mLibraryExplorerOpenRowKeys;
    private List<Path>                       mRecentFiles;
    private Map<String, QuickExport>         mQuickExports;
    private PageRefSettings                  mPageRefSettings;
    private Map<String, String>              mKeyBindingOverrides;
    private Map<String, BaseWindow.Position> mBaseWindowPositions;
    private SheetSettings                    mSheetSettings;
    private int                              mLastRecentFilesUpdateCounter;

    public static synchronized Settings getInstance() {
        if (INSTANCE == null) {
            INSTANCE = new Settings();
            // Have to do the pruning of invalid quick exports here, since the isValid() call
            // uses preferences to determine some validity.
            List<String> toRemove = new ArrayList<>();
            for (Map.Entry<String, QuickExport> entry : INSTANCE.mQuickExports.entrySet()) {
                if (!entry.getValue().isValid()) {
                    toRemove.add(entry.getKey());
                }
            }
            for (String key : toRemove) {
                INSTANCE.mQuickExports.remove(key);
            }
        }
        return INSTANCE;
    }

    private static Path getPreferencesPath() {
        String homeDir = System.getProperty("user.home", ".");
        Path   path;
        if (Platform.isMacintosh()) {
            path = Path.of(homeDir, "Library", "Preferences");
        } else if (Platform.isWindows()) {
            String localAppData = System.getenv("LOCALAPPDATA");
            path = localAppData != null ? Path.of(localAppData) : Path.of(homeDir, "AppData", "Local");
        } else {
            path = Path.of(homeDir, ".config");
        }
        return path.resolve("gcs.json").normalize().toAbsolutePath();
    }

    private Settings() {
        mLastSeenGCSVersion = new Version(GCS.VERSION);
        mGeneralSettings = new GeneralSettings();
        Library.LIBRARIES.clear();
        mLibraryExplorerDividerPosition = DEFAULT_LIBRARY_EXPLORER_DIVIDER_POSITION;
        mLibraryExplorerOpenRowKeys = new ArrayList<>();
        mRecentFiles = new ArrayList<>();
        mQuickExports = new HashMap<>();
        mPageRefSettings = new PageRefSettings();
        mKeyBindingOverrides = new HashMap<>();
        mBaseWindowPositions = new HashMap<>();
        mSheetSettings = new SheetSettings();
        Path path = getPreferencesPath();
        if (Files.isReadable(path) && Files.isRegularFile(path)) {
            try (BufferedReader in = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
                JsonMap m = Json.asMap(Json.parse(in));
                if (!m.isEmpty()) {
                    int version = m.getInt(VERSION);
                    if (version >= MINIMUM_VERSION && version <= DataFile.CURRENT_VERSION) {
                        Version loadVersion = new Version(m.getString(LAST_SEEN_GCS_VERSION));
                        if (loadVersion.compareTo(mLastSeenGCSVersion) > 0) {
                            mLastSeenGCSVersion = loadVersion;
                        }
                        if (m.has(GENERAL)) {
                            mGeneralSettings = new GeneralSettings(m.getMap(GENERAL));
                        } else {
                            // Old data, before general settings were separated on June 25, 2021
                            mGeneralSettings = new GeneralSettings(m);
                        }
                        if (m.has(LIBRARIES)) {
                            JsonMap m2 = m.getMap(LIBRARIES);
                            for (String key : m2.keySet()) {
                                Library.LIBRARIES.add(Library.fromJSON(key, m2.getMap(key)));
                            }
                        }
                        if (m.has(LIBRARY_EXPLORER)) {
                            JsonMap m2 = m.getMap(LIBRARY_EXPLORER);
                            mLibraryExplorerDividerPosition = m2.getIntWithDefault(DIVIDER_POSITION, mLibraryExplorerDividerPosition);
                            JsonArray a      = m2.getArray(OPEN_ROW_KEYS);
                            int       length = a.size();
                            for (int i = 0; i < length; i++) {
                                mLibraryExplorerOpenRowKeys.add(a.getString(i));
                            }
                        }
                        if (m.has(RECENT_FILES)) {
                            JsonArray a      = m.getArray(RECENT_FILES);
                            int       length = a.size();
                            for (int i = 0; i < length; i++) {
                                mRecentFiles.add(Path.of(a.getString(i)).normalize().toAbsolutePath());
                            }
                        }
                        if (m.has(LAST_DIRS)) {
                            Dirs.load(m.getMap(LAST_DIRS));
                        }
                        if (m.has(PAGE_REFS)) {
                            mPageRefSettings = new PageRefSettings(m.getMap(PAGE_REFS));
                        } else if (m.has(DEPRECATED_PDF_REFS)) {
                            mPageRefSettings = new PageRefSettings(m.getMap(DEPRECATED_PDF_REFS));
                        }
                        if (m.has(KEY_BINDINGS)) {
                            JsonMap m2 = m.getMap(KEY_BINDINGS);
                            for (String key : m2.keySet()) {
                                if (key != null && !key.isBlank()) {
                                    mKeyBindingOverrides.put(key, m2.getString(key));
                                }
                            }
                        }
                        if (m.has(THEME)) {
                            JsonMap m2 = m.getMap(THEME);
                            if (m2.has(COLORS)) {
                                Colors.setCurrentThemeColors(new Colors(m2.getMap(COLORS)));
                            }
                            if (m2.has(FONTS)) {
                                Fonts.setCurrentThemeFonts(new Fonts(m2.getMap(FONTS)));
                            }
                        }
                        if (m.has(WINDOW_POSITIONS)) {
                            JsonMap m2 = m.getMap(WINDOW_POSITIONS);
                            mBaseWindowPositions = new HashMap<>();
                            for (String key : m2.keySet()) {
                                mBaseWindowPositions.put(key, new BaseWindow.Position(m2.getMap(key)));
                            }
                        }
                        if (m.has(QUICK_EXPORTS)) {
                            JsonMap m2 = m.getMap(QUICK_EXPORTS);
                            for (String key : m2.keySet()) {
                                mQuickExports.put(key, new QuickExport(m2.getMap(key)));
                            }
                        }
                        if (m.has(SHEET_SETTINGS)) {
                            mSheetSettings.load(m.getMap(SHEET_SETTINGS));
                        } else {
                            mSheetSettings.load(m);
                        }
                    }
                }
            } catch (Exception exception) {
                Log.error(exception);
            }
        }
        boolean hasMaster = false;
        boolean hasUser   = false;
        for (Library lib : Library.LIBRARIES) {
            if (lib == Library.MASTER) {
                hasMaster = true;
                if (hasUser) {
                    break;
                }
            } else if (lib == Library.USER) {
                hasUser = true;
                if (hasMaster) {
                    break;
                }
            }
        }
        if (!hasMaster) {
            Library.LIBRARIES.add(Library.MASTER);
        }
        if (!hasUser) {
            Library.LIBRARIES.add(Library.USER);
        }
        Collections.sort(Library.LIBRARIES);
        mGeneralSettings.updateToolTipTiming();
    }

    public void save() {
        try {
            SafeFileUpdater trans = new SafeFileUpdater();
            trans.begin();
            try {
                Path path = getPreferencesPath();
                Files.createDirectories(path.getParent());
                File file = trans.getTransactionFile(path.toFile());
                try (JsonWriter w = new JsonWriter(new BufferedWriter(new FileWriter(file, StandardCharsets.UTF_8)), "\t")) {
                    w.startMap();
                    w.keyValue(VERSION, DataFile.CURRENT_VERSION);
                    w.keyValue(LAST_SEEN_GCS_VERSION, mLastSeenGCSVersion.toString());
                    w.key(GENERAL);
                    mGeneralSettings.save(w);
                    w.key(LIBRARIES);
                    w.startMap();
                    for (Library lib : Library.LIBRARIES) {
                        lib.toJSON(w);
                    }
                    w.endMap();
                    w.key(LIBRARY_EXPLORER);
                    w.startMap();
                    w.keyValue(DIVIDER_POSITION, mLibraryExplorerDividerPosition);
                    w.key(OPEN_ROW_KEYS);
                    w.startArray();
                    for (String key : mLibraryExplorerOpenRowKeys) {
                        w.value(key);
                    }
                    w.endArray();
                    w.endMap();
                    w.key(RECENT_FILES);
                    w.startArray();
                    for (Path p : mRecentFiles) {
                        w.value(p.toString());
                    }
                    w.endArray();
                    Dirs.save(LAST_DIRS, w);
                    w.key(PAGE_REFS);
                    mPageRefSettings.save(w);
                    w.key(KEY_BINDINGS);
                    w.startMap();
                    for (Map.Entry<String, String> entry : mKeyBindingOverrides.entrySet()) {
                        w.keyValue(entry.getKey(), entry.getValue());
                    }
                    w.endMap();
                    w.key(WINDOW_POSITIONS);
                    w.startMap();
                    long cutoff = System.currentTimeMillis() - 1000L * 60L * 60L * 24L * 45L;
                    for (Map.Entry<String, BaseWindow.Position> entry : mBaseWindowPositions.entrySet()) {
                        BaseWindow.Position info = entry.getValue();
                        if (info.mLastUpdated > cutoff) {
                            w.key(entry.getKey());
                            info.toJSON(w);
                        }
                    }
                    w.endMap();
                    w.key(THEME);
                    w.startMap();
                    w.key(COLORS);
                    Colors.currentThemeColors().save(w);
                    w.key(FONTS);
                    Fonts.currentThemeFonts().save(w);
                    w.endMap();
                    pruneQuickExports();
                    if (!mQuickExports.isEmpty()) {
                        w.key(QUICK_EXPORTS);
                        w.startMap();
                        for (Map.Entry<String, QuickExport> entry : mQuickExports.entrySet()) {
                            w.key(entry.getKey());
                            entry.getValue().toJSON(w);
                        }
                        w.endMap();
                    }
                    w.key(SHEET_SETTINGS);
                    mSheetSettings.save(w, true);
                    w.endMap();
                }
            } catch (IOException ioe) {
                trans.abort();
                throw ioe;
            }
            trans.commit();
        } catch (Exception exception) {
            Log.error(exception);
        }
    }

    public Version getLastSeenGCSVersion() {
        return mLastSeenGCSVersion;
    }

    public void setLastSeenGCSVersion(Version lastSeenGCSVersion) {
        mLastSeenGCSVersion = new Version(lastSeenGCSVersion);
    }

    public GeneralSettings getGeneralSettings() {
        return mGeneralSettings;
    }

    public int getLibraryExplorerDividerPosition() {
        return mLibraryExplorerDividerPosition;
    }

    public void setLibraryExplorerDividerPosition(int libraryExplorerDividerPosition) {
        mLibraryExplorerDividerPosition = libraryExplorerDividerPosition;
    }

    public List<String> getLibraryExplorerOpenRowKeys() {
        return mLibraryExplorerOpenRowKeys;
    }

    public void setLibraryExplorerOpenRowKeys(List<String> libraryExplorerOpenRowKeys) {
        mLibraryExplorerOpenRowKeys = libraryExplorerOpenRowKeys;
    }

    public static String linesToString(List<String> lines) {
        StringBuilder buffer = new StringBuilder();
        for (String line : lines) {
            if (!buffer.isEmpty()) {
                buffer.append('\n');
            }
            buffer.append(line);
        }
        return buffer.toString();
    }

    public int getLastRecentFilesUpdateCounter() {
        return mLastRecentFilesUpdateCounter;
    }

    public List<Path> getRecentFiles() {
        return mRecentFiles;
    }

    public void addRecentFile(Path path) {
        String extension = PathUtils.getExtension(path);
        if (Platform.isMacintosh() || Platform.isWindows()) {
            extension = extension.toLowerCase();
        }
        for (FileType fileType : FileType.ALL_OPENABLE) {
            if (fileType.matchExtension(extension)) {
                if (Files.isReadable(path)) {
                    mLastRecentFilesUpdateCounter++;
                    path = path.normalize().toAbsolutePath();
                    mRecentFiles.remove(path);
                    mRecentFiles.add(0, path);
                    if (mRecentFiles.size() > MAX_RECENT_FILES) {
                        mRecentFiles.remove(MAX_RECENT_FILES);
                    }
                }
                break;
            }
        }
    }

    public void setRecentFiles(List<Path> recentFiles) {
        mRecentFiles = recentFiles;
        mLastRecentFilesUpdateCounter++;
    }

    public PageRefSettings getPDFRefSettings() {
        return mPageRefSettings;
    }

    public String getKeyBindingOverride(String key) {
        return mKeyBindingOverrides.get(key);
    }

    public void setKeyBindingOverride(String key, String override) {
        if (override == null || override.isBlank()) {
            mKeyBindingOverrides.remove(key);
        } else {
            mKeyBindingOverrides.put(key, override);
        }
    }

    public BaseWindow.Position getBaseWindowPosition(String key) {
        return mBaseWindowPositions.get(key);
    }

    public void putBaseWindowPosition(String key, BaseWindow.Position info) {
        mBaseWindowPositions.put(key, info);
    }

    public QuickExport getQuickExport(String path) {
        return mQuickExports.get(path);
    }

    public void putQuickExport(GURPSCharacter character, QuickExport qe) {
        Path path = character.getPath();
        if (path != null) {
            mQuickExports.put(path.toAbsolutePath().toString(), qe);
        }
    }

    public void pruneQuickExports() {
        int size = mQuickExports.size();
        if (size > MAX_QUICK_EXPORTS) {
            List<QuickExport> all = new ArrayList<>(size);
            for (Map.Entry<String, QuickExport> entry : mQuickExports.entrySet()) {
                QuickExport qe = entry.getValue();
                qe.setKey(entry.getKey());
                all.add(qe);
            }
            Collections.sort(all);
            for (int i = MAX_QUICK_EXPORTS; i < size; i++) {
                mQuickExports.remove(all.get(i).getKey());
            }
        }
    }

    public SheetSettings getSheetSettings() {
        return mSheetSettings;
    }
}
