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

package com.trollworks.gcs.ui.border;

import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Component;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Insets;
import java.awt.Rectangle;
import javax.swing.SwingConstants;

/** A border consisting of a frame and optional title. */
public class TitledBorder extends LineBorder {
    private String mTitle;
    private Font   mFont;

    /** Creates a new border without a title. */
    public TitledBorder() {
        super(ThemeColor.HEADER);
    }

    /**
     * Creates a new border without a title.
     *
     * @param font  The font to use.
     * @param title The title to use.
     */
    public TitledBorder(Font font, String title) {
        this();
        mFont = font;
        mTitle = title;
    }

    /** @return The font. */
    public Font getFont() {
        return mFont;
    }

    /** @param font The font to use. */
    public void setFont(Font font) {
        mFont = font;
    }

    /** @return The title. */
    public String getTitle() {
        return mTitle;
    }

    /** @param title The new title. */
    public void setTitle(String title) {
        mTitle = title;
    }

    @Override
    public Insets getBorderInsets(Component component) {
        setThickness(Edge.TOP, (mTitle != null && mFont != null) ? TextDrawing.getPreferredSize(mFont, mTitle).height : Scale.get(component).scale(1));
        return super.getBorderInsets(component);
    }

    @Override
    public void paintBorder(Component component, Graphics graphics, int x, int y, int width, int height) {
        super.paintBorder(component, graphics, x, y, width, height);
        if (mTitle != null && mFont != null) {
            Scale scale     = Scale.get(component);
            Font  savedFont = graphics.getFont();
            Font  font      = scale.scale(mFont);
            int   one       = scale.scale(1);
            graphics.setFont(font);
            graphics.setColor(ThemeColor.ON_HEADER);
            TextDrawing.draw(graphics, new Rectangle(x + one, y, width - (one + one), TextDrawing.getPreferredSize(font, mTitle).height), mTitle, SwingConstants.CENTER, SwingConstants.TOP);
            graphics.setFont(savedFont);
        }
    }

    public int getMinimumWidth(Component component) {
        Scale scale = Scale.get(component);
        if (mTitle != null) {
            return TextDrawing.getPreferredSize(scale.scale(mFont), mTitle).width + scale.scale(1) * 4;
        }
        return scale.scale(1) * 4;
    }
}
