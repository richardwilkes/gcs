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

package com.trollworks.gcs.body;

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.StdFileDialog;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.SafeFileUpdater;
import com.trollworks.gcs.utility.VersionException;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.awt.EventQueue;
import java.awt.Point;
import java.awt.event.ItemEvent;
import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JComboBox;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.ScrollPaneConstants;

public class HitLocationEditor extends JPanel {
    private static final String                JSON_TYPE_NAME = "hit_locations";
    public static final  int                   BUTTON_SIZE    = 14;
    private              HitLocationTablePanel mLocationsPanel;
    private              JScrollPane           mScroller;

    public HitLocationEditor(HitLocationTable locations, Runnable adjustCallback, String extraTitle) {
        super(new PrecisionLayout().setColumns(4).setMargins(0).setHorizontalSpacing(10));
        setOpaque(false);
        add(new Label(I18n.Text("Hit Locations") + extraTitle));
        add(new FontAwesomeButton("\uf56e", BUTTON_SIZE, I18n.Text("Export Hit Locations"), this::exportData));
        add(new FontAwesomeButton("\uf56f", BUTTON_SIZE, I18n.Text("Import Hit Locations"), this::importData));
        addStdChoicesCombo();
        mLocationsPanel = new HitLocationTablePanel(locations, adjustCallback);
        mScroller = new JScrollPane(mLocationsPanel, ScrollPaneConstants.VERTICAL_SCROLLBAR_AS_NEEDED, ScrollPaneConstants.HORIZONTAL_SCROLLBAR_AS_NEEDED);
        int minHeight = new HitLocationPanel(new HitLocation(new JsonMap()), null).getPreferredSize().height + 8;
        add(mScroller, new PrecisionLayoutData().setHorizontalSpan(4).setFillAlignment().setGrabSpace(true).setMinimumHeight(minHeight));
        EventQueue.invokeLater(() -> mScroller.getViewport().setViewPosition(new Point(0, 0)));
    }

    private void addStdChoicesCombo() {
        Choice       none    = new Choice(null);
        List<Choice> choices = new ArrayList<>();
        choices.add(none);
        for (HitLocationTable table : HitLocationTable.getStdTables()) {
            choices.add(new Choice(table.clone()));
        }
        JComboBox<Choice> combo = new JComboBox<>(choices.toArray(new Choice[0]));
        combo.setSelectedItem(choices.get(0));
        combo.addItemListener((evt) -> {
            if (evt.getStateChange() == ItemEvent.SELECTED) {
                Choice choice = (Choice) evt.getItem();
                if (choice.mTable != null) {
                    reset(choice.mTable);
                    combo.setSelectedItem(none);
                }
            }
        });
        add(combo);
    }

    public void reset(HitLocationTable locations) {
        mLocationsPanel.reset(locations.clone());
        mLocationsPanel.getAdjustCallback().run();
        revalidate();
        EventQueue.invokeLater(() -> mScroller.getViewport().setViewPosition(new Point(0, 0)));
    }

    private void importData() {
        Path path = StdFileDialog.showOpenDialog(this, I18n.Text("Import Hit Locations…"),
                FileType.HIT_LOCATIONS.getFilter());
        if (path != null) {
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
                reset(new HitLocationTable(m.getMap(JSON_TYPE_NAME)));
            } catch (IOException ioe) {
                Log.error(ioe);
                WindowUtils.showError(this, I18n.Text("Unable to import Hit Locations."));
            }
        }
    }

    private void exportData() {
        Path path = StdFileDialog.showSaveDialog(this, I18n.Text("Export Hit Locations As…"),
                Preferences.getInstance().getLastDir().resolve(I18n.Text("hit_locations")),
                FileType.HIT_LOCATIONS.getFilter());
        if (path != null) {
            SafeFileUpdater transaction = new SafeFileUpdater();
            transaction.begin();
            try {
                File transactionFile = transaction.getTransactionFile(path.toFile());
                try (JsonWriter w = new JsonWriter(new BufferedWriter(new FileWriter(transactionFile, StandardCharsets.UTF_8)), "\t")) {
                    w.startMap();
                    w.keyValue(DataFile.TYPE, JSON_TYPE_NAME);
                    w.keyValue(DataFile.VERSION, DataFile.CURRENT_VERSION);
                    w.key(JSON_TYPE_NAME);
                    mLocationsPanel.getHitLocations().toJSON(w, null);
                    w.endMap();
                }
                transaction.commit();
            } catch (Exception exception) {
                Log.error(exception);
                transaction.abort();
                WindowUtils.showError(this, I18n.Text("Unable to export Hit Locations."));
            }
        }
    }

    private static class Choice {
        HitLocationTable mTable;

        Choice(HitLocationTable table) {
            mTable = table;
        }

        @Override
        public String toString() {
            return mTable != null ? mTable.getName() : I18n.Text("Select a Standard Table");
        }
    }
}
