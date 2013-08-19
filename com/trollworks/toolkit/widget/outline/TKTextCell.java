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

import com.trollworks.toolkit.text.TKNumericStringComparator;
import com.trollworks.toolkit.text.TKTextDrawing;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.TKKeyEventFilter;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;

import java.awt.Color;
import java.awt.Cursor;
import java.awt.Font;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.awt.font.FontRenderContext;
import java.awt.image.BufferedImage;
import java.util.StringTokenizer;

/** Represents text cells in an {@link TKOutline}. */
public class TKTextCell implements TKCell {
	/** The constant to use for comparing as text. */
	public static final int		COMPARE_AS_TEXT		= 0;
	/** The constant to use for comparing as an integer number. */
	public static final int		COMPARE_AS_INTEGER	= 1;
	/** The constant to use for comparing as a floating point number. */
	public static final int		COMPARE_AS_FLOAT	= 2;
	/** The standard horizontal margin. */
	public static final int		H_MARGIN			= 2;
	/** The standard horizontal margin width. */
	public static final int		H_MARGIN_WIDTH		= H_MARGIN * 2;
	private int					mHAlignment;
	private int					mVAlignment;
	private Color				mColor;
	private Color				mSelectedColor;
	private String				mFontKey;
	private int					mCompareType;
	private boolean				mEditable;
	private int					mTruncationPolicy;
	private TKKeyEventFilter	mKeyEventFilter;
	private boolean				mWrapped;

	/** Create a new text cell. */
	public TKTextCell() {
		this(TKAlignment.LEFT);
	}

	/**
	 * Create a new text cell.
	 * 
	 * @param editable Pass in <code>true</code> to make the cell editable.
	 */
	public TKTextCell(boolean editable) {
		this(null, null, null, TKAlignment.LEFT, TKAlignment.CENTER, COMPARE_AS_TEXT, editable, TKAlignment.CENTER);
	}

	/**
	 * Create a new text cell.
	 * 
	 * @param editable Pass in <code>true</code> to make the cell editable.
	 * @param wrapped Pass in <code>true</code> to enable wrapping.
	 */
	public TKTextCell(boolean editable, boolean wrapped) {
		this(null, null, null, TKAlignment.LEFT, TKAlignment.CENTER, COMPARE_AS_TEXT, editable, TKAlignment.CENTER, null, wrapped);
	}

	/**
	 * Create a new text cell.
	 * 
	 * @param editable Pass in <code>true</code> to make the cell editable.
	 * @param alignment The horizontal text alignment to use.
	 */
	public TKTextCell(boolean editable, int alignment) {
		this(null, null, null, alignment, TKAlignment.CENTER, COMPARE_AS_TEXT, editable, TKAlignment.CENTER);
	}

	/**
	 * Create a new text cell.
	 * 
	 * @param editable Pass in <code>true</code> to make the cell editable.
	 * @param alignment The horizontal text alignment to use.
	 * @param compareType The type of comparison to perform when asked.
	 */
	public TKTextCell(boolean editable, int alignment, int compareType) {
		this(null, null, null, alignment, TKAlignment.CENTER, compareType, editable, TKAlignment.CENTER);
	}

	/**
	 * Create a new text cell.
	 * 
	 * @param alignment The horizontal text alignment to use.
	 */
	public TKTextCell(int alignment) {
		this(null, alignment);
	}

	/**
	 * Create a new text cell.
	 * 
	 * @param font The font to use.
	 * @param alignment The horizontal text alignment to use.
	 */
	public TKTextCell(String font, int alignment) {
		this(font, null, null, alignment, TKAlignment.CENTER, COMPARE_AS_TEXT, true, TKAlignment.CENTER);
	}

	/**
	 * Create a new text cell.
	 * 
	 * @param alignment The horizontal text alignment to use.
	 * @param compareType The type of comparison to perform when asked.
	 */
	public TKTextCell(int alignment, int compareType) {
		this(null, null, null, alignment, TKAlignment.CENTER, compareType, true, TKAlignment.CENTER);
	}

	/**
	 * Create a new text cell.
	 * 
	 * @param font The font to use.
	 * @param alignment The horizontal text alignment to use.
	 * @param compareType The type of comparison to perform when asked.
	 */
	public TKTextCell(String font, int alignment, int compareType) {
		this(font, null, null, alignment, TKAlignment.CENTER, compareType, true, TKAlignment.CENTER);
	}

	/**
	 * Create a new text cell.
	 * 
	 * @param font The font to use.
	 * @param color The foreground color to use.
	 * @param selectedColor The foreground color to use when selected.
	 * @param hAlignment The horizontal text alignment to use.
	 * @param vAlignment The vertical text alignment to use.
	 * @param compareType The type of comparison to perform when asked.
	 * @param editable Pass in <code>true</code> to make the cell editable.
	 * @param truncationPolicy The truncation policy to use. One of {@link TKAlignment#LEFT},
	 *            {@link TKAlignment#CENTER}, or {@link TKAlignment#RIGHT}.
	 */
	public TKTextCell(String font, Color color, Color selectedColor, int hAlignment, int vAlignment, int compareType, boolean editable, int truncationPolicy) {
		this(font, color, selectedColor, hAlignment, vAlignment, compareType, editable, truncationPolicy, null, false);
	}

	/**
	 * Create a new text cell.
	 * 
	 * @param font The font key to use.
	 * @param color The foreground color to use.
	 * @param selectedColor The foreground color to use when selected.
	 * @param hAlignment The horizontal text alignment to use.
	 * @param vAlignment The vertical text alignment to use.
	 * @param compareType The type of comparison to perform when asked.
	 * @param editable Pass in <code>true</code> to make the cell editable.
	 * @param truncationPolicy The truncation policy to use. One of {@link TKAlignment#LEFT},
	 *            {@link TKAlignment#CENTER}, or {@link TKAlignment#RIGHT}.
	 * @param keyFilter The key event filter to use.
	 * @param wrapped Pass in <code>true</code> to enable wrapping.
	 */
	public TKTextCell(String font, Color color, Color selectedColor, int hAlignment, int vAlignment, int compareType, boolean editable, int truncationPolicy, TKKeyEventFilter keyFilter, boolean wrapped) {
		mFontKey = font;
		mColor = color;
		mSelectedColor = selectedColor;
		mHAlignment = hAlignment;
		mVAlignment = vAlignment;
		mCompareType = compareType;
		mEditable = editable;
		mTruncationPolicy = truncationPolicy;
		mKeyEventFilter = keyFilter;
		mWrapped = wrapped;
	}

	/** @return The compare type for this cell. */
	public int getCompareType() {
		return mCompareType;
	}

	/** @param compareType The compare type for this cell. */
	public void setCompareType(int compareType) {
		mCompareType = compareType;
	}

	public int compare(TKColumn column, TKRow one, TKRow two) {
		switch (mCompareType) {
			case COMPARE_AS_INTEGER:
				long firstL = TKNumberUtils.getLong(one.getData(column).toString(), 0);
				long secondL = TKNumberUtils.getLong(two.getData(column).toString(), 0);

				if (firstL < secondL) {
					return -1;
				}
				if (firstL > secondL) {
					return 1;
				}
				return 0;
			case COMPARE_AS_FLOAT:
				double first = TKNumberUtils.getDouble(one.getData(column).toString(), 0.0);
				double second = TKNumberUtils.getDouble(two.getData(column).toString(), 0.0);

				if (first < second) {
					return -1;
				}
				if (first > second) {
					return 1;
				}
				return 0;
			default:
				return TKNumericStringComparator.caselessCompareStrings(one.getDataAsText(column), two.getDataAsText(column));
		}
	}

	/**
	 * @param selected Whether or not the selected version of the color is needed.
	 * @param active Whether or not the active version of the color is needed.
	 * @param row The row.
	 * @param column The column.
	 * @return The foreground color.
	 */
	public Color getColor(boolean selected, @SuppressWarnings("unused") boolean active, @SuppressWarnings("unused") TKRow row, @SuppressWarnings("unused") TKColumn column) {
		return selected ? getSelectedColor() : getColor();
	}

	public int getPreferredWidth(TKRow row, TKColumn column) {
		int width = TKTextDrawing.getPreferredSize(TKFont.lookup(getFontKey(row, column)), null, getPresentationText(row, column)).width;
		BufferedImage icon = row == null ? column.getIcon() : row.getIcon(column);

		if (icon != null) {
			width += icon.getWidth() + H_MARGIN;
		}

		return H_MARGIN_WIDTH + width;
	}

	public int getPreferredHeight(TKRow row, TKColumn column) {
		Font font = TKFont.lookup(getFontKey(row, column));
		int minHeight = TKTextDrawing.getPreferredSize(font, null, "Mg").height; //$NON-NLS-1$
		int height = TKTextDrawing.getPreferredSize(font, null, getPresentationText(row, column)).height;
		BufferedImage icon = row == null ? column.getIcon() : row.getIcon(column);

		if (icon != null) {
			int iconHeight = icon.getHeight();

			if (height < iconHeight) {
				height = iconHeight;
			}
		}

		return minHeight > height ? minHeight : height;
	}

	public void drawCell(Graphics2D g2d, Rectangle bounds, TKRow row, TKColumn column, boolean selected, boolean active) {
		Font font = TKFont.lookup(getFontKey(row, column));
		FontRenderContext frc = g2d.getFontRenderContext();
		int ascent = g2d.getFontMetrics(font).getAscent();
		StringTokenizer tokenizer = new StringTokenizer(getPresentationText(row, column), "\n", true); //$NON-NLS-1$
		int totalHeight = getPreferredHeight(row, column);
		int lineHeight = TKTextDrawing.getPreferredSize(font, frc, "Mg").height; //$NON-NLS-1$
		int lineCount = 0;
		BufferedImage icon = row == null ? column.getIcon() : row.getIcon(column);
		int left = icon == null ? 0 : icon.getWidth() + H_MARGIN;
		int cellWidth = bounds.width - (left + H_MARGIN_WIDTH);
		int vAlignment = getVAlignment();
		int hAlignment = getHAlignment();

		left += bounds.x + H_MARGIN;

		if (icon != null) {
			int iy = bounds.y;

			if (vAlignment != TKAlignment.TOP) {
				int ivDelta = bounds.height - icon.getHeight();

				if (vAlignment == TKAlignment.CENTER) {
					ivDelta /= 2;
				}
				iy += ivDelta;
			}
			g2d.drawImage(icon, bounds.x + H_MARGIN, iy, null);
		}

		g2d.setFont(font);
		while (tokenizer.hasMoreTokens()) {
			String token = tokenizer.nextToken();

			if (token.equals("\n")) { //$NON-NLS-1$
				lineCount++;
			} else {
				String text = TKTextDrawing.truncateIfNecessary(font, frc, token, cellWidth, getTruncationPolicy());
				int x = left;
				int y = bounds.y + ascent + lineHeight * lineCount;

				if (hAlignment != TKAlignment.LEFT) {
					int hDelta = cellWidth - TKTextDrawing.getWidth(font, frc, text);

					if (hAlignment == TKAlignment.CENTER) {
						hDelta /= 2;
					}
					x += hDelta;
				}

				if (vAlignment != TKAlignment.TOP) {
					float vDelta = bounds.height - totalHeight;

					if (vAlignment == TKAlignment.CENTER) {
						vDelta /= 2;
					}
					y += vDelta;
				}

				g2d.setColor(getColor(selected, active, row, column));
				g2d.drawString(text, x, y);
			}
		}
	}

	/**
	 * @param row The row.
	 * @param column The column.
	 * @return The data of this cell as a string that is prepared for display.
	 */
	protected String getPresentationText(TKRow row, TKColumn column) {
		String text = getData(row, column, false);

		if (!mWrapped || row == null) {
			return text;
		}

		int width = column.getWidth();

		if (width == -1) {
			return text;
		}
		return TKTextDrawing.wrapToPixelWidth(TKFont.lookup(getFontKey(row, column)), null, text, width - (H_MARGIN_WIDTH + row.getOwner().getIndentWidth(row, column)));
	}

	public TKPanel getEditor(Color rowBackground, TKRow row, TKColumn column, boolean mouseInitiated) {
		TKTextField editor = new TKTextField(0, getHAlignment());
		TKKeyEventFilter filter = getKeyEventFilter();

		editor.setSingleLineOnly(false);
		editor.setFontKey(getFontKey(row, column));
		editor.setForeground(getColor());
		editor.setBackground(rowBackground);
		editor.setBorder(new TKEmptyBorder(0, H_MARGIN, 0, H_MARGIN));
		editor.setText(getData(row, column, false));
		editor.setActionCommand(TKOutline.CMD_UPDATE_FROM_EDITOR);
		if (filter != null) {
			editor.setKeyEventFilter(filter);
		}
		if (mouseInitiated) {
			editor.setSelection(Integer.MAX_VALUE, Integer.MAX_VALUE);
			editor.disableNextSelectAllForFocus();
		}
		return editor;
	}

	public boolean isRowDragHandle(TKRow row, TKColumn column) {
		return !isEditable(row, column, true);
	}

	public boolean isEditable(TKRow row, TKColumn column, boolean viaSingleClick) {
		return isEditable();
	}

	/** @return Whether this text cell is generally editable. */
	public boolean isEditable() {
		return mEditable;
	}

	/** @param editable Whether this cell is editable or not. */
	public void setEditable(boolean editable) {
		mEditable = editable;
	}

	public Object getEditedObject(TKPanel editor) {
		return ((TKTextField) editor).getText();
	}

	public Object stopEditing(TKPanel editor) {
		return getEditedObject(editor);
	}

	public Cursor getCursor(MouseEvent event, Rectangle bounds, TKRow row, TKColumn column) {
		return isEditable(row, column, true) ? Cursor.getPredefinedCursor(Cursor.TEXT_CURSOR) : Cursor.getDefaultCursor();
	}

	/** @return The truncation policy. */
	public int getTruncationPolicy() {
		return mTruncationPolicy;
	}

	/**
	 * @param truncationPolicy The truncation policy. One of {@link TKAlignment#LEFT},
	 *            {@link TKAlignment#CENTER}, or {@link TKAlignment#RIGHT}.
	 */
	public void setTruncationPolicy(int truncationPolicy) {
		mTruncationPolicy = truncationPolicy;
	}

	/**
	 * @param row The row.
	 * @param column The column.
	 * @param nullOK <code>true</code> if <code>null</code> may be returned.
	 * @return The data of this cell as a string.
	 */
	protected String getData(TKRow row, TKColumn column, boolean nullOK) {
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
	public String getFontKey(@SuppressWarnings("unused") TKRow row, @SuppressWarnings("unused") TKColumn column) {
		return getFontKey();
	}

	/** @return The font. */
	public String getFontKey() {
		if (mFontKey == null) {
			mFontKey = TKFont.TEXT_FONT_KEY;
		}
		return mFontKey;
	}

	/** @param font The font to set. */
	public void setFontKey(String font) {
		mFontKey = font;
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
		return mVAlignment;
	}

	/** @param alignment The vertical alignment. */
	public void setVAlignment(int alignment) {
		mVAlignment = alignment;
	}

	/** @return The non-selected font color for this cell. */
	public Color getColor() {
		if (mColor == null) {
			mColor = TKColor.TEXT;
		}
		return mColor;
	}

	/**
	 * Sets the non-selected font color for this cell.
	 * 
	 * @param color The non-selected font color.
	 */
	public void setColor(Color color) {
		mColor = color;
	}

	/** @return The selected font color for this cell. */
	public Color getSelectedColor() {
		if (mSelectedColor == null) {
			mSelectedColor = TKColor.HIGHLIGHTED_TEXT;
		}
		return mSelectedColor;
	}

	/**
	 * Sets the selected font color for this cell.
	 * 
	 * @param selectedColor The selected font color.
	 */
	public void setSelectedColor(Color selectedColor) {
		mSelectedColor = selectedColor;
	}

	/** @return The key event filter. */
	public TKKeyEventFilter getKeyEventFilter() {
		return mKeyEventFilter;
	}

	public String getToolTipText(MouseEvent event, Rectangle bounds, TKRow row, TKColumn column) {
		if (getPreferredWidth(row, column) - H_MARGIN > column.getWidth() - row.getOwner().getIndentWidth(row, column)) {
			return getData(row, column, true);
		}
		return null;
	}

	public boolean participatesInDynamicRowLayout() {
		return mWrapped;
	}

	public void mouseClicked(MouseEvent event, Rectangle bounds, TKRow row, TKColumn column) {
		// Does nothing
	}
}
