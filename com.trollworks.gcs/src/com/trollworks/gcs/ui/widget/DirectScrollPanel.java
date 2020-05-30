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

import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.utility.Geometry;
import com.trollworks.gcs.utility.Log;

import java.awt.Adjustable;
import java.awt.AlphaComposite;
import java.awt.Color;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.LayoutManager;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.Transparency;
import java.awt.dnd.Autoscroll;
import java.awt.dnd.DragGestureEvent;
import java.awt.event.MouseEvent;
import java.awt.event.MouseWheelEvent;
import java.awt.event.MouseWheelListener;
import javax.swing.JPanel;
import javax.swing.JScrollBar;
import javax.swing.event.ChangeEvent;
import javax.swing.event.ChangeListener;

/** Provides a panel that manages scrolling internally to itself. */
public abstract class DirectScrollPanel extends JPanel implements Autoscroll, LayoutManager, ChangeListener, MouseWheelListener {
    /** The autoscroll margin. */
    public static final int                   AUTO_SCROLL_MARGIN = 10;
    private             JScrollBar            mHSB               = new JScrollBar(Adjustable.HORIZONTAL);
    private             JScrollBar            mVSB               = new JScrollBar(Adjustable.VERTICAL);
    private             int                   mHSBHeight         = mHSB.getPreferredSize().height;
    private             int                   mVSBWidth          = mVSB.getPreferredSize().width;
    private             Dimension             mHeaderSize        = new Dimension();
    private             Dimension             mContentSize       = new Dimension();
    private             Rectangle             mHeaderBounds      = new Rectangle();
    private             Rectangle             mContentBounds     = new Rectangle();
    private             DirectScrollPanelArea mLastAreaClicked;
    private             Rectangle             mDragClip;
    private             Component             mHSBLeftCorner;
    private             Component             mHSBRightCorner;
    private             Component             mVSBTopCorner;
    private             Component             mVSBBottomCorner;
    private             boolean               mDrawingDragImage;
    private             boolean               mInAutoscroll;

    /** Creates a new {@link DirectScrollPanel}. */
    public DirectScrollPanel() {
        setBackground(Color.WHITE);
        setLayout(this);
        mHSB.setVisible(false);
        mVSB.setVisible(false);
        mHSB.getModel().addChangeListener(this);
        mVSB.getModel().addChangeListener(this);
        addMouseWheelListener(this);
        add(mHSB);
        add(mVSB);
        setAutoscrolls(true);
    }

    /** @param amount The amount to set for the vertical scroll bar. */
    public void setVerticalUnitIncrement(int amount) {
        mVSB.setUnitIncrement(amount);
    }

    /** @param amount The amount to set for the horizontal scroll bar. */
    public void setHorizontalUnitIncrement(int amount) {
        mHSB.setUnitIncrement(amount);
    }

    /** @param amount The amount to set for both scroll bars. */
    public void setUnitIncrement(int amount) {
        mVSB.setUnitIncrement(amount);
        mHSB.setUnitIncrement(amount);
    }

    /** Repaints the header view. */
    public void repaintHeaderView() {
        repaint(mHeaderBounds);
    }

    /** Paints the header view immediately. */
    public void paintHeaderViewImmediately() {
        paintImmediately(mHeaderBounds);
    }

    /**
     * Repaints the header view.
     *
     * @param x      The area within the content view to repaint.
     * @param y      The area within the content view to repaint.
     * @param width  The area within the content view to repaint.
     * @param height The area within the content view to repaint.
     */
    public void repaintHeaderView(int x, int y, int width, int height) {
        Rectangle area = new Rectangle(x - (mHSB.getValue() - mHeaderBounds.x), y + mHeaderBounds.y, width, height);
        repaint(area.intersection(mHeaderBounds));
    }

    /**
     * Paints the header view immediately.
     *
     * @param x      The area within the content view to paint.
     * @param y      The area within the content view to paint.
     * @param width  The area within the content view to paint.
     * @param height The area within the content view to paint.
     */
    public void paintHeaderViewImmediately(int x, int y, int width, int height) {
        Rectangle area = new Rectangle(x - (mHSB.getValue() - mHeaderBounds.x), y + mHeaderBounds.y, width, height);
        paintImmediately(area.intersection(mHeaderBounds));
    }

    /**
     * Repaints the header view.
     *
     * @param bounds The area within the header view to repaint.
     */
    public void repaintHeaderView(Rectangle bounds) {
        repaintHeaderView(bounds.x, bounds.y, bounds.width, bounds.height);
    }

    /**
     * Paints the header view immediately.
     *
     * @param bounds The area within the header view to paint.
     */
    public void paintHeaderViewImmediately(Rectangle bounds) {
        paintHeaderViewImmediately(bounds.x, bounds.y, bounds.width, bounds.height);
    }

    /** Repaints the content view. */
    public void repaintContentView() {
        repaint(mContentBounds);
    }

    /** Paints the content view immediately. */
    public void paintContentViewImmediately() {
        paintImmediately(mContentBounds);
    }

    /**
     * Repaints the content view.
     *
     * @param x      The area within the content view to repaint.
     * @param y      The area within the content view to repaint.
     * @param width  The area within the content view to repaint.
     * @param height The area within the content view to repaint.
     */
    public void repaintContentView(int x, int y, int width, int height) {
        Rectangle area = new Rectangle(x - (mHSB.getValue() - mContentBounds.x), y - (mVSB.getValue() - mContentBounds.y), width, height);
        repaint(area.intersection(mContentBounds));
    }

    /**
     * Paints the content view immediately.
     *
     * @param x      The area within the content view to paint.
     * @param y      The area within the content view to paint.
     * @param width  The area within the content view to paint.
     * @param height The area within the content view to paint.
     */
    public void paintContentViewImmediately(int x, int y, int width, int height) {
        Rectangle area = new Rectangle(x - (mHSB.getValue() - mContentBounds.x), y - (mVSB.getValue() - mContentBounds.y), width, height);
        paintImmediately(area.intersection(mContentBounds));
    }

    /**
     * Repaints the content view.
     *
     * @param bounds The area within the content view to repaint.
     */
    public void repaintContentView(Rectangle bounds) {
        repaintContentView(bounds.x, bounds.y, bounds.width, bounds.height);
    }

    /**
     * Paints the content view immediately.
     *
     * @param bounds The area within the content view to paint.
     */
    public void paintContentViewImmediately(Rectangle bounds) {
        paintContentViewImmediately(bounds.x, bounds.y, bounds.width, bounds.height);
    }

    /** @return The currently visible header view bounds. */
    public Rectangle getHeaderViewBounds() {
        return new Rectangle(mHSB.getValue(), 0, mHeaderBounds.width, mHeaderBounds.height);
    }

    /** @return The currently visible content view bounds. */
    public Rectangle getContentViewBounds() {
        return new Rectangle(mHSB.getValue(), mVSB.getValue(), mContentBounds.width, mContentBounds.height);
    }

    @Override
    protected void paintComponent(Graphics g) {
        super.paintComponent(g);
        int hScroll = mHSB.getValue();
        if (mHeaderBounds.height > 0) {
            Graphics2D gc = (Graphics2D) g.create();
            try {
                gc.clipRect(mHeaderBounds.x, mHeaderBounds.y, mHeaderBounds.width, mHeaderBounds.height);
                gc.translate(mHeaderBounds.x - hScroll, mHeaderBounds.y);
                drawHeader(gc);
            } finally {
                gc.dispose();
            }
        }
        Graphics2D gc = (Graphics2D) g.create();
        try {
            gc.clipRect(mContentBounds.x, mContentBounds.y, mContentBounds.width, mContentBounds.height);
            gc.translate(mContentBounds.x - hScroll, mContentBounds.y - mVSB.getValue());
            drawContents(gc);
        } finally {
            gc.dispose();
        }
    }

    /**
     * Called to draw the header area.
     *
     * @param gc The {@link Graphics2D} context.
     */
    public void drawHeader(Graphics2D gc) {
        // Does nothing
    }

    /**
     * Called to draw the content area.
     *
     * @param gc The {@link Graphics2D} context.
     */
    public void drawContents(Graphics2D gc) {
        // Does nothing
    }

    /** Sets the header and content sizes to their preferred sizes. */
    public void pack() {
        setHeaderAndContentSize(getPreferredHeaderSize(), getPreferredContentSize());
        repaint();
    }

    /** @return The dimensions that would accommodate showing the entire header. */
    public abstract Dimension getPreferredHeaderSize();

    /** @return The dimensions that would accommodate showing the entire content. */
    public abstract Dimension getPreferredContentSize();

    /** @return The size of the header area. */
    public Dimension getHeaderSize() {
        return mHeaderSize;
    }

    /** @param size The size of the header area. */
    public void setHeaderSize(Dimension size) {
        mHeaderSize.setSize(size);
        layoutContainer(this);
    }

    /** @return The size of the content area. */
    public Dimension getContentSize() {
        return mContentSize;
    }

    /** @param size The size of the content area. */
    public void setContentSize(Dimension size) {
        mContentSize.setSize(size);
        layoutContainer(this);
    }

    /**
     * @param headerSize  The size of the header area.
     * @param contentSize The size of the content area.
     */
    public void setHeaderAndContentSize(Dimension headerSize, Dimension contentSize) {
        mHeaderSize.setSize(headerSize);
        mContentSize.setSize(contentSize);
        layoutContainer(this);
    }

    /** @return The minimum viewable size of the content area. */
    @SuppressWarnings("static-method")
    public Dimension getMinimumViewableContentSize() {
        return new Dimension(20, 20);
    }

    /** Synchronize any visible scroll bars to the contents. */
    public void syncScrollBars() {
        syncScrollBar(mHSB, mContentBounds.width, mContentSize.width);
        syncScrollBar(mVSB, mContentBounds.height, mContentSize.height);
    }

    private static void syncScrollBar(JScrollBar bar, int available, int needed) {
        bar.setValues(available < needed ? Math.min(bar.getValue(), needed - available) : 0, available, 0, needed);
    }

    /**
     * Scrolls the specified coordinates into view.
     *
     * @param where The location.
     */
    public void scrollContentIntoView(Point where) {
        scrollContentIntoView(where.x, where.y);
    }

    /**
     * Scrolls the specified coordinates into view.
     *
     * @param x The x-coordinate.
     * @param y The y-coordinate.
     */
    public void scrollContentIntoView(int x, int y) {
        scrollContentIntoView(x, y, 1, 1);
    }

    /**
     * Scrolls the specified bounds into view.
     *
     * @param bounds The area.
     */
    public void scrollContentIntoView(Rectangle bounds) {
        scrollContentIntoView(bounds.x, bounds.y, bounds.width, bounds.height);
    }

    /**
     * Scrolls the specified bounds into view.
     *
     * @param x      The x-coordinate.
     * @param y      The y-coordinate.
     * @param width  The width of the area.
     * @param height The height of the area.
     */
    public void scrollContentIntoView(int x, int y, int width, int height) {
        int   xadj         = 0;
        int   yadj         = 0;
        Point tl           = fromContentView(new Point(x, y));
        int   contentRight = mContentBounds.x + mContentBounds.width;
        int   right        = tl.x + width;
        if (tl.x >= mContentBounds.x || right <= contentRight) {
            if (tl.x < mContentBounds.x) {
                xadj = tl.x - mContentBounds.x;
            } else if (right > contentRight) {
                xadj = right - contentRight;
                tl.x -= xadj;
                if (tl.x < mContentBounds.x) {
                    xadj += tl.x - mContentBounds.x;
                }
            }
        }
        int contentBottom = mContentBounds.y + mContentBounds.height;
        int bottom        = tl.y + height;
        if (tl.y >= mContentBounds.y || bottom <= contentBottom) {
            if (tl.y < mContentBounds.y) {
                yadj = tl.y - mContentBounds.y;
            } else if (bottom > contentBottom) {
                yadj = bottom - contentBottom;
                tl.y -= yadj;
                if (tl.y < mContentBounds.y) {
                    yadj += tl.y - mContentBounds.y;
                }
            }
        }
        if (xadj != 0) {
            mHSB.setValue(mHSB.getValue() + xadj);
        }
        if (yadj != 0) {
            mVSB.setValue(mVSB.getValue() + yadj);
        }
    }

    /**
     * Scrolls the content view to the specified coordinates.
     *
     * @param x The x-coordinate.
     * @param y The y-coordinate.
     */
    public void scrollContentTo(int x, int y) {
        scrollContentToX(x);
        scrollContentToY(y);
    }

    /**
     * Scrolls the content view to the specified x-coordinate.
     *
     * @param x The x-coordinate.
     */
    public void scrollContentToX(int x) {
        if (x != mHSB.getValue()) {
            mHSB.setValue(x);
        }
    }

    /**
     * Scrolls the content view to the specified y-coordinate.
     *
     * @param y The y-coordinate.
     */
    public void scrollContentToY(int y) {
        if (y != mVSB.getValue()) {
            mVSB.setValue(y);
        }
    }

    @Override
    public void layoutContainer(Container parent) {
        Insets    insets      = getInsets();
        Dimension headerSize  = getHeaderSize();
        Dimension contentSize = getContentSize();
        int       totalWidth  = getWidth();
        int       width       = totalWidth - (insets.left + insets.right);
        int       totalHeight = getHeight();
        int       height      = totalHeight - (insets.top + headerSize.height + insets.bottom);
        boolean   needHSB     = mHSBLeftCorner != null || mHSBRightCorner != null || Math.max(headerSize.width, contentSize.width) > width;
        if (needHSB) {
            height -= mHSBHeight;
        }
        boolean needVSB = mVSBTopCorner != null || mVSBBottomCorner != null || contentSize.height > height;
        if (needVSB) {
            width -= mVSBWidth;
            if (!needHSB) {
                needHSB = Math.max(headerSize.width, contentSize.width) > width;
                if (needHSB) {
                    height -= mHSBHeight;
                }
            }
        }
        mHeaderBounds.x = insets.left;
        mHeaderBounds.y = insets.top;
        mHeaderBounds.width = Math.max(width, 0);
        mHeaderBounds.height = headerSize.height;
        mContentBounds.x = insets.left;
        mContentBounds.y = mHeaderBounds.y + mHeaderBounds.height;
        mContentBounds.width = mHeaderBounds.width;
        mContentBounds.height = Math.max(height, 0);
        mHSB.setVisible(needHSB);
        if (needHSB) {
            int x   = insets.left;
            int y   = totalHeight - (insets.bottom + mHSBHeight);
            int w   = totalWidth - (x + insets.right);
            int min = mHSBHeight * 3;
            if (needVSB) {
                w -= mVSBWidth;
            }
            if (mHSBLeftCorner != null) {
                int desired   = mHSBLeftCorner.getPreferredSize().width;
                int remaining = w - desired;
                if (remaining < min) {
                    desired = Math.max(desired - (min - remaining), 0);
                }
                mHSBLeftCorner.setBounds(x, y, desired, mHSBHeight);
                x += desired;
                w -= desired;
            }
            if (mHSBRightCorner != null) {
                int desired = mHSBRightCorner.getPreferredSize().width;
                if (needVSB) {
                    w += mVSBWidth;
                }
                int remaining = w - desired;
                if (remaining < min) {
                    desired = Math.max(desired - (min - remaining), 0);
                }
                w -= desired;
                mHSBRightCorner.setBounds(x + w, y, desired, mHSBHeight);
            }
            mHSB.setBounds(x, y, w, mHSBHeight);
        }
        mVSB.setVisible(needVSB);
        if (needVSB) {
            int x   = totalWidth - (insets.right + mVSBWidth);
            int y   = insets.top;
            int h   = totalHeight - (y + insets.bottom);
            int min = mVSBWidth * 3;
            if (needHSB) {
                h -= mHSBHeight;
            }
            if (mVSBTopCorner != null) {
                int desired   = mVSBTopCorner.getPreferredSize().height;
                int remaining = h - desired;
                if (remaining < min) {
                    desired = Math.max(desired - (min - remaining), 0);
                }
                mVSBTopCorner.setBounds(x, y, mVSBWidth, desired);
                y += desired;
                h -= desired;
            }
            if (mVSBBottomCorner != null) {
                int desired = mVSBBottomCorner.getPreferredSize().height;
                if (needHSB && mHSBRightCorner == null) {
                    h += mHSBHeight;
                }
                int remaining = h - desired;
                if (remaining < min) {
                    desired = Math.max(desired - (min - remaining), 0);
                }
                h -= desired;
                mVSBBottomCorner.setBounds(x, y + h, mVSBWidth, desired);
            }
            mVSB.setBounds(x, y, mVSBWidth, h);
        }
        syncScrollBars();
    }

    @Override
    public Dimension minimumLayoutSize(Container parent) {
        return getLayoutSize(getMinimumViewableContentSize());
    }

    @Override
    public Dimension preferredLayoutSize(Container parent) {
        return getLayoutSize(getContentSize());
    }

    private Dimension getLayoutSize(Dimension contentSize) {
        Insets    insets     = getInsets();
        Dimension headerSize = getHeaderSize();
        int       width      = insets.left + insets.right + Math.max(Math.max(headerSize.width, contentSize.width), mVSBWidth);
        int       height     = insets.top + insets.bottom + headerSize.height + Math.max(contentSize.height, mHSBHeight);
        return new Dimension(width, height);
    }

    @Override
    public void addLayoutComponent(String name, Component comp) {
        // Not used.
    }

    @Override
    public void removeLayoutComponent(Component comp) {
        // Not used.
    }

    @Override
    public void stateChanged(ChangeEvent event) {
        Object source = event.getSource();
        if (source == mHSB.getModel() || source == mVSB.getModel()) {
            repaint(mHeaderBounds);
            repaint(mContentBounds);
        }
    }

    /**
     * Transforms coordinates from this panel's space to the header view's space.
     *
     * @param where The coordinates to adjust.
     * @return The object that was passed in, adjusted as necessary.
     */
    public Point toHeaderView(Point where) {
        where.x += mHSB.getValue() - mHeaderBounds.x;
        where.y -= mHeaderBounds.y;
        return where;
    }

    /**
     * Transforms coordinates from the header view's space to this panel's space.
     *
     * @param where The coordinates to adjust.
     * @return The object that was passed in, adjusted as necessary.
     */
    public Point fromHeaderView(Point where) {
        where.x -= mHSB.getValue() - mHeaderBounds.x;
        where.y += mHeaderBounds.y;
        return where;
    }

    /**
     * Transforms coordinates from this panel's space to the content view's space.
     *
     * @param where The coordinates to adjust.
     * @return The object that was passed in, adjusted as necessary.
     */
    public Point toContentView(Point where) {
        where.x += mHSB.getValue() - mContentBounds.x;
        where.y += mVSB.getValue() - mContentBounds.y;
        return where;
    }

    /**
     * Transforms coordinates from the content view's space to this panel's space.
     *
     * @param where The coordinates to adjust.
     * @return The object that was passed in, adjusted as necessary.
     */
    public Point fromContentView(Point where) {
        where.x -= mHSB.getValue() - mContentBounds.x;
        where.y -= mVSB.getValue() - mContentBounds.y;
        return where;
    }

    @Override
    public void mouseWheelMoved(MouseWheelEvent event) {
        if (event.getScrollType() == MouseWheelEvent.WHEEL_UNIT_SCROLL) {
            int        amt = event.getUnitsToScroll();
            JScrollBar bar = event.isShiftDown() ? mHSB : mVSB;
            bar.setValue(bar.getValue() + amt * bar.getUnitIncrement());
        }
    }

    /**
     * Call to determine where the {@link Point} is.
     *
     * @param where The {@link Point} to check. If this method returns anything other than {@link
     *              DirectScrollPanelArea#NONE}, the passed-in {@link Point} will have its position
     *              adjusted for the area.
     * @return The {@link DirectScrollPanelArea} the {@link Point} occurred within.
     */
    public DirectScrollPanelArea checkAndConvertToArea(Point where) {
        DirectScrollPanelArea area;
        if (mContentBounds.contains(where)) {
            area = DirectScrollPanelArea.CONTENT;
        } else {
            area = mHeaderBounds.contains(where) ? DirectScrollPanelArea.HEADER : DirectScrollPanelArea.NONE;
        }
        area.convertPoint(this, where);
        return area;
    }

    @Override
    public void autoscroll(Point pt) {
        if (mLastAreaClicked == DirectScrollPanelArea.CONTENT) {
            if (pt.x < AUTO_SCROLL_MARGIN + mContentBounds.x) {
                pt.x -= AUTO_SCROLL_MARGIN;
            } else if (pt.x > mContentBounds.x + mContentBounds.width - AUTO_SCROLL_MARGIN) {
                pt.x += AUTO_SCROLL_MARGIN;
            }
            if (pt.y < AUTO_SCROLL_MARGIN + mContentBounds.y) {
                pt.y -= AUTO_SCROLL_MARGIN;
            } else if (pt.y > mContentBounds.y + mContentBounds.height - AUTO_SCROLL_MARGIN) {
                pt.y += AUTO_SCROLL_MARGIN;
            }
            scrollContentIntoView(toContentView(pt));
        } else if (mLastAreaClicked == DirectScrollPanelArea.HEADER) {
            if (pt.x < AUTO_SCROLL_MARGIN + mContentBounds.x) {
                pt.x -= AUTO_SCROLL_MARGIN;
            } else if (pt.x > mContentBounds.x + mContentBounds.width - AUTO_SCROLL_MARGIN) {
                pt.x += AUTO_SCROLL_MARGIN;
            }
            pt.y = mContentBounds.y;
            scrollContentIntoView(toContentView(pt));
        }
    }

    @Override
    public Insets getAutoscrollInsets() {
        Dimension size = getSize();
        return new Insets(AUTO_SCROLL_MARGIN + mContentBounds.y, AUTO_SCROLL_MARGIN + mContentBounds.x, AUTO_SCROLL_MARGIN + size.height - (mContentBounds.y + mContentBounds.height), AUTO_SCROLL_MARGIN + size.width - (mContentBounds.x + mContentBounds.width));
    }

    @Override
    protected void processMouseMotionEvent(MouseEvent event) {
        mInAutoscroll = getAutoscrolls() && event.getID() == MouseEvent.MOUSE_DRAGGED;
        super.processMouseMotionEvent(event);
        mInAutoscroll = false;
    }

    @Override
    public Rectangle getVisibleRect() {
        Rectangle bounds = super.getVisibleRect();
        if (mInAutoscroll) {
            bounds = Geometry.intersection(bounds, new Rectangle(mContentBounds.x + AUTO_SCROLL_MARGIN, mHeaderBounds.y, mContentBounds.width - AUTO_SCROLL_MARGIN * 2, mHeaderBounds.height + mContentBounds.height - AUTO_SCROLL_MARGIN));
        }
        return bounds;
    }

    @Override
    protected void processMouseEvent(MouseEvent event) {
        if (event.getID() == MouseEvent.MOUSE_PRESSED) {
            mLastAreaClicked = checkAndConvertToArea(event.getPoint());
        }
        super.processMouseEvent(event);
    }

    /**
     * @return A {@link Img} containing the currently visible header and contents (no scroll bars).
     *         If something goes wrong (unable to allocate an offscreen buffer, for instance), then
     *         {@code null} may be returned.
     */
    public Img createImage() {
        int        width     = mHeaderBounds.width;
        int        height    = mHeaderBounds.height + mContentBounds.height;
        Img        offscreen = null;
        Graphics2D g2d       = null;
        synchronized (getTreeLock()) {
            try {
                offscreen = Img.create(getGraphicsConfiguration(), width, height, Transparency.TRANSLUCENT);
                g2d = offscreen.getGraphics();
                Color saved = g2d.getBackground();
                g2d.setBackground(new Color(0, true));
                g2d.clearRect(0, 0, width, height);
                g2d.setBackground(saved);
                Insets insets = getInsets();
                g2d.translate(-insets.left, -insets.top);
                Rectangle clip = new Rectangle(0, 0, width, height);
                g2d.setClip(clip);
                paint(g2d);
            } catch (Exception exception) {
                Log.error(exception);
            } finally {
                if (g2d != null) {
                    g2d.dispose();
                }
            }
        }
        return offscreen;
    }

    /** @return Whether or not a drag image is currently being drawn. */
    protected boolean isDrawingDragImage() {
        return mDrawingDragImage;
    }

    /** @return The current drag clip area. May be {@code null}. */
    protected Rectangle getDragClip() {
        return mDragClip;
    }

    /** @param bounds The current drag clip area. May be {@code null}. */
    protected void setDragClip(Rectangle bounds) {
        mDragClip = bounds != null ? new Rectangle(bounds) : null;
    }

    @Override
    public boolean isOpaque() {
        return super.isOpaque() && !mDrawingDragImage;
    }

    /**
     * @param dragOrigin The drag origin point, as returned by {@link DragGestureEvent#getDragOrigin()}.
     * @return The drag image.
     */
    public Img createDragImage(Point dragOrigin) {
        Graphics2D gc = null;
        Img        off2;
        mDrawingDragImage = true;
        mDragClip = null;
        Img off1 = createImage();
        mDrawingDragImage = false;
        if (mDragClip == null) {
            mDragClip = new Rectangle(dragOrigin.x, dragOrigin.y, 1, 1);
        }
        try {
            off2 = Img.create(getGraphicsConfiguration(), mDragClip.width, mDragClip.height, Transparency.TRANSLUCENT);
            gc = off2.getGraphics();
            gc.setClip(new Rectangle(0, 0, mDragClip.width, mDragClip.height));
            gc.setBackground(new Color(0, true));
            gc.clearRect(0, 0, mDragClip.width, mDragClip.height);
            gc.setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, 0.5f));
            Point pt = toContentView(new Point(mHeaderBounds.x, mHeaderBounds.y));
            gc.drawImage(off1, -(mDragClip.x - pt.x), -(mDragClip.y - pt.y), this);
        } catch (Exception paintException) {
            off2 = null;
            mDragClip = new Rectangle(dragOrigin.x, dragOrigin.y, 1, 1);
            Log.error(paintException);
        } finally {
            if (gc != null) {
                gc.dispose();
            }
        }
        return off2 != null ? off2 : off1;
    }

    /** @return The {@link Component} to place on the left side of the horizontal scroll bar. */
    public Component getHSBLeftCorner() {
        return mHSBLeftCorner;
    }

    /**
     * @param component The {@link Component} to place on the left side of the horizontal scroll
     *                  bar.
     */
    public void setHSBLeftCorner(Component component) {
        if (mHSBLeftCorner != component) {
            if (mHSBLeftCorner != null) {
                remove(mHSBLeftCorner);
            }
            mHSBLeftCorner = component;
            if (mHSBLeftCorner != null) {
                add(mHSBLeftCorner);
            }
        }
    }

    /** @return The {@link Component} to place on the right side of the horizontal scroll bar. */
    public Component getHSBRightCorner() {
        return mHSBRightCorner;
    }

    /**
     * @param component The {@link Component} to place on the right side of the horizontal scroll
     *                  bar.
     */
    public void setHSBRightCorner(Component component) {
        if (mHSBRightCorner != component) {
            if (mHSBRightCorner != null) {
                remove(mHSBRightCorner);
            }
            mHSBRightCorner = component;
            if (mHSBRightCorner != null) {
                add(mHSBRightCorner);
            }
        }
    }

    /** @return The {@link Component} to place on the top side of the vertical scroll bar. */
    public Component getVSBTopCorner() {
        return mVSBTopCorner;
    }

    /**
     * @param component The {@link Component} to place on the top side of the vertical scroll bar.
     */
    public void setVSBTopCorner(Component component) {
        if (mVSBTopCorner != component) {
            if (mVSBTopCorner != null) {
                remove(mVSBTopCorner);
            }
            mVSBTopCorner = component;
            if (mVSBTopCorner != null) {
                add(mVSBTopCorner);
            }
        }
    }

    /** @return The {@link Component} to place on the bottom side of the vertical scroll bar. */
    public Component getVSBBottomCorner() {
        return mVSBBottomCorner;
    }

    /**
     * @param component The {@link Component} to place on the bottom side of the vertical scroll
     *                  bar.
     */
    public void setVSBBottomCorner(Component component) {
        if (mVSBBottomCorner != component) {
            if (mVSBBottomCorner != null) {
                remove(mVSBBottomCorner);
            }
            mVSBBottomCorner = component;
            if (mVSBBottomCorner != null) {
                add(mVSBBottomCorner);
            }
        }
    }
}
