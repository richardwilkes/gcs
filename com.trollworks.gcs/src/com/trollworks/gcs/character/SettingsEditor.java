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

package com.trollworks.gcs.character;

import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.datafile.DataChangeListener;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.WeightUnits;

import java.awt.BorderLayout;
import java.awt.EventQueue;
import java.awt.FlowLayout;
import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.awt.event.WindowEvent;
import java.util.List;
import javax.swing.JButton;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JTextArea;
import javax.swing.SwingConstants;
import javax.swing.border.CompoundBorder;
import javax.swing.border.EmptyBorder;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

public class SettingsEditor extends BaseWindow implements ActionListener, DocumentListener, ItemListener, CloseHandler, NotifierTarget, DataChangeListener, Runnable {
    private GURPSCharacter           mCharacter;
    private Settings                 mSettings;
    private JCheckBox                mBaseWillOn10;
    private JCheckBox                mBasePerOn10;
    private JCheckBox                mUseMultiplicativeModifiers;
    private JCheckBox                mUseModifyingDicePlusAdds;
    private JCheckBox                mUseKnowYourOwnStrength;
    private JCheckBox                mUseReducedSwing;
    private JCheckBox                mUseThrustEqualsSwingMinus2;
    private JCheckBox                mUseSimpleMetricConversions;
    private JCheckBox                mShowCollegeInSpells;
    private JCheckBox                mShowDifficulty;
    private JCheckBox                mShowAdvantageModifierAdj;
    private JCheckBox                mShowEquipmentModifierAdj;
    private JCheckBox                mShowSpellAdj;
    private JCheckBox                mShowTitleInsteadOfNameInPageFooter;
    private JCheckBox                mExtraSpaceAroundEncumbrance;
    private JComboBox<LengthUnits>   mLengthUnitsCombo;
    private JComboBox<WeightUnits>   mWeightUnitsCombo;
    private JComboBox<DisplayOption> mUserDescriptionDisplayCombo;
    private JComboBox<DisplayOption> mModifiersDisplayCombo;
    private JComboBox<DisplayOption> mNotesDisplayCombo;
    private JTextArea                mBlockLayoutField;
    private JButton                  mResetButton;
    private boolean                  mUpdatePending;

    public static SettingsEditor find(GURPSCharacter character) {
        for (Window window : Window.getWindows()) {
            if (window.isShowing() && window instanceof SettingsEditor) {
                SettingsEditor wnd = (SettingsEditor) window;
                if (wnd.mCharacter == character) {
                    return wnd;
                }
            }
        }
        return null;
    }

    public static void display(GURPSCharacter character) {
        if (!UIUtilities.inModalState()) {
            SettingsEditor wnd = find(character);
            if (wnd == null) {
                wnd = new SettingsEditor(character);
            }
            wnd.setVisible(true);
        }
    }

    private static String createTitle(GURPSCharacter character) {
        return String.format(I18n.Text("Sheet Settings: %s"), character.getProfile().getName());
    }

    private SettingsEditor(GURPSCharacter character) {
        super(createTitle(character));
        mCharacter = character;
        mSettings = character.getSettings();
        addTopPanel();
        addResetPanel();
        adjustResetButton();
        restoreBounds();
        character.addTarget(this, Profile.ID_NAME);
        Preferences.getInstance().addChangeListener(this);
    }

    private void addTopPanel() {
        JPanel panel = new JPanel(new PrecisionLayout().setColumns(2).setMargins(10));
        mShowCollegeInSpells = addCheckBox(panel, I18n.Text("Show the College column in the spells list"), null, mSettings.showCollegeInSpells());
        mShowDifficulty = addCheckBox(panel, I18n.Text("Show the Difficulty column in the skills and spells lists"), null, mSettings.showDifficulty());
        mShowAdvantageModifierAdj = addCheckBox(panel, I18n.Text("Show the advantage modifier cost adjustments in the advantage list"), null, mSettings.showAdvantageModifierAdj());
        mShowEquipmentModifierAdj = addCheckBox(panel, I18n.Text("Show the equipment modifier cost and weight adjustments in the equipment lists"), null, mSettings.showEquipmentModifierAdj());
        mShowSpellAdj = addCheckBox(panel, I18n.Text("Show the spell ritual, cost and time adjustments in the spell list"), null, mSettings.showSpellAdj());
        mShowTitleInsteadOfNameInPageFooter = addCheckBox(panel, I18n.Text("Show the title rather than the name in the page footer"), null, mSettings.useTitleInFooter());
        mExtraSpaceAroundEncumbrance = addCheckBox(panel, I18n.Text("Add extra space around Encumbrance table rather than around FP/HP table"), null, mSettings.extraSpaceAroundEncumbrance());
        mBaseWillOn10 = addCheckBox(panel, I18n.Text("Base Will on 10 and not IQ"), null, mSettings.baseWillOn10());
        mBasePerOn10 = addCheckBox(panel, I18n.Text("Base Perception on 10 and not IQ"), null, mSettings.basePerOn10());
        mUseMultiplicativeModifiers = addCheckBox(panel, I18n.Text("Use Multiplicative Modifiers from PW102 (note: changes point value)"), null, mSettings.useMultiplicativeModifiers());
        mUseModifyingDicePlusAdds = addCheckBox(panel, I18n.Text("Use Modifying Dice + Adds from B269"), null, mSettings.useModifyingDicePlusAdds());
        mUseKnowYourOwnStrength = addCheckBox(panel, I18n.Text("Use strength rules from Knowing Your Own Strength (PY83)"), null, mSettings.useKnowYourOwnStrength());
        mUseReducedSwing = addCheckBox(panel, I18n.Text("Use the reduced swing rules from Adjusting Swing Damage in Dungeon Fantasy"), "From noschoolgrognard.blogspot.com", mSettings.useReducedSwing());
        mUseThrustEqualsSwingMinus2 = addCheckBox(panel, I18n.Text("Use Thrust = Swing - 2"), null, mSettings.useThrustEqualsSwingMinus2());
        mUseSimpleMetricConversions = addCheckBox(panel, I18n.Text("Use the simple metric conversion rules from B9"), null, mSettings.useSimpleMetricConversions());

        addLabel(panel, I18n.Text("Length Units"));
        mLengthUnitsCombo = addCombo(panel, LengthUnits.values(), mSettings.defaultLengthUnits(), I18n.Text("The units to use for display of generated lengths"));

        addLabel(panel, I18n.Text("Weight Units"));
        mWeightUnitsCombo = addCombo(panel, WeightUnits.values(), mSettings.defaultWeightUnits(), I18n.Text("The units to use for display of generated weights"));

        addLabel(panel, I18n.Text("Show User Description"));
        String tooltip = I18n.Text("Where to display this information");
        mUserDescriptionDisplayCombo = addCombo(panel, DisplayOption.values(), mSettings.userDescriptionDisplay(), tooltip);

        addLabel(panel, I18n.Text("Show Modifiers"));
        mModifiersDisplayCombo = addCombo(panel, DisplayOption.values(), mSettings.modifiersDisplay(), tooltip);

        addLabel(panel, I18n.Text("Show Notes"));
        mNotesDisplayCombo = addCombo(panel, DisplayOption.values(), mSettings.notesDisplay(), tooltip);

        JLabel label = new JLabel(I18n.Text("Block Layout"));
        label.setOpaque(false);
        panel.add(label, new PrecisionLayoutData().setHorizontalSpan(2));

        String blockLayoutTooltip = Text.wrapPlainTextForToolTip(I18n.Text("Specifies the layout of the various blocks of data on the character sheet"));
        mBlockLayoutField = new JTextArea(Preferences.linesToString(mSettings.blockLayout()));
        mBlockLayoutField.setToolTipText(blockLayoutTooltip);
        mBlockLayoutField.getDocument().addDocumentListener(this);
        mBlockLayoutField.setBorder(new CompoundBorder(new LineBorder(), new EmptyBorder(0, 4, 0, 4)));
        panel.add(mBlockLayoutField, new PrecisionLayoutData().setHorizontalSpan(2).setFillAlignment().setGrabSpace(true));

        getContentPane().add(panel);
    }

    private void addResetPanel() {
        JPanel panel = new JPanel(new FlowLayout(FlowLayout.CENTER));
        mResetButton = new JButton(I18n.Text("Reset to Global Preference Values"));
        mResetButton.addActionListener(this);
        panel.add(mResetButton);
        getContentPane().add(panel, BorderLayout.SOUTH);
    }

    private void addLabel(JPanel panel, String title) {
        JLabel label = new JLabel(title, SwingConstants.RIGHT);
        label.setOpaque(false);
        panel.add(label, new PrecisionLayoutData().setFillHorizontalAlignment());
    }

    private <E> JComboBox<E> addCombo(JPanel panel, E[] values, E choice, String tooltip) {
        JComboBox<E> combo = new JComboBox<>(values);
        combo.setOpaque(false);
        combo.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        combo.setSelectedItem(choice);
        combo.addActionListener(this);
        combo.setMaximumRowCount(combo.getItemCount());
        panel.add(combo);
        return combo;
    }

    private JCheckBox addCheckBox(JPanel panel, String title, String tooltip, boolean checked) {
        JCheckBox checkbox = new JCheckBox(title, checked);
        checkbox.setOpaque(false);
        checkbox.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        checkbox.addItemListener(this);
        panel.add(checkbox, new PrecisionLayoutData().setHorizontalSpan(2));
        return checkbox;
    }

    @Override
    public void itemStateChanged(ItemEvent event) {
        Object source = event.getSource();
        if (source == mShowCollegeInSpells) {
            mSettings.setShowCollegeInSpells(mShowCollegeInSpells.isSelected());
        } else if (source == mShowDifficulty) {
            mSettings.setShowDifficulty(mShowDifficulty.isSelected());
        } else if (source == mShowAdvantageModifierAdj) {
            mSettings.setShowAdvantageModifierAdj(mShowAdvantageModifierAdj.isSelected());
        } else if (source == mShowEquipmentModifierAdj) {
            mSettings.setShowEquipmentModifierAdj(mShowEquipmentModifierAdj.isSelected());
        } else if (source == mShowSpellAdj) {
            mSettings.setShowSpellAdj(mShowSpellAdj.isSelected());
        } else if (source == mShowTitleInsteadOfNameInPageFooter) {
            mSettings.setUseTitleInFooter(mShowTitleInsteadOfNameInPageFooter.isSelected());
        } else if (source == mExtraSpaceAroundEncumbrance) {
            mSettings.setExtraSpaceAroundEncumbrance(mExtraSpaceAroundEncumbrance.isSelected());
        } else if (source == mBaseWillOn10) {
            mSettings.setBaseWillOn10(mBaseWillOn10.isSelected());
        } else if (source == mBasePerOn10) {
            mSettings.setBasePerOn10(mBasePerOn10.isSelected());
        } else if (source == mUseMultiplicativeModifiers) {
            mSettings.setUseMultiplicativeModifiers(mUseMultiplicativeModifiers.isSelected());
        } else if (source == mUseModifyingDicePlusAdds) {
            mSettings.setUseModifyingDicePlusAdds(mUseModifyingDicePlusAdds.isSelected());
        } else if (source == mUseKnowYourOwnStrength) {
            mSettings.setUseKnowYourOwnStrength(mUseKnowYourOwnStrength.isSelected());
        } else if (source == mUseReducedSwing) {
            mSettings.setUseReducedSwing(mUseReducedSwing.isSelected());
        } else if (source == mUseThrustEqualsSwingMinus2) {
            mSettings.setUseThrustEqualsSwingMinus2(mUseThrustEqualsSwingMinus2.isSelected());
        } else if (source == mUseSimpleMetricConversions) {
            mSettings.setUseSimpleMetricConversions(mUseSimpleMetricConversions.isSelected());
        }
        adjustResetButton();
    }

    private void adjustResetButton() {
        mResetButton.setEnabled(!isSetToDefaults());
    }

    private boolean isSetToDefaults() {
        Preferences prefs      = Preferences.getInstance();
        boolean     atDefaults = mUseModifyingDicePlusAdds.isSelected() == prefs.useModifyingDicePlusAdds();
        atDefaults = atDefaults && mShowCollegeInSpells.isSelected() == prefs.showCollegeInSheetSpells();
        atDefaults = atDefaults && mShowDifficulty.isSelected() == prefs.showDifficulty();
        atDefaults = atDefaults && mShowAdvantageModifierAdj.isSelected() == prefs.showAdvantageModifierAdj();
        atDefaults = atDefaults && mShowEquipmentModifierAdj.isSelected() == prefs.showEquipmentModifierAdj();
        atDefaults = atDefaults && mShowSpellAdj.isSelected() == prefs.showSpellAdj();
        atDefaults = atDefaults && mShowTitleInsteadOfNameInPageFooter.isSelected() == prefs.useTitleInFooter();
        atDefaults = atDefaults && mExtraSpaceAroundEncumbrance.isSelected() == prefs.extraSpaceAroundEncumbrance();
        atDefaults = atDefaults && mBaseWillOn10.isSelected() == prefs.baseWillOn10();
        atDefaults = atDefaults && mBasePerOn10.isSelected() == prefs.basePerOn10();
        atDefaults = atDefaults && mUseMultiplicativeModifiers.isSelected() == prefs.useMultiplicativeModifiers();
        atDefaults = atDefaults && mUseKnowYourOwnStrength.isSelected() == prefs.useKnowYourOwnStrength();
        atDefaults = atDefaults && mUseThrustEqualsSwingMinus2.isSelected() == prefs.useThrustEqualsSwingMinus2();
        atDefaults = atDefaults && mUseReducedSwing.isSelected() == prefs.useReducedSwing();
        atDefaults = atDefaults && mUseSimpleMetricConversions.isSelected() == prefs.useSimpleMetricConversions();
        atDefaults = atDefaults && mLengthUnitsCombo.getSelectedItem() == prefs.getDefaultLengthUnits();
        atDefaults = atDefaults && mWeightUnitsCombo.getSelectedItem() == prefs.getDefaultWeightUnits();
        atDefaults = atDefaults && mUserDescriptionDisplayCombo.getSelectedItem() == prefs.getUserDescriptionDisplay();
        atDefaults = atDefaults && mModifiersDisplayCombo.getSelectedItem() == prefs.getModifiersDisplay();
        atDefaults = atDefaults && mNotesDisplayCombo.getSelectedItem() == prefs.getNotesDisplay();
        atDefaults = atDefaults && mBlockLayoutField.getText().equals(Preferences.linesToString(prefs.getBlockLayout()));
        return atDefaults;
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Object source = event.getSource();
        if (source == mLengthUnitsCombo) {
            mSettings.setDefaultLengthUnits((LengthUnits) mLengthUnitsCombo.getSelectedItem());
        } else if (source == mWeightUnitsCombo) {
            mSettings.setDefaultWeightUnits((WeightUnits) mWeightUnitsCombo.getSelectedItem());
        } else if (source == mUserDescriptionDisplayCombo) {
            mSettings.setUserDescriptionDisplay((DisplayOption) mUserDescriptionDisplayCombo.getSelectedItem());
        } else if (source == mModifiersDisplayCombo) {
            mSettings.setModifiersDisplay((DisplayOption) mModifiersDisplayCombo.getSelectedItem());
        } else if (source == mNotesDisplayCombo) {
            mSettings.setNotesDisplay((DisplayOption) mNotesDisplayCombo.getSelectedItem());
        } else if (source == mResetButton) {
            Preferences prefs = Preferences.getInstance();
            mUseModifyingDicePlusAdds.setSelected(prefs.useModifyingDicePlusAdds());
            mShowCollegeInSpells.setSelected(prefs.showCollegeInSheetSpells());
            mShowDifficulty.setSelected(prefs.showDifficulty());
            mShowAdvantageModifierAdj.setSelected(prefs.showAdvantageModifierAdj());
            mShowEquipmentModifierAdj.setSelected(prefs.showEquipmentModifierAdj());
            mShowSpellAdj.setSelected(prefs.showSpellAdj());
            mShowTitleInsteadOfNameInPageFooter.setSelected(prefs.useTitleInFooter());
            mExtraSpaceAroundEncumbrance.setSelected(prefs.extraSpaceAroundEncumbrance());
            mBaseWillOn10.setSelected(prefs.baseWillOn10());
            mBasePerOn10.setSelected(prefs.basePerOn10());
            mUseMultiplicativeModifiers.setSelected(prefs.useMultiplicativeModifiers());
            mUseKnowYourOwnStrength.setSelected(prefs.useKnowYourOwnStrength());
            mUseThrustEqualsSwingMinus2.setSelected(prefs.useThrustEqualsSwingMinus2());
            mUseReducedSwing.setSelected(prefs.useReducedSwing());
            mUseSimpleMetricConversions.setSelected(prefs.useSimpleMetricConversions());
            mLengthUnitsCombo.setSelectedItem(prefs.getDefaultLengthUnits());
            mWeightUnitsCombo.setSelectedItem(prefs.getDefaultWeightUnits());
            mUserDescriptionDisplayCombo.setSelectedItem(prefs.getUserDescriptionDisplay());
            mModifiersDisplayCombo.setSelectedItem(prefs.getModifiersDisplay());
            mNotesDisplayCombo.setSelectedItem(prefs.getNotesDisplay());
            mBlockLayoutField.setText(Preferences.linesToString(prefs.getBlockLayout()));
        }
        adjustResetButton();
    }

    @Override
    public String getWindowPrefsKey() {
        return "settings_editor";
    }

    @Override
    public boolean mayAttemptClose() {
        return true;
    }

    @Override
    public boolean attemptClose() {
        windowClosing(new WindowEvent(this, WindowEvent.WINDOW_CLOSING));
        return true;
    }

    @Override
    public void dispose() {
        mCharacter.removeTarget(this);
        Preferences.getInstance().removeChangeListener(this);
        super.dispose();
    }

    @Override
    public void dataWasChanged() {
        if (!mUpdatePending) {
            mUpdatePending = true;
            EventQueue.invokeLater(this);
        }
    }

    @Override
    public void run() {
        setTitle(createTitle(mCharacter));
        adjustResetButton();
        mUpdatePending = false;
    }

    @Override
    public int getNotificationPriority() {
        return 0;
    }

    @Override
    public void handleNotification(Object producer, String name, Object data) {
        dataWasChanged();
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
    public void changedUpdate(DocumentEvent event) {
        mSettings.setBlockLayout(List.of(mBlockLayoutField.getText().split("\n")));
        adjustResetButton();
    }
}
