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

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.menu.file.ExportToGURPSCalculatorCommand;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.scale.Scales;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.Button;
import com.trollworks.gcs.ui.widget.Checkbox;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.I18n;

import java.awt.Container;
import java.awt.Desktop;
import java.awt.Dimension;
import java.awt.event.WindowEvent;
import java.net.URI;
import java.text.MessageFormat;
import javax.swing.SwingConstants;

public final class GeneralSettingsWindow extends BaseWindow implements CloseHandler {
    private static GeneralSettingsWindow INSTANCE;
    private        EditorField           mPlayerName;
    private        EditorField           mTechLevel;
    private        EditorField           mInitialPoints;
    private        Checkbox              mAutoFillProfile;
    private        PopupMenu<Scales>     mInitialScale;
    private        EditorField           mToolTipTimeout;
    private        EditorField           mImageResolution;
    private        Checkbox              mIncludeUnspentPointsInTotal;
    private        EditorField           mGCalcKey;
    private        Button                mResetButton;

    /** Displays the general settings window. */
    public static void display() {
        if (!UIUtilities.inModalState()) {
            GeneralSettingsWindow wnd;
            synchronized (GeneralSettingsWindow.class) {
                if (INSTANCE == null) {
                    INSTANCE = new GeneralSettingsWindow();
                }
                wnd = INSTANCE;
            }
            wnd.setVisible(true);
        }
    }

    private GeneralSettingsWindow() {
        super(I18n.text("General Settings"));
        Settings  prefs   = Settings.getInstance();
        Container content = getContentPane();
        content.setLayout(new PrecisionLayout().setColumns(3).setMargins(10));

        // First row
        mPlayerName = new EditorField(FieldFactory.STRING, (f) -> {
            Settings.getInstance().setDefaultPlayerName(f.getText().trim());
            adjustResetButton();
        }, SwingConstants.LEFT, prefs.getDefaultPlayerName(),
                I18n.text("The player name to use when a new character sheet is created"));
        content.add(new Label(I18n.text("Player"), mPlayerName), new PrecisionLayoutData().setFillHorizontalAlignment());
        content.add(mPlayerName, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        mAutoFillProfile = new Checkbox(I18n.text("Fill in initial description"),
                prefs.autoFillProfile(), (b) -> {
            prefs.setAutoFillProfile(b.isChecked());
            adjustResetButton();
        });
        mAutoFillProfile.setToolTipText(I18n.text("Automatically fill in new character identity and description information with randomized choices"));
        mAutoFillProfile.setOpaque(false);
        content.add(mAutoFillProfile);

        // Second row
        mTechLevel = new EditorField(FieldFactory.STRING, (f) -> {
            Settings.getInstance().setDefaultTechLevel(f.getText().trim());
            adjustResetButton();
        }, SwingConstants.RIGHT, prefs.getDefaultTechLevel(), "99+99^", getTechLevelTooltip());
        content.add(new Label(I18n.text("Tech Level"), mTechLevel),
                new PrecisionLayoutData().setFillHorizontalAlignment());
        Wrapper wrapper = new Wrapper(new PrecisionLayout().setMargins(0).setColumns(3));
        content.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        wrapper.add(mTechLevel, new PrecisionLayoutData().setFillHorizontalAlignment());

        mInitialPoints = new EditorField(FieldFactory.POSINT6, (f) -> {
            Settings.getInstance().setInitialPoints(((Integer) f.getValue()).intValue());
            adjustResetButton();
        }, SwingConstants.RIGHT, Integer.valueOf(prefs.getInitialPoints()), Integer.valueOf(999999),
                I18n.text("The initial number of character points to start with"));
        wrapper.add(new Label(I18n.text("Initial Points"), mTechLevel),
                new PrecisionLayoutData().setFillHorizontalAlignment().setLeftMargin(5));
        wrapper.add(mInitialPoints, new PrecisionLayoutData().setFillHorizontalAlignment());

        mIncludeUnspentPointsInTotal = new Checkbox(I18n.text("Include unspent points in total"),
                prefs.includeUnspentPointsInTotal(), (b) -> {
            prefs.setIncludeUnspentPointsInTotal(b.isChecked());
            adjustResetButton();
        });
        mIncludeUnspentPointsInTotal.setToolTipText(I18n.text("Include unspent points in the character point total"));
        mIncludeUnspentPointsInTotal.setOpaque(false);
        content.add(mIncludeUnspentPointsInTotal);

        // Third row
        mInitialScale = new PopupMenu<>(Scales.values(), (p) -> {
            Settings.getInstance().setInitialUIScale(p.getSelectedItem());
            adjustResetButton();
        });
        mInitialScale.setSelectedItem(prefs.getInitialUIScale());
        content.add(new Label(I18n.text("Initial Scale"), mInitialScale), new PrecisionLayoutData().setFillHorizontalAlignment());
        wrapper = new Wrapper(new PrecisionLayout().setMargins(0).setColumns(7));
        content.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(2));
        wrapper.add(mInitialScale);

        mToolTipTimeout = new EditorField(FieldFactory.TOOLTIP_TIMEOUT, (f) -> {
            Settings.getInstance().setToolTipTimeout(((Integer) f.getValue()).intValue());
            adjustResetButton();
        }, SwingConstants.RIGHT, Integer.valueOf(prefs.getToolTipTimeout()),
                FieldFactory.getMaxValue(FieldFactory.TOOLTIP_TIMEOUT),
                I18n.text("The number of seconds before tooltips will dismiss themselves"));
        wrapper.add(new Label(I18n.text("Tooltip Timeout"), mToolTipTimeout), new PrecisionLayoutData().setFillHorizontalAlignment().setLeftMargin(5));
        wrapper.add(mToolTipTimeout, new PrecisionLayoutData().setFillHorizontalAlignment());
        wrapper.add(new Label(I18n.text("seconds"), mToolTipTimeout));

        mImageResolution = new EditorField(FieldFactory.OUTPUT_DPI, (f) -> {
            Settings.getInstance().setImageResolution(((Integer) f.getValue()).intValue());
            adjustResetButton();
        }, SwingConstants.RIGHT, Integer.valueOf(prefs.getImageResolution()),
                FieldFactory.getMaxValue(FieldFactory.OUTPUT_DPI),
                I18n.text("The resolution, in dots-per-inch, to use when saving sheets as PNG files"));
        wrapper.add(new Label(I18n.text("Image Resolution"), mImageResolution), new PrecisionLayoutData().setFillHorizontalAlignment().setLeftMargin(5));
        wrapper.add(mImageResolution, new PrecisionLayoutData().setFillHorizontalAlignment());
        wrapper.add(new Label(I18n.text("dpi"), mImageResolution));

        // Fourth row
        wrapper = new Wrapper(new PrecisionLayout().setMargins(0).setColumns(3));
        content.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
        mGCalcKey = new EditorField(FieldFactory.STRING, (f) -> {
            Settings.getInstance().setGURPSCalculatorKey(f.getText().trim());
            adjustResetButton();
        }, SwingConstants.LEFT, prefs.getGURPSCalculatorKey(), null);
        wrapper.add(new Label(I18n.text("GURPS Calculator Key"), mGCalcKey), new PrecisionLayoutData().setFillHorizontalAlignment());
        wrapper.add(mGCalcKey, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        wrapper.add(new Button(I18n.text("Find mine"), (btn) -> {
            try {
                Desktop.getDesktop().browse(new URI(ExportToGURPSCalculatorCommand.GURPS_CALCULATOR_URL));
            } catch (Exception exception) {
                Modal.showError(this, MessageFormat.format(I18n.text("Unable to open {0}"),
                        ExportToGURPSCalculatorCommand.GURPS_CALCULATOR_URL));
            }
        }));

        // Bottom row
        mResetButton = new Button(I18n.text("Reset to Factory Settings"), (btn) -> reset());
        content.add(mResetButton, new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.MIDDLE).setHorizontalSpan(3).setTopMargin(10));

        adjustResetButton();
        establishSizing();
        WindowUtils.packAndCenterWindowOn(this, null);
    }

    public static String getTechLevelTooltip() {
        return I18n.text("""
                TL0: Stone Age (Prehistory and later)
                TL1: Bronze Age (3500 B.C.+)
                TL2: Iron Age (1200 B.C.+)
                TL3: Medieval (600 A.D.+)
                TL4: Age of Sail (1450+)
                TL5: Industrial Revolution (1730+)
                TL6: Mechanized Age (1880+)
                TL7: Nuclear Age (1940+)
                TL8: Digital Age (1980+)
                TL9: Microtech Age (2025+?)
                TL10: Robotic Age (2070+?)
                TL11: Age of Exotic Matter
                TL12: Anything Goes""");
    }

    @Override
    public void establishSizing() {
        setMinimumSize(new Dimension(20, 20));
        setResizable(false);
    }

    private void reset() {
        mPlayerName.setValue(Settings.DEFAULT_DEFAULT_PLAYER_NAME);
        mTechLevel.setValue(Settings.DEFAULT_DEFAULT_TECH_LEVEL);
        mInitialPoints.setValue(Integer.valueOf(Settings.DEFAULT_INITIAL_POINTS));
        mAutoFillProfile.setChecked(Settings.DEFAULT_AUTO_FILL_PROFILE);
        mInitialScale.setSelectedItem(Settings.DEFAULT_INITIAL_UI_SCALE);
        mToolTipTimeout.setValue(Integer.valueOf(Settings.DEFAULT_TOOLTIP_TIMEOUT));
        mImageResolution.setValue(Integer.valueOf(Settings.DEFAULT_IMAGE_RESOLUTION));
        mIncludeUnspentPointsInTotal.setChecked(Settings.DEFAULT_INCLUDE_UNSPENT_POINTS_IN_TOTAL);
        mGCalcKey.setValue("");
        adjustResetButton();
    }

    private void adjustResetButton() {
        if (mResetButton != null) {
            mResetButton.setEnabled(!isSetToDefaults());
        }
    }

    private static boolean isSetToDefaults() {
        Settings prefs     = Settings.getInstance();
        boolean  atDefault = prefs.getDefaultPlayerName().equals(Settings.DEFAULT_DEFAULT_PLAYER_NAME);
        atDefault = atDefault && prefs.getDefaultTechLevel().equals(Settings.DEFAULT_DEFAULT_TECH_LEVEL);
        atDefault = atDefault && prefs.getInitialPoints() == Settings.DEFAULT_INITIAL_POINTS;
        atDefault = atDefault && prefs.autoFillProfile() == Settings.DEFAULT_AUTO_FILL_PROFILE;
        atDefault = atDefault && prefs.getInitialUIScale() == Settings.DEFAULT_INITIAL_UI_SCALE;
        atDefault = atDefault && prefs.getToolTipTimeout() == Settings.DEFAULT_TOOLTIP_TIMEOUT;
        atDefault = atDefault && prefs.getImageResolution() == Settings.DEFAULT_IMAGE_RESOLUTION;
        atDefault = atDefault && prefs.includeUnspentPointsInTotal() == Settings.DEFAULT_INCLUDE_UNSPENT_POINTS_IN_TOTAL;
        atDefault = atDefault && prefs.getGURPSCalculatorKey().isEmpty();
        return atDefault;
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
        synchronized (GeneralSettingsWindow.class) {
            INSTANCE = null;
        }
        super.dispose();
    }
}
