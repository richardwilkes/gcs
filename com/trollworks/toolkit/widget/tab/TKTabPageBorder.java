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

package com.trollworks.toolkit.widget.tab;

import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKBorder;

import java.awt.Color;
import java.awt.Graphics2D;
import java.awt.Insets;

/**
 * Manages a border around a page display in a {@link TKTabGroup}. A tab page border has a 1-pixel
 * raised bevel look which is 'knocked out' around the selected {@link TKTab}.
 */
public class TKTabPageBorder implements TKBorder {
	private TKTabGroup	mTabs;
	private Insets		mInsets;
	private Color		mColor;
	private boolean		mHasScrollbars;

	/**
	 * Constructs a new tab page border for the given tab group with the given inset on all sides
	 * and the given color.
	 * 
	 * @param tabs The tab group that this border surrounds.
	 * @param inset The border inset value.
	 * @param color The color of the tab page border.
	 * @param hasScrollbars If <code>true</code>, the page being outlined has scrollbars pushed
	 *            all the way to the edge.
	 */
	public TKTabPageBorder(TKTabGroup tabs, int inset, Color color, boolean hasScrollbars) {
		mTabs = tabs;
		mInsets = new Insets(inset, inset, inset, inset);
		mColor = color;
		mHasScrollbars = hasScrollbars;
	}

	public void paintBorder(TKPanel panel, Graphics2D g2d, int x, int y, int width, int height) {
		Color saveColor = g2d.getColor();
		TKTab selectedTab = (TKTab) mTabs.getSelection();
		int tabX = selectedTab.getX();
		int tabWidth = selectedTab.getWidth() - 1;

		width--;
		height--;

		g2d.setColor(mColor);
		g2d.drawLine(x, y, tabX, y);
		g2d.drawLine(tabX + tabWidth, y, x + width, y);
		g2d.drawLine(x, y, x, y + height);

		if (!mHasScrollbars) {
			g2d.drawLine(x + width, y, x + width, y + height);
			g2d.drawLine(x, y + height, x + width, y + height);

			g2d.setColor(TKColor.CONTROL_HIGHLIGHT);
			g2d.drawLine(x + 1, y + 1, tabX + 1, y + 1);
			g2d.drawLine(tabX + tabWidth - 1, y + 1, x + width - 1, y + 1);
			g2d.drawLine(x + 1, y + 1, x + 1, y + height - 1);

			g2d.setColor(TKColor.CONTROL_SHADOW);
			g2d.drawLine(x + width - 1, y + 2, x + width - 1, y + height - 1);
			g2d.drawLine(x + 1, y + height - 1, x + width - 1, y + height - 1);
		}

		g2d.setColor(saveColor);
	}

	public Insets getBorderInsets(TKPanel panel) {
		return mInsets;
	}
}
