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
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets.outline;

import com.trollworks.gcs.app.GCSFonts;
import com.trollworks.ttk.widgets.outline.Column;
import com.trollworks.ttk.widgets.outline.HeaderCell;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.Row;

import java.awt.Color;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Rectangle;

import javax.swing.UIManager;

/** Used to draw headers in the lists. */
public class ListHeaderCell extends HeaderCell {
	private boolean	mForSheet;

	/**
	 * Create a new header cell.
	 * 
	 * @param forSheet Whether the header will be displayed in the sheet or not.
	 */
	public ListHeaderCell(boolean forSheet) {
		super();
		mForSheet = forSheet;
	}

	@Override
	public Font getFont(Row row, Column column) {
		return UIManager.getFont(GCSFonts.KEY_LABEL);
	}

	@Override
	public void drawCell(Outline outline, Graphics gc, Rectangle bounds, Row row, Column column, boolean selected, boolean active) {
		if (mForSheet) {
			drawCellSuper(outline, gc, bounds, row, column, selected, active);
		} else {
			super.drawCell(outline, gc, bounds, row, column, selected, active);
		}
	}

	@Override
	public int getPreferredWidth(Row row, Column column) {
		int width = super.getPreferredWidth(row, column);

		if (mForSheet) {
			width -= SORTER_WIDTH + 4;
		}
		return width;
	}

	@Override
	public Color getColor(boolean selected, boolean active, Row row, Column column) {
		if (mForSheet) {
			return Color.white;
		}
		return super.getColor(selected, active, row, column);
	}
}
