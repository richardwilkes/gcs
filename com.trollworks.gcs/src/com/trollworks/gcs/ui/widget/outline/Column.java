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

import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.datatransfer.DataFlavor;
import java.awt.datatransfer.Transferable;
import java.awt.datatransfer.UnsupportedFlavorException;
import java.awt.event.MouseEvent;

/** Represents a single column within a {@link Outline} panel. */
public class Column implements Transferable {
    /** The data flavor for this class. */
    public static final DataFlavor DATA_FLAVOR = new DataFlavor(Column.class, "Outline Column");
    private             HeaderCell mHeaderCell;
    private             Cell       mRowCell;
    private             boolean    mVisible;
    private             int        mWidth;
    private             String     mName;
    private             RetinaIcon mIcon;
    private             int        mID;
    private             String     mToolTipText;

    /**
     * Create a new outline column.
     *
     * @param id   The user-supplied ID.
     * @param name The name of the column.
     */
    public Column(int id, String name) {
        this(id, name, (RetinaIcon) null);
    }

    /**
     * Create a new outline column.
     *
     * @param id      The user-supplied ID.
     * @param name    The name of the column.
     * @param rowCell The cell to use for rows.
     */
    public Column(int id, String name, Cell rowCell) {
        this(id, name, null, null, rowCell);
    }

    /**
     * Create a new outline column.
     *
     * @param id   The user-supplied ID.
     * @param name The name of the column.
     * @param icon The icon to use in the header of the column, if any.
     */
    public Column(int id, String name, RetinaIcon icon) {
        this(id, name, icon, name);
    }

    /**
     * Create a new outline column.
     *
     * @param id          The user-supplied ID.
     * @param name        The name of the column.
     * @param toolTipText The tooltip text to use for this column.
     */
    public Column(int id, String name, String toolTipText) {
        this(id, name, null, toolTipText);
    }

    /**
     * Create a new outline column.
     *
     * @param id          The user-supplied ID.
     * @param name        The name of the column.
     * @param icon        The icon to use in the header of the column, if any.
     * @param toolTipText The tooltip text to use for this column.
     */
    public Column(int id, String name, RetinaIcon icon, String toolTipText) {
        this(id, name, icon, toolTipText, new TextCell());
    }

    /**
     * Create a new outline column.
     *
     * @param id          The user-supplied ID.
     * @param name        The name of the column.
     * @param toolTipText The tooltip text to use for this column.
     * @param rowCell     The cell to use for rows.
     */
    public Column(int id, String name, String toolTipText, Cell rowCell) {
        this(id, name, null, toolTipText, rowCell);
    }

    /**
     * Create a new outline column.
     *
     * @param id          The user-supplied ID.
     * @param name        The name of the column.
     * @param icon        The icon to use in the header of the column, if any.
     * @param toolTipText The tooltip text to use for this column.
     * @param rowCell     The cell to use for rows.
     */
    public Column(int id, String name, RetinaIcon icon, String toolTipText, Cell rowCell) {
        mName = name == null || name.isEmpty() ? " " : name;
        mIcon = icon;
        mToolTipText = toolTipText;
        mHeaderCell = new HeaderCell();
        mRowCell = rowCell;
        mVisible = true;
        mWidth = -1;
        mID = id;
    }

    /** @return The user-supplied ID. */
    public int getID() {
        return mID;
    }

    /** @param id The user-supplied ID. */
    public void setID(int id) {
        mID = id;
    }

    /**
     * @return {@code true} if this column should be sorted in ascending order.
     */
    public boolean isSortAscending() {
        return getHeaderCell().isSortAscending();
    }

    /**
     * Sets the sort criteria for this column.
     *
     * @param sequence  The column's sort sequence. Use {@code -1} if it has none.
     * @param ascending Pass in {@code true} for an ascending sort.
     */
    public void setSortCriteria(int sequence, boolean ascending) {
        getHeaderCell().setSortCriteria(sequence, ascending);
    }

    /** @return The column's sort sequence, or {@code -1} if it has none. */
    public int getSortSequence() {
        return getHeaderCell().getSortSequence();
    }

    /**
     * Draws the header cell.
     *
     * @param outline The {@link Outline} being drawn.
     * @param gc      The graphics context to use.
     * @param bounds  The bounds of the cell.
     */
    public void drawHeaderCell(Outline outline, Graphics gc, Rectangle bounds) {
        getHeaderCell().drawCell(outline, gc, bounds, null, this, false, true);
    }

    /**
     * Draws a row cell.
     *
     * @param outline  The {@link Outline} being drawn.
     * @param gc       The graphics context to use.
     * @param bounds   The bounds of the cell.
     * @param row      The row the cell data is to come from.
     * @param selected Pass in {@code true} if the cell should be drawn in its selected state.
     * @param active   Pass in {@code true} if the cell should be drawn in its active state.
     */
    public void drawRowCell(Outline outline, Graphics gc, Rectangle bounds, Row row, boolean selected, boolean active) {
        getRowCell(row).drawCell(outline, gc, bounds, row, this, selected, active);
    }

    /** @return The cell used for the header. */
    public HeaderCell getHeaderCell() {
        return mHeaderCell;
    }

    /** @param cell The cell used for the header. */
    public void setHeaderCell(HeaderCell cell) {
        mHeaderCell = cell;
    }

    /**
     * @param row The row to return a cell for. This parameter may be {@code null}, for example,
     *            during sort operations.
     * @return The cell used for the specified row.
     */
    public Cell getRowCell(Row row) {
        return mRowCell;
    }

    /** @param cell The cell used for rows. */
    public void setRowCell(Cell cell) {
        mRowCell = cell;
    }

    /** @return {@code true} if this column is visible. */
    public boolean isVisible() {
        return mVisible;
    }

    /** @param visible Whether this column is visible or not. */
    public void setVisible(boolean visible) {
        mVisible = visible;
    }

    /** @return The width of this column. */
    public int getWidth() {
        return mWidth;
    }

    /**
     * @param outline The outline using this column.
     * @param width   The width of this column.
     */
    public void setWidth(Outline outline, int width) {
        int minWidth = getPreferredHeaderWidth(outline);
        if (width < minWidth && width != -1) {
            width = minWidth;
        }
        mWidth = width;
    }

    /**
     * @param outline The outline using this column.
     * @return The preferred width of this column's header.
     */
    public int getPreferredHeaderWidth(Outline outline) {
        return getHeaderCell().getPreferredWidth(outline, null, this);
    }

    /**
     * @param outline The outline using this column.
     * @return The preferred height of this column's header.
     */
    public int getPreferredHeaderHeight(Outline outline) {
        return getHeaderCell().getPreferredHeight(outline, null, this);
    }

    /**
     * @param outline The outline using this column.
     * @return The preferred width of this column.
     */
    public int getPreferredWidth(Outline outline) {
        Scale        scale          = Scale.get(outline);
        int          preferredWidth = getPreferredHeaderWidth(outline);
        OutlineModel model          = outline.getModel();
        for (Row row : model.getRows()) {
            int width = getRowCell(row).getPreferredWidth(outline, row, this) + scale.scale(model.getIndentWidth(row, this));
            if (width > preferredWidth) {
                preferredWidth = width;
            }
        }
        return preferredWidth;
    }

    /**
     * @param name The name of this column, which can be retrieved by using {@link #toString()}.
     */
    public void setName(String name) {
        mName = name == null || name.isEmpty() ? " " : name;
    }

    @Override
    public String toString() {
        return mName;
    }

    /** @return The name of this column, with any new lines replaced with spaces. */
    public String getSanitizedName() {
        return mName.replaceAll("\n", " ");
    }

    /** @return The icon used in the header of this column, if any. */
    public RetinaIcon getIcon() {
        return mIcon;
    }

    /** @param icon The icon used in the header of this column. */
    public void setIcon(RetinaIcon icon) {
        mIcon = icon;
    }

    /** @param text The tooltip text for this column. */
    public void setToolTipText(String text) {
        mToolTipText = text;
    }

    /**
     * @param event  The event that triggered the tooltip.
     * @param bounds The bounds of the header cell.
     * @return The tooltip text for this column, typically associated with the header.
     */
    public String getToolTipText(MouseEvent event, Rectangle bounds) {
        return mToolTipText;
    }

    @Override
    public DataFlavor[] getTransferDataFlavors() {
        return new DataFlavor[]{DATA_FLAVOR, DataFlavor.stringFlavor};
    }

    @Override
    public boolean isDataFlavorSupported(DataFlavor flavor) {
        return DATA_FLAVOR.equals(flavor) || DataFlavor.stringFlavor.equals(flavor);
    }

    @Override
    public Object getTransferData(DataFlavor flavor) throws UnsupportedFlavorException {
        if (DATA_FLAVOR.equals(flavor)) {
            return this;
        }
        if (DataFlavor.stringFlavor.equals(flavor)) {
            return getSanitizedName();
        }
        throw new UnsupportedFlavorException(flavor);
    }
}
