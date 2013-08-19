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

import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKHeaderCell;
import com.trollworks.toolkit.widget.outline.TKRow;

import java.awt.Color;
import java.awt.Graphics2D;
import java.awt.Rectangle;

/** Used to draw headers in a {@link CSOutline}. */
public class CSHeaderCell extends TKHeaderCell {
	private boolean	mForSheet;

	/**
	 * Create a new header cell.
	 * 
	 * @param forSheet Whether the header will be displayed in the sheet or not.
	 */
	public CSHeaderCell(boolean forSheet) {
		super();
		mForSheet = forSheet;
		if (mForSheet) {
			setColor(Color.white);
		}
	}

	@Override public String getFontKey() {
		return CSFont.KEY_LABEL;
	}

	@Override public void drawCell(Graphics2D g2d, Rectangle bounds, TKRow row, TKColumn column, boolean selected, boolean active) {
		if (mForSheet) {
			drawCellSuper(g2d, bounds, row, column, selected, active);
		} else {
			super.drawCell(g2d, bounds, row, column, selected, active);
		}
	}

	@Override public int getPreferredWidth(TKRow row, TKColumn column) {
		int width = super.getPreferredWidth(row, column);

		if (mForSheet) {
			width -= SORTER_WIDTH + 4;
		}
		return width;
	}
}
