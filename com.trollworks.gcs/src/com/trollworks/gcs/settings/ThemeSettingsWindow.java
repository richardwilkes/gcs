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
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.Theme;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.ColorWell;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.ui.widget.FontPanel;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Container;
import java.awt.Font;
import java.awt.event.WindowEvent;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JSeparator;
import javax.swing.SwingConstants;
import javax.swing.UIManager;

/** A window for editing theme settings. */
public final class ThemeSettingsWindow extends BaseWindow implements CloseHandler {
    private static ThemeSettingsWindow INSTANCE;
    private        FontPanel[]         mFontPanels;
    private        List<ColorTracker>  mColorWells;
    private        FontAwesomeButton   mResetFontsButton;
    private        FontAwesomeButton   mResetColorsButton;
    private        boolean             mIgnore;

    /** Displays the theme settings window. */
    public static void display() {
        if (!UIUtilities.inModalState()) {
            ThemeSettingsWindow wnd;
            synchronized (ThemeSettingsWindow.class) {
                if (INSTANCE == null) {
                    INSTANCE = new ThemeSettingsWindow();
                }
                wnd = INSTANCE;
            }
            wnd.setVisible(true);
        }
    }

    private ThemeSettingsWindow() {
        super(I18n.text("Theme Settings"));
        JPanel panel = new JPanel(new PrecisionLayout().setMargins(10));

        mResetFontsButton = addHeader(panel, I18n.text("Fonts"), 0, this::resetFonts);
        JPanel wrapper = new JPanel(new PrecisionLayout().setColumns(2));
        wrapper.setOpaque(false);
        String[] keys = Fonts.getKeys();
        mFontPanels = new FontPanel[keys.length];
        int i = 0;
        for (String key : keys) {
            JLabel label = new JLabel(Fonts.getDescription(key), SwingConstants.RIGHT);
            label.setOpaque(false);
            wrapper.add(label, new PrecisionLayoutData().setFillHorizontalAlignment());
            mFontPanels[i] = new FontPanel(UIManager.getFont(key));
            mFontPanels[i].setActionCommand(key);
            mFontPanels[i].addActionListener(evt -> {
                if (!mIgnore) {
                    Object source = evt.getSource();
                    if (source instanceof FontPanel) {
                        boolean adjusted = false;
                        for (FontPanel fp : mFontPanels) {
                            if (fp == source) {
                                Font   font = fp.getCurrentFont();
                                String cmd  = fp.getActionCommand();
                                if (!font.equals(UIManager.getFont(cmd))) {
                                    UIManager.put(cmd, font);
                                    Preferences.getInstance().setFontInfo(cmd, new Fonts.Info(font));
                                    adjusted = true;
                                }
                                break;
                            }
                        }
                        if (adjusted) {
                            BaseWindow.forceRepaintAndInvalidate();
                        }
                    }
                    adjustResetButtons();
                }
            });
            wrapper.add(mFontPanels[i]);
            i++;
        }
        panel.add(wrapper);

        mResetColorsButton = addHeader(panel, I18n.text("Colors"), 16, this::resetColors);
        int cols = 8;
        wrapper = new JPanel(new PrecisionLayout().setColumns(cols));
        wrapper.setOpaque(false);
        mColorWells = new ArrayList<>();
        int max = ThemeColor.ALL.size();
        cols /= 2;
        int maxPerCol  = max / cols;
        int excess     = max % (maxPerCol * cols);
        int iterations = maxPerCol;
        if (excess != 0) {
            iterations++;
        }
        for (i = 0; i < iterations; i++) {
            addColorTracker(wrapper, ThemeColor.ALL.get(i), 0);
            int index = i;
            for (int j = 1; j < ((i == maxPerCol) ? excess : cols); j++) {
                index += maxPerCol;
                if (j - 1 < excess) {
                    index++;
                }
                if (index < max) {
                    addColorTracker(wrapper, ThemeColor.ALL.get(index), 8);
                }
            }
        }
        panel.add(wrapper);

        Container   content  = getContentPane();
        JScrollPane scroller = new JScrollPane(panel);
        scroller.setBorder(null);
        content.add(scroller, BorderLayout.CENTER);
        adjustResetButtons();
        WindowUtils.packAndCenterWindowOn(this, null);
    }

    private FontAwesomeButton addHeader(Container parent, String text, int topMargin, Runnable reset) {
        JPanel header = new JPanel(new PrecisionLayout().setColumns(2).setMargins(0));
        JLabel label  = new JLabel(text);
        label.setFont(label.getFont().deriveFont(Font.BOLD));
        header.add(label);
        FontAwesomeButton resetButton = new FontAwesomeButton("\uf011", I18n.text("Reset to Factory Defaults"), reset);
        header.add(resetButton, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.END));
        header.add(new JSeparator(SwingConstants.HORIZONTAL), new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(2));
        parent.add(header, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setTopMargin(topMargin));
        return resetButton;
    }

    private void addColorTracker(Container parent, ThemeColor color, int leftMargin) {
        JLabel label = new JLabel(color.toString(), SwingConstants.RIGHT);
        label.setOpaque(false);
        parent.add(label, new PrecisionLayoutData().setFillHorizontalAlignment().setLeftMargin(leftMargin));
        ColorTracker tracker = new ColorTracker(color);
        mColorWells.add(tracker);
        parent.add(tracker, new PrecisionLayoutData().setLeftMargin(4));
    }

    private void resetFonts() {
        mIgnore = true;
        Fonts.restoreDefaults();
        for (FontPanel panel : mFontPanels) {
            panel.setCurrentFont(UIManager.getFont(panel.getActionCommand()));
        }
        mIgnore = false;
        BaseWindow.forceRepaintAndInvalidate();
        Theme.repaint();
        adjustResetButtons();
    }

    private void resetColors() {
        mIgnore = true;
        for (ColorTracker tracker : mColorWells) {
            tracker.reset();
        }
        mIgnore = false;
        BaseWindow.forceRepaintAndInvalidate();
        Theme.repaint();
        adjustResetButtons();
    }

    private void adjustResetButtons() {
        boolean enabled = false;
        for (ThemeColor color : ThemeColor.ALL) {
            if (color.getRGB() != Theme.DEFAULT.getColor(color.getIndex()).getRGB()) {
                enabled = true;
                break;
            }
        }
        mResetColorsButton.setEnabled(enabled);
        mResetFontsButton.setEnabled(!Fonts.isSetToDefaults());
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
        synchronized (ThemeSettingsWindow.class) {
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
                adjustResetButtons();
                Theme.repaint();
            }
        }
    }
}
