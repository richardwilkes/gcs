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

import com.trollworks.gcs.character.DisplayOption;
import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.scale.Scales;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.WeightUnits;

import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.beans.PropertyChangeListener;
import java.util.List;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JFormattedTextField;
import javax.swing.JLabel;
import javax.swing.JScrollPane;
import javax.swing.JTextArea;
import javax.swing.ScrollPaneConstants;
import javax.swing.SwingConstants;
import javax.swing.border.EmptyBorder;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;

/** The display preferences panel. */
public class DisplayPreferences extends PreferencePanel implements ActionListener, DocumentListener, ItemListener {
    private JCheckBox                mIncludeUnspentPointsInTotal;
    private JCheckBox                mShowCollegeInSheetSpells;
    private JCheckBox                mShowDifficulty;
    private JCheckBox                mShowAdvantageModifierAdj;
    private JCheckBox                mShowEquipmentModifierAdj;
    private JCheckBox                mShowSpellAdj;
    private JCheckBox                mShowTitleInsteadOfNameInPageFooter;
    private JComboBox<Scales>        mUIScaleCombo;
    private JComboBox<LengthUnits>   mLengthUnitsCombo;
    private JComboBox<WeightUnits>   mWeightUnitsCombo;
    private JComboBox<DisplayOption> mUserDescriptionDisplayCombo;
    private JComboBox<DisplayOption> mModifiersDisplayCombo;
    private JComboBox<DisplayOption> mNotesDisplayCombo;
    private EditorField              mToolTipTimeout;
    private JTextArea                mBlockLayoutField;

    /**
     * Creates a new {@link DisplayPreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public DisplayPreferences(PreferencesWindow owner) {
        super(I18n.Text("Display"), owner);
        setLayout(new PrecisionLayout().setColumns(2));
        Preferences prefs = Preferences.getInstance();

        mIncludeUnspentPointsInTotal = addCheckBox(I18n.Text("Character point total display includes unspent points"), prefs.includeUnspentPointsInTotal());
        mShowCollegeInSheetSpells = addCheckBox(I18n.Text("Show the College column *"), prefs.showCollegeInSheetSpells());
        mShowDifficulty = addCheckBox(I18n.Text("Show the Difficulty column *"), prefs.showDifficulty());
        mShowAdvantageModifierAdj = addCheckBox(I18n.Text("Show advantage modifier cost adjustments *"), prefs.showAdvantageModifierAdj());
        mShowEquipmentModifierAdj = addCheckBox(I18n.Text("Show equipment modifier cost & weight adjustments *"), prefs.showEquipmentModifierAdj());
        mShowSpellAdj = addCheckBox(I18n.Text("Show spell ritual, cost & time adjustments *"), prefs.showSpellAdj());
        mShowTitleInsteadOfNameInPageFooter = addCheckBox(I18n.Text("Show the title instead of the name in the footer *"), prefs.useTitleInFooter());

        addLabel(I18n.Text("Initial Scale"));
        mUIScaleCombo = addCombo(Scales.values(), prefs.getInitialUIScale(), null);

        addLabel(I18n.Text("Length Units *"));
        mLengthUnitsCombo = addCombo(LengthUnits.values(), prefs.getDefaultLengthUnits(), I18n.Text("The units to use for display of generated lengths"));

        addLabel(I18n.Text("Weight Units *"));
        mWeightUnitsCombo = addCombo(WeightUnits.values(), prefs.getDefaultWeightUnits(), I18n.Text("The units to use for display of generated weights"));

        addLabel(I18n.Text("Tooltip Timeout (seconds)"));
        mToolTipTimeout = addField(Integer.valueOf(prefs.getToolTipTimeout()), Integer.valueOf(Preferences.MAXIMUM_TOOLTIP_TIMEOUT), FieldFactory.POSINT6, (evt) -> {
            int value = ((Integer) evt.getNewValue()).intValue();
            int revised = Math.min(Math.max(value, Preferences.MINIMUM_TOOLTIP_TIMEOUT), Preferences.MAXIMUM_TOOLTIP_TIMEOUT);
            Preferences.getInstance().setToolTipTimeout(revised);
            adjustResetButton();
            if (value != revised) {
                mToolTipTimeout.setValue(Integer.valueOf(revised));
            }
        });

        addLabel(I18n.Text("Show User Description *"));
        mUserDescriptionDisplayCombo = addCombo(DisplayOption.values(), prefs.getUserDescriptionDisplay(), I18n.Text("Where to display this information"));

        addLabel(I18n.Text("Show Modifiers *"));
        mModifiersDisplayCombo = addCombo(DisplayOption.values(), prefs.getModifiersDisplay(), I18n.Text("Where to display this information"));

        addLabel(I18n.Text("Show Notes *"));
        mNotesDisplayCombo = addCombo(DisplayOption.values(), prefs.getNotesDisplay(), I18n.Text("Where to display this information"));

        String tooltip = I18n.Text("Specifies the layout of the various blocks of data on the character sheet");
        JLabel label   = new JLabel(I18n.Text("Block Layout *"));
        label.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        label.setOpaque(false);
        add(label, new PrecisionLayoutData().setHorizontalSpan(2));
        mBlockLayoutField = addTextArea(Preferences.linesToString(prefs.getBlockLayout()), tooltip);

        label = new JLabel(I18n.Text("* To change the setting on existing sheets, use the per-sheet settings available from the toolbar"));
        label.setOpaque(false);
        add(label, new PrecisionLayoutData().setHorizontalSpan(2).setHorizontalAlignment(PrecisionLayoutAlignment.MIDDLE));
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

    private JTextArea addTextArea(String value, String tooltip) {
        JTextArea field = new JTextArea(value);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setBorder(new EmptyBorder(0, 4, 0, 4));
        field.getDocument().addDocumentListener(this);
        JScrollPane scroller = new JScrollPane(field, ScrollPaneConstants.VERTICAL_SCROLLBAR_AS_NEEDED, ScrollPaneConstants.HORIZONTAL_SCROLLBAR_AS_NEEDED);
        add(scroller, new PrecisionLayoutData().setHorizontalSpan(2).setHeightHint(scroller.getPreferredSize().height).setFillAlignment().setGrabSpace(true));
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
        } else if (source == mShowDifficulty) {
            Preferences.getInstance().setShowDifficulty(mShowDifficulty.isSelected());
        } else if (source == mShowAdvantageModifierAdj) {
            Preferences.getInstance().setShowAdvantageModifierAdj(mShowAdvantageModifierAdj.isSelected());
        } else if (source == mShowEquipmentModifierAdj) {
            Preferences.getInstance().setShowEquipmentModifierAdj(mShowEquipmentModifierAdj.isSelected());
        } else if (source == mShowSpellAdj) {
            Preferences.getInstance().setShowSpellAdj(mShowSpellAdj.isSelected());
        } else if (source == mShowTitleInsteadOfNameInPageFooter) {
            Preferences.getInstance().setUseTitleInFooter(mShowTitleInsteadOfNameInPageFooter.isSelected());
        }
        adjustResetButton();
    }

    @Override
    public void reset() {
        mIncludeUnspentPointsInTotal.setSelected(Preferences.DEFAULT_INCLUDE_UNSPENT_POINTS_IN_TOTAL);
        mShowCollegeInSheetSpells.setSelected(Preferences.DEFAULT_SHOW_COLLEGE_IN_SHEET_SPELLS);
        mShowDifficulty.setSelected(Preferences.DEFAULT_SHOW_DIFFICULTY);
        mShowAdvantageModifierAdj.setSelected(Preferences.DEFAULT_SHOW_ADVANTAGE_MODIFIER_ADJ);
        mShowEquipmentModifierAdj.setSelected(Preferences.DEFAULT_SHOW_EQUIPMENT_MODIFIER_ADJ);
        mShowSpellAdj.setSelected(Preferences.DEFAULT_SHOW_SPELL_ADJ);
        mShowTitleInsteadOfNameInPageFooter.setSelected(Preferences.DEFAULT_USE_TITLE_IN_FOOTER);
        mUIScaleCombo.setSelectedItem(Preferences.DEFAULT_INITIAL_UI_SCALE);
        mLengthUnitsCombo.setSelectedItem(Preferences.DEFAULT_DEFAULT_LENGTH_UNITS);
        mWeightUnitsCombo.setSelectedItem(Preferences.DEFAULT_DEFAULT_WEIGHT_UNITS);
        mToolTipTimeout.setValue(Integer.valueOf(Preferences.DEFAULT_TOOLTIP_TIMEOUT));
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
        atDefault = atDefault && prefs.showDifficulty() == Preferences.DEFAULT_SHOW_DIFFICULTY;
        atDefault = atDefault && prefs.showAdvantageModifierAdj() == Preferences.DEFAULT_SHOW_ADVANTAGE_MODIFIER_ADJ;
        atDefault = atDefault && prefs.showEquipmentModifierAdj() == Preferences.DEFAULT_SHOW_EQUIPMENT_MODIFIER_ADJ;
        atDefault = atDefault && prefs.showSpellAdj() == Preferences.DEFAULT_SHOW_SPELL_ADJ;
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
