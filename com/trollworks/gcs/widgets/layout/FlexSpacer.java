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

package com.trollworks.gcs.widgets.layout;

import java.awt.Dimension;
import java.awt.Rectangle;

/** A spacer within a {@link FlexLayout}. */
public class FlexSpacer extends FlexCell {
	private int		mWidth;
	private int		mHeight;
	private boolean	mGrowWidth;
	private boolean	mGrowHeight;

	/**
	 * Creates a new {@link FlexSpacer}.
	 * 
	 * @param width The width of the spacer.
	 * @param height The height of the spacer.
	 * @param growWidth Whether the width of the spacer can grow.
	 * @param growHeight Whether the height of the spacer can grow.
	 */
	public FlexSpacer(int width, int height, boolean growWidth, boolean growHeight) {
		mWidth = width;
		mHeight = height;
		mGrowWidth = growWidth;
		mGrowHeight = growHeight;
	}

	@Override protected Dimension getSizeSelf(LayoutSize type) {
		int width = mWidth;
		int height = mHeight;
		if (type == LayoutSize.MAXIMUM) {
			if (mGrowWidth) {
				width = Integer.MAX_VALUE;
			}
			if (mGrowHeight) {
				height = Integer.MAX_VALUE;
			}
		}
		return new Dimension(width, height);
	}

	@Override protected void layoutSelf(Rectangle bounds) {
		// Does nothing
	}
}
