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

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import javax.swing.SwingConstants;

public class FontAwesomeIcon extends Panel {
    private String mText;
    private int    mSize;
    private int    mMargin;

    public FontAwesomeIcon(String text, int size, int margin, String tooltip) {
        super(null, false);
        setToolTipText(tooltip);
        setText(text);
        setSize(size);
        setMargin(margin);
    }

    public void setText(String text) {
        mText = text;
        repaint();
    }

    public void setSize(int size) {
        mSize = size;
    }

    public void setMargin(int margin) {
        mMargin = margin;
    }

    @Override
    public Dimension getPreferredSize() {
        Scale     scale = Scale.get(this);
        Dimension size  = TextDrawing.getPreferredSize(new Font(ThemeFont.FONT_AWESOME_SOLID, Font.PLAIN, scale.scale(mSize)), mText);
        if (mMargin != 0) {
            size.width += scale.scale(mMargin) * 2;
            size.height *= scale.scale(mMargin) * 2;
        }
        Insets insets = getInsets();
        size.width += insets.left + insets.right;
        size.height += insets.top + insets.bottom;
        return size;
    }

    @Override
    public Dimension getMinimumSize() {
        return getPreferredSize();
    }

    @Override
    public Dimension getMaximumSize() {
        return getPreferredSize();
    }

    @Override
    protected void paintComponent(Graphics g) {
        Rectangle bounds = getBounds();
        Insets    insets = getInsets();
        bounds.x = insets.left;
        bounds.y = insets.top;
        bounds.width -= insets.left + insets.right;
        bounds.height -= insets.top + insets.bottom;
        Scale      scale = Scale.get(this);
        Graphics2D gc    = GraphicsUtilities.prepare(g);
        gc.setFont(new Font(ThemeFont.FONT_AWESOME_SOLID, Font.PLAIN, scale.scale(mSize)));
        TextDrawing.draw(gc, bounds, mText, SwingConstants.CENTER, SwingConstants.CENTER);
    }
}
