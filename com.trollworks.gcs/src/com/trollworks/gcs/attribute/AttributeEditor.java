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
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.SafeFileUpdater;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.awt.EventQueue;
import java.awt.Point;
import java.awt.event.ItemEvent;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Path;
import java.util.List;
import java.util.Map;
import javax.swing.JComboBox;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.ScrollPaneConstants;

public class AttributeEditor extends JPanel {
    private static final String             JSON_TYPE_NAME = "attribute_settings";
    private              AttributeListPanel mListPanel;
    private              JScrollPane        mScroller;

    public AttributeEditor(Map<String, AttributeDef> attributes, Runnable adjustCallback, String extraTitle) {
        super(new PrecisionLayout().setColumns(5).setMargins(0).setHorizontalSpacing(10));
        setOpaque(false);
        add(new Label(I18n.Text("Attributes") + extraTitle));
        add(new FontAwesomeButton("\uf055", I18n.Text("Add Attribute"), () -> mListPanel.addAttribute()));
        add(new FontAwesomeButton("\uf56e", I18n.Text("Export Attribute Settings"), this::exportData));
        add(new FontAwesomeButton("\uf56f", I18n.Text("Import Attribute Settings"), this::importData));
        addStdChoicesCombo();
        mListPanel = new AttributeListPanel(attributes, adjustCallback);
        mScroller = new JScrollPane(mListPanel, ScrollPaneConstants.VERTICAL_SCROLLBAR_AS_NEEDED, ScrollPaneConstants.HORIZONTAL_SCROLLBAR_AS_NEEDED);
        int minHeight = new AttributePanel(null, new AttributeDef(new JsonMap(), 0), null).getPreferredSize().height + 8;
        add(mScroller, new PrecisionLayoutData().setHorizontalSpan(5).setFillAlignment().setGrabSpace(true).setMinimumHeight(minHeight));
        EventQueue.invokeLater(() -> mScroller.getViewport().setViewPosition(new Point(0, 0)));
    }

    private void addStdChoicesCombo() {
        List<AttributeSet> sets = AttributeSet.get();
        sets.add(0, new AttributeSet(I18n.Text("Select a Pre-Defined Attribute Set"), null));
        JComboBox<AttributeSet> combo = new JComboBox<>(sets.toArray(new AttributeSet[0]));
        combo.setSelectedItem(sets.get(0));
        combo.addItemListener((evt) -> {
            if (evt.getStateChange() == ItemEvent.SELECTED) {
                AttributeSet choice = (AttributeSet) evt.getItem();
                Map<String, AttributeDef> attributes = choice.getAttributes();
                if (attributes != null) {
                    reset(attributes);
                    combo.setSelectedItem(sets.get(0));
                }
            }
        });
        add(combo);
    }

    public void reset(Map<String, AttributeDef> attributes) {
        mListPanel.reset(attributes);
        mListPanel.getAdjustCallback().run();
        revalidate();
        EventQueue.invokeLater(() -> mScroller.getViewport().setViewPosition(new Point(0, 0)));
    }

    private void importData() {
        Path path = StdFileDialog.showOpenDialog(this, I18n.Text("Import Attribute Settings…"),
                FileType.ATTRIBUTE_SETTINGS.getFilter());
        if (path != null) {
            try {
                AttributeSet set = new AttributeSet(path);
                reset(set.getAttributes());
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
                File         file = transaction.getTransactionFile(path.toFile());
                AttributeSet set  = new AttributeSet(PathUtils.getLeafName(path, false), mListPanel.getAttributes());
                try (JsonWriter w = new JsonWriter(new BufferedWriter(new FileWriter(file, StandardCharsets.UTF_8)), "\t")) {
                    set.toJSON(w);
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
