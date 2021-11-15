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

import com.trollworks.gcs.body.BodyTypePanel;
import com.trollworks.gcs.body.HitLocationTable;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataChangeListener;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.SafeFileUpdater;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Window;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

/** A window for editing body type settings. */
public final class BodyTypeSettingsWindow extends SettingsWindow<HitLocationTable> implements DataChangeListener {
    private static final Map<UUID, BodyTypeSettingsWindow> INSTANCES = new HashMap<>();

    private GURPSCharacter mCharacter;
    private BodyTypePanel  mLocationsPanel;
    private boolean        mUpdatePending;

    /** Displays the hit location settings window. */
    public static void display(GURPSCharacter gchar) {
        if (!UIUtilities.inModalState()) {
            BodyTypeSettingsWindow wnd;
            synchronized (INSTANCES) {
                UUID key = gchar == null ? null : gchar.getID();
                wnd = INSTANCES.get(key);
                if (wnd == null) {
                    wnd = new BodyTypeSettingsWindow(gchar);
                    INSTANCES.put(key, wnd);
                }
            }
            wnd.setVisible(true);
        }
    }

    /** Closes the HitLocationSettingsWindow for the given character if it is open. */
    public static void closeFor(GURPSCharacter gchar) {
        for (Window window : Window.getWindows()) {
            if (window.isShowing() && window instanceof BodyTypeSettingsWindow wnd) {
                if (wnd.mCharacter == gchar) {
                    wnd.attemptClose();
                }
            }
        }
    }

    private static String createTitle(GURPSCharacter gchar) {
        return gchar == null ? I18n.text("Default Body Type") : String.format(I18n.text("Body Type for %s"), gchar.getProfile().getName());
    }

    private BodyTypeSettingsWindow(GURPSCharacter gchar) {
        super(createTitle(gchar));
        mCharacter = gchar;
        if (mCharacter != null) {
            mCharacter.addChangeListener(this);
            Settings.getInstance().addChangeListener(this);
        }
        fill();
    }

    @Override
    protected void preDispose() {
        synchronized (INSTANCES) {
            INSTANCES.remove(mCharacter == null ? null : mCharacter.getID());
        }
        if (mCharacter != null) {
            mCharacter.removeChangeListener(this);
            Settings.getInstance().removeChangeListener(this);
        }
    }

    @Override
    protected Panel createContent() {
        mLocationsPanel = new BodyTypePanel(SheetSettings.get(mCharacter).getHitLocations(), () -> {
            SheetSettings.get(mCharacter).getHitLocations().update();
            if (mCharacter == null) {
                Settings.getInstance().notifyOfChange();
            } else {
                mCharacter.notifyOfChange();
            }
            adjustResetButton();
        });
        return mLocationsPanel;
    }

    @Override
    public void establishSizing() {
        Dimension min = getMinimumSize();
        setMinimumSize(new Dimension(Math.max(min.width, 600), min.height));
    }

    @Override
    protected boolean shouldResetBeEnabled() {
        HitLocationTable prefsLocations = Settings.getInstance().getSheetSettings().getHitLocations();
        if (mCharacter == null) {
            return !prefsLocations.equals(HitLocationTable.createHumanoidTable());
        }
        return !mCharacter.getSheetSettings().getHitLocations().equals(prefsLocations);
    }

    @Override
    protected String getResetButtonTooltip() {
        return mCharacter == null ? super.getResetButtonTooltip() : I18n.text("Reset to Global Defaults");
    }

    @Override
    protected HitLocationTable getResetData() {
        return mCharacter == null ? HitLocationTable.createHumanoidTable() :
                Settings.getInstance().getSheetSettings().getHitLocations();
    }

    @Override
    protected void doResetTo(HitLocationTable data) {
        mLocationsPanel.reset(data.clone());
        mLocationsPanel.getAdjustCallback().run();
        revalidate();
        repaint();
    }

    @Override
    protected Dirs getDir() {
        return Dirs.SETTINGS;
    }

    @Override
    protected FileType getFileType() {
        return FileType.BODY_SETTINGS;
    }

    @Override
    protected HitLocationTable createSettingsFrom(Path path) throws IOException {
        return new HitLocationTable(path);
    }

    @Override
    protected void exportSettingsTo(Path path) throws IOException {
        SafeFileUpdater trans = new SafeFileUpdater();
        trans.begin();
        try {
            Files.createDirectories(path.getParent());
            File transactionFile = trans.getTransactionFile(path.toFile());
            try (JsonWriter w = new JsonWriter(new BufferedWriter(new FileWriter(transactionFile, StandardCharsets.UTF_8)), "\t")) {
                w.startMap();
                w.keyValue(DataFile.VERSION, DataFile.CURRENT_VERSION);
                w.key(SheetSettings.KEY_HIT_LOCATIONS);
                mLocationsPanel.getHitLocations().toJSON(w, null);
                w.endMap();
            }
        } catch (IOException ioe) {
            trans.abort();
            throw ioe;
        }
        trans.commit();
    }

    @Override
    public void dataWasChanged() {
        if (!mUpdatePending) {
            mUpdatePending = true;
            EventQueue.invokeLater(() -> {
                setTitle(createTitle(mCharacter));
                adjustResetButton();
                mUpdatePending = false;
            });
        }
    }
}
