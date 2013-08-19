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

package com.trollworks.gcs.ui.editor;

import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.scroll.TKScrollable;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Rectangle;

/** A simple panel that draws banded colors behinds its contents. */
public class CSBandedPanel extends TKPanel implements TKScrollable {
	private String	mTitle;

	/**
	 * Creates a new {@link CSBandedPanel}.
	 * 
	 * @param title The title for this panel.
	 */
	public CSBandedPanel(String title) {
		super(new TKColumnLayout());
		setOpaque(true);
		setBackground(Color.white);
		setBorder(new TKEmptyBorder(5, 0, 5, 5));
		mTitle = title;
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		int count = getComponentCount();
		Rectangle bounds = getLocalBounds();

		for (int i = 0; i < count; i++) {
			Rectangle compBounds = getComponent(i).getBounds();

			bounds.y = compBounds.y;
			bounds.height = compBounds.height;
			g2d.setColor(i % 2 == 0 ? TKColor.PRIMARY_BANDING : TKColor.SECONDARY_BANDING);
			g2d.fill(bounds);
		}
	}

	public int getBlockScrollIncrement(Rectangle visibleBounds, boolean vertical, boolean upLeftDirection) {
		return (upLeftDirection ? -1 : 1) * (vertical ? visibleBounds.height : visibleBounds.width);
	}

	public Dimension getPreferredViewportSize() {
		return getPreferredSize();
	}

	public int getUnitScrollIncrement(Rectangle visibleBounds, boolean vertical, boolean upLeftDirection) {
		return upLeftDirection ? -10 : 10;
	}

	public boolean shouldTrackViewportHeight() {
		return TKScrollPanel.shouldTrackViewportHeight(this);
	}

	public boolean shouldTrackViewportWidth() {
		return TKScrollPanel.shouldTrackViewportWidth(this);
	}

	@Override public String toString() {
		return mTitle;
	}
}
