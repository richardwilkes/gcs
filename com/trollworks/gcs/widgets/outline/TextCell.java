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

import com.trollworks.gcs.utility.text.NumericStringComparator;
import com.trollworks.gcs.utility.text.TextDrawing;

import java.awt.Color;
import java.awt.Cursor;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.awt.font.FontRenderContext;
import java.awt.image.BufferedImage;
import java.util.StringTokenizer;

import javax.swing.SwingConstants;
import javax.swing.UIManager;

/** Represents text cells in an {@link Outline}. */
public class TextCell implements Cell {
	/** The standard horizontal margin. */
	public static final int	H_MARGIN		= 2;
	/** The standard horizontal margin width. */
	public static final int	H_MARGIN_WIDTH	= H_MARGIN * 2;
	private int				mHAlignment;
	private boolean			mWrapped;

	/** Create a new text cell. */
	public TextCell() {
		this(SwingConstants.LEFT);
	}

	/**
	 * Create a new text cell.
	 * 
	 * @param alignment The horizontal text alignment to use.
	 */
	public TextCell(int alignment) {
		this(alignment, false);
	}

	/**
	 * Create a new text cell.
	 * 
	 * @param hAlignment The horizontal text alignment to use.
	 * @param wrapped Pass in <code>true</code> to enable wrapping.
	 */
	public TextCell(int hAlignment, boolean wrapped) {
		mHAlignment = hAlignment;
		mWrapped = wrapped;
	}

	@SuppressWarnings("unchecked") public int compare(Column column, Row one, Row two) {
		Object oneObj = one.getData(column);
		Object twoObj = two.getData(column);
		if (!(oneObj instanceof String) && oneObj.getClass() == twoObj.getClass() && oneObj instanceof Comparable) {
			return ((Comparable) oneObj).compareTo(twoObj);
		}
		return NumericStringComparator.caselessCompareStrings(one.getDataAsText(column), two.getDataAsText(column));
	}

	/**
	 * @param selected Whether or not the selected version of the color is needed.
	 * @param active Whether or not the active version of the color is needed.
	 * @param row The row.
	 * @param column The column.
	 * @return The foreground color.
	 */
	public Color getColor(boolean selected, boolean active, Row row, Column column) {
		return Outline.getListForeground(selected, active);
	}

	public int getPreferredWidth(Row row, Column column) {
		int width = TextDrawing.getPreferredSize(getFont(row, column), null, getPresentationText(row, column)).width;
		BufferedImage icon = row == null ? column.getIcon() : row.getIcon(column);
		if (icon != null) {
			width += icon.getWidth() + H_MARGIN;
		}
		return H_MARGIN_WIDTH + width;
	}

	public int getPreferredHeight(Row row, Column column) {
		Font font = getFont(row, column);
		int minHeight = TextDrawing.getPreferredSize(font, null, "Mg").height; //$NON-NLS-1$
		int height = TextDrawing.getPreferredSize(font, null, getPresentationText(row, column)).height;
		BufferedImage icon = row == null ? column.getIcon() : row.getIcon(column);
		if (icon != null) {
			int iconHeight = icon.getHeight();
			if (height < iconHeight) {
				height = iconHeight;
			}
		}
		return minHeight > height ? minHeight : height;
	}

	public void drawCell(Outline outline, Graphics gc, Rectangle bounds, Row row, Column column, boolean selected, boolean active) {
		Font font = getFont(row, column);
		FontRenderContext frc = ((Graphics2D) gc).getFontRenderContext();
		int ascent = gc.getFontMetrics(font).getAscent();
		StringTokenizer tokenizer = new StringTokenizer(getPresentationText(row, column), "\n", true); //$NON-NLS-1$
		int totalHeight = getPreferredHeight(row, column);
		int lineHeight = TextDrawing.getPreferredSize(font, frc, "Mg").height; //$NON-NLS-1$
		int lineCount = 0;
		BufferedImage icon = row == null ? column.getIcon() : row.getIcon(column);
		int left = icon == null ? 0 : icon.getWidth() + H_MARGIN;
		int cellWidth = bounds.width - (left + H_MARGIN_WIDTH);
		int vAlignment = getVAlignment();
		int hAlignment = getHAlignment();

		left += bounds.x + H_MARGIN;

		if (icon != null) {
			int iy = bounds.y;

			if (vAlignment != SwingConstants.TOP) {
				int ivDelta = bounds.height - icon.getHeight();

				if (vAlignment == SwingConstants.CENTER) {
					ivDelta /= 2;
				}
				iy += ivDelta;
			}
			gc.drawImage(icon, bounds.x + H_MARGIN, iy, null);
		}

		gc.setFont(font);
		while (tokenizer.hasMoreTokens()) {
			String token = tokenizer.nextToken();

			if (token.equals("\n")) { //$NON-NLS-1$
				lineCount++;
			} else {
				String text = TextDrawing.truncateIfNecessary(font, frc, token, cellWidth, getTruncationPolicy());
				int x = left;
				int y = bounds.y + ascent + lineHeight * lineCount;

				if (hAlignment != SwingConstants.LEFT) {
					int hDelta = cellWidth - TextDrawing.getWidth(font, frc, text);

					if (hAlignment == SwingConstants.CENTER) {
						hDelta /= 2;
					}
					x += hDelta;
				}

				if (vAlignment != SwingConstants.TOP) {
					float vDelta = bounds.height - totalHeight;

					if (vAlignment == SwingConstants.CENTER) {
						vDelta /= 2;
					}
					y += vDelta;
				}

				gc.setColor(getColor(selected, active, row, column));
				gc.drawString(text, x, y);
			}
		}
	}

	/**
	 * @param row The row.
	 * @param column The column.
	 * @return The data of this cell as a string that is prepared for display.
	 */
	protected String getPresentationText(Row row, Column column) {
		String text = getData(row, column, false);
		if (!mWrapped || row == null) {
			return text;
		}
		int width = column.getWidth();
		if (width == -1) {
			return text;
		}
		return TextDrawing.wrapToPixelWidth(getFont(row, column), null, text, width - (H_MARGIN_WIDTH + row.getOwner().getIndentWidth(row, column)));
	}

	public Cursor getCursor(MouseEvent event, Rectangle bounds, Row row, Column column) {
		return Cursor.getDefaultCursor();
	}

	/** @return The truncation policy. */
	public int getTruncationPolicy() {
		return SwingConstants.CENTER;
	}

	/**
	 * @param row The row.
	 * @param column The column.
	 * @param nullOK <code>true</code> if <code>null</code> may be returned.
	 * @return The data of this cell as a string.
	 */
	protected String getData(Row row, Column column, boolean nullOK) {
		if (row != null) {
			String text = row.getDataAsText(column);

			return text == null ? nullOK ? null : "" : text; //$NON-NLS-1$
		}
		return column.toString();
	}

	/**
	 * @param row The row.
	 * @param column The column.
	 * @return The font.
	 */
	public Font getFont(Row row, Column column) {
		return UIManager.getFont("TextField.font"); //$NON-NLS-1$
	}

	/** @return The horizontal alignment. */
	public int getHAlignment() {
		return mHAlignment;
	}

	/** @param alignment The horizontal alignment. */
	public void setHAlignment(int alignment) {
		mHAlignment = alignment;
	}

	/** @return The vertical alignment. */
	public int getVAlignment() {
		return SwingConstants.TOP;
	}

	public String getToolTipText(MouseEvent event, Rectangle bounds, Row row, Column column) {
		if (getPreferredWidth(row, column) - H_MARGIN > column.getWidth() - row.getOwner().getIndentWidth(row, column)) {
			return getData(row, column, true);
		}
		return null;
	}

	public boolean participatesInDynamicRowLayout() {
		return mWrapped;
	}

	public void mouseClicked(MouseEvent event, Rectangle bounds, Row row, Column column) {
		// Does nothing
	}
}
