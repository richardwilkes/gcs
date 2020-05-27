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
import com.trollworks.gcs.character.Settings;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.layout.FlexColumn;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.scale.Scales;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Preferences;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.WeightUnits;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JLabel;
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
    private static final String        MODULE                           = "Display";
    private static final String        TOTAL_POINTS_DISPLAY_KEY         = "TotalPointsIncludesUnspentPoints";
    public static final  String        TOTAL_POINTS_DISPLAY_PREF_KEY    = Preferences.getModuleKey(MODULE, TOTAL_POINTS_DISPLAY_KEY);
    private static final boolean       DEFAULT_TOTAL_POINTS_DISPLAY     = true;
    private static final String        DEFAULT_BLOCK_LAYOUT             = "melee\nranged\nadvantages skills\nspells\nequipment\nother_equipment\nnotes";
    private static final Scales        DEFAULT_SCALE                    = Scales.QUARTER_AGAIN_SIZE;
    private static final String        SCALE_KEY                        = "UIScale.v2";
    public static final  LengthUnits   DEFAULT_LENGTH_UNITS             = LengthUnits.FT_IN;
    public static final  WeightUnits   DEFAULT_WEIGHT_UNITS             = WeightUnits.LB;
    private static final String        TOOLTIP_TIMEOUT_KEY              = "TooltipTimeout";
    private static final int           MINIMUM_TOOLTIP_TIMEOUT          = 1;
    private static final int           MAXIMUM_TOOLTIP_TIMEOUT          = 9999;
    private static final int           DEFAULT_TOOLTIP_TIMEOUT          = 60;
    public static final  DisplayOption DEFAULT_USER_DESCRIPTION_DISPLAY = DisplayOption.TOOLTIP;
    public static final  DisplayOption DEFAULT_MODIFIERS_DISPLAY        = DisplayOption.INLINE;
    public static final  DisplayOption DEFAULT_NOTES_DISPLAY            = DisplayOption.INLINE;

    // Eliminate these after a suitable time period. Added May 25, 2020 after the 4.16 release.
    private static final String OLD_LENGTH_UNITS_KEY = "LengthUnits";
    private static final String OLD_WEIGHT_UNITS_KEY = "WeightUnits";
    private static final String OLD_BLOCK_LAYOUT_KEY = "BlockLayout";

    private JCheckBox                mIncludeUnspentPointsInTotal;
    private JComboBox<Scales>        mUIScaleCombo;
    private JComboBox<LengthUnits>   mLengthUnitsCombo;
    private JComboBox<WeightUnits>   mWeightUnitsCombo;
    private JComboBox<DisplayOption> mUserDescriptionDisplayCombo;
    private JComboBox<DisplayOption> mModifiersDisplayCombo;
    private JComboBox<DisplayOption> mNotesDisplayCombo;
    private JTextField               mToolTipTimeout;
    private JTextArea                mBlockLayoutField;

    /** Initializes the services controlled by these preferences. */
    public static void initialize() {
        Preferences prefs = Preferences.getInstance();
        for (String key : prefs.getModuleKeys(MODULE)) {
            if (OLD_LENGTH_UNITS_KEY.equals(key)) {
                prefs.setValue(MODULE, Settings.TAG_DEFAULT_LENGTH_UNITS, prefs.getStringValue(MODULE, OLD_LENGTH_UNITS_KEY, Enums.toId(DEFAULT_LENGTH_UNITS)));
                prefs.removePreference(MODULE, OLD_LENGTH_UNITS_KEY);
            } else if (OLD_WEIGHT_UNITS_KEY.equals(key)) {
                prefs.setValue(MODULE, Settings.TAG_DEFAULT_WEIGHT_UNITS, prefs.getStringValue(MODULE, OLD_WEIGHT_UNITS_KEY, Enums.toId(DEFAULT_WEIGHT_UNITS)));
                prefs.removePreference(MODULE, OLD_WEIGHT_UNITS_KEY);
            } else if (OLD_BLOCK_LAYOUT_KEY.equals(key)) {
                prefs.setValue(MODULE, Settings.TAG_BLOCK_LAYOUT, prefs.getStringValue(MODULE, OLD_BLOCK_LAYOUT_KEY, DEFAULT_BLOCK_LAYOUT));
                prefs.removePreference(MODULE, OLD_BLOCK_LAYOUT_KEY);
            }
        }
    }

    /** @return The timeout, in seconds, to use for tooltips. */
    public static int getToolTipTimeout() {
        return Preferences.getInstance().getIntValue(MODULE, TOOLTIP_TIMEOUT_KEY, DEFAULT_TOOLTIP_TIMEOUT);
    }

    public static DisplayOption userDescriptionDisplay() {
        return Enums.extract(Preferences.getInstance().getStringValue(MODULE, Settings.TAG_USER_DESCRIPTION_DISPLAY), DisplayOption.values(), DEFAULT_USER_DESCRIPTION_DISPLAY);
    }

    public static DisplayOption modifiersDisplay() {
        return Enums.extract(Preferences.getInstance().getStringValue(MODULE, Settings.TAG_MODIFIERS_DISPLAY), DisplayOption.values(), DEFAULT_MODIFIERS_DISPLAY);
    }

    public static DisplayOption notesDisplay() {
        return Enums.extract(Preferences.getInstance().getStringValue(MODULE, Settings.TAG_NOTES_DISPLAY), DisplayOption.values(), DEFAULT_NOTES_DISPLAY);
    }

    /**
     * @return Whether the character's total points are displayed with or without including earned
     *         (but unspent) points.
     */
    public static boolean shouldIncludeUnspentPointsInTotalPointDisplay() {
        return Preferences.getInstance().getBooleanValue(MODULE, TOTAL_POINTS_DISPLAY_KEY, DEFAULT_TOTAL_POINTS_DISPLAY);
    }

    /** @return The {@link Scales} to use when opening a new scalable file. */
    public static Scales initialUIScale() {
        double value = Preferences.getInstance().getDoubleValue(MODULE, SCALE_KEY, DEFAULT_SCALE.getScale().getScale());
        for (Scales one : Scales.values()) {
            if (one.getScale().getScale() == value) {
                return one;
            }
        }
        return Scales.ACTUAL_SIZE;
    }

    /** @return The default length units to use. */
    public static LengthUnits defaultLengthUnits() {
        return Enums.extract(Preferences.getInstance().getStringValue(MODULE, Settings.TAG_DEFAULT_LENGTH_UNITS), LengthUnits.values(), DEFAULT_LENGTH_UNITS);
    }

    /** @return The default weight units to use. */
    public static WeightUnits defaultWeightUnits() {
        return Enums.extract(Preferences.getInstance().getStringValue(MODULE, Settings.TAG_DEFAULT_WEIGHT_UNITS), WeightUnits.values(), DEFAULT_WEIGHT_UNITS);
    }

    /** @return The block layout settings. */
    public static String blockLayout() {
        return Preferences.getInstance().getStringValue(MODULE, Settings.TAG_BLOCK_LAYOUT, DEFAULT_BLOCK_LAYOUT);
    }

    /**
     * Creates a new {@link DisplayPreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public DisplayPreferences(PreferencesWindow owner) {
        super(I18n.Text("Display"), owner);
        FlexColumn column = new FlexColumn();

        mIncludeUnspentPointsInTotal = createCheckBox(I18n.Text("Character point total display includes unspent points"), null, shouldIncludeUnspentPointsInTotalPointDisplay());
        column.add(mIncludeUnspentPointsInTotal);

        FlexRow row = new FlexRow();
        row.add(createLabel(I18n.Text("Use"), null));
        mUIScaleCombo = createCombo(Scales.values(), initialUIScale(), null);
        row.add(mUIScaleCombo);
        JLabel label = createLabel(I18n.Text("for the initial scale when opening character sheets, templates and lists"), null, SwingConstants.LEFT);
        UIUtilities.setOnlySize(label, null);
        row.add(label);
        column.add(row);

        row = new FlexRow();
        row.add(createLabel(I18n.Text("Use"), null));
        mLengthUnitsCombo = createCombo(LengthUnits.values(), defaultLengthUnits(), I18n.Text("The units to use for display of generated lengths"));
        row.add(mLengthUnitsCombo);
        row.add(createLabel(I18n.Text("and"), null));
        mWeightUnitsCombo = createCombo(WeightUnits.values(), defaultWeightUnits(), I18n.Text("The units to use for display of generated weights"));
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
        row.add(createLabel(I18n.Text("Show User Description"), null));
        mUserDescriptionDisplayCombo = createCombo(DisplayOption.values(), userDescriptionDisplay(), I18n.Text("Where to display this information"));
        row.add(mUserDescriptionDisplayCombo);
        column.add(row);

        row = new FlexRow();
        row.add(createLabel(I18n.Text("Show Modifiers"), null));
        mModifiersDisplayCombo = createCombo(DisplayOption.values(), modifiersDisplay(), I18n.Text("Where to display this information"));
        row.add(mModifiersDisplayCombo);
        column.add(row);

        row = new FlexRow();
        row.add(createLabel(I18n.Text("Show Notes"), null));
        mNotesDisplayCombo = createCombo(DisplayOption.values(), notesDisplay(), I18n.Text("Where to display this information"));
        row.add(mNotesDisplayCombo);
        column.add(row);

        row = new FlexRow();
        String blockLayoutTooltip = I18n.Text("Specifies the layout of the various blocks of data on the character sheet");
        row.add(createLabel(I18n.Text("Block Layout"), blockLayoutTooltip));
        column.add(row);
        row = new FlexRow();
        mBlockLayoutField = createTextArea(blockLayoutTooltip, blockLayout());
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
        Object source = event.getSource();
        if (source == mUIScaleCombo) {
            Scales scale = (Scales) mUIScaleCombo.getSelectedItem();
            if (scale == null) {
                scale = Scales.ACTUAL_SIZE;
            }
            Preferences.getInstance().setValue(MODULE, SCALE_KEY, scale.getScale().getScale());
        } else if (source == mLengthUnitsCombo) {
            Preferences.getInstance().setValue(MODULE, Settings.TAG_DEFAULT_LENGTH_UNITS, Enums.toId(LengthUnits.values()[mLengthUnitsCombo.getSelectedIndex()]));
        } else if (source == mWeightUnitsCombo) {
            Preferences.getInstance().setValue(MODULE, Settings.TAG_DEFAULT_WEIGHT_UNITS, Enums.toId(WeightUnits.values()[mWeightUnitsCombo.getSelectedIndex()]));
        } else if (source == mUserDescriptionDisplayCombo) {
            Preferences.getInstance().setValue(MODULE, Settings.TAG_USER_DESCRIPTION_DISPLAY, Enums.toId(DisplayOption.values()[mUserDescriptionDisplayCombo.getSelectedIndex()]));
        } else if (source == mModifiersDisplayCombo) {
            Preferences.getInstance().setValue(MODULE, Settings.TAG_MODIFIERS_DISPLAY, Enums.toId(DisplayOption.values()[mModifiersDisplayCombo.getSelectedIndex()]));
        } else if (source == mNotesDisplayCombo) {
            Preferences.getInstance().setValue(MODULE, Settings.TAG_NOTES_DISPLAY, Enums.toId(DisplayOption.values()[mNotesDisplayCombo.getSelectedIndex()]));
        }
        adjustResetButton();
    }

    @Override
    public void reset() {
        mIncludeUnspentPointsInTotal.setSelected(DEFAULT_TOTAL_POINTS_DISPLAY);
        mUIScaleCombo.setSelectedItem(DEFAULT_SCALE);
        mLengthUnitsCombo.setSelectedItem(DEFAULT_LENGTH_UNITS);
        mWeightUnitsCombo.setSelectedItem(DEFAULT_WEIGHT_UNITS);
        mToolTipTimeout.setText(Integer.toString(DEFAULT_TOOLTIP_TIMEOUT));
        mBlockLayoutField.setText(DEFAULT_BLOCK_LAYOUT);
        mUserDescriptionDisplayCombo.setSelectedItem(DEFAULT_USER_DESCRIPTION_DISPLAY);
        mModifiersDisplayCombo.setSelectedItem(DEFAULT_MODIFIERS_DISPLAY);
        mNotesDisplayCombo.setSelectedItem(DEFAULT_NOTES_DISPLAY);
    }

    @Override
    public boolean isSetToDefaults() {
        return shouldIncludeUnspentPointsInTotalPointDisplay() == DEFAULT_TOTAL_POINTS_DISPLAY && initialUIScale() == DEFAULT_SCALE && mLengthUnitsCombo.getSelectedItem() == DEFAULT_LENGTH_UNITS && mWeightUnitsCombo.getSelectedItem() == DEFAULT_WEIGHT_UNITS && getToolTipTimeout() == DEFAULT_TOOLTIP_TIMEOUT && mBlockLayoutField.getText().equals(DEFAULT_BLOCK_LAYOUT) && userDescriptionDisplay() == DEFAULT_USER_DESCRIPTION_DISPLAY && modifiersDisplay() == DEFAULT_MODIFIERS_DISPLAY && notesDisplay() == DEFAULT_NOTES_DISPLAY;
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        Document document = event.getDocument();
        if (mBlockLayoutField.getDocument() == document) {
            Preferences.getInstance().setValue(MODULE, Settings.TAG_BLOCK_LAYOUT, mBlockLayoutField.getText());
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
        }
        adjustResetButton();
    }
}
