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

package com.trollworks.toolkit.widget.menu;

import com.trollworks.toolkit.text.TKTextDrawing;
import com.trollworks.toolkit.utility.TKColor;

import java.awt.Color;
import java.awt.Graphics2D;
import java.awt.image.BufferedImage;

/** A standard menu separator. */
public class TKMenuSeparator extends TKMenuItem {
	private boolean	mIndented;

	/** Creates a new menu separator. */
	public TKMenuSeparator() {
		this(null, null);
	}

	/**
	 * Creates a new menu separator whose divider line starts at the text in a normal menu.
	 * 
	 * @param indented Whether or not the separator should be indented.
	 */
	public TKMenuSeparator(boolean indented) {
		this(null, null);
		mIndented = indented;
	}

	/**
	 * Creates a new menu separator that has a title.
	 * 
	 * @param title The separator title.
	 */
	public TKMenuSeparator(String title) {
		this(title, null);
	}

	/**
	 * Creates a new menu separator that has an icon.
	 * 
	 * @param icon The separator icon.
	 */
	public TKMenuSeparator(BufferedImage icon) {
		this(null, icon);
	}

	/**
	 * Creates a new menu separator that has both a title and an icon.
	 * 
	 * @param title The separator title.
	 * @param icon The separator icon.
	 */
	public TKMenuSeparator(String title, BufferedImage icon) {
		super(title, icon, null, null);
		setEnabled(false);
	}

	@Override public void draw(Graphics2D g2d, int x, int y, int width, int height, Color color) {
		x += 2;
		width -= 4;

		Color oldColor = g2d.getColor();
		int middle = y + height / 2;
		int left = x + getIndent() + H_MARGIN + getMarkWidth();
		int blankWidth = 0;
		BufferedImage icon = getIcon();
		String title = getTitle();

		if (icon != null) {
			blankWidth += icon.getWidth();
		}

		if (title != null && title.length() > 0) {
			blankWidth += TKTextDrawing.getPreferredSize(getFont(), g2d.getFontRenderContext(), title).width;
		}

		if (icon != null && title != null) {
			blankWidth += GAP;
		}

		if (icon != null || title != null) {
			blankWidth += 2 * GAP;
		}

		g2d.setColor(TKColor.MENU_SHADOW);
		if (blankWidth > 0) {
			if (!mIndented) {
				g2d.drawLine(x, middle, left, middle);
			} else {
				g2d.setColor(TKColor.darker(TKColor.MENU_BACKGROUND, 10));
			}
			g2d.drawLine(left + blankWidth, middle, width, middle);
		} else if (mIndented) {
			g2d.setColor(TKColor.darker(TKColor.MENU_BACKGROUND, 10));
			g2d.drawLine(left + H_MARGIN, middle, width, middle);
		} else {
			g2d.drawLine(x, middle, width, middle);
		}

		if (!mIndented) {
			g2d.setColor(TKColor.MENU_HIGHLIGHT);
			if (blankWidth > 0) {
				g2d.drawLine(x, middle + 1, left, middle + 1);
				g2d.drawLine(left + blankWidth, middle + 1, width, middle + 1);
			} else if (!mIndented) {
				g2d.drawLine(x, middle + 1, width, middle + 1);
			}
		}

		g2d.setColor(oldColor);
		super.draw(g2d, x, y, width, height, color);
	}

	@Override public Color getTitleColor() {
		return TKColor.MENU_SEPARATOR;
	}
}
