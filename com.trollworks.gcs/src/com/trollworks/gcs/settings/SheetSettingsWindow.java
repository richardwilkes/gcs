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

package com.trollworks.gcs.settings;

import com.trollworks.gcs.character.DisplayOption;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataChangeListener;
import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.page.PageSettings;
import com.trollworks.gcs.page.PageSettingsEditor;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.Button;
import com.trollworks.gcs.ui.widget.Checkbox;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.WeightUnits;

import java.awt.BorderLayout;
import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.FlowLayout;
import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.WindowEvent;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import javax.swing.JComboBox;
import javax.swing.JTextArea;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

// TODO: Fix layout around scrolling

public final class SheetSettingsWindow extends BaseWindow implements ActionListener, DocumentListener, CloseHandler, DataChangeListener, PageSettingsEditor.ResetPageSettings {
    private static final Map<UUID, SheetSettingsWindow> INSTANCES = new HashMap<>();
    private              GURPSCharacter                 mCharacter;
    private              SheetSettings                  mSheetSettings;
    private              Checkbox                       mUseMultiplicativeModifiers;
    private              Checkbox                       mUseModifyingDicePlusAdds;
    private              Checkbox                       mUseKnowYourOwnStrength;
    private              Checkbox                       mUseReducedSwing;
    private              Checkbox                       mUseThrustEqualsSwingMinus2;
    private              Checkbox                       mUseSimpleMetricConversions;
    private              Checkbox                       mShowCollegeInSpells;
    private              Checkbox                       mShowDifficulty;
    private              Checkbox                       mShowAdvantageModifierAdj;
    private              Checkbox                       mShowEquipmentModifierAdj;
    private              Checkbox                       mShowSpellAdj;
    private              Checkbox                       mShowTitleInsteadOfNameInPageFooter;
    private              JComboBox<LengthUnits>         mLengthUnitsCombo;
    private              JComboBox<WeightUnits>         mWeightUnitsCombo;
    private              JComboBox<DisplayOption>       mUserDescriptionDisplayCombo;
    private              JComboBox<DisplayOption>       mModifiersDisplayCombo;
    private              JComboBox<DisplayOption>       mNotesDisplayCombo;
    private              JTextArea                      mBlockLayoutField;
    private              PageSettingsEditor             mPageSettingsEditor;
    private              Button                         mResetButton;
    private              boolean                        mUpdatePending;

    /** Displays the sheet settings window. */
    public static void display(GURPSCharacter gchar) {
        if (!UIUtilities.inModalState()) {
            SheetSettingsWindow wnd;
            synchronized (INSTANCES) {
                UUID key = gchar == null ? null : gchar.getID();
                wnd = INSTANCES.get(key);
                if (wnd == null) {
                    wnd = new SheetSettingsWindow(gchar);
                    INSTANCES.put(key, wnd);
                }
            }
            wnd.setVisible(true);
        }
    }

    /** Closes the SheetSettingsWindow for the given character if it is open. */
    public static void closeFor(GURPSCharacter gchar) {
        for (Window window : Window.getWindows()) {
            if (window.isShowing() && window instanceof SheetSettingsWindow) {
                SheetSettingsWindow wnd = (SheetSettingsWindow) window;
                if (wnd.mCharacter == gchar) {
                    wnd.attemptClose();
                }
            }
        }
    }

    private static String createTitle(GURPSCharacter character) {
        return character == null ? I18n.text("Sheet Settings") : String.format(I18n.text("Sheet Settings: %s"), character.getProfile().getName());
    }

    private SheetSettingsWindow(GURPSCharacter character) {
        super(createTitle(character));
        mCharacter = character;
        mSheetSettings = SheetSettings.get(mCharacter);
        addTopPanel();
        addResetPanel();
        adjustResetButton();
        establishSizing();
        WindowUtils.packAndCenterWindowOn(this, null);
        if (mCharacter != null) {
            mCharacter.addChangeListener(this);
            Settings.getInstance().addChangeListener(this);
        }
    }

    @Override
    public void establishSizing() {
        pack();
        int width = getSize().width;
        setMinimumSize(new Dimension(width, 200));
        setMaximumSize(new Dimension(width, 10000));
    }

    private void addTopPanel() {
        Panel left = new Panel(new PrecisionLayout().setColumns(2).setMargins(0));
        mShowCollegeInSpells = addCheckbox(left, I18n.text("Show the College column"), null,
                mSheetSettings.showCollegeInSpells(), (b) -> {
                    mSheetSettings.setShowCollegeInSpells(b.isChecked());
                    adjustResetButton();
                });
        mShowDifficulty = addCheckbox(left, I18n.text("Show the Difficulty column"), null,
                mSheetSettings.showDifficulty(), (b) -> {
                    mSheetSettings.setShowDifficulty(b.isChecked());
                    adjustResetButton();
                });
        mShowAdvantageModifierAdj = addCheckbox(left,
                I18n.text("Show advantage modifier cost adjustments"), null,
                mSheetSettings.showAdvantageModifierAdj(), (b) -> {
                    mSheetSettings.setShowAdvantageModifierAdj(b.isChecked());
                    adjustResetButton();
                });
        mShowEquipmentModifierAdj = addCheckbox(left,
                I18n.text("Show equipment modifier cost & weight adjustments"), null,
                mSheetSettings.showEquipmentModifierAdj(), (b) -> {
                    mSheetSettings.setShowEquipmentModifierAdj(b.isChecked());
                    adjustResetButton();
                });
        mShowSpellAdj = addCheckbox(left, I18n.text("Show spell ritual, cost & time adjustments"),
                null, mSheetSettings.showSpellAdj(), (b) -> {
                    mSheetSettings.setShowSpellAdj(b.isChecked());
                    adjustResetButton();
                });
        mShowTitleInsteadOfNameInPageFooter = addCheckbox(left,
                I18n.text("Show the title instead of the name in the footer"), null,
                mSheetSettings.useTitleInFooter(), (b) -> {
                    mSheetSettings.setUseTitleInFooter(b.isChecked());
                    adjustResetButton();
                });
        String tooltip = I18n.text("Where to display this information");
        mUserDescriptionDisplayCombo = addCombo(left, DisplayOption.values(), mSheetSettings.userDescriptionDisplay(), I18n.text("Show User Description"), tooltip);
        mModifiersDisplayCombo = addCombo(left, DisplayOption.values(), mSheetSettings.modifiersDisplay(), I18n.text("Show Modifiers"), tooltip);
        mNotesDisplayCombo = addCombo(left, DisplayOption.values(), mSheetSettings.notesDisplay(), I18n.text("Show Notes"), tooltip);
        String blockLayoutTooltip = I18n.text("Specifies the layout of the various blocks of data on the character sheet");
        mBlockLayoutField = new MultiLineTextField(Settings.linesToString(mSheetSettings.blockLayout()), blockLayoutTooltip, this);
        left.add(new Label(I18n.text("Block Layout")), new PrecisionLayoutData().setHorizontalSpan(2));
        left.add(mBlockLayoutField, new PrecisionLayoutData().setFillAlignment().setGrabSpace(true).setHorizontalSpan(2));

        Panel right = new Panel(new PrecisionLayout().setColumns(2).setMargins(0));
        mUseMultiplicativeModifiers = addCheckbox(right,
                I18n.text("Use Multiplicative Modifiers (PW102; changes point value)"), null,
                mSheetSettings.useMultiplicativeModifiers(), (b) -> {
                    mSheetSettings.setUseMultiplicativeModifiers(b.isChecked());
                    adjustResetButton();
                });
        mUseModifyingDicePlusAdds = addCheckbox(right, I18n.text("Use Modifying Dice + Adds (B269)"),
                null, mSheetSettings.useModifyingDicePlusAdds(), (b) -> {
                    mSheetSettings.setUseModifyingDicePlusAdds(b.isChecked());
                    adjustResetButton();
                });
        mUseKnowYourOwnStrength = addCheckbox(right,
                I18n.text("Use strength rules from Knowing Your Own Strength (PY83)"), null,
                mSheetSettings.useKnowYourOwnStrength(), (b) -> {
                    mSheetSettings.setUseKnowYourOwnStrength(b.isChecked());
                    adjustResetButton();
                });
        mUseReducedSwing = addCheckbox(right, I18n.text("Use the reduced swing rules"),
                I18n.text("From \"Adjusting Swing Damage in Dungeon Fantasy\" found on noschoolgrognard.blogspot.com"),
                mSheetSettings.useReducedSwing(), (b) -> {
                    mSheetSettings.setUseReducedSwing(b.isChecked());
                    adjustResetButton();
                });
        mUseThrustEqualsSwingMinus2 = addCheckbox(right, I18n.text("Use Thrust = Swing - 2"), null,
                mSheetSettings.useThrustEqualsSwingMinus2(), (b) -> {
                    mSheetSettings.setUseThrustEqualsSwingMinus2(b.isChecked());
                    adjustResetButton();
                });
        mUseSimpleMetricConversions = addCheckbox(right,
                I18n.text("Use the simple metric conversion rules (B9)"), null,
                mSheetSettings.useSimpleMetricConversions(), (b) -> {
                    mSheetSettings.setUseSimpleMetricConversions(b.isChecked());
                    adjustResetButton();
                });
        mLengthUnitsCombo = addCombo(right, LengthUnits.values(), mSheetSettings.defaultLengthUnits(), I18n.text("Length Units"), I18n.text("The units to use for display of generated lengths"));
        mWeightUnitsCombo = addCombo(right, WeightUnits.values(), mSheetSettings.defaultWeightUnits(), I18n.text("Weight Units"), I18n.text("The units to use for display of generated weights"));
        mPageSettingsEditor = new PageSettingsEditor(mSheetSettings.getPageSettings(), this::adjustResetButton, this);
        right.add(mPageSettingsEditor, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment().setHorizontalSpan(2).setTopMargin(10));

        Panel panel = new Panel(new PrecisionLayout().setColumns(2).setMargins(10).setHorizontalSpacing(10));
        panel.add(left, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        panel.add(right, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));

        panel.setSize(panel.getPreferredSize());

        getContentPane().add(new ScrollPanel(panel), BorderLayout.CENTER);
    }

    private void addResetPanel() {
        Panel panel = new Panel(new FlowLayout(FlowLayout.CENTER));
        mResetButton = new Button(mCharacter == null ? I18n.text("Reset to Factory Settings") : I18n.text("Reset to Defaults"), (btn) -> reset());
        panel.add(mResetButton);
        getContentPane().add(panel, BorderLayout.SOUTH);
    }

    private static void addLabel(Panel panel, String title) {
        panel.add(new Label(title, SwingConstants.RIGHT), new PrecisionLayoutData().setFillHorizontalAlignment());
    }

    private <E> JComboBox<E> addCombo(Panel panel, E[] values, E choice, String title, String tooltip) {
        JComboBox<E> combo = new JComboBox<>(values);
        combo.setOpaque(false);
        combo.setToolTipText(tooltip);
        combo.setSelectedItem(choice);
        combo.addActionListener(this);
        combo.setMaximumRowCount(combo.getItemCount());
        panel.add(new Label(title, combo), new PrecisionLayoutData().setFillHorizontalAlignment());
        panel.add(combo);
        return combo;
    }

    private static Checkbox addCheckbox(Panel panel, String title, String tooltip, boolean checked, Checkbox.ClickFunction clickFunction) {
        Checkbox checkbox = new Checkbox(title, checked, clickFunction);
        checkbox.setToolTipText(tooltip);
        panel.add(checkbox, new PrecisionLayoutData().setHorizontalSpan(2));
        return checkbox;
    }

    private void adjustResetButton() {
        mResetButton.setEnabled(!isSetToDefaults());
    }

    private boolean isSetToDefaults() {
        SheetSettings defaults   = new SheetSettings(mCharacter);
        boolean       atDefaults = mUseModifyingDicePlusAdds.isChecked() == defaults.useModifyingDicePlusAdds();
        atDefaults = atDefaults && mShowCollegeInSpells.isChecked() == defaults.showCollegeInSpells();
        atDefaults = atDefaults && mShowDifficulty.isChecked() == defaults.showDifficulty();
        atDefaults = atDefaults && mShowAdvantageModifierAdj.isChecked() == defaults.showAdvantageModifierAdj();
        atDefaults = atDefaults && mShowEquipmentModifierAdj.isChecked() == defaults.showEquipmentModifierAdj();
        atDefaults = atDefaults && mShowSpellAdj.isChecked() == defaults.showSpellAdj();
        atDefaults = atDefaults && mShowTitleInsteadOfNameInPageFooter.isChecked() == defaults.useTitleInFooter();
        atDefaults = atDefaults && mUseMultiplicativeModifiers.isChecked() == defaults.useMultiplicativeModifiers();
        atDefaults = atDefaults && mUseKnowYourOwnStrength.isChecked() == defaults.useKnowYourOwnStrength();
        atDefaults = atDefaults && mUseThrustEqualsSwingMinus2.isChecked() == defaults.useThrustEqualsSwingMinus2();
        atDefaults = atDefaults && mUseReducedSwing.isChecked() == defaults.useReducedSwing();
        atDefaults = atDefaults && mUseSimpleMetricConversions.isChecked() == defaults.useSimpleMetricConversions();
        atDefaults = atDefaults && mLengthUnitsCombo.getSelectedItem() == defaults.defaultLengthUnits();
        atDefaults = atDefaults && mWeightUnitsCombo.getSelectedItem() == defaults.defaultWeightUnits();
        atDefaults = atDefaults && mUserDescriptionDisplayCombo.getSelectedItem() == defaults.userDescriptionDisplay();
        atDefaults = atDefaults && mModifiersDisplayCombo.getSelectedItem() == defaults.modifiersDisplay();
        atDefaults = atDefaults && mNotesDisplayCombo.getSelectedItem() == defaults.notesDisplay();
        atDefaults = atDefaults && mBlockLayoutField.getText().equals(Settings.linesToString(defaults.blockLayout()));
        atDefaults = atDefaults && mSheetSettings.getPageSettings().equals(defaults.getPageSettings());
        return atDefaults;
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Object source = event.getSource();
        if (source == mLengthUnitsCombo) {
            mSheetSettings.setDefaultLengthUnits((LengthUnits) mLengthUnitsCombo.getSelectedItem());
        } else if (source == mWeightUnitsCombo) {
            mSheetSettings.setDefaultWeightUnits((WeightUnits) mWeightUnitsCombo.getSelectedItem());
        } else if (source == mUserDescriptionDisplayCombo) {
            mSheetSettings.setUserDescriptionDisplay((DisplayOption) mUserDescriptionDisplayCombo.getSelectedItem());
        } else if (source == mModifiersDisplayCombo) {
            mSheetSettings.setModifiersDisplay((DisplayOption) mModifiersDisplayCombo.getSelectedItem());
        } else if (source == mNotesDisplayCombo) {
            mSheetSettings.setNotesDisplay((DisplayOption) mNotesDisplayCombo.getSelectedItem());
        }
        adjustResetButton();
    }

    private void reset() {
        SheetSettings defaults = new SheetSettings(mCharacter);
        mUseModifyingDicePlusAdds.setChecked(defaults.useModifyingDicePlusAdds());
        mShowCollegeInSpells.setChecked(defaults.showCollegeInSpells());
        mShowDifficulty.setChecked(defaults.showDifficulty());
        mShowAdvantageModifierAdj.setChecked(defaults.showAdvantageModifierAdj());
        mShowEquipmentModifierAdj.setChecked(defaults.showEquipmentModifierAdj());
        mShowSpellAdj.setChecked(defaults.showSpellAdj());
        mShowTitleInsteadOfNameInPageFooter.setChecked(defaults.useTitleInFooter());
        mUseMultiplicativeModifiers.setChecked(defaults.useMultiplicativeModifiers());
        mUseKnowYourOwnStrength.setChecked(defaults.useKnowYourOwnStrength());
        mUseThrustEqualsSwingMinus2.setChecked(defaults.useThrustEqualsSwingMinus2());
        mUseReducedSwing.setChecked(defaults.useReducedSwing());
        mUseSimpleMetricConversions.setChecked(defaults.useSimpleMetricConversions());
        mLengthUnitsCombo.setSelectedItem(defaults.defaultLengthUnits());
        mWeightUnitsCombo.setSelectedItem(defaults.defaultWeightUnits());
        mUserDescriptionDisplayCombo.setSelectedItem(defaults.userDescriptionDisplay());
        mModifiersDisplayCombo.setSelectedItem(defaults.modifiersDisplay());
        mNotesDisplayCombo.setSelectedItem(defaults.notesDisplay());
        mBlockLayoutField.setText(Settings.linesToString(defaults.blockLayout()));
        mPageSettingsEditor.reset();
        adjustResetButton();
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
        synchronized (INSTANCES) {
            INSTANCES.remove(mCharacter == null ? null : mCharacter.getID());
        }
        if (mCharacter != null) {
            mCharacter.removeChangeListener(this);
            Settings.getInstance().removeChangeListener(this);
        }
        super.dispose();
    }

    @Override
    public void dataWasChanged() {
        if (!mUpdatePending) {
            mUpdatePending = true;
            EventQueue.invokeLater(() -> {
                setTitle(createTitle(mCharacter));
                adjustResetButton();
                mUpdatePending = false;
            });
        }
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
        mSheetSettings.setBlockLayout(List.of(mBlockLayoutField.getText().split("\n")));
        adjustResetButton();
    }

    @Override
    public void resetPageSettings(PageSettings settings) {
        settings.copy(new SheetSettings(mCharacter).getPageSettings());
    }
}
