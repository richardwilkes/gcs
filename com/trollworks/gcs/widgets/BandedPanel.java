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

package com.trollworks.gcs.widgets;

import com.trollworks.gcs.utility.Colors;
import com.trollworks.gcs.widgets.layout.ColumnLayout;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Rectangle;

import javax.swing.Scrollable;
import javax.swing.SwingConstants;

/** A simple panel that draws banded colors behind its contents. */
public class BandedPanel extends ActionPanel implements Scrollable {
	private String	mTitle;

	/**
	 * Creates a new {@link BandedPanel}.
	 * 
	 * @param title The title for this panel.
	 */
	public BandedPanel(String title) {
		super(new ColumnLayout(1, 0, 0));
		setOpaque(true);
		setBackground(Color.white);
		mTitle = title;
	}

	@Override protected void paintComponent(Graphics gc) {
		super.paintComponent(GraphicsUtilities.prepare(gc));

		int count = getComponentCount();
		Rectangle bounds = getBounds();
		bounds.x = 0;
		bounds.y = 0;

		for (int i = 0; i < count; i++) {
			Rectangle compBounds = getComponent(i).getBounds();
			bounds.y = compBounds.y;
			bounds.height = compBounds.height;
			gc.setColor(Colors.getBanding(i % 2 == 0));
			gc.fillRect(bounds.x, bounds.y, bounds.width, bounds.height);
		}
	}

	public int getScrollableBlockIncrement(Rectangle visibleRect, int orientation, int direction) {
		return orientation == SwingConstants.VERTICAL ? visibleRect.height : visibleRect.width;
	}

	public Dimension getPreferredScrollableViewportSize() {
		return getPreferredSize();
	}

	public int getScrollableUnitIncrement(Rectangle visibleRect, int orientation, int direction) {
		return 10;
	}

	public boolean getScrollableTracksViewportHeight() {
		return UIUtilities.shouldTrackViewportHeight(this);
	}

	public boolean getScrollableTracksViewportWidth() {
		return UIUtilities.shouldTrackViewportWidth(this);
	}

	@Override public String toString() {
		return mTitle;
	}
}
