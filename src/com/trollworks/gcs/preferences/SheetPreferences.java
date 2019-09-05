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

import com.trollworks.gcs.character.Profile;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.layout.Alignment;
import com.trollworks.toolkit.ui.layout.FlexColumn;
import com.trollworks.toolkit.ui.layout.FlexComponent;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.ui.preferences.PreferencePanel;
import com.trollworks.toolkit.ui.preferences.PreferencesWindow;
import com.trollworks.toolkit.ui.scale.Scales;
import com.trollworks.toolkit.ui.widget.StdFileDialog;
import com.trollworks.toolkit.utility.Dice;
import com.trollworks.toolkit.utility.FileType;
import com.trollworks.toolkit.utility.I18n;
import com.trollworks.toolkit.utility.PathUtils;
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
import java.io.File;
import java.util.ArrayList;
import java.util.List;

import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JTextField;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.filechooser.FileNameExtensionFilter;
import javax.swing.text.Document;

/** The sheet preferences panel. */
public class SheetPreferences extends PreferencePanel implements ActionListener, DocumentListener, ItemListener {
    static final String              MODULE                             = "Sheet";
    private static final String      OPTIONAL_DICE_RULES_KEY            = "UseOptionDiceRules";
    /** The optional dice rules preference key. */
    public static final String       OPTIONAL_DICE_RULES_PREF_KEY       = Preferences.getModuleKey(MODULE, OPTIONAL_DICE_RULES_KEY);
    private static final boolean     DEFAULT_OPTIONAL_DICE_RULES        = false;
    private static final String      OPTIONAL_IQ_RULES_KEY              = "UseOptionIQRules";
    /** The optional IQ rules preference key. */
    public static final String       OPTIONAL_IQ_RULES_PREF_KEY         = Preferences.getModuleKey(MODULE, OPTIONAL_IQ_RULES_KEY);
    private static final boolean     DEFAULT_OPTIONAL_IQ_RULES          = false;
    private static final String      OPTIONAL_MODIFIER_RULES_KEY        = "UseOptionModifierRules";
    /** The optional modifier rules preference key. */
    public static final String       OPTIONAL_MODIFIER_RULES_PREF_KEY   = Preferences.getModuleKey(MODULE, OPTIONAL_MODIFIER_RULES_KEY);
    private static final boolean     DEFAULT_OPTIONAL_MODIFIER_RULES    = false;
    private static final String      OPTIONAL_STRENGTH_RULES_KEY        = "UseOptionalStrengthRules";
    /** The optional Strength rules preference key. */
    public static final String       OPTIONAL_STRENGTH_RULES_PREF_KEY   = Preferences.getModuleKey(MODULE, OPTIONAL_STRENGTH_RULES_KEY);
    private static final boolean     DEFAULT_OPTIONAL_STRENGTH_RULES    = false;
    private static final String      OPTIONAL_REDUCED_SWING_KEY         = "UseOptionalReducedSwing";
    /** The optional Reduced Swing rules preference key. */
    public static final String       OPTIONAL_REDUCED_SWING_PREF_KEY    = Preferences.getModuleKey(MODULE, OPTIONAL_REDUCED_SWING_KEY);
    private static final boolean     DEFAULT_OPTIONAL_REDUCED_SWING     = false;
    private static final String      AUTO_NAME_KEY                      = "AutoNameNewCharacters";
    /** The auto-naming preference key. */
    public static final String       AUTO_NAME_PREF_KEY                 = Preferences.getModuleKey(MODULE, AUTO_NAME_KEY);
    private static final boolean     DEFAULT_AUTO_NAME                  = true;
    /** The optional Thrust Damage rules preference key. */
    private static final String      OPTIONAL_THRUST_DAMAGE_KEY         = "UseOptionalThurstDamage";
    public static final String       OPTIONAL_THRUST_DAMAGE_PREF_KEY    = Preferences.getModuleKey(MODULE, OPTIONAL_THRUST_DAMAGE_KEY);
    private static final boolean     DEFAULT_OPTIONAL_THRUST_DAMAGE     = false;
    private static final String      LENGTH_UNITS_KEY                   = "LengthUnits";
    /** The default length units preference key. */
    public static final String       LENGTH_UNITS_PREF_KEY              = Preferences.getModuleKey(MODULE, LENGTH_UNITS_KEY);
    private static final LengthUnits DEFAULT_LENGTH_UNITS               = LengthUnits.FT_IN;
    private static final String      WEIGHT_UNITS_KEY                   = "WeightUnits";
    /** The default weight units preference key. */
    public static final String       WEIGHT_UNITS_PREF_KEY              = Preferences.getModuleKey(MODULE, WEIGHT_UNITS_KEY);
    private static final WeightUnits DEFAULT_WEIGHT_UNITS               = WeightUnits.LB;
    private static final String      TOTAL_POINTS_DISPLAY_KEY           = "TotalPointsIncludesUnspentPoints";
    /** The total points includes unspent points preference key. */
    public static final String       TOTAL_POINTS_DISPLAY_PREF_KEY      = Preferences.getModuleKey(MODULE, TOTAL_POINTS_DISPLAY_KEY);
    private static final boolean     DEFAULT_TOTAL_POINTS_DISPLAY       = true;
    private static final String      GURPS_METRIC_RULES_KEY             = "UseGurpsMetricRules";
    /** The GURPS Metric preference key. */
    public static final String       GURPS_METRIC_RULES_PREF_KEY        = Preferences.getModuleKey(MODULE, GURPS_METRIC_RULES_KEY);

    public static final String       SHOW_USER_DESC_AS_TOOL_TIP_KEY     = "ShowUserDescAsToolTip";
    private static final boolean     DEFAULT_SHOW_USER_DESC_AS_TOOL_TIP = true;

    private static final boolean     DEFAULT_GURPS_METRIC_RULES         = true;
    private static final Scales      DEFAULT_SCALE                      = Scales.ACTUAL_SIZE;
    private static final String      SCALE_KEY                          = "UIScale";
    private static final String      INITIAL_POINTS_KEY                 = "InitialPoints";
    private static final int         DEFAULT_INITIAL_POINTS             = 100;
    private JTextField               mPlayerName;
    private JTextField               mCampaign;
    private JTextField               mTechLevel;
    private JTextField               mInitialPoints;
    private PortraitPreferencePanel  mPortrait;
    private JComboBox<Scales>        mUIScaleCombo;
    private JComboBox<String>        mLengthUnitsCombo;
    private JComboBox<String>        mWeightUnitsCombo;
    private JCheckBox                mUseOptionalDiceRules;
    private JCheckBox                mUseOptionalIQRules;
    private JCheckBox                mUseOptionalModifierRules;
    private JCheckBox                mUseOptionalStrengthRules;
    private JCheckBox                mUseOptionalReducedSwing;
    private JCheckBox                mIncludeUnspentPointsInTotal;
    private JCheckBox                mUseGurpsMetricRules;
    private JCheckBox                mAutoName;
    private JCheckBox                mShowUserDescAsToolTips;
    private JCheckBox                mUseOptionalThrustDamage;

    /** Initializes the services controlled by these preferences. */
    public static void initialize() {
        adjustOptionalDiceRulesProperty(areOptionalDiceRulesUsed());
    }

    /** @return The default length units to use. */
    public static LengthUnits getLengthUnits() {
        return Enums.extract(Preferences.getInstance().getStringValue(MODULE, LENGTH_UNITS_KEY), LengthUnits.values(), DEFAULT_LENGTH_UNITS);
    }

    /** @return The default weight units to use. */
    public static WeightUnits getWeightUnits() {
        return Enums.extract(Preferences.getInstance().getStringValue(MODULE, WEIGHT_UNITS_KEY), WeightUnits.values(), DEFAULT_WEIGHT_UNITS);
    }

    private static void adjustOptionalDiceRulesProperty(boolean use) {
        Dice.setConvertModifiersToExtraDice(use);
    }

    /** @return Whether the optional dice rules from B269 are in use. */
    public static boolean areOptionalDiceRulesUsed() {
        return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_DICE_RULES_KEY, DEFAULT_OPTIONAL_DICE_RULES);
    }

    /**
     * @return Whether the optional IQ rules (Will &amp; Perception are not based on IQ) are in use.
     */
    public static boolean areOptionalIQRulesUsed() {
        return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_IQ_RULES_KEY, DEFAULT_OPTIONAL_IQ_RULES);
    }

    /** @return Whether the optional modifier rules from PW102 are in use. */
    public static boolean areOptionalModifierRulesUsed() {
        return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_MODIFIER_RULES_KEY, DEFAULT_OPTIONAL_MODIFIER_RULES);
    }

    /** @return Whether the optional strength rules (KYOS) are in use. */
    public static boolean areOptionalStrengthRulesUsed() {
        return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_STRENGTH_RULES_KEY, DEFAULT_OPTIONAL_STRENGTH_RULES);
    }

    /** @return Whether the optional thrust damage (sw-2) rules are in use. */
    public static boolean areOptionalThrustDamageUsed() {
        return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_THRUST_DAMAGE_KEY, DEFAULT_OPTIONAL_THRUST_DAMAGE);
    }

    /**
     * @return Whether the optional reduced swing rules are in use. Reduces KYOS damages if used
     *         together.
     */
    public static boolean areOptionalReducedSwingUsed() {
        return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_REDUCED_SWING_KEY, DEFAULT_OPTIONAL_REDUCED_SWING);
    }

    /**
     * @return Whether the character's total points are displayed with or without including earned
     *         (but unspent) points.
     */
    public static boolean shouldIncludeUnspentPointsInTotalPointDisplay() {
        return Preferences.getInstance().getBooleanValue(MODULE, TOTAL_POINTS_DISPLAY_KEY, DEFAULT_TOTAL_POINTS_DISPLAY);
    }

    /** @return Whether the GURPS Metrics rules are used for weight and height conversion. */
    public static boolean areGurpsMetricRulesUsed() {
        return Preferences.getInstance().getBooleanValue(MODULE, GURPS_METRIC_RULES_KEY, DEFAULT_GURPS_METRIC_RULES);
    }

    /** @return Whether a new character should be automatically named. */
    public static boolean isNewCharacterAutoNamed() {
        return Preferences.getInstance().getBooleanValue(MODULE, AUTO_NAME_KEY, DEFAULT_AUTO_NAME);
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

    /** @return The initial points to start a new character with. */
    public static int getInitialPoints() {
        return Preferences.getInstance().getIntValue(MODULE, INITIAL_POINTS_KEY, DEFAULT_INITIAL_POINTS);
    }

    public static boolean showUserDescAsTooltip() {
        return Preferences.getInstance().getBooleanValue(MODULE, SHOW_USER_DESC_AS_TOOL_TIP_KEY, DEFAULT_SHOW_USER_DESC_AS_TOOL_TIP);
    }

    /**
     * Creates a new {@link SheetPreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public SheetPreferences(PreferencesWindow owner) {
        super(I18n.Text("Sheet"), owner);
        FlexColumn column = new FlexColumn();

        FlexGrid   grid   = new FlexGrid();
        column.add(grid);

        int rowIndex = 0;
        mPortrait = createPortrait();
        FlexComponent comp = new FlexComponent(mPortrait, Alignment.LEFT_TOP, Alignment.LEFT_TOP);
        grid.add(comp, rowIndex, 0, 4, 1);

        String playerTooltip = I18n.Text("The player name to use when a new character sheet is created");
        grid.add(createFlexLabel(I18n.Text("Player"), playerTooltip), rowIndex, 1);
        mPlayerName = createTextField(playerTooltip, Profile.getDefaultPlayerName());
        grid.add(mPlayerName, rowIndex++, 2);

        String campaignTooltiop = I18n.Text("The campaign to use when a new character sheet is created");
        grid.add(createFlexLabel(I18n.Text("Campaign"), campaignTooltiop), rowIndex, 1);
        mCampaign = createTextField(campaignTooltiop, Profile.getDefaultCampaign());
        grid.add(mCampaign, rowIndex++, 2);

        String techLevelTooltip = I18n.Text("<html><body>TL0: Stone Age (Prehistory and later)<br>TL1: Bronze Age (3500 B.C.+)<br>TL2: Iron Age (1200 B.C.+)<br>TL3: Medieval (600 A.D.+)<br>TL4: Age of Sail (1450+)<br>TL5: Industrial Revolution (1730+)<br>TL6: Mechanized Age (1880+)<br>TL7: Nuclear Age (1940+)<br>TL8: Digital Age (1980+)<br>TL9: Microtech Age (2025+?)<br>TL10: Robotic Age (2070+?)<br>TL11: Age of Exotic Matter<br>TL12: Anything Goes</body></html>");
        grid.add(createFlexLabel(I18n.Text("Tech Level"), techLevelTooltip), rowIndex, 1);
        mTechLevel = createTextField(techLevelTooltip, Profile.getDefaultTechLevel());
        grid.add(mTechLevel, rowIndex++, 2);

        String initialPointsTooltip = I18n.Text("The initial number of character points to start with");
        grid.add(createFlexLabel(I18n.Text("Initial Points"), initialPointsTooltip), rowIndex, 1);
        mInitialPoints = createTextField(initialPointsTooltip, Integer.toString(getInitialPoints()));
        grid.add(mInitialPoints, rowIndex++, 2);

        grid.add(new FlexSpacer(0, 0, false, false), rowIndex, 1);
        grid.add(new FlexSpacer(0, 0, true, false), rowIndex, 2);

        FlexRow row = new FlexRow();
        row.add(createLabel(I18n.Text("Use"), null));
        mLengthUnitsCombo = createLengthUnitsPopup();
        row.add(mLengthUnitsCombo);
        row.add(createLabel(I18n.Text("and"), null));
        mWeightUnitsCombo = createWeightUnitsPopup();
        row.add(mWeightUnitsCombo);
        row.add(createLabel(I18n.Text("for display of generated units"), null));
        column.add(row);

        mAutoName = createCheckBox(I18n.Text("Automatically name new characters"), null, isNewCharacterAutoNamed());
        column.add(mAutoName);

        mUseOptionalIQRules = createCheckBox(I18n.Text("Use optional rule: Will and Perception are not based upon IQ"), null, areOptionalIQRulesUsed());
        column.add(mUseOptionalIQRules);

        mUseOptionalModifierRules = createCheckBox(I18n.Text("Use optional rule \"Multiplicative Modifiers\" from PW102 (note: changes point value)"), null, areOptionalModifierRulesUsed());
        column.add(mUseOptionalModifierRules);

        mUseOptionalDiceRules = createCheckBox(I18n.Text("Use optional rule \"Modifying Dice + Adds\" from B269"), null, areOptionalDiceRulesUsed());
        column.add(mUseOptionalDiceRules);

        mUseOptionalStrengthRules = createCheckBox(I18n.Text("Use optional strength rules from the \"Knowing Your Own Strength\" Pyramid article"), null, areOptionalStrengthRulesUsed());
        column.add(mUseOptionalStrengthRules);

        mUseOptionalReducedSwing = createCheckBox(I18n.Text("Use optional Reduced Swing rules from the  \"Adjusting Swing Damage in Dungeon Fantasy\" No School Grognard article"), null, areOptionalReducedSwingUsed());
        column.add(mUseOptionalReducedSwing);

        mUseOptionalThrustDamage = createCheckBox(I18n.Text("Use optional Thrust damage (Thrust = Swing -2)"), null, areOptionalThrustDamageUsed());
        column.add(mUseOptionalThrustDamage);

        mIncludeUnspentPointsInTotal = createCheckBox(I18n.Text("Character point total display includes unspent points"), null, shouldIncludeUnspentPointsInTotalPointDisplay());
        column.add(mIncludeUnspentPointsInTotal);

        mUseGurpsMetricRules = createCheckBox(I18n.Text("Use GURPS Metric rules for height, weight, encumbrance and lifting things"), null, areGurpsMetricRulesUsed());
        column.add(mUseGurpsMetricRules);

        mShowUserDescAsToolTips = createCheckBox(I18n.Text("Show Advantage User Description as a Tooltip"), null, showUserDescAsTooltip());
        column.add(mShowUserDescAsToolTips);

        row = new FlexRow();
        row.add(createLabel(I18n.Text("Use"), null));
        mUIScaleCombo = createUIScalePopup();
        row.add(mUIScaleCombo);
        row.add(createLabel(I18n.Text("for the initial scale when opening character sheets, templates and lists"), null, SwingConstants.LEFT));
        column.add(row);

        column.add(new FlexSpacer(0, 0, false, true));

        column.apply(this);
    }

    private FlexComponent createFlexLabel(String title, String tooltip) {
        return new FlexComponent(createLabel(title, tooltip), Alignment.RIGHT_BOTTOM, Alignment.CENTER);
    }

    private PortraitPreferencePanel createPortrait() {
        PortraitPreferencePanel panel = new PortraitPreferencePanel(Profile.getPortraitFromPortraitPath(Profile.getDefaultPortraitPath()));
        panel.addActionListener(this);
        add(panel);
        return panel;
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

    public static File choosePortrait() {
        List<FileNameExtensionFilter> filters = new ArrayList<>();
        filters.add(new FileNameExtensionFilter(I18n.Text("All Readable Image Files"), FileType.PNG_EXTENSION, FileType.JPEG_EXTENSION, "jpeg", FileType.GIF_EXTENSION));
        filters.add(FileType.getPngFilter());
        filters.add(FileType.getJpegFilter());
        filters.add(FileType.getGifFilter());
        return StdFileDialog.showOpenDialog(null, I18n.Text("Select A Portrait"), filters);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Object source = event.getSource();
        if (source == mPortrait) {
            File file = choosePortrait();
            if (file != null) {
                setPortrait(PathUtils.getFullPath(file));
            }
        } else if (source == mUIScaleCombo) {
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
        mPlayerName.setText(System.getProperty("user.name"));
        mCampaign.setText("");
        mTechLevel.setText(Profile.DEFAULT_TECH_LEVEL);
        mInitialPoints.setText(Integer.toString(DEFAULT_INITIAL_POINTS));
        setPortrait(Profile.DEFAULT_PORTRAIT);
        mUIScaleCombo.setSelectedItem(DEFAULT_SCALE);
        mLengthUnitsCombo.setSelectedIndex(DEFAULT_LENGTH_UNITS.ordinal());
        mWeightUnitsCombo.setSelectedIndex(DEFAULT_WEIGHT_UNITS.ordinal());
        mAutoName.setSelected(DEFAULT_AUTO_NAME);
        mUseOptionalDiceRules.setSelected(DEFAULT_OPTIONAL_DICE_RULES);
        mUseOptionalIQRules.setSelected(DEFAULT_OPTIONAL_IQ_RULES);
        mUseOptionalModifierRules.setSelected(DEFAULT_OPTIONAL_MODIFIER_RULES);
        mUseOptionalStrengthRules.setSelected(DEFAULT_OPTIONAL_STRENGTH_RULES);
        mUseOptionalThrustDamage.setSelected(DEFAULT_OPTIONAL_THRUST_DAMAGE);
        mUseOptionalReducedSwing.setSelected(DEFAULT_OPTIONAL_REDUCED_SWING);
        mIncludeUnspentPointsInTotal.setSelected(DEFAULT_TOTAL_POINTS_DISPLAY);
        mUseGurpsMetricRules.setSelected(DEFAULT_GURPS_METRIC_RULES);
    }

    @Override
    public boolean isSetToDefaults() {
        return Profile.getDefaultPlayerName().equals(System.getProperty("user.name")) && Profile.getDefaultCampaign().equals("") && Profile.getDefaultPortraitPath().equals(Profile.DEFAULT_PORTRAIT) && Profile.getDefaultTechLevel().equals(Profile.DEFAULT_TECH_LEVEL) && getInitialPoints() == DEFAULT_INITIAL_POINTS && getInitialUIScale() == DEFAULT_SCALE && areOptionalDiceRulesUsed() == DEFAULT_OPTIONAL_DICE_RULES && areOptionalIQRulesUsed() == DEFAULT_OPTIONAL_IQ_RULES && areOptionalModifierRulesUsed() == DEFAULT_OPTIONAL_MODIFIER_RULES && areOptionalStrengthRulesUsed() == DEFAULT_OPTIONAL_STRENGTH_RULES && areOptionalReducedSwingUsed() == DEFAULT_OPTIONAL_REDUCED_SWING && isNewCharacterAutoNamed() == DEFAULT_AUTO_NAME;
    }

    private void setPortrait(String path) {
        StdImage image = Profile.getPortraitFromPortraitPath(path);
        Profile.setDefaultPortraitPath(path);
        mPortrait.setPortrait(image);
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        Document document = event.getDocument();
        if (mPlayerName.getDocument() == document) {
            Profile.setDefaultPlayerName(mPlayerName.getText());
        } else if (mCampaign.getDocument() == document) {
            Profile.setDefaultCampaign(mCampaign.getText());
        } else if (mTechLevel.getDocument() == document) {
            Profile.setDefaultTechLevel(mTechLevel.getText());
        } else if (mInitialPoints.getDocument() == document) {
            Preferences.getInstance().setValue(MODULE, INITIAL_POINTS_KEY, Numbers.extractInteger(mInitialPoints.getText(), 0, true));
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
        if (source == mUseOptionalDiceRules) {
            boolean checked = mUseOptionalDiceRules.isSelected();
            adjustOptionalDiceRulesProperty(checked);
            Preferences.getInstance().setValue(MODULE, OPTIONAL_DICE_RULES_KEY, checked);
        } else if (source == mUseOptionalIQRules) {
            Preferences.getInstance().setValue(MODULE, OPTIONAL_IQ_RULES_KEY, mUseOptionalIQRules.isSelected());
        } else if (source == mUseOptionalModifierRules) {
            Preferences.getInstance().setValue(MODULE, OPTIONAL_MODIFIER_RULES_KEY, mUseOptionalModifierRules.isSelected());
        } else if (source == mUseOptionalStrengthRules) {
            Preferences.getInstance().setValue(MODULE, OPTIONAL_STRENGTH_RULES_KEY, mUseOptionalStrengthRules.isSelected());
        } else if (source == mUseOptionalThrustDamage) {
            Preferences.getInstance().setValue(MODULE, OPTIONAL_THRUST_DAMAGE_KEY, mUseOptionalThrustDamage.isSelected());
        } else if (source == mUseOptionalReducedSwing) {
            Preferences.getInstance().setValue(MODULE, OPTIONAL_REDUCED_SWING_KEY, mUseOptionalReducedSwing.isSelected());
        } else if (source == mIncludeUnspentPointsInTotal) {
            Preferences.getInstance().setValue(MODULE, TOTAL_POINTS_DISPLAY_KEY, mIncludeUnspentPointsInTotal.isSelected());
        } else if (source == mUseGurpsMetricRules) {
            Preferences.getInstance().setValue(MODULE, GURPS_METRIC_RULES_KEY, mUseGurpsMetricRules.isSelected());
        } else if (source == mShowUserDescAsToolTips) {
            Preferences.getInstance().setValue(MODULE, SHOW_USER_DESC_AS_TOOL_TIP_KEY, mShowUserDescAsToolTips.isSelected());
        } else if (source == mAutoName) {
            Preferences.getInstance().setValue(MODULE, AUTO_NAME_KEY, mAutoName.isSelected());
        }
        adjustResetButton();
    }
}
