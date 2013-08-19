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
import com.trollworks.toolkit.widget.button.TKToggleButton;

import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.image.BufferedImage;

/**
 * A toggle button that has a unique 'tab' look and is used to control sub-panel display in tabbed
 * panels.
 */
public class TKTab extends TKToggleButton {
	private boolean	mFirst;

	/**
	 * Creates a tab with the specified text, icon, selection state, and horizontal alignment.
	 * 
	 * @param text The title of the tab.
	 * @param icon An optional icon for the tab.
	 * @param selected Whether the tab is currently selected.
	 * @param alignment The text alignment for the tab title.
	 */
	public TKTab(String text, BufferedImage icon, boolean selected, int alignment) {
		super(text, icon, selected, alignment);
		setPushButtonMode(true);
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		Insets insets = getInsets();
		int x = insets.left;
		int y = insets.top;
		int w = getWidth() - 1;
		int h = getHeight();

		if (isSelected()) {
			g2d.setColor(TKColor.CONTROL_FILL);
		} else {
			g2d.setColor(TKColor.CONTROL_PRESSED_FILL);
			y += 1;
		}

		g2d.fillRect(x, y, w, h);
		g2d.setColor(TKColor.CONTROL_LINE);
		g2d.drawLine(x, y, x, y + h);
		g2d.drawLine(x, y, x + w, y);
		g2d.drawLine(x + w, y, x + w, y + h);

		g2d.setColor(TKColor.CONTROL_HIGHLIGHT);
		g2d.drawLine(x + 1, y + 1, x + w - 1, y + 1);

		if (isSelected() || isFirst()) {
			g2d.drawLine(x + 1, y + 1, x + 1, y + h);
		}

		if (isSelected()) {
			g2d.setColor(TKColor.CONTROL_SHADOW);
			g2d.drawLine(x + w - 1, y + 1, x + w - 1, y + h);
		}

		drawLabel(g2d, clips);
	}

	/** @return <code>true</code>, if this tab has been marked as the first in a tab group. */
	public boolean isFirst() {
		return mFirst;
	}

	/**
	 * Sets a flag indicating whether this tab begins a tab group.
	 * 
	 * @param first the new value for the first flag.
	 */
	public void setFirst(boolean first) {
		mFirst = first;
	}

	@Override protected int getAvailableTextWidth() {
		return super.getAvailableTextWidth() - 4;
	}

	@Override public String getToolTipText() {
		return getText();
	}
}
