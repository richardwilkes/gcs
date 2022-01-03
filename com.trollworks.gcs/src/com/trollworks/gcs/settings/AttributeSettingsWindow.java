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

import com.trollworks.gcs.attribute.Attribute;
import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.attribute.AttributeListPanel;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataChangeListener;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.FontIconButton;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.Toolbar;
import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.SafeFileUpdater;
import com.trollworks.gcs.utility.VersionException;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Window;
import java.io.BufferedReader;
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

/** A window for editing attribute settings. */
public final class AttributeSettingsWindow extends SettingsWindow<Map<String, AttributeDef>> implements DataChangeListener {
    private static final Map<UUID, AttributeSettingsWindow> INSTANCES = new HashMap<>();

    private GURPSCharacter     mCharacter;
    private AttributeListPanel mListPanel;
    private boolean            mUpdatePending;

    /** Displays the attribute settings window. */
    public static void display(GURPSCharacter gchar) {
        if (!UIUtilities.inModalState()) {
            AttributeSettingsWindow wnd;
            synchronized (INSTANCES) {
                UUID key = gchar == null ? null : gchar.getID();
                wnd = INSTANCES.get(key);
                if (wnd == null) {
                    wnd = new AttributeSettingsWindow(gchar);
                    INSTANCES.put(key, wnd);
                }
            }
            wnd.setVisible(true);
        }
    }

    /** Closes the AttributeSettingsWindow for the given character if it is open. */
    public static void closeFor(GURPSCharacter gchar) {
        for (Window window : Window.getWindows()) {
            if (window.isShowing() && window instanceof AttributeSettingsWindow wnd) {
                if (wnd.mCharacter == gchar) {
                    wnd.attemptClose();
                }
            }
        }
    }

    private static String createTitle(GURPSCharacter gchar) {
        return gchar == null ? I18n.text("Default Attributes") : String.format(I18n.text("Attributes for %s"), gchar.getProfile().getName());
    }

    private AttributeSettingsWindow(GURPSCharacter gchar) {
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
    protected void addToToolBar(Toolbar toolbar) {
        toolbar.add(new FontIconButton(FontAwesome.PLUS_CIRCLE, I18n.text("Add Attribute"),
                (b) -> mListPanel.addAttribute()));
        super.addToToolBar(toolbar);
    }

    @Override
    protected Panel createContent() {
        mListPanel = new AttributeListPanel(SheetSettings.get(mCharacter).getAttributes(), () -> {
            adjustResetButton();
            if (mCharacter == null) {
                Settings.getInstance().notifyOfChange();
            } else {
                mCharacter.notifyOfChange();
            }
        });
        return mListPanel;
    }

    @Override
    public void establishSizing() {
        Dimension min = getMinimumSize();
        setMinimumSize(new Dimension(Math.max(min.width, 600), Math.max(min.height, 100)));
    }

    @Override
    protected boolean shouldResetBeEnabled() {
        Map<String, AttributeDef> prefsAttributes = Settings.getInstance().getSheetSettings().getAttributes();
        if (mCharacter != null) {
            Map<String, Attribute> oldAttributes = mCharacter.getAttributes();
            Map<String, Attribute> newAttributes = new HashMap<>();
            for (String key : mCharacter.getSheetSettings().getAttributes().keySet()) {
                Attribute attribute = oldAttributes.get(key);
                newAttributes.put(key, attribute != null ? attribute : new Attribute(key));
            }
            if (!oldAttributes.equals(newAttributes)) {
                oldAttributes.clear();
                oldAttributes.putAll(newAttributes);
                mCharacter.notifyOfChange();
            }
        }
        mListPanel.adjustButtons();
        if (mCharacter == null) {
            return !prefsAttributes.equals(AttributeDef.createStandardAttributes());
        }
        return !mCharacter.getSheetSettings().getAttributes().equals(prefsAttributes);
    }

    @Override
    protected String getResetButtonTooltip() {
        return mCharacter == null ? super.getResetButtonTooltip() : I18n.text("Reset to Global Defaults");
    }

    @Override
    protected Map<String, AttributeDef> getResetData() {
        if (mCharacter == null) {
            return AttributeDef.createStandardAttributes();
        }
        return AttributeDef.cloneMap(Settings.getInstance().getSheetSettings().getAttributes());
    }

    @Override
    protected void doResetTo(Map<String, AttributeDef> attributes) {
        mListPanel.reset(attributes);
        mListPanel.getAdjustCallback().run();
        revalidate();
        repaint();
    }

    @Override
    protected Dirs getDir() {
        return Dirs.SETTINGS;
    }

    @Override
    protected FileType getFileType() {
        return FileType.ATTRIBUTE_SETTINGS;
    }

    @Override
    protected Map<String, AttributeDef> createSettingsFrom(Path path) throws IOException {
        try (BufferedReader in = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            JsonMap m       = Json.asMap(Json.parse(in));
            int     version = m.getInt(DataFile.VERSION);
            if (version > DataFile.CURRENT_VERSION) {
                throw VersionException.createTooNew();
            }
            String key = SheetSettings.KEY_ATTRIBUTES;
            if (!m.has(SheetSettings.KEY_ATTRIBUTES)) {
                key = "attribute_settings"; // older key
            }
            if (!m.has(key)) {
                throw new IOException("invalid data type");
            }
            return AttributeDef.load(m.getArray(key));
        }
    }

    @Override
    protected void exportSettingsTo(Path path) throws IOException {
        SafeFileUpdater trans = new SafeFileUpdater();
        trans.begin();
        try {
            Files.createDirectories(path.getParent());
            File file = trans.getTransactionFile(path.toFile());
            try (JsonWriter w = new JsonWriter(new BufferedWriter(new FileWriter(file, StandardCharsets.UTF_8)), "\t")) {
                w.startMap();
                w.keyValue(Settings.VERSION, DataFile.CURRENT_VERSION);
                w.key(SheetSettings.KEY_ATTRIBUTES);
                AttributeDef.writeOrdered(w, mListPanel.getAttributes());
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
