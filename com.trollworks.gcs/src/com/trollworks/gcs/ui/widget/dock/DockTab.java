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
import com.trollworks.gcs.menu.file.Saveable;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.DataModifiedListener;
import com.trollworks.gcs.ui.widget.IconButton;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.NumericComparator;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Color;
import java.awt.GradientPaint;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.LayoutManager;
import java.awt.Point;
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DragGestureEvent;
import java.awt.dnd.DragGestureListener;
import java.awt.dnd.DragSource;
import java.awt.event.ContainerEvent;
import java.awt.event.ContainerListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.geom.Path2D;
import javax.swing.Icon;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.SwingConstants;

/** Provides a tab that contains the {@link Dockable}'s icon, title, and close button, if any. */
public class DockTab extends JPanel implements ContainerListener, MouseListener, DragGestureListener, DataModifiedListener, Comparable<DockTab> {
    private Dockable mDockable;
    private JLabel   mTitle;

    /**
     * Creates a new {@link DockTab} for the specified {@link Dockable}.
     *
     * @param dockable The {@link Dockable} to work with.
     */
    public DockTab(Dockable dockable) {
        super(new PrecisionLayout().setMargins(2, 4, 2, 4).setMiddleVerticalAlignment());
        mDockable = dockable;
        setOpaque(false);
        setBorder(new EmptyBorder(2, 1, 0, 1));
        addContainerListener(this);
        mTitle = new JLabel(getFullTitle(), dockable.getTitleIcon(), SwingConstants.LEFT);
        IconButton closeButton = new IconButton(Images.DOCK_CLOSE, I18n.Text("Close"), this::attemptClose);
        add(mTitle, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHeightHint(Math.max(mTitle.getPreferredSize().height, closeButton.getPreferredSize().height)));
        if (dockable instanceof CloseHandler) {
            add(closeButton, new PrecisionLayoutData().setEndHorizontalAlignment());
        }
        if (dockable instanceof Saveable) {
            ((Saveable) dockable).addDataModifiedListener(this);
        }
        addMouseListener(this);
        setToolTipText(Text.wrapPlainTextForToolTip(dockable.getTitleTooltip()));
        DragSource.getDefaultDragSource().createDefaultDragGestureRecognizer(this, DnDConstants.ACTION_MOVE, this);
    }

    /**
     * @return {@code true} if this {@link DockTab} is the current one for the {@link
     *         DockContainer}.
     */
    public boolean isCurrent() {
        DockContainer dc = getDockContainer();
        return dc != null && dc.getCurrentDockable() == mDockable;
    }

    /** @return The {@link Dockable} this tab represents. */
    public Dockable getDockable() {
        return mDockable;
    }

    /** @return The icon used by the tab. */
    public Icon getIcon() {
        return mDockable.getTitleIcon();
    }

    /** @return The full title used by the tab. */
    public String getFullTitle() {
        StringBuilder buffer = new StringBuilder();
        if (mDockable instanceof Saveable) {
            if (((Saveable) mDockable).isModified()) {
                buffer.append('*');
            }
        }
        buffer.append(mDockable.getTitle());
        return buffer.toString();
    }

    /** Update the title and icon from the {@link Dockable}. */
    public void updateTitle() {
        mTitle.setText(getFullTitle());
        mTitle.setIcon(mDockable.getTitleIcon());
        mTitle.revalidate();
    }

    @Override
    protected void paintComponent(Graphics g) {
        Insets        insets = getInsets();
        Path2D.Double path   = new Path2D.Double();
        int           bottom = getHeight();
        path.moveTo(0, bottom);
        path.lineTo(0, 6);
        path.curveTo(0, 6, 0, 1, 6, 1);
        int width = getWidth();
        path.lineTo(width - 7, 1);
        path.curveTo(width - 7, 1, width - 1, 1, width - 1, 7);
        path.lineTo(width - 1, bottom);
        DockContainer dc   = getDockContainer();
        Color         base = DockColors.BACKGROUND;
        if (dc != null) {
            if (dc.getCurrentDockable() == mDockable) {
                base = dc.isActive() ? DockColors.ACTIVE_TAB_BACKGROUND : DockColors.CURRENT_TAB_BACKGROUND;
            }
        }
        Graphics2D gc = (Graphics2D) g;
        gc.setPaint(new GradientPaint(new Point(insets.left, insets.top), base, new Point(insets.left, getHeight() - (insets.top + insets.bottom)), Colors.adjustBrightness(base, -0.1f)));
        gc.fill(path);
        gc.setColor(DockColors.SHADOW);
        gc.draw(path);
    }

    @Override
    public PrecisionLayout getLayout() {
        return (PrecisionLayout) super.getLayout();
    }

    @Override
    public void setLayout(LayoutManager mgr) {
        if (mgr instanceof PrecisionLayout) {
            super.setLayout(mgr);
        } else {
            throw new IllegalArgumentException("Must use a PrecisionLayout.");
        }
    }

    @Override
    public void componentAdded(ContainerEvent event) {
        getLayout().setColumns(getComponentCount());
    }

    @Override
    public void componentRemoved(ContainerEvent event) {
        getLayout().setColumns(getComponentCount());
    }

    private DockContainer getDockContainer() {
        return UIUtilities.getAncestorOfType(this, DockContainer.class);
    }

    public void attemptClose() {
        DockContainer dc = getDockContainer();
        if (dc != null) {
            dc.attemptClose(mDockable);
        }
    }

    @Override
    public void dragGestureRecognized(DragGestureEvent dge) {
        DockableTransferable transferable = new DockableTransferable(mDockable);
        if (DragSource.isDragImageSupported()) {
            Point offset = new Point(dge.getDragOrigin());
            offset.x = -offset.x;
            offset.y = -offset.y;
            dge.startDrag(null, DragSource.isDragImageSupported() ? UIUtilities.getImage(this) : null, offset, transferable, null);
        } else {
            dge.startDrag(null, transferable);
        }
    }

    @Override
    public void mouseEntered(MouseEvent event) {
        // Unused
    }

    @Override
    public void mousePressed(MouseEvent event) {
        DockContainer dc = getDockContainer();
        if (dc.getCurrentDockable() != mDockable) {
            dc.setCurrentDockable(mDockable);
            dc.acquireFocus();
        } else if (!dc.isActive()) {
            dc.acquireFocus();
        }
    }

    @Override
    public void mouseReleased(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseClicked(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseExited(MouseEvent event) {
        // Unused
    }

    @Override
    public void dataModificationStateChanged(Object obj, boolean modified) {
        String title = getFullTitle();
        if (!title.equals(mTitle.getText())) {
            mTitle.setText(title);
            mTitle.revalidate();
        }
    }

    @Override
    public int compareTo(DockTab other) {
        int result = NumericComparator.caselessCompareStrings(mDockable.getTitle(), other.mDockable.getTitle());
        if (result == 0) {
            int h1 = hashCode();
            int h2 = other.hashCode();
            if (h1 < h2) {
                result = -1;
            } else if (h1 > h2) {
                result = 1;
            }
        }
        return result;
    }
}
