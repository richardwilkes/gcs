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

package com.trollworks.gcs.ui.common;

import com.trollworks.gcs.model.CMRow;
import com.trollworks.toolkit.text.TKNumericStringComparator;
import com.trollworks.toolkit.text.TKTextDrawing;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.outline.TKCell;
import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKRow;

import java.awt.Color;
import java.awt.Cursor;
import java.awt.Font;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;

/**
 * A {@link TKCell} for displaying both a piece of primary information and a piece of secondary
 * information for a {@link CMRow}.
 */
public class CSMultiCell implements TKCell {
	private static final int	H_MARGIN	= 2;

	/**
	 * @param row The row to use.
	 * @return The primary text to display.
	 */
	protected String getPrimaryText(CMRow row) {
		return row.toString();
	}

	/**
	 * @param row The row to use.
	 * @return The secondary text to display.
	 */
	protected String getSecondaryText(CMRow row) {
		String modifierNotes = row.getModifierNotes();
		String notes = row.getNotes();
		return modifierNotes.length() == 0 ? notes : modifierNotes + '\n' + notes;
	}

	public void drawCell(Graphics2D g2d, Rectangle bounds, TKRow row, TKColumn column, boolean selected, boolean active) {
		CMRow theRow = (CMRow) row;
		Rectangle insetBounds = new Rectangle(bounds.x + H_MARGIN, bounds.y, bounds.width - H_MARGIN * 2, bounds.height);
		String notes = getSecondaryText(theRow);
		Font font = TKFont.lookup(CSFont.KEY_FIELD);
		Color color;
		int pos;

		if (theRow.isSatisfied()) {
			if (selected && active) {
				color = TKColor.HIGHLIGHTED_TEXT;
			} else {
				color = TKColor.TEXT;
			}
		} else {
			color = Color.red;
		}
		g2d.setColor(color);
		g2d.setFont(font);
		pos = TKTextDrawing.draw(g2d, insetBounds, getPrimaryText(theRow), TKAlignment.LEFT, TKAlignment.TOP);
		if (notes.trim().length() > 0) {
			insetBounds.height -= pos - insetBounds.y;
			insetBounds.y = pos;
			g2d.setFont(TKFont.lookup(CSFont.KEY_FIELD_NOTES));
			TKTextDrawing.draw(g2d, insetBounds, notes, TKAlignment.LEFT, TKAlignment.TOP);
		}
	}

	public int getPreferredWidth(TKRow row, TKColumn column) {
		CMRow theRow = (CMRow) row;
		int width = TKTextDrawing.getWidth(TKFont.lookup(CSFont.KEY_FIELD), null, getPrimaryText(theRow));
		String notes = getSecondaryText(theRow);

		if (notes.trim().length() > 0) {
			int notesWidth = TKTextDrawing.getWidth(TKFont.lookup(CSFont.KEY_FIELD_NOTES), null, notes);

			if (notesWidth > width) {
				width = notesWidth;
			}
		}
		return width + H_MARGIN * 2;
	}

	public int getPreferredHeight(TKRow row, TKColumn column) {
		CMRow theRow = (CMRow) row;
		Font font = TKFont.lookup(CSFont.KEY_FIELD);
		int height = TKTextDrawing.getPreferredSize(font, null, wrap(theRow, column, getPrimaryText(theRow), font)).height;
		String notes = getSecondaryText(theRow);

		if (notes.trim().length() > 0) {
			font = TKFont.lookup(CSFont.KEY_FIELD_NOTES);
			height += TKTextDrawing.getPreferredSize(font, null, wrap(theRow, column, notes, font)).height;
		}
		return height;
	}

	private String wrap(CMRow row, TKColumn column, String text, Font font) {
		int width = column.getWidth();

		if (width == -1) {
			return text;
		}
		return TKTextDrawing.wrapToPixelWidth(font, null, text, width - (row.getOwner().getIndentWidth(row, column) + H_MARGIN * 2));
	}

	public int compare(TKColumn column, TKRow one, TKRow two) {
		CMRow r1 = (CMRow) one;
		CMRow r2 = (CMRow) two;
		int result = TKNumericStringComparator.caselessCompareStrings(getPrimaryText(r1), getPrimaryText(r2));

		if (result == 0) {
			result = TKNumericStringComparator.caselessCompareStrings(getSecondaryText(r1), getSecondaryText(r2));
		}
		return result;
	}

	public boolean isRowDragHandle(TKRow row, TKColumn column) {
		return true;
	}

	public boolean isEditable(TKRow row, TKColumn column, boolean viaSingleClick) {
		return false;
	}

	public TKPanel getEditor(Color rowBackground, TKRow row, TKColumn column, boolean mouseInitiated) {
		return null;
	}

	public Object getEditedObject(TKPanel editor) {
		return null;
	}

	public Object stopEditing(TKPanel editor) {
		return null;
	}

	public Cursor getCursor(MouseEvent event, Rectangle bounds, TKRow row, TKColumn column) {
		return Cursor.getDefaultCursor();
	}

	public String getToolTipText(MouseEvent event, Rectangle bounds, TKRow row, TKColumn column) {
		CMRow theRow = (CMRow) row;

		return theRow.isSatisfied() ? null : theRow.getReasonForUnsatisfied();
	}

	public boolean participatesInDynamicRowLayout() {
		return true;
	}

	public void mouseClicked(MouseEvent event, Rectangle bounds, TKRow row, TKColumn column) {
		// Does nothing
	}
}
