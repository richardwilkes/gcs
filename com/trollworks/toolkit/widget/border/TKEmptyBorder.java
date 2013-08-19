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

/** An empty border that takes up space but doesn't draw. */
public class TKEmptyBorder implements TKBorder {
	private int	mTop;
	private int	mLeft;
	private int	mBottom;
	private int	mRight;

	/**
	 * Creates a new empty border.
	 * 
	 * @param margin The inset for each side.
	 */
	public TKEmptyBorder(int margin) {
		mTop = margin;
		mLeft = margin;
		mBottom = margin;
		mRight = margin;
	}

	/**
	 * Creates a new empty border.
	 * 
	 * @param top The top inset of the border.
	 * @param left The left inset of the border.
	 * @param bottom The bottom inset of the border.
	 * @param right The right inset of the border.
	 */
	public TKEmptyBorder(int top, int left, int bottom, int right) {
		mTop = top;
		mLeft = left;
		mBottom = bottom;
		mRight = right;
	}

	/**
	 * Creates a new empty border.
	 * 
	 * @param insets The insets to use for the border.
	 */
	public TKEmptyBorder(Insets insets) {
		mTop = insets.top;
		mLeft = insets.left;
		mBottom = insets.bottom;
		mRight = insets.right;
	}

	public void paintBorder(TKPanel panel, Graphics2D g2d, int x, int y, int width, int height) {
		// No border to paint...
	}

	public Insets getBorderInsets(TKPanel panel) {
		return new Insets(mTop, mLeft, mBottom, mRight);
	}
}
