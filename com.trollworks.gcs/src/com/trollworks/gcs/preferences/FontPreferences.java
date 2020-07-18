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

import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.FontPanel;
import com.trollworks.gcs.utility.I18n;

import java.awt.Font;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import javax.swing.JLabel;
import javax.swing.SwingConstants;
import javax.swing.UIManager;

/** The font preferences panel. */
public class FontPreferences extends PreferencePanel implements ActionListener {
    private FontPanel[] mFontPanels;
    private boolean     mIgnore;

    /**
     * Creates a new {@link FontPreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public FontPreferences(PreferencesWindow owner) {
        super(I18n.Text("Fonts"), owner);
        setLayout(new PrecisionLayout().setColumns(2));
        String[] keys = Fonts.getKeys();
        mFontPanels = new FontPanel[keys.length];
        int i = 0;
        for (String key : keys) {
            JLabel label = new JLabel(Fonts.getDescription(key), SwingConstants.RIGHT);
            label.setOpaque(false);
            add(label, new PrecisionLayoutData().setFillHorizontalAlignment());
            mFontPanels[i] = new FontPanel(UIManager.getFont(key));
            mFontPanels[i].setActionCommand(key);
            mFontPanels[i].addActionListener(this);
            add(mFontPanels[i]);
            i++;
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        if (!mIgnore) {
            Object source = event.getSource();
            if (source instanceof FontPanel) {
                boolean adjusted = false;
                for (FontPanel panel : mFontPanels) {
                    if (panel == source) {
                        Font   font = panel.getCurrentFont();
                        String key  = panel.getActionCommand();
                        if (!font.equals(UIManager.getFont(key))) {
                            UIManager.put(key, font);
                            Preferences.getInstance().setFontInfo(key, new Fonts.Info(font));
                            adjusted = true;
                        }
                        break;
                    }
                }
                if (adjusted) {
                    BaseWindow.forceRepaintAndInvalidate();
                    Fonts.notifyOfFontChanges();
                }
            }
            adjustResetButton();
        }
    }

    @Override
    public void reset() {
        Fonts.restoreDefaults();
        mIgnore = true;
        for (FontPanel panel : mFontPanels) {
            panel.setCurrentFont(UIManager.getFont(panel.getActionCommand()));
        }
        mIgnore = false;
        BaseWindow.forceRepaintAndInvalidate();
        Fonts.notifyOfFontChanges();
    }

    @Override
    public boolean isSetToDefaults() {
        return Fonts.isSetToDefaults();
    }
}
