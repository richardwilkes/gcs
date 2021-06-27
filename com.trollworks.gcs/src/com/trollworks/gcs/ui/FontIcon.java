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

package com.trollworks.gcs.ui;

import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Color;
import java.awt.Component;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import javax.swing.Icon;
import javax.swing.SwingConstants;

public class FontIcon implements Icon {
    private String    mGlyph;
    private ThemeFont mFont;
    private Color     mColor;

    /**
     * @param glyph The glyph to use.
     * @param font  The font to use.
     */
    public FontIcon(String glyph, ThemeFont font) {
        mGlyph = glyph;
        mFont = font;
    }

    /**
     * @param glyph The glyph to use.
     * @param font  The font to use.
     * @param color The color to use. May be {@code null}, in which case the current graphics color
     *              is used.
     */
    public FontIcon(String glyph, ThemeFont font, Color color) {
        mGlyph = glyph;
        mFont = font;
        mColor = color;
    }

    @Override
    public void paintIcon(Component component, Graphics g, int x, int y) {
        Scale      scale = Scale.get(component);
        Font       font  = scale.scale(mFont.getFont());
        Dimension  size  = TextDrawing.getPreferredSize(font, mGlyph);
        Graphics2D gc    = GraphicsUtilities.prepare(g.create());
        if (mColor != null) {
            gc.setColor(mColor);
        }
        gc.setFont(font);
        TextDrawing.draw(gc, new Rectangle(x, y, size.width, size.height), mGlyph,
                SwingConstants.CENTER, SwingConstants.CENTER);
        gc.dispose();
    }

    @Override
    public int getIconWidth() {
        return TextDrawing.getPreferredSize(mFont.getFont(), mGlyph).width;
    }

    @Override
    public int getIconHeight() {
        return TextDrawing.getPreferredSize(mFont.getFont(), mGlyph).height;
    }
}
