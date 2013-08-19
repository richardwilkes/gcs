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

package com.trollworks.gcs.widgets.outline;

import com.trollworks.gcs.app.GCSFonts;
import com.trollworks.ttk.widgets.outline.Column;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.Row;
import com.trollworks.ttk.widgets.outline.TextCell;

import java.awt.Color;
import java.awt.Font;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;

import javax.swing.UIManager;

/** Represents cells in a {@link Outline}. */
public class ListTextCell extends TextCell {
	/**
	 * Create a new text cell.
	 * 
	 * @param alignment The horizontal text alignment to use.
	 * @param wrapped Pass in <code>true</code> to enable wrapping.
	 */
	public ListTextCell(int alignment, boolean wrapped) {
		super(alignment, wrapped);
	}

	@Override
	public Font getFont(Row row, Column column) {
		return UIManager.getFont(GCSFonts.KEY_FIELD);
	}

	@Override
	public Color getColor(boolean selected, boolean active, Row row, Column column) {
		if (row instanceof ListRow && !((ListRow) row).isSatisfied()) {
			return Color.red;
		}
		return super.getColor(selected, active, row, column);
	}

	@Override
	public String getToolTipText(MouseEvent event, Rectangle bounds, Row row, Column column) {
		if (!(row instanceof ListRow) || ((ListRow) row).isSatisfied()) {
			return super.getToolTipText(event, bounds, row, column);
		}
		return ((ListRow) row).getReasonForUnsatisfied();
	}
}
