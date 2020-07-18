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
import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.StdFileDialog;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;

import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.nio.file.Path;
import javax.swing.JCheckBox;
import javax.swing.JLabel;
import javax.swing.JTextField;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;

/** The sheet preferences panel. */
public class SheetPreferences extends PreferencePanel implements ActionListener, DocumentListener, ItemListener {
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

    /**
     * Creates a new {@link SheetPreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public SheetPreferences(PreferencesWindow owner) {
        super(I18n.Text("Sheet"), owner);
        setLayout(new PrecisionLayout().setColumns(3));
        Preferences prefs = Preferences.getInstance();

        mPortrait = new PortraitPreferencePanel(Profile.getPortraitFromPortraitPath(prefs.getDefaultPortraitPath()));
        mPortrait.addActionListener(this);
        add(mPortrait, new PrecisionLayoutData().setVerticalSpan(11).setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));

        String playerTooltip = I18n.Text("The player name to use when a new character sheet is created");
        addLabel(I18n.Text("Player"), playerTooltip);
        mPlayerName = addTextField(playerTooltip, prefs.getDefaultPlayerName());

        String techLevelTooltip = I18n.Text("<html><body>TL0: Stone Age (Prehistory and later)<br>TL1: Bronze Age (3500 B.C.+)<br>TL2: Iron Age (1200 B.C.+)<br>TL3: Medieval (600 A.D.+)<br>TL4: Age of Sail (1450+)<br>TL5: Industrial Revolution (1730+)<br>TL6: Mechanized Age (1880+)<br>TL7: Nuclear Age (1940+)<br>TL8: Digital Age (1980+)<br>TL9: Microtech Age (2025+?)<br>TL10: Robotic Age (2070+?)<br>TL11: Age of Exotic Matter<br>TL12: Anything Goes</body></html>");
        addLabel(I18n.Text("Tech Level"), techLevelTooltip);
        mTechLevel = addTextField(techLevelTooltip, prefs.getDefaultTechLevel());

        String initialPointsTooltip = I18n.Text("The initial number of character points to start with");
        addLabel(I18n.Text("Initial Points"), initialPointsTooltip);
        mInitialPoints = addTextField(initialPointsTooltip, Integer.toString(prefs.getInitialPoints()));

        mAutoNameNewCharacters = addCheckBox(I18n.Text("Automatically name new characters"), null, prefs.autoNameNewCharacters());
        mBaseWillAndPerOn10 = addCheckBox(I18n.Text("Base Will and Perception on 10 and not IQ *"), null, prefs.baseWillAndPerOn10());
        mUseMultiplicativeModifiers = addCheckBox(I18n.Text("Use Multiplicative Modifiers from PW102 (note: changes point value) *"), null, prefs.useMultiplicativeModifiers());
        mUseModifyingDicePlusAdds = addCheckBox(I18n.Text("Use Modifying Dice + Adds from B269 *"), null, prefs.useModifyingDicePlusAdds());
        mUseKnowYourOwnStrength = addCheckBox(I18n.Text("Use strength rules from Knowing Your Own Strength (PY83) *"), null, prefs.useKnowYourOwnStrength());
        mUseReducedSwing = addCheckBox(I18n.Text("Use the reduced swing rules from Adjusting Swing Damage in Dungeon Fantasy *"), "From noschoolgrognard.blogspot.com", prefs.useReducedSwing());
        mUseThrustEqualsSwingMinus2 = addCheckBox(I18n.Text("Use Thrust = Swing - 2 *"), null, prefs.useThrustEqualsSwingMinus2());
        mUseSimpleMetricConversions = addCheckBox(I18n.Text("Use the simple metric conversion rules from B9 *"), null, prefs.useSimpleMetricConversions());

        JLabel label = new JLabel(I18n.Text("* To change the setting on existing sheets, use the per-sheet settings available from the toolbar"));
        label.setOpaque(false);
        add(label, new PrecisionLayoutData().setHorizontalSpan(3).setAlignment(PrecisionLayoutAlignment.MIDDLE, PrecisionLayoutAlignment.END).setGrabVerticalSpace(true));
    }

    private JLabel addLabel(String title, String tooltip) {
        JLabel label = new JLabel(title, SwingConstants.RIGHT);
        label.setOpaque(false);
        label.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        add(label, new PrecisionLayoutData().setFillHorizontalAlignment());
        return label;
    }

    private JTextField addTextField(String tooltip, String value) {
        JTextField field = new JTextField(value);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.getDocument().addDocumentListener(this);
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment());
        return field;
    }

    private JCheckBox addCheckBox(String title, String tooltip, boolean checked) {
        JCheckBox checkbox = new JCheckBox(title, checked);
        checkbox.setOpaque(false);
        checkbox.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        checkbox.addItemListener(this);
        add(checkbox, new PrecisionLayoutData().setHorizontalSpan(2));
        return checkbox;
    }

    public static Path choosePortrait() {
        return StdFileDialog.showOpenDialog(null, I18n.Text("Select A Portrait"), FileType.IMAGE_FILTERS);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Object source = event.getSource();
        if (source == mPortrait) {
            Path path = choosePortrait();
            if (path != null) {
                setPortrait(path.normalize().toAbsolutePath().toString());
            }
        }
        adjustResetButton();
    }

    private void setPortrait(String path) {
        Img image = Profile.getPortraitFromPortraitPath(path);
        Preferences.getInstance().setDefaultPortraitPath(path);
        mPortrait.setPortrait(image);
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
        if (source == mBaseWillAndPerOn10) {
            prefs.setBaseWillAndPerOn10(mBaseWillAndPerOn10.isSelected());
        } else if (source == mUseMultiplicativeModifiers) {
            prefs.setUseMultiplicativeModifiers(mUseMultiplicativeModifiers.isSelected());
        } else if (source == mUseModifyingDicePlusAdds) {
            prefs.setUseModifyingDicePlusAdds(mUseModifyingDicePlusAdds.isSelected());
        } else if (source == mUseKnowYourOwnStrength) {
            prefs.setUseKnowYourOwnStrength(mUseKnowYourOwnStrength.isSelected());
        } else if (source == mUseReducedSwing) {
            prefs.setUseReducedSwing(mUseReducedSwing.isSelected());
        } else if (source == mUseThrustEqualsSwingMinus2) {
            prefs.setUseThrustEqualsSwingMinus2(mUseThrustEqualsSwingMinus2.isSelected());
        } else if (source == mUseSimpleMetricConversions) {
            prefs.setUseSimpleMetricConversions(mUseSimpleMetricConversions.isSelected());
        } else if (source == mAutoNameNewCharacters) {
            prefs.setAutoNameNewCharacters(mAutoNameNewCharacters.isSelected());
        }
        adjustResetButton();
    }

    @Override
    public void reset() {
        mPlayerName.setText(Preferences.DEFAULT_DEFAULT_PLAYER_NAME);
        mTechLevel.setText(Preferences.DEFAULT_DEFAULT_TECH_LEVEL);
        mInitialPoints.setText(Integer.toString(Preferences.DEFAULT_INITIAL_POINTS));
        setPortrait(Preferences.DEFAULT_DEFAULT_PORTRAIT_PATH);
        mAutoNameNewCharacters.setSelected(Preferences.DEFAULT_AUTO_NAME_NEW_CHARACTERS);
        mUseModifyingDicePlusAdds.setSelected(Preferences.DEFAULT_USE_MODIFYING_DICE_PLUS_ADDS);
        mBaseWillAndPerOn10.setSelected(Preferences.DEFAULT_BASE_WILL_AND_PER_ON_10);
        mUseMultiplicativeModifiers.setSelected(Preferences.DEFAULT_USE_MULTIPLICATIVE_MODIFIERS);
        mUseKnowYourOwnStrength.setSelected(Preferences.DEFAULT_USE_KNOW_YOUR_OWN_STRENGTH);
        mUseThrustEqualsSwingMinus2.setSelected(Preferences.DEFAULT_USE_THRUST_EQUALS_SWING_MINUS_2);
        mUseReducedSwing.setSelected(Preferences.DEFAULT_USE_REDUCED_SWING);
        mUseSimpleMetricConversions.setSelected(Preferences.DEFAULT_USE_SIMPLE_METRIC_CONVERSIONS);
    }

    @Override
    public boolean isSetToDefaults() {
        Preferences prefs     = Preferences.getInstance();
        boolean     atDefault = prefs.getDefaultPlayerName().equals(Preferences.DEFAULT_DEFAULT_PLAYER_NAME);
        atDefault = atDefault && prefs.getDefaultPortraitPath().equals(Preferences.DEFAULT_DEFAULT_PORTRAIT_PATH);
        atDefault = atDefault && prefs.getDefaultTechLevel().equals(Preferences.DEFAULT_DEFAULT_TECH_LEVEL);
        atDefault = atDefault && prefs.getInitialPoints() == Preferences.DEFAULT_INITIAL_POINTS;
        atDefault = atDefault && prefs.useModifyingDicePlusAdds() == Preferences.DEFAULT_USE_MODIFYING_DICE_PLUS_ADDS;
        atDefault = atDefault && prefs.baseWillAndPerOn10() == Preferences.DEFAULT_BASE_WILL_AND_PER_ON_10;
        atDefault = atDefault && prefs.useMultiplicativeModifiers() == Preferences.DEFAULT_USE_MULTIPLICATIVE_MODIFIERS;
        atDefault = atDefault && prefs.useKnowYourOwnStrength() == Preferences.DEFAULT_USE_KNOW_YOUR_OWN_STRENGTH;
        atDefault = atDefault && prefs.useReducedSwing() == Preferences.DEFAULT_USE_REDUCED_SWING;
        atDefault = atDefault && prefs.autoNameNewCharacters() == Preferences.DEFAULT_AUTO_NAME_NEW_CHARACTERS;
        return atDefault;
    }
}
