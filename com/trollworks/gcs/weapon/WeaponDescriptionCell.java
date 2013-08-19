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
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.app.GCSFonts;
import com.trollworks.ttk.text.NumericStringComparator;
import com.trollworks.ttk.text.TextDrawing;
import com.trollworks.ttk.widgets.outline.Cell;
import com.trollworks.ttk.widgets.outline.Column;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.Row;

import java.awt.Cursor;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;

import javax.swing.SwingConstants;
import javax.swing.UIManager;

/**
 * A {@link Cell} for displaying both a piece of primary information and a piece of secondary
 * information for a {@link WeaponDisplayRow}.
 */
public class WeaponDescriptionCell implements Cell {
	private static final int	H_MARGIN	= 2;

	/**
	 * @param row The row to use.
	 * @return The primary text to display.
	 */
	protected String getPrimaryText(WeaponDisplayRow row) {
		return row.getWeapon().toString();
	}

	/**
	 * @param row The row to use.
	 * @return The secondary text to display.
	 */
	protected String getSecondaryText(WeaponDisplayRow row) {
		return row.getWeapon().getNotes();
	}

	@Override
	public void drawCell(Outline outline, Graphics gc, Rectangle bounds, Row row, Column column, boolean selected, boolean active) {
		WeaponDisplayRow theRow = (WeaponDisplayRow) row;
		Rectangle insetBounds = new Rectangle(bounds.x + H_MARGIN, bounds.y, bounds.width - H_MARGIN * 2, bounds.height);
		String notes = getSecondaryText(theRow);
		Font font = UIManager.getFont(GCSFonts.KEY_FIELD);
		int pos;

		gc.setColor(Outline.getListForeground(selected, active));
		gc.setFont(font);
		pos = TextDrawing.draw(gc, insetBounds, getPrimaryText(theRow), SwingConstants.LEFT, SwingConstants.TOP);
		if (notes.trim().length() > 0) {
			insetBounds.height -= pos - insetBounds.y;
			insetBounds.y = pos;
			gc.setFont(UIManager.getFont(GCSFonts.KEY_FIELD_NOTES));
			TextDrawing.draw(gc, insetBounds, notes, SwingConstants.LEFT, SwingConstants.TOP);
		}
	}

	@Override
	public int getPreferredWidth(Row row, Column column) {
		WeaponDisplayRow theRow = (WeaponDisplayRow) row;
		int width = TextDrawing.getWidth(UIManager.getFont(GCSFonts.KEY_FIELD), getPrimaryText(theRow));
		String notes = getSecondaryText(theRow);

		if (notes.trim().length() > 0) {
			int notesWidth = TextDrawing.getWidth(UIManager.getFont(GCSFonts.KEY_FIELD_NOTES), notes);

			if (notesWidth > width) {
				width = notesWidth;
			}
		}
		return width + H_MARGIN * 2;
	}

	@Override
	public int getPreferredHeight(Row row, Column column) {
		WeaponDisplayRow theRow = (WeaponDisplayRow) row;
		Font font = UIManager.getFont(GCSFonts.KEY_FIELD);
		int height = TextDrawing.getPreferredSize(font, wrap(theRow, column, getPrimaryText(theRow), font)).height;
		String notes = getSecondaryText(theRow);

		if (notes.trim().length() > 0) {
			font = UIManager.getFont(GCSFonts.KEY_FIELD_NOTES);
			height += TextDrawing.getPreferredSize(font, wrap(theRow, column, notes, font)).height;
		}
		return height;
	}

	private String wrap(WeaponDisplayRow row, Column column, String text, Font font) {
		int width = column.getWidth();

		if (width == -1) {
			return text;
		}
		return TextDrawing.wrapToPixelWidth(font, text, width - (row.getOwner().getIndentWidth(row, column) + H_MARGIN * 2));
	}

	@Override
	public int compare(Column column, Row one, Row two) {
		WeaponDisplayRow r1 = (WeaponDisplayRow) one;
		WeaponDisplayRow r2 = (WeaponDisplayRow) two;
		int result = NumericStringComparator.caselessCompareStrings(getPrimaryText(r1), getPrimaryText(r2));

		if (result == 0) {
			result = NumericStringComparator.caselessCompareStrings(getSecondaryText(r1), getSecondaryText(r2));
		}
		return result;
	}

	@Override
	public Cursor getCursor(MouseEvent event, Rectangle bounds, Row row, Column column) {
		return Cursor.getDefaultCursor();
	}

	@Override
	public String getToolTipText(MouseEvent event, Rectangle bounds, Row row, Column column) {
		return null;
	}

	@Override
	public boolean participatesInDynamicRowLayout() {
		return true;
	}

	@Override
	public void mouseClicked(MouseEvent event, Rectangle bounds, Row row, Column column) {
		// Does nothing
	}
}
