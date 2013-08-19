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

package com.trollworks.gcs.utility.text;

import com.trollworks.gcs.utility.Fonts;

import java.awt.Component;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.font.FontRenderContext;
import java.util.ArrayList;
import java.util.StringTokenizer;

import javax.swing.SwingConstants;

/** General text drawing utilities. */
public class TextDrawing {
	private static final String	SPACE		= " ";		//$NON-NLS-1$
	private static final String	NEWLINE		= "\n";	//$NON-NLS-1$
	private static final String	WRAPPABLE	= " \t/\\"; //$NON-NLS-1$
	private static final char	ELLIPSIS	= '\u2026';

	/**
	 * Draws the text. Embedded return characters may be present.
	 * 
	 * @param gc The graphics context.
	 * @param bounds The bounding rectangle to draw the text within.
	 * @param text The text to draw.
	 * @param hAlign The horizontal alignment to use, from {@link Component}.
	 * @param vAlign The vertical alignment to use, from {@link Component}.
	 * @return The bottom of the drawn text.
	 */
	public static final int draw(Graphics gc, Rectangle bounds, String text, int hAlign, int vAlign) {
		int y = bounds.y;

		if (text.length() > 0) {
			ArrayList<String> list = new ArrayList<String>();
			Font font = gc.getFont();
			FontRenderContext frc = ((Graphics2D) gc).getFontRenderContext();
			FontMetrics fm = gc.getFontMetrics();
			int ascent = fm.getAscent();
			int descent = fm.getDescent();
			// Don't use fm.getHeight(), as the PC adds too much dead space
			int fHeight = ascent + descent;
			StringTokenizer tokenizer = new StringTokenizer(text, " \n", true); //$NON-NLS-1$
			StringBuilder buffer = new StringBuilder(text.length());
			int textHeight = 0;
			int width;

			while (tokenizer.hasMoreTokens()) {
				String token = tokenizer.nextToken();

				if (token.equals(NEWLINE)) {
					text = buffer.toString();
					textHeight += fHeight;
					list.add(text);
					buffer.setLength(0);
				} else {
					width = getSimpleWidth(font, frc, buffer.toString() + token);
					if (width > bounds.width && buffer.length() > 0) {
						text = buffer.toString();
						textHeight += fHeight;
						list.add(text);
						buffer.setLength(0);
					}
					buffer.append(token);
				}
			}

			if (buffer.length() > 0) {
				text = buffer.toString();
				textHeight += fHeight;
				list.add(text);
			}

			if (vAlign == SwingConstants.CENTER) {
				y = bounds.y + (bounds.height - (textHeight - descent / 2)) / 2;
			} else if (vAlign == SwingConstants.BOTTOM) {
				y = bounds.y + bounds.height - textHeight;
			}

			for (String piece : list) {
				int x = bounds.x;

				if (hAlign == SwingConstants.CENTER) {
					x = x + (bounds.width - getSimpleWidth(font, frc, piece)) / 2;
				} else if (hAlign == SwingConstants.RIGHT) {
					x = x + bounds.width - (1 + getSimpleWidth(font, frc, piece));
				}
				gc.drawString(piece, x, y + ascent);
				y += fHeight;
			}
		}
		return y;
	}

	/**
	 * Embedded return characters may be present.
	 * 
	 * @param font The font the text will be in.
	 * @param frc The font render context.
	 * @param text The text to calculate a size for.
	 * @return The preferred size of the text in the specified font.
	 */
	public static final Dimension getPreferredSize(Font font, FontRenderContext frc, String text) {
		FontMetrics fm = Fonts.getFontMetrics(font);
		// Don't use fm.getHeight(), as the PC adds too much dead space
		int fHeight = fm.getAscent() + fm.getDescent();
		StringTokenizer tokenizer = new StringTokenizer(text, NEWLINE, true);
		boolean veryFirst = true;
		boolean first = true;
		int width = 0;
		int height = 0;

		if (frc == null) {
			frc = Fonts.getRenderContext();
		}
		while (tokenizer.hasMoreTokens()) {
			String token = tokenizer.nextToken();
			int bWidth;

			if (token.equals(NEWLINE)) {
				if (first && !veryFirst) {
					first = false;
					continue;
				}
				token = SPACE;
			} else {
				first = true;
			}
			veryFirst = false;
			height += fHeight;
			bWidth = getSimpleWidth(font, frc, token);
			if (width < bWidth) {
				width = bWidth;
			}
		}
		return new Dimension(width, height);
	}

	/**
	 * @param font The font the text will be in.
	 * @param frc The font render context.
	 * @param text The text to calculate a size for.
	 * @return The width of the text in the specified font.
	 */
	public static final int getWidth(Font font, FontRenderContext frc, String text) {
		StringTokenizer tokenizer = new StringTokenizer(text, NEWLINE, true);
		boolean veryFirst = true;
		boolean first = true;
		int width = 0;

		if (frc == null) {
			frc = Fonts.getRenderContext();
		}
		while (tokenizer.hasMoreTokens()) {
			String token = tokenizer.nextToken();
			int bWidth;

			if (token.equals(NEWLINE)) {
				if (first && !veryFirst) {
					first = false;
					continue;
				}
				token = SPACE;
			} else {
				first = true;
			}
			veryFirst = false;
			bWidth = getSimpleWidth(font, frc, token);
			if (width < bWidth) {
				width = bWidth;
			}
		}
		return width;
	}

	private static final int getSimpleWidth(Font font, FontRenderContext frc, String text) {
		return (int) (font.getStringBounds(text, frc).getWidth() + 0.5);
	}

	/**
	 * If the text doesn't fit in the specified width, it will be shortened and an ellipse ("...")
	 * will be added. This method does not work properly on text with embedded line endings.
	 * 
	 * @param font The font to use.
	 * @param frc The font render context.
	 * @param text The text to work on.
	 * @param width The maximum pixel width.
	 * @param truncationPolicy One of {@link SwingConstants#LEFT}, {@link SwingConstants#CENTER},
	 *            or {@link SwingConstants#RIGHT}.
	 * @return The adjusted text.
	 */
	public static final String truncateIfNecessary(Font font, FontRenderContext frc, String text, int width, int truncationPolicy) {
		if (frc == null) {
			frc = Fonts.getRenderContext();
		}
		if (getSimpleWidth(font, frc, text) > width) {
			StringBuilder buffer = new StringBuilder(text);
			int max = buffer.length();

			if (truncationPolicy == SwingConstants.LEFT) {
				buffer.insert(0, ELLIPSIS);
				while (max-- > 0 && getSimpleWidth(font, frc, buffer.toString()) > width) {
					buffer.deleteCharAt(1);
				}
			} else if (truncationPolicy == SwingConstants.CENTER) {
				int left = max / 2;
				int right = left + 1;
				boolean leftSide = false;

				buffer.insert(left--, ELLIPSIS);
				while (max-- > 0 && getSimpleWidth(font, frc, buffer.toString()) > width) {
					if (leftSide) {
						buffer.deleteCharAt(left--);
						if (--right < max + 1) {
							leftSide = false;
						}
					} else {
						buffer.deleteCharAt(right);
						if (left >= 0) {
							leftSide = true;
						}
					}
				}
			} else if (truncationPolicy == SwingConstants.RIGHT) {
				buffer.append(ELLIPSIS);
				while (max-- > 0 && getSimpleWidth(font, frc, buffer.toString()) > width) {
					buffer.deleteCharAt(max);
				}
			}
			text = buffer.toString();
		}
		return text;
	}

	/**
	 * @param font The font to use.
	 * @param frc The font render context.
	 * @param text The text to wrap.
	 * @param width The maximum pixel width to allow.
	 * @return A new, wrapped version of the text.
	 */
	public static String wrapToPixelWidth(Font font, FontRenderContext frc, String text, int width) {
		StringBuilder buffer = new StringBuilder(text.length() * 2);
		StringBuilder lineBuffer = new StringBuilder(text.length());
		StringTokenizer tokenizer = new StringTokenizer(text + NEWLINE, NEWLINE, true);

		if (frc == null) {
			frc = Fonts.getRenderContext();
		}
		while (tokenizer.hasMoreTokens()) {
			String token = tokenizer.nextToken();

			if (token.equals(NEWLINE)) {
				buffer.append(token);
			} else {
				StringTokenizer tokenizer2 = new StringTokenizer(token, WRAPPABLE, true);
				int lineWidth = 0;
				boolean wrapped = false;

				lineBuffer.setLength(0);
				while (tokenizer2.hasMoreTokens()) {
					String token2 = tokenizer2.nextToken();
					int tokenWidth = getSimpleWidth(font, frc, token2);

					if (wrapped && lineWidth == 0 && token2.equals(SPACE)) {
						continue;
					}
					if (lineWidth == 0 || lineWidth + tokenWidth <= width) {
						lineBuffer.append(token2);
						lineWidth += tokenWidth;
					} else {
						buffer.append(lineBuffer);
						buffer.append(NEWLINE);
						wrapped = true;
						lineBuffer.setLength(0);
						if (!token2.equals(SPACE)) {
							lineBuffer.append(token2);
							lineWidth = tokenWidth;
						} else {
							lineWidth = 0;
						}
					}
				}
				if (lineWidth > 0) {
					buffer.append(lineBuffer);
				}
			}
		}
		buffer.setLength(buffer.length() - 1);
		return buffer.toString();
	}

	/**
	 * @param font The font to use.
	 * @param frc The font render context.
	 * @param text The text to wrap.
	 * @param width The maximum pixel width to allow.
	 * @return An array of strings that have been wrapped at the specified width and trailing line
	 *         feeds have been removed.
	 */
	public static final ArrayList<String> breakAtPixelWidth(Font font, FontRenderContext frc, String text, int width) {
		ArrayList<String> data = new ArrayList<String>();
		StringBuilder buffer = new StringBuilder(text.length());
		StringTokenizer tokenizer = new StringTokenizer(text, " \t\n", true); //$NON-NLS-1$
		boolean wrapped = false;

		if (frc == null) {
			frc = Fonts.getRenderContext();
		}

		while (tokenizer.hasMoreTokens()) {
			String token = tokenizer.nextToken();

			if (NEWLINE.equals(token)) {
				data.add(buffer.toString());
				buffer.setLength(0);
				wrapped = false;
			} else {
				int marker = buffer.length();

				if (wrapped && SPACE.equals(token) && buffer.length() == 0) {
					continue;
				}
				buffer.append(token);
				if (getSimpleWidth(font, frc, buffer.toString()) > width) {
					buffer.setLength(marker);
					data.add(buffer.toString());
					buffer.setLength(0);
					wrapped = true;
					if (!SPACE.equals(token)) {
						buffer.append(token);
					}
				}
			}
		}
		if (buffer.length() > 0) {
			data.add(buffer.toString());
		}
		return data;
	}
}
