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

package com.trollworks.toolkit.widget.button;

import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.window.TKBaseWindow;
import com.trollworks.toolkit.window.TKDialog;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.image.BufferedImage;

/** A standard push button. */
public class TKButton extends TKBaseButton {
	/** Creates a button with no icon and no text. */
	public TKButton() {
		this(null, null, TKAlignment.CENTER);
	}

	/**
	 * Creates a button with the specified text.
	 * 
	 * @param text The text to use.
	 */
	public TKButton(String text) {
		this(text, null, TKAlignment.CENTER);
	}

	/**
	 * Creates a button with the specified text and horizontal alignment.
	 * 
	 * @param text The text to use.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKButton(String text, int alignment) {
		this(text, null, alignment);
	}

	/**
	 * Creates a button with the specified icon.
	 * 
	 * @param icon The icon to use.
	 */
	public TKButton(BufferedImage icon) {
		this(null, icon, TKAlignment.CENTER);
	}

	/**
	 * Creates a button with the specified icon and horizontal alignment.
	 * 
	 * @param icon The icon to use.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKButton(BufferedImage icon, int alignment) {
		this(null, icon, alignment);
	}

	/**
	 * Creates a button with the specified text and icon.
	 * 
	 * @param text The text to use.
	 * @param icon The icon to use.
	 */
	public TKButton(String text, BufferedImage icon) {
		this(text, icon, TKAlignment.CENTER);
	}

	/**
	 * Creates a button with the specified text, icon and horizontal alignment.
	 * 
	 * @param text The text to use.
	 * @param icon The icon to use.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKButton(String text, BufferedImage icon, int alignment) {
		super(text, icon, alignment);
		setBorderDisplayed(true);
	}

	@Override protected Dimension getPreferredSizeSelf() {
		Dimension size = super.getPreferredSizeSelf();
		int extraHeight = 0;
		int extraWidth = 0;

		if (getImage() != null) {
			extraHeight = 6;
			extraWidth = 6;
		}

		if (getText().length() > 0) {
			FontMetrics fm = getFontMetrics(getFont());
			int tmp = fm.getAscent() + fm.getDescent() + 8 - size.height; // Don't use
			// fm.getHeight(), as
			// the PC adds too much
			// dead space

			if (tmp > extraHeight) {
				extraHeight = tmp;
			}

			tmp = (int) Math.ceil(size.height / 2.0 + 6.0);
			if (tmp > extraWidth) {
				extraWidth = tmp;
			}
		}

		size.height += extraHeight;
		size.width += extraWidth;

		return size;
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		Color oldColor = g2d.getColor();
		Insets insets = getInsets();
		int x = insets.left;
		int y = insets.top;
		int w = getWidth() - 1;
		int h = getHeight() - 1;
		int ix = x + w - 1;
		int iy = y + h - 1;
		boolean pressed = isPressed();

		g2d.setColor(pressed ? TKColor.CONTROL_PRESSED_FILL : mInRollOver ? TKColor.CONTROL_ROLL : TKColor.CONTROL_FILL);
		g2d.fillRect(x, y, w, h);
		if (isBorderDisplayed()) {
			g2d.setColor(TKColor.CONTROL_LINE);
			g2d.drawRect(x, y, w, h);

			if (pressed) {
				g2d.setColor(TKColor.CONTROL_HIGHLIGHT);
			} else {
				g2d.setColor(TKColor.CONTROL_SHADOW);
			}

			g2d.drawLine(ix, y + 1, ix, iy);
			g2d.drawLine(x + 1, iy, ix, iy);

			if (!pressed) {
				g2d.setColor(TKColor.CONTROL_HIGHLIGHT);
				g2d.drawLine(x + 1, y + 1, x + 1, iy);
				g2d.drawLine(x + 1, y + 1, ix, y + 1);
			}
		}

		if (isEnabled()) {
			TKBaseWindow window = getBaseWindow();

			if (window instanceof TKDialog && window.isInForeground() && ((TKDialog) window).getDefaultButton() == this) {
				x += 2;
				y += 2;
				w -= 4;
				h -= 4;
				g2d.setColor(TKColor.DEFAULT_BUTTON_INDICATOR);
				g2d.drawRect(x, y, w, h);
			}
		}

		g2d.setColor(oldColor);
		super.paintPanel(g2d, clips);
	}
}
