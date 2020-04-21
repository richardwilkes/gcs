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

import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import javax.swing.JColorChooser;
import javax.swing.JPanel;

public class ColorWell extends JPanel implements MouseListener {
    private Color                mColor;
    private ColorChangedListener mListener;

    public ColorWell(Color color, ColorChangedListener listener) {
        this(color, listener, null);
    }

    public ColorWell(Color color, ColorChangedListener listener, String tooltip) {
        mColor = color;
        mListener = listener;
        setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        setBorder(new LineBorder());
        UIUtilities.setOnlySize(this, new Dimension(22, 22));
        addMouseListener(this);
    }

    public Color getWellColor() {
        return mColor;
    }

    public void setWellColor(Color color) {
        mColor = color;
        repaint();
    }

    @Override
    protected void paintComponent(Graphics g) {
        super.paintComponent(g);
        Rectangle bounds = UIUtilities.getLocalInsetBounds(this);
        g.setColor(Color.WHITE);
        g.fillRect(bounds.x, bounds.y, bounds.width, bounds.height);
        g.setColor(Color.LIGHT_GRAY);
        int xs     = bounds.width / 4;
        int ys     = bounds.height / 4;
        int offset = 0;
        for (int y = bounds.y; y < bounds.y + bounds.height; y += ys) {
            for (int x = bounds.x + offset; x < bounds.x + bounds.width; x += xs * 2) {
                g.fillRect(x, y, xs, ys);
            }
            offset = offset == 0 ? xs : 0;
        }
        g.setColor(mColor);
        g.fillRect(bounds.x, bounds.y, bounds.width, bounds.height);
    }

    @Override
    public void mouseClicked(MouseEvent event) {
        Color color = JColorChooser.showDialog(this, null, mColor);
        if (color != null) {
            if (!mColor.equals(color)) {
                mColor = color;
                repaint();
                if (mListener != null) {
                    mListener.colorChanged(mColor);
                }
            }
        }
    }

    @Override
    public void mousePressed(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseReleased(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseEntered(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseExited(MouseEvent event) {
        // Unused
    }

    public interface ColorChangedListener {
        void colorChanged(Color color);
    }
}
