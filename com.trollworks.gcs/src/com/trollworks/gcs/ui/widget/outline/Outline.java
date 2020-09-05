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

package com.trollworks.gcs.ui.widget.outline;

import com.trollworks.gcs.menu.edit.Deletable;
import com.trollworks.gcs.menu.edit.SelectAllCapable;
import com.trollworks.gcs.menu.edit.Undoable;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.Selection;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.ui.print.PrintUtilities;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.widget.ActionPanel;
import com.trollworks.gcs.ui.widget.Icons;
import com.trollworks.gcs.ui.widget.dock.Dock;
import com.trollworks.gcs.ui.widget.dock.DockableTransferable;
import com.trollworks.gcs.utility.Geometry;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.text.Text;

import java.awt.AlphaComposite;
import java.awt.Color;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.GraphicsEnvironment;
import java.awt.Insets;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.Shape;
import java.awt.Transparency;
import java.awt.dnd.Autoscroll;
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DragGestureEvent;
import java.awt.dnd.DragGestureListener;
import java.awt.dnd.DragSource;
import java.awt.dnd.DropTarget;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;
import java.awt.dnd.DropTargetListener;
import java.awt.event.ActionEvent;
import java.awt.event.ComponentEvent;
import java.awt.event.ComponentListener;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import java.util.ArrayList;
import java.util.Collection;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import javax.swing.JScrollPane;
import javax.swing.Scrollable;
import javax.swing.SwingConstants;
import javax.swing.ToolTipManager;
import javax.swing.undo.StateEdit;
import javax.swing.undo.UndoableEdit;

/** A panel that can show both hierarchical and tabular data. */
public class Outline extends ActionPanel implements OutlineModelListener, ComponentListener, FocusListener, Autoscroll, Scrollable, Deletable, SelectAllCapable, DragGestureListener, DropTargetListener, MouseListener, MouseMotionListener, KeyListener {
    /** The default double-click action command. */
    public static final  String            CMD_OPEN_SELECTION                = "Outline.OpenSelection";
    /** The default selection changed action command. */
    public static final  String            CMD_SELECTION_CHANGED             = "Outline.SelectionChanged";
    /** The default potential content size change action command. */
    public static final  String            CMD_POTENTIAL_CONTENT_SIZE_CHANGE = "Outline.ContentSizeMayHaveChanged";
    private static final int               DIVIDER_HIT_SLOP                  = 2;
    private static final int               AUTO_SCROLL_MARGIN                = 10;
    private              OutlineModel      mModel;
    /** The header panel. */
    protected            OutlineHeader     mHeaderPanel;
    private              boolean           mDrawRowDividers;
    private              boolean           mDrawColumnDividers;
    private              boolean           mDrawingDragImage;
    private              Rectangle         mDragClip;
    private              Column            mDividerDrag;
    private              int               mColumnStart;
    private              String            mSelectionChangedCommand;
    private              String            mPotentialContentSizeChangeCommand;
    private              boolean           mAllowColumnResize;
    private              boolean           mAllowRowDrag;
    private              boolean           mUseBanding;
    private              List<Column>      mSavedColumns;
    private              Row               mRollRow;
    private              Row               mDragParentRow;
    private              int               mDragChildInsertIndex;
    private              boolean           mDragWasAcceptable;
    private              boolean           mDragFocus;
    private              boolean           mDynamicRowHeight;
    private              Set<OutlineProxy> mProxies;
    /** The first row index this outline will display. */
    protected            int               mFirstRow;
    /** The last row index this outline will display. */
    protected            int               mLastRow;
    private              int               mSelectOnMouseUp;
    private              boolean           mUserSortable;
    private              boolean           mIgnoreClick;
    private              Deletable         mDeletableProxy;
    private              Dock              mAlternateDragDestination;
    private              String            mLastTooltipText;
    private              int               mLastTooltipX;

    /** Creates a new outline. */
    public Outline() {
        this(true);
    }

    /**
     * Creates a new outline.
     *
     * @param model The model to use.
     */
    public Outline(OutlineModel model) {
        this(model, true);
    }

    /**
     * Creates a new outline.
     *
     * @param showIndent Pass in {@code true} if the outline should show hierarchy and controls for
     *                   it.
     */
    public Outline(boolean showIndent) {
        this(new OutlineModel(), showIndent);
    }

    /**
     * Creates a new outline.
     *
     * @param model      The model to use.
     * @param showIndent Pass in {@code true} if the outline should show hierarchy and controls for
     *                   it.
     */
    public Outline(OutlineModel model, boolean showIndent) {
        mModel = model;
        mProxies = new HashSet<>();
        mUserSortable = true;
        mAllowColumnResize = true;
        mAllowRowDrag = true;
        mDrawRowDividers = true;
        mDrawColumnDividers = true;
        mUseBanding = true;
        mSelectionChangedCommand = CMD_SELECTION_CHANGED;
        mPotentialContentSizeChangeCommand = CMD_POTENTIAL_CONTENT_SIZE_CHANGE;
        mDragChildInsertIndex = -1;
        mLastRow = -1;
        mModel.setShowIndent(showIndent);
        mModel.setIndentWidth(Icons.getDisclosure(true, true).getIconWidth());

        setActionCommand(CMD_OPEN_SELECTION);
        setBackground(Color.white);
        setOpaque(true);
        setFocusable(true);
        addFocusListener(this);
        addMouseListener(this);
        addMouseMotionListener(this);
        addKeyListener(this);
        addComponentListener(this);
        setAutoscrolls(true);
        ToolTipManager.sharedInstance().registerComponent(this);

        if (!GraphicsUtilities.inHeadlessPrintMode() && !GraphicsEnvironment.isHeadless()) {
            DragSource.getDefaultDragSource().createDefaultDragGestureRecognizer(this, DnDConstants.ACTION_COPY_OR_MOVE, this);
            setDropTarget(new DropTarget(this, this));
        }

        if (!(this instanceof OutlineProxy)) {
            mModel.addListener(this);
        }
    }

    /** Causes the {@link RowFilter} to be re-applied. */
    public void reapplyRowFilter() {
        if (mModel.getRowFilter() != null) {
            mModel.reapplyRowFilter();
            revalidateView();
        }
    }

    /** @return The underlying data model. */
    public OutlineModel getModel() {
        return mModel;
    }

    /** @return This outline. */
    public Outline getRealOutline() {
        return this;
    }

    /** @param proxy The proxy to add. */
    protected void addProxy(OutlineProxy proxy) {
        mProxies.add(proxy);
        mModel.addListener(proxy);
    }

    /** Removes all proxies from this outline. */
    public void clearProxies() {
        for (OutlineProxy proxy : mProxies) {
            mModel.removeListener(proxy);
        }
        mProxies.clear();
    }

    /** @return Whether rows will resize vertically when their content changes. */
    public boolean dynamicRowHeight() {
        return mDynamicRowHeight;
    }

    /** @param dynamic Sets whether rows will resize vertically when their content changes. */
    public void setDynamicRowHeight(boolean dynamic) {
        mDynamicRowHeight = dynamic;
    }

    /** @return {@code true} if hierarchy indention (and controls) will be shown. */
    public boolean showIndent() {
        return mModel.showIndent();
    }

    /** @return Whether to draw the row dividers or not. */
    public boolean shouldDrawRowDividers() {
        return mDrawRowDividers;
    }

    /** @param draw Whether to draw the row dividers or not. */
    public void setDrawRowDividers(boolean draw) {
        mDrawRowDividers = draw;
    }

    /** @return Whether to draw the column dividers or not. */
    public boolean shouldDrawColumnDividers() {
        return mDrawColumnDividers;
    }

    /** @param draw Whether to draw the column dividers or not. */
    public void setDrawColumnDividers(boolean draw) {
        mDrawColumnDividers = draw;
    }

    @Override
    public Dimension getPreferredSize() {
        Scale        scale          = Scale.get(this);
        int          one            = scale.scale(1);
        Insets       insets         = getInsets();
        Dimension    size           = new Dimension(insets.left + insets.right, insets.top + insets.bottom);
        List<Column> columns        = mModel.getColumns();
        boolean      needRevalidate = false;

        for (Column col : columns) {
            int width = col.getWidth();
            if (width == -1) {
                width = col.getPreferredWidth(this);
                col.setWidth(this, width);
                needRevalidate = true;
            }
            if (col.isVisible()) {
                size.width += width + (mDrawColumnDividers ? one : 0);
            }
        }
        if (mDrawColumnDividers && !mModel.getColumns().isEmpty()) {
            size.width -= one;
        }

        if (needRevalidate) {
            revalidateView();
        }

        boolean needHeightAdjust = false;
        for (int i = getFirstRowToDisplay(); i <= getLastRowToDisplay(); i++) {
            Row row = mModel.getRowAtIndex(i);
            if (!mModel.isRowFiltered(row)) {
                int height = row.getHeight();
                if (height == -1) {
                    height = row.getPreferredHeight(this, columns);
                    row.setHeight(height);
                }
                size.height += height + (mDrawRowDividers ? one : 0);
                needHeightAdjust = true;
            }
        }
        if (mDrawRowDividers && needHeightAdjust) {
            size.height -= one;
        }

        if (isMinimumSizeSet()) {
            Dimension minSize = getMinimumSize();
            if (size.width < minSize.width) {
                size.width = minSize.width;
            }
            if (size.height < minSize.height) {
                size.height = minSize.height;
            }
        }
        return size;
    }

    private void drawDragRowInsertionMarker(Graphics gc, Row parent, int insertAtIndex) {
        Scale     scale  = Scale.get(this);
        int       one    = scale.scale(1);
        Rectangle bounds = getDragRowInsertionMarkerBounds(parent, insertAtIndex);
        gc.setColor(Color.red);
        int three = scale.scale(3);
        gc.fillRect(bounds.x, bounds.y + three, bounds.width, one);
        int height = bounds.height;
        for (int i = 0; i < three; i++) {
            gc.fillRect(bounds.x + i, bounds.y + i, 1, height);
            height -= 2;
        }
    }

    private Rectangle getDragRowInsertionMarkerBounds(Row parent, int insertAtIndex) {
        Scale     scale    = Scale.get(this);
        int       one      = scale.scale(1);
        int       rowCount = mModel.getRowCount();
        Rectangle bounds;

        if (insertAtIndex < 0 || rowCount == 0) {
            bounds = new Rectangle();
        } else {
            int    insertAt = getAbsoluteInsertionIndex(parent, insertAtIndex);
            Column col      = mModel.getHierarchyColumn();
            int    indent   = getColumnStart(col) + scale.scale(mModel.getIndentWidth()) + (parent != null ? scale.scale(mModel.getIndentWidth(parent, col)) : 0);
            if (insertAt != -1 && insertAt < rowCount) {
                bounds = getRowBounds(mModel.getRowAtIndex(insertAt));
                if (mDrawRowDividers && insertAt != 0) {
                    bounds.y -= one;
                }
            } else {
                bounds = getRowBounds(mModel.getRowAtIndex(rowCount - 1));
                bounds.y += bounds.height;
            }
            bounds.x += indent;
            bounds.width -= indent;
        }
        int three = scale.scale(3);
        bounds.y -= three;
        bounds.height = three + one + three;
        return bounds;
    }

    /** @return The first row to display. By default, this would be 0. */
    public int getFirstRowToDisplay() {
        return Math.max(mFirstRow, 0);
    }

    /** @param index The first row to display. */
    public void setFirstRowToDisplay(int index) {
        mFirstRow = index;
    }

    /** @return The last row to display. */
    public int getLastRowToDisplay() {
        int max = mModel.getRowCount() - 1;
        return mLastRow < 0 || mLastRow > max ? max : mLastRow;
    }

    /**
     * @param index The last row to display. If set to a negative value, then the last row index in
     *              the outline will be returned from {@link #getLastRowToDisplay()}.
     */
    public void setLastRowToDisplay(int index) {
        mLastRow = index;
    }

    @Override
    protected void paintComponent(Graphics gc) {
        Scale scale = Scale.get(this);
        int   one   = scale.scale(1);

        super.paintComponent(GraphicsUtilities.prepare(gc));
        drawBackground(gc);

        Shape     origClip   = gc.getClip();
        Rectangle clip       = gc.getClipBounds();
        Insets    insets     = getInsets();
        Rectangle bounds     = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
        boolean   active     = isFocusOwner();
        int       first      = getFirstRowToDisplay();
        int       last       = getLastRowToDisplay();
        boolean   isPrinting = PrintUtilities.isPrinting(this);
        boolean   showIndent = showIndent();

        for (int rowIndex = first; rowIndex <= last; rowIndex++) {
            Row row = mModel.getRowAtIndex(rowIndex);
            if (!mModel.isRowFiltered(row)) {
                bounds.height = row.getHeight();
                if (bounds.y >= clip.y || bounds.y + bounds.height + (mDrawRowDividers ? one : 0) >= clip.y) {
                    if (bounds.y > clip.y + clip.height) {
                        break;
                    }

                    boolean rowSelected = !isPrinting && mModel.isRowSelected(row);
                    if (!mDrawingDragImage || rowSelected) {
                        Rectangle colBounds = new Rectangle(bounds);
                        int       shift     = 0;

                        for (Column col : mModel.getColumns()) {
                            if (col.isVisible()) {
                                colBounds.width = col.getWidth();
                                if (clip.intersects(colBounds)) {
                                    gc.clipRect(colBounds.x, colBounds.y, colBounds.width, colBounds.height);
                                    boolean isHierCol = mModel.isHierarchyColumn(col);
                                    if (showIndent && isHierCol) {
                                        shift = scale.scale(mModel.getIndentWidth(row, col));
                                        colBounds.x += shift;
                                        colBounds.width -= shift;
                                        if (row.canHaveChildren()) {
                                            RetinaIcon disclosure = getDisclosureControl(row);
                                            disclosure.paintIcon(this, gc, colBounds.x - scale.scale(disclosure.getIconWidth()), colBounds.y + (colBounds.height - scale.scale(disclosure.getIconHeight())) / 2);
                                        }
                                    }
                                    col.drawRowCell(this, gc, colBounds, row, rowSelected, active);
                                    if (showIndent && isHierCol) {
                                        colBounds.x -= shift;
                                        colBounds.width += shift;
                                    }
                                    gc.setClip(origClip);
                                }
                                colBounds.x += colBounds.width + one;
                            }
                        }

                        if (mDrawingDragImage) {
                            if (mDragClip == null) {
                                mDragClip = new Rectangle(bounds);
                            } else {
                                mDragClip.add(bounds);
                            }
                        }
                    }
                }
                bounds.y += bounds.height + (mDrawRowDividers ? one : 0);
            }
        }

        if (mDragChildInsertIndex != -1) {
            drawDragRowInsertionMarker(gc, mDragParentRow, mDragChildInsertIndex);
        }
        Row dragTargetRow = getDragTargetRow();
        if (dragTargetRow != null) {
            Graphics2D g2d = (Graphics2D) gc;
            g2d.setColor(Color.RED);
            g2d.draw(Geometry.inset(1, getRowBounds(dragTargetRow)));
        }
    }

    private void drawBackground(Graphics gc) {
        Scale scale = Scale.get(this);
        int   one   = scale.scale(1);

        super.paintComponent(gc);

        Rectangle clip       = gc.getClipBounds();
        Insets    insets     = getInsets();
        int       top        = insets.top;
        int       bottom     = getHeight() - (top + insets.bottom);
        Rectangle bounds     = new Rectangle(insets.left, top, getWidth() - (insets.left + insets.right), bottom);
        boolean   active     = isFocusOwner();
        int       first      = getFirstRowToDisplay();
        int       last       = getLastRowToDisplay();
        boolean   isPrinting = PrintUtilities.isPrinting(this);

        for (int rowIndex = first; rowIndex <= last; rowIndex++) {
            Row row = mModel.getRowAtIndex(rowIndex);
            if (!mModel.isRowFiltered(row)) {
                bounds.height = row.getHeight();
                if (bounds.y >= clip.y || bounds.y + bounds.height + (mDrawRowDividers ? one : 0) >= clip.y) {
                    if (bounds.y > clip.y + clip.height) {
                        break;
                    }
                    boolean rowSelected = !isPrinting && mModel.isRowSelected(row);
                    if (!mDrawingDragImage || rowSelected) {
                        gc.setColor(getBackground(rowIndex, rowSelected, active));
                        gc.fillRect(bounds.x, bounds.y, bounds.width, bounds.height);
                        if (mDrawRowDividers) {
                            gc.setColor(ThemeColor.DIVIDER);
                            gc.fillRect(bounds.x, bounds.y + bounds.height, bounds.width, one);
                        }
                    }
                }
                bounds.y += bounds.height + (mDrawRowDividers ? one : 0);
            }
        }

        if (mDrawColumnDividers) {
            int x = insets.left;
            gc.setColor(ThemeColor.DIVIDER);
            List<Column> columns = mModel.getColumns();
            int          count   = columns.size() - 1;
            while (count > 0 && !columns.get(count).isVisible()) {
                count--;
            }
            for (int i = 0; i < count; i++) {
                Column col = columns.get(i);
                if (col.isVisible()) {
                    x += col.getWidth();
                    gc.fillRect(x, top, one, bottom);
                    x += one;
                }
            }
        }

        if (!isPrinting && mDragFocus) {
            gc.setColor(Colors.getListBackground(true, true));
            bounds = getVisibleRect();
            gc.fillRect(bounds.x, bounds.y, bounds.width, one);
            gc.fillRect(bounds.x, bounds.y + bounds.height - one, bounds.width, one);
            gc.fillRect(bounds.x, bounds.y + one, one, bounds.height - (one + one));
            gc.fillRect(bounds.x + bounds.width - one, bounds.y + one, one, bounds.height - (one + one));
        }
    }

    @Override
    public void repaint(Rectangle bounds) {
        super.repaint(bounds);
        // We have to check for null here, since repaint() will be called during initialization of
        // our super class.
        if (mProxies != null) {
            for (OutlineProxy proxy : mProxies) {
                proxy.repaintProxy(bounds);
            }
        }
    }

    /**
     * Repaints the header panel, along with any proxy header panels.
     *
     * @param bounds The bounds to repaint.
     */
    void repaintHeader(Rectangle bounds) {
        if (mHeaderPanel != null) {
            getHeaderPanel().repaintInternal(bounds);
            for (OutlineProxy proxy : mProxies) {
                if (proxy.mHeaderPanel != null) {
                    proxy.getHeaderPanel().repaintInternal(bounds);
                }
            }
        }
    }

    /** Repaints the header panel, if present. */
    void repaintHeader() {
        if (mHeaderPanel != null) {
            Rectangle bounds = mHeaderPanel.getBounds();
            bounds.x = 0;
            bounds.y = 0;
            mHeaderPanel.repaintInternal(bounds);
        }
    }

    /**
     * Repaints the specified column index.
     *
     * @param columnIndex The index of the column to repaint.
     */
    public void repaintColumn(int columnIndex) {
        repaintColumn(mModel.getColumnAtIndex(columnIndex));
    }

    /**
     * Repaints the specified column.
     *
     * @param column The column to repaint.
     */
    public void repaintColumn(Column column) {
        if (column.isVisible()) {
            Rectangle bounds = new Rectangle(getColumnStart(column), 0, column.getWidth(), getHeight());
            repaint(bounds);
            if (mHeaderPanel != null) {
                bounds.height = mHeaderPanel.getHeight();
                mHeaderPanel.repaint(bounds);
            }
        }
    }

    /** Repaints both the outline and its header. */
    public void repaintView() {
        repaint();
        if (mHeaderPanel != null) {
            mHeaderPanel.repaint();
        }
    }

    /** Repaints the current selection. */
    public void repaintSelection() {
        repaintSelectionInternal();
        for (OutlineProxy proxy : mProxies) {
            proxy.repaintSelectionInternal();
        }
    }

    /** Repaints the current selection. */
    protected void repaintSelectionInternal() {
        Scale        scale   = Scale.get(this);
        int          one     = scale.scale(1);
        Insets       insets  = getInsets();
        Rectangle    bounds  = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
        int          last    = getLastRowToDisplay();
        List<Column> columns = mModel.getColumns();
        for (int i = getFirstRowToDisplay(); i <= last; i++) {
            Row row = mModel.getRowAtIndex(i);
            if (!mModel.isRowFiltered(row)) {
                int height = row.getHeight();
                if (height == -1) {
                    height = row.getPreferredHeight(this, columns);
                    row.setHeight(height);
                }
                if (mDrawRowDividers) {
                    height += one;
                }
                if (mModel.isRowSelected(row)) {
                    bounds.height = height;
                    repaint(bounds);
                }
                bounds.y += height;
            }
        }
    }

    /**
     * @param rowIndex The index of the row.
     * @param selected Whether the row should be considered "selected".
     * @param active   Whether the outline should be considered "active".
     * @return The background color for the specified row index.
     */
    public Color getBackground(int rowIndex, boolean selected, boolean active) {
        if (selected) {
            return Colors.getListBackground(true, active);
        }
        return (useBanding() && (rowIndex % 2 != 0)) ? ThemeColor.BANDING : Color.WHITE;
    }

    @Override
    public int getScrollableBlockIncrement(Rectangle visibleRect, int orientation, int direction) {
        return orientation == SwingConstants.VERTICAL ? visibleRect.height : visibleRect.width;
    }

    @Override
    public Dimension getPreferredScrollableViewportSize() {
        return getPreferredSize();
    }

    @Override
    public int getScrollableUnitIncrement(Rectangle visibleRect, int orientation, int direction) {
        Scale scale = Scale.get(this);
        if (orientation == SwingConstants.VERTICAL) {
            Insets insets = getInsets();
            int    y      = visibleRect.y - insets.top;
            int    rowIndex;
            int    rowTop;
            if (direction < 0) {
                rowIndex = overRowIndex(y);
                if (rowIndex > -1) {
                    rowTop = getRowIndexStart(rowIndex);
                    if (rowTop < y) {
                        return y - rowTop;
                    } else {
                        List<Row> rows  = mModel.getRows();
                        int       first = getFirstRowToDisplay();
                        do {
                            if (--rowIndex <= first) {
                                break;
                            }
                        } while (mModel.isRowFiltered(rows.get(rowIndex)));
                        if (rowIndex >= first) {
                            return y - getRowIndexStart(rowIndex);
                        }
                    }
                }
            } else {
                y += visibleRect.height;
                rowIndex = overRowIndex(y);
                if (rowIndex > -1) {
                    int one = scale.scale(1);
                    rowTop = getRowIndexStart(rowIndex);
                    int rowBottom = rowTop + mModel.getRowAtIndex(rowIndex).getHeight() + (mDrawRowDividers ? one : 0);
                    if (rowBottom > y) {
                        return rowBottom - (y - one);
                    } else if (++rowIndex < mModel.getRowCount()) {
                        return getRowIndexStart(rowIndex) + mModel.getRowAtIndex(rowIndex).getHeight() + (mDrawRowDividers ? one : 0) - (y - one);
                    }
                }
            }
        }
        return scale.scale(10);
    }

    @Override
    public boolean getScrollableTracksViewportHeight() {
        return UIUtilities.shouldTrackViewportHeight(this);
    }

    @Override
    public boolean getScrollableTracksViewportWidth() {
        return UIUtilities.shouldTrackViewportWidth(this);
    }

    /**
     * Determines if the specified x-coordinate is over a column's divider.
     *
     * @param x The coordinate to check.
     * @return The column, or {@code null} if none is found.
     */
    public Column overColumnDivider(int x) {
        Scale scale = Scale.get(this);
        int   slop  = scale.scale(DIVIDER_HIT_SLOP);
        int   one   = scale.scale(1);
        int   pos   = getInsets().left;
        int   count = mModel.getColumnCount();
        for (int i = 0; i < count; i++) {
            Column col = mModel.getColumnAtIndex(i);
            if (col.isVisible()) {
                pos += col.getWidth() + (mDrawColumnDividers ? one : 0);
                if (x >= pos - slop && x <= pos + slop) {
                    return col;
                }
            }
        }
        return null;
    }

    /**
     * Determines if the specified x-coordinate is over a column's divider.
     *
     * @param x The coordinate to check.
     * @return The column index, or {@code -1} if none is found.
     */
    public int overColumnDividerIndex(int x) {
        Scale scale = Scale.get(this);
        int   slop  = scale.scale(DIVIDER_HIT_SLOP);
        int   one   = scale.scale(1);
        int   pos   = getInsets().left;
        int   count = mModel.getColumnCount();
        for (int i = 0; i < count; i++) {
            Column col = mModel.getColumnAtIndex(i);
            if (col.isVisible()) {
                pos += col.getWidth() + (mDrawColumnDividers ? one : 0);
                if (x >= pos - slop && x <= pos + slop) {
                    return i;
                }
            }
        }
        return -1;
    }

    /**
     * Determines if the specified x-coordinate is over a column.
     *
     * @param x The coordinate to check.
     * @return The column, or {@code null} if none is found.
     */
    public Column overColumn(int x) {
        int one   = Scale.get(this).scale(1);
        int pos   = getInsets().left;
        int count = mModel.getColumnCount();
        for (int i = 0; i < count; i++) {
            Column col = mModel.getColumnAtIndex(i);
            if (col.isVisible()) {
                pos += col.getWidth() + (mDrawColumnDividers ? one : 0);
                if (x < pos) {
                    return col;
                }
            }
        }
        return null;
    }

    /**
     * Determines if the specified x-coordinate is over a column.
     *
     * @param x The coordinate to check.
     * @return The column index, or {@code -1} if none is found.
     */
    public int overColumnIndex(int x) {
        int one   = Scale.get(this).scale(1);
        int pos   = getInsets().left;
        int count = mModel.getColumnCount();
        for (int i = 0; i < count; i++) {
            Column col = mModel.getColumnAtIndex(i);
            if (col.isVisible()) {
                pos += col.getWidth() + (mDrawColumnDividers ? one : 0);
                if (x < pos) {
                    return i;
                }
            }
        }
        return -1;
    }

    private RetinaIcon getDisclosureControl(Row row) {
        return Icons.getDisclosure(row.isOpen(), row == mRollRow);
    }

    /**
     * @param x      The x-coordinate.
     * @param y      The y-coordinate.
     * @param column The column the coordinates are currently over.
     * @param row    The row the coordinates are currently over.
     * @return {@code true} if the coordinates are over a disclosure triangle.
     */
    public boolean overDisclosureControl(int x, int y, Column column, Row row) {
        if (showIndent() && column != null && row != null && row.canHaveChildren() && mModel.isHierarchyColumn(column)) {
            Scale scale = Scale.get(this);
            int   right = getColumnStart(column) + scale.scale(mModel.getIndentWidth(row, column));
            return x <= right && x >= right - scale.scale(getDisclosureControl(row).getIconWidth());
        }
        return false;
    }

    /**
     * @param columnIndex The index of the column.
     * @return The starting x-coordinate for the specified column index.
     */
    public int getColumnIndexStart(int columnIndex) {
        int one = Scale.get(this).scale(1);
        int pos = getInsets().left;
        for (int i = 0; i < columnIndex; i++) {
            Column column = mModel.getColumnAtIndex(i);
            if (column.isVisible()) {
                pos += column.getWidth() + (mDrawColumnDividers ? one : 0);
            }
        }
        return pos;
    }

    /**
     * @param column The column.
     * @return The starting x-coordinate for the specified column.
     */
    public int getColumnStart(Column column) {
        int one   = Scale.get(this).scale(1);
        int pos   = getInsets().left;
        int count = mModel.getColumnCount();
        for (int i = 0; i < count; i++) {
            Column col = mModel.getColumnAtIndex(i);
            if (col == column) {
                break;
            }
            if (col.isVisible()) {
                pos += col.getWidth() + (mDrawColumnDividers ? one : 0);
            }
        }
        return pos;
    }

    /**
     * @param column The column.
     * @return An {@link Img} containing the drag image for the specified column.
     */
    public Img getColumnDragImage(Column column) {
        Img offscreen = null;
        synchronized (getTreeLock()) {
            Scale      scale   = Scale.get(this);
            int        one     = scale.scale(1);
            int        twoOnes = one + one;
            Graphics2D gc      = null;
            try {
                Rectangle bounds = new Rectangle(0, 0, column.getWidth() + (mDrawColumnDividers ? twoOnes : 0), getVisibleRect().height + (mHeaderPanel != null ? mHeaderPanel.getHeight() + one : 0));
                offscreen = Img.create(getGraphicsConfiguration(), bounds.width, bounds.height, Transparency.TRANSLUCENT);
                gc = GraphicsUtilities.prepare(offscreen.getGraphics());
                gc.setClip(bounds);
                gc.setBackground(new Color(0, true));
                gc.clearRect(0, 0, bounds.width, bounds.height);
                gc.setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, 0.5f));
                gc.setColor(getBackground());
                gc.fill(bounds);
                gc.setColor(ThemeColor.DIVIDER);
                if (mDrawRowDividers) {
                    gc.fillRect(bounds.x, bounds.y, bounds.width, one);
                    gc.fillRect(bounds.x, bounds.y + bounds.height - one, bounds.width, one);
                    bounds.y += one;
                    bounds.height -= twoOnes;
                }
                if (mDrawColumnDividers) {
                    gc.fillRect(bounds.x, bounds.y, one, bounds.height);
                    gc.fillRect(bounds.x + bounds.width - one, bounds.y, one, bounds.height);
                    bounds.x += one;
                    bounds.width -= twoOnes;
                }
                drawOneColumn(gc, column, bounds);
            } catch (Exception exception) {
                Log.error(exception);
            } finally {
                if (gc != null) {
                    gc.dispose();
                }
            }
        }
        return offscreen;
    }

    private void drawOneColumn(Graphics2D g2d, Column column, Rectangle bounds) {
        Scale scale   = Scale.get(this);
        int   one     = scale.scale(1);
        Shape oldClip = g2d.getClip();
        int   last    = getLastRowToDisplay();
        int   y       = bounds.y;
        int   maxY    = bounds.y + bounds.height;

        if (mHeaderPanel != null) {
            bounds.height = mHeaderPanel.getHeight();
            g2d.setColor(mHeaderPanel.getBackground());
            g2d.fill(bounds);
            column.drawHeaderCell(this, g2d, bounds);
            bounds.y += mHeaderPanel.getHeight();
            g2d.setColor(ThemeColor.DIVIDER);
            g2d.fillRect(bounds.x, bounds.y, bounds.width, one);
            bounds.y += one;
        }

        for (int i = getFirstRowToDisplay(); i <= last; i++) {
            Row row = mModel.getRowAtIndex(i);
            if (!mModel.isRowFiltered(row)) {
                bounds.height = row.getHeight();
                if (maxY < bounds.y) {
                    break;
                }
                if (y <= bounds.y) {
                    g2d.setClip(bounds);
                    g2d.setColor(getBackground(i, false, true));
                    g2d.fill(bounds);
                    column.drawRowCell(this, g2d, bounds, row, false, true);
                    g2d.setClip(oldClip);
                    if (mDrawRowDividers) {
                        g2d.setColor(ThemeColor.DIVIDER);
                        g2d.fillRect(bounds.x, bounds.y + bounds.height, bounds.width, one);
                    }
                }
                bounds.y += bounds.height + (mDrawRowDividers ? one : 0);
            }
        }
    }

    /**
     * Determines if the specified y-coordinate is over a row.
     *
     * @param y The coordinate to check.
     * @return The row, or {@code null} if none is found.
     */
    public Row overRow(int y) {
        int       one  = Scale.get(this).scale(1);
        List<Row> rows = mModel.getRows();
        int       pos  = getInsets().top;
        int       last = getLastRowToDisplay();
        for (int i = getFirstRowToDisplay(); i <= last; i++) {
            Row row = rows.get(i);
            if (!mModel.isRowFiltered(row)) {
                pos += row.getHeight() + (mDrawRowDividers ? one : 0);
                if (y < pos) {
                    return row;
                }
            }
        }
        return null;
    }

    /**
     * Determines if the specified y-coordinate is over a row.
     *
     * @param y The coordinate to check.
     * @return The row index, or {@code -1} if none is found.
     */
    public int overRowIndex(int y) {
        int       one  = Scale.get(this).scale(1);
        List<Row> rows = mModel.getRows();
        int       pos  = getInsets().top;
        int       last = getLastRowToDisplay();
        for (int i = getFirstRowToDisplay(); i <= last; i++) {
            Row row = rows.get(i);
            if (!mModel.isRowFiltered(row)) {
                pos += row.getHeight() + (mDrawRowDividers ? one : 0);
                if (y < pos) {
                    return i;
                }
            }
        }
        return -1;
    }

    /**
     * @param y The coordinate to check.
     * @return The row index to insert at, from {@code 0} to {@link OutlineModel#getRowCount()} .
     */
    public int getRowInsertionIndex(int y) {
        int       one  = Scale.get(this).scale(1);
        List<Row> rows = mModel.getRows();
        int       pos  = getInsets().top;
        int       last = getLastRowToDisplay();
        for (int i = getFirstRowToDisplay(); i <= last; i++) {
            Row row = rows.get(i);
            if (!mModel.isRowFiltered(row)) {
                int height = row.getHeight();
                int tmp    = pos + height / 2;
                if (y <= tmp) {
                    return i;
                }
                pos += height + (mDrawRowDividers ? one : 0);
            }
        }
        return last;
    }

    /**
     * @param index The index of the row.
     * @return The starting y-coordinate for the specified row index.
     */
    public int getRowIndexStart(int index) {
        int       one  = Scale.get(this).scale(1);
        List<Row> rows = mModel.getRows();
        int       pos  = getInsets().top;
        for (int i = getFirstRowToDisplay(); i < index; i++) {
            Row row = rows.get(i);
            if (!mModel.isRowFiltered(row)) {
                pos += row.getHeight() + (mDrawRowDividers ? one : 0);
            }
        }
        return pos;
    }

    /**
     * @param row The row.
     * @return The starting y-coordinate for the specified row.
     */
    public int getRowStart(Row row) {
        int       one  = Scale.get(this).scale(1);
        List<Row> rows = mModel.getRows();
        int       pos  = getInsets().top;
        int       last = getLastRowToDisplay();
        for (int i = getFirstRowToDisplay(); i <= last; i++) {
            Row oneRow = rows.get(i);
            if (row == oneRow) {
                break;
            }
            if (!mModel.isRowFiltered(oneRow)) {
                pos += oneRow.getHeight() + (mDrawRowDividers ? one : 0);
            }
        }
        return pos;
    }

    /**
     * @param rowIndex The index of the row.
     * @return The bounds of the row at the specified index.
     */
    public Rectangle getRowIndexBounds(int rowIndex) {
        Insets    insets = getInsets();
        Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
        bounds.y = getRowIndexStart(rowIndex);
        bounds.height = mModel.getRowAtIndex(rowIndex).getHeight();
        return bounds;
    }

    /**
     * @param row The row.
     * @return The bounds of the specified row.
     */
    public Rectangle getRowBounds(Row row) {
        Insets    insets = getInsets();
        Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
        bounds.y = getRowStart(row);
        bounds.height = row.getHeight();
        return bounds;
    }

    /**
     * @param row    The row to use.
     * @param column The column to use.
     * @return The bounds of the specified cell.
     */
    public Rectangle getCellBounds(Row row, Column column) {
        Rectangle bounds = getRowBounds(row);
        bounds.x = getColumnStart(column);
        bounds.width = column.getWidth();
        return bounds;
    }

    /**
     * @param row    The row to use.
     * @param column The column to use.
     * @return The bounds of the specified cell, adjusted for any necessary indent.
     */
    public Rectangle getAdjustedCellBounds(Row row, Column column) {
        Rectangle bounds = getCellBounds(row, column);
        if (mModel.isHierarchyColumn(column)) {
            Scale scale  = Scale.get(this);
            int   indent = scale.scale(mModel.getIndentWidth(row, column));
            bounds.x += indent;
            bounds.width -= indent;
            int one = scale.scale(1);
            if (bounds.width < one) {
                bounds.width = one;
            }
        }
        return bounds;
    }

    /** Sets the width of all visible columns to their preferred width. */
    public void sizeColumnsToFit() {
        List<Column> columns = new ArrayList<>();
        for (Column column : mModel.getColumns()) {
            if (column.isVisible()) {
                int width = column.getPreferredWidth(this);
                if (width != column.getWidth()) {
                    column.setWidth(this, width);
                    columns.add(column);
                }
            }
        }
        if (!columns.isEmpty()) {
            processColumnWidthChanges(columns);
        }
    }

    /**
     * Repaint the specified row in this outline, as well as its proxies.
     *
     * @param row The row to repaint.
     */
    protected void repaintProxyRow(Row row) {
        repaint(getRowBounds(row));
        for (OutlineProxy proxy : mProxies) {
            proxy.repaintProxy(proxy.getRowBounds(row));
        }
    }

    /** Notifies all action listeners of a selection change. */
    protected void notifyOfSelectionChange() {
        notifyActionListeners(new ActionEvent(getRealOutline(), ActionEvent.ACTION_PERFORMED, getSelectionChangedActionCommand()));
    }

    /** @return The selection changed action command. */
    public String getSelectionChangedActionCommand() {
        return mSelectionChangedCommand;
    }

    /** @param command The selection changed action command. */
    public void setSelectionChangedActionCommand(String command) {
        mSelectionChangedCommand = command;
    }

    /** @return The potential user-initiated content size change action command. */
    public String getPotentialContentSizeChangeActionCommand() {
        return mPotentialContentSizeChangeCommand;
    }

    /** @param command The potential user-initiated content size change action command. */
    public void setPotentialContentSizeChangeActionCommand(String command) {
        mPotentialContentSizeChangeCommand = command;
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
        return mDeletableProxy != null && mDeletableProxy.canDeleteSelection();
    }

    @Override
    public void deleteSelection() {
        if (mDeletableProxy != null) {
            mDeletableProxy.deleteSelection();
        }
    }

    @Override
    public boolean canSelectAll() {
        return mModel.canSelectAll();
    }

    @Override
    public void selectAll() {
        mModel.select();
    }

    /**
     * Arranges the columns in the same order as the columns passed in.
     *
     * @param columns The column order.
     */
    public void setColumnOrder(List<Column> columns) {
        List<Column> list = new ArrayList<>(columns);
        List<Column> cols = mModel.getColumns();
        cols.removeAll(columns);
        list.addAll(cols);
        cols.clear();
        cols.addAll(list);
        repaint();
        getHeaderPanel().repaint();
    }

    /** @return The header panel for this table. */
    public OutlineHeader getHeaderPanel() {
        if (mHeaderPanel == null) {
            mHeaderPanel = new OutlineHeader(this);
        }
        return mHeaderPanel;
    }

    /** @param scrollTo The row index to scroll to. */
    protected void keyScroll(int scrollTo) {
        Outline real = getRealOutline();
        if (!real.keyScrollInternal(scrollTo)) {
            for (OutlineProxy proxy : real.mProxies) {
                if (proxy.keyScrollInternal(scrollTo)) {
                    break;
                }
            }
        }
    }

    protected boolean keyScrollInternal(int scrollTo) {
        if (scrollTo >= getFirstRowToDisplay() && scrollTo <= getLastRowToDisplay()) {
            requestFocus();
            scrollRectToVisible(getRowIndexBounds(scrollTo));
            return true;
        }
        return false;
    }

    @Override
    public void mouseClicked(MouseEvent event) {
        if (!mIgnoreClick) {
            requestFocus();
            Row rollRow = null;
            try {
                boolean local      = event.getSource() == this;
                int     x          = event.getX();
                int     y          = event.getY();
                Row     rowHit;
                Column  column;
                int     clickCount = event.getClickCount();

                rollRow = mRollRow;
                if (clickCount == 1) {
                    if (local) {
                        column = overColumn(x);
                        if (column != null) {
                            rowHit = overRow(y);
                            if (rowHit != null && !overDisclosureControl(x, y, column, rowHit)) {
                                Cell cell = column.getRowCell(rowHit);
                                if (cell != null) {
                                    cell.mouseClicked(event, getCellBounds(rowHit, column), rowHit, column);
                                }
                            }
                        }
                    }
                } else if (clickCount == 2) {
                    column = overColumnDivider(x);
                    if (column == null) {
                        if (local) {
                            rowHit = overRow(y);
                            if (rowHit != null) {
                                column = overColumn(x);
                                if ((column == null || !overDisclosureControl(x, y, column, rowHit)) && mModel.isRowSelected(rowHit)) {
                                    notifyActionListeners();
                                }
                            }
                        }
                    } else if (allowColumnResize()) {
                        int width = column.getPreferredWidth(this);
                        if (width != column.getWidth()) {
                            adjustColumnWidth(column, width);
                        }
                    }
                }
            } finally {
                repaintChangedRollRow(rollRow);
            }
        }
    }

    private void repaintChangedRollRow(Row rollRow) {
        if (rollRow != mRollRow) {
            Scale     scale  = Scale.get(this);
            Column    column = mModel.getHierarchyColumn();
            Rectangle bounds;
            if (mRollRow != null) {
                bounds = getCellBounds(mRollRow, column);
                bounds.width = scale.scale(mModel.getIndentWidth(mRollRow, column));
                repaint(bounds);
            }
            if (rollRow != null) {
                bounds = getCellBounds(rollRow, column);
                bounds.width = scale.scale(mModel.getIndentWidth(rollRow, column));
                repaint(bounds);
            }
            mRollRow = rollRow;
        }
    }

    @Override
    public void mouseEntered(MouseEvent event) {
        // Not used.
    }

    @Override
    public void mouseExited(MouseEvent event) {
        // Not used.
    }

    @Override
    public void mousePressed(MouseEvent event) {
        if (isEnabled()) {
            mIgnoreClick = false;
            Row rollRow = null;
            try {
                boolean local = event.getSource() == this;
                int     x     = event.getX();
                int     y     = event.getY();
                Row     rowHit;
                int     rowIndexHit;
                Column  column;
                requestFocus();
                mSelectOnMouseUp = -1;
                mDividerDrag = overColumnDivider(x);
                if (mDividerDrag != null) {
                    if (allowColumnResize()) {
                        mColumnStart = getColumnStart(mDividerDrag);
                    }
                } else if (local) {
                    column = overColumn(x);
                    rowHit = overRow(y);
                    if (column != null && rowHit != null) {
                        if (overDisclosureControl(x, y, column, rowHit)) {
                            Rectangle bounds = getCellBounds(rowHit, column);
                            bounds.width = Scale.get(this).scale(mModel.getIndentWidth(rowHit, column));
                            rollRow = mRollRow;
                            repaint(bounds);
                            rowHit.setOpen(!rowHit.isOpen());
                            return;
                        }
                    }
                    int method = Selection.MOUSE_NONE;

                    rowIndexHit = overRowIndex(y);

                    if (event.isShiftDown()) {
                        method |= Selection.MOUSE_EXTEND;
                    }
                    if ((event.getModifiersEx() & getToolkit().getMenuShortcutKeyMaskEx()) != 0 && !event.isPopupTrigger()) {
                        method |= Selection.MOUSE_FLIP;
                    }
                    mSelectOnMouseUp = mModel.getSelection().selectByMouse(rowIndexHit, method);
                    reapplyRowFilter();
                    if (event.isPopupTrigger()) {
                        mSelectOnMouseUp = -1;
                        mIgnoreClick = true;
                        showContextMenu(event);
                    }
                }
            } finally {
                repaintChangedRollRow(rollRow);
            }
        }
    }

    @Override
    public void mouseReleased(MouseEvent event) {
        if (isEnabled()) {
            Row rollRow = null;
            try {
                rollRow = mRollRow;
                if (mDividerDrag != null && allowColumnResize()) {
                    dragColumnDivider(event.getX());
                }
                mDividerDrag = null;
                if (mSelectOnMouseUp != -1) {
                    mModel.select(mSelectOnMouseUp, false);
                    mSelectOnMouseUp = -1;
                }
                if (event.isPopupTrigger()) {
                    mSelectOnMouseUp = -1;
                    mIgnoreClick = true;
                    showContextMenu(event);
                }
            } finally {
                repaintChangedRollRow(rollRow);
            }
        }
    }

    @Override
    public void mouseDragged(MouseEvent event) {
        if (isEnabled()) {
            try {
                int x = event.getX();
                mSelectOnMouseUp = -1;
                if (mDividerDrag != null && allowColumnResize()) {
                    dragColumnDivider(x);
                    if (UIUtilities.getAncestorOfType(this, JScrollPane.class) != null) {
                        Point pt = event.getPoint();
                        if (!(event.getSource() instanceof Outline)) {
                            // Column resizing is occurring in the header, most likely
                            pt.y = getVisibleRect().y + 1;
                        }
                        scrollRectToVisible(new Rectangle(pt.x, pt.y, 1, 1));
                    }
                }
            } finally {
                repaintChangedRollRow(null);
            }
        }
    }

    @Override
    public void mouseMoved(MouseEvent event) {
        if (isEnabled()) {
            Row rollRow = null;
            try {
                boolean local = event.getSource() == this;
                int     x     = event.getX();
                int     y     = event.getY();
                Row     rowHit;
                Column  column;

                Cursor cursor = Cursor.getDefaultCursor();

                if (overColumnDivider(x) != null) {
                    if (allowColumnResize()) {
                        cursor = Cursor.getPredefinedCursor(Cursor.W_RESIZE_CURSOR);
                    }
                } else if (local) {
                    column = overColumn(x);
                    if (column != null) {
                        rowHit = overRow(y);
                        if (rowHit != null) {
                            if (overDisclosureControl(x, y, column, rowHit)) {
                                rollRow = rowHit;
                            } else {
                                Cell cell = column.getRowCell(rowHit);
                                cursor = cell.getCursor(event, getCellBounds(rowHit, column), rowHit, column);
                            }
                        }
                    }
                }
                setCursor(cursor);
            } finally {
                repaintChangedRollRow(rollRow);
            }
        }
    }

    @Override
    public void keyPressed(KeyEvent event) {
        if (!event.isConsumed() && (event.getModifiersEx() & getToolkit().getMenuShortcutKeyMaskEx()) == 0) {
            Selection selection = mModel.getSelection();
            boolean   shiftDown = event.isShiftDown();
            int       index;
            switch (event.getKeyCode()) {
            case KeyEvent.VK_LEFT:
                index = selection.firstSelectedIndex();
                if (index >= 0) {
                    if (selection.getCount() == 1) {
                        Row row = mModel.getRowAtIndex(index);
                        if (row.canHaveChildren() && row.isOpen()) {
                            row.setOpen(false);
                            repaintSelection();
                        } else {
                            Row parentRow = row.getParent();
                            if (parentRow != null) {
                                index = getModel().getIndexOfRow(parentRow);
                                selection.select(index, false);
                                keyScroll(index);
                            }
                        }
                    } else {
                        while (index >= 0) {
                            mModel.getRowAtIndex(index).setOpen(false);
                            index = selection.nextSelectedIndex(index + 1);
                            repaintSelection();
                        }
                    }
                }
                break;
            case KeyEvent.VK_RIGHT:
                index = selection.firstSelectedIndex();
                while (index >= 0) {
                    mModel.getRowAtIndex(index).setOpen(true);
                    index = selection.nextSelectedIndex(index + 1);
                    repaintSelection();
                }
                break;
            case KeyEvent.VK_UP:
                index = selection.selectUp(shiftDown);
                if (index != -1) {
                    OutlineModel model  = getModel();
                    RowFilter    filter = model.getRowFilter();
                    if (filter != null) {
                        while (filter.isRowFiltered(model.getRowAtIndex(index))) {
                            int last = index;
                            index = selection.selectUp(shiftDown);
                            if (index == last || index == -1) {
                                break;
                            }
                        }
                        model.reapplyRowFilter();
                    }
                    keyScroll(index);
                }
                break;
            case KeyEvent.VK_DOWN:
                index = selection.selectDown(shiftDown);
                if (index != -1) {
                    OutlineModel model  = getModel();
                    RowFilter    filter = model.getRowFilter();
                    if (filter != null) {
                        while (filter.isRowFiltered(model.getRowAtIndex(index))) {
                            int last = index;
                            index = selection.selectDown(shiftDown);
                            if (index == last || index == -1) {
                                break;
                            }
                        }
                        model.reapplyRowFilter();
                    }
                    keyScroll(index);
                }
                break;
            case KeyEvent.VK_HOME:
                selectToHome(selection, shiftDown);
                break;
            case KeyEvent.VK_END:
                selectToEnd(selection, shiftDown);
                break;
            default:
                return;
            }
            event.consume();
        }
    }

    private void selectToHome(Selection selection, boolean shiftDown) {
        OutlineModel model = getModel();
        int          count = model.getRowCount();
        if (count > 0) {
            RowFilter filter = model.getRowFilter();
            if (filter != null) {
                int i = 0;
                while (i < count && filter.isRowFiltered(model.getRowAtIndex(i))) {
                    i++;
                }
                if (i == count) {
                    return;
                }
                if (shiftDown && !selection.isEmpty()) {
                    int anchor = selection.getAnchor();
                    if (anchor < 0) {
                        anchor = selection.lastSelectedIndex();
                    }
                    selection.select(i, anchor, true);
                } else {
                    selection.select(i, false);
                }
                model.reapplyRowFilter();
                keyScroll(i);
            } else {
                selection.select(0, shiftDown);
                keyScroll(0);
            }
        }
    }

    private void selectToEnd(Selection selection, boolean shiftDown) {
        OutlineModel model = getModel();
        int          count = model.getRowCount();
        if (count > 0) {
            RowFilter filter = model.getRowFilter();
            if (filter != null) {
                int i = count - 1;
                while (i >= 0 && filter.isRowFiltered(model.getRowAtIndex(i))) {
                    i--;
                }
                if (i < 0) {
                    return;
                }
                if (shiftDown && !selection.isEmpty()) {
                    int anchor = selection.getAnchor();
                    if (anchor < 0) {
                        anchor = selection.firstSelectedIndex();
                    }
                    selection.select(anchor, i, true);
                } else {
                    selection.select(i, false);
                }
                model.reapplyRowFilter();
                keyScroll(i);
            } else {
                selection.select(count - 1, shiftDown);
                keyScroll(count - 1);
            }
        }
    }

    @Override
    public void keyTyped(KeyEvent event) {
        if (!event.isConsumed() && (event.getModifiersEx() & getToolkit().getMenuShortcutKeyMaskEx()) == 0) {
            char ch = event.getKeyChar();
            if (ch == '\n' || ch == '\r') {
                if (mModel.hasSelection()) {
                    notifyActionListeners();
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

    @Override
    public void keyReleased(KeyEvent event) {
        // Not used.
    }

    /**
     * @param viewPt The location within the view.
     * @return The row cell at the specified point.
     */
    public Cell getCellAt(Point viewPt) {
        return getCellAt(viewPt.x, viewPt.y);
    }

    /**
     * @param x The x-coordinate within the view.
     * @param y The y-coordinate within the view.
     * @return The row cell at the specified coordinates.
     */
    public Cell getCellAt(int x, int y) {
        Column column = overColumn(x);
        if (column != null) {
            Row row = overRow(y);
            if (row != null) {
                return column.getRowCell(row);
            }
        }
        return null;
    }

    private void dragColumnDivider(int x) {
        Scale scale = Scale.get(this);
        int   slop  = scale.scale(DIVIDER_HIT_SLOP);
        int   old   = mDividerDrag.getWidth();
        if (x <= mColumnStart + slop * 2) {
            x = mColumnStart + slop * 2 + scale.scale(1);
        }
        x -= mColumnStart;
        if (old != x) {
            adjustColumnWidth(mDividerDrag, x);
        }
    }

    /**
     * @param column The column to adjust.
     * @param width  The new column width.
     */
    public void adjustColumnWidth(Column column, int width) {
        List<Column> columns = new ArrayList<>(1);
        column.setWidth(this, width);
        columns.add(column);
        processColumnWidthChanges(columns);
    }

    private void processColumnWidthChanges(List<Column> columns) {
        updateRowHeightsIfNeeded(columns);
        revalidateView();
    }

    /**
     * @param x The x-coordinate.
     * @param y The y-coordinate.
     * @return The drag image for this table when dragging rows.
     */
    protected Img getDragImage(int x, int y) {
        Graphics2D gc = null;
        mDrawingDragImage = true;
        mDragClip = null;
        Img off1 = getImage();
        mDrawingDragImage = false;
        if (mDragClip == null) {
            mDragClip = new Rectangle(x, y, 1, 1);
        }
        Img off2;
        try {
            off2 = Img.create(getGraphicsConfiguration(), mDragClip.width, mDragClip.height, Transparency.TRANSLUCENT);
            gc = off2.getGraphics();
            gc.setClip(new Rectangle(0, 0, mDragClip.width, mDragClip.height));
            gc.setBackground(new Color(0, true));
            gc.clearRect(0, 0, mDragClip.width, mDragClip.height);
            gc.setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, 0.5f));
            Rectangle bounds = getVisibleRect();
            gc.translate(-(mDragClip.x - bounds.x), -(mDragClip.y - bounds.y));
            gc.drawImage(off1, 0, 0, this);
        } catch (Exception paintException) {
            Log.error(paintException);
            off2 = null;
            mDragClip = new Rectangle(x, y, 1, 1);
        } finally {
            if (gc != null) {
                gc.dispose();
            }
        }
        return off2 != null ? off2 : off1;
    }

    /**
     * @return An {@link Img} containing the current contents of this component, minus the specified
     *         component and its children.
     */
    public Img getImage() {
        Img offscreen = null;
        synchronized (getTreeLock()) {
            Graphics2D gc = null;
            try {
                Rectangle bounds = getVisibleRect();
                offscreen = Img.create(getGraphicsConfiguration(), bounds.width, bounds.height, Transparency.TRANSLUCENT);
                gc = offscreen.getGraphics();
                Color saved = gc.getBackground();
                gc.setBackground(new Color(0, true));
                gc.clearRect(0, 0, bounds.width, bounds.height);
                gc.setBackground(saved);
                Rectangle clip = new Rectangle(0, 0, bounds.width, bounds.height);
                gc.setClip(clip);
                gc.translate(-bounds.x, -bounds.y);
                paint(gc);
            } catch (Exception exception) {
                Log.error(exception);
            } finally {
                if (gc != null) {
                    gc.dispose();
                }
            }
        }
        return offscreen;
    }

    @Override
    public boolean isOpaque() {
        return super.isOpaque() && !mDrawingDragImage;
    }

    /**
     * Displays a context menu.
     *
     * @param event The triggering mouse event.
     */
    protected void showContextMenu(MouseEvent event) {
        // Does nothing by default.
    }

    /** @return {@code true} if column resizing is allowed. */
    public boolean allowColumnResize() {
        return mAllowColumnResize;
    }

    /** @param allow Whether column resizing is on or off. */
    public void setAllowColumnResize(boolean allow) {
        mAllowColumnResize = allow;
    }

    /** @return {@code true} if row dragging is allowed. */
    public boolean allowRowDrag() {
        return mAllowRowDrag;
    }

    /** @param allow Whether row dragging is on or off. */
    public void setAllowRowDrag(boolean allow) {
        mAllowRowDrag = allow;
    }

    /** Revalidates the view and header panel if it exists. */
    public void revalidateView() {
        revalidate();
        if (mHeaderPanel != null) {
            mHeaderPanel.revalidate();
            mHeaderPanel.repaint();
        }
        repaint();
    }

    @Override
    public void setBounds(int x, int y, int width, int height) {
        super.setBounds(x, y, width, height);
        if (mHeaderPanel != null) {
            Dimension size = mHeaderPanel.getPreferredSize();
            mHeaderPanel.setResizeOK(true);
            mHeaderPanel.setBounds(getX(), mHeaderPanel.getY(), getWidth(), size.height);
            mHeaderPanel.setResizeOK(false);
        }
    }

    /**
     * Sort on a column.
     *
     * @param column    The column to sort on.
     * @param ascending Pass in {@code true} for an ascending sort.
     * @param add       Pass in {@code true} to add this column to the end of the sort order, or
     *                  {@code false} to make this column the primary and only sort column.
     */
    public void setSort(Column column, boolean ascending, boolean add) {
        StateEdit edit  = new StateEdit(mModel, I18n.Text("Sort"));
        int       count = mModel.getColumnCount();
        int       i;

        if (add) {
            if (column.getSortSequence() == -1) {
                int highest = -1;
                for (i = 0; i < count; i++) {
                    int sortOrder = mModel.getColumnAtIndex(i).getSortSequence();
                    if (sortOrder > highest) {
                        highest = sortOrder;
                    }
                }
                column.setSortCriteria(highest + 1, ascending);
            } else {
                column.setSortCriteria(column.getSortSequence(), ascending);
            }
        } else {
            for (i = 0; i < count; i++) {
                Column col = mModel.getColumnAtIndex(i);
                if (column == col) {
                    col.setSortCriteria(0, ascending);
                } else {
                    col.setSortCriteria(-1, col.isSortAscending());
                }
            }
        }
        mModel.sort();
        edit.end();
        postUndo(edit);
    }

    @Override
    public void focusGained(FocusEvent event) {
        repaintSelection();
        repaintFocus();
    }

    @Override
    public void focusLost(FocusEvent event) {
        repaintSelection();
        repaintFocus();
    }

    private void repaintFocus() {
        Scale     scale  = Scale.get(this);
        int       one    = scale.scale(1);
        Rectangle bounds = getVisibleRect();
        paintImmediately(bounds.x, bounds.y, bounds.width, one);
        paintImmediately(bounds.x, bounds.y + bounds.height - one, bounds.width, one);
        paintImmediately(bounds.x, bounds.y + one, one, bounds.height - (one + one));
        paintImmediately(bounds.x + bounds.width - one, bounds.y + one, one, bounds.height - (one + one));
    }

    /**
     * @param outline The outline to check.
     * @return Whether the specified outline refers to this outline or a proxy of it.
     */
    public boolean isSelfOrProxy(Outline outline) {
        Outline self = getRealOutline();
        if (outline == self) {
            return true;
        }
        for (OutlineProxy proxy : self.mProxies) {
            if (outline == proxy) {
                return true;
            }
        }
        return false;
    }

    @Override
    public Point getToolTipLocation(MouseEvent event) {
        int    x      = event.getX();
        Column column = overColumn(x);
        if (column != null) {
            Row row = overRow(event.getY());
            if (row != null) {
                Rectangle bounds = getCellBounds(row, column);
                String    text   = getToolTipText(event);
                if (mLastTooltipText == null || !mLastTooltipText.equals(text)) {
                    mLastTooltipText = text;
                    mLastTooltipX = x;
                }
                return new Point(mLastTooltipX, bounds.y + bounds.height);
            }
        }
        return null;
    }

    @Override
    public String getToolTipText(MouseEvent event) {
        Column column = overColumn(event.getX());
        if (column != null) {
            Row row = overRow(event.getY());
            if (row != null) {
                return Text.wrapPlainTextForToolTip(column.getRowCell(row).getToolTipText(this, event, getCellBounds(row, column), row, column));
            }
        }
        return super.getToolTipText(event);
    }

    /** @return {@code true} if background banding is enabled. */
    public boolean useBanding() {
        return mUseBanding;
    }

    /** @param useBanding Whether to use background banding or not. */
    public void setUseBanding(boolean useBanding) {
        mUseBanding = useBanding;
    }

    @Override
    public void dragGestureRecognized(DragGestureEvent dge) {
        if (mDividerDrag == null && mModel.hasSelection() && allowRowDrag() && hasFocus()) {
            Point        pt        = dge.getDragOrigin();
            RowSelection selection = new RowSelection(mModel, mModel.getSelectionAsList(true).toArray(new Row[0]));
            if (DragSource.isDragImageSupported()) {
                Img   dragImage   = getDragImage(pt.x, pt.y);
                Point imageOffset = new Point(mDragClip.x - pt.x, mDragClip.y - pt.y);
                dge.startDrag(null, dragImage, imageOffset, selection, null);
            } else {
                dge.startDrag(null, selection);
            }
        }
    }

    /**
     * @param dtde The drop target drag event.
     * @return {@code true} if the contents of the drag can be dropped into this outline.
     */
    protected boolean isDragAcceptable(DropTargetDragEvent dtde) {
        boolean result = false;
        mAlternateDragDestination = null;
        try {
            if (dtde.isDataFlavorSupported(Column.DATA_FLAVOR)) {
                Column column = (Column) dtde.getTransferable().getTransferData(Column.DATA_FLAVOR);
                result = isColumnDragAcceptable(dtde, column);
                if (result) {
                    mModel.setDragColumn(column);
                }
            } else if (dtde.isDataFlavorSupported(RowSelection.DATA_FLAVOR)) {
                Row[] rows = (Row[]) dtde.getTransferable().getTransferData(RowSelection.DATA_FLAVOR);
                result = isRowDragAcceptable(dtde, rows);
                if (result) {
                    mModel.setDragRows(rows);
                }
            } else if (dtde.isDataFlavorSupported(DockableTransferable.DATA_FLAVOR)) {
                mAlternateDragDestination = UIUtilities.getAncestorOfType(this, Dock.class);
            }
        } catch (Exception exception) {
            Log.error(exception);
        }
        return result;
    }

    /** @return Whether or not the user can sort by clicking in the column header. */
    public boolean isUserSortable() {
        return mUserSortable;
    }

    /** @param sortable Whether or not the user can sort by clicking in the column header. */
    public void setUserSortable(boolean sortable) {
        mUserSortable = sortable;
    }

    /**
     * @param dtde   The drop target drag event.
     * @param column The column.
     * @return {@code true} if the contents of the drag can be dropped into this outline.
     */
    protected boolean isColumnDragAcceptable(DropTargetDragEvent dtde, Column column) {
        return mModel.getColumns().contains(column);
    }

    /**
     * @param dtde The drop target drag event.
     * @param rows The rows.
     * @return {@code true} if the contents of the drag can be dropped into this outline.
     */
    protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, Row[] rows) {
        return rows.length > 0 && mModel.getRows().contains(rows[0]);
    }

    @Override
    public void dragEnter(DropTargetDragEvent dtde) {
        mDragWasAcceptable = isDragAcceptable(dtde);
        if (mDragWasAcceptable) {
            if (mModel.getDragColumn() != null) {
                dtde.acceptDrag(dragEnterColumn(dtde));
                return;
            }
            Row[] rows = mModel.getDragRows();
            if (rows != null && rows.length > 0) {
                dtde.acceptDrag(dragEnterRow(dtde));
                return;
            }
        } else if (mAlternateDragDestination != null) {
            UIUtilities.updateDropTargetDragPointTo(dtde, mAlternateDragDestination);
            mAlternateDragDestination.dragEnter(dtde);
            return;
        }
        dtde.rejectDrag();
    }

    /**
     * Called when a column drag is entered.
     *
     * @param dtde The drag event.
     * @return The value to return via {@link DropTargetDragEvent#acceptDrag(int)}.
     */
    protected int dragEnterColumn(DropTargetDragEvent dtde) {
        mSavedColumns = new ArrayList<>(mModel.getColumns());
        return DnDConstants.ACTION_MOVE;
    }

    /**
     * Called when a row drag is entered.
     *
     * @param dtde The drag event.
     * @return The value to return via {@link DropTargetDragEvent#acceptDrag(int)}.
     */
    protected int dragEnterRow(DropTargetDragEvent dtde) {
        addDragHighlight(this);
        return DnDConstants.ACTION_MOVE;
    }

    /**
     * Called when a row drag is entered over a proxy to this outline.
     *
     * @param dtde  The drag event.
     * @param proxy The proxy.
     */
    @SuppressWarnings("static-method")
    protected void dragEnterRow(DropTargetDragEvent dtde, OutlineProxy proxy) {
        addDragHighlight(proxy);
    }

    @Override
    public void dragOver(DropTargetDragEvent dtde) {
        if (mDragWasAcceptable) {
            if (mModel.getDragColumn() != null) {
                dtde.acceptDrag(dragOverColumn(dtde));
                return;
            }
            Row[] rows = mModel.getDragRows();
            if (rows != null && rows.length > 0) {
                dtde.acceptDrag(dragOverRow(dtde));
                return;
            }
        } else if (mAlternateDragDestination != null) {
            UIUtilities.updateDropTargetDragPointTo(dtde, mAlternateDragDestination);
            mAlternateDragDestination.dragOver(dtde);
            return;
        }
        dtde.rejectDrag();
    }

    /**
     * Called when a column drag is in progress.
     *
     * @param dtde The drag event.
     * @return The value to return via {@link DropTargetDragEvent#acceptDrag(int)}.
     */
    protected int dragOverColumn(DropTargetDragEvent dtde) {
        int x    = UIUtilities.convertDropTargetDragPointTo(dtde, this).x;
        int over = overColumnIndex(x);
        int cur  = mModel.getIndexOfColumn(mModel.getDragColumn());
        if (over != cur && over != -1) {
            int midway = getColumnIndexStart(over) + mModel.getColumnAtIndex(over).getWidth() / 2;
            if (over < cur && x < midway || over > cur && x > midway) {
                List<Column> columns = mModel.getColumns();
                if (cur < over) {
                    for (int i = cur; i < over; i++) {
                        columns.set(i, mModel.getColumnAtIndex(i + 1));
                    }
                } else {
                    for (int j = cur; j > over; j--) {
                        columns.set(j, mModel.getColumnAtIndex(j - 1));
                    }
                }
                columns.set(over, mModel.getDragColumn());
                repaint();
                if (mHeaderPanel != null) {
                    mHeaderPanel.repaint();
                }
            }
        }
        return DnDConstants.ACTION_MOVE;
    }

    protected boolean isDropOnRow(Row[] dragRows) {
        return false;
    }

    /**
     * Called when a row drag is in progress.
     *
     * @param dtde The drag event.
     * @return The value to return via {@link DropTargetDragEvent#acceptDrag(int)}.
     */
    protected int dragOverRow(DropTargetDragEvent dtde) {
        Row[] dragRows = mModel.getDragRows();
        if (isDropOnRow(dragRows)) {
            Point pt = UIUtilities.convertDropTargetDragPointTo(dtde, this);
            setDragTargetRow(overRow(pt.y));
            return getDragTargetRow() != null ? DnDConstants.ACTION_MOVE : DnDConstants.ACTION_NONE;
        }
        Scale     scale                 = Scale.get(this);
        int       one                   = scale.scale(1);
        Row       savedParentRow        = mDragParentRow;
        int       savedChildInsertIndex = mDragChildInsertIndex;
        Row       parentRow             = null;
        int       childInsertIndex      = -1;
        Point     pt                    = UIUtilities.convertDropTargetDragPointTo(dtde, this);
        int       y                     = getInsets().top;
        int       last                  = getLastRowToDisplay();
        boolean   isFromSelf            = dragRows != null && dragRows.length > 0 && mModel.getRows().contains(dragRows[0]);
        int       indentWidth           = scale.scale(mModel.getIndentWidth());
        Rectangle bounds;
        int       indent;
        Row       row;

        for (int i = getFirstRowToDisplay(); i <= last; i++) {
            row = mModel.getRowAtIndex(i);
            if (!mModel.isRowFiltered(row)) {
                int height = row.getHeight();
                if (pt.y <= y + height / 2) {
                    if (!isFromSelf || !mModel.isExtendedRowSelected(i) || i != 0 && !mModel.isExtendedRowSelected(i - 1)) {
                        parentRow = row.getParent();
                        childInsertIndex = parentRow != null ? parentRow.getIndexOfChild(row) : i;
                        break;
                    }
                } else if (pt.y <= y + height) {
                    if (row.canHaveChildren()) {
                        bounds = getRowBounds(row);
                        indent = indentWidth + scale.scale(mModel.getIndentWidth(row, mModel.getHierarchyColumn()));
                        if (pt.x >= bounds.x + indent && (!isFromSelf || !mModel.isExtendedRowSelected(row))) {
                            parentRow = row;
                            childInsertIndex = 0;
                            break;
                        }
                    }
                    if (!isFromSelf || !mModel.isExtendedRowSelected(i) || i < last && !mModel.isExtendedRowSelected(i + 1)) {
                        parentRow = row.getParent();
                        if (parentRow != null) {
                            if (!isFromSelf || !mModel.isExtendedRowSelected(i)) {
                                childInsertIndex = parentRow.getIndexOfChild(row) + 1;
                                break;
                            }
                        } else {
                            childInsertIndex = i + 1;
                            break;
                        }
                    }
                }
                y += height + (mDrawRowDividers ? one : 0);
            }
        }
        if (childInsertIndex == -1) {
            if (last > 0) {
                row = mModel.getRowAtIndex(last);
                if (row.canHaveChildren()) {
                    bounds = getRowBounds(row);
                    indent = indentWidth + scale.scale(mModel.getIndentWidth(row, mModel.getHierarchyColumn()));
                    if (pt.x >= bounds.x + indent && (!isFromSelf || !mModel.isExtendedRowSelected(row))) {
                        parentRow = row;
                        childInsertIndex = 0;
                    }
                }
                if (childInsertIndex == -1) {
                    parentRow = row.getParent();
                    if (parentRow != null && (!isFromSelf || !mModel.isExtendedRowSelected(parentRow))) {
                        childInsertIndex = parentRow.getIndexOfChild(row) + 1;
                    } else {
                        parentRow = null;
                        childInsertIndex = last + 1;
                    }
                }
            } else {
                parentRow = null;
                childInsertIndex = 0;
            }
        }

        if (!isDragToRowAcceptable(parentRow)) {
            mDragParentRow = null;
            mDragChildInsertIndex = 0;
            if (savedParentRow != null || mDragChildInsertIndex != savedChildInsertIndex) {
                repaint(getDragRowInsertionMarkerBounds(savedParentRow, savedChildInsertIndex));
            }
            return DnDConstants.ACTION_NONE;
        }

        if (mDragParentRow != parentRow || mDragChildInsertIndex != childInsertIndex) {
            Graphics gc = getGraphics();
            mDragParentRow = parentRow;
            mDragChildInsertIndex = childInsertIndex;
            drawDragRowInsertionMarker(gc, mDragParentRow, mDragChildInsertIndex);
            gc.dispose();
        }

        if (mDragParentRow != savedParentRow || mDragChildInsertIndex != savedChildInsertIndex) {
            repaint(getDragRowInsertionMarkerBounds(savedParentRow, savedChildInsertIndex));
        }
        return DnDConstants.ACTION_MOVE;
    }

    @SuppressWarnings("static-method")
    protected boolean isDragToRowAcceptable(Row parentRow) {
        return true;
    }

    @Override
    public void dropActionChanged(DropTargetDragEvent dtde) {
        if (mDragWasAcceptable) {
            Row[] rows;

            if (mModel.getDragColumn() != null) {
                dtde.acceptDrag(dropActionChangedColumn(dtde));
                return;
            }
            rows = mModel.getDragRows();
            if (rows != null && rows.length > 0) {
                dtde.acceptDrag(dropActionChangedRow(dtde));
                return;
            }
        } else if (mAlternateDragDestination != null) {
            UIUtilities.updateDropTargetDragPointTo(dtde, mAlternateDragDestination);
            mAlternateDragDestination.dropActionChanged(dtde);
            return;
        }
        dtde.rejectDrag();
    }

    /**
     * Called when a column drop action is changed.
     *
     * @param dtde The drag event.
     * @return The value to return via {@link DropTargetDragEvent#acceptDrag(int)}.
     */
    @SuppressWarnings("static-method")
    protected int dropActionChangedColumn(DropTargetDragEvent dtde) {
        return DnDConstants.ACTION_MOVE;
    }

    /**
     * Called when a row drop action is changed.
     *
     * @param dtde The drag event.
     * @return The value to return via {@link DropTargetDragEvent#acceptDrag(int)}.
     */
    @SuppressWarnings("static-method")
    protected int dropActionChangedRow(DropTargetDragEvent dtde) {
        return DnDConstants.ACTION_MOVE;
    }

    @Override
    public void dragExit(DropTargetEvent dte) {
        if (mDragWasAcceptable) {
            if (mModel.getDragColumn() != null) {
                dragExitColumn(dte);
            } else {
                Row[] rows = mModel.getDragRows();

                if (rows != null && rows.length > 0) {
                    dragExitRow(dte);
                }
            }
        } else if (mAlternateDragDestination != null) {
            mAlternateDragDestination.dragExit(dte);
        }
    }

    /**
     * Called when a column drag leaves the outline.
     *
     * @param dte The drop target event.
     */
    protected void dragExitColumn(DropTargetEvent dte) {
        List<Column> columns = mModel.getColumns();

        if (columns.equals(mSavedColumns)) {
            repaintColumn(mModel.getDragColumn());
        } else {
            columns.clear();
            columns.addAll(mSavedColumns);
            repaint();
            if (mHeaderPanel != null) {
                mHeaderPanel.repaint();
            }
        }
        mSavedColumns = null;
        mModel.setDragColumn(null);
    }

    /**
     * Called when a row drag leaves the outline.
     *
     * @param dte The drop target event.
     */
    protected void dragExitRow(DropTargetEvent dte) {
        setDragTargetRow(null);
        repaint(getDragRowInsertionMarkerBounds(mDragParentRow, mDragChildInsertIndex));
        removeDragHighlight(this);
        mDragParentRow = null;
        mDragChildInsertIndex = -1;
        mModel.setDragRows(null);
    }

    /**
     * Called when a row drag leaves a proxy of this outline.
     *
     * @param dte   The drop target event.
     * @param proxy The proxy.
     */
    @SuppressWarnings("static-method")
    protected void dragExitRow(DropTargetEvent dte, OutlineProxy proxy) {
        removeDragHighlight(proxy);
    }

    @Override
    public void drop(DropTargetDropEvent dtde) {
        if (mDragWasAcceptable) {
            dtde.acceptDrop(dtde.getDropAction());
            if (mModel.getDragColumn() != null) {
                dropColumn(dtde);
            } else {
                Row[] rows = mModel.getDragRows();
                if (rows != null && rows.length > 0) {
                    dropRow(dtde);
                }
            }
            dtde.dropComplete(true);
        } else if (mAlternateDragDestination != null) {
            UIUtilities.updateDropTargetDropPointTo(dtde, mAlternateDragDestination);
            mAlternateDragDestination.drop(dtde);
        }
    }

    /**
     * Called when a column drag leaves the outline.
     *
     * @param dtde The drop target drop event.
     */
    protected void dropColumn(DropTargetDropEvent dtde) {
        repaintColumn(mModel.getDragColumn());
        mSavedColumns = null;
        mModel.setDragColumn(null);
    }

    /**
     * Called to convert the foreign drag rows to this outline's row type and remove them from the
     * other outline. The default implementation merely removes them from the other outline and
     * returns the original drag rows, plus any of their children that should be added due to the
     * row being open. All rows put in the list will have had their owner set to this outline.
     *
     * @param list A list to hold the converted rows.
     */
    public void convertDragRowsToSelf(List<Row> list) {
        Row[] rows = mModel.getDragRows();
        rows[0].getOwner().removeRows(rows);
        for (Row element : rows) {
            mModel.collectRowsAndSetOwner(list, element, false);
        }
    }

    /**
     * Called when a row drop occurs.
     *
     * @param dtde The drop target drop event.
     */
    protected void dropRow(DropTargetDropEvent dtde) {
        requestFocus();
        if (dropOnRow(dtde)) {
            return;
        }
        removeDragHighlight(this);
        if (mDragChildInsertIndex != -1) {
            StateEdit edit         = new StateEdit(mModel, I18n.Text("Row Drag & Drop"));
            Row[]     dragRows     = mModel.getDragRows();
            boolean   isFromSelf   = dragRows != null && dragRows.length > 0 && mModel.getRows().contains(dragRows[0]);
            int       count        = mModel.getRowCount();
            List<Row> rows         = new ArrayList<>(count);
            List<Row> selection    = new ArrayList<>(count);
            List<Row> needSelected = new ArrayList<>(count);
            List<Row> modelRows;
            int       i;
            int       insertAt;
            Row       row;

            // Collect up the selected rows
            if (isFromSelf) {
                for (i = 0; i < count; i++) {
                    row = mModel.getRowAtIndex(i);
                    if (mModel.isExtendedRowSelected(row)) {
                        selection.add(row);
                    }
                }
            } else {
                convertDragRowsToSelf(selection);
            }

            // Re-order the visible rows
            insertAt = mDragParentRow != null && !mDragParentRow.isOpen() ? -1 : getAbsoluteInsertionIndex(mDragParentRow, mDragChildInsertIndex);
            for (i = 0; i < count; i++) {
                row = mModel.getRowAtIndex(i);
                if (i == insertAt) {
                    rows.addAll(selection);
                }
                if (!isFromSelf || !mModel.isExtendedRowSelected(row)) {
                    rows.add(row);
                }
            }
            if (count == insertAt) {
                rows.addAll(selection);
            }

            // Prune the selected rows that don't need to have their parents updated
            needSelected.addAll(selection);
            count = selection.size() - 1;
            for (i = count; i >= 0; i--) {
                Row parent;

                row = selection.get(i);
                parent = row.getParent();
                if (insertAt == -1) {
                    row.setOwner(null);
                }
                if (parent != null && (isFromSelf ? !mModel.isExtendedRowSelected(parent) : !selection.contains(parent))) {
                    row.removeFromParent();
                } else if (parent != null) {
                    selection.remove(i);
                }
            }

            // Update the parents of the remaining selected rows
            if (mDragParentRow != null) {
                count = selection.size() - 1;
                for (i = count; i >= 0; i--) {
                    mDragParentRow.insertChild(mDragChildInsertIndex, selection.get(i));
                }
            }

            mModel.deselect();
            mModel.clearSort();
            mDragParentRow = null;
            mDragChildInsertIndex = -1;
            modelRows = mModel.getRows();
            modelRows.clear();
            modelRows.addAll(rows);
            mModel.getSelection().setSize(modelRows.size());
            setSize(getPreferredSize());
            mModel.select(needSelected, false);
            edit.end();
            postUndo(edit);
            repaint();
            contentSizeMayHaveChanged();
            rowsWereDropped();
        }
        mModel.setDragRows(null);
    }

    protected boolean dropOnRow(DropTargetDropEvent dtde) {
        return false;
    }

    /**
     * Called when a row drop occurs in a proxy of this outline.
     *
     * @param dtde  The drop target drop event.
     * @param proxy The proxy.
     */
    @SuppressWarnings("static-method")
    protected void dropRow(DropTargetDropEvent dtde, OutlineProxy proxy) {
        removeDragHighlight(proxy);
    }

    /** Called after a row drop. */
    protected void rowsWereDropped() {
        // Does nothing.
    }

    protected static void addDragHighlight(Outline outline) {
        outline.mDragFocus = true;
        outline.repaintFocus();
    }

    protected static void removeDragHighlight(Outline outline) {
        outline.mDragFocus = false;
        outline.repaintFocus();
    }

    private int getAbsoluteInsertionIndex(Row parent, int childInsertIndex) {
        int insertAt;
        int count;
        if (parent == null) {
            count = mModel.getRowCount();
            insertAt = childInsertIndex;
            while (insertAt < count && mModel.getRowAtIndex(insertAt).getParent() != null) {
                insertAt++;
            }
        } else {
            int i = parent.getChildCount();
            if (i == 0 || !parent.isOpen()) {
                insertAt = mModel.getIndexOfRow(parent) + 1;
            } else if (childInsertIndex < i) {
                insertAt = mModel.getIndexOfRow(parent.getChild(childInsertIndex));
            } else {
                Row row = parent.getChild(i - 1);
                count = mModel.getRowCount();
                insertAt = mModel.getIndexOfRow(row) + 1;
                while (insertAt < count && mModel.getRowAtIndex(insertAt).isDescendantOf(row)) {
                    insertAt++;
                }
            }
        }
        return insertAt;
    }

    /**
     * Causes all row heights to be recalculated, if necessary.
     *
     * @param columns The columns that had their width altered.
     */
    public void updateRowHeightsIfNeeded(Collection<Column> columns) {
        if (dynamicRowHeight()) {
            for (Column column : columns) {
                if (column.getRowCell(null).participatesInDynamicRowLayout()) {
                    updateRowHeights();
                    break;
                }
            }
        }
    }

    /** @param row Causes the row height to be recalculated. */
    public void updateRowHeight(Row row) {
        List<Row> rows = new ArrayList<>(1);
        rows.add(row);
        updateRowHeights(rows);
    }

    /** Causes all row heights to be recalculated. */
    public void updateRowHeights() {
        updateRowHeights(mModel.getRows());
    }

    /**
     * Causes row heights to be recalculated.
     *
     * @param rows The rows to update.
     */
    public void updateRowHeights(Collection<? extends Row> rows) {
        List<Column> columns        = mModel.getColumns();
        boolean      needRevalidate = false;
        for (Row row : rows) {
            int height     = row.getHeight();
            int prefHeight = row.getPreferredHeight(this, columns);
            if (height != prefHeight) {
                row.setHeight(prefHeight);
                needRevalidate = true;
            }
        }
        if (needRevalidate) {
            contentSizeMayHaveChanged();
            revalidateView();
        }
    }

    @Override
    public Insets getAutoscrollInsets() {
        int         margin     = Scale.get(this).scale(AUTO_SCROLL_MARGIN);
        JScrollPane scrollPane = UIUtilities.getAncestorOfType(this, JScrollPane.class);
        if (scrollPane != null) {
            Rectangle bounds = scrollPane.getViewport().getViewRect();
            return new Insets(bounds.y + margin, bounds.x + margin, getHeight() - (bounds.y + bounds.height) + margin, getWidth() - (bounds.x + bounds.width) + margin);
        }
        return new Insets(margin, margin, margin, margin);
    }

    @Override
    public void autoscroll(Point pt) {
        int       margin = Scale.get(this).scale(AUTO_SCROLL_MARGIN);
        Insets    insets = getAutoscrollInsets();
        Dimension size   = getSize();
        if (pt.x < insets.left) {
            pt.x -= margin;
        } else if (pt.x > size.width - insets.right) {
            pt.x += margin;
        }
        if (pt.y < insets.top) {
            pt.y -= margin;
        } else if (pt.y > size.height - insets.bottom) {
            pt.y += margin;
        }
        scrollRectToVisible(new Rectangle(pt.x, pt.y, 1, 1));
    }

    /**
     * Called whenever the contents of this outline changed due to a user action such that its
     * preferred size might be different now.
     */
    public void contentSizeMayHaveChanged() {
        notifyActionListeners(new ActionEvent(getRealOutline(), ActionEvent.ACTION_PERFORMED, getPotentialContentSizeChangeActionCommand()));
    }

    @Override
    public void rowsAdded(OutlineModel model, Row[] rows) {
        contentSizeMayHaveChanged();
        revalidateView();
    }

    @Override
    public void rowsWillBeRemoved(OutlineModel model, Row[] rows) {
        // Nothing to do.
    }

    @Override
    public void rowsWereRemoved(OutlineModel model, Row[] rows) {
        for (Row element : rows) {
            if (element == mRollRow) {
                mRollRow = null;
                break;
            }
        }
        contentSizeMayHaveChanged();
        revalidateView();
    }

    @Override
    public void rowWasModified(OutlineModel model, Row row, Column column) {
        repaint();
    }

    @Override
    public void sortCleared(OutlineModel model) {
        repaintHeader();
    }

    @Override
    public void sorted(OutlineModel model) {
        if (isFocusOwner()) {
            scrollSelectionIntoView();
        }
        repaintView();
    }

    /** Scrolls the selection into view, if possible. */
    public void scrollSelectionIntoView() {
        int first = mModel.getFirstSelectedRowIndex();
        if (first >= getFirstRowToDisplay() && first <= getLastRowToDisplay()) {
            scrollSelectionIntoViewInternal();
        } else if (mProxies != null) {
            for (Outline proxy : mProxies) {
                if (first >= proxy.getFirstRowToDisplay() && first <= proxy.getLastRowToDisplay()) {
                    proxy.scrollSelectionIntoViewInternal();
                    break;
                }
            }
        }
    }

    private void scrollSelectionIntoViewInternal() {
        Selection selection = mModel.getSelection();
        int       first     = selection.nextSelectedIndex(getFirstRowToDisplay());
        int       max       = getLastRowToDisplay();

        if (first != -1 && first <= max) {
            Rectangle bounds = getRowIndexBounds(first);
            int       tmp    = first;
            int       last;

            do {
                last = tmp;
                tmp = selection.nextSelectedIndex(last + 1);
            } while (tmp != -1 && tmp <= max);

            if (first != last) {
                bounds = Geometry.union(bounds, getRowIndexBounds(last));
            }
            scrollRectToVisible(bounds);
        }
    }

    @Override
    public void lockedStateWillChange(OutlineModel model) {
        // Nothing to do...
    }

    @Override
    public void lockedStateDidChange(OutlineModel model) {
        // Nothing to do...
    }

    @Override
    public void selectionWillChange(OutlineModel model) {
        repaintSelectionInternal();
    }

    @Override
    public void selectionDidChange(OutlineModel model) {
        repaintSelectionInternal();
        if (!(this instanceof OutlineProxy)) {
            notifyOfSelectionChange();
        }
    }

    @Override
    public void undoWillHappen(OutlineModel model) {
        // Nothing to do.
    }

    @Override
    public void undoDidHappen(OutlineModel model) {
        contentSizeMayHaveChanged();
        revalidateView();
    }

    /** @param undo The undo to post. */
    public void postUndo(UndoableEdit undo) {
        Undoable undoable = UIUtilities.getSelfOrAncestorOfType(this, Undoable.class);
        if (undoable != null) {
            undoable.getUndoManager().addEdit(undo);
        }
    }

    /**
     * @param index The row index to look for.
     * @return The {@link Outline} most suitable for displaying the index.
     */
    public Outline getBestOutlineForRowIndex(int index) {
        Outline outline = getRealOutline();

        for (Outline other : outline.mProxies) {
            if (other.mFirstRow <= index && other.mLastRow >= index) {
                return other;
            }
        }
        return outline;
    }

    @Override
    public void componentHidden(ComponentEvent event) {
        // Not used.
    }

    @Override
    public void componentMoved(ComponentEvent event) {
        if (isFocusOwner()) {
            repaint();
        }
    }

    @Override
    public void componentResized(ComponentEvent event) {
        // Not used.
    }

    @Override
    public void componentShown(ComponentEvent event) {
        // Not used.
    }

    public Row getDragTargetRow() {
        return mModel.getDragTargetRow();
    }

    public void setDragTargetRow(Row row) {
        Row dragTargetRow = mModel.getDragTargetRow();
        if (dragTargetRow != null) {
            repaint(getRowBounds(dragTargetRow));
        }
        mModel.setDragTargetRow(row);
        if (row != null) {
            repaint(getRowBounds(row));
        }
    }
}
