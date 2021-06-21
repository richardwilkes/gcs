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

import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.ui.Theme;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.Button;
import com.trollworks.gcs.ui.widget.FontPanel;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;

import java.awt.BorderLayout;
import java.awt.Dimension;
import java.awt.FlowLayout;
import java.awt.Font;
import java.awt.event.WindowEvent;
import java.util.ArrayList;
import java.util.List;

/** A window for editing font settings. */
public final class FontSettingsWindow extends BaseWindow implements CloseHandler {
    private static FontSettingsWindow INSTANCE;
    private        List<FontTracker>  mFontPanels;
    private        Button             mResetButton;
    private        boolean            mIgnore;

    /** Displays the theme settings window. */
    public static void display() {
        if (!UIUtilities.inModalState()) {
            FontSettingsWindow wnd;
            synchronized (FontSettingsWindow.class) {
                if (INSTANCE == null) {
                    INSTANCE = new FontSettingsWindow();
                }
                wnd = INSTANCE;
            }
            wnd.setVisible(true);
        }
    }

    private FontSettingsWindow() {
        super(I18n.text("Font Settings"));
        Panel panel = new Panel(new PrecisionLayout().setColumns(2).setMargins(20, 20, 0, 20), false);
        mFontPanels = new ArrayList<>();
        for (ThemeFont font : ThemeFont.ALL) {
            if (font.isEditable()) {
                FontTracker tracker = new FontTracker(font);
                panel.add(new Label(font.toString()), new PrecisionLayoutData().setFillHorizontalAlignment());
                panel.add(tracker);
                mFontPanels.add(tracker);
            }
        }
        getContentPane().add(new ScrollPanel(panel), BorderLayout.CENTER);
        addResetPanel();
        adjustResetButton();
        establishSizing();
        WindowUtils.packAndCenterWindowOn(this, null);
    }

    private void addResetPanel() {
        Panel panel = new Panel(new FlowLayout(FlowLayout.CENTER));
        panel.setBorder(new EmptyBorder(20));
        mResetButton = new Button(I18n.text("Reset to Factory Settings"), (btn) -> resetFonts());
        panel.add(mResetButton);
        getContentPane().add(panel, BorderLayout.SOUTH);
    }

    @Override
    public void establishSizing() {
        pack();
        int width = getSize().width;
        setMinimumSize(new Dimension(width, 200));
        setMaximumSize(new Dimension(width, getPreferredSize().height));
    }

    private void resetFonts() {
        mIgnore = true;
        for (FontTracker tracker : mFontPanels) {
            tracker.reset();
        }
        mIgnore = false;
        BaseWindow.forceRevalidateAndRepaint();
        adjustResetButton();
    }

    private void adjustResetButton() {
        boolean enabled = false;
        for (ThemeFont font : ThemeFont.ALL) {
            if (font.isEditable() && !font.getFont().equals(Theme.DEFAULT.getFont(font.getIndex()))) {
                enabled = true;
                break;
            }
        }
        mResetButton.setEnabled(enabled);
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
        synchronized (FontSettingsWindow.class) {
            INSTANCE = null;
        }
        super.dispose();
    }

    private class FontTracker extends FontPanel {
        private int mIndex;

        FontTracker(ThemeFont font) {
            super(font.getFont());
            mIndex = font.getIndex();
            addActionListener((evt) -> {
                if (!mIgnore) {
                    Theme.current().setFont(mIndex, getCurrentFont());
                    adjustResetButton();
                    BaseWindow.forceRevalidateAndRepaint();
                }
            });
        }

        void reset() {
            Font font = Theme.DEFAULT.getFont(mIndex);
            setCurrentFont(font);
            Theme.current().setFont(mIndex, font);
        }
    }
}
