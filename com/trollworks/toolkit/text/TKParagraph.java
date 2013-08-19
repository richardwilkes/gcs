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

package com.trollworks.toolkit.text;

import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKFont;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.Shape;
import java.awt.font.FontRenderContext;
import java.util.ArrayList;

/** Represents a paragraph of text within a document. */
public class TKParagraph {
	private static final int	INVALID_MARKER_DEVIATION	= 2;
	private StringBuilder		mData;
	private ArrayList<String>	mLines;
	private ArrayList<LineData>	mLineWidths;
	private int					mWrapWidth;
	private Font				mFont;
	private int					mAscent;
	private int					mDescent;
	private int					mLineHeight;
	private int					mAlignment;

	/** Creates a new, empty paragraph. */
	public TKParagraph() {
		mData = new StringBuilder();
		mAscent = -1;
	}

	/**
	 * Creates a new paragraph from the text. The text should not have embedded line-endings in it.
	 * Use <code>createParagraphs()</code> if you need to create paragraphs from a string that may
	 * have embedded line-endings.
	 * 
	 * @param text The text of the paragraph.
	 */
	public TKParagraph(String text) {
		mData = new StringBuilder(text);
		mAscent = -1;
	}

	/** @return The internal {@link StringBuilder} used to track the text. */
	StringBuilder getDataBuffer() {
		return mData;
	}

	/**
	 * Draw the paragraph. The font and foreground color should be applied to the graphics object
	 * before calling this method.
	 * 
	 * @param g2d The graphics object to use.
	 * @param bounds The drawing area.
	 * @param alignment The horizontal alignment for the text.
	 * @param drawCursor Pass in <code>true</code> to draw the cursor.
	 * @param drawSelection Pass in <code>true</code> to draw the selected area with a highlight.
	 * @param selectionColor The color to use for drawing selections.
	 * @param selectedTextColor The color to use for selected text.
	 * @param selectionStart The selection starting position.
	 * @param selectionEnd The selection ending position.
	 * @param drawInvalidMarker Whether or not to draw the invalid marker.
	 * @return The last vertical position used within the drawing area.
	 */
	public int draw(Graphics2D g2d, Rectangle bounds, int alignment, boolean drawCursor, boolean drawSelection, Color selectionColor, Color selectedTextColor, int selectionStart, int selectionEnd, boolean drawInvalidMarker) {
		Rectangle clip = g2d.getClipBounds();
		int cY = clip.y;
		int bY = bounds.y;
		int minY = Math.max(cY, bY);
		int maxY = Math.min(cY + clip.height, bY + bounds.height);
		int wrapWidth = bounds.width;
		int top = bY;
		int left = bounds.x;
		int charCount = 0;
		int lineCount;

		prepareLayouts(g2d.getFont(), alignment, wrapWidth);
		checkLazyLoad();
		lineCount = mLines.size();
		for (int i = 0; i < lineCount && top < maxY; i++) {
			String line = mLines.get(i);
			LineData data = mLineWidths.get(i);
			int count = line.length();

			if (top + mLineHeight > minY) {
				Color color = g2d.getColor();
				int lineWidth = data.mWidths[data.mWidths.length - 1];
				Rectangle selBounds = null;
				int x = left;
				int tmp;

				if (mAlignment > TKAlignment.LEFT) {
					tmp = wrapWidth - lineWidth;
					if (mAlignment == TKAlignment.CENTER) {
						tmp /= 2;
					}
					x += tmp;
				}

				if (selectionStart <= data.mEnd && selectionEnd >= data.mStart) {
					boolean cursorOnly = selectionStart == selectionEnd;

					if (drawCursor && cursorOnly && selectionStart >= data.mStart && selectionStart <= data.mEnd) {
						g2d.setColor(Color.black);
						tmp = x + data.mWidths[selectionStart - data.mStart];
						if (mAlignment == TKAlignment.RIGHT && selectionEnd == data.mEnd) {
							tmp--;
						}
						g2d.drawLine(tmp, top, tmp, top + mLineHeight);
					} else if (drawSelection && !cursorOnly) {
						tmp = data.mWidths[Math.max(selectionStart, data.mStart) - data.mStart];
						g2d.setColor(selectionColor);
						selBounds = new Rectangle(x + tmp, top, data.mWidths[Math.min(selectionEnd, data.mEnd) - data.mStart] - tmp, mLineHeight);
						g2d.fill(selBounds);
					}
				}

				if (drawInvalidMarker) {
					boolean down = true;
					int iEnd = lineWidth;
					int y1 = top + mAscent;
					int x1;
					int x2;
					int y2;

					g2d.setColor(TKColor.INVALID_MARKER);
					for (x1 = 0; x1 < iEnd; x1 += INVALID_MARKER_DEVIATION) {
						x2 = x1 + INVALID_MARKER_DEVIATION;
						y2 = down ? INVALID_MARKER_DEVIATION : 0;
						down = !down;
						g2d.drawLine(x1 + x, y1, x2 + x, y2);
						y1 = y2;
					}
				}

				g2d.setColor(color);
				g2d.drawString(line, x, top + mAscent);
				if (selBounds != null) {
					Shape savedClip = g2d.getClip();

					g2d.clip(selBounds);
					g2d.setColor(selectedTextColor);
					g2d.drawString(line, x, top + mAscent);
					g2d.setColor(color);
					g2d.setClip(savedClip);
				}
			}
			top += mLineHeight;
			charCount += count + 1;
		}
		return top;
	}

	/**
	 * @param characterPosition The character to position to get the bounds of.
	 * @return The bounding rectangle, relative to the paragraph, of the character at the specified
	 *         position within the text. An empty rectangle will be returned if this paragraph has
	 *         never had {@link #prepareLayouts(Font,int,int)} called either directly or indirectly,
	 *         such as through a call to
	 *         {@link #draw(Graphics2D,Rectangle,int,boolean,boolean,Color,Color,int,int,boolean)}.
	 */
	public Rectangle getCharacterBounds(int characterPosition) {
		checkLazyLoad();
		if (mLines != null) {
			int count = mLineWidths.size();

			for (int i = 0; i < count; i++) {
				LineData data = mLineWidths.get(i);

				if (characterPosition <= data.mEnd) {
					int left;

					if (mAlignment > TKAlignment.LEFT) {
						left = mWrapWidth - data.mWidths[data.mWidths.length - 1];
						if (mAlignment == TKAlignment.CENTER) {
							left /= 2;
						}
					} else {
						left = 0;
					}
					characterPosition -= data.mStart;
					return new Rectangle(left + data.mWidths[characterPosition], i * mLineHeight, characterPosition < data.mWidths.length - 1 ? data.mWidths[characterPosition + 1] - data.mWidths[characterPosition] : 1, mLineHeight);
				}
			}
		}
		return new Rectangle();
	}

	/**
	 * @return The height of this paragraph. 0 if this paragraph has never had
	 *         {@link #prepareLayouts(Font,int,int)} called either directly or indirectly, such as
	 *         through a call to
	 *         {@link #draw(Graphics2D,Rectangle,int,boolean,boolean,Color,Color,int,int,boolean)}.
	 */
	public int getHeight() {
		checkLazyLoad();
		return mLines != null ? mLineHeight * mLines.size() : mLineHeight;
	}

	/**
	 * @return The width of this paragraph. 0 if this paragraph has never had
	 *         {@link #prepareLayouts(Font,int,int)} called either directly or indirectly, such as
	 *         through a call to
	 *         {@link #draw(Graphics2D,Rectangle,int,boolean,boolean,Color,Color,int,int,boolean)}.
	 */
	public int getWidth() {
		checkLazyLoad();
		if (mLines != null) {
			int width = 0;

			for (LineData data : mLineWidths) {
				int lWidth = data.mWidths[data.mWidths.length - 1];

				if (lWidth > width) {
					width = lWidth;
				}
			}
			return width;
		}
		return 0;
	}

	/**
	 * @return The dimensions of this paragraph. 0 for width and height if this paragraph has never
	 *         had {@link #prepareLayouts(Font,int,int)} called either directly or indirectly, such
	 *         as through a call to
	 *         {@link #draw(Graphics2D,Rectangle,int,boolean,boolean,Color,Color,int,int,boolean)}.
	 */
	public Dimension getSize() {
		checkLazyLoad();
		if (mLines != null) {
			int width = 0;

			for (LineData data : mLineWidths) {
				int lWidth = data.mWidths[data.mWidths.length - 1];

				if (lWidth > width) {
					width = lWidth;
				}
			}
			return new Dimension(width, mLineHeight * mLines.size());
		}
		return new Dimension(0, mLineHeight);
	}

	/** @return The length of the paragraph without the trailing return. */
	public int getLength() {
		return mData.length();
	}

	/** @return The text of this paragraph without the trailing return. */
	public String getText() {
		return mData.toString();
	}

	/**
	 * Returns the position of the nearest character within the text at the specified location.
	 * 
	 * @param x The x-coordinate.
	 * @param y The y-coordinate.
	 * @return The text position nearest the coordinates.
	 */
	public int getTextPosition(int x, int y) {
		checkLazyLoad();
		if (mLines != null && mData.length() > 0) {
			int index = y / mLineHeight;
			LineData data;

			if (index < 0) {
				return 0;
			}
			if (index >= mLines.size()) {
				return getLength();
			}

			data = mLineWidths.get(index);

			if (mAlignment > TKAlignment.LEFT) {
				int left = mWrapWidth - data.mWidths[data.mWidths.length - 1];

				if (mAlignment == TKAlignment.CENTER) {
					left /= 2;
				}
				x -= left;
			}

			for (int i = 0; i < data.mWidths.length - 1; i++) {
				if (x < data.mWidths[i] + (data.mWidths[i + 1] - data.mWidths[i]) / 2) {
					return data.mStart + i;
				}
			}
			return data.mEnd;
		}
		return 0;
	}

	/**
	 * Prepares this paragraph's text layouts.
	 * 
	 * @param font The font to use.
	 * @param alignment The horizontal alignment for the text.
	 * @param wrapWidth The wrapping width to use.
	 */
	public void prepareLayouts(Font font, int alignment, int wrapWidth) {
		if (mLines == null || mWrapWidth != wrapWidth || alignment != mAlignment || !font.equals(mFont)) {
			FontRenderContext frc = TKFont.getRenderContext();
			int pos = 0;

			mAlignment = alignment;
			mWrapWidth = wrapWidth;
			mFont = font;
			mAscent = -1;
			mLines = wrapWidth > 0 ? TKTextDrawing.breakAtPixelWidth(font, frc, mData.toString(), wrapWidth) : new ArrayList<String>();
			mLineWidths = new ArrayList<LineData>();
			if (mLines.isEmpty()) {
				mLines.add(""); //$NON-NLS-1$
			}
			for (String line : mLines) {
				int length = line.length();
				LineData data = new LineData(length + 1);

				mLineWidths.add(data);
				data.mStart = pos;
				pos += length;
				data.mEnd = pos++;
				data.mWidths[0] = 0;
				for (int i = 1; i <= length; i++) {
					data.mWidths[i] = TKTextDrawing.getWidth(mFont, frc, line.substring(0, i));
				}
			}
		}
	}

	private void checkLazyLoad() {
		if (mAscent == -1) {
			FontMetrics fm = TKFont.getFontMetrics(mFont);

			mAscent = fm.getAscent();
			mDescent = fm.getDescent();
			mLineHeight = mAscent + mDescent;
		}
	}

	private class LineData {
		/** The width of the line for text up to the specified index. */
		public int[]	mWidths;
		/** The starting text position within the paragraph. */
		public int		mStart;
		/** The ending text position within the paragraph. */
		public int		mEnd;

		/** @param count The number of widths to store. */
		public LineData(int count) {
			mWidths = new int[count];
		}
	}
}
