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

import com.trollworks.gcs.character.DisplayOption;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.layout.FlexColumn;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.scale.Scales;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.WeightUnits;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.util.List;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JTextArea;
import javax.swing.JTextField;
import javax.swing.SwingConstants;
import javax.swing.border.CompoundBorder;
import javax.swing.border.EmptyBorder;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;

/** The display preferences panel. */
public class DisplayPreferences extends PreferencePanel implements ActionListener, DocumentListener, ItemListener {
    private JCheckBox                mIncludeUnspentPointsInTotal;
    private JCheckBox                mShowCollegeInSheetSpells;
    private JCheckBox                mShowTitleInsteadOfNameInPageFooter;
    private JComboBox<Scales>        mUIScaleCombo;
    private JComboBox<LengthUnits>   mLengthUnitsCombo;
    private JComboBox<WeightUnits>   mWeightUnitsCombo;
    private JComboBox<DisplayOption> mUserDescriptionDisplayCombo;
    private JComboBox<DisplayOption> mModifiersDisplayCombo;
    private JComboBox<DisplayOption> mNotesDisplayCombo;
    private JTextField               mToolTipTimeout;
    private JTextArea                mBlockLayoutField;

    /**
     * Creates a new {@link DisplayPreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public DisplayPreferences(PreferencesWindow owner) {
        super(I18n.Text("Display"), owner);
        Preferences prefs  = Preferences.getInstance();
        FlexColumn  column = new FlexColumn();

        mIncludeUnspentPointsInTotal = createCheckBox(I18n.Text("Character point total display includes unspent points"), null, prefs.includeUnspentPointsInTotal());
        column.add(mIncludeUnspentPointsInTotal);

        mShowCollegeInSheetSpells = createCheckBox(I18n.Text("Show the College column on character sheets"), null, prefs.showCollegeInSheetSpells());
        column.add(mShowCollegeInSheetSpells);

        mShowTitleInsteadOfNameInPageFooter = createCheckBox(I18n.Text("Show the title rather than the name in the page footer on character sheets"), null, prefs.useTitleInFooter());
        column.add(mShowTitleInsteadOfNameInPageFooter);

        FlexRow row = new FlexRow();
        row.add(createLabel(I18n.Text("Use"), null));
        mUIScaleCombo = createCombo(Scales.values(), prefs.getInitialUIScale(), null);
        row.add(mUIScaleCombo);
        JLabel label = createLabel(I18n.Text("for the initial scale when opening character sheets, templates and lists"), null, SwingConstants.LEFT);
        UIUtilities.setOnlySize(label, null);
        row.add(label);
        column.add(row);

        row = new FlexRow();
        row.add(createLabel(I18n.Text("Use"), null));
        mLengthUnitsCombo = createCombo(LengthUnits.values(), prefs.getDefaultLengthUnits(), I18n.Text("The units to use for display of generated lengths"));
        row.add(mLengthUnitsCombo);
        row.add(createLabel(I18n.Text("and"), null));
        mWeightUnitsCombo = createCombo(WeightUnits.values(), prefs.getDefaultWeightUnits(), I18n.Text("The units to use for display of generated weights"));
        row.add(mWeightUnitsCombo);
        row.add(createLabel(I18n.Text("for display of generated units"), null));
        column.add(row);

        row = new FlexRow();
        String seconds = I18n.Text("Seconds");
        row.add(createLabel(I18n.Text("Tooltip Timeout"), seconds));
        mToolTipTimeout = createTextField(seconds, Integer.valueOf(prefs.getToolTipTimeout()).toString());
        Dimension size = mToolTipTimeout.getPreferredSize();
        size.width = TextDrawing.getWidth(mToolTipTimeout.getFont(), Integer.valueOf(Preferences.MAXIMUM_TOOLTIP_TIMEOUT).toString()) * 3 / 2;
        UIUtilities.setOnlySize(mToolTipTimeout, size);
        row.add(mToolTipTimeout);
        row.add(createLabel(I18n.Text("seconds"), null));
        column.add(row);

        row = new FlexRow();
        row.add(createLabel(I18n.Text("Show User Description"), null));
        mUserDescriptionDisplayCombo = createCombo(DisplayOption.values(), prefs.getUserDescriptionDisplay(), I18n.Text("Where to display this information"));
        row.add(mUserDescriptionDisplayCombo);
        column.add(row);

        row = new FlexRow();
        row.add(createLabel(I18n.Text("Show Modifiers"), null));
        mModifiersDisplayCombo = createCombo(DisplayOption.values(), prefs.getModifiersDisplay(), I18n.Text("Where to display this information"));
        row.add(mModifiersDisplayCombo);
        column.add(row);

        row = new FlexRow();
        row.add(createLabel(I18n.Text("Show Notes"), null));
        mNotesDisplayCombo = createCombo(DisplayOption.values(), prefs.getNotesDisplay(), I18n.Text("Where to display this information"));
        row.add(mNotesDisplayCombo);
        column.add(row);

        row = new FlexRow();
        String blockLayoutTooltip = I18n.Text("Specifies the layout of the various blocks of data on the character sheet");
        row.add(createLabel(I18n.Text("Block Layout"), blockLayoutTooltip));
        column.add(row);
        row = new FlexRow();
        mBlockLayoutField = createTextArea(blockLayoutTooltip, Preferences.linesToString(prefs.getBlockLayout()));
        row.add(mBlockLayoutField);
        row.setFillVertical(true);
        column.add(row);

        column.apply(this);
    }

    private <E> JComboBox<E> createCombo(E[] values, E choice, String tooltip) {
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

    private JTextArea createTextArea(String tooltip, String value) {
        JTextArea field = new JTextArea(value);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.getDocument().addDocumentListener(this);
        field.setBorder(new CompoundBorder(new LineBorder(), new EmptyBorder(0, 4, 0, 4)));
        add(field);
        return field;
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Preferences prefs  = Preferences.getInstance();
        Object      source = event.getSource();
        if (source == mUIScaleCombo) {
            prefs.setInitialUIScale((Scales) mUIScaleCombo.getSelectedItem());
        } else if (source == mLengthUnitsCombo) {
            prefs.setDefaultLengthUnits((LengthUnits) mLengthUnitsCombo.getSelectedItem());
        } else if (source == mWeightUnitsCombo) {
            prefs.setDefaultWeightUnits((WeightUnits) mWeightUnitsCombo.getSelectedItem());
        } else if (source == mUserDescriptionDisplayCombo) {
            prefs.setUserDescriptionDisplay((DisplayOption) mUserDescriptionDisplayCombo.getSelectedItem());
        } else if (source == mModifiersDisplayCombo) {
            prefs.setModifiersDisplay((DisplayOption) mModifiersDisplayCombo.getSelectedItem());
        } else if (source == mNotesDisplayCombo) {
            prefs.setNotesDisplay((DisplayOption) mNotesDisplayCombo.getSelectedItem());
        }
        adjustResetButton();
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        Preferences prefs    = Preferences.getInstance();
        Document    document = event.getDocument();
        if (mBlockLayoutField.getDocument() == document) {
            prefs.setBlockLayout(List.of(mBlockLayoutField.getText().split("\n")));
        } else if (mToolTipTimeout.getDocument() == document) {
            prefs.setToolTipTimeout(Numbers.extractInteger(mToolTipTimeout.getText(), Preferences.DEFAULT_TOOLTIP_TIMEOUT, Preferences.MINIMUM_TOOLTIP_TIMEOUT, Preferences.MAXIMUM_TOOLTIP_TIMEOUT, true));
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
        Object source = event.getSource();
        if (source == mIncludeUnspentPointsInTotal) {
            Preferences.getInstance().setIncludeUnspentPointsInTotal(mIncludeUnspentPointsInTotal.isSelected());
        } else if (source == mShowCollegeInSheetSpells) {
            Preferences.getInstance().setShowCollegeInSheetSpells(mShowCollegeInSheetSpells.isSelected());
        } else if (source == mShowTitleInsteadOfNameInPageFooter) {
            Preferences.getInstance().setUseTitleInFooter(mShowTitleInsteadOfNameInPageFooter.isSelected());
        }
        adjustResetButton();
    }

    @Override
    public void reset() {
        mIncludeUnspentPointsInTotal.setSelected(Preferences.DEFAULT_INCLUDE_UNSPENT_POINTS_IN_TOTAL);
        mShowCollegeInSheetSpells.setSelected(Preferences.DEFAULT_SHOW_COLLEGE_IN_SHEET_SPELLS);
        mShowTitleInsteadOfNameInPageFooter.setSelected(Preferences.DEFAULT_USE_TITLE_IN_FOOTER);
        mUIScaleCombo.setSelectedItem(Preferences.DEFAULT_INITIAL_UI_SCALE);
        mLengthUnitsCombo.setSelectedItem(Preferences.DEFAULT_DEFAULT_LENGTH_UNITS);
        mWeightUnitsCombo.setSelectedItem(Preferences.DEFAULT_DEFAULT_WEIGHT_UNITS);
        mToolTipTimeout.setText(Integer.toString(Preferences.DEFAULT_TOOLTIP_TIMEOUT));
        mBlockLayoutField.setText(Preferences.linesToString(Preferences.DEFAULT_BLOCK_LAYOUT));
        mUserDescriptionDisplayCombo.setSelectedItem(Preferences.DEFAULT_USER_DESCRIPTION_DISPLAY);
        mModifiersDisplayCombo.setSelectedItem(Preferences.DEFAULT_MODIFIERS_DISPLAY);
        mNotesDisplayCombo.setSelectedItem(Preferences.DEFAULT_NOTES_DISPLAY);
    }

    @Override
    public boolean isSetToDefaults() {
        Preferences prefs     = Preferences.getInstance();
        boolean     atDefault = prefs.includeUnspentPointsInTotal() == Preferences.DEFAULT_INCLUDE_UNSPENT_POINTS_IN_TOTAL;
        atDefault = atDefault && prefs.showCollegeInSheetSpells() == Preferences.DEFAULT_SHOW_COLLEGE_IN_SHEET_SPELLS;
        atDefault = atDefault && prefs.getInitialUIScale() == Preferences.DEFAULT_INITIAL_UI_SCALE;
        atDefault = atDefault && prefs.getDefaultLengthUnits() == Preferences.DEFAULT_DEFAULT_LENGTH_UNITS;
        atDefault = atDefault && prefs.getDefaultWeightUnits() == Preferences.DEFAULT_DEFAULT_WEIGHT_UNITS;
        atDefault = atDefault && prefs.getToolTipTimeout() == Preferences.DEFAULT_TOOLTIP_TIMEOUT;
        atDefault = atDefault && prefs.getBlockLayout().equals(Preferences.DEFAULT_BLOCK_LAYOUT);
        atDefault = atDefault && prefs.getUserDescriptionDisplay() == Preferences.DEFAULT_USER_DESCRIPTION_DISPLAY;
        atDefault = atDefault && prefs.getModifiersDisplay() == Preferences.DEFAULT_MODIFIERS_DISPLAY;
        atDefault = atDefault && prefs.getNotesDisplay() == Preferences.DEFAULT_NOTES_DISPLAY;
        return atDefault;
    }
}
