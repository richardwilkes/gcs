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

import com.trollworks.toolkit.widget.TKPanel;

import java.awt.Graphics2D;
import java.awt.Insets;

/** All panel borders must implement this interface. */
public interface TKBorder {
	/**
	 * Paints the border for a panel.
	 * 
	 * @param panel The panel to paint the border for.
	 * @param g2d The graphics object to use.
	 * @param x The x coordinate of the border.
	 * @param y The y coordinate of the border.
	 * @param width The width of the border.
	 * @param height The height of the border.
	 */
	public void paintBorder(TKPanel panel, Graphics2D g2d, int x, int y, int width, int height);

	/**
	 * @param panel The panel this border is being used for.
	 * @return The insets of this border.
	 */
	public Insets getBorderInsets(TKPanel panel);
}
