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

import com.trollworks.toolkit.utility.TKColor;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Rectangle;

/** Provides a simple colored panel. */
public class TKColorSwatch extends TKPanel {
	private Color	mColor;

	/**
	 * Creates the color swatch.
	 * 
	 * @param color The color to display.
	 */
	public TKColorSwatch(Color color) {
		super();
		mColor = color;
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		Rectangle bounds = getLocalInsetBounds();

		g2d.setPaint(mColor);
		g2d.fillRect(bounds.x, bounds.y, bounds.width, bounds.height);
		g2d.setColor(TKColor.CONTROL_LINE);
		g2d.drawRect(bounds.x, bounds.y, bounds.width, bounds.height);
	}

	@Override protected Dimension getMaximumSizeSelf() {
		return getPreferredSizeSelf();
	}

	@Override protected Dimension getMinimumSizeSelf() {
		return getPreferredSizeSelf();
	}

	@Override protected Dimension getPreferredSizeSelf() {
		return new Dimension(25, 12);
	}

	/** @return This swatch's color. */
	public Color getColor() {
		return mColor;
	}

	/** @param newColor This swatch's color. */
	public void setColor(Color newColor) {
		mColor = newColor;
		revalidate();
	}
}
