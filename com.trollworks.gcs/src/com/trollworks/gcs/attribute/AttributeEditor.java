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

package com.trollworks.gcs.attribute;

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
import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Map;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.ScrollPaneConstants;

public class AttributeEditor extends JPanel {
    private static final int                JSON_VERSION   = 1;
    private static final String             JSON_TYPE_NAME = "attribute_settings";
    public static final  int                BUTTON_SIZE    = 14;
    private              AttributeListPanel mListPanel;
    private              JScrollPane        mScroller;

    public AttributeEditor(Map<String, AttributeDef> attributes, Runnable adjustCallback) {
        super(new PrecisionLayout().setColumns(4).setMargins(0).setHorizontalSpacing(10));
        setOpaque(false);
        add(new Label(I18n.Text("Attributes")));
        add(new FontAwesomeButton("\uf055", BUTTON_SIZE, I18n.Text("Add Attribute"), () -> mListPanel.addAttribute()));
        add(new FontAwesomeButton("\uf56e", BUTTON_SIZE, I18n.Text("Export Attribute Settings"), this::exportData));
        add(new FontAwesomeButton("\uf56f", BUTTON_SIZE, I18n.Text("Import Attribute Settings"), this::importData));
        mListPanel = new AttributeListPanel(attributes, adjustCallback);
        mScroller = new JScrollPane(mListPanel, ScrollPaneConstants.VERTICAL_SCROLLBAR_AS_NEEDED, ScrollPaneConstants.HORIZONTAL_SCROLLBAR_AS_NEEDED);
        int minHeight = new AttributePanel(null, new AttributeDef(new JsonMap(), 0), null).getPreferredSize().height + 8;
        add(mScroller, new PrecisionLayoutData().setHorizontalSpan(4).setFillAlignment().setGrabSpace(true).setMinimumHeight(minHeight));
        EventQueue.invokeLater(() -> mScroller.getViewport().setViewPosition(new Point(0, 0)));
    }

    public void reset(Map<String, AttributeDef> attributes) {
        mListPanel.reset(attributes);
        EventQueue.invokeLater(() -> mScroller.getViewport().setViewPosition(new Point(0, 0)));
    }

    private void importData() {
        Path path = StdFileDialog.showOpenDialog(this, I18n.Text("Import Attribute Settings…"),
                FileType.ATTRIBUTE_SETTINGS.getFilter());
        if (path != null) {
            try (BufferedReader fileReader = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
                JsonMap   m     = Json.asMap(Json.parse(fileReader));
                LoadState state = new LoadState();
                state.mDataFileVersion = m.getInt(LoadState.ATTRIBUTE_VERSION);
                if (state.mDataFileVersion > JSON_VERSION) {
                    throw VersionException.createTooNew();
                }
                if (!JSON_TYPE_NAME.equals(m.getString(DataFile.KEY_TYPE))) {
                    throw new IOException("invalid data type");
                }
                mListPanel.reset(AttributeDef.load(m.getArray(JSON_TYPE_NAME)));
                mListPanel.getAdjustCallback().run();
                revalidate();
                EventQueue.invokeLater(() -> mScroller.getViewport().setViewPosition(new Point(0, 0)));
            } catch (IOException ioe) {
                Log.error(ioe);
                WindowUtils.showError(this, I18n.Text("Unable to import attribute settings."));
            }
        }
    }

    private void exportData() {
        Path path = StdFileDialog.showSaveDialog(this, I18n.Text("Export Attribute Settings As…"),
                Preferences.getInstance().getLastDir().resolve(I18n.Text("attribute_settings")),
                FileType.ATTRIBUTE_SETTINGS.getFilter());
        if (path != null) {
            SafeFileUpdater transaction = new SafeFileUpdater();
            transaction.begin();
            try {
                File transactionFile = transaction.getTransactionFile(path.toFile());
                try (JsonWriter w = new JsonWriter(new BufferedWriter(new FileWriter(transactionFile, StandardCharsets.UTF_8)), "\t")) {
                    w.startMap();
                    w.keyValue(DataFile.KEY_TYPE, JSON_TYPE_NAME);
                    w.keyValue(LoadState.ATTRIBUTE_VERSION, JSON_VERSION);
                    w.key(JSON_TYPE_NAME);
                    AttributeDef.writeOrdered(w, mListPanel.getAttributes());
                    w.endMap();
                }
                transaction.commit();
            } catch (Exception exception) {
                Log.error(exception);
                transaction.abort();
                WindowUtils.showError(this, I18n.Text("Unable to export attribute settings."));
            }
        }
    }
}
