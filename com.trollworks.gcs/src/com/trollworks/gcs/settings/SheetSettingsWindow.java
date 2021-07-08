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
import com.trollworks.gcs.page.PageSettingsEditor;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.Checkbox;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.LayoutConstants;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.Separator;
import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.WeightUnits;

import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Window;
import java.io.IOException;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

public final class SheetSettingsWindow extends SettingsWindow<SheetSettings> implements DocumentListener, DataChangeListener {
    private static final Map<UUID, SheetSettingsWindow> INSTANCES = new HashMap<>();

    private GURPSCharacter               mCharacter;
    private SheetSettings                mSheetSettings;
    private Checkbox                     mUseMultiplicativeModifiers;
    private Checkbox                     mUseModifyingDicePlusAdds;
    private Checkbox                     mUseSimpleMetricConversions;
    private Checkbox                     mShowCollegeInSpells;
    private Checkbox                     mShowDifficulty;
    private Checkbox                     mShowAdvantageModifierAdj;
    private Checkbox                     mShowEquipmentModifierAdj;
    private Checkbox                     mShowSpellAdj;
    private Checkbox                     mShowTitleInsteadOfNameInPageFooter;
    private PopupMenu<DamageProgression> mDamageProgressionPopup;
    private PopupMenu<LengthUnits>       mLengthUnitsPopup;
    private PopupMenu<WeightUnits>       mWeightUnitsPopup;
    private PopupMenu<DisplayOption>     mUserDescriptionDisplayPopup;
    private PopupMenu<DisplayOption>     mModifiersDisplayPopup;
    private PopupMenu<DisplayOption>     mNotesDisplayPopup;
    private MultiLineTextField           mBlockLayoutField;
    private PageSettingsEditor           mPageSettingsEditor;
    private boolean                      mUpdatePending;

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
        if (mCharacter != null) {
            mCharacter.addChangeListener(this);
            Settings.getInstance().addChangeListener(this);
        }
        fill();
    }

    @Override
    protected void preDispose() {
        synchronized (INSTANCES) {
            INSTANCES.remove(mCharacter == null ? null : mCharacter.getID());
        }
        if (mCharacter != null) {
            mCharacter.removeChangeListener(this);
            Settings.getInstance().removeChangeListener(this);
        }
    }

    @Override
    protected Panel createContent() {
        Panel left = new Panel(new PrecisionLayout().setMargins(0).
                setVerticalSpacing(LayoutConstants.WINDOW_BORDER_INSET));
        left.add(createDamageProgressionPanel());
        left.add(createCheckBoxGroupPanel());
        left.add(createBlockLayoutPanel(), new PrecisionLayoutData().setFillAlignment().setGrabSpace(true));

        Panel right = new Panel(new PrecisionLayout().setMargins(0).
                setVerticalSpacing(LayoutConstants.WINDOW_BORDER_INSET));
        right.add(createUnitsOfMeasurePanel(), new PrecisionLayoutData().
                setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        right.add(createWhereToDisplayPanel(), new PrecisionLayoutData().
                setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        mPageSettingsEditor = new PageSettingsEditor(mSheetSettings.getPageSettings(), this::adjustResetButton);
        right.add(mPageSettingsEditor, new PrecisionLayoutData().setFillHorizontalAlignment().
                setGrabHorizontalSpace(true));

        Panel panel = new Panel(new PrecisionLayout().setColumns(2).
                setMargins(LayoutConstants.WINDOW_BORDER_INSET).
                setHorizontalSpacing(LayoutConstants.WINDOW_BORDER_INSET));
        panel.add(left, new PrecisionLayoutData().setFillVerticalAlignment().setGrabVerticalSpace(true));
        panel.add(right, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        panel.setSize(panel.getPreferredSize());
        return panel;
    }

    private static <E> PopupMenu<E> addPopupMenu(Panel panel, E[] values, E choice, String title, String tooltip, PopupMenu.SelectionListener<E> listener) {
        PopupMenu<E> popup = new PopupMenu<>(values, listener);
        popup.setToolTipText(tooltip);
        popup.setSelectedItem(choice, false);
        panel.add(new Label(title), new PrecisionLayoutData().setFillHorizontalAlignment());
        panel.add(popup);
        return popup;
    }

    private static Checkbox addCheckbox(Panel panel, String title, String tooltip, boolean checked, Checkbox.ClickFunction clickFunction) {
        Checkbox checkbox = new Checkbox(title, checked, clickFunction);
        checkbox.setToolTipText(tooltip);
        panel.add(checkbox, new PrecisionLayoutData().setHorizontalSpan(2));
        return checkbox;
    }

    private Panel createDamageProgressionPanel() {
        Panel             panel                    = new Panel(new PrecisionLayout().setColumns(2).setMargins(0));
        DamageProgression currentDamageProgression = mSheetSettings.getDamageProgression();
        mDamageProgressionPopup = addPopupMenu(panel, DamageProgression.values(),
                currentDamageProgression, I18n.text("Damage Progression"),
                currentDamageProgression.getTooltip(), (p) -> {
                    DamageProgression progression = mDamageProgressionPopup.getSelectedItem();
                    if (progression != null) {
                        mDamageProgressionPopup.setToolTipText(progression.getTooltip());
                        mSheetSettings.setDamageProgression(progression);
                        adjustResetButton();
                    }
                });
        return panel;
    }

    private Panel createCheckBoxGroupPanel() {
        Panel panel = new Panel(new PrecisionLayout().setMargins(0));
        mShowCollegeInSpells = addCheckbox(panel, I18n.text("Show the College column"), null,
                mSheetSettings.showCollegeInSpells(), (b) -> {
            System.out.println(b + ": " + b.isChecked());
                    mSheetSettings.setShowCollegeInSpells(b.isChecked());
                    adjustResetButton();
                });
        mShowDifficulty = addCheckbox(panel, I18n.text("Show the Difficulty column"), null,
                mSheetSettings.showDifficulty(), (b) -> {
                    mSheetSettings.setShowDifficulty(b.isChecked());
                    adjustResetButton();
                });
        mShowAdvantageModifierAdj = addCheckbox(panel,
                I18n.text("Show advantage modifier cost adjustments"), null,
                mSheetSettings.showAdvantageModifierAdj(), (b) -> {
                    mSheetSettings.setShowAdvantageModifierAdj(b.isChecked());
                    adjustResetButton();
                });
        mShowEquipmentModifierAdj = addCheckbox(panel,
                I18n.text("Show equipment modifier cost & weight adjustments"), null,
                mSheetSettings.showEquipmentModifierAdj(), (b) -> {
                    mSheetSettings.setShowEquipmentModifierAdj(b.isChecked());
                    adjustResetButton();
                });
        mShowSpellAdj = addCheckbox(panel, I18n.text("Show spell ritual, cost & time adjustments"),
                null, mSheetSettings.showSpellAdj(), (b) -> {
                    mSheetSettings.setShowSpellAdj(b.isChecked());
                    adjustResetButton();
                });
        mShowTitleInsteadOfNameInPageFooter = addCheckbox(panel,
                I18n.text("Show the title instead of the name in the footer"), null,
                mSheetSettings.useTitleInFooter(), (b) -> {
                    mSheetSettings.setUseTitleInFooter(b.isChecked());
                    adjustResetButton();
                });
        mUseMultiplicativeModifiers = addCheckbox(panel,
                I18n.text("Use Multiplicative Modifiers (PW102; changes point value)"), null,
                mSheetSettings.useMultiplicativeModifiers(), (b) -> {
                    mSheetSettings.setUseMultiplicativeModifiers(b.isChecked());
                    adjustResetButton();
                });
        mUseModifyingDicePlusAdds = addCheckbox(panel, I18n.text("Use Modifying Dice + Adds (B269)"),
                null, mSheetSettings.useModifyingDicePlusAdds(), (b) -> {
                    mSheetSettings.setUseModifyingDicePlusAdds(b.isChecked());
                    adjustResetButton();
                });
        return panel;
    }

    private Panel createBlockLayoutPanel() {
        Panel panel  = new Panel(new PrecisionLayout().setMargins(0));
        Label header = new Label(I18n.text("Block Layout"));
        header.setThemeFont(Fonts.HEADER);
        panel.add(header);
        mBlockLayoutField = new MultiLineTextField(Settings.linesToString(mSheetSettings.blockLayout()),
                I18n.text("Specifies the layout of the various blocks of data on the character sheet"), this);
        panel.add(mBlockLayoutField, new PrecisionLayoutData().setFillAlignment().setGrabSpace(true));
        return panel;
    }

    private Panel createUnitsOfMeasurePanel() {
        Panel panel  = new Panel(new PrecisionLayout().setColumns(2).setMargins(0));
        Label header = new Label(I18n.text("Units of Measurement"));
        header.setThemeFont(Fonts.HEADER);
        panel.add(header, new PrecisionLayoutData().setHorizontalSpan(2));
        panel.add(new Separator(), new PrecisionLayoutData().setFillHorizontalAlignment().
                setGrabHorizontalSpace(true).setHorizontalSpan(2).
                setBottomMargin(LayoutConstants.TOOLBAR_VERTICAL_INSET / 2));
        mLengthUnitsPopup = addPopupMenu(panel, LengthUnits.values(),
                mSheetSettings.defaultLengthUnits(), I18n.text("Length Units"),
                I18n.text("The units to use for display of generated lengths"), (p) -> {
                    mSheetSettings.setDefaultLengthUnits(mLengthUnitsPopup.getSelectedItem());
                    adjustResetButton();
                });
        mWeightUnitsPopup = addPopupMenu(panel, WeightUnits.values(),
                mSheetSettings.defaultWeightUnits(), I18n.text("Weight Units"),
                I18n.text("The units to use for display of generated weights"), (p) -> {
                    mSheetSettings.setDefaultWeightUnits(mWeightUnitsPopup.getSelectedItem());
                    adjustResetButton();
                });
        mUseSimpleMetricConversions = addCheckbox(panel,
                I18n.text("Use the simple metric conversion rules (B9)"), null,
                mSheetSettings.useSimpleMetricConversions(), (b) -> {
                    mSheetSettings.setUseSimpleMetricConversions(b.isChecked());
                    adjustResetButton();
                });
        return panel;
    }

    private Panel createWhereToDisplayPanel() {
        Panel panel  = new Panel(new PrecisionLayout().setColumns(2).setMargins(0));
        Label header = new Label(I18n.text("Where to display…"));
        header.setThemeFont(Fonts.HEADER);
        panel.add(header, new PrecisionLayoutData().setHorizontalSpan(2));
        panel.add(new Separator(), new PrecisionLayoutData().setFillHorizontalAlignment().
                setGrabHorizontalSpace(true).setHorizontalSpan(2).
                setBottomMargin(LayoutConstants.TOOLBAR_VERTICAL_INSET / 2));
        String tooltip = I18n.text("Where to display this information");
        mUserDescriptionDisplayPopup = addPopupMenu(panel, DisplayOption.values(),
                mSheetSettings.userDescriptionDisplay(), I18n.text("User Description"),
                tooltip, (p) -> {
                    mSheetSettings.setUserDescriptionDisplay(mUserDescriptionDisplayPopup.getSelectedItem());
                    adjustResetButton();
                });
        mModifiersDisplayPopup = addPopupMenu(panel, DisplayOption.values(),
                mSheetSettings.modifiersDisplay(), I18n.text("Modifiers"), tooltip, (p) -> {
                    mSheetSettings.setModifiersDisplay(mModifiersDisplayPopup.getSelectedItem());
                    adjustResetButton();
                });
        mNotesDisplayPopup = addPopupMenu(panel, DisplayOption.values(), mSheetSettings.notesDisplay(),
                I18n.text("Notes"), tooltip, (p) -> {
                    mSheetSettings.setNotesDisplay(mNotesDisplayPopup.getSelectedItem());
                    adjustResetButton();
                });
        return panel;
    }

    @Override
    public void establishSizing() {
        pack();
        int width = getSize().width;
        setMinimumSize(new Dimension(width, 200));
        setMaximumSize(new Dimension(width, 10000));
    }

    @Override
    protected boolean shouldResetBeEnabled() {
        SheetSettings defaults   = new SheetSettings(mCharacter);
        boolean       atDefaults = mUseModifyingDicePlusAdds.isChecked() == defaults.useModifyingDicePlusAdds();
        atDefaults = atDefaults && mShowCollegeInSpells.isChecked() == defaults.showCollegeInSpells();
        atDefaults = atDefaults && mShowDifficulty.isChecked() == defaults.showDifficulty();
        atDefaults = atDefaults && mShowAdvantageModifierAdj.isChecked() == defaults.showAdvantageModifierAdj();
        atDefaults = atDefaults && mShowEquipmentModifierAdj.isChecked() == defaults.showEquipmentModifierAdj();
        atDefaults = atDefaults && mShowSpellAdj.isChecked() == defaults.showSpellAdj();
        atDefaults = atDefaults && mShowTitleInsteadOfNameInPageFooter.isChecked() == defaults.useTitleInFooter();
        atDefaults = atDefaults && mUseMultiplicativeModifiers.isChecked() == defaults.useMultiplicativeModifiers();
        atDefaults = atDefaults && mDamageProgressionPopup.getSelectedItem() == defaults.getDamageProgression();
        atDefaults = atDefaults && mUseSimpleMetricConversions.isChecked() == defaults.useSimpleMetricConversions();
        atDefaults = atDefaults && mLengthUnitsPopup.getSelectedItem() == defaults.defaultLengthUnits();
        atDefaults = atDefaults && mWeightUnitsPopup.getSelectedItem() == defaults.defaultWeightUnits();
        atDefaults = atDefaults && mUserDescriptionDisplayPopup.getSelectedItem() == defaults.userDescriptionDisplay();
        atDefaults = atDefaults && mModifiersDisplayPopup.getSelectedItem() == defaults.modifiersDisplay();
        atDefaults = atDefaults && mNotesDisplayPopup.getSelectedItem() == defaults.notesDisplay();
        atDefaults = atDefaults && mBlockLayoutField.getText().equals(Settings.linesToString(defaults.blockLayout()));
        atDefaults = atDefaults && mSheetSettings.getPageSettings().equals(defaults.getPageSettings());
        return !atDefaults;
    }

    @Override
    protected String getResetButtonTooltip() {
        return mCharacter == null ? super.getResetButtonTooltip() : I18n.text("Reset to Global Defaults");
    }

    @Override
    protected SheetSettings getResetData() {
        return new SheetSettings(mCharacter);
    }

    @Override
    protected void doResetTo(SheetSettings data) {
        mSheetSettings.copyFrom(data);
        mUseModifyingDicePlusAdds.setChecked(data.useModifyingDicePlusAdds());
        mShowCollegeInSpells.setChecked(data.showCollegeInSpells());
        mShowDifficulty.setChecked(data.showDifficulty());
        mShowAdvantageModifierAdj.setChecked(data.showAdvantageModifierAdj());
        mShowEquipmentModifierAdj.setChecked(data.showEquipmentModifierAdj());
        mShowSpellAdj.setChecked(data.showSpellAdj());
        mShowTitleInsteadOfNameInPageFooter.setChecked(data.useTitleInFooter());
        mUseMultiplicativeModifiers.setChecked(data.useMultiplicativeModifiers());
        DamageProgression progression = data.getDamageProgression();
        mDamageProgressionPopup.setSelectedItem(progression, true);
        mUseSimpleMetricConversions.setChecked(data.useSimpleMetricConversions());
        mLengthUnitsPopup.setSelectedItem(data.defaultLengthUnits(), true);
        mWeightUnitsPopup.setSelectedItem(data.defaultWeightUnits(), true);
        mUserDescriptionDisplayPopup.setSelectedItem(data.userDescriptionDisplay(), true);
        mModifiersDisplayPopup.setSelectedItem(data.modifiersDisplay(), true);
        mNotesDisplayPopup.setSelectedItem(data.notesDisplay(), true);
        mBlockLayoutField.setText(Settings.linesToString(data.blockLayout()));
        mPageSettingsEditor.resetTo(data.getPageSettings());
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
    protected Dirs getDir() {
        return Dirs.SETTINGS;
    }

    @Override
    protected FileType getFileType() {
        return FileType.SHEET_SETTINGS;
    }

    @Override
    protected SheetSettings createSettingsFrom(Path path) throws IOException {
        return new SheetSettings(path);
    }

    @Override
    protected void exportSettingsTo(Path path) throws IOException {
        mSheetSettings.save(path);
    }
}
