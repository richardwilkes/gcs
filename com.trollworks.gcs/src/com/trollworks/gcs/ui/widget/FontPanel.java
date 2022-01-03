/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.ui.FontStyle;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.layout.PrecisionLayout;

import java.awt.Font;
import java.awt.GraphicsEnvironment;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;

/** A standard font selection panel. */
public class FontPanel extends ActionPanel implements ActionListener {
    private PopupMenu<Integer>   mFontSizeMenu;
    private PopupMenu<String>    mFontNameMenu;
    private PopupMenu<FontStyle> mFontStyleMenu;
    private boolean              mNoNotify;

    /**
     * Creates a new font panel.
     *
     * @param font The font to start with.
     */
    public FontPanel(Font font) {
        super(new PrecisionLayout().setColumns(3).setMargins(0));
        setOpaque(false);

        mFontNameMenu = new PopupMenu<>(GraphicsEnvironment.getLocalGraphicsEnvironment().getAvailableFontFamilyNames(),
                (p) -> notifyActionListeners());
        add(mFontNameMenu);

        Integer[] sizes = new Integer[20];
        for (int i = 0; i < 20; i++) {
            sizes[i] = Integer.valueOf(5 + i);
        }
        mFontSizeMenu = new PopupMenu<>(sizes, (p) -> notifyActionListeners());
        add(mFontSizeMenu);

        mFontStyleMenu = new PopupMenu<>(FontStyle.values(), (p) -> notifyActionListeners());
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
        String name = mFontNameMenu.getSelectedItem();
        if (name == null) {
            name = Fonts.ROBOTO;
        }
        FontStyle style = mFontStyleMenu.getSelectedItem();
        if (style == null) {
            style = FontStyle.PLAIN;
        }
        Integer size = mFontSizeMenu.getSelectedItem();
        if (size == null) {
            size = Integer.valueOf(12);
        }
        return new Font(name, style.ordinal(), size.intValue());
    }

    /** @param font The new font. */
    public void setCurrentFont(Font font) {
        mNoNotify = true;
        mFontNameMenu.setSelectedItem(font.getName(), true);
        if (mFontNameMenu.getSelectedItem() == null) {
            mFontNameMenu.setSelectedIndex(0, true);
        }
        mFontSizeMenu.setSelectedItem(Integer.valueOf(font.getSize()), true);
        if (mFontSizeMenu.getSelectedItem() == null) {
            mFontSizeMenu.setSelectedIndex(3, true);
        }
        mFontStyleMenu.setSelectedItem(FontStyle.from(font), true);
        if (mFontStyleMenu.getSelectedItem() == null) {
            mFontStyleMenu.setSelectedIndex(0, true);
        }
        mNoNotify = false;
        notifyActionListeners();
    }
}
