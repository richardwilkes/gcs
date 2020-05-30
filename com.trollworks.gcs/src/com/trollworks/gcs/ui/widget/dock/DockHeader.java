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

import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.IconButton;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Insets;
import java.awt.LayoutManager;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DropTarget;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;
import java.awt.dnd.DropTargetListener;
import javax.swing.JPanel;
import javax.swing.border.CompoundBorder;

/** The header for a {@link DockContainer}. */
public class DockHeader extends JPanel implements LayoutManager, DropTargetListener {
    private static final int            MINIMUM_TAB_WIDTH = 60;
    private static final int            GAP               = 4;
    private              IconButton     mMaximizeRestoreButton;
    private              ShowTabsButton mShowTabsButton;
    private              Dockable       mDragDockable;
    private              int            mDragInsertIndex;

    /**
     * Creates a new {@link DockHeader} for the specified {@link DockContainer}.
     *
     * @param dc The {@link DockContainer} to work with.
     */
    public DockHeader(DockContainer dc) {
        super.setLayout(this);
        setOpaque(true);
        setBackground(DockColors.BACKGROUND);
        setBorder(new CompoundBorder(new LineBorder(DockColors.SHADOW, 0, 0, 1, 0), new EmptyBorder(0, 4, 0, 4)));
        for (Dockable dockable : dc.getDockables()) {
            add(new DockTab(dockable));
        }
        mShowTabsButton = new ShowTabsButton();
        add(mShowTabsButton);
        mMaximizeRestoreButton = new IconButton(Images.DOCK_MAXIMIZE, "", this::maximize);
        add(mMaximizeRestoreButton);
        setDropTarget(new DropTarget(this, DnDConstants.ACTION_MOVE, this));
        adjustToRestoredState();
    }

    void addTab(Dockable dockable, int index) {
        add(new DockTab(dockable), index);
        revalidate();
        repaint();
    }

    void close(Dockable dockable) {
        int count = getComponentCount();
        for (int i = 0; i < count; i++) {
            Component child = getComponent(i);
            if (child instanceof DockTab) {
                if (((DockTab) child).getDockable() == dockable) {
                    remove(child);
                    return;
                }
            }
        }
    }

    private DockContainer getDockContainer() {
        return UIUtilities.getAncestorOfType(this, DockContainer.class);
    }

    /** @param index The index of the tab whose title and icon should be updated. */
    public void updateTitle(int index) {
        int count = getComponentCount();
        if (index >= 0 && index < count) {
            Component child = getComponent(index);
            if (child instanceof DockTab) {
                ((DockTab) child).updateTitle();
            }
        }
    }

    private void maximize() {
        getDockContainer().maximize();
    }

    private void restore() {
        getDockContainer().restore();
    }

    /** Called when the owning {@link DockContainer} is set to the maximized state. */
    void adjustToMaximizedState() {
        mMaximizeRestoreButton.setClickFunction(this::restore);
        mMaximizeRestoreButton.setIcon(Images.DOCK_RESTORE);
        mMaximizeRestoreButton.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Restore")));
    }

    /** Called when the owning {@link DockContainer} is restored from the maximized state. */
    void adjustToRestoredState() {
        mMaximizeRestoreButton.setClickFunction(this::maximize);
        mMaximizeRestoreButton.setIcon(Images.DOCK_MAXIMIZE);
        mMaximizeRestoreButton.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Maximize")));
    }

    @Override
    public void setLayout(LayoutManager mgr) {
        // Don't allow overrides
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
        Insets insets = getInsets();
        int    count  = getComponentCount();
        int    width  = 0;
        int    height = 0;
        for (int i = 0; i < count; i++) {
            Component component = getComponent(i);
            if (component != mShowTabsButton) {
                Dimension size = component.getPreferredSize();
                width += size.width + GAP;
                if (height < size.height) {
                    height = size.height;
                }
            }
        }
        width -= GAP;
        if (width < 0) {
            width = 0;
        }
        return new Dimension(insets.left + width + insets.right, insets.top + height + insets.bottom);
    }

    @Override
    public Dimension minimumLayoutSize(Container parent) {
        Insets  insets   = getInsets();
        int     count    = getComponentCount();
        int     width    = 0;
        int     height   = 0;
        boolean foundTab = false;
        for (int i = 0; i < count; i++) {
            Component component = getComponent(i);
            if (component instanceof DockTab) {
                if (foundTab) {
                    continue;
                }
                foundTab = true;
            }
            Dimension size = component.getPreferredSize();
            width += (component instanceof DockTab ? MINIMUM_TAB_WIDTH : size.width) + GAP;
            if (height < size.height) {
                height = size.height;
            }
        }
        width -= GAP;
        if (width < 0) {
            width = 0;
        }
        return new Dimension(insets.left + width + insets.right, insets.top + height + insets.bottom);
    }

    @Override
    public void layoutContainer(Container parent) {
        int         extra         = getWidth() - preferredLayoutSize(parent).width;
        Insets      insets        = getInsets();
        int         count         = getComponentCount();
        Component[] comps         = getComponents();
        int[]       widths        = new int[count];
        int[]       heights       = new int[count];
        int         showTabsIndex = -1;
        mShowTabsButton.clearHidden();
        for (int i = 0; i < count; i++) {
            Dimension size = comps[i].getPreferredSize();
            widths[i] = size.width;
            heights[i] = size.height;
            if (comps[i] == mShowTabsButton) {
                showTabsIndex = i;
            }
        }
        if (extra < 0) {
            int     current   = getDockContainer().getCurrentTabIndex();
            int     remaining = -extra;
            boolean found     = true;
            // Shrink the non-current tabs down
            while (found && remaining > 0) {
                int tabs = 0;
                found = false;
                for (int i = 0; i < count; i++) {
                    if (i != current && comps[i] instanceof DockTab && widths[i] > MINIMUM_TAB_WIDTH) {
                        tabs++;
                    }
                }
                if (tabs > 0) {
                    int perTab = Math.max(remaining / tabs, 1);
                    for (int i = 0; i < count && remaining > 0; i++) {
                        if (i != current && comps[i] instanceof DockTab && widths[i] > MINIMUM_TAB_WIDTH) {
                            found = true;
                            remaining -= perTab;
                            widths[i] -= perTab;
                            if (widths[i] <= MINIMUM_TAB_WIDTH) {
                                remaining += MINIMUM_TAB_WIDTH - widths[i];
                                widths[i] = MINIMUM_TAB_WIDTH;
                            }
                        }
                    }
                }
            }
            if (remaining > 0) {
                // Still not small enough... start trimming out tabs
                remaining += widths[showTabsIndex] + GAP;
                for (int i = count - 1; i >= 0 && remaining > 0; i--) {
                    if (i != current && comps[i] instanceof DockTab) {
                        remaining -= widths[showTabsIndex];
                        mShowTabsButton.addHidden((DockTab) comps[i]);
                        widths[showTabsIndex] = mShowTabsButton.getPreferredWidth();
                        remaining += widths[showTabsIndex];
                        remaining -= widths[i] + GAP;
                    }
                }
                if (remaining > 0) {
                    // STILL not small enough... reduce the size of the current tab, too
                    widths[current] -= remaining;
                    if (widths[current] < MINIMUM_TAB_WIDTH) {
                        widths[current] = MINIMUM_TAB_WIDTH;
                    }
                    remaining = 0;
                }
                extra = -remaining;
            } else {
                extra = 0;
            }
        }
        int     x           = insets.left;
        int     height      = getHeight();
        boolean insertExtra = true;
        for (int i = 0; i < count; i++) {
            if (mShowTabsButton.isHidden(comps[i])) {
                comps[i].setVisible(false);
            } else {
                comps[i].setVisible(true);
                if (insertExtra && !(comps[i] instanceof DockTab)) {
                    insertExtra = false;
                    x += extra;
                }
                comps[i].setBounds(x, insets.top + (height - (insets.top + heights[i] + insets.bottom)) / 2, widths[i], heights[i]);
                x += widths[i] + GAP;
            }
        }
    }

    @Override
    protected void paintComponent(Graphics g) {
        super.paintComponent(g);
        if (mDragDockable != null && mDragInsertIndex >= 0) {
            Insets insets = getInsets();
            int    count  = getComponentCount();
            int    x;
            if (mDragInsertIndex < count) {
                Component child = getComponent(mDragInsertIndex);
                if (child instanceof DockTab) {
                    x = child.getX() - GAP;
                } else if (mDragInsertIndex > 0) {
                    child = getComponent(mDragInsertIndex - 1);
                    x = child.getX() + child.getWidth();
                } else {
                    x = insets.left;
                }
            } else {
                if (count > 0) {
                    Component child = getComponent(count - 1);
                    x = child.getX() + child.getWidth() + GAP / 2;
                } else {
                    x = insets.left;
                }
            }
            g.setColor(DockColors.DROP_AREA_OUTER_BORDER);
            g.drawLine(x, insets.top, x, getHeight() - (insets.top + insets.bottom));
            g.drawLine(x + 3, insets.top, x + 3, getHeight() - (insets.top + insets.bottom));
            g.setColor(DockColors.DROP_AREA_INNER_BORDER);
            g.drawLine(x + 1, insets.top, x + 1, getHeight() - (insets.top + insets.bottom));
            g.drawLine(x + 2, insets.top, x + 2, getHeight() - (insets.top + insets.bottom));
        }
    }

    private Dockable getDockableInDrag(DropTargetDragEvent dtde) {
        if (dtde.getDropAction() == DnDConstants.ACTION_MOVE) {
            try {
                if (dtde.isDataFlavorSupported(DockableTransferable.DATA_FLAVOR)) {
                    Dockable      dockable = (Dockable) dtde.getTransferable().getTransferData(DockableTransferable.DATA_FLAVOR);
                    DockContainer dc       = dockable.getDockContainer();
                    if (dc != null && dc.getDock() == getDockContainer().getDock()) {
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
            mDragInsertIndex = -1;
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
        if (mDragDockable != null && mDragInsertIndex != -1) {
            getDockContainer().stack(mDragDockable, mDragInsertIndex);
            dtde.acceptDrop(DnDConstants.ACTION_MOVE);
            dtde.dropComplete(true);
        } else {
            dtde.rejectDrop();
            dtde.dropComplete(false);
        }
        clearDragState();
    }

    private void clearDragState() {
        repaint();
        mDragDockable = null;
        mDragInsertIndex = -1;
    }

    private void updateForDragOver(Point where) {
        int count    = getComponentCount();
        int insertAt = count;
        for (int i = 0; i < count; i++) {
            Component child = getComponent(i);
            if (child instanceof DockTab) {
                Rectangle bounds = child.getBounds();
                if (where.x < bounds.x + bounds.width / 2) {
                    insertAt = i;
                    break;
                }
                if (where.x < bounds.x + bounds.width) {
                    insertAt = i + 1;
                    break;
                }
            } else {
                insertAt = i;
                break;
            }
        }
        if (insertAt != mDragInsertIndex) {
            mDragInsertIndex = insertAt;
            repaint();
        }
    }
}
