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
import java.awt.Insets;
import java.awt.Rectangle;

/** A simple divider. */
public class TKDivider extends TKPanel {
	private boolean	mVertical;
	private int		mThickness;

	/** Creates a one-pixel thick horizontal black divider. */
	public TKDivider() {
		this(false, null, 1);
	}

	/**
	 * Creates a beveled divider.
	 * 
	 * @param vertical Pass in <code>true</code> for a vertical divider and <code>false</code>
	 *            for a horizontal one.
	 */
	public TKDivider(boolean vertical) {
		this(vertical, null, 0);
	}

	/**
	 * Creates a one-pixel thick horizontal divider.
	 * 
	 * @param color The color to draw the divider.
	 */
	public TKDivider(Color color) {
		this(false, color, 1);
	}

	/**
	 * Creates a one-pixel thick divider.
	 * 
	 * @param vertical Pass in <code>true</code> for a vertical divider and <code>false</code>
	 *            for a horizontal one.
	 * @param color The color to draw the divider.
	 */
	public TKDivider(boolean vertical, Color color) {
		this(vertical, color, 1);
	}

	/**
	 * Creates a black divider.
	 * 
	 * @param vertical Pass in <code>true</code> for a vertical divider and <code>false</code>
	 *            for a horizontal one.
	 * @param thickness The thickness to make this divider.
	 */
	public TKDivider(boolean vertical, int thickness) {
		this(vertical, null, thickness);
	}

	/**
	 * Creates a divider.
	 * 
	 * @param vertical Pass in <code>true</code> for a vertical divider and <code>false</code>
	 *            for a horizontal one.
	 * @param color The color to draw the divider.
	 * @param thickness The thickness to make this divider.
	 */
	public TKDivider(boolean vertical, Color color, int thickness) {
		super();
		mVertical = vertical;
		mThickness = thickness < 1 ? 0 : thickness;
		setForeground(color == null ? Color.black : color);
	}

	@Override protected Dimension getMaximumSizeSelf() {
		Insets insets = getInsets();

		if (isVertical()) {
			return new Dimension(getThickness() + insets.left + insets.right, MAX_SIZE);
		}
		return new Dimension(MAX_SIZE, getThickness() + insets.top + insets.bottom);
	}

	@Override protected Dimension getMinimumSizeSelf() {
		Insets insets = getInsets();

		if (isVertical()) {
			return new Dimension(getThickness() + insets.left + insets.right, 1 + insets.top + insets.bottom);
		}
		return new Dimension(1 + insets.left + insets.right, getThickness() + insets.top + insets.bottom);
	}

	@Override protected Dimension getPreferredSizeSelf() {
		return getMinimumSizeSelf();
	}

	/** @return The thickness of this divider. */
	public int getThickness() {
		return mThickness > 0 ? mThickness : 2;
	}

	/** @return <code>true</code> if this divider is vertically oriented. */
	public boolean isVertical() {
		return mVertical;
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		Rectangle bounds = getLocalInsetBounds();

		if (mThickness == 0) {
			if (isVertical()) {
				g2d.setColor(TKColor.CONTROL_SHADOW);
				g2d.drawLine(bounds.x, bounds.y + 1, bounds.x, bounds.y + bounds.height - 2);
				g2d.setColor(TKColor.CONTROL_HIGHLIGHT);
				g2d.drawLine(bounds.x + 1, bounds.y + 1, bounds.x + 1, bounds.y + bounds.height - 2);
			} else {
				g2d.setColor(TKColor.CONTROL_SHADOW);
				g2d.drawLine(bounds.x + 1, bounds.y, bounds.x + bounds.width - 2, bounds.y);
				g2d.setColor(TKColor.CONTROL_HIGHLIGHT);
				g2d.drawLine(bounds.x + 1, bounds.y + 1, bounds.x + bounds.width - 2, bounds.y + 1);
			}
		} else {
			int thickness = getThickness();

			if (isVertical()) {
				bounds.x += (bounds.width - thickness) / 2;
				bounds.width = thickness;
			} else {
				bounds.y += (bounds.height - thickness) / 2;
				bounds.height = thickness;
			}

			g2d.setColor(getForeground());
			g2d.fill(bounds);
		}
	}

	/** @param thickness The thickness of this divider. */
	public void setThickness(int thickness) {
		if (mThickness > 0 && mThickness != thickness && thickness > 0) {
			mThickness = thickness;
			revalidate();
		}
	}
}
