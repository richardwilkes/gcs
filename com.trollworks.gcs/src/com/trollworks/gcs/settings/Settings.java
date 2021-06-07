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
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.pdfview.PDFRef;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.Theme;
import com.trollworks.gcs.ui.scale.Scales;
import com.trollworks.gcs.ui.widget.BaseWindow;
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
import com.trollworks.gcs.utility.text.Enums;

import java.awt.GraphicsEnvironment;
import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import javax.swing.ToolTipManager;

public final class Settings extends ChangeableData {
    private static final int MINIMUM_VERSION = 1;

    private static final String DEPRECATED_AUTO_NAME_NEW_CHARACTERS = "auto_name_new_characters"; // March 21, 2021
    private static final String DEPRECATED_PNG_RESOLUTION           = "png_resolution"; // June 6, 2021

    private static final String AUTO_FILL_PROFILE               = "auto_fill_profile";
    private static final String DEFAULT_PLAYER_NAME             = "default_player_name";
    private static final String DEFAULT_TECH_LEVEL              = "default_tech_level";
    private static final String DIVIDER_POSITION                = "divider_position";
    private static final String FONTS                           = "fonts";
    private static final String GURPS_CALCULATOR_KEY            = "gurps_calculator_key";
    private static final String IMAGE_RESOLUTION                = "image_resolution";
    private static final String INCLUDE_UNSPENT_POINTS_IN_TOTAL = "include_unspent_points_in_total";
    private static final String INITIAL_POINTS                  = "initial_points";
    private static final String INITIAL_UI_SCALE                = "initial_ui_scale";
    private static final String KEY_BINDINGS                    = "key_bindings";
    private static final String LAST_DIR                        = "last_dir";
    private static final String LAST_SEEN_GCS_VERSION           = "last_seen_gcs_version";
    private static final String LIBRARIES                       = "libraries";
    private static final String LIBRARY_EXPLORER                = "library_explorer";
    private static final String OPEN_ROW_KEYS                   = "open_row_keys";
    private static final String PDF_REFS                        = "pdf_refs";
    private static final String QUICK_EXPORTS                   = "quick_exports";
    private static final String RECENT_FILES                    = "recent_files";
    private static final String SHEET_SETTINGS                  = "sheet_settings";
    private static final String THEME                           = "theme";
    private static final String TOOLTIP_TIMEOUT                 = "tooltip_timeout";
    private static final String VERSION                         = "version";
    private static final String WINDOW_POSITIONS                = "window_positions";

    public static final boolean DEFAULT_AUTO_FILL_PROFILE                 = true;
    public static final boolean DEFAULT_INCLUDE_UNSPENT_POINTS_IN_TOTAL   = true;
    public static final int     DEFAULT_IMAGE_RESOLUTION                  = 200;
    public static final int     DEFAULT_INITIAL_POINTS                    = 250;
    public static final int     DEFAULT_LIBRARY_EXPLORER_DIVIDER_POSITION = 300;
    public static final int     DEFAULT_TOOLTIP_TIMEOUT                   = 60;
    public static final Scales  DEFAULT_INITIAL_UI_SCALE                  = Scales.QUARTER_AGAIN_SIZE;
    public static final String  DEFAULT_DEFAULT_PLAYER_NAME               = System.getProperty("user.name", "");
    public static final String  DEFAULT_DEFAULT_TECH_LEVEL                = "3";

    public static final int MAX_QUICK_EXPORTS = 100;
    public static final int MAX_RECENT_FILES  = 20;

    private static Settings                         INSTANCE;
    private        Version                          mLastSeenGCSVersion;
    private        int                              mInitialPoints;
    private        int                              mToolTipTimeout;
    private        int                              mLibraryExplorerDividerPosition;
    private        List<String>                     mLibraryExplorerOpenRowKeys;
    private        Scales                           mInitialUIScale;
    private        List<Path>                       mRecentFiles;
    private        Map<String, QuickExport>         mQuickExports;
    private        Path                             mLastDir;
    private        Map<String, PDFRef>              mPdfRefs;
    private        Map<String, String>              mKeyBindingOverrides;
    private        Map<String, Fonts.Info>          mFontInfo;
    private        Map<String, BaseWindow.Position> mBaseWindowPositions;
    private        String                           mGURPSCalculatorKey;
    private        String                           mDefaultPlayerName;
    private        String                           mDefaultTechLevel;
    private        SheetSettings                    mSheetSettings;
    private        int                              mLastRecentFilesUpdateCounter;
    private        int                              mImageResolution;
    private        boolean                          mIncludeUnspentPointsInTotal;
    private        boolean                          mAutoFillProfile;

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
            path = Paths.get(homeDir, "Library", "Preferences");
        } else if (Platform.isWindows()) {
            String localAppData = System.getenv("LOCALAPPDATA");
            path = localAppData != null ? Paths.get(localAppData) : Paths.get(homeDir, "AppData", "Local");
        } else {
            path = Paths.get(homeDir, ".config");
        }
        return path.resolve("gcs.json").normalize().toAbsolutePath();
    }

    private Settings() {
        mLastSeenGCSVersion = new Version(GCS.VERSION);
        Library.LIBRARIES.clear();
        mInitialPoints = DEFAULT_INITIAL_POINTS;
        mToolTipTimeout = DEFAULT_TOOLTIP_TIMEOUT;
        mLibraryExplorerDividerPosition = DEFAULT_LIBRARY_EXPLORER_DIVIDER_POSITION;
        mLibraryExplorerOpenRowKeys = new ArrayList<>();
        mInitialUIScale = DEFAULT_INITIAL_UI_SCALE;
        mRecentFiles = new ArrayList<>();
        mQuickExports = new HashMap<>();
        mLastDir = Paths.get(System.getProperty("user.home", ".")).normalize().toAbsolutePath();
        mGURPSCalculatorKey = "";
        mDefaultPlayerName = DEFAULT_DEFAULT_PLAYER_NAME;
        mDefaultTechLevel = DEFAULT_DEFAULT_TECH_LEVEL;
        mImageResolution = DEFAULT_IMAGE_RESOLUTION;
        mPdfRefs = new HashMap<>();
        mFontInfo = new HashMap<>();
        mKeyBindingOverrides = new HashMap<>();
        mBaseWindowPositions = new HashMap<>();
        mIncludeUnspentPointsInTotal = DEFAULT_INCLUDE_UNSPENT_POINTS_IN_TOTAL;
        mAutoFillProfile = DEFAULT_AUTO_FILL_PROFILE;
        mSheetSettings = new SheetSettings();
        Path path = getPreferencesPath();
        if (Files.isReadable(path) && Files.isRegularFile(path)) {
            try (BufferedReader in = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
                JsonMap m = Json.asMap(Json.parse(in));
                if (!m.isEmpty()) {
                    int version = m.getInt(VERSION);
                    if (version >= MINIMUM_VERSION && version <= DataFile.CURRENT_VERSION) {
                        LoadState state = new LoadState();
                        state.mDataFileVersion = version;
                        Version loadVersion = new Version(m.getString(LAST_SEEN_GCS_VERSION));
                        if (loadVersion.compareTo(mLastSeenGCSVersion) > 0) {
                            mLastSeenGCSVersion = loadVersion;
                        }
                        if (m.has(LIBRARIES)) {
                            JsonMap m2 = m.getMap(LIBRARIES);
                            for (String key : m2.keySet()) {
                                Library.LIBRARIES.add(Library.fromJSON(key, m2.getMap(key)));
                            }
                        }
                        mInitialPoints = m.getIntWithDefault(INITIAL_POINTS, mInitialPoints);
                        mToolTipTimeout = m.getIntWithDefault(TOOLTIP_TIMEOUT, mToolTipTimeout);
                        if (m.has(LIBRARY_EXPLORER)) {
                            JsonMap m2 = m.getMap(LIBRARY_EXPLORER);
                            mLibraryExplorerDividerPosition = m2.getIntWithDefault(DIVIDER_POSITION, mLibraryExplorerDividerPosition);
                            JsonArray a      = m2.getArray(OPEN_ROW_KEYS);
                            int       length = a.size();
                            for (int i = 0; i < length; i++) {
                                mLibraryExplorerOpenRowKeys.add(a.getString(i));
                            }
                        }
                        mInitialUIScale = Enums.extract(m.getStringWithDefault(INITIAL_UI_SCALE, ""), Scales.values(), mInitialUIScale);
                        if (m.has(RECENT_FILES)) {
                            JsonArray a      = m.getArray(RECENT_FILES);
                            int       length = a.size();
                            for (int i = 0; i < length; i++) {
                                mRecentFiles.add(Paths.get(a.getString(i)).normalize().toAbsolutePath());
                            }
                        }
                        mLastDir = Paths.get(m.getStringWithDefault(LAST_DIR, mLastDir.toString())).normalize().toAbsolutePath();
                        if (m.has(PDF_REFS)) {
                            JsonMap m2 = m.getMap(PDF_REFS);
                            for (String key : m2.keySet()) {
                                mPdfRefs.put(key, new PDFRef(m2.getMap(key)));
                            }
                        }
                        if (m.has(KEY_BINDINGS)) {
                            JsonMap m2 = m.getMap(KEY_BINDINGS);
                            for (String key : m2.keySet()) {
                                if (key != null && !key.isBlank()) {
                                    mKeyBindingOverrides.put(key, m2.getString(key));
                                }
                            }
                        }
                        if (m.has(FONTS)) {
                            JsonMap m2 = m.getMap(FONTS);
                            for (String key : m2.keySet()) {
                                mFontInfo.put(key, new Fonts.Info(m2.getMap(key)));
                            }
                        }
                        if (m.has(WINDOW_POSITIONS)) {
                            JsonMap m2 = m.getMap(WINDOW_POSITIONS);
                            mBaseWindowPositions = new HashMap<>();
                            for (String key : m2.keySet()) {
                                mBaseWindowPositions.put(key, new BaseWindow.Position(m2.getMap(key)));
                            }
                        }
                        mGURPSCalculatorKey = m.getStringWithDefault(GURPS_CALCULATOR_KEY, mGURPSCalculatorKey);
                        mDefaultPlayerName = m.getStringWithDefault(DEFAULT_PLAYER_NAME, mDefaultPlayerName);
                        mDefaultTechLevel = m.getStringWithDefault(DEFAULT_TECH_LEVEL, mDefaultTechLevel);
                        if (m.has(DEPRECATED_PNG_RESOLUTION)) {
                            mImageResolution = m.getIntWithDefault(DEPRECATED_PNG_RESOLUTION, mImageResolution);
                        } else {
                            mImageResolution = m.getIntWithDefault(IMAGE_RESOLUTION, mImageResolution);
                        }
                        mIncludeUnspentPointsInTotal = m.getBooleanWithDefault(INCLUDE_UNSPENT_POINTS_IN_TOTAL, mIncludeUnspentPointsInTotal);
                        if (m.has(DEPRECATED_AUTO_NAME_NEW_CHARACTERS)) {
                            mAutoFillProfile = m.getBooleanWithDefault(DEPRECATED_AUTO_NAME_NEW_CHARACTERS, mAutoFillProfile);
                        } else {
                            mAutoFillProfile = m.getBooleanWithDefault(AUTO_FILL_PROFILE, mAutoFillProfile);
                        }
                        if (m.has(THEME)) {
                            Theme.set(new Theme(m.getMap(THEME)));
                        }
                        if (m.has(QUICK_EXPORTS)) {
                            JsonMap m2 = m.getMap(QUICK_EXPORTS);
                            for (String key : m2.keySet()) {
                                mQuickExports.put(key, new QuickExport(m2.getMap(key)));
                            }
                        }
                        if (m.has(SHEET_SETTINGS)) {
                            mSheetSettings.load(m.getMap(SHEET_SETTINGS), state);
                        } else {
                            mSheetSettings.load(m, state);
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
        if (!GraphicsEnvironment.isHeadless()) {
            ToolTipManager.sharedInstance().setDismissDelay(mToolTipTimeout * 1000);
        }
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
                    w.key(LIBRARIES);
                    w.startMap();
                    for (Library lib : Library.LIBRARIES) {
                        lib.toJSON(w);
                    }
                    w.endMap();
                    w.keyValue(INITIAL_POINTS, mInitialPoints);
                    w.keyValue(TOOLTIP_TIMEOUT, mToolTipTimeout);
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
                    w.keyValue(INITIAL_UI_SCALE, Enums.toId(mInitialUIScale));
                    w.key(RECENT_FILES);
                    w.startArray();
                    for (Path p : mRecentFiles) {
                        w.value(p.toString());
                    }
                    w.endArray();
                    w.keyValue(LAST_DIR, mLastDir.toString());
                    w.key(PDF_REFS);
                    w.startMap();
                    for (Map.Entry<String, PDFRef> entry : mPdfRefs.entrySet()) {
                        w.key(entry.getKey());
                        entry.getValue().toJSON(w);
                    }
                    w.endMap();
                    w.key(KEY_BINDINGS);
                    w.startMap();
                    for (Map.Entry<String, String> entry : mKeyBindingOverrides.entrySet()) {
                        w.keyValue(entry.getKey(), entry.getValue());
                    }
                    w.endMap();
                    w.key(FONTS);
                    w.startMap();
                    for (Map.Entry<String, Fonts.Info> entry : mFontInfo.entrySet()) {
                        w.key(entry.getKey());
                        entry.getValue().toJSON(w);
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
                    w.keyValue(GURPS_CALCULATOR_KEY, mGURPSCalculatorKey);
                    w.keyValue(DEFAULT_PLAYER_NAME, mDefaultPlayerName);
                    w.keyValue(DEFAULT_TECH_LEVEL, mDefaultTechLevel);
                    w.keyValue(IMAGE_RESOLUTION, mImageResolution);
                    w.keyValue(INCLUDE_UNSPENT_POINTS_IN_TOTAL, mIncludeUnspentPointsInTotal);
                    w.keyValue(AUTO_FILL_PROFILE, mAutoFillProfile);
                    w.key(THEME);
                    Theme.current().save(w);
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
                    mSheetSettings.toJSON(w);
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

    public int getInitialPoints() {
        return mInitialPoints;
    }

    public void setInitialPoints(int initialPoints) {
        mInitialPoints = initialPoints;
    }

    public int getToolTipTimeout() {
        return mToolTipTimeout;
    }

    public void setToolTipTimeout(int toolTipTimeout) {
        if (mToolTipTimeout != toolTipTimeout) {
            mToolTipTimeout = toolTipTimeout;
            ToolTipManager.sharedInstance().setDismissDelay(mToolTipTimeout * 1000);
        }
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

    public Scales getInitialUIScale() {
        return mInitialUIScale;
    }

    public void setInitialUIScale(Scales initialUIScale) {
        mInitialUIScale = initialUIScale;
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

    public Path getLastDir() {
        if (!Files.isDirectory(mLastDir)) {
            mLastDir = Paths.get(System.getProperty("user.home", ".")).normalize().toAbsolutePath();
        }
        return mLastDir;
    }

    public void setLastDir(Path lastDir) {
        mLastDir = lastDir.normalize().toAbsolutePath();
    }

    public String getGURPSCalculatorKey() {
        return mGURPSCalculatorKey;
    }

    public void setGURPSCalculatorKey(String key) {
        mGURPSCalculatorKey = key;
    }

    public String getDefaultPlayerName() {
        return mDefaultPlayerName;
    }

    public void setDefaultPlayerName(String defaultPlayerName) {
        mDefaultPlayerName = defaultPlayerName;
    }

    public String getDefaultTechLevel() {
        return mDefaultTechLevel;
    }

    public void setDefaultTechLevel(String defaultTechLevel) {
        mDefaultTechLevel = defaultTechLevel;
    }

    public int getImageResolution() {
        return mImageResolution;
    }

    public void setImageResolution(int resolution) {
        mImageResolution = resolution;
    }

    public List<PDFRef> allPdfRefs(boolean requireExistence) {
        List<PDFRef> list = new ArrayList<>();
        for (String key : mPdfRefs.keySet()) {
            PDFRef ref = lookupPdfRef(key, requireExistence);
            if (ref != null) {
                list.add(ref);
            }
        }
        Collections.sort(list);
        return list;
    }

    public PDFRef lookupPdfRef(String id, boolean requireExistence) {
        PDFRef ref = mPdfRefs.get(id);
        if (ref == null) {
            return null;
        }
        if (requireExistence) {
            Path path = ref.getPath();
            if (!Files.isReadable(path) || !Files.isRegularFile(path)) {
                return null;
            }
        }
        return ref;
    }

    public void putPdfRef(PDFRef ref) {
        mPdfRefs.put(ref.getID(), ref);
    }

    public void removePdfRef(PDFRef ref) {
        mPdfRefs.remove(ref.getID());
    }

    public Fonts.Info getFontInfo(String key) {
        return mFontInfo.get(key);
    }

    public void setFontInfo(String key, Fonts.Info fontInfo) {
        mFontInfo.put(key, fontInfo);
        notifyOfChange();
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

    public boolean includeUnspentPointsInTotal() {
        return mIncludeUnspentPointsInTotal;
    }

    public void setIncludeUnspentPointsInTotal(boolean includeUnspentPointsInTotal) {
        if (mIncludeUnspentPointsInTotal != includeUnspentPointsInTotal) {
            mIncludeUnspentPointsInTotal = includeUnspentPointsInTotal;
            notifyOfChange();
        }
    }

    public boolean autoFillProfile() {
        return mAutoFillProfile;
    }

    public void setAutoFillProfile(boolean autoFillProfile) {
        mAutoFillProfile = autoFillProfile;
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
