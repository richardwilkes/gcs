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

import com.trollworks.gcs.character.Profile;
import com.trollworks.gcs.character.Settings;
import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.ui.layout.Alignment;
import com.trollworks.gcs.ui.layout.FlexColumn;
import com.trollworks.gcs.ui.layout.FlexComponent;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.StdFileDialog;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.Preferences;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.io.File;
import javax.swing.JCheckBox;
import javax.swing.JTextField;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;

/** The sheet preferences panel. */
public class SheetPreferences extends PreferencePanel implements ActionListener, DocumentListener, ItemListener {
    public static final  String  MODULE                                  = "Sheet";
    private static final boolean DEFAULT_BASE_WILL_AND_PER_ON_10         = false;
    private static final boolean DEFAULT_USE_MULTIPLICATIVE_MODIFIERS    = false;
    private static final boolean DEFAULT_USE_MODIFYING_DICE_PLUS_ADDS    = false;
    private static final boolean DEFAULT_USE_KNOW_YOUR_OWN_STRENGTH      = false;
    private static final boolean DEFAULT_USE_REDUCED_SWING               = false;
    private static final boolean DEFAULT_USE_THRUST_EQUALS_SWING_MINUS_2 = false;
    private static final boolean DEFAULT_USE_SIMPLE_METRIC_CONVERSIONS   = true;
    private static final String  AUTO_NAME_NEW_CHARACTERS_KEY            = "auto_name_new_characters";
    private static final boolean DEFAULT_AUTO_NAME_NEW_CHARACTERS        = true;
    private static final String  INITIAL_POINTS_KEY                      = "initial_points";
    private static final int     DEFAULT_INITIAL_POINTS                  = 100;

    // Eliminate these after a suitable time period. Added May 25, 2020 after the 4.16 release.
    private static final String OLD_USE_OPTIONAL_IQ_RULES_KEY   = "UseOptionIQRules";
    private static final String OLD_OPTIONAL_MODIFIER_RULES_KEY = "UseOptionModifierRules";
    private static final String OLD_OPTIONAL_DICE_RULES_KEY     = "UseOptionDiceRules";
    private static final String OLD_OPTIONAL_STRENGTH_RULES_KEY = "UseOptionalStrengthRules";
    private static final String OLD_OPTIONAL_REDUCED_SWING_KEY  = "UseOptionalReducedSwing";
    private static final String OLD_OPTIONAL_THRUST_DAMAGE_KEY  = "UseOptionalThrustDamage";
    private static final String OLD_GURPS_METRIC_RULES_KEY      = "UseGurpsMetricRules";
    private static final String OLD_AUTO_NAME_KEY               = "AutoNameNewCharacters";
    private static final String OLD_INITIAL_POINTS_KEY          = "InitialPoints";

    private JTextField              mPlayerName;
    private JTextField              mTechLevel;
    private JTextField              mInitialPoints;
    private PortraitPreferencePanel mPortrait;
    private JCheckBox               mBaseWillAndPerOn10;
    private JCheckBox               mUseMultiplicativeModifiers;
    private JCheckBox               mUseModifyingDicePlusAdds;
    private JCheckBox               mUseKnowYourOwnStrength;
    private JCheckBox               mUseReducedSwing;
    private JCheckBox               mUseThrustEqualsSwingMinus2;
    private JCheckBox               mUseSimpleMetricConversions;
    private JCheckBox               mAutoNameNewCharacters;

    /** Initializes the services controlled by these preferences. */
    public static void initialize() {
        Preferences prefs = Preferences.getInstance();
        for (String key : prefs.getModuleKeys(MODULE)) {
            if (OLD_USE_OPTIONAL_IQ_RULES_KEY.equals(key)) {
                prefs.setValue(MODULE, Settings.TAG_BASE_WILL_AND_PER_ON_10, prefs.getBooleanValue(MODULE, OLD_USE_OPTIONAL_IQ_RULES_KEY, DEFAULT_BASE_WILL_AND_PER_ON_10));
                prefs.removePreference(MODULE, OLD_USE_OPTIONAL_IQ_RULES_KEY);
            } else if (OLD_OPTIONAL_MODIFIER_RULES_KEY.equals(key)) {
                prefs.setValue(MODULE, Settings.TAG_USE_MULTIPLICATIVE_MODIFIERS, prefs.getBooleanValue(MODULE, OLD_OPTIONAL_MODIFIER_RULES_KEY, DEFAULT_USE_MULTIPLICATIVE_MODIFIERS));
                prefs.removePreference(MODULE, OLD_OPTIONAL_MODIFIER_RULES_KEY);
            } else if (OLD_OPTIONAL_DICE_RULES_KEY.equals(key)) {
                prefs.setValue(MODULE, Settings.TAG_USE_MODIFYING_DICE_PLUS_ADDS, prefs.getBooleanValue(MODULE, OLD_OPTIONAL_DICE_RULES_KEY, DEFAULT_USE_MODIFYING_DICE_PLUS_ADDS));
                prefs.removePreference(MODULE, OLD_OPTIONAL_DICE_RULES_KEY);
            } else if (OLD_OPTIONAL_STRENGTH_RULES_KEY.equals(key)) {
                prefs.setValue(MODULE, Settings.TAG_USE_KNOW_YOUR_OWN_STRENGTH, prefs.getBooleanValue(MODULE, OLD_OPTIONAL_STRENGTH_RULES_KEY, DEFAULT_USE_KNOW_YOUR_OWN_STRENGTH));
                prefs.removePreference(MODULE, OLD_OPTIONAL_STRENGTH_RULES_KEY);
            } else if (OLD_OPTIONAL_REDUCED_SWING_KEY.equals(key)) {
                prefs.setValue(MODULE, Settings.TAG_USE_REDUCED_SWING, prefs.getBooleanValue(MODULE, OLD_OPTIONAL_REDUCED_SWING_KEY, DEFAULT_USE_REDUCED_SWING));
                prefs.removePreference(MODULE, OLD_OPTIONAL_REDUCED_SWING_KEY);
            } else if (OLD_OPTIONAL_THRUST_DAMAGE_KEY.equals(key)) {
                prefs.setValue(MODULE, Settings.TAG_USE_THRUST_EQUALS_SWING_MINUS_2, prefs.getBooleanValue(MODULE, OLD_OPTIONAL_THRUST_DAMAGE_KEY, DEFAULT_USE_THRUST_EQUALS_SWING_MINUS_2));
                prefs.removePreference(MODULE, OLD_OPTIONAL_THRUST_DAMAGE_KEY);
            } else if (OLD_GURPS_METRIC_RULES_KEY.equals(key)) {
                prefs.setValue(MODULE, Settings.TAG_USE_SIMPLE_METRIC_CONVERSIONS, prefs.getBooleanValue(MODULE, OLD_GURPS_METRIC_RULES_KEY, DEFAULT_USE_SIMPLE_METRIC_CONVERSIONS));
                prefs.removePreference(MODULE, OLD_GURPS_METRIC_RULES_KEY);
            } else if (OLD_AUTO_NAME_KEY.equals(key)) {
                prefs.setValue(MODULE, AUTO_NAME_NEW_CHARACTERS_KEY, prefs.getBooleanValue(MODULE, OLD_AUTO_NAME_KEY, DEFAULT_AUTO_NAME_NEW_CHARACTERS));
                prefs.removePreference(MODULE, OLD_AUTO_NAME_KEY);
            } else if (OLD_INITIAL_POINTS_KEY.equals(key)) {
                prefs.setValue(MODULE, INITIAL_POINTS_KEY, prefs.getIntValue(MODULE, OLD_INITIAL_POINTS_KEY, DEFAULT_INITIAL_POINTS));
                prefs.removePreference(MODULE, OLD_INITIAL_POINTS_KEY);
            }
        }
    }

    /** @return Whether Will and Perception should be based on 10 rather than IQ. */
    public static boolean baseWillAndPerOn10() {
        return Preferences.getInstance().getBooleanValue(MODULE, Settings.TAG_BASE_WILL_AND_PER_ON_10, DEFAULT_BASE_WILL_AND_PER_ON_10);
    }

    /** @return Whether to use the multiplicative modifier rules from PW102. */
    public static boolean useMultiplicativeModifiers() {
        return Preferences.getInstance().getBooleanValue(MODULE, Settings.TAG_USE_MULTIPLICATIVE_MODIFIERS, DEFAULT_USE_MULTIPLICATIVE_MODIFIERS);
    }

    /** @return Whether to use the dice modification rules from B269. */
    public static boolean useModifyingDicePlusAdds() {
        return Preferences.getInstance().getBooleanValue(MODULE, Settings.TAG_USE_MODIFYING_DICE_PLUS_ADDS, DEFAULT_USE_MODIFYING_DICE_PLUS_ADDS);
    }

    /** @return Whether to use the Know Your Own Strength rules from PY83. */
    public static boolean useKnowYourOwnStrength() {
        return Preferences.getInstance().getBooleanValue(MODULE, Settings.TAG_USE_KNOW_YOUR_OWN_STRENGTH, DEFAULT_USE_KNOW_YOUR_OWN_STRENGTH);
    }

    /**
     * @return Whether to use the Adjusting Swing Damage rules from noschoolgrognard.blogspot.com.
     *         Reduces KYOS damages if used together.
     */
    public static boolean useReducedSwing() {
        return Preferences.getInstance().getBooleanValue(MODULE, Settings.TAG_USE_REDUCED_SWING, DEFAULT_USE_REDUCED_SWING);
    }

    /** @return Whether to set thrust damage to swing-2. */
    public static boolean useThrustEqualsSwingMinus2() {
        return Preferences.getInstance().getBooleanValue(MODULE, Settings.TAG_USE_THRUST_EQUALS_SWING_MINUS_2, DEFAULT_USE_THRUST_EQUALS_SWING_MINUS_2);
    }

    /** @return Whether to use the simple metric conversion rules from B9. */
    public static boolean useSimpleMetricConversions() {
        return Preferences.getInstance().getBooleanValue(MODULE, Settings.TAG_USE_SIMPLE_METRIC_CONVERSIONS, DEFAULT_USE_SIMPLE_METRIC_CONVERSIONS);
    }

    /** @return Whether a new character should be automatically named. */
    public static boolean autoNameNewCharacters() {
        return Preferences.getInstance().getBooleanValue(MODULE, AUTO_NAME_NEW_CHARACTERS_KEY, DEFAULT_AUTO_NAME_NEW_CHARACTERS);
    }

    /** @return The initial points to start a new character with. */
    public static int getInitialPoints() {
        return Preferences.getInstance().getIntValue(MODULE, INITIAL_POINTS_KEY, DEFAULT_INITIAL_POINTS);
    }

    /**
     * Creates a new {@link SheetPreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public SheetPreferences(PreferencesWindow owner) {
        super(I18n.Text("Sheet"), owner);
        FlexColumn column = new FlexColumn();

        FlexGrid grid = new FlexGrid();
        column.add(grid);

        int rowIndex = 0;
        mPortrait = createPortrait();
        FlexComponent comp = new FlexComponent(mPortrait, Alignment.LEFT_TOP, Alignment.LEFT_TOP);
        grid.add(comp, rowIndex, 0, 4, 1);

        String playerTooltip = I18n.Text("The player name to use when a new character sheet is created");
        grid.add(createFlexLabel(I18n.Text("Player"), playerTooltip), rowIndex, 1);
        mPlayerName = createTextField(playerTooltip, Profile.getDefaultPlayerName());
        grid.add(mPlayerName, rowIndex++, 2);

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

        mAutoNameNewCharacters = createCheckBox(I18n.Text("Automatically name new characters"), null, autoNameNewCharacters());
        column.add(mAutoNameNewCharacters);

        mBaseWillAndPerOn10 = createCheckBox(I18n.Text("Base Will and Perception on 10 and not IQ"), null, baseWillAndPerOn10());
        column.add(mBaseWillAndPerOn10);

        mUseMultiplicativeModifiers = createCheckBox(I18n.Text("Use Multiplicative Modifiers from PW102 (note: changes point value)"), null, useMultiplicativeModifiers());
        column.add(mUseMultiplicativeModifiers);

        mUseModifyingDicePlusAdds = createCheckBox(I18n.Text("Use Modifying Dice + Adds from B269"), null, useModifyingDicePlusAdds());
        column.add(mUseModifyingDicePlusAdds);

        mUseKnowYourOwnStrength = createCheckBox(I18n.Text("Use strength rules from Knowing Your Own Strength (PY83)"), null, useKnowYourOwnStrength());
        column.add(mUseKnowYourOwnStrength);

        mUseReducedSwing = createCheckBox(I18n.Text("Use the reduced swing rules from Adjusting Swing Damage in Dungeon Fantasy"), "From noschoolgrognard.blogspot.com", useReducedSwing());
        column.add(mUseReducedSwing);

        mUseThrustEqualsSwingMinus2 = createCheckBox(I18n.Text("Use Thrust = Swing - 2"), null, useThrustEqualsSwingMinus2());
        column.add(mUseThrustEqualsSwingMinus2);

        mUseSimpleMetricConversions = createCheckBox(I18n.Text("Use the simple metric conversion rules from B9"), null, useSimpleMetricConversions());
        column.add(mUseSimpleMetricConversions);

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
        return StdFileDialog.showOpenDialog(null, I18n.Text("Select A Portrait"), FileType.IMAGE_FILTERS);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Object source = event.getSource();
        if (source == mPortrait) {
            File file = choosePortrait();
            if (file != null) {
                setPortrait(PathUtils.getFullPath(file));
            }
        }
        adjustResetButton();
    }

    @Override
    public void reset() {
        mPlayerName.setText(System.getProperty("user.name"));
        mTechLevel.setText(Profile.DEFAULT_TECH_LEVEL);
        mInitialPoints.setText(Integer.toString(DEFAULT_INITIAL_POINTS));
        setPortrait(Profile.DEFAULT_PORTRAIT);
        mAutoNameNewCharacters.setSelected(DEFAULT_AUTO_NAME_NEW_CHARACTERS);
        mUseModifyingDicePlusAdds.setSelected(DEFAULT_USE_MODIFYING_DICE_PLUS_ADDS);
        mBaseWillAndPerOn10.setSelected(DEFAULT_BASE_WILL_AND_PER_ON_10);
        mUseMultiplicativeModifiers.setSelected(DEFAULT_USE_MULTIPLICATIVE_MODIFIERS);
        mUseKnowYourOwnStrength.setSelected(DEFAULT_USE_KNOW_YOUR_OWN_STRENGTH);
        mUseThrustEqualsSwingMinus2.setSelected(DEFAULT_USE_THRUST_EQUALS_SWING_MINUS_2);
        mUseReducedSwing.setSelected(DEFAULT_USE_REDUCED_SWING);
        mUseSimpleMetricConversions.setSelected(DEFAULT_USE_SIMPLE_METRIC_CONVERSIONS);
    }

    @Override
    public boolean isSetToDefaults() {
        return Profile.getDefaultPlayerName().equals(System.getProperty("user.name")) && Profile.getDefaultPortraitPath().equals(Profile.DEFAULT_PORTRAIT) && Profile.getDefaultTechLevel().equals(Profile.DEFAULT_TECH_LEVEL) && getInitialPoints() == DEFAULT_INITIAL_POINTS && useModifyingDicePlusAdds() == DEFAULT_USE_MODIFYING_DICE_PLUS_ADDS && baseWillAndPerOn10() == DEFAULT_BASE_WILL_AND_PER_ON_10 && useMultiplicativeModifiers() == DEFAULT_USE_MULTIPLICATIVE_MODIFIERS && useKnowYourOwnStrength() == DEFAULT_USE_KNOW_YOUR_OWN_STRENGTH && useReducedSwing() == DEFAULT_USE_REDUCED_SWING && autoNameNewCharacters() == DEFAULT_AUTO_NAME_NEW_CHARACTERS;
    }

    private void setPortrait(String path) {
        Img image = Profile.getPortraitFromPortraitPath(path);
        Profile.setDefaultPortraitPath(path);
        mPortrait.setPortrait(image);
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        Document document = event.getDocument();
        if (mPlayerName.getDocument() == document) {
            Profile.setDefaultPlayerName(mPlayerName.getText());
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
        if (source == mBaseWillAndPerOn10) {
            Preferences.getInstance().setValue(MODULE, Settings.TAG_BASE_WILL_AND_PER_ON_10, mBaseWillAndPerOn10.isSelected());
        } else if (source == mUseMultiplicativeModifiers) {
            Preferences.getInstance().setValue(MODULE, Settings.TAG_USE_MULTIPLICATIVE_MODIFIERS, mUseMultiplicativeModifiers.isSelected());
        } else if (source == mUseModifyingDicePlusAdds) {
            Preferences.getInstance().setValue(MODULE, Settings.TAG_USE_MODIFYING_DICE_PLUS_ADDS, mUseModifyingDicePlusAdds.isSelected());
        } else if (source == mUseKnowYourOwnStrength) {
            Preferences.getInstance().setValue(MODULE, Settings.TAG_USE_KNOW_YOUR_OWN_STRENGTH, mUseKnowYourOwnStrength.isSelected());
        } else if (source == mUseReducedSwing) {
            Preferences.getInstance().setValue(MODULE, Settings.TAG_USE_REDUCED_SWING, mUseReducedSwing.isSelected());
        } else if (source == mUseThrustEqualsSwingMinus2) {
            Preferences.getInstance().setValue(MODULE, Settings.TAG_USE_THRUST_EQUALS_SWING_MINUS_2, mUseThrustEqualsSwingMinus2.isSelected());
        } else if (source == mUseSimpleMetricConversions) {
            Preferences.getInstance().setValue(MODULE, Settings.TAG_USE_SIMPLE_METRIC_CONVERSIONS, mUseSimpleMetricConversions.isSelected());
        } else if (source == mAutoNameNewCharacters) {
            Preferences.getInstance().setValue(MODULE, AUTO_NAME_NEW_CHARACTERS_KEY, mAutoNameNewCharacters.isSelected());
        }
        adjustResetButton();
    }
}
