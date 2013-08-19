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

import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.scroll.TKScrollable;

import java.awt.Dimension;
import java.awt.LayoutManager;
import java.awt.Rectangle;

/** A simple panel that implements the {@link TKScrollable} interface. */
public class TKScrollablePanel extends TKPanel implements TKScrollable {
	private Dimension	mPreferredViewportSize;

	/** Creates a new {@link TKScrollablePanel}. */
	public TKScrollablePanel() {
		this(null);
	}

	/**
	 * Creates a new {@link TKScrollablePanel}.
	 * 
	 * @param layout The layout to use.
	 */
	public TKScrollablePanel(LayoutManager layout) {
		super(layout);
	}

	public int getBlockScrollIncrement(Rectangle visibleBounds, boolean vertical, boolean upLeftDirection) {
		int amt = vertical ? visibleBounds.height : visibleBounds.width;

		return upLeftDirection ? -amt : amt;
	}

	public int getUnitScrollIncrement(Rectangle visibleBounds, boolean vertical, boolean upLeftDirection) {
		int amt = (vertical ? visibleBounds.height : visibleBounds.width) / 10;

		if (amt < 1) {
			amt = 1;
		}
		return upLeftDirection ? -amt : amt;
	}

	/** @param size The preferred viewport size. */
	public void setPreferredViewportSize(Dimension size) {
		mPreferredViewportSize = size;
	}

	public Dimension getPreferredViewportSize() {
		return mPreferredViewportSize == null ? getPreferredSize() : mPreferredViewportSize;
	}

	public boolean shouldTrackViewportHeight() {
		return TKScrollPanel.shouldTrackViewportHeight(this);
	}

	public boolean shouldTrackViewportWidth() {
		return TKScrollPanel.shouldTrackViewportWidth(this);
	}

}
