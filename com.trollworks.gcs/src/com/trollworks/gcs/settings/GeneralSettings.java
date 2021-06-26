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

import com.trollworks.gcs.datafile.ChangeableData;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.ui.scale.Scales;
import com.trollworks.gcs.utility.SafeFileUpdater;
import com.trollworks.gcs.utility.json.Json;
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
import javax.swing.ToolTipManager;

public final class GeneralSettings {
    private static final String DEPRECATED_KEY_AUTO_NAME_NEW_CHARACTERS = "auto_name_new_characters"; // March 21, 2021
    private static final String DEPRECATED_KEY_PNG_RESOLUTION           = "png_resolution"; // June 6, 2021

    private static final String KEY_AUTO_FILL_PROFILE   = "auto_fill_profile";
    private static final String KEY_DEFAULT_PLAYER_NAME = "default_player_name";
    private static final String KEY_DEFAULT_TECH_LEVEL  = "default_tech_level";
    private static final String KEY_GCALC               = "gurps_calculator_key";
    private static final String KEY_IMAGE_RESOLUTION    = "image_resolution";
    private static final String KEY_UNSPENT_POINTS      = "include_unspent_points_in_total";
    private static final String KEY_INITIAL_POINTS      = "initial_points";
    private static final String KEY_INITIAL_UI_SCALE    = "initial_ui_scale";
    private static final String KEY_TOOLTIP_TIMEOUT     = "tooltip_timeout";

    private Scales  mInitialUIScale;
    private String  mDefaultPlayerName;
    private String  mDefaultTechLevel;
    private String  mGCalcKey;
    private int     mInitialPoints;
    private int     mToolTipTimeout;
    private int     mImageResolution;
    private boolean mAutoFillProfile;
    private boolean mIncludeUnspentPointsInTotal;

    public GeneralSettings() {
        mInitialUIScale = Scales.QUARTER_AGAIN_SIZE;
        mDefaultPlayerName = System.getProperty("user.name", "");
        mDefaultTechLevel = "3";
        mGCalcKey = "";
        mInitialPoints = 250;
        mToolTipTimeout = 60;
        mImageResolution = 200;
        mAutoFillProfile = true;
        mIncludeUnspentPointsInTotal = true;
    }

    public GeneralSettings(Path path) throws IOException {
        this();
        try (BufferedReader in = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            JsonMap m = Json.asMap(Json.parse(in));
            if (!m.isEmpty()) {
                int version = m.getInt(Settings.VERSION);
                if (version >= Settings.MINIMUM_VERSION && version <= DataFile.CURRENT_VERSION) {
                    load(m.getMap(Settings.GENERAL));
                }
            }
        }
    }

    public GeneralSettings(JsonMap m) {
        this();
        load(m);
    }

    public void copyFrom(GeneralSettings other) {
        mInitialUIScale = other.mInitialUIScale;
        mDefaultPlayerName = other.mDefaultPlayerName;
        mDefaultTechLevel = other.mDefaultTechLevel;
        mGCalcKey = other.mGCalcKey;
        mInitialPoints = other.mInitialPoints;
        mToolTipTimeout = other.mToolTipTimeout;
        mImageResolution = other.mImageResolution;
        mAutoFillProfile = other.mAutoFillProfile;
        mIncludeUnspentPointsInTotal = other.mIncludeUnspentPointsInTotal;
    }

    private void load(JsonMap m) {
        mInitialUIScale = Enums.extract(m.getStringWithDefault(KEY_INITIAL_UI_SCALE, ""), Scales.values(), mInitialUIScale);
        mDefaultPlayerName = m.getStringWithDefault(KEY_DEFAULT_PLAYER_NAME, mDefaultPlayerName);
        mDefaultTechLevel = m.getStringWithDefault(KEY_DEFAULT_TECH_LEVEL, mDefaultTechLevel);
        mGCalcKey = m.getStringWithDefault(KEY_GCALC, mGCalcKey);
        mInitialPoints = m.getIntWithDefault(KEY_INITIAL_POINTS, mInitialPoints);
        mToolTipTimeout = m.getIntWithDefault(KEY_TOOLTIP_TIMEOUT, mToolTipTimeout);
        if (m.has(DEPRECATED_KEY_PNG_RESOLUTION)) {
            mImageResolution = m.getIntWithDefault(DEPRECATED_KEY_PNG_RESOLUTION, mImageResolution);
        } else {
            mImageResolution = m.getIntWithDefault(KEY_IMAGE_RESOLUTION, mImageResolution);
        }
        if (m.has(DEPRECATED_KEY_AUTO_NAME_NEW_CHARACTERS)) {
            mAutoFillProfile = m.getBooleanWithDefault(DEPRECATED_KEY_AUTO_NAME_NEW_CHARACTERS, mAutoFillProfile);
        } else {
            mAutoFillProfile = m.getBooleanWithDefault(KEY_AUTO_FILL_PROFILE, mAutoFillProfile);
        }
        mIncludeUnspentPointsInTotal = m.getBooleanWithDefault(KEY_UNSPENT_POINTS, mIncludeUnspentPointsInTotal);
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
                w.key(Settings.GENERAL);
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
        w.keyValue(KEY_INITIAL_UI_SCALE, Enums.toId(mInitialUIScale));
        w.keyValue(KEY_DEFAULT_PLAYER_NAME, mDefaultPlayerName);
        w.keyValue(KEY_DEFAULT_TECH_LEVEL, mDefaultTechLevel);
        w.keyValue(KEY_GCALC, mGCalcKey);
        w.keyValue(KEY_INITIAL_POINTS, mInitialPoints);
        w.keyValue(KEY_TOOLTIP_TIMEOUT, mToolTipTimeout);
        w.keyValue(KEY_IMAGE_RESOLUTION, mImageResolution);
        w.keyValue(KEY_AUTO_FILL_PROFILE, mAutoFillProfile);
        w.keyValue(KEY_UNSPENT_POINTS, mIncludeUnspentPointsInTotal);
        w.endMap();
    }

    public Scales getInitialUIScale() {
        return mInitialUIScale;
    }

    public void setInitialUIScale(Scales initialUIScale) {
        mInitialUIScale = initialUIScale;
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

    public String getGCalcKey() {
        return mGCalcKey;
    }

    public void setGCalcKey(String key) {
        mGCalcKey = key;
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
            updateToolTipTimeout();
        }
    }

    public void updateToolTipTimeout() {
        if (!GraphicsEnvironment.isHeadless()) {
            ToolTipManager.sharedInstance().setDismissDelay(mToolTipTimeout * 1000);
        }
    }

    public int getImageResolution() {
        return mImageResolution;
    }

    public void setImageResolution(int resolution) {
        mImageResolution = resolution;
    }

    public boolean autoFillProfile() {
        return mAutoFillProfile;
    }

    public void setAutoFillProfile(boolean autoFillProfile) {
        mAutoFillProfile = autoFillProfile;
    }

    public boolean includeUnspentPointsInTotal() {
        return mIncludeUnspentPointsInTotal;
    }

    public void setIncludeUnspentPointsInTotal(boolean includeUnspentPointsInTotal, ChangeableData notifier) {
        if (mIncludeUnspentPointsInTotal != includeUnspentPointsInTotal) {
            mIncludeUnspentPointsInTotal = includeUnspentPointsInTotal;
            if (notifier != null) {
                notifier.notifyOfChange();
            }
        }
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }
        GeneralSettings that = (GeneralSettings) other;
        if (mInitialPoints != that.mInitialPoints) {
            return false;
        }
        if (mToolTipTimeout != that.mToolTipTimeout) {
            return false;
        }
        if (mImageResolution != that.mImageResolution) {
            return false;
        }
        if (mAutoFillProfile != that.mAutoFillProfile) {
            return false;
        }
        if (mIncludeUnspentPointsInTotal != that.mIncludeUnspentPointsInTotal) {
            return false;
        }
        if (mInitialUIScale != that.mInitialUIScale) {
            return false;
        }
        if (!mDefaultPlayerName.equals(that.mDefaultPlayerName)) {
            return false;
        }
        if (!mDefaultTechLevel.equals(that.mDefaultTechLevel)) {
            return false;
        }
        return mGCalcKey.equals(that.mGCalcKey);
    }

    @Override
    public int hashCode() {
        int result = mInitialUIScale.hashCode();
        result = 31 * result + mDefaultPlayerName.hashCode();
        result = 31 * result + mDefaultTechLevel.hashCode();
        result = 31 * result + mGCalcKey.hashCode();
        result = 31 * result + mInitialPoints;
        result = 31 * result + mToolTipTimeout;
        result = 31 * result + mImageResolution;
        result = 31 * result + (mAutoFillProfile ? 1 : 0);
        result = 31 * result + (mIncludeUnspentPointsInTotal ? 1 : 0);
        return result;
    }
}
