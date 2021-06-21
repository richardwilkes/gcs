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
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.Button;
import com.trollworks.gcs.ui.widget.ColorWell;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.FlowLayout;
import java.awt.event.WindowEvent;
import java.util.ArrayList;
import java.util.List;

/** A window for editing color settings. */
public final class ColorSettingsWindow extends BaseWindow implements CloseHandler {
    private static ColorSettingsWindow INSTANCE;
    private        List<ColorTracker>  mColorWells;
    private        Button              mResetButton;
    private        boolean             mIgnore;

    /** Displays the color settings window. */
    public static void display() {
        if (!UIUtilities.inModalState()) {
            ColorSettingsWindow wnd;
            synchronized (ColorSettingsWindow.class) {
                if (INSTANCE == null) {
                    INSTANCE = new ColorSettingsWindow();
                }
                wnd = INSTANCE;
            }
            wnd.setVisible(true);
        }
    }

    private ColorSettingsWindow() {
        super(I18n.text("Color Settings"));
        int   cols  = 8;
        Panel panel = new Panel(new PrecisionLayout().setColumns(cols).setMargins(20, 20, 0, 20), false);
        mColorWells = new ArrayList<>();
        int max = ThemeColor.ALL.size();
        cols /= 2;
        int maxPerCol  = max / cols;
        int excess     = max % (maxPerCol * cols);
        int iterations = maxPerCol;
        if (excess != 0) {
            iterations++;
        }
        for (int i = 0; i < iterations; i++) {
            addColorTracker(panel, ThemeColor.ALL.get(i), 0);
            int index = i;
            for (int j = 1; j < ((i == maxPerCol) ? excess : cols); j++) {
                index += maxPerCol;
                if (j - 1 < excess) {
                    index++;
                }
                if (index < max) {
                    addColorTracker(panel, ThemeColor.ALL.get(index), 8);
                }
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
        mResetButton = new Button(I18n.text("Reset to Factory Settings"), (btn) -> resetColors());
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

    private void addColorTracker(Container parent, ThemeColor color, int leftMargin) {
        ColorTracker tracker = new ColorTracker(color);
        mColorWells.add(tracker);
        parent.add(new Label(color.toString()), new PrecisionLayoutData().setFillHorizontalAlignment().setLeftMargin(leftMargin));
        parent.add(tracker, new PrecisionLayoutData().setLeftMargin(4));
    }

    private void resetColors() {
        mIgnore = true;
        for (ColorTracker tracker : mColorWells) {
            tracker.reset();
        }
        mIgnore = false;
        Theme.repaint();
        adjustResetButton();
    }

    private void adjustResetButton() {
        boolean enabled = false;
        for (ThemeColor color : ThemeColor.ALL) {
            if (color.getRGB() != Theme.DEFAULT.getColor(color.getIndex()).getRGB()) {
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
        synchronized (ColorSettingsWindow.class) {
            INSTANCE = null;
        }
        super.dispose();
    }

    private class ColorTracker extends ColorWell implements ColorWell.ColorChangedListener {
        private int mIndex;

        ColorTracker(ThemeColor color) {
            super(new Color(color.getRGB(), true), null);
            mIndex = color.getIndex();
            setColorChangedListener(this);
        }

        void reset() {
            Color color = Theme.DEFAULT.getColor(mIndex);
            setWellColor(color);
            Theme.current().setColor(mIndex, color);
        }

        @Override
        public void colorChanged(Color color) {
            if (!mIgnore) {
                Theme.current().setColor(mIndex, color);
                adjustResetButton();
                Theme.repaint();
            }
        }
    }
}
