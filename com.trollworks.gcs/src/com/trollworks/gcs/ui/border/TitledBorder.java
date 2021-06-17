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

package com.trollworks.gcs.ui.border;

import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Color;
import java.awt.Component;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import javax.swing.SwingConstants;

/** A border consisting of a frame and optional title. */
public class TitledBorder extends LineBorder {
    private String    mTitle;
    private ThemeFont mFont;
    private Color     mTitleColor;

    /** Creates a new border without a title. */
    public TitledBorder() {
        super(ThemeColor.HEADER);
        mTitleColor = ThemeColor.ON_HEADER;
    }

    /**
     * Creates a new border.
     *
     * @param font  The font to use.
     * @param title The title to use.
     */
    public TitledBorder(ThemeFont font, String title) {
        this();
        mFont = font;
        mTitle = title;
    }

    /**
     * Creates a new border.
     *
     * @param borderColor The border color.
     * @param titleColor  The title color.
     * @param font        The font to use.
     * @param title       The title to use.
     */
    public TitledBorder(Color borderColor, Color titleColor, ThemeFont font, String title) {
        super(borderColor);
        mFont = font;
        mTitle = title;
        mTitleColor = titleColor;
    }

    /** @return The font. */
    public ThemeFont getFont() {
        return mFont;
    }

    /** @param font The font to use. */
    public void setFont(ThemeFont font) {
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
        setThickness(Edge.TOP, (mTitle != null && mFont != null) ? TextDrawing.getPreferredSize(mFont.getFont(), mTitle).height : Scale.get(component).scale(1));
        return super.getBorderInsets(component);
    }

    @Override
    public void paintBorder(Component component, Graphics g, int x, int y, int width, int height) {
        super.paintBorder(component, g, x, y, width, height);
        if (mTitle != null && mFont != null) {
            Graphics2D gc        = GraphicsUtilities.prepare(g);
            Scale      scale     = Scale.get(component);
            Font       savedFont = gc.getFont();
            Font       font      = scale.scale(mFont.getFont());
            int        one       = scale.scale(1);
            gc.setFont(font);
            gc.setColor(mTitleColor);
            TextDrawing.draw(gc, new Rectangle(x + one, y, width - (one + one), TextDrawing.getPreferredSize(font, mTitle).height), mTitle, SwingConstants.CENTER, SwingConstants.TOP);
            gc.setFont(savedFont);
        }
    }

    public int getMinimumWidth(Component component) {
        Scale scale = Scale.get(component);
        if (mTitle != null && mFont != null) {
            return TextDrawing.getPreferredSize(scale.scale(mFont.getFont()), mTitle).width + scale.scale(1) * 4;
        }
        return scale.scale(1) * 4;
    }
}
