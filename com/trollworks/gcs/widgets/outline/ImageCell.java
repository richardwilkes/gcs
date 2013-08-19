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

package com.trollworks.gcs.widgets.outline;

import com.trollworks.gcs.utility.text.NumericStringComparator;

import java.awt.Cursor;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.awt.image.BufferedImage;

import javax.swing.SwingConstants;

/** Represents image cells in a {@link Outline}. */
public class ImageCell implements Cell {
	private int	mHAlignment;
	private int	mVAlignment;

	/** Create a new image cell renderer. */
	public ImageCell() {
		this(SwingConstants.CENTER, SwingConstants.CENTER);
	}

	/**
	 * Create a new image cell renderer.
	 * 
	 * @param hAlignment The image horizontal alignment to use.
	 * @param vAlignment The image vertical alignment to use.
	 */
	public ImageCell(int hAlignment, int vAlignment) {
		mHAlignment = hAlignment;
		mVAlignment = vAlignment;
	}

	public int compare(Column column, Row one, Row two) {
		String oneText = one.getDataAsText(column);
		String twoText = two.getDataAsText(column);

		return NumericStringComparator.caselessCompareStrings(oneText != null ? oneText : "", twoText != null ? twoText : ""); //$NON-NLS-1$ //$NON-NLS-2$
	}

	/**
	 * @param row The row to use.
	 * @param column The column to use.
	 * @param selected Whether the row is selected.
	 * @param active Whether the outline is active.
	 * @return The icon, if any.
	 */
	protected BufferedImage getIcon(Row row, Column column, @SuppressWarnings("unused") boolean selected, @SuppressWarnings("unused") boolean active) {
		Object data = row.getData(column);

		return data instanceof BufferedImage ? (BufferedImage) data : null;
	}

	public void drawCell(Outline outline, Graphics gc, Rectangle bounds, Row row, Column column, boolean selected, boolean active) {
		if (row != null) {
			BufferedImage image = getIcon(row, column, selected, active);

			if (image != null) {
				int x = bounds.x;
				int y = bounds.y;

				if (mHAlignment != SwingConstants.LEFT) {
					int hDelta = bounds.width - image.getWidth();

					if (mHAlignment == SwingConstants.CENTER) {
						hDelta /= 2;
					}
					x += hDelta;
				}

				if (mVAlignment != SwingConstants.TOP) {
					int vDelta = bounds.height - image.getHeight();

					if (mVAlignment == SwingConstants.CENTER) {
						vDelta /= 2;
					}
					y += vDelta;
				}

				gc.drawImage(image, x, y, null);
			}
		}
	}

	public int getPreferredWidth(Row row, Column column) {
		Object data = row != null ? row.getData(column) : null;

		return data instanceof BufferedImage ? ((BufferedImage) data).getWidth() : 0;
	}

	public int getPreferredHeight(Row row, Column column) {
		Object data = row != null ? row.getData(column) : null;

		return data instanceof BufferedImage ? ((BufferedImage) data).getHeight() : 0;
	}

	public Cursor getCursor(MouseEvent event, Rectangle bounds, Row row, Column column) {
		return Cursor.getDefaultCursor();
	}

	public String getToolTipText(MouseEvent event, Rectangle bounds, Row row, Column column) {
		return null;
	}

	public boolean participatesInDynamicRowLayout() {
		return false;
	}

	public void mouseClicked(MouseEvent event, Rectangle bounds, Row row, Column column) {
		// Does nothing
	}
}
