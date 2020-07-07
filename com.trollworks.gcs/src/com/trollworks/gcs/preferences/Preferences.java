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

package com.trollworks.gcs.preferences;

import com.trollworks.gcs.GCS;
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.DisplayOption;
import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.pdfview.PdfRef;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.print.PageOrientation;
import com.trollworks.gcs.ui.print.PrintManager;
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
import com.trollworks.gcs.utility.notification.Notifier;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.WeightUnits;

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
import java.util.UUID;
import javax.swing.ToolTipManager;

/** Provides the implementation of preferences. Note: not all preferences emit notifications. */
public class Preferences {
    private static final int CURRENT_VERSION        = 2;
    private static final int VERSION_AS_OF_GCS_4_18 = 2;
    private static final int MINIMUM_VERSION        = 1;

    private static final String AUTO_NAME_NEW_CHARACTERS        = "auto_name_new_characters";
    private static final String BASE_WILL_AND_PER_ON_10         = "base_will_and_per_on_10";
    private static final String BLOCK_LAYOUT                    = "block_layout";
    private static final String DEFAULT_LENGTH_UNITS            = "default_length_units";
    private static final String DEFAULT_PAGE_SETTINGS           = "default_page_settings";
    private static final String DEFAULT_PLAYER_NAME             = "default_player_name";
    private static final String DEFAULT_PORTRAIT_PATH           = "default_portrait_path";
    private static final String DEFAULT_TECH_LEVEL              = "default_tech_level";
    private static final String DEFAULT_WEIGHT_UNITS            = "default_weight_units";
    private static final String DIVIDER_POSITION                = "divider_position";
    private static final String FONTS                           = "fonts";
    private static final String GURPS_CALCULATOR_KEY            = "gurps_calculator_key";
    private static final String ID                              = "id";
    private static final String INCLUDE_UNSPENT_POINTS_IN_TOTAL = "include_unspent_points_in_total";
    private static final String INITIAL_POINTS                  = "initial_points";
    private static final String INITIAL_UI_SCALE                = "initial_ui_scale";
    private static final String KEY_BINDINGS                    = "key_bindings";
    private static final String LAST_DIR                        = "last_dir";
    private static final String LAST_SEEN_GCS_VERSION           = "last_seen_gcs_version";
    private static final String LIBRARIES                       = "libraries";
    private static final String LIBRARY_EXPLORER                = "library_explorer";
    private static final String MODIFIERS_DISPLAY               = "modifiers_display";
    private static final String NOTES_DISPLAY                   = "notes_display";
    private static final String OPEN_ROW_KEYS                   = "open_row_keys";
    private static final String PDF_REFS                        = "pdf_refs";
    private static final String PNG_RESOLUTION                  = "png_resolution";
    private static final String RECENT_FILES                    = "recent_files";
    private static final String SHOW_COLLEGE_IN_SHEET_SPELLS    = "show_college_in_sheet_spells";
    private static final String TOOLTIP_TIMEOUT                 = "tooltip_timeout";
    private static final String USE_KNOW_YOUR_OWN_STRENGTH      = "use_know_your_own_strength";
    private static final String USE_MODIFYING_DICE_PLUS_ADDS    = "use_modifying_dice_plus_adds";
    private static final String USE_MULTIPLICATIVE_MODIFIERS    = "use_multiplicative_modifiers";
    private static final String USE_NATIVE_PRINT_DIALOGS        = "use_native_print_dialogs";
    private static final String USE_REDUCED_SWING               = "use_reduced_swing";
    private static final String USE_SIMPLE_METRIC_CONVERSIONS   = "use_simple_metric_conversions";
    private static final String USE_THRUST_EQUALS_SWING_MINUS_2 = "use_thrust_equals_swing_minus_2";
    private static final String USER_DESCRIPTION_DISPLAY        = "user_description_display";
    private static final String VERSION                         = "version";
    private static final String WINDOW_POSITIONS                = "window_positions";

    // No longer used as of the 4.18 release -- remove at some point in the future
    private static final String LAST_GCS_VERSION    = "last_gcs_version";
    private static final String MASTER_LIBRARY_PATH = "master_library_path";
    private static final String USER_LIBRARY_PATH   = "user_library_path";

    public static final String KEY_PREFIX           = "prefs.";
    public static final String KEY_PER_SHEET_PREFIX = KEY_PREFIX + "sheet.";

    public static final String KEY_BASE_WILL_AND_PER_ON_10         = KEY_PER_SHEET_PREFIX + BASE_WILL_AND_PER_ON_10;
    public static final String KEY_BLOCK_LAYOUT                    = KEY_PER_SHEET_PREFIX + BLOCK_LAYOUT;
    public static final String KEY_DEFAULT_LENGTH_UNITS            = KEY_PER_SHEET_PREFIX + DEFAULT_LENGTH_UNITS;
    public static final String KEY_DEFAULT_WEIGHT_UNITS            = KEY_PER_SHEET_PREFIX + DEFAULT_WEIGHT_UNITS;
    public static final String KEY_INCLUDE_UNSPENT_POINTS_IN_TOTAL = KEY_PREFIX + INCLUDE_UNSPENT_POINTS_IN_TOTAL;
    public static final String KEY_MODIFIERS_DISPLAY               = KEY_PER_SHEET_PREFIX + MODIFIERS_DISPLAY;
    public static final String KEY_NOTES_DISPLAY                   = KEY_PER_SHEET_PREFIX + NOTES_DISPLAY;
    public static final String KEY_SHOW_COLLEGE_IN_SHEET_SPELLS    = KEY_PER_SHEET_PREFIX + SHOW_COLLEGE_IN_SHEET_SPELLS;
    public static final String KEY_USE_KNOW_YOUR_OWN_STRENGTH      = KEY_PER_SHEET_PREFIX + USE_KNOW_YOUR_OWN_STRENGTH;
    public static final String KEY_USE_MODIFYING_DICE_PLUS_ADDS    = KEY_PER_SHEET_PREFIX + USE_MODIFYING_DICE_PLUS_ADDS;
    public static final String KEY_USE_MULTIPLICATIVE_MODIFIERS    = KEY_PER_SHEET_PREFIX + USE_MULTIPLICATIVE_MODIFIERS;
    public static final String KEY_USE_REDUCED_SWING               = KEY_PER_SHEET_PREFIX + USE_REDUCED_SWING;
    public static final String KEY_USE_SIMPLE_METRIC_CONVERSIONS   = KEY_PER_SHEET_PREFIX + USE_SIMPLE_METRIC_CONVERSIONS;
    public static final String KEY_USE_THRUST_EQUALS_SWING_MINUS_2 = KEY_PER_SHEET_PREFIX + USE_THRUST_EQUALS_SWING_MINUS_2;
    public static final String KEY_USER_DESCRIPTION_DISPLAY        = KEY_PER_SHEET_PREFIX + USER_DESCRIPTION_DISPLAY;

    public static final boolean       DEFAULT_AUTO_NAME_NEW_CHARACTERS          = true;
    public static final boolean       DEFAULT_BASE_WILL_AND_PER_ON_10           = false;
    public static final boolean       DEFAULT_INCLUDE_UNSPENT_POINTS_IN_TOTAL   = true;
    public static final int           DEFAULT_INITIAL_POINTS                    = 100;
    public static final int           DEFAULT_LIBRARY_EXPLORER_DIVIDER_POSITION = 300;
    public static final boolean       DEFAULT_SHOW_COLLEGE_IN_SHEET_SPELLS      = false;
    public static final boolean       DEFAULT_USE_KNOW_YOUR_OWN_STRENGTH        = false;
    public static final boolean       DEFAULT_USE_MODIFYING_DICE_PLUS_ADDS      = false;
    public static final boolean       DEFAULT_USE_MULTIPLICATIVE_MODIFIERS      = false;
    public static final boolean       DEFAULT_USE_NATIVE_PRINT_DIALOGS          = false;
    public static final boolean       DEFAULT_USE_REDUCED_SWING                 = false;
    public static final boolean       DEFAULT_USE_SIMPLE_METRIC_CONVERSIONS     = true;
    public static final boolean       DEFAULT_USE_THRUST_EQUALS_SWING_MINUS_2   = false;
    public static final DisplayOption DEFAULT_MODIFIERS_DISPLAY                 = DisplayOption.INLINE;
    public static final DisplayOption DEFAULT_NOTES_DISPLAY                     = DisplayOption.INLINE;
    public static final DisplayOption DEFAULT_USER_DESCRIPTION_DISPLAY          = DisplayOption.TOOLTIP;
    public static final int           DEFAULT_PNG_RESOLUTION                    = 200;
    public static final int           DEFAULT_TOOLTIP_TIMEOUT                   = 60;
    public static final LengthUnits   DEFAULT_DEFAULT_LENGTH_UNITS              = LengthUnits.FT_IN;
    public static final List<String>  DEFAULT_BLOCK_LAYOUT                      = List.of(CharacterSheet.REACTIONS_KEY, CharacterSheet.MELEE_KEY, CharacterSheet.RANGED_KEY, CharacterSheet.ADVANTAGES_KEY + " " + CharacterSheet.SKILLS_KEY, CharacterSheet.SPELLS_KEY, CharacterSheet.EQUIPMENT_KEY, CharacterSheet.OTHER_EQUIPMENT_KEY, CharacterSheet.NOTES_KEY);
    public static final Scales        DEFAULT_INITIAL_UI_SCALE                  = Scales.QUARTER_AGAIN_SIZE;
    public static final String        DEFAULT_DEFAULT_PLAYER_NAME               = System.getProperty("user.name", "");
    public static final String        DEFAULT_DEFAULT_PORTRAIT_PATH             = "!\000";
    public static final String        DEFAULT_DEFAULT_TECH_LEVEL                = "3";
    public static final WeightUnits   DEFAULT_DEFAULT_WEIGHT_UNITS              = WeightUnits.LB;

    public static final int MAX_RECENT_FILES        = 20;
    public static final int MINIMUM_TOOLTIP_TIMEOUT = 1;
    public static final int MAXIMUM_TOOLTIP_TIMEOUT = 9999;

    private static Preferences                      INSTANCE;
    private        Notifier                         mNotifier;
    private        UUID                             mID;
    private        Version                          mLastSeenGCSVersion;
    private        int                              mInitialPoints;
    private        int                              mToolTipTimeout;
    private        int                              mLibraryExplorerDividerPosition;
    private        List<String>                     mLibraryExplorerOpenRowKeys;
    private        DisplayOption                    mUserDescriptionDisplay;
    private        DisplayOption                    mModifiersDisplay;
    private        DisplayOption                    mNotesDisplay;
    private        Scales                           mInitialUIScale;
    private        LengthUnits                      mDefaultLengthUnits;
    private        WeightUnits                      mDefaultWeightUnits;
    private        List<String>                     mBlockLayout;
    private        List<Path>                       mRecentFiles;
    private        Path                             mLastDir;
    private        Map<String, PdfRef>              mPdfRefs;
    private        Map<String, String>              mKeyBindingOverrides;
    private        Map<String, Fonts.Info>          mFontInfo;
    private        Map<String, BaseWindow.Position> mBaseWindowPositions;
    private        PrintManager                     mDefaultPageSettings;
    private        String                           mGURPSCalculatorKey;
    private        String                           mDefaultPlayerName;
    private        String                           mDefaultTechLevel;
    private        String                           mDefaultPortraitPath;
    private        int                              mLastRecentFilesUpdateCounter;
    private        int                              mPNGResolution;
    private        boolean                          mIncludeUnspentPointsInTotal;
    private        boolean                          mBaseWillAndPerOn10;
    private        boolean                          mUseMultiplicativeModifiers;
    private        boolean                          mUseModifyingDicePlusAdds;
    private        boolean                          mUseKnowYourOwnStrength;
    private        boolean                          mUseReducedSwing;
    private        boolean                          mUseThrustEqualsSwingMinus2;
    private        boolean                          mUseSimpleMetricConversions;
    private        boolean                          mAutoNameNewCharacters;
    private        boolean                          mUseNativePrintDialogs;
    private        boolean                          mShowCollegeInSheetSpells;

    public static synchronized Preferences getInstance() {
        if (INSTANCE == null) {
            INSTANCE = new Preferences();
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

    private Preferences() {
        mNotifier = new Notifier();
        mID = UUID.randomUUID();
        mLastSeenGCSVersion = new Version(GCS.VERSION);
        Library.LIBRARIES.clear();
        mInitialPoints = DEFAULT_INITIAL_POINTS;
        mToolTipTimeout = DEFAULT_TOOLTIP_TIMEOUT;
        mLibraryExplorerDividerPosition = DEFAULT_LIBRARY_EXPLORER_DIVIDER_POSITION;
        mLibraryExplorerOpenRowKeys = new ArrayList<>();
        mUserDescriptionDisplay = DEFAULT_USER_DESCRIPTION_DISPLAY;
        mModifiersDisplay = DEFAULT_MODIFIERS_DISPLAY;
        mNotesDisplay = DEFAULT_NOTES_DISPLAY;
        mInitialUIScale = DEFAULT_INITIAL_UI_SCALE;
        mDefaultLengthUnits = DEFAULT_DEFAULT_LENGTH_UNITS;
        mDefaultWeightUnits = DEFAULT_DEFAULT_WEIGHT_UNITS;
        mBlockLayout = new ArrayList<>(DEFAULT_BLOCK_LAYOUT);
        mRecentFiles = new ArrayList<>();
        mLastDir = Paths.get(System.getProperty("user.home", ".")).normalize().toAbsolutePath();
        mGURPSCalculatorKey = "";
        mDefaultPlayerName = DEFAULT_DEFAULT_PLAYER_NAME;
        mDefaultTechLevel = DEFAULT_DEFAULT_TECH_LEVEL;
        mDefaultPortraitPath = DEFAULT_DEFAULT_PORTRAIT_PATH;
        mPNGResolution = DEFAULT_PNG_RESOLUTION;
        mPdfRefs = new HashMap<>();
        mFontInfo = new HashMap<>();
        mKeyBindingOverrides = new HashMap<>();
        mBaseWindowPositions = new HashMap<>();
        try {
            mDefaultPageSettings = new PrintManager(PageOrientation.PORTRAIT, 0.25, LengthUnits.IN);
        } catch (Exception exception) {
            mDefaultPageSettings = null;
        }
        mIncludeUnspentPointsInTotal = DEFAULT_INCLUDE_UNSPENT_POINTS_IN_TOTAL;
        mBaseWillAndPerOn10 = DEFAULT_BASE_WILL_AND_PER_ON_10;
        mUseMultiplicativeModifiers = DEFAULT_USE_MULTIPLICATIVE_MODIFIERS;
        mUseModifyingDicePlusAdds = DEFAULT_USE_MODIFYING_DICE_PLUS_ADDS;
        mUseKnowYourOwnStrength = DEFAULT_USE_KNOW_YOUR_OWN_STRENGTH;
        mUseReducedSwing = DEFAULT_USE_REDUCED_SWING;
        mUseThrustEqualsSwingMinus2 = DEFAULT_USE_THRUST_EQUALS_SWING_MINUS_2;
        mUseSimpleMetricConversions = DEFAULT_USE_SIMPLE_METRIC_CONVERSIONS;
        mAutoNameNewCharacters = DEFAULT_AUTO_NAME_NEW_CHARACTERS;
        mUseNativePrintDialogs = DEFAULT_USE_NATIVE_PRINT_DIALOGS;
        mShowCollegeInSheetSpells = DEFAULT_SHOW_COLLEGE_IN_SHEET_SPELLS;
        Path path = getPreferencesPath();
        if (Files.isReadable(path) && Files.isRegularFile(path)) {
            try (BufferedReader in = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
                JsonMap m = Json.asMap(Json.parse(in));
                if (!m.isEmpty()) {
                    int version = m.getInt(VERSION);
                    if (version >= MINIMUM_VERSION && version <= CURRENT_VERSION) {
                        mID = UUID.fromString(m.getStringWithDefault(ID, mID.toString()));
                        Version loadVersion;
                        if (version >= VERSION_AS_OF_GCS_4_18) {
                            loadVersion = new Version(m.getString(LAST_SEEN_GCS_VERSION));
                        } else {
                            loadVersion = new Version(m.getString(LAST_GCS_VERSION));
                        }
                        if (loadVersion.compareTo(mLastSeenGCSVersion) > 0) {
                            mLastSeenGCSVersion = loadVersion;
                        }
                        if (version >= VERSION_AS_OF_GCS_4_18) {
                            if (m.has(LIBRARIES)) {
                                JsonMap m2 = m.getMap(LIBRARIES);
                                for (String key : m2.keySet()) {
                                    Library.LIBRARIES.add(Library.fromJSON(key, m2.getMap(key)));
                                }
                            }
                        } else {
                            Library.MASTER.setPath(Paths.get(m.getStringWithDefault(MASTER_LIBRARY_PATH, Library.MASTER.getPathNoCreate().toString())));
                            Library.USER.setPath(Paths.get(m.getStringWithDefault(USER_LIBRARY_PATH, Library.USER.getPathNoCreate().toString())));
                        }
                        mInitialPoints = m.getIntWithDefault(INITIAL_POINTS, mInitialPoints);
                        mToolTipTimeout = m.getIntWithDefault(TOOLTIP_TIMEOUT, mToolTipTimeout);
                        if (m.has(LIBRARY_EXPLORER)) {
                            JsonMap m2 = m.getMap(LIBRARY_EXPLORER);
                            mLibraryExplorerDividerPosition = m2.getIntWithDefault(DIVIDER_POSITION, mLibraryExplorerDividerPosition);
                            JsonArray a      = m2.getArray(OPEN_ROW_KEYS);
                            int       length = a.size();
                            mLibraryExplorerOpenRowKeys = new ArrayList<>();
                            for (int i = 0; i < length; i++) {
                                mLibraryExplorerOpenRowKeys.add(a.getString(i));
                            }
                        }
                        mUserDescriptionDisplay = Enums.extract(m.getStringWithDefault(USER_DESCRIPTION_DISPLAY, ""), DisplayOption.values(), mUserDescriptionDisplay);
                        mModifiersDisplay = Enums.extract(m.getStringWithDefault(MODIFIERS_DISPLAY, ""), DisplayOption.values(), mModifiersDisplay);
                        mNotesDisplay = Enums.extract(m.getStringWithDefault(NOTES_DISPLAY, ""), DisplayOption.values(), mNotesDisplay);
                        mInitialUIScale = Enums.extract(m.getStringWithDefault(INITIAL_UI_SCALE, ""), Scales.values(), mInitialUIScale);
                        mDefaultLengthUnits = Enums.extract(m.getStringWithDefault(DEFAULT_LENGTH_UNITS, ""), LengthUnits.values(), mDefaultLengthUnits);
                        mDefaultWeightUnits = Enums.extract(m.getStringWithDefault(DEFAULT_WEIGHT_UNITS, ""), WeightUnits.values(), mDefaultWeightUnits);
                        if (m.has(BLOCK_LAYOUT)) {
                            JsonArray a      = m.getArray(BLOCK_LAYOUT);
                            int       length = a.size();
                            mBlockLayout = new ArrayList<>();
                            for (int i = 0; i < length; i++) {
                                mBlockLayout.add(a.getString(i));
                            }
                            if (version < VERSION_AS_OF_GCS_4_18) {
                                mBlockLayout.add(0, CharacterSheet.REACTIONS_KEY);
                            }
                        }
                        if (m.has(RECENT_FILES)) {
                            JsonArray a      = m.getArray(RECENT_FILES);
                            int       length = a.size();
                            mRecentFiles = new ArrayList<>();
                            for (int i = 0; i < length; i++) {
                                mRecentFiles.add(Paths.get(a.getString(i)).normalize().toAbsolutePath());
                            }
                        }
                        mLastDir = Paths.get(m.getStringWithDefault(LAST_DIR, mLastDir.toString())).normalize().toAbsolutePath();
                        if (m.has(PDF_REFS)) {
                            JsonMap m2 = m.getMap(PDF_REFS);
                            mPdfRefs = new HashMap<>();
                            for (String key : m2.keySet()) {
                                mPdfRefs.put(key, new PdfRef(m2.getMap(key)));
                            }
                        }
                        if (m.has(KEY_BINDINGS)) {
                            JsonMap m2 = m.getMap(KEY_BINDINGS);
                            mKeyBindingOverrides = new HashMap<>();
                            for (String key : m2.keySet()) {
                                mKeyBindingOverrides.put(key, m2.getString(key));
                            }
                        }
                        if (m.has(FONTS)) {
                            JsonMap m2 = m.getMap(FONTS);
                            mFontInfo = new HashMap<>();
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
                        mDefaultPortraitPath = m.getStringWithDefault(DEFAULT_PORTRAIT_PATH, mDefaultPortraitPath);
                        mPNGResolution = m.getIntWithDefault(PNG_RESOLUTION, mPNGResolution);
                        mIncludeUnspentPointsInTotal = m.getBooleanWithDefault(INCLUDE_UNSPENT_POINTS_IN_TOTAL, mIncludeUnspentPointsInTotal);
                        mBaseWillAndPerOn10 = m.getBooleanWithDefault(BASE_WILL_AND_PER_ON_10, mBaseWillAndPerOn10);
                        mUseMultiplicativeModifiers = m.getBooleanWithDefault(USE_MULTIPLICATIVE_MODIFIERS, mUseMultiplicativeModifiers);
                        mUseModifyingDicePlusAdds = m.getBooleanWithDefault(USE_MODIFYING_DICE_PLUS_ADDS, mUseModifyingDicePlusAdds);
                        mUseKnowYourOwnStrength = m.getBooleanWithDefault(USE_KNOW_YOUR_OWN_STRENGTH, mUseKnowYourOwnStrength);
                        mUseReducedSwing = m.getBooleanWithDefault(USE_REDUCED_SWING, mUseReducedSwing);
                        mUseThrustEqualsSwingMinus2 = m.getBooleanWithDefault(USE_THRUST_EQUALS_SWING_MINUS_2, mUseThrustEqualsSwingMinus2);
                        mUseSimpleMetricConversions = m.getBooleanWithDefault(USE_SIMPLE_METRIC_CONVERSIONS, mUseSimpleMetricConversions);
                        mAutoNameNewCharacters = m.getBooleanWithDefault(AUTO_NAME_NEW_CHARACTERS, mAutoNameNewCharacters);
                        mUseNativePrintDialogs = m.getBooleanWithDefault(USE_NATIVE_PRINT_DIALOGS, mUseNativePrintDialogs);
                        mShowCollegeInSheetSpells = m.getBooleanWithDefault(SHOW_COLLEGE_IN_SHEET_SPELLS, mShowCollegeInSheetSpells);
                        if (m.has(DEFAULT_PAGE_SETTINGS)) {
                            mDefaultPageSettings = new PrintManager(m.getMap(DEFAULT_PAGE_SETTINGS));
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

    public Notifier getNotifier() {
        return mNotifier;
    }

    public void startBatch() {
        mNotifier.startBatch();
    }

    public void endBatch() {
        mNotifier.endBatch();
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
                    w.keyValue(VERSION, CURRENT_VERSION);
                    w.keyValue(ID, mID.toString());
                    w.keyValue(LAST_SEEN_GCS_VERSION, mLastSeenGCSVersion.toString());
                    w.key(LIBRARIES);
                    w.startMap();
                    for (Library lib : Library.LIBRARIES) {
                        lib.toJSON(w);
                    }
                    w.endMap();
                    w.keyValue(MASTER_LIBRARY_PATH, Library.MASTER.getPathNoCreate().toString());
                    w.keyValue(USER_LIBRARY_PATH, Library.USER.getPathNoCreate().toString());
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
                    w.keyValue(USER_DESCRIPTION_DISPLAY, Enums.toId(mUserDescriptionDisplay));
                    w.keyValue(MODIFIERS_DISPLAY, Enums.toId(mModifiersDisplay));
                    w.keyValue(NOTES_DISPLAY, Enums.toId(mNotesDisplay));
                    w.keyValue(INITIAL_UI_SCALE, Enums.toId(mInitialUIScale));
                    w.keyValue(DEFAULT_LENGTH_UNITS, Enums.toId(mDefaultLengthUnits));
                    w.keyValue(DEFAULT_WEIGHT_UNITS, Enums.toId(mDefaultWeightUnits));
                    w.key(BLOCK_LAYOUT);
                    w.startArray();
                    for (String line : mBlockLayout) {
                        w.value(line);
                    }
                    w.endArray();
                    w.key(RECENT_FILES);
                    w.startArray();
                    for (Path p : mRecentFiles) {
                        w.value(p.toString());
                    }
                    w.endArray();
                    w.keyValue(LAST_DIR, mLastDir.toString());
                    w.key(PDF_REFS);
                    w.startMap();
                    for (Map.Entry<String, PdfRef> entry : mPdfRefs.entrySet()) {
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
                            w.startMap();
                            w.keyValue("x", info.mBounds.x);
                            w.keyValue("y", info.mBounds.y);
                            w.keyValue("width", info.mBounds.width);
                            w.keyValue("height", info.mBounds.height);
                            w.keyValue("last_updated", info.mLastUpdated);
                            w.endMap();
                        }
                    }
                    w.endMap();
                    w.keyValue(GURPS_CALCULATOR_KEY, mGURPSCalculatorKey);
                    w.keyValue(DEFAULT_PLAYER_NAME, mDefaultPlayerName);
                    w.keyValue(DEFAULT_TECH_LEVEL, mDefaultTechLevel);
                    w.keyValue(DEFAULT_PORTRAIT_PATH, mDefaultPortraitPath);
                    w.keyValue(PNG_RESOLUTION, mPNGResolution);
                    w.keyValue(INCLUDE_UNSPENT_POINTS_IN_TOTAL, mIncludeUnspentPointsInTotal);
                    w.keyValue(BASE_WILL_AND_PER_ON_10, mBaseWillAndPerOn10);
                    w.keyValue(USE_MULTIPLICATIVE_MODIFIERS, mUseMultiplicativeModifiers);
                    w.keyValue(USE_MODIFYING_DICE_PLUS_ADDS, mUseModifyingDicePlusAdds);
                    w.keyValue(USE_KNOW_YOUR_OWN_STRENGTH, mUseKnowYourOwnStrength);
                    w.keyValue(USE_REDUCED_SWING, mUseReducedSwing);
                    w.keyValue(USE_THRUST_EQUALS_SWING_MINUS_2, mUseThrustEqualsSwingMinus2);
                    w.keyValue(USE_SIMPLE_METRIC_CONVERSIONS, mUseSimpleMetricConversions);
                    w.keyValue(SHOW_COLLEGE_IN_SHEET_SPELLS, mShowCollegeInSheetSpells);
                    w.keyValue(AUTO_NAME_NEW_CHARACTERS, mAutoNameNewCharacters);
                    w.keyValue(USE_NATIVE_PRINT_DIALOGS, mUseNativePrintDialogs);
                    if (mDefaultPageSettings != null) {
                        w.key(DEFAULT_PAGE_SETTINGS);
                        mDefaultPageSettings.toJSON(w, LengthUnits.IN);
                    }
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

    public UUID getID() {
        return mID;
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

    public DisplayOption getUserDescriptionDisplay() {
        return mUserDescriptionDisplay;
    }

    public boolean setUserDescriptionDisplay(DisplayOption userDescriptionDisplay) {
        if (mUserDescriptionDisplay != userDescriptionDisplay) {
            mUserDescriptionDisplay = userDescriptionDisplay;
            mNotifier.notify(this, KEY_USER_DESCRIPTION_DISPLAY);
            return true;
        }
        return false;
    }

    public DisplayOption getModifiersDisplay() {
        return mModifiersDisplay;
    }

    public boolean setModifiersDisplay(DisplayOption modifiersDisplay) {
        if (mModifiersDisplay != modifiersDisplay) {
            mModifiersDisplay = modifiersDisplay;
            mNotifier.notify(this, KEY_MODIFIERS_DISPLAY);
            return true;
        }
        return false;
    }

    public DisplayOption getNotesDisplay() {
        return mNotesDisplay;
    }

    public boolean setNotesDisplay(DisplayOption notesDisplay) {
        if (mNotesDisplay != notesDisplay) {
            mNotesDisplay = notesDisplay;
            mNotifier.notify(this, KEY_NOTES_DISPLAY);
            return true;
        }
        return false;
    }

    public Scales getInitialUIScale() {
        return mInitialUIScale;
    }

    public void setInitialUIScale(Scales initialUIScale) {
        mInitialUIScale = initialUIScale;
    }

    public LengthUnits getDefaultLengthUnits() {
        return mDefaultLengthUnits;
    }

    public boolean setDefaultLengthUnits(LengthUnits defaultLengthUnits) {
        if (mDefaultLengthUnits != defaultLengthUnits) {
            mDefaultLengthUnits = defaultLengthUnits;
            mNotifier.notify(this, KEY_DEFAULT_LENGTH_UNITS);
            return true;
        }
        return false;
    }

    public WeightUnits getDefaultWeightUnits() {
        return mDefaultWeightUnits;
    }

    public boolean setDefaultWeightUnits(WeightUnits defaultWeightUnits) {
        if (mDefaultWeightUnits != defaultWeightUnits) {
            mDefaultWeightUnits = defaultWeightUnits;
            mNotifier.notify(this, KEY_DEFAULT_WEIGHT_UNITS);
            return true;
        }
        return false;
    }

    public List<String> getBlockLayout() {
        return mBlockLayout;
    }

    public boolean setBlockLayout(List<String> blockLayout) {
        if (!mBlockLayout.equals(blockLayout)) {
            mBlockLayout = new ArrayList<>(blockLayout);
            mNotifier.notify(this, KEY_BLOCK_LAYOUT);
            return true;
        }
        return false;
    }

    public static String linesToString(List<String> lines) {
        StringBuilder buffer = new StringBuilder();
        for (String line : lines) {
            if (buffer.length() != 0) {
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
        for (FileType fileType : FileType.OPENABLE) {
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

    public void setGURPSCalculatorKey(String GURPSCalculatorKey) {
        mGURPSCalculatorKey = GURPSCalculatorKey;
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

    public String getDefaultPortraitPath() {
        return mDefaultPortraitPath;
    }

    public void setDefaultPortraitPath(String defaultPortraitPath) {
        mDefaultPortraitPath = defaultPortraitPath;
    }

    public int getPNGResolution() {
        return mPNGResolution;
    }

    public void setPNGResolution(int PNGResolution) {
        mPNGResolution = PNGResolution;
    }

    public List<PdfRef> allPdfRefs(boolean requireExistence) {
        List<PdfRef> list = new ArrayList<>();
        for (String key : mPdfRefs.keySet()) {
            PdfRef ref = lookupPdfRef(key, requireExistence);
            if (ref != null) {
                list.add(ref);
            }
        }
        Collections.sort(list);
        return list;
    }

    public PdfRef lookupPdfRef(String id, boolean requireExistence) {
        PdfRef ref = mPdfRefs.get(id);
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

    public void putPdfRef(PdfRef ref) {
        mPdfRefs.put(ref.getId(), ref);
    }

    public void removePdfRef(PdfRef ref) {
        mPdfRefs.remove(ref.getId());
    }

    public void clearPdfRefs() {
        mPdfRefs = new HashMap<>();
    }

    public boolean arePdfRefsSetToDefault() {
        return mPdfRefs.isEmpty();
    }

    public Fonts.Info getFontInfo(String key) {
        return mFontInfo.get(key);
    }

    public void setFontInfo(String key, Fonts.Info fontInfo) {
        mFontInfo.put(key, fontInfo);
    }

    public String getKeyBindingOverride(String key) {
        return mKeyBindingOverrides.get(key);
    }

    public void setKeyBindingOverride(String key, String override) {
        if (override == null) {
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

    public PrintManager getDefaultPageSettings() {
        return mDefaultPageSettings;
    }

    public void setDefaultPageSettings(PrintManager defaultPageSettings) {
        mDefaultPageSettings = defaultPageSettings != null ? new PrintManager(defaultPageSettings) : null;
    }

    public boolean includeUnspentPointsInTotal() {
        return mIncludeUnspentPointsInTotal;
    }

    public boolean setIncludeUnspentPointsInTotal(boolean includeUnspentPointsInTotal) {
        if (mIncludeUnspentPointsInTotal != includeUnspentPointsInTotal) {
            mIncludeUnspentPointsInTotal = includeUnspentPointsInTotal;
            mNotifier.notify(this, KEY_INCLUDE_UNSPENT_POINTS_IN_TOTAL);
            return true;
        }
        return false;
    }

    /** @return Whether Will and Perception should be based on 10 rather than IQ. */
    public boolean baseWillAndPerOn10() {
        return mBaseWillAndPerOn10;
    }

    public boolean setBaseWillAndPerOn10(boolean baseWillAndPerOn10) {
        if (mBaseWillAndPerOn10 != baseWillAndPerOn10) {
            mBaseWillAndPerOn10 = baseWillAndPerOn10;
            mNotifier.notify(this, KEY_BASE_WILL_AND_PER_ON_10);
            return true;
        }
        return false;
    }

    /** @return Whether to show the college column in the sheet display. */
    public boolean showCollegeInSheetSpells() {
        return mShowCollegeInSheetSpells;
    }

    public boolean setShowCollegeInSheetSpells(boolean show) {
        if (mShowCollegeInSheetSpells != show) {
            mShowCollegeInSheetSpells = show;
            mNotifier.notify(this, KEY_SHOW_COLLEGE_IN_SHEET_SPELLS);
            return true;
        }
        return false;
    }

    /** @return Whether to use the multiplicative modifier rules from PW102. */
    public boolean useMultiplicativeModifiers() {
        return mUseMultiplicativeModifiers;
    }

    public boolean setUseMultiplicativeModifiers(boolean useMultiplicativeModifiers) {
        if (mUseMultiplicativeModifiers != useMultiplicativeModifiers) {
            mUseMultiplicativeModifiers = useMultiplicativeModifiers;
            mNotifier.notify(this, KEY_USE_MULTIPLICATIVE_MODIFIERS);
            return true;
        }
        return false;
    }

    /** @return Whether to use the dice modification rules from B269. */
    public boolean useModifyingDicePlusAdds() {
        return mUseModifyingDicePlusAdds;
    }

    public boolean setUseModifyingDicePlusAdds(boolean useModifyingDicePlusAdds) {
        if (mUseModifyingDicePlusAdds != useModifyingDicePlusAdds) {
            mUseModifyingDicePlusAdds = useModifyingDicePlusAdds;
            mNotifier.notify(this, KEY_USE_MODIFYING_DICE_PLUS_ADDS);
            return true;
        }
        return false;
    }

    /** @return Whether to use the Know Your Own Strength rules from PY83. */
    public boolean useKnowYourOwnStrength() {
        return mUseKnowYourOwnStrength;
    }

    public boolean setUseKnowYourOwnStrength(boolean useKnowYourOwnStrength) {
        if (mUseKnowYourOwnStrength != useKnowYourOwnStrength) {
            mUseKnowYourOwnStrength = useKnowYourOwnStrength;
            mNotifier.notify(this, KEY_USE_KNOW_YOUR_OWN_STRENGTH);
            return true;
        }
        return false;
    }

    /**
     * @return Whether to use the Adjusting Swing Damage rules from noschoolgrognard.blogspot.com.
     *         Reduces KYOS damages if used together.
     */
    public boolean useReducedSwing() {
        return mUseReducedSwing;
    }

    public boolean setUseReducedSwing(boolean useReducedSwing) {
        if (mUseReducedSwing != useReducedSwing) {
            mUseReducedSwing = useReducedSwing;
            mNotifier.notify(this, KEY_USE_REDUCED_SWING);
            return true;
        }
        return false;
    }

    /** @return Whether to set thrust damage to swing-2. */
    public boolean useThrustEqualsSwingMinus2() {
        return mUseThrustEqualsSwingMinus2;
    }

    public boolean setUseThrustEqualsSwingMinus2(boolean useThrustEqualsSwingMinus2) {
        if (mUseThrustEqualsSwingMinus2 != useThrustEqualsSwingMinus2) {
            mUseThrustEqualsSwingMinus2 = useThrustEqualsSwingMinus2;
            mNotifier.notify(this, KEY_USE_THRUST_EQUALS_SWING_MINUS_2);
            return true;
        }
        return false;
    }

    /** @return Whether to use the simple metric conversion rules from B9. */
    public boolean useSimpleMetricConversions() {
        return mUseSimpleMetricConversions;
    }

    public boolean setUseSimpleMetricConversions(boolean useSimpleMetricConversions) {
        if (mUseSimpleMetricConversions != useSimpleMetricConversions) {
            mUseSimpleMetricConversions = useSimpleMetricConversions;
            mNotifier.notify(this, KEY_USE_SIMPLE_METRIC_CONVERSIONS);
            return true;
        }
        return false;
    }

    /** @return Whether a new character should be automatically named. */
    public boolean autoNameNewCharacters() {
        return mAutoNameNewCharacters;
    }

    public void setAutoNameNewCharacters(boolean autoNameNewCharacters) {
        mAutoNameNewCharacters = autoNameNewCharacters;
    }

    public boolean useNativePrintDialogs() {
        return mUseNativePrintDialogs;
    }

    public void setUseNativePrintDialogs(boolean useNativePrintDialogs) {
        mUseNativePrintDialogs = useNativePrintDialogs;
    }
}
