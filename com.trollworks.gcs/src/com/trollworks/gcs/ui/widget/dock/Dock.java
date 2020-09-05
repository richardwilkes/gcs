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

import com.trollworks.gcs.ui.MouseCapture;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.image.Cursors;
import com.trollworks.gcs.utility.Log;

import java.awt.AWTEvent;
import java.awt.Component;
import java.awt.Graphics;
import java.awt.IllegalComponentStateException;
import java.awt.KeyboardFocusManager;
import java.awt.LayoutManager;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DropTarget;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;
import java.awt.dnd.DropTargetListener;
import java.awt.event.AWTEventListener;
import java.awt.event.HierarchyEvent;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JPanel;

/** Provides an area where {@link Dockable} components can be displayed and rearranged. */
public class Dock extends JPanel implements MouseListener, MouseMotionListener, PropertyChangeListener, DropTargetListener {
    private static final String         PERMANENT_FOCUS_OWNER_KEY = "permanentFocusOwner";
    private static final int            GRIP_GAP                  = 1;
    private static final int            GRIP_WIDTH                = 4;
    private static final int            GRIP_HEIGHT               = 2;
    private static final int            GRIP_LENGTH               = GRIP_HEIGHT * 5 + GRIP_GAP * 4;
    public static final  int            DIVIDER_SIZE              = GRIP_WIDTH + 4;
    private static final int            DRAG_THRESHOLD            = 5;
    private static final long           DRAG_DELAY                = 250;
    private              long           mDividerDragStartedAt;
    private              int            mDividerDragStartX;
    private              int            mDividerDragStartY;
    private              DockLayout     mDividerDragLayout;
    private              int            mDividerDragInitialEventPosition;
    private              int            mDividerDragInitialDividerPosition;
    private              boolean        mDividerDragIsValid;
    private              Dockable       mDragDockable;
    private              DockLayoutNode mDragOverNode;
    private              DockLocation   mDragOverLocation;
    private              DockContainer  mMaximizedContainer;
    private              MouseMonitor   mMouseMonitor;

    /** Creates a new, empty {@link Dock}. */
    public Dock() {
        super(new DockLayout(), true);
        setBorder(null);
        addMouseListener(this);
        addMouseMotionListener(this);
        setFocusCycleRoot(true);
        setDropTarget(new DropTarget(this, DnDConstants.ACTION_MOVE, this));
        mMouseMonitor = new MouseMonitor();
        addHierarchyListener((event) -> {
            if (event.getID() == HierarchyEvent.HIERARCHY_CHANGED && (event.getChangeFlags() & HierarchyEvent.DISPLAYABILITY_CHANGED) == HierarchyEvent.DISPLAYABILITY_CHANGED) {
                if (isDisplayable()) {
                    getToolkit().addAWTEventListener(mMouseMonitor, AWTEvent.MOUSE_EVENT_MASK);
                } else {
                    getToolkit().removeAWTEventListener(mMouseMonitor);
                }
            }
        });
    }

    /**
     * Docks a {@link Dockable} within this {@link Dock}. If the {@link Dockable} already exists in
     * this {@link Dock}, it will be moved to the new location.
     *
     * @param dockable The {@link Dockable} to install into this {@link Dock}.
     * @param location The location within the top level to install the {@link Dockable}.
     */
    public void dock(Dockable dockable, DockLocation location) {
        dock(dockable, getLayout(), location);
    }

    /**
     * Docks a {@link Dockable} within this {@link Dock}. If the {@link Dockable} already exists in
     * this {@link Dock}, it will be moved to the new location.
     *
     * @param dockable                 The {@link Dockable} to install into this {@link Dock}.
     * @param target                   The target {@link Dockable}.
     * @param locationRelativeToTarget The location relative to the target to install the {@link
     *                                 Dockable}. You may pass in {@code null} to have it stack with
     *                                 the target.
     */
    public void dock(Dockable dockable, Dockable target, DockLocation locationRelativeToTarget) {
        DockContainer dc = target.getDockContainer();
        if (dc != null && dc.getDock() == this) {
            dock(dockable, dc, locationRelativeToTarget);
        }
    }

    /**
     * Docks a {@link Dockable} within this {@link Dock}. If the {@link Dockable} already exists in
     * this {@link Dock}, it will be moved to the new location.
     *
     * @param dockable                 The {@link Dockable} to install into this {@link Dock}.
     * @param target                   The target {@link DockLayoutNode}.
     * @param locationRelativeToTarget The location relative to the target to install the {@link
     *                                 Dockable}. If the target is a {@link DockContainer}, you may
     *                                 pass in {@code null} to have it stack with the target.
     */
    public void dock(Dockable dockable, DockLayoutNode target, DockLocation locationRelativeToTarget) {
        DockLayout layout = getLayout();
        if (layout.contains(target)) {
            DockContainer dc = dockable.getDockContainer();
            if (dc == target) {
                if (dc.getDockables().size() == 1) {
                    // It's already where it needs to be
                    return;
                }
            }
            if (dc != null) {
                // Remove it from it's old position
                ArrayList<DockLayoutNode> layouts = new ArrayList<>();
                if (target instanceof DockLayout) {
                    while (target != null) {
                        layouts.add(target);
                        target = ((DockLayout) target).getParent();
                    }
                    target = layouts.get(0);
                    for (DockLayoutNode child : ((DockLayout) target).getChildren()) {
                        if (child != dc) {
                            layouts.add(1, child);
                        }
                    }
                }
                dc.close(dockable);
                if (target instanceof DockLayout) {
                    int i     = 1;
                    int count = layouts.size();
                    while (!layout.contains(target)) {
                        if (i >= count) {
                            target = layout;
                            break;
                        }
                        target = layouts.get(i++);
                    }
                }
            }
            dc = new DockContainer(this, dockable);
            layout.dock(dc, target, locationRelativeToTarget);
            addImpl(dc, null, -1);
            revalidate();
            dc.setCurrentDockable(dockable);
        }
    }

    @Override
    protected void paintComponent(Graphics gc) {
        Rectangle bounds = gc.getClipBounds();
        if (bounds == null) {
            bounds = new Rectangle(0, 0, getWidth(), getHeight());
        }
        gc.setColor(getBackground());
        gc.fillRect(bounds.x, bounds.y, bounds.width, bounds.height);
        drawDividers(gc, getLayout(), bounds);
    }

    @Override
    protected void paintChildren(Graphics gc) {
        super.paintChildren(gc);
        if (mDragOverNode != null) {
            Rectangle bounds = getDragOverBounds();
            gc.setColor(DockColors.DROP_AREA);
            gc.fillRect(bounds.x, bounds.y, bounds.width, bounds.height);
            gc.setColor(DockColors.DROP_AREA_INNER_BORDER);
            gc.drawRect(bounds.x + 1, bounds.y + 1, bounds.width - 3, bounds.height - 3);
            gc.setColor(DockColors.DROP_AREA_OUTER_BORDER);
            gc.drawRect(bounds.x, bounds.y, bounds.width - 1, bounds.height - 1);
        }
    }

    @Override
    protected boolean isPaintingOrigin() {
        // This is necessary to ensure we do the correct overdraw when dragging.
        return true;
    }

    private void drawDividers(Graphics gc, DockLayout layout, Rectangle clip) {
        if (clip.intersects(layout.getX() - 1, layout.getY() - 1, layout.getWidth() + 2, layout.getHeight() + 2)) {
            DockLayoutNode[] children = layout.getChildren();
            if (layout.isFull()) {
                if (layout.isHorizontal()) {
                    drawHorizontalGripper(gc, children[1]);
                } else {
                    drawVerticalGripper(gc, children[1]);
                }
            }
            drawDockLayoutNode(gc, children[0], clip);
            drawDockLayoutNode(gc, children[1], clip);
        }
    }

    private static void drawHorizontalGripper(Graphics gc, DockLayoutNode secondary) {
        int x      = secondary.getX() - DIVIDER_SIZE + (DIVIDER_SIZE - GRIP_WIDTH) / 2;
        int y      = secondary.getY() + (secondary.getHeight() - GRIP_LENGTH) / 2;
        int top    = GRIP_HEIGHT / 2;
        int bottom = GRIP_HEIGHT - top;
        for (int yy = y; yy < y + GRIP_LENGTH; yy += GRIP_HEIGHT + GRIP_GAP) {
            gc.setColor(DockColors.HIGHLIGHT);
            gc.fillRect(x, yy, GRIP_WIDTH, top);
            gc.setColor(DockColors.SHADOW);
            gc.fillRect(x, yy + top, GRIP_WIDTH, bottom);
        }
    }

    private static void drawVerticalGripper(Graphics gc, DockLayoutNode secondary) {
        int x      = secondary.getX() + (secondary.getWidth() - GRIP_LENGTH) / 2;
        int y      = secondary.getY() - DIVIDER_SIZE + (DIVIDER_SIZE - GRIP_WIDTH) / 2;
        int top    = GRIP_HEIGHT / 2;
        int bottom = GRIP_HEIGHT - top;
        for (int xx = x; xx < x + GRIP_LENGTH; xx += GRIP_HEIGHT + GRIP_GAP) {
            gc.setColor(DockColors.HIGHLIGHT);
            gc.fillRect(xx, y, top, GRIP_WIDTH);
            gc.setColor(DockColors.SHADOW);
            gc.fillRect(xx + top, y, bottom, GRIP_WIDTH);
        }
    }

    private void drawDockLayoutNode(Graphics gc, DockLayoutNode node, Rectangle clip) {
        if (node instanceof DockLayout) {
            drawDividers(gc, (DockLayout) node, clip);
        } else if (node != null) {
            int layoutWidth = node.getWidth();
            if (layoutWidth > 0) {
                int layoutHeight = node.getHeight();
                if (layoutHeight > 0) {
                    gc.setColor(DockColors.SHADOW);
                    gc.drawRect(node.getX() - 1, node.getY() - 1, node.getWidth() + 1, node.getHeight() + 1);
                }
            }
        }
    }

    @Override
    public void mouseEntered(MouseEvent event) {
        updateCursor(event);
    }

    @Override
    public void mouseMoved(MouseEvent event) {
        updateCursor(event);
    }

    @Override
    public void mousePressed(MouseEvent event) {
        DockLayoutNode over = over(event.getX(), event.getY());
        if (over instanceof DockLayout) {
            mDividerDragLayout = (DockLayout) over;
            mDividerDragStartedAt = event.getWhen();
            mDividerDragStartX = event.getX();
            mDividerDragStartY = event.getY();
            mDividerDragInitialEventPosition = mDividerDragLayout.isHorizontal() ? event.getX() : event.getY();
            mDividerDragInitialDividerPosition = mDividerDragLayout.getDividerPosition();
            mDividerDragIsValid = false;
            MouseCapture.start(this, mDividerDragLayout.isHorizontal() ? Cursors.HORIZONTAL_RESIZE : Cursors.VERTICAL_RESIZE);
        }
    }

    @Override
    public void mouseDragged(MouseEvent event) {
        dragDivider(event);
    }

    private void dragDivider(MouseEvent event) {
        if (mDividerDragLayout != null) {
            if (!mDividerDragIsValid) {
                mDividerDragIsValid = Math.abs(mDividerDragStartX - event.getX()) > DRAG_THRESHOLD || Math.abs(mDividerDragStartY - event.getY()) > DRAG_THRESHOLD || event.getWhen() - mDividerDragStartedAt > DRAG_DELAY;
            }
            if (mDividerDragIsValid) {
                int pos = mDividerDragInitialDividerPosition - (mDividerDragInitialEventPosition - (mDividerDragLayout.isHorizontal() ? event.getX() : event.getY()));
                mDividerDragLayout.setDividerPosition(Math.max(pos, 0));
            }
        }
    }

    @Override
    public void mouseReleased(MouseEvent event) {
        if (mDividerDragLayout != null) {
            if (mDividerDragIsValid) {
                dragDivider(event);
            }
            mDividerDragLayout = null;
            MouseCapture.stop(this);
        }
    }

    @Override
    public void mouseClicked(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseExited(MouseEvent event) {
        setCursor(null);
    }

    private void updateCursor(MouseEvent event) {
        DockLayoutNode over = over(event.getX(), event.getY());
        if (over instanceof DockLayout) {
            setCursor(((DockLayout) over).isHorizontal() ? Cursors.HORIZONTAL_RESIZE : Cursors.VERTICAL_RESIZE);
        } else {
            setCursor(null);
        }
    }

    private static boolean containedBy(DockLayoutNode node, int x, int y) {
        if (node != null) {
            int edgeX = node.getX();
            if (x >= edgeX && x < edgeX + node.getWidth()) {
                int edgeY = node.getY();
                return y >= edgeY && y < edgeY + node.getHeight();
            }
        }
        return false;
    }

    private DockLayoutNode over(int x, int y) {
        return over(getLayout(), x, y);
    }

    private DockLayoutNode over(DockLayoutNode node, int x, int y) {
        if (containedBy(node, x, y)) {
            if (node instanceof DockLayout) {
                DockLayout layout = (DockLayout) node;
                for (DockLayoutNode child : layout.getChildren()) {
                    if (containedBy(child, x, y)) {
                        return over(child, x, y);
                    }
                }
                if (layout.isFull()) {
                    return node;
                }
            } else if (node instanceof DockContainer) {
                return node;
            }
        }
        return null;
    }

    @Override
    public void setLayout(LayoutManager mgr) {
        if (mgr instanceof DockLayout) {
            super.setLayout(mgr);
        } else {
            throw new IllegalArgumentException("Must use a DockLayout.");
        }
    }

    @Override
    public DockLayout getLayout() {
        return (DockLayout) super.getLayout();
    }

    @Override
    public String toString() {
        StringBuilder buffer = new StringBuilder();
        dump(buffer, 0, getLayout());
        buffer.setLength(buffer.length() - 1);
        return buffer.toString();
    }

    private static void dump(StringBuilder buffer, int depth, DockLayoutNode node) {
        if (node instanceof DockLayout) {
            DockLayout layout = (DockLayout) node;
            pad(buffer, depth);
            buffer.append(layout);
            buffer.append('\n');
            depth++;
            for (DockLayoutNode child : layout.getChildren()) {
                if (child != null) {
                    dump(buffer, depth, child);
                }
            }
        } else if (node instanceof DockContainer) {
            pad(buffer, depth);
            buffer.append(node);
            buffer.append('\n');
        }
    }

    private static void pad(StringBuilder buffer, int depth) {
        buffer.append("...".repeat(Math.max(0, depth)));
    }

    /**
     * Use one of {@link #dock(Dockable, DockLocation)}, {@link #dock(Dockable, Dockable,
     * DockLocation)}, or {@link #dock(Dockable, DockLayoutNode, DockLocation)} instead.
     */
    @Override
    public final Component add(Component comp) {
        throw createIllegalComponentStateException();
    }

    /**
     * Use one of {@link #dock(Dockable, DockLocation)}, {@link #dock(Dockable, Dockable,
     * DockLocation)}, or {@link #dock(Dockable, DockLayoutNode, DockLocation)} instead.
     */
    @Override
    public final Component add(Component comp, int index) {
        throw createIllegalComponentStateException();
    }

    /**
     * Use one of {@link #dock(Dockable, DockLocation)}, {@link #dock(Dockable, Dockable,
     * DockLocation)}, or {@link #dock(Dockable, DockLayoutNode, DockLocation)} instead.
     */
    @Override
    public final void add(Component comp, Object constraints) {
        throw createIllegalComponentStateException();
    }

    /**
     * Use one of {@link #dock(Dockable, DockLocation)}, {@link #dock(Dockable, Dockable,
     * DockLocation)}, or {@link #dock(Dockable, DockLayoutNode, DockLocation)} instead.
     */
    @Override
    public final void add(Component comp, Object constraints, int index) {
        throw createIllegalComponentStateException();
    }

    /**
     * Use one of {@link #dock(Dockable, DockLocation)}, {@link #dock(Dockable, Dockable,
     * DockLocation)}, or {@link #dock(Dockable, DockLayoutNode, DockLocation)} instead.
     */
    @Override
    public final Component add(String name, Component comp) {
        throw createIllegalComponentStateException();
    }

    private static IllegalComponentStateException createIllegalComponentStateException() {
        return new IllegalComponentStateException("Use one of the dock() methods instead");
    }

    @Override
    public void addNotify() {
        super.addNotify();
        KeyboardFocusManager.getCurrentKeyboardFocusManager().addPropertyChangeListener(PERMANENT_FOCUS_OWNER_KEY, this);
    }

    @Override
    public void removeNotify() {
        super.removeNotify();
        KeyboardFocusManager.getCurrentKeyboardFocusManager().removePropertyChangeListener(PERMANENT_FOCUS_OWNER_KEY, this);
    }

    @Override
    public void propertyChange(PropertyChangeEvent event) {
        getLayout().forEachDockContainer(DockContainer::updateActiveHighlight);
    }

    /** @return The {@link DockContainer} with the current keyboard focus, or {@code null}. */
    public DockContainer getFocusedDockContainer() {
        return getLayout().getFocusedDockContainer();
    }

    /** @return The current maximized {@link DockContainer}, or {@code null}. */
    public DockContainer getMaximizedContainer() {
        return mMaximizedContainer;
    }

    /**
     * Causes the {@link DockContainer} to fill the entire {@link Dock} area.
     *
     * @param dc The {@link DockContainer} to maximize.
     */
    public void maximize(DockContainer dc) {
        if (mMaximizedContainer != null) {
            mMaximizedContainer.getHeader().adjustToRestoredState();
        }
        mMaximizedContainer = dc;
        mMaximizedContainer.getHeader().adjustToMaximizedState();
        getLayout().forEachDockContainer((target) -> target.setVisible(target == mMaximizedContainer));
        revalidate();
        mMaximizedContainer.acquireFocus();
        repaint();
    }

    /** Restores the current maximized {@link DockContainer} to its normal state. */
    public void restore() {
        if (mMaximizedContainer != null) {
            mMaximizedContainer.getHeader().adjustToRestoredState();
            mMaximizedContainer = null;
            getLayout().forEachDockContainer((dc) -> dc.setVisible(true));
            revalidate();
            repaint();
        }
    }

    /** @return All the {@link Dockable}s contained in this {@link Dock}. */
    public List<Dockable> getDockables() {
        List<Dockable> dockables = new ArrayList<>();
        getLayout().forEachDockContainer((dc) -> dockables.addAll(dc.getDockables()));
        return dockables;
    }

    private Dockable getDockableInDrag(DropTargetDragEvent dtde) {
        if (dtde.getDropAction() == DnDConstants.ACTION_MOVE) {
            try {
                if (dtde.isDataFlavorSupported(DockableTransferable.DATA_FLAVOR)) {
                    Dockable      dockable = (Dockable) dtde.getTransferable().getTransferData(DockableTransferable.DATA_FLAVOR);
                    DockContainer dc       = dockable.getDockContainer();
                    if (dc != null && dc.getDock() == this) {
                        return dockable;
                    }
                }
            } catch (Exception exception) {
                Log.error(exception);
            }
        }
        return null;
    }

    @Override
    public void dragEnter(DropTargetDragEvent dtde) {
        mDragDockable = getDockableInDrag(dtde);
        if (mDragDockable != null) {
            mDragOverNode = null;
            mDragOverLocation = null;
            updateForDragOver(dtde.getLocation());
            dtde.acceptDrag(DnDConstants.ACTION_MOVE);
        } else {
            dtde.rejectDrag();
        }
    }

    @Override
    public void dragOver(DropTargetDragEvent dtde) {
        if (mDragDockable != null) {
            updateForDragOver(dtde.getLocation());
            dtde.acceptDrag(DnDConstants.ACTION_MOVE);
        } else {
            dtde.rejectDrag();
        }
    }

    @Override
    public void dragExit(DropTargetEvent dte) {
        if (mDragDockable != null) {
            clearDragState();
        }
    }

    @Override
    public void dropActionChanged(DropTargetDragEvent dtde) {
        mDragDockable = getDockableInDrag(dtde);
        if (mDragDockable != null) {
            updateForDragOver(dtde.getLocation());
            dtde.acceptDrag(DnDConstants.ACTION_MOVE);
        } else {
            clearDragState();
            dtde.rejectDrag();
        }
    }

    @Override
    public void drop(DropTargetDropEvent dtde) {
        if (mDragDockable != null) {
            if (mDragOverNode != null) {
                dock(mDragDockable, mDragOverNode, mDragOverLocation);
                revalidate();
            }
            dtde.acceptDrop(DnDConstants.ACTION_MOVE);
            dtde.dropComplete(true);
        } else {
            dtde.dropComplete(false);
        }
        clearDragState();
    }

    private void clearDragState() {
        if (mDragOverNode != null) {
            repaint(getDragOverBounds());
        }
        mDragDockable = null;
        mDragOverNode = null;
        mDragOverLocation = null;
    }

    private void updateForDragOver(Point where) {
        int            ex       = where.x;
        int            ey       = where.y;
        DockLocation   location = null;
        DockLayoutNode over     = over(ex, ey);
        if (over != null) {
            int x      = over.getX();
            int y      = over.getY();
            int width  = over.getWidth();
            int height = over.getHeight();
            ex -= x;
            ey -= y;
            if (ex < width / 2) {
                location = DockLocation.WEST;
            } else {
                location = DockLocation.EAST;
                ex = width - ex;
            }
            if (ey < height / 2) {
                if (ex > ey) {
                    location = DockLocation.NORTH;
                }
            } else if (ex > height - ey) {
                location = DockLocation.SOUTH;
            }
        }
        if (over != mDragOverNode || location != mDragOverLocation) {
            if (mDragOverNode != null) {
                repaint(getDragOverBounds());
            }
            mDragOverNode = over;
            mDragOverLocation = location;
            if (mDragOverNode != null) {
                repaint(getDragOverBounds());
            }
        }
    }

    private Rectangle getDragOverBounds() {
        Rectangle bounds = new Rectangle(mDragOverNode.getX(), mDragOverNode.getY(), mDragOverNode.getWidth(), mDragOverNode.getHeight());
        switch (mDragOverLocation) {
        case NORTH -> bounds.height = Math.max(bounds.height / 2, 1);
        case SOUTH -> {
            int halfHeight = Math.max(bounds.height / 2, 1);
            bounds.y += bounds.height - halfHeight;
            bounds.height = halfHeight;
        }
        case EAST -> {
            int halfWidth = Math.max(bounds.width / 2, 1);
            bounds.x += bounds.width - halfWidth;
            bounds.width = halfWidth;
        }
        default -> bounds.width = Math.max(bounds.width / 2, 1);
        }
        return bounds;
    }

    /**
     * Creates a new, untitled window title.
     *
     * @param baseTitle The base untitled name.
     * @param exclude   A {@link Dockable} to exclude from naming decisions. May be {@code null}.
     * @return The new {@link Dockable} title.
     */
    public String getNextUntitledDockableName(String baseTitle, Dockable exclude) {
        List<Dockable> dockables = getDockables();
        int            value     = 0;
        String         title;
        boolean        again;
        do {
            again = false;
            title = baseTitle;
            if (++value > 1) {
                title += " " + value;
            }
            for (Dockable dockable : dockables) {
                if (dockable != exclude && title.equals(dockable.getTitle())) {
                    again = true;
                    break;
                }
            }
        } while (again);
        return title;
    }

    class MouseMonitor implements AWTEventListener {
        @Override
        public void eventDispatched(AWTEvent event) {
            if (event.getID() == MouseEvent.MOUSE_PRESSED) {
                Object source = event.getSource();
                if (source instanceof Component) {
                    Component comp = (Component) source;
                    if (Dock.this == UIUtilities.getAncestorOfType(comp, Dock.class)) {
                        Dockable dockable;
                        dockable = comp instanceof Dockable ? (Dockable) comp : UIUtilities.getAncestorOfType(comp, Dockable.class);
                        if (dockable != null) {
                            DockContainer dc = dockable.getDockContainer();
                            if (dc != null) {
                                dc.setCurrentDockable(dockable, comp);
                            }
                        }
                    }
                }
            }
        }
    }
}
