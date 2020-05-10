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

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.LayoutManager;

/** Provides all {@link Dock} layout management. */
public class DockLayout implements DockLayoutNode, LayoutManager {
    private DockLayout       mParent;
    private DockLayoutNode[] mChildren        = new DockLayoutNode[2];
    private int              mX;
    private int              mY;
    private int              mWidth;
    private int              mHeight;
    private int              mDividerPosition = -1;
    private boolean          mHorizontal;

    /** @param processor A processor to execute for each {@link DockContainer}. */
    public void forEachDockContainer(DockContainerProcessor processor) {
        for (DockLayoutNode child : mChildren) {
            if (child instanceof DockContainer) {
                processor.processDockContainer((DockContainer) child);
            } else if (child instanceof DockLayout) {
                ((DockLayout) child).forEachDockContainer(processor);
            }
        }
    }

    /** @return The {@link DockContainer} with the current keyboard focus, or {@code null}. */
    public DockContainer getFocusedDockContainer() {
        return getRootLayout().getFocusedDockContainerInternal();
    }

    private DockContainer getFocusedDockContainerInternal() {
        for (DockLayoutNode child : mChildren) {
            if (child instanceof DockContainer) {
                DockContainer dc = (DockContainer) child;
                if (dc.isActive()) {
                    return dc;
                }
            } else if (child instanceof DockLayout) {
                DockContainer dc = ((DockLayout) child).getFocusedDockContainerInternal();
                if (dc != null) {
                    return dc;
                }
            }
        }
        return null;
    }

    /** @return The root {@link DockLayout}, which may be this object. */
    public DockLayout getRootLayout() {
        DockLayout root = this;
        while (root.mParent != null) {
            root = root.mParent;
        }
        return root;
    }

    /** @return The {@link Dock} this {@link DockLayout} is associated with. */
    public Dock getDock() {
        return getRootLayout().getDockInternal();
    }

    private Dock getDockInternal() {
        for (DockLayoutNode child : mChildren) {
            if (child instanceof DockContainer) {
                return (Dock) ((DockContainer) child).getParent();
            } else if (child instanceof DockLayout) {
                Dock dock = ((DockLayout) child).getDockInternal();
                if (dock != null) {
                    return dock;
                }
            }
        }
        return null;
    }

    /**
     * @param dc The {@link DockContainer} to search for.
     * @return The {@link DockLayout} that contains the {@link DockContainer}, or {@code null} if it
     *         is not present. Note that this method will always start at the root and work its way
     *         down, even if called on a sub-node.
     */
    public DockLayout findLayout(DockContainer dc) {
        return getRootLayout().findLayoutInternal(dc);
    }

    private DockLayout findLayoutInternal(DockContainer dc) {
        for (DockLayoutNode child : mChildren) {
            if (child instanceof DockContainer) {
                if (child == dc) {
                    return this;
                }
            } else if (child instanceof DockLayout) {
                DockLayout layout = ((DockLayout) child).findLayoutInternal(dc);
                if (layout != null) {
                    return layout;
                }
            }
        }
        return null;
    }

    /**
     * @param node The {@link DockLayoutNode} to look for.
     * @return {@code true} if the node is this {@link DockLayout} or one of its descendants.
     */
    public boolean contains(DockLayoutNode node) {
        if (node == this) {
            return true;
        }
        for (DockLayoutNode child : mChildren) {
            if (child == node) {
                return true;
            }
            if (child instanceof DockLayout) {
                if (((DockLayout) child).contains(node)) {
                    return true;
                }
            }
        }
        return false;
    }

    /**
     * Docks a {@link DockContainer} within this {@link DockLayout}. If the {@link DockContainer}
     * already exists in this {@link DockLayout}, it will be moved to the new location.
     *
     * @param dc                       The {@link DockContainer} to install into this {@link
     *                                 DockLayout}.
     * @param target                   The target {@link DockLayoutNode}.
     * @param locationRelativeToTarget The location relative to the target to install the {@link
     *                                 DockContainer}.
     */
    void dock(DockContainer dc, DockLayoutNode target, DockLocation locationRelativeToTarget) {
        // Does the container already exist in our hierarchy?
        DockLayout existingLayout = findLayout(dc);
        if (existingLayout != null) {
            // Yes. Is it the same layout?
            DockLayout targetLayout;
            if (target instanceof DockLayout) {
                targetLayout = (DockLayout) target;
            } else if (target instanceof DockContainer) {
                targetLayout = findLayout((DockContainer) target);
            } else {
                targetLayout = null;
            }
            if (targetLayout == existingLayout) {
                // Yes. Reposition the target within this layout.
                int[] order = locationRelativeToTarget.getOrder();
                if (targetLayout.mChildren[order[0]] != dc) {
                    targetLayout.mChildren[order[1]] = targetLayout.mChildren[order[0]];
                    targetLayout.mChildren[order[0]] = dc;
                }
                targetLayout.mHorizontal = locationRelativeToTarget.isHorizontal();
                return;
            }
            // Not in the same layout. Remove the container from the hierarchy so we can re-add it.
            existingLayout.remove(dc);
        }

        if (target instanceof DockLayout) {
            ((DockLayout) target).dock(dc, locationRelativeToTarget);
        } else if (target instanceof DockContainer) {
            DockContainer tdc    = (DockContainer) target;
            DockLayout    layout = findLayout(tdc);
            layout.dockWithContainer(dc, tdc, locationRelativeToTarget);
        }
    }

    private void dockWithContainer(DockContainer dc, DockLayoutNode target, DockLocation locationRelativeToTarget) {
        boolean horizontal = locationRelativeToTarget.isHorizontal();
        int[]   order      = locationRelativeToTarget.getOrder();
        if (mChildren[order[0]] != null) {
            if (mChildren[order[1]] == null) {
                mChildren[order[1]] = mChildren[order[0]];
                mChildren[order[0]] = dc;
                mHorizontal = horizontal;
            } else {
                DockLayout layout = new DockLayout();
                layout.mParent = this;
                layout.mChildren[order[0]] = dc;
                layout.mHorizontal = horizontal;
                int which = target == mChildren[order[0]] ? 0 : 1;
                layout.mChildren[order[1]] = mChildren[order[which]];
                mChildren[order[which]] = layout;
                if (order[which] == 0) {
                    layout.mDividerPosition = mDividerPosition;
                    mDividerPosition = -1;
                } else {
                    layout.mDividerPosition = -1;
                }
            }
        } else {
            mChildren[order[0]] = dc;
            mHorizontal = horizontal;
        }
    }

    private void dock(DockContainer dc, DockLocation locationRelativeToTarget) {
        int[] order = locationRelativeToTarget.getOrder();
        if (mChildren[order[0]] != null) {
            if (mChildren[order[1]] == null) {
                mChildren[order[1]] = mChildren[order[0]];
            } else {
                mChildren[order[1]] = pushDown();
                mDividerPosition = -1;
            }
        }
        mChildren[order[0]] = dc;
        mHorizontal = locationRelativeToTarget.isHorizontal();
    }

    private DockLayout pushDown() {
        DockLayout layout = new DockLayout();
        layout.mParent = this;
        int length = mChildren.length;
        for (int i = 0; i < length; i++) {
            if (mChildren[i] instanceof DockLayout) {
                ((DockLayout) mChildren[i]).mParent = layout;
            }
            layout.mChildren[i] = mChildren[i];
        }
        layout.mHorizontal = mHorizontal;
        layout.mDividerPosition = mDividerPosition;
        return layout;
    }

    /**
     * @param node The node to remove.
     * @return {@code true} if the node was found and removed.
     */
    public boolean remove(DockLayoutNode node) {
        if (node == mChildren[0]) {
            mChildren[0] = null;
            pullUp(mChildren[1]);
            return true;
        } else if (node == mChildren[1]) {
            mChildren[1] = null;
            pullUp(mChildren[0]);
            return true;
        }
        for (DockLayoutNode child : mChildren) {
            if (child instanceof DockLayout) {
                if (((DockLayout) child).remove(node)) {
                    return true;
                }
            }
        }
        return false;
    }

    private void pullUp(DockLayoutNode node) {
        if (mParent != null) {
            if (mParent.mChildren[0] == this) {
                mParent.mChildren[0] = node;
            } else if (mParent.mChildren[1] == this) {
                mParent.mChildren[1] = node;
            }
            if (node instanceof DockLayout) {
                ((DockLayout) node).mParent = mParent;
            } else if (node == null) {
                if (mParent.mChildren[0] == null && mParent.mChildren[1] == null) {
                    mParent.pullUp(null);
                }
            }
        }
    }

    /** @return The parent {@link DockLayout}. */
    public DockLayout getParent() {
        return mParent;
    }

    /**
     * @return The immediate children of this {@link DockLayout}. Note that the array may contain
     *         {@code null} values.
     */
    public DockLayoutNode[] getChildren() {
        return mChildren;
    }

    /** @return {@code true} if this {@link DockLayout} lays its children out horizontally. */
    public boolean isHorizontal() {
        return mHorizontal;
    }

    /** @return {@code true} if this {@link DockLayout} lays its children out vertically. */
    public boolean isVertical() {
        return !mHorizontal;
    }

    @Override
    public Dimension getPreferredSize() {
        int width  = 0;
        int height = 0;
        if (mChildren[0] != null) {
            Dimension size = mChildren[0].getPreferredSize();
            width = size.width;
            height = size.height;
        }
        if (mChildren[1] != null) {
            Dimension size = mChildren[1].getPreferredSize();
            if (width < size.width) {
                width = size.width;
            }
            if (height < size.height) {
                height = size.height;
            }
            if (mHorizontal) {
                width *= 2;
                width += Dock.DIVIDER_SIZE;
            } else {
                height *= 2;
                height += Dock.DIVIDER_SIZE;
            }
        }
        return new Dimension(width, height);
    }

    @Override
    public int getX() {
        return mX;
    }

    @Override
    public int getY() {
        return mY;
    }

    @Override
    public int getWidth() {
        return mWidth;
    }

    @Override
    public int getHeight() {
        return mHeight;
    }

    /** @return {@code true} if this {@link DockLayout} has no children. */
    public boolean isEmpty() {
        return mChildren[0] == null && mChildren[1] == null;
    }

    /** @return {@code true} if both child nodes of this {@link DockLayout} are occupied. */
    public boolean isFull() {
        return mChildren[0] != null && mChildren[1] != null;
    }

    /**
     * @return The maximum value the divider can be set to. Will always return 0 if {@link
     *         #isFull()} returns {@code false}.
     */
    public int getDividerMaximum() {
        if (isFull()) {
            return Math.max((mHorizontal ? mWidth : mHeight) - Dock.DIVIDER_SIZE, 0);
        }
        return 0;
    }

    public int getRawDividerPosition() {
        return mDividerPosition;
    }

    /** @return The current divider position. */
    public int getDividerPosition() {
        if (isFull()) {
            if (mDividerPosition == -1) {
                if (mHorizontal) {
                    return mChildren[0].getWidth();
                }
                return mChildren[0].getHeight();
            }
            return Math.min(mDividerPosition, getDividerMaximum());
        }
        return 0;
    }

    /** @return {@code true} if the divider is currently set and not in its default mode. */
    public boolean isDividerPositionSet() {
        return mDividerPosition != -1;
    }

    /**
     * @param position The new divider position to set. Use a value less than 0 to reset the divider
     *                 to its default mode, which splits the available space evenly between the
     *                 children.
     */
    public void setDividerPosition(int position) {
        int old = mDividerPosition;
        mDividerPosition = position < 0 ? -1 : position;
        if (mDividerPosition != old && isFull()) {
            setBounds(mX, mY, mWidth, mHeight);
            revalidate();
            getDock().repaint(mX - 1, mY - 1, mWidth + 2, mHeight + 2);
        }
    }

    @Override
    public void revalidate() {
        for (DockLayoutNode child : mChildren) {
            if (child != null) {
                child.revalidate();
            }
        }
    }

    @Override
    public void setBounds(int x, int y, int width, int height) {
        mX = x;
        mY = y;
        mWidth = width;
        mHeight = height;
        Dock          dock               = getDock();
        DockContainer maximizedContainer = dock != null ? dock.getMaximizedContainer() : null;
        if (maximizedContainer != null) {
            forEachDockContainer((dc) -> {
                for (DockLayoutNode child : mChildren) {
                    if (child != null) {
                        child.setBounds(-32000, -32000, child.getWidth(), child.getHeight());
                    }
                }
            });
            maximizedContainer.setBounds(mX, mY, mWidth, mHeight);
        } else if (isFull()) {
            int available = Math.max((mHorizontal ? width : height) - Dock.DIVIDER_SIZE, 0);
            int primary;
            if (mDividerPosition == -1) {
                primary = available / 2;
            } else {
                if (mDividerPosition > available) {
                    mDividerPosition = available;
                }
                primary = mDividerPosition;
            }
            if (mHorizontal) {
                mChildren[0].setBounds(x, y, primary, height);
                mChildren[1].setBounds(x + primary + Dock.DIVIDER_SIZE, y, available - primary, height);
            } else {
                mChildren[0].setBounds(x, y, width, primary);
                mChildren[1].setBounds(x, y + primary + Dock.DIVIDER_SIZE, width, available - primary);
            }
        } else {
            DockLayoutNode node = mChildren[0] != null ? mChildren[0] : mChildren[1];
            if (node != null) {
                node.setBounds(x, y, width, height);
            }
        }
    }

    @Override
    public void addLayoutComponent(String name, Component comp) {
        // Unused
    }

    @Override
    public void removeLayoutComponent(Component comp) {
        if (comp instanceof DockLayoutNode) {
            remove((DockLayoutNode) comp);
        }
    }

    @Override
    public Dimension preferredLayoutSize(Container parent) {
        return getPreferredSize();
    }

    @Override
    public Dimension minimumLayoutSize(Container parent) {
        return new Dimension(Dock.DIVIDER_SIZE, Dock.DIVIDER_SIZE);
    }

    @Override
    public void layoutContainer(Container parent) {
        setBounds(0, 0, parent.getWidth(), parent.getHeight());
    }

    @Override
    public String toString() {
        StringBuilder buffer = new StringBuilder();
        buffer.append(mHorizontal ? "Horizontal" : "Vertical");
        buffer.append(" Dock Layout [id:");
        buffer.append(Integer.toHexString(hashCode()));
        if (mDividerPosition != -1) {
            buffer.append(" d:");
            buffer.append(mDividerPosition);
        }
        buffer.append(" x:");
        buffer.append(mX);
        buffer.append(" y:");
        buffer.append(mY);
        buffer.append(" w:");
        buffer.append(mWidth);
        buffer.append(" h:");
        buffer.append(mHeight);
        if (mParent != null) {
            buffer.append(" p:");
            buffer.append(Integer.toHexString(mParent.hashCode()));
        }
        buffer.append(']');
        return buffer.toString();
    }
}
