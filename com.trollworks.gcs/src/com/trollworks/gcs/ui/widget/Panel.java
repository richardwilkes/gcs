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
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.ThemeFont;

import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.LayoutManager;
import javax.swing.JPanel;

public class Panel extends JPanel {
    private ThemeFont mThemeFont;

    public Panel() {
        init(true);
    }

    public Panel(LayoutManager layout) {
        super(layout);
        init(true);
    }

    public Panel(LayoutManager layout, boolean opaque) {
        super(layout);
        init(opaque);
    }

    private void init(boolean opaque) {
        setStdColors();
        setThemeFont(ThemeFont.LABEL_PRIMARY);
        setOpaque(opaque);
        setBorder(null);
    }

    protected void setStdColors() {
        setBackground(ThemeColor.BACKGROUND);
        setForeground(ThemeColor.ON_BACKGROUND);
    }

    protected void paintComponent(Graphics g) {
        Graphics2D gc = GraphicsUtilities.prepare(g);
        if (isOpaque()) {
            gc.setColor(getBackground());
            gc.fillRect(0, 0, getWidth(), getHeight());
        }
        gc.setColor(getForeground());
    }

    @Override
    public ToolTip createToolTip() {
        return new ToolTip(this);
    }

    public final ThemeFont getThemeFont() {
        return mThemeFont;
    }

    public final void setThemeFont(ThemeFont font) {
        mThemeFont = font;
    }

    @Override
    public final Font getFont() {
        return mThemeFont != null ? mThemeFont.getFont() : super.getFont();
    }
}
