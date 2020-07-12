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

package com.trollworks.gcs.ui.widget.tree;

import com.trollworks.gcs.menu.edit.Deletable;
import com.trollworks.gcs.menu.edit.Openable;
import com.trollworks.gcs.menu.edit.SelectAllCapable;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.ui.widget.DirectScrollPanel;
import com.trollworks.gcs.ui.widget.DirectScrollPanelArea;
import com.trollworks.gcs.ui.widget.dock.Dock;
import com.trollworks.gcs.ui.widget.dock.DockableTransferable;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.utility.task.Tasks;

import java.awt.AlphaComposite;
import java.awt.Color;
import java.awt.Composite;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.Shape;
import java.awt.Transparency;
import java.awt.datatransfer.Transferable;
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DragGestureEvent;
import java.awt.dnd.DragGestureListener;
import java.awt.dnd.DragSource;
import java.awt.dnd.DragSourceDragEvent;
import java.awt.dnd.DragSourceDropEvent;
import java.awt.dnd.DragSourceEvent;
import java.awt.dnd.DragSourceListener;
import java.awt.dnd.DropTarget;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;
import java.awt.dnd.DropTargetListener;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.TimeUnit;
import javax.swing.UIManager;

/** Provides a flexible tree widget. */
public class TreePanel extends DirectScrollPanel implements Runnable, Openable, Deletable, SelectAllCapable, DropTargetListener, DragSourceListener, DragGestureListener, FocusListener, KeyListener, MouseListener, MouseMotionListener, NotifierTarget {
    private static final float                     DRAG_OPACITY            = 0.75f;
    /** The amount of 'slop' to allow for hit detection. */
    public static final  int                       HIT_SLOP                = 4;
    private static final int                       DRAG_FOCUS_WIDTH        = 3;
    private static final int                       DRAG_INSERT_WIDTH       = 3;
    /** The amount of indent per level of hierarchy. */
    public static final  int                       INDENT                  = TextTreeColumn.HMARGIN + 16 + TextTreeColumn.ICON_GAP;
    private              TreeRoot                  mRoot;
    private              ArrayList<TreeColumn>     mColumns                = new ArrayList<>();
    private              DirectScrollPanelArea     mViewArea;
    private              Deletable                 mDeletableProxy;
    private              Openable                  mOpenableProxy;
    private              Color                     mDividerColor           = Color.GRAY;
    private              Color                     mHierarchyLineColor     = new Color(224, 224, 224);
    private              HashSet<TreeContainerRow> mOpenRows               = new HashSet<>();
    private              HashSet<TreeRow>          mSelectedRows           = new HashSet<>();
    private              Map<TreeRow, Integer>     mRowHeightMap           = new HashMap<>();
    private              int                       mRowHeight              = TextTreeColumn.VMARGIN + TextDrawing.getFontHeight(Fonts.getDefaultFont()) + TextTreeColumn.VMARGIN;
    private              int                       mMouseOverColumnDivider = -1;
    private              int                       mDragColumnDivider      = -1;
    private              int                       mAllowedRowDragTypes    = DnDConstants.ACTION_COPY_OR_MOVE;
    private              int                       mAllowedRowDropTypes    = DnDConstants.ACTION_COPY_OR_MOVE;
    private              TreeSorter                mSorter                 = new TreeSorter();
    private              TreeColumn                mSortColumn;
    private              TreeColumn                mSourceDragColumn;
    private              TreeDragState             mDragState;
    private              TreeRow                   mAnchorRow;
    private              TreeRow                   mRowToSelectOnMouseUp;
    private              TreeRow                   mResizeRow;
    private              Dock                      mAlternateDragDestination;
    private              boolean                   mShowDisclosureControls = true;
    private              boolean                   mUseBanding             = true;
    private              boolean                   mAllowColumnResize      = true;
    private              boolean                   mAllowColumnDrag        = true;
    private              boolean                   mAllowColumnContextMenu = true;
    private              boolean                   mAllowRowDropFromExternal;
    private              boolean                   mUserSortable           = true;
    private              boolean                   mShowColumnDivider      = true;
    private              boolean                   mShowRowDivider         = true;
    private              boolean                   mShowHeader             = true;
    private              boolean                   mDropReceived;
    private              boolean                   mResizePending;
    private              boolean                   mIgnoreNextDragGesture;

    /**
     * Creates a new {@link TreePanel}.
     *
     * @param root The {@link TreeRoot} to use.
     */
    public TreePanel(TreeRoot root) {
        mRoot = root;
        mRoot.getNotifier().add(this, TreeNotificationKeys.ROW_REMOVED);
        setUnitIncrement(mRowHeight + getRowDividerHeight());
        setFocusable(true);
        addFocusListener(this);
        addMouseListener(this);
        addMouseMotionListener(this);
        addKeyListener(this);
        if (!GraphicsUtilities.inHeadlessPrintMode()) {
            DragSource.getDefaultDragSource().createDefaultDragGestureRecognizer(this, DnDConstants.ACTION_COPY_OR_MOVE, this);
            setDropTarget(new DropTarget(this, this));
        }
    }

    /** @return The {@link TreeRoot} being displayed. */
    public TreeRoot getRoot() {
        return mRoot;
    }

    public void setRoot(TreeRoot root) {
        if (root != mRoot) {
            mRoot.getNotifier().remove(this);
            mRoot = root;
            mRoot.getNotifier().add(this, TreeNotificationKeys.ROW_REMOVED);
            repaint();
        }
    }

    @Override
    public void mouseEntered(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseExited(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseMoved(MouseEvent event) {
        Point                 where = event.getPoint();
        DirectScrollPanelArea area  = checkAndConvertToArea(where);
        if (area != DirectScrollPanelArea.NONE) {
            setColumnDividerHighlight(where.x);
        }
    }

    private boolean setColumnDividerHighlight(int x) {
        int oldDividerIndex = mMouseOverColumnDivider;
        int dividerIndex    = mAllowColumnResize ? overColumnDivider(x) : -1;
        setCursor(dividerIndex == -1 ? Cursor.getDefaultCursor() : Cursor.getPredefinedCursor(Cursor.W_RESIZE_CURSOR));
        mMouseOverColumnDivider = dividerIndex;
        if (oldDividerIndex != dividerIndex) {
            repaintDivider(oldDividerIndex);
            repaintDivider(dividerIndex);
        }
        return dividerIndex != -1;
    }

    private void clearColumnDividerHighlight() {
        if (mMouseOverColumnDivider != -1) {
            repaintDivider(mMouseOverColumnDivider);
            mMouseOverColumnDivider = -1;
        }
        setCursor(Cursor.getDefaultCursor());
    }

    private void repaintDivider(int index) {
        if (index != -1) {
            int x = getColumnDividerPosition(index);
            repaintHeaderView(x - 1, 0, 3, Integer.MAX_VALUE);
            repaintContentView(x - 1, 0, 3, Integer.MAX_VALUE);
        }
    }

    @Override
    public void mousePressed(MouseEvent event) {
        requestFocus();
        Point where = event.getPoint();
        mIgnoreNextDragGesture = false;
        mRowToSelectOnMouseUp = null;
        mViewArea = checkAndConvertToArea(where);
        if (mViewArea != DirectScrollPanelArea.NONE) {
            Point dragStart = new Point(where);
            if (setColumnDividerHighlight(dragStart.x)) {
                mDragColumnDivider = mMouseOverColumnDivider;
            } else {
                boolean isPopupTrigger = event.isPopupTrigger();
                switch (mViewArea) {
                case CONTENT:
                    TreeContainerRow disclosureRow = overDisclosureControl(dragStart.x, dragStart.y);
                    if (disclosureRow != null) {
                        setOpen(!isOpen(disclosureRow), disclosureRow);
                    } else {
                        boolean handled = false;
                        TreeRow row     = overRow(dragStart.y);
                        if (row != null) {
                            TreeColumn column = overColumn(dragStart.x);
                            if (column != null) {
                                handled = column.mousePress(row, new Point(dragStart));
                                if (handled) {
                                    mIgnoreNextDragGesture = true;
                                }
                            }
                            if (!handled) {
                                if (mAnchorRow != null && event.isShiftDown()) {
                                    select(mAnchorRow, row, true);
                                } else if ((event.getModifiersEx() & getToolkit().getMenuShortcutKeyMaskEx()) != 0 && !isPopupTrigger) {
                                    if (isSelected(row)) {
                                        deselect(row);
                                    } else {
                                        select(row, true);
                                    }
                                } else if (!isSelected(row)) {
                                    select(row, false);
                                } else if (mSelectedRows.size() != 1) {
                                    mRowToSelectOnMouseUp = row;
                                }
                            }
                        } else {
                            deselect();
                        }
                        if (!handled && isPopupTrigger) {
                            mRowToSelectOnMouseUp = null;
                            showContextMenuForContent(where);
                        }
                    }
                    break;
                case HEADER:
                    TreeColumn column = overColumn(where.x);
                    if (isPopupTrigger) {
                        if (column != null && mAllowColumnContextMenu) {
                            showContextMenuForColumn(where, column);
                        }
                    } else if (mUserSortable) {
                        mSortColumn = column;
                    }
                    break;
                default:
                    break;
                }
            }
        }
    }

    /**
     * Called to display a context menu for a {@link TreeColumn}.
     *
     * @param where  The point that was clicked, in header area coordinates.
     * @param column The {@link TreeColumn} that was clicked on.
     */
    protected void showContextMenuForColumn(Point where, TreeColumn column) {
        // Does nothing by default.
    }

    /**
     * Called to display a context menu for the content area.
     *
     * @param where The point that was clicked, in content area coordinates.
     */
    protected void showContextMenuForContent(Point where) {
        // Does nothing by default.
    }

    @Override
    public void mouseDragged(MouseEvent event) {
        if (mDragColumnDivider != -1) {
            Point where = event.getPoint();
            mViewArea.convertPoint(this, where);
            dragColumnDividerAndScrollIntoView(where);
        }
    }

    @Override
    public void mouseReleased(MouseEvent event) {
        Rectangle bounds = getBounds();
        bounds.x = 0;
        bounds.y = 0;
        Point   where  = event.getPoint();
        boolean inside = bounds.contains(where);
        mViewArea.convertPoint(this, where);
        if (mDragColumnDivider != -1) {
            mDragColumnDivider = -1;
            if (inside) {
                setColumnDividerHighlight(where.x);
            } else {
                clearColumnDividerHighlight();
            }
        }
        if (mRowToSelectOnMouseUp != null) {
            select(mRowToSelectOnMouseUp, false);
            mRowToSelectOnMouseUp = null;
        }
        if (mSortColumn != null) {
            if (mUserSortable && mSortColumn == overColumn(where.x)) {
                mSorter.setSort(mSortColumn, !event.isShiftDown());
                mSorter.sort(mRoot);
                repaint();
            }
            mSortColumn = null;
        }
    }

    @Override
    public void mouseClicked(MouseEvent event) {
        requestFocus();
        if (!mIgnoreNextDragGesture && event.getClickCount() == 2) {
            Point where = event.getPoint();
            mViewArea = checkAndConvertToArea(where);
            if (mViewArea != DirectScrollPanelArea.NONE) {
                int which = overColumnDivider(where.x);
                if (mAllowColumnResize && which != -1) {
                    sizeColumnToFit(mColumns.get(which));
                } else if (mViewArea == DirectScrollPanelArea.CONTENT && canOpenSelection()) {
                    openSelection();
                }
            }
        }
    }

    /** @return The current {@link TreeSorter}. */
    public TreeSorter getTreeSorter() {
        return mSorter;
    }

    private void dragColumnDividerAndScrollIntoView(Point where) {
        if (mAllowColumnResize) {
            int x = where.x;
            dragColumnDivider(x);
            scrollContentIntoView(x - AUTO_SCROLL_MARGIN, getContentViewBounds().y, AUTO_SCROLL_MARGIN * 2, AUTO_SCROLL_MARGIN * 2);
        }
    }

    private void dragColumnDivider(int x) {
        TreeColumn column = mColumns.get(mDragColumnDivider);
        int        start  = getColumnStart(mDragColumnDivider);
        int        min    = start + column.getMinimumWidth(this);
        if (mShowDisclosureControls && mDragColumnDivider == 0) {
            min += INDENT;
        }
        if (x <= min) {
            x = min + 1;
        }
        x -= start;
        if (x != column.getWidth()) {
            if (mResizeRow == null) {
                mResizeRow = overRow(getContentViewBounds().y);
            }
            column.setWidth(x);
            invalidateAllRowHeights();
            if (isDynamicRowHeightEnabled()) {
                if (!mResizePending) {
                    mResizePending = true;
                    Tasks.scheduleOnUIThread(this, 150, TimeUnit.MILLISECONDS, this);
                }
            } else {
                run();
            }
            repaint();
        }
    }

    /** @return Whether or not the user can sort by clicking in the column header. */
    public final boolean isUserSortable() {
        return mUserSortable;
    }

    /** @param sortable Whether or not the user can sort by clicking in the column header. */
    public final void setUserSortable(boolean sortable) {
        mUserSortable = sortable;
    }

    /** @return Whether or not column resizing by the user is permitted. */
    public final boolean allowColumnResize() {
        return mAllowColumnResize;
    }

    /** @param allow Whether or not column resizing by the user is permitted. */
    public final void setAllowColumnResize(boolean allow) {
        mAllowColumnResize = allow;
    }

    /** @return Whether or not context menus on the column header are permitted. */
    public final boolean allowColumnContextMenu() {
        return mAllowColumnContextMenu;
    }

    /** @param allow Whether or not context menus on the column header are permitted. */
    public final void setAllowColumnContextMenu(boolean allow) {
        mAllowColumnContextMenu = allow;
    }

    /** @return Whether or not column dragging by the user is permitted. */
    public final boolean allowColumnDrag() {
        return mAllowColumnDrag;
    }

    /** @param allow Whether or not column dragging by the user is permitted. */
    public final void setAllowColumnDrag(boolean allow) {
        mAllowColumnDrag = allow;
    }

    /** @return The types of row dragging permitted to be initiated. */
    public final int getAllowedRowDragTypes() {
        return mAllowedRowDragTypes;
    }

    /** @param rowDragTypes The types of row dragging permitted to be initiated. */
    public final void setAllowedRowDragTypes(int rowDragTypes) {
        mAllowedRowDragTypes = rowDragTypes & DnDConstants.ACTION_COPY_OR_MOVE;
    }

    /** @return The types of row dragging permitted to be dropped. */
    public final int getAllowedRowDropTypes() {
        return mAllowedRowDropTypes;
    }

    /** @param rowDropTypes The types of row dragging permitted to be dropped. */
    public final void setAllowedRowDropTypes(int rowDropTypes) {
        mAllowedRowDropTypes = rowDropTypes & DnDConstants.ACTION_COPY_OR_MOVE;
    }

    /** @return Whether or not row drops from external sources is permitted. */
    public final boolean allowRowDropFromExternal() {
        return mAllowRowDropFromExternal;
    }

    /** @param allow Whether or not row drops from external sources is permitted. */
    public final void setAllowRowDropFromExternal(boolean allow) {
        mAllowRowDropFromExternal = allow;
    }

    @Override
    public void focusGained(FocusEvent event) {
        repaint();
    }

    @Override
    public void focusLost(FocusEvent event) {
        repaint();
    }

    /** @return Whether or not dynamic row heights are in use. */
    public final boolean isDynamicRowHeightEnabled() {
        return mRowHeight < 1;
    }

    /**
     * @return The row height for each row. If this is {@code 0}, then each row potentially has a
     *         separate height, as calculated by asking each column to compute a height for each
     *         row.
     */
    public final int getRowHeight() {
        return mRowHeight;
    }

    /**
     * @param height The row height for each row. If this is {@code 0}, then each row may have a
     *               separate height, as calculated by asking each column to compute a height for
     *               each row. <b>WARNING</b>: Using anything other than fixed row heights can cause
     *               severe performance problems when large numbers of rows are visible.
     */
    public final void setRowHeight(int height) {
        if (mRowHeight != height) {
            mRowHeight = height;
            invalidateAllRowHeights();
        }
    }

    /** @return The number of {@link TreeColumn}s. */
    public final int getColumnCount() {
        return mColumns.size();
    }

    /** @return The current {@link TreeColumn}s. */
    public final List<TreeColumn> getColumns() {
        return Collections.unmodifiableList(mColumns);
    }

    /**
     * @param index The index of the {@link TreeColumn} to retrieve.
     * @return The specified {@link TreeColumn}.
     */
    public final TreeColumn getColumn(int index) {
        return mColumns.get(index);
    }

    /** @param column The {@link TreeColumn} to add. */
    public void addColumn(TreeColumn column) {
        mColumns.add(column);
        notify(TreeNotificationKeys.COLUMN_ADDED, new TreeColumn[]{column});
        sizeColumnToFit(column);
    }

    /** @param columns The {@link TreeColumn}s to add. */
    public void addColumn(List<TreeColumn> columns) {
        if (mColumns.addAll(columns)) {
            notify(TreeNotificationKeys.COLUMN_ADDED, columns.toArray(new TreeColumn[0]));
            sizeColumnsToFit(columns);
        }
    }

    /**
     * @param index  The index to insert at.
     * @param column The {@link TreeColumn} to add.
     */
    public void addColumn(int index, TreeColumn column) {
        mColumns.add(index, column);
        notify(TreeNotificationKeys.COLUMN_ADDED, new TreeColumn[]{column});
        sizeColumnToFit(column);
    }

    /**
     * @param index   The index to insert at.
     * @param columns The {@link TreeColumn}s to add.
     */
    public void addColumn(int index, List<TreeColumn> columns) {
        if (mColumns.addAll(index, columns)) {
            notify(TreeNotificationKeys.COLUMN_ADDED, columns.toArray(new TreeColumn[0]));
            sizeColumnsToFit(columns);
        }
    }

    /** @param column The {@link TreeColumn} to remove. */
    public void removeColumn(TreeColumn column) {
        if (mColumns.remove(column)) {
            notify(TreeNotificationKeys.COLUMN_REMOVED, new TreeColumn[]{column});
        }
    }

    /** @param columns The {@link TreeColumn}s to remove. */
    public void removeColumn(Collection<TreeColumn> columns) {
        if (mColumns.removeAll(columns)) {
            notify(TreeNotificationKeys.COLUMN_REMOVED, columns.toArray(new TreeColumn[0]));
        }
    }

    /** @param columns The {@link TreeColumn}s to use. */
    protected void restoreColumns(List<TreeColumn> columns) {
        mColumns.clear();
        mColumns.addAll(columns);
        repaint();
    }

    @Override
    public Dimension getPreferredHeaderSize() {
        int width  = 0;
        int height = 0;
        if (mShowHeader) {
            for (TreeColumn column : mColumns) {
                Dimension headerSize = column.calculatePreferredHeaderSize(this);
                width += headerSize.width;
                height = Math.max(headerSize.height, height);
            }
            int columnDividerWidth = getColumnDividerWidth();
            int size               = mColumns.size() - 1;
            if (size > 0) {
                width += size * columnDividerWidth;
            }
            height += columnDividerWidth;
        }
        return new Dimension(width, height);
    }

    @Override
    public Dimension getPreferredContentSize() {
        int width         = 0;
        int height        = 0;
        int dividerHeight = getRowDividerHeight();
        for (TreeRow row : new TreeRowViewIterator(this, mRoot.getChildren())) {
            height += getRowHeight(row) + dividerHeight;
        }
        for (TreeColumn column : mColumns) {
            width += column.getWidth();
        }
        int size = mColumns.size() - 1;
        if (size > 0) {
            width += size * getColumnDividerWidth();
        }
        return new Dimension(width, height);
    }

    /**
     * @param index The index of the {@link TreeColumn} to check.
     * @return The x-coordinate of the left side of the {@link TreeColumn}.
     */
    public int getColumnStart(int index) {
        int width = index * getColumnDividerWidth();
        for (int i = 0; i < index; i++) {
            width += mColumns.get(i).getWidth();
        }
        return width;
    }

    /**
     * @param column The {@link TreeColumn} to check.
     * @return The x-coordinate of the left side of the {@link TreeColumn}.
     */
    public int getColumnStart(TreeColumn column) {
        int count = mColumns.size();
        int width = 0;
        for (int i = 0; i < count; i++) {
            TreeColumn one = mColumns.get(i);
            if (one == column) {
                return width + getColumnDividerWidth() * i;
            }
            width += one.getWidth();
        }
        return 0;
    }

    /**
     * @param row The {@link TreeRow} to determine the height of.
     * @return The height of the specified {@link TreeRow}.
     */
    public int getRowHeight(TreeRow row) {
        if (mRowHeight > 0) {
            return mRowHeight;
        }
        int height = mRowHeightMap.get(row).intValue();
        if (height == 0) {
            height = calculateHeight(row);
            mRowHeightMap.put(row, Integer.valueOf(height));
        }
        return height;
    }

    private int calculateHeight(TreeRow row) {
        int height = 0;
        int count  = mColumns.size();
        for (int i = 0; i < count; i++) {
            TreeColumn column   = mColumns.get(i);
            int        colWidth = column.getWidth();
            if (mShowDisclosureControls && i == 0) {
                colWidth -= INDENT * row.getDepth();
            }
            int pref = column.calculatePreferredHeight(row, colWidth);
            if (pref > height) {
                height = pref;
            }
        }
        return height;
    }

    /** @param row The {@link TreeRow} to invalidate the cached height of. */
    public void invalidateRowHeight(TreeRow row) {
        if (mRowHeight < 1 && mRowHeightMap.remove(row) != null) {
            notify(TreeNotificationKeys.ROW_HEIGHT, new TreeRow[]{row});
        }
    }

    /** Invalidates the height of all {@link TreeRow}s. */
    public void invalidateAllRowHeights() {
        if (mRowHeight < 1) {
            TreeRow[] rows = mRowHeightMap.keySet().toArray(new TreeRow[0]);
            mRowHeightMap.clear();
            notify(TreeNotificationKeys.ROW_HEIGHT, rows);
        }
    }

    @Override
    public void drawHeader(Graphics2D gc) {
        super.drawHeader(gc);
        boolean   active       = isFocusOwner();
        Rectangle clipBounds   = gc.getClipBounds();
        int       left         = clipBounds.x;
        int       right        = clipBounds.x + clipBounds.width;
        Rectangle colBounds    = getHeaderViewBounds();
        int       dividerWidth = getColumnDividerWidth();
        colBounds.x = 0;
        colBounds.height -= dividerWidth;
        int count = mColumns.size();
        for (int i = 0; i < count; i++) {
            TreeColumn column = mColumns.get(i);
            colBounds.width = i == count - 1 ? Math.max(column.getWidth(), right - colBounds.x) : column.getWidth();
            if (colBounds.x < right && colBounds.x + colBounds.width > left) {
                column.drawHeader(gc, this, colBounds, active);
            }
            colBounds.x += colBounds.width + dividerWidth;
        }
        if (dividerWidth > 0) {
            Color savedColor = gc.getColor();
            gc.setColor(mDividerColor);
            int bottom = colBounds.y + colBounds.height;
            gc.drawLine(left, bottom, right, bottom);
            gc.setColor(savedColor);
            drawColumnDividers(gc);
        }
        if (!isDrawingDragImage() && mDragState != null && mDragState.isHeaderFocus()) {
            Composite savedComposite = gc.getComposite();
            Color     savedColor     = gc.getColor();
            setHighlightColorAndComposite(gc);
            Rectangle bounds = getHeaderViewBounds();
            gc.fillRect(bounds.x + DRAG_FOCUS_WIDTH, bounds.y, bounds.width - DRAG_FOCUS_WIDTH * 2, DRAG_FOCUS_WIDTH);
            if (!mDragState.isContentsFocus()) {
                gc.fillRect(bounds.x + DRAG_FOCUS_WIDTH, bounds.y + bounds.height - DRAG_FOCUS_WIDTH, bounds.width - DRAG_FOCUS_WIDTH * 2, DRAG_FOCUS_WIDTH);
            }
            gc.fillRect(bounds.x, bounds.y, DRAG_FOCUS_WIDTH, bounds.height);
            gc.fillRect(bounds.x + bounds.width - DRAG_FOCUS_WIDTH, bounds.y, DRAG_FOCUS_WIDTH, bounds.height);
            gc.setColor(savedColor);
            gc.setComposite(savedComposite);
        }
    }

    @Override
    public void drawContents(Graphics2D gc) {
        super.drawContents(gc);
        boolean   active = isFocusOwner();
        Rectangle bounds = getContentViewBounds();
        int       min    = bounds.y;
        int       max    = min + bounds.height;
        int       y      = 0;
        if (mResizePending && mResizeRow != null) {
            y = min - getRowBounds(mResizeRow).y;
        }
        int       count            = 0;
        int       dividerHeight    = getRowDividerHeight();
        Rectangle dragClip         = getDragClip();
        boolean   drawingDragImage = isDrawingDragImage();
        for (TreeRow row : new TreeRowViewIterator(this, mRoot.getChildren())) {
            int height = getRowHeight(row) + dividerHeight;
            if (y + height > min) {
                if (drawingDragImage) {
                    if (mSelectedRows.contains(row)) {
                        if (dragClip == null) {
                            dragClip = new Rectangle(bounds.x, y, bounds.width, height);
                        } else {
                            dragClip.height = y + height - dragClip.y;
                        }
                        drawRow(gc, row, count, y, active);
                    }
                } else {
                    drawRow(gc, row, count, y, active);
                }
            }
            y += height;
            if (y >= max) {
                break;
            }
            count++;
        }
        drawColumnDividers(gc);
        if (drawingDragImage) {
            setDragClip(dragClip);
        } else if (mDragState != null) {
            Composite savedComposite = gc.getComposite();
            Color     savedColor     = gc.getColor();
            if (mDragState.isContentsFocus()) {
                setHighlightColorAndComposite(gc);
                if (!mDragState.isHeaderFocus()) {
                    gc.fillRect(bounds.x + DRAG_FOCUS_WIDTH, bounds.y, bounds.width - DRAG_FOCUS_WIDTH * 2, DRAG_FOCUS_WIDTH);
                }
                gc.fillRect(bounds.x + DRAG_FOCUS_WIDTH, bounds.y + bounds.height - DRAG_FOCUS_WIDTH, bounds.width - DRAG_FOCUS_WIDTH * 2, DRAG_FOCUS_WIDTH);
                gc.fillRect(bounds.x, bounds.y, DRAG_FOCUS_WIDTH, bounds.height);
                gc.fillRect(bounds.x + bounds.width - DRAG_FOCUS_WIDTH, bounds.y, DRAG_FOCUS_WIDTH, bounds.height);
            }
            if (mDragState instanceof TreeRowDragState) {
                TreeRowDragState state  = (TreeRowDragState) mDragState;
                TreeContainerRow parent = state.getParentRow();
                if (parent != null) {
                    int index = state.getChildInsertIndex();
                    if (index >= 0) {
                        y = getInsertionMarkerPosition(parent, index);
                        gc.setColor(Color.RED);
                        gc.setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, DRAG_OPACITY));
                        int indent = INDENT * parent.getDepth();
                        if (mShowDisclosureControls) {
                            indent += INDENT;
                        }
                        gc.fillRect(indent, y - DRAG_INSERT_WIDTH / 2, bounds.x + bounds.width - indent, DRAG_INSERT_WIDTH);
                    }
                }
            }
            gc.setColor(savedColor);
            gc.setComposite(savedComposite);
        }
    }

    private void drawColumnDividers(Graphics2D gc) {
        if (mShowColumnDivider) {
            int count = mColumns.size() - 1;
            if (count > 0) {
                Color     savedColor = gc.getColor();
                Rectangle clipBounds = gc.getClipBounds();
                int       left       = clipBounds.x;
                int       right      = left + clipBounds.width;
                int       top        = clipBounds.y;
                int       bottom     = top + clipBounds.height;
                int       x          = 0;
                gc.setColor(mDividerColor);
                for (int i = 0; i < count; i++) {
                    x += mColumns.get(i).getWidth();
                    if (x >= left && x < right) {
                        gc.drawLine(x, top, x, bottom);
                    }
                    x++;
                }
                if (mMouseOverColumnDivider != -1) {
                    x = 0;
                    for (int i = 0; i < count; i++) {
                        x += mColumns.get(i).getWidth();
                        if (i == mMouseOverColumnDivider) {
                            if (x >= left && x < right) {
                                Composite savedComposite = gc.getComposite();
                                gc.setColor(Color.BLACK);
                                gc.setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, 0.25f));
                                gc.drawLine(x - 1, top, x - 1, bottom);
                                gc.drawLine(x + 1, top, x + 1, bottom);
                                gc.setComposite(savedComposite);
                                gc.setColor(Color.BLACK);
                                gc.drawLine(x, top, x, bottom);
                            }
                            break;
                        }
                        x++;
                    }
                }
                gc.setColor(savedColor);
            }
        }
    }

    private static void setHighlightColorAndComposite(Graphics2D gc) {
        gc.setColor(UIManager.getColor("List.selectionBackground"));
        gc.setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, DRAG_OPACITY));
    }

    /**
     * Draws the specified {@link TreeRow}.
     *
     * @param gc       The {@link Graphics2D} context to use.
     * @param row      The {@link TreeRow} to draw.
     * @param position The {@link TreeRow}'s position in the linear view.
     * @param top      The y-coordinate for the top of the {@link TreeRow}.
     * @param active   Whether or not the active state should be displayed.
     */
    public void drawRow(Graphics2D gc, TreeRow row, int position, int top, boolean active) {
        Shape     clip               = gc.getClip();
        Color     savedColor         = gc.getColor();
        int       rowHeight          = getRowHeight(row);
        int       height             = rowHeight + getRowDividerHeight();
        int       x                  = 0;
        Rectangle clipBounds         = gc.getClipBounds();
        int       left               = clipBounds.x;
        int       right              = clipBounds.x + clipBounds.width;
        int       columnDividerWidth = getColumnDividerWidth();
        int       count              = mColumns.size();
        boolean   selected           = isSelected(row);
        gc.setColor(getDefaultRowBackground(position, selected, active));
        gc.fillRect(left, top, clipBounds.width, height);
        if (mShowRowDivider) {
            gc.setColor(mDividerColor);
            int bottom = top + height - 1;
            gc.drawLine(0, bottom, right, bottom);
        }
        Color fg = getDefaultRowForeground(position, selected, active);
        gc.setColor(fg);
        for (int i = 0; i < count; i++) {
            TreeColumn column   = mColumns.get(i);
            int        colWidth = column.getWidth();
            if (i == count - 1) {
                colWidth = Math.max(colWidth, right - x);
            }
            if (x + colWidth > left && x < right) {
                int tmpX = x;
                gc.clipRect(x, top, colWidth, height);
                Composite savedComposite = null;
                if (mSourceDragColumn == column) {
                    savedComposite = gc.getComposite();
                    gc.setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, DRAG_OPACITY));
                }
                if (i == 0) {
                    int depth  = row.getDepth();
                    int indent = INDENT * depth;
                    colWidth -= indent;
                    tmpX += indent;
                    if ((mShowDisclosureControls || depth > 0) && mHierarchyLineColor != null) {
                        TreeRow firstRow = mRoot.getChild(0);
                        if (row != firstRow || mRoot.getChildCount() > 1 || firstRow instanceof TreeContainerRow) {
                            int xx = x + indent - INDENT / 2;
                            gc.setColor(mHierarchyLineColor);
                            gc.drawLine(xx, top + rowHeight / 2, x + indent, top + rowHeight / 2);
                            int yt = row == firstRow ? top + rowHeight / 2 : top;
                            int yb = top + (row.getIndex() == row.getParent().getChildCount() - 1 ? rowHeight / 2 : height);
                            if (yt != yb) {
                                gc.drawLine(xx, yt, xx, yb);
                            }
                            TreeContainerRow one = row.getParent();
                            while (one != mRoot) {
                                xx -= INDENT;
                                if (one.getIndex() != one.getParent().getChildCount() - 1) {
                                    gc.drawLine(xx, top, xx, top + height);
                                }
                                one = one.getParent();
                            }
                            gc.setColor(fg);
                        }
                    }
                    if (mShowDisclosureControls && row instanceof TreeContainerRow) {
                        RetinaIcon icon     = isOpen((TreeContainerRow) row) ? Images.COLLAPSE : Images.EXPAND;
                        int        imgWidth = icon.getIconWidth();
                        int        yt       = rowHeight - icon.getIconHeight();
                        if ((yt & 1) == 1) {
                            yt++;
                        }
                        yt = top + yt / 2;
                        int xt = INDENT - imgWidth;
                        if ((xt & 1) == 0) {
                            xt--;
                        }
                        xt = x + indent - (imgWidth + xt / 2);
                        icon.paintIcon(this, gc, xt, yt);
                    }
                }
                column.draw(gc, this, row, position, top, tmpX, colWidth, selected, active);
                if (mSourceDragColumn == column) {
                    gc.setComposite(savedComposite);
                }
                gc.setClip(clip);
            }
            x += column.getWidth() + columnDividerWidth;
        }
        gc.setColor(savedColor);
    }

    /**
     * @param index The column divider index to retrieve.
     * @return The x-coordinate of the divider.
     */
    public int getColumnDividerPosition(int index) {
        int divider = getColumnDividerWidth();
        int pos     = 0;
        for (int i = 0; i <= index; i++) {
            pos += mColumns.get(i).getWidth() + divider;
        }
        return pos - divider;
    }

    /**
     * @param x The coordinate to check.
     * @return The column divider index, or {@code -1}.
     */
    public int overColumnDivider(int x) {
        int divider = getColumnDividerWidth();
        int count   = mColumns.size() - 1;
        int pos     = 0;
        for (int i = 0; i < count; i++) {
            TreeColumn column = mColumns.get(i);
            pos += column.getWidth();
            if (x >= pos - HIT_SLOP && x <= pos + HIT_SLOP) {
                return i;
            }
            pos += divider;
        }
        return -1;
    }

    /**
     * @param x The x-coordinate to check.
     * @param y The y-coordinate to check.
     * @return The {@link TreeContainerRow} whose disclosure control is at the specified
     *         coordinates, or {@code null}.
     */
    public TreeContainerRow overDisclosureControl(int x, int y) {
        if (mShowDisclosureControls && !mColumns.isEmpty()) {
            if (x < mColumns.get(0).getWidth() + getColumnDividerWidth()) {
                TreeRow row = overRow(y);
                if (row instanceof TreeContainerRow) {
                    int right = INDENT * row.getDepth();
                    if (x <= right && x >= right - INDENT) {
                        return (TreeContainerRow) row;
                    }
                }
            }
        }
        return null;
    }

    /** @param rows The {@link TreeRow}s to repaint. */
    public void repaintRows(TreeRow... rows) {
        repaintRows(Arrays.asList(rows));
    }

    /** @param rows The {@link TreeRow}s to repaint. */
    public void repaintRows(Collection<TreeRow> rows) {
        Set<TreeRow> set           = new HashSet<>(rows);
        Rectangle    bounds        = getContentViewBounds();
        int          min           = bounds.y;
        int          max           = bounds.y + bounds.height;
        int          x             = bounds.x;
        int          width         = bounds.width;
        int          y             = 0;
        int          dividerHeight = getRowDividerHeight();
        for (TreeRow row : new TreeRowViewIterator(this, mRoot.getChildren())) {
            int height = getRowHeight(row) + dividerHeight;
            if (y + height > min) {
                if (set.contains(row)) {
                    repaintContentView(x, y, width, height);
                    set.remove(row);
                    if (set.isEmpty()) {
                        break;
                    }
                }
            }
            y += height;
            if (y >= max) {
                break;
            }
        }
    }

    /**
     * @param row The {@link TreeRow} to determine the bounds of.
     * @return The bounds of the {@link TreeRow} or {@code null} if it is not currently viewable.
     */
    public Rectangle getRowBounds(TreeRow row) {
        int y             = 0;
        int dividerHeight = getRowDividerHeight();
        for (TreeRow one : new TreeRowViewIterator(this, mRoot.getChildren())) {
            int height = getRowHeight(one) + dividerHeight;
            if (one == row) {
                return new Rectangle(0, y, getContentSize().width, height);
            }
            y += height;
        }
        return null;
    }

    /**
     * Finds the {@link TreeColumn} at the specified x-coordinate.
     *
     * @param x The coordinate to check.
     * @return The {@link TreeColumn} at the specified x-coordinate, or {@code null}.
     */
    public TreeColumn overColumn(int x) {
        if (x >= 0 && !mColumns.isEmpty()) {
            int divider = getColumnDividerWidth();
            int count   = mColumns.size() - 1;
            int pos     = 0;
            for (int i = 0; i < count; i++) {
                TreeColumn column = mColumns.get(i);
                pos += column.getWidth() + divider;
                if (x < pos) {
                    return column;
                }
            }
            return mColumns.get(count);
        }
        return null;
    }

    /**
     * Finds the {@link TreeRow} at the specified y-coordinate.
     *
     * @param y The y-coordinate to look for.
     * @return The {@link TreeRow} at the specified y-coordinate, or {@code null}.
     */
    public TreeRow overRow(int y) {
        if (y < 0) {
            return null;
        }
        int dividerHeight = getRowDividerHeight();
        for (TreeRow row : new TreeRowViewIterator(this, mRoot.getChildren())) {
            y -= getRowHeight(row) + dividerHeight;
            if (y < 0) {
                return row;
            }
        }
        return null;
    }

    public boolean showHeader() {
        return mShowHeader;
    }

    public void setShowHeader(boolean visible) {
        if (visible != mShowHeader) {
            mShowHeader = visible;
            notify(TreeNotificationKeys.HEADER, Boolean.valueOf(mShowHeader));
        }
    }

    public boolean showColumnDivider() {
        return mShowColumnDivider;
    }

    public void setShowColumnDivider(boolean visible) {
        if (visible != mShowColumnDivider) {
            mShowColumnDivider = visible;
            notify(TreeNotificationKeys.COLUMN_DIVIDER, Boolean.valueOf(mShowColumnDivider));
        }
    }

    public boolean showRowDivider() {
        return mShowRowDivider;
    }

    public void setShowRowDivider(boolean visible) {
        if (visible != mShowRowDivider) {
            mShowRowDivider = visible;
            notify(TreeNotificationKeys.ROW_DIVIDER, Boolean.valueOf(mShowRowDivider));
        }
    }

    /** @return The height of the row divider. */
    public int getRowDividerHeight() {
        return mShowRowDivider ? 1 : 0;
    }

    /** @return The {@link Color} to use for divider lines. */
    public Color getDividerColor() {
        return mDividerColor;
    }

    /** @param color The {@link Color} to use for divider lines. */
    public void setDividerColor(Color color) {
        mDividerColor = color;
    }

    /** @return The width of the column divider. */
    public int getColumnDividerWidth() {
        return mShowColumnDivider ? 1 : 0;
    }

    /** @return The {@link Color} to use for the hierarchy lines. */
    public Color getHierarchyLineColor() {
        return mHierarchyLineColor;
    }

    /**
     * @param color The {@link Color} to use for the hierarchy lines. If {@code null}, then no
     *              hierarchy lines will be drawn.
     */
    public void setHierarchyLineColor(Color color) {
        mHierarchyLineColor = color;
    }

    /**
     * @param position The {@link TreeRow}'s position in the linear view.
     * @param selected Whether or not the selected version of the color is needed.
     * @param active   Whether or not the active version of the color is needed.
     * @return The foreground color.
     */
    @SuppressWarnings("static-method")
    public Color getDefaultRowForeground(int position, boolean selected, boolean active) {
        return Colors.getListForeground(selected, active);
    }

    /**
     * @param position The {@link TreeRow}'s position in the linear view.
     * @param selected Whether or not the selected version of the color is needed.
     * @param active   Whether or not the active version of the color is needed.
     * @return The background color.
     */
    public Color getDefaultRowBackground(int position, boolean selected, boolean active) {
        if (selected) {
            return Colors.getListBackground(true, active);
        }
        return mUseBanding ? Colors.getBanding(position % 2 == 0) : Color.WHITE;
    }

    /** @return {@code true} if background banding is enabled. */
    public boolean useBanding() {
        return mUseBanding;
    }

    /** @param useBanding Whether to use background banding or not. */
    public void setUseBanding(boolean useBanding) {
        mUseBanding = useBanding;
    }

    /** @return Whether or not the disclosure controls are being shown. */
    public boolean areDisclosureControlsShowing() {
        return mShowDisclosureControls;
    }

    /** @param show Whether or not the disclosure controls are being shown. */
    public void setDisclosureControlsShowing(boolean show) {
        mShowDisclosureControls = show;
    }

    /**
     * Sets all {@link TreeContainerRow}s contained by this {@link TreePanel} to the opposite open
     * state of the first one.
     */
    public void toggleDisclosure() {
        List<TreeContainerRow> rows = mRoot.getRecursiveChildContainers(null);
        if (!rows.isEmpty()) {
            setOpen(!isOpen(rows.get(0)), rows);
        }
    }

    /**
     * @param row The {@link TreeContainerRow} to check.
     * @return Whether the specified {@link TreeContainerRow} is logically 'open', i.e. showing its
     *         children.
     */
    public boolean isOpen(TreeContainerRow row) {
        return mOpenRows.contains(row);
    }

    /**
     * @param open Whether the specified {@link TreeContainerRow}s should be logically 'open', i.e.
     *             showing their children.
     * @param rows The {@link TreeContainerRow}s to disclose.
     */
    public void setOpen(boolean open, TreeContainerRow... rows) {
        setOpen(open, Arrays.asList(rows));
    }

    /**
     * @param open Whether the specified {@link TreeContainerRow}s should be logically 'open', i.e.
     *             showing their children.
     * @param rows The {@link TreeContainerRow}s to disclose.
     */
    public void setOpen(boolean open, Collection<TreeContainerRow> rows) {
        HashSet<TreeContainerRow> modified = new HashSet<>();
        for (TreeContainerRow row : rows) {
            if (row.getTreeRoot() == mRoot) {
                if (mOpenRows.contains(row) != open) {
                    modified.add(row);
                }
            }
        }
        if (!modified.isEmpty()) {
            List<TreeRow>      selectionRemoved = new ArrayList<>();
            TreeContainerRow[] data             = modified.toArray(new TreeContainerRow[0]);
            if (open) {
                mOpenRows.addAll(modified);
            } else {
                for (TreeContainerRow container : data) {
                    for (TreeRow row : new TreeRowViewIterator(this, container.getChildren())) {
                        if (mSelectedRows.contains(row)) {
                            selectionRemoved.add(row);
                        }
                    }
                }
                mOpenRows.removeAll(modified);
            }
            notify(open ? TreeNotificationKeys.ROW_OPENED : TreeNotificationKeys.ROW_CLOSED, data);
            if (!selectionRemoved.isEmpty()) {
                TreeRow[] oldSelection = mSelectedRows.toArray(new TreeRow[0]);
                mSelectedRows.removeAll(selectionRemoved);
                if (mAnchorRow != null && !mSelectedRows.contains(mAnchorRow)) {
                    mAnchorRow = getFirstSelectedRow();
                }
                notify(TreeNotificationKeys.ROW_SELECTION, oldSelection);
            }
            repaintContentView();
            run();
        }
    }

    /**
     * Ensures that each {@link TreeRow} specified will have all of its parents open such that it
     * can be viewed by the user.
     *
     * @param rows The rows whose parents must be open.
     */
    public void setParentsOpen(Collection<TreeRow> rows) {
        Set<TreeContainerRow> parents = new HashSet<>();
        for (TreeRow row : rows) {
            collectClosedParents(row, parents);
        }
        if (!parents.isEmpty()) {
            setOpen(true, parents);
        }
    }

    private void collectClosedParents(TreeRow row, Set<TreeContainerRow> parents) {
        TreeContainerRow parent = row.getParent();
        if (parent != null) {
            if (!isOpen(parent)) {
                parents.add(parent);
            }
            collectClosedParents(parent, parents);
        }
    }

    /**
     * @param row The {@link TreeRow} to check.
     * @return Whether or not the specified {@link TreeRow} is selected.
     */
    public boolean isSelected(TreeRow row) {
        return mSelectedRows.contains(row);
    }

    /**
     * @param row The {@link TreeRow} to check.
     * @return Whether or not the specified {@link TreeRow} or any of its ancestors is selected.
     */
    public boolean isRowOrAncestorSelected(TreeRow row) {
        return mSelectedRows.contains(row) || isAncestorSelected(row);
    }

    /**
     * @param row The {@link TreeRow} to check.
     * @return Whether or not any ancestor of the specified {@link TreeRow} is selected.
     */
    public boolean isAncestorSelected(TreeRow row) {
        row = row.getParent();
        while (row != null) {
            if (mSelectedRows.contains(row)) {
                return true;
            }
            row = row.getParent();
        }
        return false;
    }

    @Override
    public boolean canSelectAll() {
        for (TreeRow row : new TreeRowViewIterator(this, mRoot.getChildren())) {
            if (!mSelectedRows.contains(row)) {
                return true;
            }
        }
        return false;
    }

    @Override
    public void selectAll() {
        TreeRow[] oldSelection = mSelectedRows.toArray(new TreeRow[0]);
        boolean   added        = false;
        mAnchorRow = null;
        for (TreeRow row : new TreeRowViewIterator(this, mRoot.getChildren())) {
            if (mAnchorRow == null) {
                mAnchorRow = row;
            }
            if (mSelectedRows.add(row)) {
                added = true;
            }
        }
        if (added) {
            repaintContentView();
            notify(TreeNotificationKeys.ROW_SELECTION, oldSelection);
        }
    }

    /**
     * Selects the specified {@link TreeRow}s, deselecting any others. The anchor will be set to the
     * top row in the selection.
     *
     * @param rows The {@link TreeRow}s to select.
     */
    public void select(Collection<TreeRow> rows) {
        if (!mSelectedRows.equals(rows)) {
            TreeRow[] oldSelection = mSelectedRows.toArray(new TreeRow[0]);
            mSelectedRows.clear();
            mSelectedRows.addAll(rows);
            mAnchorRow = getFirstSelectedRow();
            repaintRows(mSelectedRows);
            notify(TreeNotificationKeys.ROW_SELECTION, oldSelection);
        }
    }

    /**
     * Selects a specific {@link TreeRow}. If the selection is replaced or there was no prior
     * selection, the anchor is set to the {@link TreeRow} passed in.
     *
     * @param row The {@link TreeRow} to select.
     * @param add Pass in {@code true} to add to the current selection, {@code false} to replace the
     *            current selection.
     */
    public void select(TreeRow row, boolean add) {
        TreeRow[] savedRows = mSelectedRows.toArray(new TreeRow[0]);
        boolean   modified  = false;
        if (!add) {
            modified = !mSelectedRows.isEmpty();
            repaintRows(mSelectedRows);
            mSelectedRows.clear();
        }
        if (mSelectedRows.isEmpty()) {
            mAnchorRow = row;
        }
        if (!mSelectedRows.contains(row)) {
            mSelectedRows.add(row);
            modified = true;
        }
        if (modified) {
            repaintRows(mSelectedRows);
            notify(TreeNotificationKeys.ROW_SELECTION, savedRows);
        }
    }

    /**
     * Selects the range from one {@link TreeRow} to another. If the selection is replaced or there
     * was no prior selection, the anchor is set to the first {@link TreeRow} passed in.
     *
     * @param fromRow The first {@link TreeRow} to select.
     * @param toRow   The last {@link TreeRow} to select.
     * @param add     Pass in {@code true} to add the range to the current selection, {@code false}
     *                to replace the current selection.
     */
    public void select(TreeRow fromRow, TreeRow toRow, boolean add) {
        if (fromRow == toRow) {
            select(fromRow, add);
        } else {
            TreeRow[] savedRows = mSelectedRows.toArray(new TreeRow[0]);
            if (!add) {
                repaintRows(mSelectedRows);
                mSelectedRows.clear();
            }
            if (mSelectedRows.isEmpty()) {
                mAnchorRow = fromRow;
            }
            boolean found = false;
            for (TreeRow row : new TreeRowViewIterator(this, mRoot.getChildren())) {
                if (found) {
                    mSelectedRows.add(row);
                    if (row == toRow) {
                        break;
                    }
                } else if (row == fromRow) {
                    found = true;
                    mSelectedRows.add(row);
                } else if (row == toRow) {
                    found = true;
                    mSelectedRows.add(row);
                    toRow = fromRow;
                }
            }
            repaintRows(mSelectedRows);
            notify(TreeNotificationKeys.ROW_SELECTION, savedRows);
        }
    }

    /** Deselect all. */
    public void deselect() {
        if (!mSelectedRows.isEmpty()) {
            TreeRow[] rows = mSelectedRows.toArray(new TreeRow[0]);
            repaintRows(mSelectedRows);
            mSelectedRows.clear();
            mAnchorRow = null;
            notify(TreeNotificationKeys.ROW_SELECTION, rows);
        }
    }

    /**
     * Deselects a specific {@link TreeRow}.
     *
     * @param row The {@link TreeRow} to deselect.
     */
    public void deselect(TreeRow row) {
        if (mSelectedRows.contains(row)) {
            TreeRow[] rows = mSelectedRows.toArray(new TreeRow[0]);
            mSelectedRows.remove(row);
            if (mAnchorRow == row) {
                mAnchorRow = getFirstSelectedRow();
            }
            repaintRows(row);
            notify(TreeNotificationKeys.ROW_SELECTION, rows);
        }
    }

    /**
     * Does a logical "select up" (such as from the keyboard up-arrow).
     *
     * @param extend Whether to extend the current selection or replace it.
     * @return The {@link TreeRow} that should be scrolled into view, or {@code null} if there isn't
     *         one.
     */
    public TreeRow selectUp(boolean extend) {
        TreeRow scrollTo = null;
        if (mRoot.getChildCount() > 0) {
            int count = mSelectedRows.size();
            if (extend && count > 0) {
                scrollTo = getLastSelectedRow();
                if (scrollTo == mAnchorRow) {
                    scrollTo = getBeforeFirstSelectedRow();
                    if (scrollTo != null) {
                        select(scrollTo, true);
                    }
                } else {
                    deselect(scrollTo);
                    scrollTo = getLastSelectedRow();
                }
            } else {
                if (count == 0) {
                    scrollTo = getLastDisclosedRow();
                } else {
                    scrollTo = count == 1 ? getBeforeFirstSelectedRow() : getFirstSelectedRow();
                    if (scrollTo == null) {
                        scrollTo = getFirstRow();
                    }
                }
                if (scrollTo != null) {
                    select(scrollTo, false);
                }
            }
        }
        return scrollTo;
    }

    /**
     * Does a logical "select down" (such as from the keyboard down-arrow).
     *
     * @param extend Whether to extend the current selection or replace it.
     * @return The {@link TreeRow} that should be scrolled into view, or {@code null} if there isn't
     *         one.
     */
    public TreeRow selectDown(boolean extend) {
        TreeRow scrollTo = null;
        if (mRoot.getChildCount() > 0) {
            int count = mSelectedRows.size();
            if (extend && count > 0) {
                scrollTo = getFirstSelectedRow();
                if (scrollTo == mAnchorRow) {
                    scrollTo = getAfterLastSelectedRow();
                    if (scrollTo != null) {
                        select(scrollTo, true);
                    }
                } else {
                    deselect(scrollTo);
                    scrollTo = getFirstSelectedRow();
                }
            } else {
                if (count == 0) {
                    scrollTo = getFirstRow();
                } else {
                    scrollTo = count == 1 ? getAfterLastSelectedRow() : getLastSelectedRow();
                    if (scrollTo == null) {
                        scrollTo = getLastDisclosedRow();
                    }
                }
                if (scrollTo != null) {
                    select(scrollTo, false);
                }
            }
        }
        return scrollTo;
    }

    /**
     * Does a logical "select home" (such as from the keyboard HOME key).
     *
     * @param extend Whether to extend the current selection or replace it.
     * @return The {@link TreeRow} that should be scrolled into view, or {@code null} if there isn't
     *         one.
     */
    public TreeRow selectToHome(boolean extend) {
        TreeRow scrollTo = null;
        if (mRoot.getChildCount() > 0) {
            scrollTo = getFirstRow();
            if (extend && !mSelectedRows.isEmpty()) {
                if (mAnchorRow == null) {
                    mAnchorRow = getLastSelectedRow();
                }
                select(scrollTo, mAnchorRow, true);
            } else {
                select(scrollTo, false);
            }
        }
        return scrollTo;
    }

    /**
     * Does a logical "select end" (such as from the keyboard END key).
     *
     * @param extend Whether to extend the current selection or replace it.
     * @return The {@link TreeRow} that should be scrolled into view, or {@code null} if there isn't
     *         one.
     */
    public TreeRow selectToEnd(boolean extend) {
        TreeRow scrollTo = null;
        if (mRoot.getChildCount() > 0) {
            scrollTo = getLastDisclosedRow();
            if (extend && !mSelectedRows.isEmpty()) {
                if (mAnchorRow == null) {
                    mAnchorRow = getFirstSelectedRow();
                }
                select(mAnchorRow, scrollTo, true);
            } else {
                select(scrollTo, false);
            }
        }
        return scrollTo;
    }

    /** @return The first {@link TreeRow}, or {@code null}. */
    public TreeRow getFirstRow() {
        List<TreeRow> children = mRoot.getChildren();
        if (!children.isEmpty()) {
            return children.get(0);
        }
        return null;
    }

    /** @return The last disclosed {@link TreeRow}, or {@code null}. */
    public TreeRow getLastDisclosedRow() {
        return getLastChild(mRoot);
    }

    private TreeRow getLastChild(TreeRow row) {
        if (row instanceof TreeContainerRow) {
            TreeContainerRow container = (TreeContainerRow) row;
            if (row instanceof TreeRoot || isOpen(container)) {
                List<TreeRow> children = container.getChildren();
                int           count    = children.size();
                if (count > 0) {
                    return getLastChild(children.get(count - 1));
                }
                return row instanceof TreeRoot ? null : row;
            }
        }
        return row;
    }

    private TreeRow getBeforeFirstSelectedRow() {
        TreeRow prior = null;
        for (TreeRow row : new TreeRowViewIterator(this, mRoot.getChildren())) {
            if (mSelectedRows.contains(row)) {
                return prior;
            }
            prior = row;
        }
        return null;
    }

    /** @return The currently selected {@link TreeRow}s. */
    public List<TreeRow> getExplicitlySelectedRows() {
        List<TreeRow> selection = new ArrayList<>();
        if (!mSelectedRows.isEmpty()) {
            for (TreeRow row : new TreeRowViewIterator(this, mRoot.getChildren())) {
                if (mSelectedRows.contains(row)) {
                    selection.add(row);
                }
            }
        }
        return selection;
    }

    /**
     * @return The currently selected {@link TreeRow}s. If any ancestor of a {@link TreeRow} is
     *         selected, then it will not appear in the result.
     */
    public List<TreeRow> getSelectedRows() {
        List<TreeRow> selection = new ArrayList<>();
        if (!mSelectedRows.isEmpty()) {
            for (TreeRow row : new TreeRowViewIterator(this, mRoot.getChildren())) {
                if (mSelectedRows.contains(row) && !isAncestorSelected(row)) {
                    selection.add(row);
                }
            }
        }
        return selection;
    }

    /** @return The first selected {@link TreeRow}, or {@code null}. */
    public TreeRow getFirstSelectedRow() {
        if (!mSelectedRows.isEmpty()) {
            for (TreeRow row : new TreeRowViewIterator(this, mRoot.getChildren())) {
                if (mSelectedRows.contains(row)) {
                    return row;
                }
            }
        }
        return null;
    }

    /** @return The last selected {@link TreeRow}, or {@code null}. */
    public TreeRow getLastSelectedRow() {
        if (!mSelectedRows.isEmpty()) {
            Set<TreeRow> rows = new HashSet<>(mSelectedRows);
            for (TreeRow row : new TreeRowViewIterator(this, mRoot.getChildren())) {
                if (rows.contains(row)) {
                    rows.remove(row);
                    if (rows.isEmpty()) {
                        return row;
                    }
                }
            }
        }
        return null;
    }

    private TreeRow getAfterLastSelectedRow() {
        if (!mSelectedRows.isEmpty()) {
            Set<TreeRow> rows = new HashSet<>(mSelectedRows);
            for (TreeRow row : new TreeRowViewIterator(this, mRoot.getChildren())) {
                if (rows.isEmpty()) {
                    return row;
                }
                rows.remove(row);
            }
        }
        return null;
    }

    private static List<TreeContainerRow> getTreeContainerRows(Collection<TreeRow> rows, boolean includeChildren) {
        List<TreeContainerRow> result = new ArrayList<>();
        for (TreeRow row : rows) {
            if (row instanceof TreeContainerRow) {
                TreeContainerRow container = (TreeContainerRow) row;
                result.add(container);
                if (includeChildren) {
                    container.getRecursiveChildContainers(result);
                }
            }
        }
        return result;
    }

    @Override
    public void keyPressed(KeyEvent event) {
        if (!event.isConsumed() && (event.getModifiersEx() & getToolkit().getMenuShortcutKeyMaskEx()) == 0) {
            switch (event.getKeyCode()) {
            case KeyEvent.VK_LEFT:
                setOpen(false, getTreeContainerRows(mSelectedRows, event.isAltDown()));
                break;
            case KeyEvent.VK_RIGHT:
                setOpen(true, getTreeContainerRows(mSelectedRows, event.isAltDown()));
                break;
            case KeyEvent.VK_UP:
                keyScroll(selectUp(event.isShiftDown()));
                break;
            case KeyEvent.VK_DOWN:
                keyScroll(selectDown(event.isShiftDown()));
                break;
            case KeyEvent.VK_HOME:
                keyScroll(selectToHome(event.isShiftDown()));
                break;
            case KeyEvent.VK_END:
                keyScroll(selectToEnd(event.isShiftDown()));
                break;
            default:
                return;
            }
            event.consume();
        }
    }

    private void keyScroll(TreeRow row) {
        if (row != null) {
            Rectangle bounds     = getRowBounds(row);
            Rectangle viewBounds = getContentViewBounds();
            bounds.x = viewBounds.x;
            bounds.width = 1;
            scrollContentIntoView(bounds);
        }
    }

    @Override
    public void keyReleased(KeyEvent event) {
        // Not used.
    }

    @Override
    public void keyTyped(KeyEvent event) {
        if (!event.isConsumed() && (event.getModifiersEx() & getToolkit().getMenuShortcutKeyMaskEx()) == 0) {
            char ch = event.getKeyChar();
            if (ch == '\n' || ch == '\r') {
                if (canOpenSelection()) {
                    openSelection();
                }
                event.consume();
            } else if (ch == '\b' || ch == KeyEvent.VK_DELETE) {
                if (canDeleteSelection()) {
                    deleteSelection();
                }
                event.consume();
            }
        }
    }

    /** @return The proxy for responding to {@link Deletable} messages. */
    public Deletable getDeletableProxy() {
        return mDeletableProxy;
    }

    /** @param deletable The proxy for responding to {@link Deletable} messages. */
    public void setDeletableProxy(Deletable deletable) {
        mDeletableProxy = deletable;
    }

    @Override
    public boolean canDeleteSelection() {
        if (mDeletableProxy != null && !mSelectedRows.isEmpty()) {
            return mDeletableProxy.canDeleteSelection();
        }
        return false;
    }

    @Override
    public void deleteSelection() {
        if (mDeletableProxy != null && !mSelectedRows.isEmpty()) {
            mDeletableProxy.deleteSelection();
        }
    }

    /** @return The proxy for responding to {@link Openable} messages. */
    public Openable getOpenableProxy() {
        return mOpenableProxy;
    }

    /** @param openable The proxy for responding to {@link Openable} messages. */
    public void setOpenableProxy(Openable openable) {
        mOpenableProxy = openable;
    }

    @Override
    public boolean canOpenSelection() {
        if (mOpenableProxy != null && !mSelectedRows.isEmpty()) {
            return mOpenableProxy.canOpenSelection();
        }
        return false;
    }

    @Override
    public void openSelection() {
        if (mOpenableProxy != null && !mSelectedRows.isEmpty()) {
            mOpenableProxy.openSelection();
        }
    }

    protected void notify(String key, Object extra) {
        mRoot.getNotifier().notify(this, key, extra);
    }

    @Override
    public int getNotificationPriority() {
        return Integer.MAX_VALUE;
    }

    @Override
    public void handleNotification(Object producer, String name, Object data) {
        if (TreeNotificationKeys.ROW_REMOVED.equals(name)) {
            for (TreeRow row : new TreeRowIterator((TreeRow[]) data)) {
                mSelectedRows.remove(row);
                if (row instanceof TreeContainerRow) {
                    mOpenRows.remove(row);
                }
                if (mRowHeight < 1) {
                    mRowHeightMap.remove(row);
                }
            }
        }
    }

    /** @return The source {@link TreeColumn} being dragged. */
    public TreeColumn getSourceDragColumn() {
        return mSourceDragColumn;
    }

    /** @param column The source {@link TreeColumn} being dragged. */
    protected void setSourceDragColumn(TreeColumn column) {
        if (mSourceDragColumn != null) {
            repaintColumn(mSourceDragColumn);
        }
        mSourceDragColumn = column;
        if (mSourceDragColumn != null) {
            repaintColumn(mSourceDragColumn);
        }
    }

    /** @param column The {@link TreeColumn} to repaint. */
    public void repaintColumn(TreeColumn column) {
        Point pt    = fromHeaderView(new Point(getColumnStart(column), 0));
        int   width = column.getWidth();
        int   count = mColumns.size();
        if (count > 0 && column == mColumns.get(count - 1)) {
            width = Math.max(width, getWidth() - pt.x);
        }
        Rectangle bounds = new Rectangle(pt.x, 0, width, getHeight());
        repaint(bounds);
    }

    private Img createColumnDragImage(TreeColumn column) {
        Graphics2D gc   = null;
        Img        off1 = createImage();
        Img        off2;
        try {
            int width  = column.getWidth();
            int height = getHeight();
            off2 = Img.create(getGraphicsConfiguration(), width, height, Transparency.TRANSLUCENT);
            gc = off2.getGraphics();
            gc.setClip(new Rectangle(0, 0, width, height));
            gc.setBackground(new Color(0, true));
            gc.clearRect(0, 0, width, height);
            gc.setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, DRAG_OPACITY));
            Point pt = fromHeaderView(new Point());
            gc.drawImage(off1, -(getColumnStart(column) + pt.x), 0, this);
        } catch (Exception paintException) {
            Log.error(paintException);
            off2 = null;
        } finally {
            if (gc != null) {
                gc.dispose();
            }
        }
        return off2 != null ? off2 : off1;
    }

    @Override
    public void dragGestureRecognized(DragGestureEvent event) {
        mDropReceived = false;
        int   dragAction = event.getDragAction();
        Point where      = new Point(event.getDragOrigin());
        switch (checkAndConvertToArea(where)) {
        case CONTENT:
            if (!mIgnoreNextDragGesture && mDragColumnDivider == -1 && !mSelectedRows.isEmpty() && (dragAction & mAllowedRowDragTypes) != 0) {
                TreeRowSelection selection = new TreeRowSelection(getSelectedRows(), mOpenRows);
                if (DragSource.isDragImageSupported()) {
                    Img       dragImage = createDragImage(where);
                    Point     imageOffset;
                    Rectangle dragClip  = getDragClip();
                    imageOffset = dragClip != null ? new Point(dragClip.x - where.x, dragClip.y - where.y) : new Point();
                    event.startDrag(null, dragImage, imageOffset, selection, this);
                } else {
                    event.startDrag(null, selection, this);
                }
            }
            break;
        case HEADER:
            if (mAllowColumnDrag && dragAction == DnDConstants.ACTION_MOVE && mSortColumn != null) {
                setSourceDragColumn(mSortColumn);
                if (DragSource.isDragImageSupported()) {
                    event.startDrag(null, createColumnDragImage(mSortColumn), new Point(-(where.x - getColumnStart(mSortColumn)), -where.y), mSortColumn, this);
                } else {
                    event.startDrag(null, mSortColumn, this);
                }
            }
            mSortColumn = null;
            break;
        default:
            break;
        }
    }

    /**
     * @param event The {@link DropTargetDragEvent}.
     * @return A newly created {@link TreeDragState}, or {@code null} if the drag was not
     *         acceptable.
     */
    protected TreeDragState checkDragAcceptability(DropTargetDragEvent event) {
        mAlternateDragDestination = null;
        try {
            if (event.isDataFlavorSupported(TreeColumn.DATA_FLAVOR)) {
                TreeColumn column = (TreeColumn) event.getTransferable().getTransferData(TreeColumn.DATA_FLAVOR);
                if (isColumnDragAcceptable(event, column)) {
                    return new TreeColumnDragState(this, column);
                }
            }
            if (event.isDataFlavorSupported(TreeRowSelection.DATA_FLAVOR)) {
                TreeRowSelection rowSelection = (TreeRowSelection) event.getTransferable().getTransferData(TreeRowSelection.DATA_FLAVOR);
                List<TreeRow>    rows         = rowSelection.getRows();
                if (!rows.isEmpty() && isRowDragAcceptable(event, rows)) {
                    return new TreeRowDragState(this, rowSelection);
                }
            }
            if (event.isDataFlavorSupported(DockableTransferable.DATA_FLAVOR)) {
                mAlternateDragDestination = UIUtilities.getAncestorOfType(this, Dock.class);
            }
        } catch (Exception exception) {
            Log.error(exception);
        }
        return null;
    }

    /**
     * @param event  The {@link DropTargetDragEvent}.
     * @param column The {@link TreeColumn}.
     * @return Whether or not the {@link TreeColumn} can be dropped here.
     */
    protected boolean isColumnDragAcceptable(DropTargetDragEvent event, TreeColumn column) {
        return mColumns.contains(column);
    }

    /**
     * Override to allow external sources of {@link TreeRow} drags.
     *
     * @param event The {@link DropTargetDragEvent}.
     * @param rows  The {@link TreeRow}s.
     * @return Whether or not the {@link TreeRow}s can be dropped here.
     */
    protected boolean isRowDragAcceptable(DropTargetDragEvent event, List<TreeRow> rows) {
        return mAllowRowDropFromExternal || rows.get(0).getTreeRoot() == mRoot;
    }

    @Override
    public void dragEnter(DropTargetDragEvent event) {
        mDragState = checkDragAcceptability(event);
        if (mDragState != null) {
            redrawDragHighlight(mDragState.isHeaderFocus(), mDragState.isContentsFocus());
            mDragState.dragEnter(event);
        } else if (mAlternateDragDestination != null) {
            UIUtilities.convertPoint(event.getLocation(), this, mAlternateDragDestination);
            mAlternateDragDestination.dragEnter(event);
        } else {
            event.rejectDrag();
        }
    }

    @Override
    public void dragOver(DropTargetDragEvent event) {
        if (mDragState != null) {
            mDragState.dragOver(event);
        } else if (mAlternateDragDestination != null) {
            UIUtilities.convertPoint(event.getLocation(), this, mAlternateDragDestination);
            mAlternateDragDestination.dragOver(event);
        } else {
            event.rejectDrag();
        }
    }

    @Override
    public void dropActionChanged(DropTargetDragEvent event) {
        if (mDragState != null) {
            mDragState.dropActionChanged(event);
        } else if (mAlternateDragDestination != null) {
            UIUtilities.convertPoint(event.getLocation(), this, mAlternateDragDestination);
            mAlternateDragDestination.dropActionChanged(event);
        } else {
            event.rejectDrag();
        }
    }

    @Override
    public void dragExit(DropTargetEvent event) {
        if (mDragState != null) {
            mDragState.dragExit(event);
        } else if (mAlternateDragDestination != null) {
            mAlternateDragDestination.dragExit(event);
        }
        clearDragState();
    }

    @Override
    public void drop(DropTargetDropEvent event) {
        if (mDragState != null) {
            event.acceptDrop(event.getDropAction());
            mDropReceived = true;
            event.dropComplete(mDragState.drop(event));
        } else if (mAlternateDragDestination != null) {
            UIUtilities.convertPoint(event.getLocation(), this, mAlternateDragDestination);
            mAlternateDragDestination.drop(event);
        } else {
            mDropReceived = false;
            event.rejectDrop();
        }
        clearDragState();
    }

    private void clearDragState() {
        if (mDragState != null) {
            boolean forHeader   = mDragState.isHeaderFocus();
            boolean forContents = mDragState.isContentsFocus();
            mDragState = null;
            redrawDragHighlight(forHeader, forContents);
        }
        mAlternateDragDestination = null;
    }

    @Override
    public void dragEnter(DragSourceDragEvent event) {
        // Not used.
    }

    @Override
    public void dragOver(DragSourceDragEvent event) {
        // Not used.
    }

    @Override
    public void dropActionChanged(DragSourceDragEvent event) {
        // Not used.
    }

    @Override
    public void dragExit(DragSourceEvent event) {
        // Not used.
    }

    @Override
    public void dragDropEnd(DragSourceDropEvent event) {
        if (event.getDropSuccess()) {
            if (!mDropReceived) {
                if ((event.getDropAction() & mAllowedRowDragTypes) == DnDConstants.ACTION_MOVE) {
                    externalMoveDrop(event.getDragSourceContext().getTransferable());
                }
            }
        }
        setSourceDragColumn(null);
    }

    /**
     * Called when a drag to an external target resolved as an {@link DnDConstants#ACTION_MOVE}. By
     * default, a call to {@link #deleteSelection()} is performed if the {@link Transferable}
     * contains the {@link TreeRowSelection#DATA_FLAVOR}.
     *
     * @param transferable The {@link Transferable}.
     */
    protected void externalMoveDrop(Transferable transferable) {
        if (transferable.isDataFlavorSupported(TreeRowSelection.DATA_FLAVOR)) {
            deleteSelection();
        }
    }

    private void redrawDragHighlight(boolean forHeader, boolean forContents) {
        if (forHeader) {
            Rectangle bounds = getHeaderViewBounds();
            int       x      = bounds.x;
            int       y      = bounds.y;
            int       w      = bounds.width;
            int       h      = bounds.height;
            paintHeaderViewImmediately(x, y, w, DRAG_FOCUS_WIDTH);
            paintHeaderViewImmediately(x, y + h - DRAG_FOCUS_WIDTH, w, DRAG_FOCUS_WIDTH);
            paintHeaderViewImmediately(x, y + DRAG_FOCUS_WIDTH, DRAG_FOCUS_WIDTH, h - DRAG_FOCUS_WIDTH * 2);
            paintHeaderViewImmediately(x + w - DRAG_FOCUS_WIDTH, y + DRAG_FOCUS_WIDTH, DRAG_FOCUS_WIDTH, h - DRAG_FOCUS_WIDTH * 2);
        }
        if (forContents) {
            Rectangle bounds = getContentViewBounds();
            int       x      = bounds.x;
            int       y      = bounds.y;
            int       w      = bounds.width;
            int       h      = bounds.height;
            paintContentViewImmediately(x, y, w, DRAG_FOCUS_WIDTH);
            paintContentViewImmediately(x, y + h - DRAG_FOCUS_WIDTH, w, DRAG_FOCUS_WIDTH);
            paintContentViewImmediately(x, y + DRAG_FOCUS_WIDTH, DRAG_FOCUS_WIDTH, h - DRAG_FOCUS_WIDTH * 2);
            paintContentViewImmediately(x + w - DRAG_FOCUS_WIDTH, y + DRAG_FOCUS_WIDTH, DRAG_FOCUS_WIDTH, h - DRAG_FOCUS_WIDTH * 2);
        }
    }

    /**
     * @param parentRow The parent {@link TreeContainerRow} to insert at.
     * @param insertAt  The child index within the parent to insert at.
     */
    protected void adjustInsertionMarker(TreeContainerRow parentRow, int insertAt) {
        if (mDragState instanceof TreeRowDragState) {
            TreeRowDragState state            = (TreeRowDragState) mDragState;
            TreeContainerRow originalParent   = state.getParentRow();
            int              originalInsertAt = state.getChildInsertIndex();
            if (originalParent != parentRow || originalInsertAt != insertAt) {
                state.setParentRow(parentRow);
                state.setChildInsertIndex(insertAt);
                Rectangle bounds = getContentViewBounds();
                if (originalParent != null && originalInsertAt >= 0) {
                    int y = getInsertionMarkerPosition(originalParent, originalInsertAt);
                    paintContentViewImmediately(bounds.x, y - DRAG_INSERT_WIDTH / 2, bounds.width, DRAG_INSERT_WIDTH);
                }
                if (parentRow != null && insertAt >= 0) {
                    int y = getInsertionMarkerPosition(parentRow, insertAt);
                    paintContentViewImmediately(bounds.x, y - DRAG_INSERT_WIDTH / 2, bounds.width, DRAG_INSERT_WIDTH);
                }
            }
        }
    }

    private int getInsertionMarkerPosition(TreeContainerRow parent, int insertAt) {
        int y             = 0;
        int dividerHeight = getRowDividerHeight();
        if (mRoot != parent) {
            for (TreeRow one : new TreeRowViewIterator(this, mRoot.getChildren())) {
                y += getRowHeight(one) + dividerHeight;
                if (one == parent) {
                    break;
                }
            }
        }
        if (insertAt > 0) {
            for (TreeRow one : new TreeRowViewIterator(this, parent.getChildren())) {
                y += getRowHeight(one) + dividerHeight;
                if (one.getParent() == parent) {
                    if (--insertAt == 0) {
                        break;
                    }
                }
            }
        }
        return y > 0 ? y - 1 : 0;
    }

    /**
     * @param column The {@link TreeColumn} to determine the preferred width of.
     * @return The preferred width.
     */
    public int getPreferredColumnWidth(TreeColumn column) {
        Dimension size    = column.calculatePreferredHeaderSize(this);
        boolean   isFirst = column == mColumns.get(0);
        for (TreeRow row : new TreeRowViewIterator(this, mRoot.getChildren())) {
            int width = column.calculatePreferredWidth(row);
            if (isFirst) {
                width += row.getDepth() * INDENT;
            }
            if (width > size.width) {
                size.width = width;
            }
        }
        return size.width;
    }

    /** @param column The {@link TreeColumn} to set to its preferred width. */
    public void sizeColumnToFit(TreeColumn column) {
        int width = getPreferredColumnWidth(column);
        if (width != column.getWidth()) {
            column.setWidth(width);
            invalidateAllRowHeights();
            run();
            repaint();
        }
    }

    /** @param columns Sets the width of these {@link TreeColumn}s to their preferred width. */
    public void sizeColumnsToFit(Collection<TreeColumn> columns) {
        boolean revalidate = false;
        for (TreeColumn column : columns) {
            int width = getPreferredColumnWidth(column);
            if (width != column.getWidth()) {
                column.setWidth(width);
                revalidate = true;
            }
        }
        if (revalidate) {
            invalidateAllRowHeights();
            run();
            repaint();
        }
    }

    /** Sets the width of all {@link TreeColumn}s to their preferred width. */
    public void sizeColumnsToFit() {
        sizeColumnsToFit(mColumns);
    }

    @Override
    public final void run() {
        mResizePending = false;
        setHeaderAndContentSize(getPreferredHeaderSize(), getPreferredContentSize());
        if (mResizeRow != null) {
            scrollContentToY(getRowBounds(mResizeRow).y);
            mResizeRow = null;
        }
    }
}
