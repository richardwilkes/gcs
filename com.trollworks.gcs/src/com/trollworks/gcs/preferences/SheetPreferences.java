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
import com.trollworks.gcs.ui.layout.Alignment;
import com.trollworks.gcs.ui.layout.FlexColumn;
import com.trollworks.gcs.ui.layout.FlexComponent;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.StdFileDialog;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.nio.file.Path;
import javax.swing.JCheckBox;
import javax.swing.JTextField;
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
        Preferences prefs  = Preferences.getInstance();
        FlexColumn  column = new FlexColumn();

        FlexGrid grid = new FlexGrid();
        column.add(grid);

        int rowIndex = 0;
        mPortrait = createPortrait();
        FlexComponent comp = new FlexComponent(mPortrait, Alignment.LEFT_TOP, Alignment.LEFT_TOP);
        grid.add(comp, rowIndex, 0, 4, 1);

        String playerTooltip = I18n.Text("The player name to use when a new character sheet is created");
        grid.add(createFlexLabel(I18n.Text("Player"), playerTooltip), rowIndex, 1);
        mPlayerName = createTextField(playerTooltip, prefs.getDefaultPlayerName());
        grid.add(mPlayerName, rowIndex++, 2);

        String techLevelTooltip = I18n.Text("<html><body>TL0: Stone Age (Prehistory and later)<br>TL1: Bronze Age (3500 B.C.+)<br>TL2: Iron Age (1200 B.C.+)<br>TL3: Medieval (600 A.D.+)<br>TL4: Age of Sail (1450+)<br>TL5: Industrial Revolution (1730+)<br>TL6: Mechanized Age (1880+)<br>TL7: Nuclear Age (1940+)<br>TL8: Digital Age (1980+)<br>TL9: Microtech Age (2025+?)<br>TL10: Robotic Age (2070+?)<br>TL11: Age of Exotic Matter<br>TL12: Anything Goes</body></html>");
        grid.add(createFlexLabel(I18n.Text("Tech Level"), techLevelTooltip), rowIndex, 1);
        mTechLevel = createTextField(techLevelTooltip, prefs.getDefaultTechLevel());
        grid.add(mTechLevel, rowIndex++, 2);

        String initialPointsTooltip = I18n.Text("The initial number of character points to start with");
        grid.add(createFlexLabel(I18n.Text("Initial Points"), initialPointsTooltip), rowIndex, 1);
        mInitialPoints = createTextField(initialPointsTooltip, Integer.toString(prefs.getInitialPoints()));
        grid.add(mInitialPoints, rowIndex++, 2);

        grid.add(new FlexSpacer(0, 0, false, false), rowIndex, 1);
        grid.add(new FlexSpacer(0, 0, true, false), rowIndex, 2);

        mAutoNameNewCharacters = createCheckBox(I18n.Text("Automatically name new characters"), null, prefs.autoNameNewCharacters());
        column.add(mAutoNameNewCharacters);

        mBaseWillAndPerOn10 = createCheckBox(I18n.Text("Base Will and Perception on 10 and not IQ"), null, prefs.baseWillAndPerOn10());
        column.add(mBaseWillAndPerOn10);

        mUseMultiplicativeModifiers = createCheckBox(I18n.Text("Use Multiplicative Modifiers from PW102 (note: changes point value)"), null, prefs.useMultiplicativeModifiers());
        column.add(mUseMultiplicativeModifiers);

        mUseModifyingDicePlusAdds = createCheckBox(I18n.Text("Use Modifying Dice + Adds from B269"), null, prefs.useModifyingDicePlusAdds());
        column.add(mUseModifyingDicePlusAdds);

        mUseKnowYourOwnStrength = createCheckBox(I18n.Text("Use strength rules from Knowing Your Own Strength (PY83)"), null, prefs.useKnowYourOwnStrength());
        column.add(mUseKnowYourOwnStrength);

        mUseReducedSwing = createCheckBox(I18n.Text("Use the reduced swing rules from Adjusting Swing Damage in Dungeon Fantasy"), "From noschoolgrognard.blogspot.com", prefs.useReducedSwing());
        column.add(mUseReducedSwing);

        mUseThrustEqualsSwingMinus2 = createCheckBox(I18n.Text("Use Thrust = Swing - 2"), null, prefs.useThrustEqualsSwingMinus2());
        column.add(mUseThrustEqualsSwingMinus2);

        mUseSimpleMetricConversions = createCheckBox(I18n.Text("Use the simple metric conversion rules from B9"), null, prefs.useSimpleMetricConversions());
        column.add(mUseSimpleMetricConversions);

        column.add(new FlexSpacer(0, 0, false, true));

        column.apply(this);
    }

    private FlexComponent createFlexLabel(String title, String tooltip) {
        return new FlexComponent(createLabel(title, tooltip), Alignment.RIGHT_BOTTOM, Alignment.CENTER);
    }

    private PortraitPreferencePanel createPortrait() {
        PortraitPreferencePanel panel = new PortraitPreferencePanel(Profile.getPortraitFromPortraitPath(Preferences.getInstance().getDefaultPortraitPath()));
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
