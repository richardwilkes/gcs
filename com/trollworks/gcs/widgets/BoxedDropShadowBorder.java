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

package com.trollworks.gcs.widgets;

import com.trollworks.gcs.utility.text.TextDrawing;

import java.awt.Color;
import java.awt.Component;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.font.FontRenderContext;

import javax.swing.SwingConstants;
import javax.swing.border.Border;

/** A border consisting of a frame, drop shadow, and optional title. */
public class BoxedDropShadowBorder implements Border {
	private static BoxedDropShadowBorder	INSTANCE	= null;
	private String							mTitle;
	private Font							mFont;

	/** Creates a new border without a title. */
	public BoxedDropShadowBorder() {
		super();
	}

	/**
	 * Creates a new border with a title.
	 * 
	 * @param font The font to use.
	 * @param title The title to use.
	 */
	public BoxedDropShadowBorder(Font font, String title) {
		super();
		mFont = font;
		mTitle = title;
	}

	/** @return The shared, non-titled border. */
	public static synchronized BoxedDropShadowBorder get() {
		if (INSTANCE == null) {
			INSTANCE = new BoxedDropShadowBorder();
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

	public Insets getBorderInsets(Component component) {
		Insets insets = new Insets(1, 1, 3, 3);
		if (mTitle != null) {
			insets.top += TextDrawing.getPreferredSize(mFont, null, mTitle).height;
		}
		return insets;
	}

	public boolean isBorderOpaque() {
		return false;
	}

	public void paintBorder(Component component, Graphics gc, int x, int y, int width, int height) {
		Color savedColor = gc.getColor();
		gc.setColor(Color.lightGray);
		gc.drawLine(x + width - 2, y + 2, x + width - 2, y + height - 1);
		gc.drawLine(x + width - 1, y + 2, x + width - 1, y + height - 1);
		gc.drawLine(x + 2, y + height - 2, x + width - 1, y + height - 2);
		gc.drawLine(x + 2, y + height - 1, x + width - 1, y + height - 1);
		gc.setColor(Color.black);
		gc.drawRect(x, y, width - 3, height - 3);
		if (mTitle != null) {
			Font savedFont = gc.getFont();
			gc.setFont(mFont);
			FontRenderContext frc = ((Graphics2D) gc).getFontRenderContext();
			int th = TextDrawing.getPreferredSize(mFont, frc, mTitle).height;
			Rectangle bounds = new Rectangle(x, y, width - 3, th + 2);
			gc.fillRect(x, y, width - 3, th + 1);
			gc.setColor(Color.white);
			TextDrawing.draw(gc, bounds, mTitle, SwingConstants.CENTER, SwingConstants.TOP);
			gc.setFont(savedFont);
		}
		gc.setColor(savedColor);
	}
}
