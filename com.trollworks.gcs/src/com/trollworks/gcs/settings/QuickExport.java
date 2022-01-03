/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.settings;

import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.menu.file.ExportToGCalcCommand;
import com.trollworks.gcs.menu.file.ExportToPNGCommand;
import com.trollworks.gcs.menu.file.ExportToTextTemplateCommand;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.nio.file.Path;

public class QuickExport implements Comparable<QuickExport> {
    private static final String KEY_TEMPLATE_PATH   = "template_path";
    private static final String KEY_EXPORT_PATH     = "export_path";
    private static final String KEY_LAST_USED       = "last_used";
    public static final  String GCALC_EXPORT_MARKER = "::gcalc::";
    public static final  String PNG_EXPORT_MARKER   = "::png::";
    private              String mTemplatePath;
    private              String mExportPath;
    private              String mKey; // not part of the json
    private              long   mLastUsed;

    /** Create a new QuickExport for GCalc. */
    public QuickExport() {
        mTemplatePath = GCALC_EXPORT_MARKER;
        mExportPath = "";
        mLastUsed = System.currentTimeMillis();
    }

    /** Create a new QuickExport for export to PNG. */
    public QuickExport(Path exportPath) {
        mTemplatePath = PNG_EXPORT_MARKER;
        mExportPath = exportPath.toAbsolutePath().toString();
        mLastUsed = System.currentTimeMillis();
    }

    /** Create a new QuickExport for export to output template. */
    public QuickExport(Path templatePath, Path exportPath) {
        mTemplatePath = templatePath.toAbsolutePath().toString();
        mExportPath = exportPath.toAbsolutePath().toString();
        mLastUsed = System.currentTimeMillis();
    }

    public QuickExport(JsonMap m) {
        mTemplatePath = m.getString(KEY_TEMPLATE_PATH);
        mExportPath = m.getString(KEY_EXPORT_PATH);
        mLastUsed = m.getLongWithDefault(KEY_LAST_USED, System.currentTimeMillis());
    }

    public void toJSON(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(KEY_TEMPLATE_PATH, mTemplatePath);
        w.keyValueNot(KEY_EXPORT_PATH, mExportPath, "");
        w.keyValue(KEY_LAST_USED, mLastUsed);
        w.endMap();
    }

    public String getTemplatePath() {
        return mTemplatePath;
    }

    public String getExportPath() {
        return mExportPath;
    }

    public boolean isGCalcExport() {
        return GCALC_EXPORT_MARKER.equals(mTemplatePath);
    }

    public boolean isPNGExport() {
        return PNG_EXPORT_MARKER.equals(mTemplatePath);
    }

    public boolean isValid() {
        if (isGCalcExport()) {
            return !Settings.getInstance().getGeneralSettings().getGCalcKey().isBlank();
        }
        if (!isPNGExport() && (mTemplatePath.isBlank() || !Path.of(mTemplatePath).getParent().toFile().isDirectory())) {
            return false;
        }
        return !mExportPath.isBlank() && Path.of(mExportPath).getParent().toFile().isDirectory();
    }

    public void export(SheetDockable dockable) {
        if (isGCalcExport()) {
            ExportToGCalcCommand.performExport(dockable);
        } else if (isPNGExport()) {
            ExportToPNGCommand.performExport(dockable, Path.of(mExportPath));
        } else {
            ExportToTextTemplateCommand.performExport(dockable, Path.of(mTemplatePath), Path.of(mExportPath));
        }
    }

    @Override
    public int compareTo(QuickExport other) {
        int result = Long.compare(other.mLastUsed, mLastUsed); // Reversed
        if (result == 0) {
            result = Integer.compare(hashCode(), other.hashCode());
        }
        return result;
    }

    protected String getKey() {
        return mKey;
    }

    protected void setKey(String key) {
        mKey = key;
    }
}
