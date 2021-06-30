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
import com.trollworks.gcs.ui.UIUtilities;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.event.MouseWheelEvent;
import java.awt.event.MouseWheelListener;
import javax.swing.JLayeredPane;
import javax.swing.JViewport;
import javax.swing.Scrollable;
import javax.swing.SwingConstants;
import javax.swing.event.ChangeEvent;
import javax.swing.event.ChangeListener;

public class ScrollPanel extends JLayeredPane implements LayoutManager, ChangeListener, MouseWheelListener {
    private static final int     MINIMUM_SIZE    = 4;
    private static final int     CORNER_RADIUS   = 8;
    private static final int     BAR_WIDTH       = 10;
    private static final int     BAR_INSET       = 13;
    private static final Integer SCROLLBAR_LAYER = Integer.valueOf(50);

    private JViewport mViewport;
    private JViewport mRowHeader;
    private JViewport mColumnHeader;
    private Scrollbar mVerticalScrollbar;
    private Scrollbar mHorizontalScrollbar;

    public ScrollPanel(Component view) {
        super.setLayout(this);
        mViewport = new JViewport();
        mViewport.setBackground(Colors.BACKGROUND);
        add(mViewport, DEFAULT_LAYER);
        mVerticalScrollbar = new Scrollbar(true);
        add(mVerticalScrollbar, SCROLLBAR_LAYER);
        mHorizontalScrollbar = new Scrollbar(false);
        add(mHorizontalScrollbar, SCROLLBAR_LAYER);
        mViewport.addChangeListener(this);
        mVerticalScrollbar.addChangeListener(this);
        mHorizontalScrollbar.addChangeListener(this);
        addMouseWheelListener(this);
        if (view != null) {
            setViewportView(view);
        }
    }

    public ScrollPanel(Component header, Component view) {
        this(view);
        setColumnHeaderView(header);
    }

    public final void setLayout(LayoutManager layout) {
        // Ignore. We set out layout once during intialization via a super call.
    }

    @Override
    public boolean isValidateRoot() {
        return true;
    }

    public JViewport getRowHeader() {
        return mRowHeader;
    }

    public void setRowHeader(JViewport header) {
        if (mRowHeader != header) {
            if (mRowHeader != null) {
                remove(mRowHeader);
            }
            mRowHeader = header;
            if (mRowHeader != null) {
                add(mRowHeader, DEFAULT_LAYER);
            }
            revalidate();
            repaint();
        }
    }

    public void setRowHeaderView(Component view) {
        if (mRowHeader == null) {
            JViewport viewport = new JViewport();
            viewport.setBackground(Colors.BACKGROUND);
            add(viewport, DEFAULT_LAYER);
        }
        mRowHeader.setView(view);
    }

    public JViewport getColumnHeader() {
        return mColumnHeader;
    }

    public void setColumnHeader(JViewport header) {
        if (mColumnHeader != header) {
            if (mColumnHeader != null) {
                remove(mColumnHeader);
            }
            mColumnHeader = header;
            if (mColumnHeader != null) {
                add(mColumnHeader, DEFAULT_LAYER);
            }
            revalidate();
            repaint();
        }
    }

    public void setColumnHeaderView(Component view) {
        if (mColumnHeader == null) {
            JViewport viewport = new JViewport();
            viewport.setBackground(Colors.BACKGROUND);
            setColumnHeader(viewport);
        }
        mColumnHeader.setView(view);
    }

    public final JViewport getViewport() {
        return mViewport;
    }

    public void setViewportView(Component view) {
        mViewport.setView(view);
        revalidate();
        repaint();
    }

    public int getUnitIncrement(boolean vertical, boolean upOrLeft) {
        Dimension extent = mViewport.getExtentSize();
        Component view   = mViewport.getView();
        if (view instanceof Scrollable) {
            Point pos = mViewport.getViewPosition();
            return ((Scrollable) view).getScrollableUnitIncrement(new Rectangle(pos.x, pos.y,
                    extent.width, extent.height), vertical ? SwingConstants.VERTICAL :
                    SwingConstants.HORIZONTAL, upOrLeft ? -1 : 1);
        }
        return 1;
    }

    public int getBlockIncrement(boolean vertical, boolean upOrLeft) {
        Dimension extent = mViewport.getExtentSize();
        Component view   = mViewport.getView();
        if (view instanceof Scrollable) {
            Point pos = mViewport.getViewPosition();
            return ((Scrollable) view).getScrollableBlockIncrement(new Rectangle(pos.x, pos.y,
                    extent.width, extent.height), vertical ? SwingConstants.VERTICAL :
                    SwingConstants.HORIZONTAL, upOrLeft ? -1 : 1);
        }
        return vertical ? extent.height : extent.width;
    }

    @Override
    public final void addLayoutComponent(String name, Component comp) {
        // Unused
    }

    @Override
    public final void removeLayoutComponent(Component comp) {
        // Unused
    }

    @Override
    public final Dimension preferredLayoutSize(Container parent) {
        Dimension size   = mViewport.getPreferredSize();
        Insets    insets = getInsets();
        int       w      = Math.max(size.width, Scrollbar.MINIMUM_SIZE) + insets.left + insets.right;
        int       h      = Math.max(size.height, Scrollbar.MINIMUM_SIZE) + insets.top + insets.bottom;
        if ((mRowHeader != null) && mRowHeader.isVisible()) {
            w += mRowHeader.getPreferredSize().width;
        }
        if ((mColumnHeader != null) && mColumnHeader.isVisible()) {
            h += mColumnHeader.getPreferredSize().height;
        }
        return new Dimension(w, h);
    }

    @Override
    public final Dimension minimumLayoutSize(Container parent) {
        Dimension size   = mViewport.getMinimumSize();
        Insets    insets = getInsets();
        int       w      = Math.max(size.width, Scrollbar.MINIMUM_SIZE) + insets.left + insets.right;
        int       h      = Math.max(size.height, Scrollbar.MINIMUM_SIZE) + insets.top + insets.bottom;
        if ((mRowHeader != null) && mRowHeader.isVisible()) {
            size = mRowHeader.getMinimumSize();
            w += size.width;
            h = Math.max(h, size.height);
        }
        if ((mColumnHeader != null) && mColumnHeader.isVisible()) {
            size = mColumnHeader.getMinimumSize();
            w = Math.max(w, size.width);
            h += size.height;
        }
        return new Dimension(w, h);
    }

    @Override
    public final void layoutContainer(Container parent) {
        Rectangle bounds       = UIUtilities.getLocalInsetBounds(this);
        Rectangle columnBounds = new Rectangle(0, bounds.y, 0, 0);
        if (mColumnHeader != null && mColumnHeader.isVisible()) {
            int height = Math.min(bounds.height, mColumnHeader.getPreferredSize().height);
            columnBounds.height = height;
            bounds.y += height;
            bounds.height -= height;
        }
        Rectangle rowBounds = new Rectangle(bounds.x, bounds.y, 0, bounds.height);
        if (mRowHeader != null && mRowHeader.isVisible()) {
            int width = Math.min(bounds.width, mRowHeader.getPreferredSize().width);
            rowBounds.width = width;
            mRowHeader.setBounds(rowBounds);
            bounds.x += width;
            bounds.width -= width;
        }
        if (mColumnHeader != null) {
            columnBounds.width = bounds.width;
            columnBounds.x = bounds.x;
            mColumnHeader.setBounds(columnBounds);
        }
        Dimension viewSize     = mViewport.getViewSize();
        boolean   isVBarNeeded = bounds.height < viewSize.height;
        boolean   isHBarNeeded = bounds.width < viewSize.width;
        boolean   bothNeeded   = isVBarNeeded && isHBarNeeded;
        int       height       = isVBarNeeded ? bounds.height : 0;
        int       width        = isHBarNeeded ? bounds.width : 0;

        mVerticalScrollbar.setRange(mVerticalScrollbar.getValue(), bounds.height, viewSize.height);
        mVerticalScrollbar.setBounds(bounds.x + bounds.width - Scrollbar.MINIMUM_SIZE, bounds.y,
                Scrollbar.MINIMUM_SIZE, bothNeeded ? height - Scrollbar.MINIMUM_SIZE : height);
        mHorizontalScrollbar.setRange(mHorizontalScrollbar.getValue(), bounds.width, viewSize.width);
        mHorizontalScrollbar.setBounds(bounds.x, bounds.y + bounds.height - Scrollbar.MINIMUM_SIZE,
                bothNeeded ? width - Scrollbar.MINIMUM_SIZE : width, Scrollbar.MINIMUM_SIZE);
        mViewport.setBounds(bounds);
    }

    @Override
    public void stateChanged(ChangeEvent event) {
        Object source = event.getSource();
        if (source == mVerticalScrollbar || source == mHorizontalScrollbar) {
            mViewport.setViewPosition(new Point(mHorizontalScrollbar.getValue(), mVerticalScrollbar.getValue()));
        } else if (source == mViewport) {
            Point viewPosition = mViewport.getViewPosition();
            if (mRowHeader != null) {
                mRowHeader.setViewPosition(new Point(0, viewPosition.y));
            }
            if (mColumnHeader != null) {
                mColumnHeader.setViewPosition(new Point(viewPosition.x, 0));
            }
            repaint();
        }
    }

    @Override
    public void mouseWheelMoved(MouseWheelEvent event) {
        int wheelRotation = event.getWheelRotation();
        if (wheelRotation != 0) {
            boolean isVertical = !event.isShiftDown();
            boolean upOrLeft   = wheelRotation < 0;
            int increment = event.getScrollType() == MouseWheelEvent.WHEEL_UNIT_SCROLL ?
                    getUnitIncrement(isVertical, upOrLeft) * event.getUnitsToScroll() :
                    getBlockIncrement(isVertical, upOrLeft) * wheelRotation;
            if (increment != 0) {
                Scrollbar bar = isVertical ? mVerticalScrollbar : mHorizontalScrollbar;
                bar.setRange(bar.getValue() + increment, bar.getExtent(), bar.getMax());
            }
        }
    }
}
