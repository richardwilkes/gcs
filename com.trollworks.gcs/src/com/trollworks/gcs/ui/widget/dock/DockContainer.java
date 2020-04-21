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

package com.trollworks.gcs.ui.widget.dock;

import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.ui.UIUtilities;

import java.awt.Color;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Insets;
import java.awt.KeyboardFocusManager;
import java.awt.LayoutManager;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JPanel;

/**
 * All {@link Dockable}s are wrapped in a {@link DockContainer} when placed within a {@link Dock}.
 */
public class DockContainer extends JPanel implements DockLayoutNode, LayoutManager {
    private Dock           mDock;
    private DockHeader     mHeader;
    private List<Dockable> mDockables = new ArrayList<>();
    private int            mCurrent;
    private boolean        mActive;

    /**
     * Creates a new {@link DockContainer} for the specified {@link Dockable}.
     *
     * @param dock     The {@link Dock} that owns this {@link DockContainer}.
     * @param dockable The {@link Dockable} to wrap.
     */
    public DockContainer(Dock dock, Dockable dockable) {
        mDock = dock;
        setLayout(this);
        setOpaque(true);
        setBackground(Color.WHITE);
        mHeader = new DockHeader(this);
        add(mHeader);
        add(dockable);
        mDockables.add(dockable);
        mHeader.addTab(dockable, 0);
        setMinimumSize(new Dimension(0, 0));
    }

    /** @return The {@link Dock} this {@link DockContainer} resides in. */
    public Dock getDock() {
        return mDock;
    }

    /** @return The current list of {@link Dockable}s in this {@link DockContainer}. */
    public List<Dockable> getDockables() {
        return mDockables;
    }

    /** @param dockable The {@link Dockable} whose title and icon should be updated. */
    public void updateTitle(Dockable dockable) {
        int index = mDockables.indexOf(dockable);
        if (index != -1) {
            mHeader.updateTitle(index);
        }
    }

    /** @param dockable The {@link Dockable} to stack into this {@link DockContainer}. */
    public void stack(Dockable dockable) {
        stack(dockable, -1);
    }

    /**
     * @param dockable The {@link Dockable} to stack into this {@link DockContainer}.
     * @param index    The position within this container to place it. Values out of range will
     *                 result in the {@link Dockable} being placed at the end.
     */
    public void stack(Dockable dockable, int index) {
        DockContainer dc = dockable.getDockContainer();
        if (dc != null) {
            if (dc == this && mDockables.size() == 1) {
                setCurrentDockable(dockable);
                acquireFocus();
                return;
            }
            dc.close(dockable);
        }
        if (index < 0 || index >= mDockables.size()) {
            mDockables.add(dockable);
        } else {
            mDockables.add(index, dockable);
        }
        add(dockable);
        mHeader.addTab(dockable, mDockables.indexOf(dockable));
        setCurrentDockable(dockable);
        acquireFocus();
    }

    /** Transfers focus to this container if it doesn't already have the focus. */
    public void acquireFocus() {
        Component focusOwner = KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner();
        Component content    = getCurrentDockable();
        while (focusOwner != null && focusOwner != content) {
            focusOwner = focusOwner.getParent();
        }
        if (focusOwner == null) {
            EventQueue.invokeLater(() -> transferFocus());
        }
    }

    /** @return The {@link DockHeader} for this {@link DockContainer}. */
    public DockHeader getHeader() {
        return mHeader;
    }

    /**
     * Calls the owning {@link Dock}'s {@link Dock#maximize(DockContainer)} method with this {@link
     * DockContainer} as the argument.
     */
    public void maximize() {
        mDock.maximize(this);
    }

    /** Calls the owning {@link Dock}'s {@link Dock#restore()} method. */
    public void restore() {
        mDock.restore();
    }

    /** @return The current tab index. */
    public int getCurrentTabIndex() {
        return mCurrent;
    }

    /** @return The current {@link Dockable}. */
    public Dockable getCurrentDockable() {
        return mCurrent >= 0 && mCurrent < mDockables.size() ? mDockables.get(mCurrent) : null;
    }

    /** @param dockable The {@link Dockable} to make current. */
    public void setCurrentDockable(Dockable dockable) {
        int index = mDockables.indexOf(dockable);
        if (index != -1) {
            int wasCurrent = mCurrent;
            mCurrent = index;
            for (Dockable one : mDockables) {
                one.setVisible(dockable == one);
            }
            mHeader.revalidate();
            repaint();
            acquireFocus();
            if (mActive && wasCurrent != mCurrent) {
                dockable.activated();
            }
        }
    }

    protected void setCurrentDockable(Dockable dockable, Component focusOn) {
        int index = mDockables.indexOf(dockable);
        if (index != -1) {
            int wasCurrent = mCurrent;
            mCurrent = index;
            for (Dockable one : mDockables) {
                one.setVisible(dockable == one);
            }
            mHeader.revalidate();
            repaint();
            focusOn.requestFocus();
            if (mActive && wasCurrent != mCurrent) {
                dockable.activated();
            }
        }
    }

    @Override
    public String toString() {
        StringBuilder buffer = new StringBuilder();
        if (getParent() == null) {
            buffer.append("FLOATING ");
        }
        buffer.append("Dock Container [x:");
        buffer.append(getX());
        buffer.append(" y:");
        buffer.append(getY());
        buffer.append(" w:");
        buffer.append(getWidth());
        buffer.append(" h:");
        buffer.append(getHeight());
        int count = mDockables.size();
        for (int i = 0; i < count; i++) {
            buffer.append(' ');
            if (i == mCurrent) {
                buffer.append('*');
            }
            buffer.append('d');
            buffer.append(i);
            buffer.append(':');
            buffer.append(mDockables.get(i).getTitle());
        }
        buffer.append("]");
        return buffer.toString();
    }

    /**
     * Attempt to close a {@link Dockable} within this {@link DockContainer}. This only has an
     * affect if the {@link Dockable} is contained by this {@link DockContainer} and implements the
     * {@link CloseHandler} interface. Note that the {@link CloseHandler} must call this {@link
     * DockContainer}'s {@link #close(Dockable)} method to actually close the tab.
     */
    public void attemptClose(Dockable dockable) {
        if (dockable instanceof CloseHandler) {
            if (mDockables.contains(dockable)) {
                CloseHandler closeable = (CloseHandler) dockable;
                if (closeable.mayAttemptClose()) {
                    closeable.attemptClose();
                }
            }
        }
    }

    /**
     * Closes the specified {@link Dockable}. If the last {@link Dockable} within this {@link
     * DockContainer} is closed, then this {@link DockContainer} is also removed from the {@link
     * Dock}.
     */
    public void close(Dockable dockable) {
        int index = mDockables.indexOf(dockable);
        if (index != -1) {
            remove(dockable);
            mDockables.remove(dockable);
            mHeader.close(dockable);
            if (mDockables.isEmpty()) {
                restore();
                mDock.remove(this);
                mDock.revalidate();
                mDock.repaint();
                mDock = null;
            } else {
                if (index > 0) {
                    index--;
                }
                setCurrentDockable(mDockables.get(index));
            }
        }
    }

    /**
     * @return {@code true} if this {@link DockContainer} or one of its children has the keyboard
     *         focus.
     */
    public boolean isActive() {
        return mActive;
    }

    /** Called by the {@link Dock} to update the active highlight. */
    void updateActiveHighlight() {
        boolean active = UIUtilities.getSelfOrAncestorOfType(KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner(), DockContainer.class) == this;
        if (mActive != active) {
            mActive = active;
            mHeader.repaint();
            if (mActive) {
                Dockable dockable = getCurrentDockable();
                if (dockable != null) {
                    dockable.activated();
                }
            }
        }
    }

    @Override
    public void addLayoutComponent(String name, Component comp) {
        // Unused
    }

    @Override
    public void removeLayoutComponent(Component comp) {
        // Unused
    }

    @Override
    public Dimension preferredLayoutSize(Container parent) {
        Dimension size   = mHeader.getPreferredSize();
        int       width  = size.width;
        int       height = size.height;
        if (!mDockables.isEmpty()) {
            size = getCurrentDockable().getPreferredSize();
            if (width < size.width) {
                width = size.width;
            }
            height += size.height;
        }
        Insets insets = parent.getInsets();
        return new Dimension(insets.left + width + insets.right, insets.top + height + insets.bottom);
    }

    @Override
    public Dimension minimumLayoutSize(Container parent) {
        Dimension size    = mHeader.getMinimumSize();
        int       width   = size.width;
        int       height  = size.height;
        Dockable  current = getCurrentDockable();
        if (current != null) {
            size = current.getMinimumSize();
            if (width < size.width) {
                width = size.width;
            }
            height += size.height;
        }
        Insets insets = parent.getInsets();
        return new Dimension(insets.left + width + insets.right, insets.top + height + insets.bottom);
    }

    @Override
    public void layoutContainer(Container parent) {
        Insets insets = parent.getInsets();
        int    height = mHeader.getPreferredSize().height;
        int    width  = parent.getWidth() - (insets.left + insets.right);
        mHeader.setBounds(insets.left, insets.top, width, height);
        Dockable current = getCurrentDockable();
        if (current != null) {
            int remaining = getHeight() - (insets.top + height);
            if (remaining < 0) {
                remaining = 0;
            }
            current.setBounds(insets.left, insets.top + height, width, remaining);
        }
    }
}
