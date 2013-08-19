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

import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.Shape;
import java.awt.geom.Rectangle2D;
import java.awt.image.BufferedImage;

/** The default item renderer. */
public class TKDefaultItemRenderer extends TKLabel implements TKItemRenderer {
	/** Creates a new item renderer using the app font. */
	public TKDefaultItemRenderer() {
		this(null);
	}

	/**
	 * Creates a new item renderer using the specified font.
	 * 
	 * @param font The font to use.
	 */
	public TKDefaultItemRenderer(String font) {
		super(null, null, TKAlignment.LEFT, true, font != null ? font : TKFont.TEXT_FONT_KEY);
		setBorder(new TKEmptyBorder(0, 2, 0, 2));
		setVerticalAlignment(TKAlignment.TOP);
	}

	public void drawItem(Graphics2D g2d, Rectangle2D bounds, Object item, int index, boolean selected, boolean active) {
		Shape clip = g2d.getClip();
		Rectangle itemBounds = bounds.getBounds();

		prepareForItem(item, index, selected, active);
		setBounds(itemBounds);
		g2d.clipRect(itemBounds.x, itemBounds.y, itemBounds.width, itemBounds.height);
		g2d.translate(itemBounds.x, itemBounds.y);
		g2d.setFont(TKFont.lookup(getFontKeyForItem(item, index)));
		paintPanel(g2d, new Rectangle[] { getLocalBounds() });
		g2d.translate(-itemBounds.x, -itemBounds.y);
		g2d.setClip(clip);
	}

	public Color getBackgroundForItem(Object item, int index, boolean selected, boolean active) {
		return selected ? active ? TKColor.HIGHLIGHT : TKColor.INACTIVE_HIGHLIGHT : TKColor.TEXT_BACKGROUND;
	}

	/**
	 * @param item The item to return the font for.
	 * @param index The item's index.
	 * @return The font that should be used for the specified item. By default, returns
	 *         <code>getFont()</code>.
	 */
	public String getFontKeyForItem(@SuppressWarnings("unused") Object item, @SuppressWarnings("unused") int index) {
		return getFontKey();
	}

	/**
	 * @param item The item to return the color for.
	 * @param index The item's index.
	 * @param selected <code>true</code> if the item is selected.
	 * @param active <code>true</code> if this item is in an active window.
	 * @return The foreground color that should be used for this item.
	 */
	public Color getForegroundForItem(@SuppressWarnings("unused") Object item, @SuppressWarnings("unused") int index, boolean selected, @SuppressWarnings("unused") boolean active) {
		return selected ? TKColor.HIGHLIGHTED_TEXT : TKColor.TEXT;
	}

	/**
	 * @param item The item to return the image for.
	 * @param index The item's index.
	 * @return The image representation that should be used for this item. By default,
	 *         <code>null</code> is returned.
	 */
	public BufferedImage getImageForItem(@SuppressWarnings("unused") Object item, @SuppressWarnings("unused") int index) {
		return null;
	}

	/**
	 * @param item The item to return the string representation for.
	 * @param index The item's index.
	 * @return The string representation of the specified item. By default,
	 *         <code>item.toString()</code> is returned.
	 */
	public String getStringForItem(Object item, @SuppressWarnings("unused") int index) {
		return item.toString();
	}

	public Dimension getItemPreferredSize(Object item, int index) {
		prepareForItem(item, index, false, true);
		return getPreferredSize();
	}

	private void prepareForItem(Object item, int index, boolean selected, boolean active) {
		setImage(getImageForItem(item, index));
		setText(getStringForItem(item, index));
		setForeground(getForegroundForItem(item, index, selected, active));
		setFontKey(getFontKeyForItem(item, index));
	}
}
