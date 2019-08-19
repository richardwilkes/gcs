/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.preferences;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.layout.FlexColumn;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.preferences.PreferencePanel;
import com.trollworks.toolkit.ui.preferences.PreferencesWindow;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.text.Text;

import java.awt.Dimension;

import javax.swing.JTextField;
import javax.swing.ToolTipManager;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;

public class SystemPreferences extends PreferencePanel implements DocumentListener {
    @Localize("System")
    @Localize(locale = "de", value = "")
    @Localize(locale = "ru", value = "")
    @Localize(locale = "es", value = "")
    private static String SYSTEM;

    @Localize("Tooltip Timeout (in milliseconds)")
    private static String TOOLTIP_TIMEOUT;

    static {
        Localization.initialize();
    }

    static final String      MODULE                  = "System"; //$NON-NLS-1$
    private static final int DEFAULT_TOOLTIP_TIMEOUT = 60000;

    private JTextField       mToolTipTimeout;

    public static int getToolTipTimeout() {
        return Preferences.getInstance().getIntValue(MODULE, TOOLTIP_TIMEOUT, DEFAULT_TOOLTIP_TIMEOUT);
    }

    public SystemPreferences(PreferencesWindow owner) {
        super(SYSTEM, owner);

        FlexColumn column = new FlexColumn();

        FlexGrid   grid   = new FlexGrid();
        column.add(grid);
        grid.setFillHorizontal(true);

        FlexRow row = new FlexRow();
        row.add(createLabel(TOOLTIP_TIMEOUT, TOOLTIP_TIMEOUT));
        mToolTipTimeout = createTextField(TOOLTIP_TIMEOUT, Integer.toString(getToolTipTimeout()));
        row.add(mToolTipTimeout);
        column.add(row);

        column.apply(this);
    }

    private JTextField createTextField(String tooltip, String value) {
        JTextField field = new JTextField(value);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.getDocument().addDocumentListener(this);
        Dimension size    = field.getPreferredSize();
        Dimension maxSize = field.getMaximumSize();
        maxSize.height = size.height;
        field.setMaximumSize(maxSize);
        add(field);
        return field;
    }

    @Override
    public void reset() {
        mToolTipTimeout.setText(Integer.toString(DEFAULT_TOOLTIP_TIMEOUT));
    }

    @Override
    public boolean isSetToDefaults() {
        return getToolTipTimeout() == DEFAULT_TOOLTIP_TIMEOUT;
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        Document document = event.getDocument();
        if (mToolTipTimeout.getDocument() == document) {
            Preferences.getInstance().setValue(MODULE, TOOLTIP_TIMEOUT, Numbers.extractInteger(mToolTipTimeout.getText(), 0, true));
            ToolTipManager.sharedInstance().setDismissDelay(getToolTipTimeout());
        }
        adjustResetButton();
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        changedUpdate(event);
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        changedUpdate(event);
    }

}
