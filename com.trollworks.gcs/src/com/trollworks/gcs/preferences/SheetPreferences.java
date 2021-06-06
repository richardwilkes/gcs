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

package com.trollworks.gcs.preferences;

import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;

import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import javax.swing.JCheckBox;
import javax.swing.JTextField;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;

/** The sheet preferences panel. */
public class SheetPreferences extends PreferencePanel implements DocumentListener, ItemListener {
    private JTextField mPlayerName;
    private JTextField mTechLevel;
    private JTextField mInitialPoints;
    private JCheckBox  mAutoFillProfile;

    /**
     * Creates a new SheetPreferences.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public SheetPreferences(PreferencesWindow owner) {
        super(I18n.text("Sheet"), owner);
        setLayout(new PrecisionLayout().setColumns(2));
        Preferences prefs = Preferences.getInstance();

        mPlayerName = addTextField(I18n.text("Player"), prefs.getDefaultPlayerName(), I18n.text("The player name to use when a new character sheet is created"));
        mTechLevel = addTextField(I18n.text("Tech Level"), prefs.getDefaultTechLevel(), I18n.text("""
                <html><body>
                TL0: Stone Age (Prehistory and later)<br>
                TL1: Bronze Age (3500 B.C.+)<br>
                TL2: Iron Age (1200 B.C.+)<br>
                TL3: Medieval (600 A.D.+)<br>
                TL4: Age of Sail (1450+)<br>
                TL5: Industrial Revolution (1730+)<br>
                TL6: Mechanized Age (1880+)<br>
                TL7: Nuclear Age (1940+)<br>
                TL8: Digital Age (1980+)<br>
                TL9: Microtech Age (2025+?)<br>
                TL10: Robotic Age (2070+?)<br>
                TL11: Age of Exotic Matter<br>
                TL12: Anything Goes
                </body></html>"""));
        mInitialPoints = addTextField(I18n.text("Initial Points"), Integer.toString(prefs.getInitialPoints()), I18n.text("The initial number of character points to start with"));
        mAutoFillProfile = addCheckBox(I18n.text("Automatically fill in new character identity and description information with randomized choices"), prefs.autoFillProfile());
    }

    private JTextField addTextField(String title, String value, String tooltip) {
        JTextField field = new JTextField(value);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.getDocument().addDocumentListener(this);
        add(new LinkedLabel(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment());
        return field;
    }

    private JCheckBox addCheckBox(String title, boolean checked) {
        JCheckBox checkbox = new JCheckBox(title, checked);
        checkbox.setOpaque(false);
        checkbox.addItemListener(this);
        add(checkbox, new PrecisionLayoutData().setHorizontalSpan(2));
        return checkbox;
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        Preferences prefs    = Preferences.getInstance();
        Document    document = event.getDocument();
        if (mPlayerName.getDocument() == document) {
            prefs.setDefaultPlayerName(mPlayerName.getText());
        } else if (mTechLevel.getDocument() == document) {
            prefs.setDefaultTechLevel(mTechLevel.getText());
        } else if (mInitialPoints.getDocument() == document) {
            prefs.setInitialPoints(Numbers.extractInteger(mInitialPoints.getText(), 0, true));
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

    @Override
    public void itemStateChanged(ItemEvent event) {
        Preferences prefs  = Preferences.getInstance();
        Object      source = event.getSource();
        if (source == mAutoFillProfile) {
            prefs.setAutoFillProfile(mAutoFillProfile.isSelected());
        }
        adjustResetButton();
    }

    @Override
    public void reset() {
        mPlayerName.setText(Preferences.DEFAULT_DEFAULT_PLAYER_NAME);
        mTechLevel.setText(Preferences.DEFAULT_DEFAULT_TECH_LEVEL);
        mInitialPoints.setText(Integer.toString(Preferences.DEFAULT_INITIAL_POINTS));
        mAutoFillProfile.setSelected(Preferences.DEFAULT_AUTO_FILL_PROFILE);
    }

    @Override
    public boolean isSetToDefaults() {
        Preferences prefs     = Preferences.getInstance();
        boolean     atDefault = prefs.getDefaultPlayerName().equals(Preferences.DEFAULT_DEFAULT_PLAYER_NAME);
        atDefault = atDefault && prefs.getDefaultTechLevel().equals(Preferences.DEFAULT_DEFAULT_TECH_LEVEL);
        atDefault = atDefault && prefs.getInitialPoints() == Preferences.DEFAULT_INITIAL_POINTS;
        atDefault = atDefault && prefs.autoFillProfile() == Preferences.DEFAULT_AUTO_FILL_PROFILE;
        return atDefault;
    }
}
