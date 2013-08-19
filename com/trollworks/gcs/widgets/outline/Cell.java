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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets.outline;

import java.awt.Cursor;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;

/** Represents cells in a {@link Outline}. */
public interface Cell {
	/**
	 * Draws the cell.
	 * 
	 * @param outline The {@link Outline} being drawn.
	 * @param gc The graphics context to use.
	 * @param bounds The bounds of the cell.
	 * @param row The row to draw.
	 * @param column The column to draw.
	 * @param selected Pass in <code>true</code> if the cell should be drawn in its selected
	 *            state.
	 * @param active Pass in <code>true</code> if the cell should be drawn in its active state.
	 */
	public void drawCell(Outline outline, Graphics gc, Rectangle bounds, Row row, Column column, boolean selected, boolean active);

	/**
	 * @param row The row to get data from.
	 * @param column The column to get data from.
	 * @return The preferred width of the cell for the specified data.
	 */
	public int getPreferredWidth(Row row, Column column);

	/**
	 * @param row The row to get data from.
	 * @param column The column to get data from.
	 * @return The preferred height of the cell for the specified data.
	 */
	public int getPreferredHeight(Row row, Column column);

	/**
	 * Compare the column in row one with row two.
	 * 
	 * @param column The column to compare.
	 * @param one The first row.
	 * @param two The second row.
	 * @return <code>< 0</code> if row one is less than row two, <code>0</code> if they are
	 *         equal, and <code>> 0</code> if row one is greater than row two.
	 */
	public int compare(Column column, Row one, Row two);

	/**
	 * @param event The {@link MouseEvent} that caused the tooltip to be shown.
	 * @param bounds The bounds of the cell.
	 * @param row The row to get data from.
	 * @param column The column to get data from.
	 * @return The cursor appropriate for the location within the cell.
	 */
	public Cursor getCursor(MouseEvent event, Rectangle bounds, Row row, Column column);

	/**
	 * @param event The {@link MouseEvent} that caused the tooltip to be shown.
	 * @param bounds The bounds of the cell.
	 * @param row The row to get data from.
	 * @param column The column to get data from.
	 * @return The tooltip string for this cell.
	 */
	public String getToolTipText(MouseEvent event, Rectangle bounds, Row row, Column column);

	/**
	 * @return <code>true</code> if this cell wants to participate in dynamic row layout.
	 */
	public boolean participatesInDynamicRowLayout();

	/**
	 * Called when a mouse click has occurred on the cell.
	 * 
	 * @param event The {@link MouseEvent} that caused the call.
	 * @param bounds The bounds of the cell.
	 * @param row The row to get data from.
	 * @param column The column to get data from.
	 */
	public void mouseClicked(MouseEvent event, Rectangle bounds, Row row, Column column);
}
