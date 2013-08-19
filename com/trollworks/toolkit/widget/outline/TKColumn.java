/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.toolkit.widget.outline;

import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.datatransfer.DataFlavor;
import java.awt.datatransfer.Transferable;
import java.awt.datatransfer.UnsupportedFlavorException;
import java.awt.event.MouseEvent;
import java.awt.image.BufferedImage;

/** Represents a single column within a {@link TKOutline} panel. */
public class TKColumn implements Transferable {
	/** The data flavor for this class. */
	public static final DataFlavor	DATA_FLAVOR	= new DataFlavor(TKColumn.class, "Outline Column"); //$NON-NLS-1$
	private TKHeaderCell			mHeaderCell;
	private TKCell					mRowCell;
	private boolean					mVisible;
	private int						mWidth;
	private String					mName;
	private BufferedImage			mIcon;
	private int						mID;
	private String					mToolTipText;

	/**
	 * Create a new outline column.
	 * 
	 * @param id The user-supplied ID.
	 * @param name The name of the column.
	 */
	public TKColumn(int id, String name) {
		this(id, name, (BufferedImage) null);
	}

	/**
	 * Create a new outline column.
	 * 
	 * @param id The user-supplied ID.
	 * @param name The name of the column.
	 * @param rowCell The cell to use for rows.
	 */
	public TKColumn(int id, String name, TKCell rowCell) {
		this(id, name, null, null, rowCell);
	}

	/**
	 * Create a new outline column.
	 * 
	 * @param id The user-supplied ID.
	 * @param name The name of the column.
	 * @param icon The icon to use in the header of the column, if any.
	 */
	public TKColumn(int id, String name, BufferedImage icon) {
		this(id, name, icon, name);
	}

	/**
	 * Create a new outline column.
	 * 
	 * @param id The user-supplied ID.
	 * @param name The name of the column.
	 * @param toolTipText The tooltip text to use for this column.
	 */
	public TKColumn(int id, String name, String toolTipText) {
		this(id, name, null, toolTipText);
	}

	/**
	 * Create a new outline column.
	 * 
	 * @param id The user-supplied ID.
	 * @param name The name of the column.
	 * @param icon The icon to use in the header of the column, if any.
	 * @param toolTipText The tooltip text to use for this column.
	 */
	public TKColumn(int id, String name, BufferedImage icon, String toolTipText) {
		this(id, name, icon, toolTipText, new TKTextCell());
	}

	/**
	 * Create a new outline column.
	 * 
	 * @param id The user-supplied ID.
	 * @param name The name of the column.
	 * @param toolTipText The tooltip text to use for this column.
	 * @param rowCell The cell to use for rows.
	 */
	public TKColumn(int id, String name, String toolTipText, TKCell rowCell) {
		this(id, name, null, toolTipText, rowCell);
	}

	/**
	 * Create a new outline column.
	 * 
	 * @param id The user-supplied ID.
	 * @param name The name of the column.
	 * @param icon The icon to use in the header of the column, if any.
	 * @param toolTipText The tooltip text to use for this column.
	 * @param rowCell The cell to use for rows.
	 */
	public TKColumn(int id, String name, BufferedImage icon, String toolTipText, TKCell rowCell) {
		mName = name == null || name.length() == 0 ? " " : name; //$NON-NLS-1$
		mIcon = icon;
		mToolTipText = toolTipText;
		mHeaderCell = new TKHeaderCell();
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
	 * @return <code>true</code> if this column should be sorted in ascending order.
	 */
	public boolean isSortAscending() {
		return getHeaderCell().isSortAscending();
	}

	/**
	 * Sets the sort criteria for this column.
	 * 
	 * @param sequence The column's sort sequence. Use <code>-1</code> if it has none.
	 * @param ascending Pass in <code>true</code> for an ascending sort.
	 */
	public void setSortCriteria(int sequence, boolean ascending) {
		getHeaderCell().setSortCriteria(sequence, ascending);
	}

	/** @return The column's sort sequence, or <code>-1</code> if it has none. */
	public int getSortSequence() {
		return getHeaderCell().getSortSequence();
	}

	/**
	 * Draws the header cell.
	 * 
	 * @param g2d The graphics context to use.
	 * @param bounds The bounds of the cell.
	 */
	public void drawHeaderCell(Graphics2D g2d, Rectangle bounds) {
		getHeaderCell().drawCell(g2d, bounds, null, this, false, true);
	}

	/**
	 * Draws a row cell.
	 * 
	 * @param g2d The graphics context to use.
	 * @param bounds The bounds of the cell.
	 * @param row The row the cell data is to come from.
	 * @param selected Pass in <code>true</code> if the cell should be drawn in its selected
	 *            state.
	 * @param active Pass in <code>true</code> if the cell should be drawn in its active state.
	 */
	public void drawRowCell(Graphics2D g2d, Rectangle bounds, TKRow row, boolean selected, boolean active) {
		getRowCell(row).drawCell(g2d, bounds, row, this, selected, active);
	}

	/** @return The cell used for the header. */
	public TKHeaderCell getHeaderCell() {
		return mHeaderCell;
	}

	/** @param cell The cell used for the header. */
	public void setHeaderCell(TKHeaderCell cell) {
		mHeaderCell = cell;
	}

	/**
	 * @param row The row to return a cell for. This parameter may be <code>null</code>, for
	 *            example, during sort operations.
	 * @return The cell used for the specified row.
	 */
	public TKCell getRowCell(@SuppressWarnings("unused") TKRow row) {
		return mRowCell;
	}

	/** @param cell The cell used for rows. */
	public void setRowCell(TKCell cell) {
		mRowCell = cell;
	}

	/** @return <code>true</code> if this column is visible. */
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

	/** @param width The width of this column. */
	public void setWidth(int width) {
		int minWidth = getPreferredHeaderWidth();

		if (width < minWidth && width != -1) {
			width = minWidth;
		}
		mWidth = width;
	}

	/** @return The preferred width of this column's header. */
	public int getPreferredHeaderWidth() {
		return getHeaderCell().getPreferredWidth(null, this);
	}

	/** @return The preferred height of this column's header. */
	public int getPreferredHeaderHeight() {
		return getHeaderCell().getPreferredHeight(null, this);
	}

	/**
	 * @param outline The outline using this column.
	 * @return The preferred width of this column.
	 */
	public int getPreferredWidth(TKOutline outline) {
		int preferredWidth = getPreferredHeaderWidth();
		TKOutlineModel model = outline.getModel();

		for (TKRow row : model.getRows()) {
			int width = getRowCell(row).getPreferredWidth(row, this) + model.getIndentWidth(row, this);

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
		mName = name == null || name.length() == 0 ? " " : name; //$NON-NLS-1$
	}

	@Override public String toString() {
		return mName;
	}

	/** @return The name of this column, with any new lines replaced with spaces. */
	public String getSanitizedName() {
		return mName.replaceAll("\n", " "); //$NON-NLS-1$ //$NON-NLS-2$
	}

	/** @return The icon used in the header of this column, if any. */
	public BufferedImage getIcon() {
		return mIcon;
	}

	/** @param icon The icon used in the header of this column. */
	public void setIcon(BufferedImage icon) {
		mIcon = icon;
	}

	/** @param text The tooltip text for this column. */
	public void setToolTipText(String text) {
		mToolTipText = text;
	}

	/**
	 * @param event The event that triggered the tooltip.
	 * @param bounds The bounds of the header cell.
	 * @return The tooltip text for this column, typically associated with the header.
	 */
	public String getToolTipText(@SuppressWarnings("unused") MouseEvent event, @SuppressWarnings("unused") Rectangle bounds) {
		return mToolTipText;
	}

	public DataFlavor[] getTransferDataFlavors() {
		return new DataFlavor[] { DATA_FLAVOR, DataFlavor.stringFlavor };
	}

	public boolean isDataFlavorSupported(DataFlavor flavor) {
		return DATA_FLAVOR.equals(flavor) || DataFlavor.stringFlavor.equals(flavor);
	}

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
