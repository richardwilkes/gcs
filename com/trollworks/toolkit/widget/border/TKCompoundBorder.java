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

/** A compound border which merges two borders into a single border by nesting one within the other. */
public class TKCompoundBorder implements TKBorder {
	private TKBorder	mInnerBorder;
	private TKBorder	mOuterBorder;

	/**
	 * Creates a new compound border.
	 * 
	 * @param outerBorder The outer border.
	 * @param innerBorder The inner border.
	 */
	public TKCompoundBorder(TKBorder outerBorder, TKBorder innerBorder) {
		mInnerBorder = innerBorder;
		mOuterBorder = outerBorder;
	}

	public void paintBorder(TKPanel panel, Graphics2D g2d, int x, int y, int width, int height) {
		Insets insets = mOuterBorder.getBorderInsets(panel);

		mOuterBorder.paintBorder(panel, g2d, x, y, width, height);
		mInnerBorder.paintBorder(panel, g2d, x + insets.left, y + insets.top, width - (insets.left + insets.right), height - (insets.top + insets.bottom));
	}

	public Insets getBorderInsets(TKPanel panel) {
		Insets inside = mInnerBorder.getBorderInsets(panel);
		Insets outside = mOuterBorder.getBorderInsets(panel);

		inside.top += outside.top;
		inside.left += outside.left;
		inside.bottom += outside.bottom;
		inside.right += outside.right;
		return inside;
	}
}
