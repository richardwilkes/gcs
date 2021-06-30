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
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.RenderingHints;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import java.util.ArrayList;
import java.util.List;
import javax.swing.event.ChangeEvent;
import javax.swing.event.ChangeListener;

public class Scrollbar extends Panel implements MouseListener, MouseMotionListener {
    public static final  int MINIMUM_SIZE  = 16;
    private static final int CORNER_RADIUS = 8;
    private static final int MIN_THUMB     = 4;
    private static final int THUMB_INDENT  = 3;

    private final List<ChangeListener> mChangeListeners;
    private final ChangeEvent          mChangeEvent;
    private       int                  mValue;
    private       int                  mExtent;
    private       int                  mMax;
    private       int                  mDragOffset;
    private       boolean              mVertical;
    private       boolean              mOverThumb;
    private       boolean              mTrackingThumb;

    public Scrollbar(boolean vertical) {
        super(null, false);
        mChangeListeners = new ArrayList<>();
        mChangeEvent = new ChangeEvent(this);
        mVertical = vertical;
        addMouseListener(this);
        addMouseMotionListener(this);
    }

    public final boolean isVertical() {
        return mVertical;
    }

    public final boolean isHorizontal() {
        return !mVertical;
    }

    public final void addChangeListener(ChangeListener listener) {
        synchronized (mChangeListeners) {
            mChangeListeners.add(listener);
        }
    }

    public final void removeChangeListener(ChangeListener listener) {
        synchronized (mChangeListeners) {
            mChangeListeners.remove(listener);
        }
    }

    private void notifyChangeListeners() {
        List<ChangeListener> listeners = null;
        synchronized (mChangeListeners) {
            if (!mChangeListeners.isEmpty()) {
                listeners = new ArrayList<>(mChangeListeners);
            }
        }
        if (listeners != null) {
            for (ChangeListener listener : listeners) {
                listener.stateChanged(mChangeEvent);
            }
        }
    }

    @Override
    public final Dimension getPreferredSize() {
        if (isPreferredSizeSet()) {
            return super.getPreferredSize();
        }
        return new Dimension(MINIMUM_SIZE, MINIMUM_SIZE);
    }

    @Override
    public final Dimension getMinimumSize() {
        if (isMinimumSizeSet()) {
            return super.getMinimumSize();
        }
        return new Dimension(MINIMUM_SIZE, MINIMUM_SIZE);
    }

    public int getValue() {
        return mValue;
    }

    public int getExtent() {
        return mExtent;
    }

    public int getMax() {
        return mMax;
    }

    public final void setRange(int value, int extent, int max) {
        if (value < 0) {
            value = 0;
        }
        if (max < 0) {
            max = 0;
        }
        if (extent > max) {
            extent = max;
        }
        if (value + extent > max) {
            value = max - extent;
        }
        if (mValue != value || mExtent != extent || mMax != max) {
            mValue = value;
            mExtent = extent;
            mMax = max;
            repaint();
            notifyChangeListeners();
        }
    }

    public Rectangle getThumbBounds() {
        if (mMax == 0) {
            return new Rectangle();
        }
        Rectangle bounds = UIUtilities.getLocalInsetBounds(this);
        if (mVertical) {
            int start = (int) (bounds.height * (mValue / (double) mMax));
            int size  = (int) (bounds.height * (mExtent / (double) mMax));
            if (size < MIN_THUMB) {
                size = MIN_THUMB;
            }
            return new Rectangle(THUMB_INDENT, start, bounds.width - 2 * THUMB_INDENT, size);
        }
        int start = (int) (bounds.width * (mValue / (double) mMax));
        int size  = (int) (bounds.width * (mExtent / (double) mMax));
        if (size < MIN_THUMB) {
            size = MIN_THUMB;
        }
        return new Rectangle(start, THUMB_INDENT, size, bounds.height - 2 * THUMB_INDENT);
    }

    @Override
    protected void paintComponent(Graphics g) {
        Rectangle bounds = getThumbBounds();
        if (bounds.width > 0 && bounds.height > 0) {
            Graphics2D     gc    = GraphicsUtilities.prepare(g);
            RenderingHints hints = GraphicsUtilities.setMaximumQualityForGraphics(gc);
            gc.setColor(mOverThumb ? Colors.SCROLL_ROLLOVER : Colors.SCROLL);
            gc.fillRoundRect(bounds.x, bounds.y, bounds.width, bounds.height, CORNER_RADIUS, CORNER_RADIUS);
            gc.setColor(Colors.SCROLL_EDGE);
            gc.drawRoundRect(bounds.x, bounds.y, bounds.width, bounds.height, CORNER_RADIUS, CORNER_RADIUS);
            gc.setRenderingHints(hints);
        }
    }

    private void checkForOverThumb(Point pt) {
        boolean wasOver = mOverThumb;
        mOverThumb = getThumbBounds().contains(pt);
        if (mOverThumb != wasOver) {
            repaint();
        }
    }

    @Override
    public void mouseClicked(MouseEvent event) {
        // Unused
    }

    @Override
    public void mousePressed(MouseEvent event) {
        Point     pt = event.getPoint();
        Rectangle tb = getThumbBounds();
        if (!tb.contains(pt)) {
            mDragOffset = 0;
            adjustValueForPoint(pt);
            tb = getThumbBounds();
        }
        mDragOffset = mVertical ? tb.y - pt.y : tb.x - pt.x;
        mOverThumb = true;
        mTrackingThumb = true;
        repaint();
    }

    @Override
    public void mouseReleased(MouseEvent event) {
        mTrackingThumb = false;
        checkForOverThumb(event.getPoint());
    }

    @Override
    public void mouseEntered(MouseEvent event) {
        if (!mTrackingThumb) {
            checkForOverThumb(event.getPoint());
        }
    }

    @Override
    public void mouseExited(MouseEvent event) {
        if (!mTrackingThumb && mOverThumb) {
            mOverThumb = false;
            repaint();
        }
    }

    @Override
    public void mouseDragged(MouseEvent event) {
        adjustValueForPoint(event.getPoint());
        repaint();
    }

    @Override
    public void mouseMoved(MouseEvent event) {
        checkForOverThumb(event.getPoint());
    }

    private void adjustValueForPoint(Point pt) {
        Rectangle bounds = UIUtilities.getLocalInsetBounds(this);
        Rectangle tb     = getThumbBounds();
        int       pos;
        double    max;
        if (mVertical) {
            pos = pt.y;
            max = bounds.height - tb.height;
        } else {
            pos = pt.x;
            max = bounds.width - tb.width;
        }
        double range = mMax - mExtent;
        setRange(range == 0 ? 0 : (int) (0.5 + (mMax - mExtent) * (pos + mDragOffset) / max), mExtent, mMax);
    }
}
