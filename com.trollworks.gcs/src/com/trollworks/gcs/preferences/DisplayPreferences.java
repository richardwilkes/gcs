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

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.scale.Scales;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.beans.PropertyChangeListener;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JFormattedTextField;
import javax.swing.JLabel;
import javax.swing.SwingConstants;

/** The display preferences panel. */
public class DisplayPreferences extends PreferencePanel implements ActionListener, ItemListener {
    private JCheckBox         mIncludeUnspentPointsInTotal;
    private JComboBox<Scales> mUIScaleCombo;
    private EditorField       mToolTipTimeout;

    /**
     * Creates a new DisplayPreferences.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public DisplayPreferences(PreferencesWindow owner) {
        super(I18n.text("Display"), owner);
        setLayout(new PrecisionLayout().setColumns(2));
        Preferences prefs = Preferences.getInstance();

        mIncludeUnspentPointsInTotal = addCheckBox(I18n.text("Character point total display includes unspent points"), prefs.includeUnspentPointsInTotal());

        addLabel(I18n.text("Initial Scale"));
        mUIScaleCombo = addCombo(Scales.values(), prefs.getInitialUIScale(), null);

        addLabel(I18n.text("Tooltip Timeout (seconds)"));
        mToolTipTimeout = addField(Integer.valueOf(prefs.getToolTipTimeout()), Integer.valueOf(Preferences.MAXIMUM_TOOLTIP_TIMEOUT), FieldFactory.POSINT6, (evt) -> {
            int value   = ((Integer) evt.getNewValue()).intValue();
            int revised = Math.min(Math.max(value, Preferences.MINIMUM_TOOLTIP_TIMEOUT), Preferences.MAXIMUM_TOOLTIP_TIMEOUT);
            Preferences.getInstance().setToolTipTimeout(revised);
            adjustResetButton();
            if (value != revised) {
                mToolTipTimeout.setValue(Integer.valueOf(revised));
            }
        });
    }

    private void addLabel(String title) {
        JLabel label = new JLabel(title, SwingConstants.RIGHT);
        label.setOpaque(false);
        add(label, new PrecisionLayoutData().setFillHorizontalAlignment());
    }

    private JCheckBox addCheckBox(String title, boolean checked) {
        JCheckBox checkbox = new JCheckBox(title, checked);
        checkbox.setOpaque(false);
        checkbox.addItemListener(this);
        add(checkbox, new PrecisionLayoutData().setHorizontalSpan(2));
        return checkbox;
    }

    private EditorField addField(Object value, Object protoValue, JFormattedTextField.AbstractFormatterFactory formatter, PropertyChangeListener listener) {
        EditorField field = new EditorField(formatter, listener, SwingConstants.LEFT, value, protoValue, null);
        add(field);
        return field;
    }

    private <E> JComboBox<E> addCombo(E[] values, E choice, String tooltip) {
        JComboBox<E> combo = new JComboBox<>(values);
        combo.setOpaque(false);
        combo.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        combo.setSelectedItem(choice);
        combo.addActionListener(this);
        combo.setMaximumRowCount(combo.getItemCount());
        UIUtilities.setToPreferredSizeOnly(combo);
        add(combo);
        return combo;
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Preferences prefs  = Preferences.getInstance();
        Object      source = event.getSource();
        if (source == mUIScaleCombo) {
            prefs.setInitialUIScale((Scales) mUIScaleCombo.getSelectedItem());
        }
        adjustResetButton();
    }

    @Override
    public void itemStateChanged(ItemEvent event) {
        Object source = event.getSource();
        if (source == mIncludeUnspentPointsInTotal) {
            Preferences.getInstance().setIncludeUnspentPointsInTotal(mIncludeUnspentPointsInTotal.isSelected());
        }
        adjustResetButton();
    }

    @Override
    public void reset() {
        mIncludeUnspentPointsInTotal.setSelected(Preferences.DEFAULT_INCLUDE_UNSPENT_POINTS_IN_TOTAL);
        mUIScaleCombo.setSelectedItem(Preferences.DEFAULT_INITIAL_UI_SCALE);
        mToolTipTimeout.setValue(Integer.valueOf(Preferences.DEFAULT_TOOLTIP_TIMEOUT));
    }

    @Override
    public boolean isSetToDefaults() {
        Preferences prefs     = Preferences.getInstance();
        boolean     atDefault = prefs.includeUnspentPointsInTotal() == Preferences.DEFAULT_INCLUDE_UNSPENT_POINTS_IN_TOTAL;
        atDefault = atDefault && prefs.getInitialUIScale() == Preferences.DEFAULT_INITIAL_UI_SCALE;
        atDefault = atDefault && prefs.getToolTipTimeout() == Preferences.DEFAULT_TOOLTIP_TIMEOUT;
        return atDefault;
    }
}
