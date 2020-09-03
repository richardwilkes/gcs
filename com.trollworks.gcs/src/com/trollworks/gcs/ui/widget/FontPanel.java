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

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.FlowLayout;
import java.awt.Font;
import java.awt.GraphicsEnvironment;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import javax.swing.JComboBox;

/** A standard font selection panel. */
public class FontPanel extends ActionPanel implements ActionListener {
    private JComboBox<Integer>         mFontSizeMenu;
    private JComboBox<String>          mFontNameMenu;
    private JComboBox<Fonts.FontStyle> mFontStyleMenu;
    private boolean                    mNoNotify;

    /**
     * Creates a new font panel.
     *
     * @param font The font to start with.
     */
    public FontPanel(Font font) {
        super(new FlowLayout(FlowLayout.LEFT, 5, 0));
        setOpaque(false);

        mFontNameMenu = new JComboBox<>(GraphicsEnvironment.getLocalGraphicsEnvironment().getAvailableFontFamilyNames());
        mFontNameMenu.setOpaque(false);
        mFontNameMenu.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Changes the font")));
        mFontNameMenu.setMaximumRowCount(25);
        mFontNameMenu.addActionListener(this);
        UIUtilities.setToPreferredSizeOnly(mFontNameMenu);
        add(mFontNameMenu);

        Integer[] sizes = new Integer[10];
        for (int i = 0; i < 7; i++) {
            sizes[i] = Integer.valueOf(6 + i);
        }
        sizes[7] = Integer.valueOf(14);
        sizes[8] = Integer.valueOf(16);
        sizes[9] = Integer.valueOf(18);
        mFontSizeMenu = new JComboBox<>(sizes);
        mFontSizeMenu.setOpaque(false);
        mFontSizeMenu.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Changes the font size")));
        mFontSizeMenu.setMaximumRowCount(sizes.length);
        mFontSizeMenu.addActionListener(this);
        UIUtilities.setToPreferredSizeOnly(mFontSizeMenu);
        add(mFontSizeMenu);

        mFontStyleMenu = new JComboBox<>(Fonts.FontStyle.values());
        mFontStyleMenu.setOpaque(false);
        mFontStyleMenu.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Changes the font style")));
        mFontStyleMenu.addActionListener(this);
        UIUtilities.setToPreferredSizeOnly(mFontStyleMenu);
        add(mFontStyleMenu);

        setCurrentFont(font);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        notifyActionListeners();
    }

    @Override
    public void notifyActionListeners(ActionEvent event) {
        if (!mNoNotify) {
            super.notifyActionListeners(event);
        }
    }

    /** @return The font this panel has been set to. */
    public Font getCurrentFont() {
        String name = (String) mFontNameMenu.getSelectedItem();
        if (name == null) {
            name = "SansSerif";
        }
        Fonts.FontStyle style = (Fonts.FontStyle) mFontStyleMenu.getSelectedItem();
        if (style == null) {
            style = Fonts.FontStyle.PLAIN;
        }
        Integer size = (Integer) mFontSizeMenu.getSelectedItem();
        if (size == null) {
            size = Integer.valueOf(12);
        }
        return new Font(name, style.ordinal(), size.intValue());
    }

    /** @param font The new font. */
    public void setCurrentFont(Font font) {
        mNoNotify = true;
        mFontNameMenu.setSelectedItem(font.getName());
        if (mFontNameMenu.getSelectedItem() == null) {
            mFontNameMenu.setSelectedIndex(0);
        }
        mFontSizeMenu.setSelectedItem(Integer.valueOf(font.getSize()));
        if (mFontSizeMenu.getSelectedItem() == null) {
            mFontSizeMenu.setSelectedIndex(3);
        }
        mFontStyleMenu.setSelectedItem(Fonts.FontStyle.from(font));
        if (mFontStyleMenu.getSelectedItem() == null) {
            mFontStyleMenu.setSelectedIndex(0);
        }
        mNoNotify = false;
        notifyActionListeners();
    }
}
