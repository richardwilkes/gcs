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

import com.trollworks.toolkit.widget.TKPanel;

import java.awt.Color;
import java.awt.Cursor;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.event.ActionListener;
import java.awt.event.MouseEvent;

/** Represents cells in a {@link TKOutline}. */
public interface TKCell {
	/**
	 * Draws the cell.
	 * 
	 * @param g2d The graphics context to use.
	 * @param bounds The bounds of the cell.
	 * @param row The row to draw.
	 * @param column The column to draw.
	 * @param selected Pass in <code>true</code> if the cell should be drawn in its selected
	 *            state.
	 * @param active Pass in <code>true</code> if the cell should be drawn in its active state.
	 */
	public void drawCell(Graphics2D g2d, Rectangle bounds, TKRow row, TKColumn column, boolean selected, boolean active);

	/**
	 * @param row The row to get data from.
	 * @param column The column to get data from.
	 * @return The preferred width of the cell for the specified data.
	 */
	public int getPreferredWidth(TKRow row, TKColumn column);

	/**
	 * @param row The row to get data from.
	 * @param column The column to get data from.
	 * @return The preferred height of the cell for the specified data.
	 */
	public int getPreferredHeight(TKRow row, TKColumn column);

	/**
	 * Compare the column in row one with row two.
	 * 
	 * @param column The column to compare.
	 * @param one The first row.
	 * @param two The second row.
	 * @return <code>< 0</code> if row one is less than row two, <code>0</code> if they are
	 *         equal, and <code>> 0</code> if row one is greater than row two.
	 */
	public int compare(TKColumn column, TKRow one, TKRow two);

	/**
	 * @param row The row the might be dragged.
	 * @param column The column that would be used to drag the row.
	 * @return <code>true</code> if this cell can be used to drag the row around.
	 */
	public boolean isRowDragHandle(TKRow row, TKColumn column);

	/**
	 * @param row The row to check.
	 * @param column The column to check.
	 * @param viaSingleClick Pass in <code>true</code> if the check is to determine whether the
	 *            cell is editable via a single click, <code>false</code> to determine whether the
	 *            cell is editable at all.
	 * @return <code>true</code> if this cell can be edited.
	 */
	public boolean isEditable(TKRow row, TKColumn column, boolean viaSingleClick);

	/**
	 * The outline will add itself as an {@link ActionListener} and will be expecting to see an
	 * action command equal to {@link TKOutline#CMD_UPDATE_FROM_EDITOR} whenever it should update
	 * itself from the editor. This action is only necessary when updates should be applied prior to
	 * a call to {@link #stopEditing(TKPanel)}.
	 * 
	 * @param rowBackground The background color of the row the cell is in.
	 * @param row The row to get data from.
	 * @param column The column to get data from.
	 * @param mouseInitiated <code>true</code> if a mouse click initiated the need for an editor.
	 * @return An editor for the data.
	 */
	public TKPanel getEditor(Color rowBackground, TKRow row, TKColumn column, boolean mouseInitiated);

	/**
	 * Asks the editor for the edited version of the object.
	 * 
	 * @param editor The editor that was returned from
	 *            {@link #getEditor(Color, TKRow, TKColumn, boolean)}.
	 * @return The edited version of the object.
	 */
	public Object getEditedObject(TKPanel editor);

	/**
	 * Tells the cell to stop editing.
	 * 
	 * @param editor The editor that was returned from
	 *            {@link #getEditor(Color, TKRow, TKColumn, boolean)}.
	 * @return The edited version of the object.
	 */
	public Object stopEditing(TKPanel editor);

	/**
	 * @param event The {@link MouseEvent} that caused the tooltip to be shown.
	 * @param bounds The bounds of the cell.
	 * @param row The row to get data from.
	 * @param column The column to get data from.
	 * @return The cursor appropriate for the location within the cell.
	 */
	public Cursor getCursor(MouseEvent event, Rectangle bounds, TKRow row, TKColumn column);

	/**
	 * @param event The {@link MouseEvent} that caused the tooltip to be shown.
	 * @param bounds The bounds of the cell.
	 * @param row The row to get data from.
	 * @param column The column to get data from.
	 * @return The tooltip string for this cell.
	 */
	public String getToolTipText(MouseEvent event, Rectangle bounds, TKRow row, TKColumn column);

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
	public void mouseClicked(MouseEvent event, Rectangle bounds, TKRow row, TKColumn column);
}
