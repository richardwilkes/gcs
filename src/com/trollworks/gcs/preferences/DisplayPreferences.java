/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.preferences;

import com.trollworks.toolkit.ui.TextDrawing;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.border.LineBorder;
import com.trollworks.toolkit.ui.layout.FlexColumn;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.preferences.PreferencePanel;
import com.trollworks.toolkit.ui.preferences.PreferencesWindow;
import com.trollworks.toolkit.ui.scale.Scales;
import com.trollworks.toolkit.utility.I18n;
import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.toolkit.utility.text.Enums;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.text.Text;
import com.trollworks.toolkit.utility.units.LengthUnits;
import com.trollworks.toolkit.utility.units.WeightUnits;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;

import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JTextArea;
import javax.swing.JTextField;
import javax.swing.SwingConstants;
import javax.swing.ToolTipManager;
import javax.swing.border.CompoundBorder;
import javax.swing.border.EmptyBorder;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;

/** The display preferences panel. */
public class DisplayPreferences extends PreferencePanel implements ActionListener, DocumentListener, ItemListener {
    private static final String      MODULE                             = "Display";
    private static final String      TOTAL_POINTS_DISPLAY_KEY           = "TotalPointsIncludesUnspentPoints";
    /** The total points includes unspent points preference key. */
    public static final String       TOTAL_POINTS_DISPLAY_PREF_KEY      = Preferences.getModuleKey(MODULE, TOTAL_POINTS_DISPLAY_KEY);
    private static final boolean     DEFAULT_TOTAL_POINTS_DISPLAY       = true;
    private static final String      BLOCK_LAYOUT_KEY                   = "BlockLayout";
    /** The block layout preference key. */
    public static final String       BLOCK_LAYOUT_PREF_KEY              = Preferences.getModuleKey(MODULE, BLOCK_LAYOUT_KEY);
    private static final String      DEFAULT_BLOCK_LAYOUT               = "melee\nranged\nadvantages skills\nspells\nequipment\nother_equipment\nnotes";
    private static final Scales      DEFAULT_SCALE                      = Scales.ACTUAL_SIZE;
    private static final String      SCALE_KEY                          = "UIScale";
    private static final String      LENGTH_UNITS_KEY                   = "LengthUnits";
    /** The default length units preference key. */
    public static final String       LENGTH_UNITS_PREF_KEY              = Preferences.getModuleKey(MODULE, LENGTH_UNITS_KEY);
    private static final LengthUnits DEFAULT_LENGTH_UNITS               = LengthUnits.FT_IN;
    private static final String      WEIGHT_UNITS_KEY                   = "WeightUnits";
    /** The default weight units preference key. */
    public static final String       WEIGHT_UNITS_PREF_KEY              = Preferences.getModuleKey(MODULE, WEIGHT_UNITS_KEY);
    private static final WeightUnits DEFAULT_WEIGHT_UNITS               = WeightUnits.LB;
    private static final String      TOOLTIP_TIMEOUT_KEY                = "TooltipTimeout";
    private static final int         MINIMUM_TOOLTIP_TIMEOUT            = 1;
    private static final int         MAXIMUM_TOOLTIP_TIMEOUT            = 9999;
    private static final int         DEFAULT_TOOLTIP_TIMEOUT            = 60;
    public static final String       SHOW_USER_DESC_AS_TOOL_TIP_KEY     = "ShowUserDescAsToolTip";
    private static final boolean     DEFAULT_SHOW_USER_DESC_AS_TOOL_TIP = true;
    private JCheckBox                mIncludeUnspentPointsInTotal;
    private JComboBox<Scales>        mUIScaleCombo;
    private JComboBox<String>        mLengthUnitsCombo;
    private JComboBox<String>        mWeightUnitsCombo;
    private JTextField               mToolTipTimeout;
    private JTextArea                mBlockLayoutField;
    private JCheckBox                mShowUserDescAsToolTips;

    /** Initializes the services controlled by these preferences. */
    public static void initialize() {
        // Converting preferences that used to live in OutputPreferences
        Preferences prefs = Preferences.getInstance();
        for (String key : prefs.getModuleKeys(OutputPreferences.MODULE)) {
            if (BLOCK_LAYOUT_KEY.equals(key)) {
                prefs.setValue(MODULE, BLOCK_LAYOUT_KEY, prefs.getStringValue(OutputPreferences.MODULE, BLOCK_LAYOUT_KEY, DEFAULT_BLOCK_LAYOUT));
                prefs.removePreference(OutputPreferences.MODULE, BLOCK_LAYOUT_KEY);
            }
        }
        // Converting preferences that used to live in SheetPreferences
        for (String key : prefs.getModuleKeys(SheetPreferences.MODULE)) {
            if (TOTAL_POINTS_DISPLAY_KEY.equals(key)) {
                prefs.setValue(MODULE, TOTAL_POINTS_DISPLAY_KEY, prefs.getBooleanValue(SheetPreferences.MODULE, TOTAL_POINTS_DISPLAY_KEY, DEFAULT_TOTAL_POINTS_DISPLAY));
                prefs.removePreference(SheetPreferences.MODULE, TOTAL_POINTS_DISPLAY_KEY);
            } else if (SCALE_KEY.equals(key)) {
                prefs.setValue(MODULE, SCALE_KEY, prefs.getDoubleValue(SheetPreferences.MODULE, SCALE_KEY, DEFAULT_SCALE.getScale().getScale()));
                prefs.removePreference(SheetPreferences.MODULE, SCALE_KEY);
            } else if (LENGTH_UNITS_KEY.equals(key)) {
                prefs.setValue(MODULE, LENGTH_UNITS_KEY, prefs.getStringValue(SheetPreferences.MODULE, LENGTH_UNITS_KEY, Enums.toId(DEFAULT_LENGTH_UNITS)));
                prefs.removePreference(SheetPreferences.MODULE, LENGTH_UNITS_KEY);
            } else if (WEIGHT_UNITS_KEY.equals(key)) {
                prefs.setValue(MODULE, WEIGHT_UNITS_KEY, prefs.getStringValue(SheetPreferences.MODULE, WEIGHT_UNITS_KEY, Enums.toId(DEFAULT_WEIGHT_UNITS)));
                prefs.removePreference(SheetPreferences.MODULE, WEIGHT_UNITS_KEY);
            } else if (SHOW_USER_DESC_AS_TOOL_TIP_KEY.equals(key)) {
                prefs.setValue(MODULE, SHOW_USER_DESC_AS_TOOL_TIP_KEY, prefs.getBooleanValue(SheetPreferences.MODULE, SHOW_USER_DESC_AS_TOOL_TIP_KEY, DEFAULT_SHOW_USER_DESC_AS_TOOL_TIP));
                prefs.removePreference(SheetPreferences.MODULE, SHOW_USER_DESC_AS_TOOL_TIP_KEY);
            }
        }
        // Converting preferences that used to live in SystemPreferences
        String systemModule = "System";
        for (String key : prefs.getModuleKeys(systemModule)) {
            if (TOOLTIP_TIMEOUT_KEY.equals(key)) {
                int timeout = prefs.getIntValue(systemModule, TOOLTIP_TIMEOUT_KEY, DEFAULT_TOOLTIP_TIMEOUT) / 1000;
                if (timeout < MINIMUM_TOOLTIP_TIMEOUT) {
                    timeout = MINIMUM_TOOLTIP_TIMEOUT;
                } else if (timeout > MAXIMUM_TOOLTIP_TIMEOUT) {
                    timeout = MAXIMUM_TOOLTIP_TIMEOUT;
                }
                prefs.setValue(MODULE, TOOLTIP_TIMEOUT_KEY, timeout);
                prefs.removePreference(systemModule, TOOLTIP_TIMEOUT_KEY);
            }
        }
    }

    /** @return The timeout, in seconds, to use for tooltips. */
    public static int getToolTipTimeout() {
        return Preferences.getInstance().getIntValue(MODULE, TOOLTIP_TIMEOUT_KEY, DEFAULT_TOOLTIP_TIMEOUT);
    }

    /** @return Whether to show the user description as a tooltip. */
    public static boolean showUserDescAsTooltip() {
        return Preferences.getInstance().getBooleanValue(MODULE, SHOW_USER_DESC_AS_TOOL_TIP_KEY, DEFAULT_SHOW_USER_DESC_AS_TOOL_TIP);
    }

    /**
     * @return Whether the character's total points are displayed with or without including earned
     *         (but unspent) points.
     */
    public static boolean shouldIncludeUnspentPointsInTotalPointDisplay() {
        return Preferences.getInstance().getBooleanValue(MODULE, TOTAL_POINTS_DISPLAY_KEY, DEFAULT_TOTAL_POINTS_DISPLAY);
    }

    /** @return The {@link Scales} to use when opening a new scalable file. */
    public static Scales getInitialUIScale() {
        double value = Preferences.getInstance().getDoubleValue(MODULE, SCALE_KEY, DEFAULT_SCALE.getScale().getScale());
        for (Scales one : Scales.values()) {
            if (one.getScale().getScale() == value) {
                return one;
            }
        }
        return Scales.ACTUAL_SIZE;
    }

    /** @return The default length units to use. */
    public static LengthUnits getLengthUnits() {
        return Enums.extract(Preferences.getInstance().getStringValue(MODULE, LENGTH_UNITS_KEY), LengthUnits.values(), DEFAULT_LENGTH_UNITS);
    }

    /** @return The default weight units to use. */
    public static WeightUnits getWeightUnits() {
        return Enums.extract(Preferences.getInstance().getStringValue(MODULE, WEIGHT_UNITS_KEY), WeightUnits.values(), DEFAULT_WEIGHT_UNITS);
    }

    /** @return The block layout settings. */
    public static String getBlockLayout() {
        return Preferences.getInstance().getStringValue(MODULE, BLOCK_LAYOUT_KEY, DEFAULT_BLOCK_LAYOUT);
    }

    /**
     * Creates a new {@link DisplayPreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public DisplayPreferences(PreferencesWindow owner) {
        super(I18n.Text("Display"), owner);
        FlexColumn column = new FlexColumn();

        FlexGrid   grid   = new FlexGrid();
        column.add(grid);

        mIncludeUnspentPointsInTotal = createCheckBox(I18n.Text("Character point total display includes unspent points"), null, shouldIncludeUnspentPointsInTotalPointDisplay());
        column.add(mIncludeUnspentPointsInTotal);

        mShowUserDescAsToolTips = createCheckBox(I18n.Text("Show advantage user description as a tooltip"), null, showUserDescAsTooltip());
        column.add(mShowUserDescAsToolTips);

        FlexRow row = new FlexRow();
        row.add(createLabel(I18n.Text("Use"), null));
        mUIScaleCombo = createUIScalePopup();
        row.add(mUIScaleCombo);
        row.add(createLabel(I18n.Text("for the initial scale when opening character sheets, templates and lists"), null, SwingConstants.LEFT));
        column.add(row);

        row = new FlexRow();
        row.add(createLabel(I18n.Text("Use"), null));
        mLengthUnitsCombo = createLengthUnitsPopup();
        row.add(mLengthUnitsCombo);
        row.add(createLabel(I18n.Text("and"), null));
        mWeightUnitsCombo = createWeightUnitsPopup();
        row.add(mWeightUnitsCombo);
        row.add(createLabel(I18n.Text("for display of generated units"), null));
        column.add(row);

        row = new FlexRow();
        String seconds = I18n.Text("Seconds");
        row.add(createLabel(I18n.Text("Tooltip Timeout"), seconds));
        mToolTipTimeout = createTextField(seconds, Integer.valueOf(getToolTipTimeout()).toString());
        Dimension size = mToolTipTimeout.getPreferredSize();
        size.width = TextDrawing.getWidth(mToolTipTimeout.getFont(), Integer.valueOf(MAXIMUM_TOOLTIP_TIMEOUT).toString()) * 3 / 2;
        UIUtilities.setOnlySize(mToolTipTimeout, size);
        row.add(mToolTipTimeout);
        row.add(createLabel(I18n.Text("seconds"), null));
        column.add(row);

        row = new FlexRow();
        String blockLayoutTooltip = I18n.Text("Specifies the layout of the various blocks of data on the character sheet");
        row.add(createLabel(I18n.Text("Block Layout"), blockLayoutTooltip));
        column.add(row);
        row               = new FlexRow();
        mBlockLayoutField = createTextArea(blockLayoutTooltip, getBlockLayout());
        row.add(mBlockLayoutField);
        row.setFillVertical(true);
        column.add(row);

        column.apply(this);
    }

    private JComboBox<Scales> createUIScalePopup() {
        JComboBox<Scales> combo = new JComboBox<>(Scales.values());
        setupCombo(combo, null);
        combo.setSelectedItem(getInitialUIScale());
        combo.addActionListener(this);
        combo.setMaximumRowCount(combo.getItemCount());
        UIUtilities.setOnlySize(combo, combo.getPreferredSize());
        return combo;
    }

    private JComboBox<String> createLengthUnitsPopup() {
        JComboBox<String> combo = new JComboBox<>();
        setupCombo(combo, I18n.Text("The units to use for display of generated lengths"));
        for (LengthUnits unit : LengthUnits.values()) {
            combo.addItem(unit.getDescription());
        }
        combo.setSelectedIndex(getLengthUnits().ordinal());
        combo.addActionListener(this);
        combo.setMaximumRowCount(combo.getItemCount());
        UIUtilities.setOnlySize(combo, combo.getPreferredSize());
        return combo;
    }

    private JComboBox<String> createWeightUnitsPopup() {
        JComboBox<String> combo = new JComboBox<>();
        setupCombo(combo, I18n.Text("The units to use for display of generated weights"));
        for (WeightUnits unit : WeightUnits.values()) {
            combo.addItem(unit.getDescription());
        }
        combo.setSelectedIndex(getWeightUnits().ordinal());
        combo.addActionListener(this);
        combo.setMaximumRowCount(combo.getItemCount());
        UIUtilities.setOnlySize(combo, combo.getPreferredSize());
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
        Object source = event.getSource();
        if (source == mUIScaleCombo) {
            Preferences.getInstance().setValue(MODULE, SCALE_KEY, ((Scales) mUIScaleCombo.getSelectedItem()).getScale().getScale());
        } else if (source == mLengthUnitsCombo) {
            Preferences.getInstance().setValue(MODULE, LENGTH_UNITS_KEY, Enums.toId(LengthUnits.values()[mLengthUnitsCombo.getSelectedIndex()]));
        } else if (source == mWeightUnitsCombo) {
            Preferences.getInstance().setValue(MODULE, WEIGHT_UNITS_KEY, Enums.toId(WeightUnits.values()[mWeightUnitsCombo.getSelectedIndex()]));
        }
        adjustResetButton();
    }

    @Override
    public void reset() {
        mIncludeUnspentPointsInTotal.setSelected(DEFAULT_TOTAL_POINTS_DISPLAY);
        mUIScaleCombo.setSelectedItem(DEFAULT_SCALE);
        mLengthUnitsCombo.setSelectedIndex(DEFAULT_LENGTH_UNITS.ordinal());
        mWeightUnitsCombo.setSelectedIndex(DEFAULT_WEIGHT_UNITS.ordinal());
        mToolTipTimeout.setText(Integer.toString(DEFAULT_TOOLTIP_TIMEOUT));
        mBlockLayoutField.setText(DEFAULT_BLOCK_LAYOUT);
    }

    @Override
    public boolean isSetToDefaults() {
        return shouldIncludeUnspentPointsInTotalPointDisplay() == DEFAULT_TOTAL_POINTS_DISPLAY && showUserDescAsTooltip() == DEFAULT_SHOW_USER_DESC_AS_TOOL_TIP && getInitialUIScale() == DEFAULT_SCALE && mLengthUnitsCombo.getSelectedIndex() == DEFAULT_LENGTH_UNITS.ordinal() && mWeightUnitsCombo.getSelectedIndex() == DEFAULT_WEIGHT_UNITS.ordinal() && getToolTipTimeout() == DEFAULT_TOOLTIP_TIMEOUT && mBlockLayoutField.getText().equals(DEFAULT_BLOCK_LAYOUT);
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        Document document = event.getDocument();
        if (mBlockLayoutField.getDocument() == document) {
            Preferences.getInstance().setValue(MODULE, BLOCK_LAYOUT_KEY, mBlockLayoutField.getText());
        } else if (mToolTipTimeout.getDocument() == document) {
            int value = Numbers.extractInteger(mToolTipTimeout.getText(), DEFAULT_TOOLTIP_TIMEOUT, MINIMUM_TOOLTIP_TIMEOUT, MAXIMUM_TOOLTIP_TIMEOUT, true);
            Preferences.getInstance().setValue(MODULE, TOOLTIP_TIMEOUT_KEY, value);
            ToolTipManager.sharedInstance().setDismissDelay(getToolTipTimeout() * 1000);
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
            Preferences.getInstance().setValue(MODULE, TOTAL_POINTS_DISPLAY_KEY, mIncludeUnspentPointsInTotal.isSelected());
        } else if (source == mShowUserDescAsToolTips) {
            Preferences.getInstance().setValue(MODULE, SHOW_USER_DESC_AS_TOOL_TIP_KEY, mShowUserDescAsToolTips.isSelected());
        }
        adjustResetButton();
    }
}
