/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.widgets.outline;

import com.trollworks.gcs.app.GCSFonts;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.ui.widget.outline.HeaderCell;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.Row;

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
