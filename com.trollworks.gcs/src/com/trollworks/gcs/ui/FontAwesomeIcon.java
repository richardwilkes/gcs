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

import java.awt.Component;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import javax.swing.Icon;
import javax.swing.SwingConstants;

public class FontAwesomeIcon implements Icon {
    private String     mValue;
    private int        mSize;
    private ThemeColor mColor;

    public FontAwesomeIcon(String value, int size, ThemeColor color) {
        mValue = value;
        mSize = size;
        mColor = color;
    }

    @Override
    public void paintIcon(Component component, Graphics g, int x, int y) {
        Scale      scale = Scale.get(component);
        Dimension  size  = TextDrawing.getPreferredSize(new Font(Fonts.FONT_AWESOME_SOLID, Font.PLAIN, scale.scale(mSize)), mValue);
        Graphics2D gc    = GraphicsUtilities.prepare(g.create());
        gc.setColor(mColor);
        gc.setFont(new Font(Fonts.FONT_AWESOME_SOLID, Font.PLAIN, scale.scale(mSize)));
        TextDrawing.draw(gc, new Rectangle(x, y, size.width, size.height), mValue, SwingConstants.CENTER, SwingConstants.CENTER);
        gc.dispose();
    }

    @Override
    public int getIconWidth() {
        return TextDrawing.getPreferredSize(new Font(Fonts.FONT_AWESOME_SOLID, Font.PLAIN, mSize), mValue).width;
    }

    @Override
    public int getIconHeight() {
        return TextDrawing.getPreferredSize(new Font(Fonts.FONT_AWESOME_SOLID, Font.PLAIN, mSize), mValue).height;
    }
}
