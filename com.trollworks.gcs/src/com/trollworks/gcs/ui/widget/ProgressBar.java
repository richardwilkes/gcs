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

import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.UIUtilities;

import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.RenderingHints;

public class ProgressBar extends Panel {
    private static final int  MIN_BAR_WIDTH   = 20;
    private static final int  THUMB_WIDTH     = MIN_BAR_WIDTH / 2;
    private static final int  BAR_HEIGHT      = 8;
    private static final int  ARC             = BAR_HEIGHT;
    private static final long ANIMATION_SPEED = 1000;

    private int  mCurrent;
    private int  mMax;
    private long mAnimation;

    public ProgressBar(int max) {
        super(null, false);
        setBackground(Colors.CONTROL);
        setForeground(Colors.ACCENT);
        mMax = Math.max(max, 0);
        mAnimation = -1;
    }

    public int getCurrent() {
        return mCurrent;
    }

    public void setCurrent(int value) {
        if (value < 0) {
            value = 0;
        } else if (value > mMax) {
            value = mMax;
        }
        if (mCurrent != value) {
            mCurrent = value;
            repaint();
        }
    }

    public int getMaximum() {
        return mMax;
    }

    public void setMaximum(int value) {
        value = Math.max(value, 0);
        if (mMax != value) {
            mMax = value;
            if (mMax == 0) {
                mAnimation = -1;
            }
            if (mCurrent > mMax) {
                mCurrent = mMax;
            }
            repaint();
        }
    }

    @Override
    public Dimension getMinimumSize() {
        if (isMinimumSizeSet()) {
            return super.getMinimumSize();
        }
        Insets insets = getInsets();
        return new Dimension(MIN_BAR_WIDTH + insets.left + insets.right,
                BAR_HEIGHT + insets.top + insets.bottom);
    }

    @Override
    public Dimension getPreferredSize() {
        if (isPreferredSizeSet()) {
            return super.getPreferredSize();
        }
        Insets insets = getInsets();
        return new Dimension(Math.max(MIN_BAR_WIDTH, 100) + insets.left + insets.right,
                BAR_HEIGHT + insets.top + insets.bottom);
    }

    @Override
    public Dimension getMaximumSize() {
        if (isMaximumSizeSet()) {
            return super.getMaximumSize();
        }
        Insets insets = getInsets();
        return new Dimension(20000, BAR_HEIGHT + insets.top + insets.bottom);
    }

    @Override
    protected void paintComponent(Graphics g) {
        Rectangle bounds = UIUtilities.getLocalInsetBounds(this);
        bounds.x++; // Pull both ends in 1 pixel so that the rounding doesn't get truncated
        bounds.width -= 2;
        Rectangle meter = new Rectangle(bounds.x, bounds.y, 0, bounds.height);
        if (mMax == 0) {
            meter.width = THUMB_WIDTH;
            if (mAnimation == -1) {
                mAnimation = System.currentTimeMillis();
            } else {
                int  max     = bounds.width - THUMB_WIDTH;
                long elapsed = (System.currentTimeMillis() - mAnimation) % (2 * ANIMATION_SPEED);
                if (elapsed >= ANIMATION_SPEED) {
                    elapsed = ANIMATION_SPEED - (elapsed - ANIMATION_SPEED);
                }
                meter.x = (int) (max * (elapsed / (double) ANIMATION_SPEED));
            }
        } else if (mCurrent > 0) {
            meter.width = (int) (bounds.width * (mCurrent / (double) mMax));
        }
        Graphics2D     gc    = GraphicsUtilities.prepare(g);
        RenderingHints hints = GraphicsUtilities.setMaximumQualityForGraphics(gc);
        gc.setColor(getBackground());
        gc.fillRoundRect(bounds.x, bounds.y, bounds.width, bounds.height, ARC, ARC);
        if (meter.width > 0) {
            gc.setColor(getForeground());
            gc.fillRoundRect(meter.x, meter.y, meter.width, meter.height, ARC, ARC);
        }
        gc.setColor(Colors.CONTROL_EDGE);
        gc.drawRoundRect(bounds.x, bounds.y, bounds.width, bounds.height - 1, ARC, ARC);
        if (meter.width > 0) {
            gc.drawRoundRect(meter.x, meter.y, meter.width, meter.height - 1, ARC, ARC);
        }
        gc.setRenderingHints(hints);
        if (mMax == 0) {
            EventQueue.invokeLater(this::repaint);
        }
    }
}
