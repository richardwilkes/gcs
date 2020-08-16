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

import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.Theme;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.ColorWell;
import com.trollworks.gcs.ui.widget.FontPanel;
import com.trollworks.gcs.utility.I18n;

import java.awt.Color;
import java.awt.Font;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JSeparator;
import javax.swing.SwingConstants;
import javax.swing.UIManager;

/** The theme preferences panel. */
public class ThemePreferences extends PreferencePanel implements ActionListener {
    private FontPanel[]        mFontPanels;
    private List<ColorTracker> mColorWells;
    private boolean            mIgnore;

    /**
     * Creates a new {@link ThemePreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public ThemePreferences(PreferencesWindow owner) {
        super(I18n.Text("Theme"), owner);
        setLayout(new PrecisionLayout());

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
            mFontPanels[i].addActionListener(this);
            wrapper.add(mFontPanels[i]);
            i++;
        }
        addHeader(I18n.Text("Fonts"), 0);
        add(wrapper);

        int cols = 8;
        wrapper = new JPanel(new PrecisionLayout().setColumns(cols));
        wrapper.setOpaque(false);
        mColorWells = new ArrayList<>();
        int max = ThemeColor.ALL.size();
        cols /= 2;
        int maxPerCol = max / cols;
        int excess    = max % (maxPerCol * cols);
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
        addHeader(I18n.Text("Colors"), 16);
        add(wrapper);
    }

    private void addHeader(String text, int topMargin) {
        JLabel label = new JLabel(text);
        label.setFont(label.getFont().deriveFont(Font.BOLD));
        add(label, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setTopMargin(topMargin));
        add(new JSeparator(SwingConstants.HORIZONTAL), new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private void addColorTracker(JPanel wrapper, ThemeColor color, int leftMargin) {
        JLabel label = new JLabel(color.toString(), SwingConstants.RIGHT);
        label.setOpaque(false);
        wrapper.add(label, new PrecisionLayoutData().setFillHorizontalAlignment().setLeftMargin(leftMargin));
        ColorTracker tracker = new ColorTracker(color);
        mColorWells.add(tracker);
        wrapper.add(tracker, new PrecisionLayoutData().setLeftMargin(4));
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        if (!mIgnore) {
            Object source = event.getSource();
            if (source instanceof FontPanel) {
                boolean adjusted = false;
                for (FontPanel panel : mFontPanels) {
                    if (panel == source) {
                        Font   font = panel.getCurrentFont();
                        String key  = panel.getActionCommand();
                        if (!font.equals(UIManager.getFont(key))) {
                            UIManager.put(key, font);
                            Preferences.getInstance().setFontInfo(key, new Fonts.Info(font));
                            adjusted = true;
                        }
                        break;
                    }
                }
                if (adjusted) {
                    BaseWindow.forceRepaintAndInvalidate();
                    Fonts.notifyOfFontChanges();
                }
            }
            adjustResetButton();
        }
    }

    @Override
    public void reset() {
        mIgnore = true;
        for (ColorTracker tracker : mColorWells) {
            tracker.reset();
        }
        Fonts.restoreDefaults();
        for (FontPanel panel : mFontPanels) {
            panel.setCurrentFont(UIManager.getFont(panel.getActionCommand()));
        }
        mIgnore = false;
        BaseWindow.forceRepaintAndInvalidate();
        Fonts.notifyOfFontChanges();
        Theme.repaint();
    }

    @Override
    public boolean isSetToDefaults() {
        for (ThemeColor color : ThemeColor.ALL) {
            if (color.getRGB() != Theme.DEFAULT.getColor(color.getIndex()).getRGB()) {
                return false;
            }
        }
        return Fonts.isSetToDefaults();
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
