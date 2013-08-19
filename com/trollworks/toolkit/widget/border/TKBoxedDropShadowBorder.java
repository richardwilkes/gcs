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
import com.trollworks.toolkit.widget.TKPanel;

import java.awt.Color;
import java.awt.Font;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.font.FontRenderContext;

/** A border consisting of a frame, drop shadow, and optional title. */
public class TKBoxedDropShadowBorder implements TKBorder {
	private static TKBoxedDropShadowBorder	INSTANCE	= null;
	private String							mTitle;
	private Font							mFont;

	/** Creates a new border without a title. */
	public TKBoxedDropShadowBorder() {
		super();
	}

	/**
	 * Creates a new border with a title.
	 * 
	 * @param font The font to use.
	 * @param title The title to use.
	 */
	public TKBoxedDropShadowBorder(Font font, String title) {
		super();
		mFont = font;
		mTitle = title;
	}

	/** @return The shared, non-titled border. */
	public static synchronized TKBoxedDropShadowBorder get() {
		if (INSTANCE == null) {
			INSTANCE = new TKBoxedDropShadowBorder();
		}
		return INSTANCE;
	}

	/** @return The title. */
	public String getTitle() {
		return mTitle;
	}

	/** @param title The new title. */
	public void setTitle(String title) {
		mTitle = title;
	}

	public void paintBorder(TKPanel panel, Graphics2D g2d, int x, int y, int width, int height) {
		Color savedColor = g2d.getColor();

		g2d.setColor(Color.lightGray);
		g2d.drawLine(x + width - 2, y + 2, x + width - 2, y + height - 1);
		g2d.drawLine(x + width - 1, y + 2, x + width - 1, y + height - 1);
		g2d.drawLine(x + 2, y + height - 2, x + width - 1, y + height - 2);
		g2d.drawLine(x + 2, y + height - 1, x + width - 1, y + height - 1);
		g2d.setColor(Color.black);
		g2d.drawRect(x, y, width - 3, height - 3);
		if (mTitle != null) {
			Font savedFont = g2d.getFont();
			g2d.setFont(mFont);
			FontRenderContext frc = g2d.getFontRenderContext();
			int th = TKTextDrawing.getPreferredSize(mFont, frc, mTitle).height;
			Rectangle bounds = new Rectangle(x, y, width - 3, th + 2);

			g2d.fillRect(x, y, width - 3, th + 1);
			g2d.setColor(Color.white);
			TKTextDrawing.draw(g2d, bounds, mTitle, TKAlignment.CENTER, TKAlignment.TOP);
			g2d.setFont(savedFont);
		}
		g2d.setColor(savedColor);
	}

	public Insets getBorderInsets(TKPanel panel) {
		Insets insets = new Insets(1, 1, 3, 3);

		if (mTitle != null) {
			insets.top += TKTextDrawing.getPreferredSize(mFont, null, mTitle).height;
		}
		return insets;
	}
}
