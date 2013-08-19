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

package com.trollworks.toolkit.widget;

import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.text.TKTextDrawing;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKDragUtil;
import com.trollworks.toolkit.utility.TKFont;

import java.awt.Color;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.datatransfer.Transferable;
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DragGestureEvent;
import java.awt.dnd.DragGestureListener;
import java.awt.dnd.DragGestureRecognizer;
import java.awt.dnd.DragSource;
import java.awt.font.FontRenderContext;
import java.awt.geom.AffineTransform;
import java.awt.image.BufferedImage;
import java.util.ArrayList;
import java.util.StringTokenizer;

/** A label capable of displaying text, an image, or both. */
public class TKLabel extends TKPanel implements DragGestureListener {
	private static final int		IMAGE_TEXT_GAP	= 4;
	private Transferable			mDragData;
	private String					mText;
	private BufferedImage			mImage;
	private BufferedImage			mDisabledImage;
	private int						mHorizontalAlignment;
	private int						mVerticalAlignment;
	private Color					mShadowColor;
	private boolean					mDisabledImageSet;
	private Color					mDisabledForeground;
	private boolean					mWrapText;
	private int						mTruncationPolicy;
	private DragGestureRecognizer	mDragGestureRecognizer;
	private boolean					mVerticalDisplay;

	/**
	 * Creates a label with no image and no text. The label is left-aligned and centered vertically
	 * in its display area.
	 */
	public TKLabel() {
		this(null, null, TKAlignment.LEFT, false, (String) null);
	}

	/**
	 * Creates a label with no image and no text. The label is centered vertically in its display
	 * area.
	 * 
	 * @param alignment The horizontal alignment to use.
	 */
	public TKLabel(int alignment) {
		this(null, null, alignment, false, (String) null);
	}

	/**
	 * Creates a label with no image and no text. The label is centered vertically in its display
	 * area.
	 * 
	 * @param alignment The horizontal alignment to use.
	 * @param font The dynamic font to use.
	 */
	public TKLabel(int alignment, String font) {
		this(null, null, alignment, false, font);
	}

	/**
	 * Creates a label with no image and the specified text. The label is left-aligned and centered
	 * vertically in its display area.
	 * 
	 * @param text The text to be displayed.
	 */
	public TKLabel(String text) {
		this(text, null, TKAlignment.LEFT, false, (String) null);
	}

	/**
	 * Creates a label with no image and the specified text. The label is left-aligned and centered
	 * vertically in its display area.
	 * 
	 * @param text The text to be displayed.
	 * @param wrapped Pass in <code>true</code> to have the text wrap.
	 */
	public TKLabel(String text, boolean wrapped) {
		this(text, null, TKAlignment.LEFT, wrapped, (String) null);
	}

	/**
	 * Creates a label with no image and the specified text. The label is centered vertically in its
	 * display area.
	 * 
	 * @param text The text to be displayed.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKLabel(String text, int alignment) {
		this(text, null, alignment, false, (String) null);
	}

	/**
	 * Creates a label with no image and the specified text. The label is centered vertically in its
	 * display area.
	 * 
	 * @param text The text to be displayed.
	 * @param alignment The horizontal alignment to use.
	 * @param wrapped Pass in <code>true</code> to have the text wrap.
	 */
	public TKLabel(String text, int alignment, boolean wrapped) {
		this(text, null, alignment, wrapped, (String) null);
	}

	/**
	 * Creates a label with no image and the specified text. The label is left-aligned and centered
	 * vertically in its display area.
	 * 
	 * @param text The text to be displayed.
	 * @param font The dynamic font to use.
	 */
	public TKLabel(String text, String font) {
		this(text, null, TKAlignment.LEFT, false, font);
	}

	/**
	 * Creates a label with no image and the specified text. The label is left-aligned and centered
	 * vertically in its display area.
	 * 
	 * @param text The text to be displayed.
	 * @param font The dynamic font to use.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKLabel(String text, String font, int alignment) {
		this(text, null, alignment, false, font);
	}

	/**
	 * Creates a label with no image and the specified text. The label is left-aligned and centered
	 * vertically in its display area.
	 * 
	 * @param text The text to be displayed.
	 * @param font The dynamic font to use.
	 * @param alignment The horizontal alignment to use.
	 * @param wrapped Pass in <code>true</code> to have the text wrap.
	 */
	public TKLabel(String text, String font, int alignment, boolean wrapped) {
		this(text, null, alignment, wrapped, font);
	}

	/**
	 * Creates a label with the specified image and no text. The label is centered in its display
	 * area.
	 * 
	 * @param image The image to be displayed.
	 */
	public TKLabel(BufferedImage image) {
		this(null, image, TKAlignment.CENTER, false, (String) null);
	}

	/**
	 * Creates a label with the specified image and no text. The label is centered vertically in its
	 * display area.
	 * 
	 * @param image The image to be displayed.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKLabel(BufferedImage image, int alignment) {
		this(null, image, alignment, false, (String) null);
	}

	/**
	 * Creates a label with the specified image and text. The label is centered in its display area
	 * and the text is on the right side of the image.
	 * 
	 * @param text The text to be displayed.
	 * @param image The image to be displayed.
	 */
	public TKLabel(String text, BufferedImage image) {
		this(text, image, TKAlignment.CENTER, false, (String) null);
	}

	/**
	 * Creates a label with the specified image and text. The label is centered vertically in its
	 * display area and the text is on the right side of the image.
	 * 
	 * @param text The text to be displayed.
	 * @param image The image to be displayed.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKLabel(String text, BufferedImage image, int alignment) {
		this(text, image, alignment, false, (String) null);
	}

	/**
	 * Creates a label with the specified image and text. The label is centered vertically in its
	 * display area and the text is on the right side of the image.
	 * 
	 * @param text The text to be displayed.
	 * @param image The image to be displayed.
	 * @param alignment The horizontal alignment to use.
	 * @param wrapped Pass in <code>true</code> to have the text wrap.
	 * @param font The dynamic font to use.
	 */
	public TKLabel(String text, BufferedImage image, int alignment, boolean wrapped, String font) {
		super();
		setFontKey(font == null ? TKFont.TEXT_FONT_KEY : font);
		setText(text);
		setImage(image);
		setHorizontalAlignment(alignment);
		setVerticalAlignment(TKAlignment.CENTER);
		setWrapText(wrapped);
		setCursor(Cursor.getDefaultCursor());
		setTruncationPolicy(TKAlignment.RIGHT);
	}

	/** @return Whether this label will display vertically or not. */
	public boolean isVerticalDisplay() {
		return mVerticalDisplay;
	}

	/** @param enabled Whether this label will display vertically or not. */
	public void setVerticalDisplay(boolean enabled) {
		mVerticalDisplay = enabled;
	}

	/**
	 * @param truncation The truncation policy. One of {@link TKAlignment#LEFT},
	 *            {@link TKAlignment#CENTER}, or {@link TKAlignment#RIGHT}.
	 */
	public void setTruncationPolicy(int truncation) {
		mTruncationPolicy = truncation;
	}

	/** @return The truncation policy. */
	public int getTruncationPolicy() {
		return mTruncationPolicy;
	}

	/** @return The current image being displayed, if any. */
	public BufferedImage getCurrentImage() {
		return isEnabled() ? getImage() : getDisabledImage();
	}

	/** @return The disabled foreground color. */
	public Color getDisabledForeground() {
		return mDisabledForeground == null ? getForeground() : mDisabledForeground;
	}

	/** @param color The disabled foreground color. */
	public void setDisabledForeground(Color color) {
		mDisabledForeground = color;
		repaint();
	}

	/**
	 * @return The disabled image the label displays. If it has not been previously set, a disabled
	 *         version of the image is computed.
	 */
	public BufferedImage getDisabledImage() {
		if (!mDisabledImageSet && mDisabledImage == null && mImage != null) {
			mDisabledImage = TKImage.createDisabledImage(mImage);
		}
		return mDisabledImage;
	}

	/** @param disabledImage The disabled image the label will display. */
	public void setDisabledImage(BufferedImage disabledImage) {
		BufferedImage oldValue = mDisabledImage;

		mDisabledImage = disabledImage;
		mDisabledImageSet = mDisabledImage != null;
		if (mDisabledImage != oldValue) {
			if (mDisabledImage == null || oldValue == null || mDisabledImage.getWidth() != oldValue.getWidth() || mDisabledImage.getHeight() != oldValue.getHeight()) {
				revalidate();
			}
			repaint();
		}
	}

	/** @return The horizontal alignment of the label's contents. */
	public int getHorizontalAlignment() {
		return mHorizontalAlignment;
	}

	/** @param alignment The horizontal alignment of the label's contents. */
	public void setHorizontalAlignment(int alignment) {
		if (alignment != mHorizontalAlignment) {
			mHorizontalAlignment = alignment;
			repaint();
		}
	}

	/** @return The image the label displays. */
	public BufferedImage getImage() {
		return mImage;
	}

	/** @param image The image the label will display. */
	public void setImage(BufferedImage image) {
		BufferedImage oldValue = mImage;

		mImage = image;
		if (mImage != oldValue) {
			if (!mDisabledImageSet) {
				mDisabledImage = null;
			}

			if (mImage == null || oldValue == null || mImage.getWidth() != oldValue.getWidth() || mImage.getHeight() != oldValue.getHeight()) {
				revalidate();
			}
			repaint();
		}
	}

	@Override protected Dimension getMaximumSizeSelf() {
		return getPreferredSize();
	}

	@Override protected Dimension getMinimumSizeSelf() {
		return getPreferredSize();
	}

	@Override protected Dimension getPreferredSizeSelf() {
		Insets insets = getInsets();
		BufferedImage image = getCurrentImage();
		String text = mText;
		Dimension size = new Dimension();

		if (image != null) {
			size.width = image.getWidth();
			size.height = image.getHeight();
		}

		if (image == null || text.length() > 0) {
			Dimension textSize;

			if (!mWrapText) {
				int where = text.indexOf('\n');

				if (where != -1) {
					text = text.substring(0, where);
				}
			}

			textSize = TKTextDrawing.getPreferredSize(getFont(), null, text.length() == 0 ? "M" : text); //$NON-NLS-1$
			if (mVerticalDisplay) {
				int tmp = textSize.width;

				textSize.width = textSize.height;
				textSize.height = tmp;
			}
			size.width += textSize.width;
			if (image != null) {
				size.width += IMAGE_TEXT_GAP;
			}
			if (size.height < textSize.height) {
				size.height = textSize.height;
			}
			size.width++; // Account for first pixel clipping
		}

		size.width += insets.left + insets.right;
		size.height += insets.top + insets.bottom;
		return size;
	}

	/** @return The text the label displays. */
	public String getText() {
		return mText;
	}

	/** @param text The text the label will display. */
	public void setText(String text) {
		String oldValue = mText;

		mText = text == null ? "" : text; //$NON-NLS-1$
		if (!mText.equals(oldValue)) {
			revalidate();
			repaint();
		}
	}

	/** @return The vertical alignment of the label's contents. */
	public int getVerticalAlignment() {
		return mVerticalAlignment;
	}

	/** @param alignment The vertical alignment of the label's contents. */
	public void setVerticalAlignment(int alignment) {
		if (alignment != mVerticalAlignment) {
			mVerticalAlignment = alignment;
			repaint();
		}
	}

	/** @return Whether the text in this label should be wrapped. */
	public boolean getWrapText() {
		return mWrapText;
	}

	/** @param wrap Whether the text in this label should be wrapped. */
	public void setWrapText(boolean wrap) {
		if (wrap != mWrapText) {
			mWrapText = wrap;
			revalidate();
		}
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		drawLabel(g2d, clips);
	}

	/**
	 * Calculates the amount of horizontal space reserved for a text title in this label.
	 * 
	 * @return The width available for a title string
	 */
	protected int getAvailableTextWidth() {
		Rectangle bounds = getLocalInsetBounds();
		int availTextWidth = mVerticalDisplay ? bounds.height : bounds.width;
		BufferedImage image = getCurrentImage();

		if (image != null) {
			availTextWidth -= IMAGE_TEXT_GAP + (mVerticalDisplay ? image.getHeight() : image.getWidth());
		}
		return availTextWidth;
	}

	/**
	 * The label painting method is contained in this routine so that subclasses can call label's
	 * paint directly without executing the paint code of intervening classes in the hierarchy.
	 * 
	 * @param g2d The graphics context
	 * @param clips An array of clip rectangles which need repainting
	 */
	protected void drawLabel(Graphics2D g2d, @SuppressWarnings("unused") Rectangle[] clips) {
		Rectangle bounds = getLocalInsetBounds();
		BufferedImage image = getCurrentImage();
		int availTextWidth = getAvailableTextWidth();
		String text = mText;
		int imageWidth = 0;
		int imageHeight = 0;
		int imageGap = 0;
		int textHeight = 0;
		int textWidth = 0;
		int x;
		int y;

		if (image != null) {
			imageWidth = image.getWidth();
			imageHeight = image.getHeight();
		}

		if (text.length() > 0) {
			ArrayList<String> list = new ArrayList<String>();
			Font font = getFont();
			FontRenderContext frc = g2d.getFontRenderContext();
			FontMetrics fm = getFontMetrics(font);
			int ascent = fm.getAscent();
			int fHeight = ascent + fm.getDescent(); // Don't use fm.getHeight(), as the PC adds too
			// much dead space
			int width;

			if (image != null) {
				imageGap = IMAGE_TEXT_GAP;
			}

			if (!mWrapText) {
				int where = text.indexOf('\n');

				if (where != -1) {
					text = text.substring(0, where);
				}

				text = TKTextDrawing.truncateIfNecessary(font, frc, text, availTextWidth, mTruncationPolicy);
				textHeight = fHeight;
				textWidth = TKTextDrawing.getPreferredSize(font, frc, text).width;
				list.add(text);
			} else {
				StringTokenizer tokenizer = new StringTokenizer(text, " \n", true); //$NON-NLS-1$
				StringBuilder buffer = new StringBuilder(text.length());

				while (tokenizer.hasMoreTokens()) {
					String token = tokenizer.nextToken();

					if (token.equals("\n")) { //$NON-NLS-1$
						text = buffer.toString();
						textHeight += fHeight;
						width = TKTextDrawing.getWidth(font, frc, text);
						if (width > textWidth) {
							textWidth = width;
						}
						list.add(text);
						buffer.setLength(0);
					} else {
						width = TKTextDrawing.getWidth(font, frc, buffer.toString() + token);
						if (width > availTextWidth && buffer.length() > 0) {
							text = buffer.toString();
							textHeight += fHeight;
							width = TKTextDrawing.getWidth(font, frc, text);
							if (width > textWidth) {
								textWidth = width;
							}
							list.add(text);
							buffer.setLength(0);
						}
						buffer.append(token);
					}
				}

				if (buffer.length() > 0) {
					text = buffer.toString();
					textHeight += fHeight;
					width = TKTextDrawing.getWidth(font, frc, text);
					if (width > textWidth) {
						textWidth = width;
					}
					list.add(text);
				}
			}

			textHeight += fm.getLeading();

			g2d.setColor(isEnabled() ? getForeground() : getDisabledForeground());
			if (mVerticalDisplay) {
				AffineTransform savedTransform = g2d.getTransform();

				if (mVerticalAlignment <= TKAlignment.TOP) {
					y = bounds.x;
				} else if (mVerticalAlignment == TKAlignment.CENTER) {
					y = bounds.x + (bounds.width - textHeight) / 2;
				} else {
					y = bounds.x + bounds.width - textHeight;
				}
				for (String piece : list) {
					x = bounds.y + bounds.height - (imageWidth + imageGap);
					x++; // Account for first pixel clipping
					if (mHorizontalAlignment == TKAlignment.CENTER) {
						x = x - (bounds.height - (imageWidth + imageGap + TKTextDrawing.getWidth(font, frc, piece))) / 2;
					} else if (mHorizontalAlignment >= TKAlignment.RIGHT) {
						x = x - bounds.height + 1 + imageWidth + imageGap + TKTextDrawing.getWidth(font, frc, piece);
					}

					g2d.rotate(Math.toRadians(270.0), y + ascent, x);
					if (mShadowColor != null) {
						Color oldColor = g2d.getColor();

						g2d.setColor(mShadowColor);
						g2d.drawString(piece, x + 2, y + ascent + 2);
						g2d.setColor(oldColor);
					}
					g2d.drawString(piece, y + ascent, x);
					g2d.setTransform(savedTransform);
					y += fHeight;
				}
			} else {
				if (mVerticalAlignment <= TKAlignment.TOP) {
					y = bounds.y;
				} else if (mVerticalAlignment == TKAlignment.CENTER) {
					y = bounds.y + (bounds.height - textHeight) / 2;
				} else {
					y = bounds.y + bounds.height - textHeight;
				}
				for (String piece : list) {
					x = bounds.x + imageWidth + imageGap;
					x++; // Account for first pixel clipping
					if (mHorizontalAlignment == TKAlignment.CENTER) {
						x = x + (bounds.width - (imageWidth + imageGap + TKTextDrawing.getWidth(font, frc, piece))) / 2;
					} else if (mHorizontalAlignment >= TKAlignment.RIGHT) {
						x = x + bounds.width - (1 + imageWidth + imageGap + TKTextDrawing.getWidth(font, frc, piece));
					}

					if (mShadowColor != null) {
						Color oldColor = g2d.getColor();

						g2d.setColor(mShadowColor);
						g2d.drawString(piece, x + 2, y + ascent + 2);
						g2d.setColor(oldColor);
					}
					g2d.drawString(piece, x, y + ascent);
					y += fHeight;
				}
			}
		}

		if (image != null) {
			if (mHorizontalAlignment <= TKAlignment.LEFT) {
				x = bounds.x;
			} else if (mHorizontalAlignment == TKAlignment.CENTER) {
				x = bounds.x + (bounds.width - (imageWidth + imageGap + textWidth)) / 2;
			} else {
				x = bounds.x + bounds.width - (imageWidth + imageGap + textWidth);
			}
			if (mVerticalAlignment <= TKAlignment.TOP) {
				y = bounds.y;
			} else if (mVerticalAlignment == TKAlignment.CENTER) {
				y = bounds.y + (bounds.height - imageHeight) / 2;
			} else {
				y = bounds.y + bounds.height - imageHeight;
			}
			drawImage(g2d, image, x, y);
		}
	}

	/**
	 * Draws the image for this label.
	 * 
	 * @param g2d The graphics context to use.
	 * @param image The image to draw.
	 * @param x The x-coordinate to use.
	 * @param y The y-coordinate to use.
	 */
	protected void drawImage(Graphics2D g2d, BufferedImage image, int x, int y) {
		g2d.drawImage(image, x, y, null);
	}

	/**
	 * @param color The "shadow" color for this label. If the color is <code>null</code>, then no
	 *            shadow will be drawn.
	 */
	public void setShadowColor(Color color) {
		if (color != mShadowColor && (color == null || !color.equals(mShadowColor))) {
			mShadowColor = color;
			repaint();
		}
	}

	/** @return The draggable data. */
	public Transferable getDraggableData() {
		return mDragData;
	}

	/** @param data The draggable data. */
	public void setDraggableData(Transferable data) {
		mDragData = data;
		if (data != null) {
			if (mDragGestureRecognizer == null) {
				mDragGestureRecognizer = DragSource.getDefaultDragSource().createDefaultDragGestureRecognizer(this, DnDConstants.ACTION_COPY_OR_MOVE, this);
			}
		} else {
			if (mDragGestureRecognizer != null) {
				mDragGestureRecognizer.removeDragGestureListener(this);
				mDragGestureRecognizer = null;
			}
		}
	}

	/** @return The draggable image. */
	public BufferedImage getDraggableImage() {
		return getImage(null);
	}

	@Override public boolean getDefaultToolbarEnabledState() {
		return true;
	}

	public void dragGestureRecognized(DragGestureEvent dge) {
		TKDragUtil.prepDrag();
		if (mDragData != null) {
			if (DragSource.isDragImageSupported()) {
				dge.startDrag(null, getDraggableImage(), dge.getDragOrigin(), mDragData, null);
			} else {
				dge.startDrag(null, mDragData);
			}
		}
	}
}
