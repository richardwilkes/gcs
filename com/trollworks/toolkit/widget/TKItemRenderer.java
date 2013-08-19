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

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.geom.Rectangle2D;

/** The interface an item renderer must implement. */
public interface TKItemRenderer {
	/**
	 * Draw the specified item.
	 * 
	 * @param g2d The graphics context.
	 * @param bounds The bounds for the item.
	 * @param item The item to draw.
	 * @param index The item's index.
	 * @param selected <code>true</code> if the item is selected.
	 * @param active <code>true</code> if the item is in an active window.
	 */
	public void drawItem(Graphics2D g2d, Rectangle2D bounds, Object item, int index, boolean selected, boolean active);

	/**
	 * @param item The item to return the color for.
	 * @param index The item's index.
	 * @param selected <code>true</code> if the item is selected.
	 * @param active <code>true</code> if the item is in an active window.
	 * @return The background color for the specified item.
	 */
	public Color getBackgroundForItem(Object item, int index, boolean selected, boolean active);

	/**
	 * @param item The item to return the size for.
	 * @param index The item's index.
	 * @return The preferred size for the specified item.
	 */
	public Dimension getItemPreferredSize(Object item, int index);
}
