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

import com.trollworks.gcs.attribute.Attribute;
import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.attribute.AttributeEditor;
import com.trollworks.gcs.datafile.DataChangeListener;
import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.utility.I18n;
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
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import javax.swing.JButton;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JTextArea;
import javax.swing.ScrollPaneConstants;
import javax.swing.SwingConstants;
import javax.swing.border.EmptyBorder;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

public class SettingsEditor extends BaseWindow implements ActionListener, DocumentListener, ItemListener, CloseHandler, DataChangeListener, Runnable {
    private GURPSCharacter           mCharacter;
    private Settings                 mSettings;
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
    private JComboBox<LengthUnits>   mLengthUnitsCombo;
    private JComboBox<WeightUnits>   mWeightUnitsCombo;
    private JComboBox<DisplayOption> mUserDescriptionDisplayCombo;
    private JComboBox<DisplayOption> mModifiersDisplayCombo;
    private JComboBox<DisplayOption> mNotesDisplayCombo;
    private JTextArea                mBlockLayoutField;
    private AttributeEditor          mAttributeEditor;
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
        character.addChangeListener(this);
        Preferences.getInstance().addChangeListener(this);
    }

    private void addTopPanel() {
        JPanel left = new JPanel(new PrecisionLayout().setColumns(2));
        mShowCollegeInSpells = addCheckBox(left, I18n.Text("Show the College column"), null, mSettings.showCollegeInSpells());
        mShowDifficulty = addCheckBox(left, I18n.Text("Show the Difficulty column"), null, mSettings.showDifficulty());
        mShowAdvantageModifierAdj = addCheckBox(left, I18n.Text("Show advantage modifier cost adjustments"), null, mSettings.showAdvantageModifierAdj());
        mShowEquipmentModifierAdj = addCheckBox(left, I18n.Text("Show equipment modifier cost & weight adjustments"), null, mSettings.showEquipmentModifierAdj());
        mShowSpellAdj = addCheckBox(left, I18n.Text("Show spell ritual, cost & time adjustments"), null, mSettings.showSpellAdj());
        mShowTitleInsteadOfNameInPageFooter = addCheckBox(left, I18n.Text("Show the title instead of the name in the footer"), null, mSettings.useTitleInFooter());
        addLabel(left, I18n.Text("Show User Description"));
        String tooltip = I18n.Text("Where to display this information");
        mUserDescriptionDisplayCombo = addCombo(left, DisplayOption.values(), mSettings.userDescriptionDisplay(), tooltip);
        addLabel(left, I18n.Text("Show Modifiers"));
        mModifiersDisplayCombo = addCombo(left, DisplayOption.values(), mSettings.modifiersDisplay(), tooltip);
        addLabel(left, I18n.Text("Show Notes"));
        mNotesDisplayCombo = addCombo(left, DisplayOption.values(), mSettings.notesDisplay(), tooltip);

        JPanel right = new JPanel(new PrecisionLayout().setColumns(2));
        mUseMultiplicativeModifiers = addCheckBox(right, I18n.Text("Use Multiplicative Modifiers (PW102; changes point value)"), null, mSettings.useMultiplicativeModifiers());
        mUseModifyingDicePlusAdds = addCheckBox(right, I18n.Text("Use Modifying Dice + Adds (B269)"), null, mSettings.useModifyingDicePlusAdds());
        mUseKnowYourOwnStrength = addCheckBox(right, I18n.Text("Use strength rules from Knowing Your Own Strength (PY83)"), null, mSettings.useKnowYourOwnStrength());
        mUseReducedSwing = addCheckBox(right, I18n.Text("Use the reduced swing rules"), "From \"Adjusting Swing Damage in Dungeon Fantasy\" found on noschoolgrognard.blogspot.com", mSettings.useReducedSwing());
        mUseThrustEqualsSwingMinus2 = addCheckBox(right, I18n.Text("Use Thrust = Swing - 2"), null, mSettings.useThrustEqualsSwingMinus2());
        mUseSimpleMetricConversions = addCheckBox(right, I18n.Text("Use the simple metric conversion rules (B9)"), null, mSettings.useSimpleMetricConversions());
        addLabel(right, I18n.Text("Length Units"));
        mLengthUnitsCombo = addCombo(right, LengthUnits.values(), mSettings.defaultLengthUnits(), I18n.Text("The units to use for display of generated lengths"));
        addLabel(right, I18n.Text("Weight Units"));
        mWeightUnitsCombo = addCombo(right, WeightUnits.values(), mSettings.defaultWeightUnits(), I18n.Text("The units to use for display of generated weights"));

        JPanel top = new JPanel(new PrecisionLayout().setColumns(2));
        top.add(left, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        top.add(right, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));

        JPanel panel = new JPanel(new PrecisionLayout().setMargins(10));
        panel.add(top, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment());

        mAttributeEditor = new AttributeEditor(mSettings.getAttributes(), () -> {
            Map<String, Attribute> oldAttributes = mCharacter.getAttributes();
            Map<String, Attribute> newAttributes = new HashMap<>();
            for (String key : mCharacter.getSettings().getAttributes().keySet()) {
                Attribute attribute = oldAttributes.get(key);
                newAttributes.put(key, attribute != null ? attribute : new Attribute(key));
            }
            oldAttributes.clear();
            oldAttributes.putAll(newAttributes);
            mCharacter.notifyOfChange();
            adjustResetButton();
        });
        panel.add(mAttributeEditor, new PrecisionLayoutData().setFillAlignment().setGrabSpace(true));

        String blockLayoutTooltip = Text.wrapPlainTextForToolTip(I18n.Text("Specifies the layout of the various blocks of data on the character sheet"));
        mBlockLayoutField = new JTextArea(Preferences.linesToString(mSettings.blockLayout()));
        mBlockLayoutField.setToolTipText(blockLayoutTooltip);
        mBlockLayoutField.setBorder(new EmptyBorder(0, 4, 0, 4));
        mBlockLayoutField.getDocument().addDocumentListener(this);
        JScrollPane scroller = new JScrollPane(mBlockLayoutField, ScrollPaneConstants.VERTICAL_SCROLLBAR_AS_NEEDED, ScrollPaneConstants.HORIZONTAL_SCROLLBAR_AS_NEEDED);
        panel.add(new Label(I18n.Text("Block Layout")), new PrecisionLayoutData());
        panel.add(scroller, new PrecisionLayoutData().setHeightHint(scroller.getPreferredSize().height).setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        getContentPane().add(panel, BorderLayout.CENTER);
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
        atDefaults = atDefaults && mSettings.getAttributes().equals(Preferences.getInstance().getAttributes());
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
            reset();
        }
        adjustResetButton();
    }

    private void reset() {
        Preferences prefs = Preferences.getInstance();
        mUseModifyingDicePlusAdds.setSelected(prefs.useModifyingDicePlusAdds());
        mShowCollegeInSpells.setSelected(prefs.showCollegeInSheetSpells());
        mShowDifficulty.setSelected(prefs.showDifficulty());
        mShowAdvantageModifierAdj.setSelected(prefs.showAdvantageModifierAdj());
        mShowEquipmentModifierAdj.setSelected(prefs.showEquipmentModifierAdj());
        mShowSpellAdj.setSelected(prefs.showSpellAdj());
        mShowTitleInsteadOfNameInPageFooter.setSelected(prefs.useTitleInFooter());
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
        mAttributeEditor.reset(AttributeDef.cloneMap(prefs.getAttributes()));
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
        mCharacter.removeChangeListener(this);
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
