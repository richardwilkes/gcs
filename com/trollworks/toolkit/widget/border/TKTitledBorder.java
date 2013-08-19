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

package com.trollworks.toolkit.widget.border;

import com.trollworks.toolkit.text.TKTextDrawing;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKPanel;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.font.FontRenderContext;

/** A border with a title and a line around it. */
public class TKTitledBorder implements TKBorder {
	private static final int	GAP	= 3;
	private Color				mTitleColor;
	private Font				mTitleFont;
	private String				mTitle;
	private Dimension			mCachedSize;
	private boolean				mInsetBorder;

	/**
	 * Creates a new titled border.
	 * 
	 * @param title The title of the border.
	 */
	public TKTitledBorder(String title) {
		this(title, null, null, true);
	}

	/**
	 * Creates a new titled border.
	 * 
	 * @param title The title of the border.
	 * @param insetBorder Pass in <code>true</code> if the normal border insets should be
	 *            reported, or <code>false</code> to report the left, right and bottom insets as
	 *            zero rather than their actual values.
	 */
	public TKTitledBorder(String title, boolean insetBorder) {
		this(title, null, null, insetBorder);
	}

	/**
	 * Creates a new titled border.
	 * 
	 * @param title The title of the border.
	 * @param font The font for the title.
	 */
	public TKTitledBorder(String title, Font font) {
		this(title, font, null, true);
	}

	/**
	 * Creates a new titled border.
	 * 
	 * @param title The title of the border.
	 * @param font The font for the title.
	 * @param insetBorder Pass in <code>true</code> if the normal border insets should be
	 *            reported, or <code>false</code> to report the left, right and bottom insets as
	 *            zero rather than their actual values.
	 */
	public TKTitledBorder(String title, Font font, boolean insetBorder) {
		this(title, font, null, insetBorder);
	}

	/**
	 * Creates a new titled border.
	 * 
	 * @param title The title of the border.
	 * @param font The font for the title.
	 * @param titleColor The color to use for the title.
	 * @param insetBorder Pass in <code>true</code> if the normal border insets should be
	 *            reported, or <code>false</code> to report the left, right and bottom insets as
	 *            zero rather than their actual values.
	 */
	public TKTitledBorder(String title, Font font, Color titleColor, boolean insetBorder) {
		mTitle = title;
		mTitleColor = titleColor == null ? Color.black : titleColor;
		mTitleFont = font == null ? TKFont.lookup(TKFont.CONTROL_FONT_KEY) : font;
		mInsetBorder = insetBorder;
	}

	public void paintBorder(TKPanel panel, Graphics2D g2d, int x, int y, int width, int height) {
		Color oldColor = g2d.getColor();
		Color bgColor = panel.getBackground();
		Color shadow = TKColor.darker(bgColor, 45);

		height--;
		width--;
		g2d.setColor(shadow);
		if (width > GAP * 2 && mTitle != null) {
			Font oldFont = g2d.getFont();
			FontRenderContext frc = g2d.getFontRenderContext();
			String title = TKTextDrawing.truncateIfNecessary(mTitleFont, frc, mTitle, width - GAP * 2, TKAlignment.RIGHT);
			Dimension size = TKTextDrawing.getPreferredSize(mTitleFont, frc, title);
			int origY = y;
			int tmp;

			tmp = size.height / 2;
			height = height - tmp;
			y += tmp;

			g2d.drawLine(x + 1, y, x + GAP - 1, y);
			g2d.drawLine(x + GAP + size.width + GAP - 1, y, x + width - 1, y);

			g2d.setColor(mTitleColor);
			g2d.setFont(mTitleFont);
			TKTextDrawing.draw(g2d, x + GAP, origY, title);
			g2d.setFont(oldFont);
		} else {
			g2d.drawLine(x + 1, y, x + width - 1, y);
		}

		g2d.setColor(shadow);
		g2d.drawLine(x, y + 1, x, y + height - 1);
		g2d.drawLine(x + width, y + 1, x + width, y + height - 1);
		g2d.drawLine(x + 1, y + height, x + width - 1, y + height);

		g2d.setColor(oldColor);
	}

	public Insets getBorderInsets(TKPanel panel) {
		int inset = mInsetBorder ? 1 : 0;

		return new Insets(getTitleSize().height + 1, inset, inset, inset);
	}

	/** @return The size of the title. */
	public Dimension getTitleSize() {
		if (mCachedSize == null) {
			if (mTitle == null) {
				mCachedSize = new Dimension(1, 1);
			} else {
				mCachedSize = TKTextDrawing.getPreferredSize(mTitleFont, null, mTitle);
			}
		}
		return mCachedSize;
	}

	/**
	 * Sets the font for the title.
	 * 
	 * @param font The font to use.
	 */
	public void setFont(Font font) {
		mTitleFont = font;
		mCachedSize = null;
	}

	/**
	 * Sets the title.
	 * 
	 * @param title The title to use.
	 */
	public void setTitle(String title) {
		mTitle = title;
		mCachedSize = null;
	}
}
