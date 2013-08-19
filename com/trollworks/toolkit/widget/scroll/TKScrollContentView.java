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

package com.trollworks.toolkit.widget.scroll;

import com.trollworks.toolkit.widget.TKPanel;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.LayoutManager;
import java.awt.Rectangle;
import java.util.ArrayList;

/** Provides a viewport onto the actual scrolling content in a scroll panel. */
public class TKScrollContentView extends TKPanel implements LayoutManager {
	private ArrayList<TKScrollContentView>	mSlaves	= new ArrayList<TKScrollContentView>();
	private TKPanel							mContent;

	/** Creates a new scrolling content view. */
	public TKScrollContentView() {
		super();
		setLayout(this);
		setOpaque(true);
		setMinimumSize(new Dimension(10, 10));
	}

	/**
	 * Creates a new slave scrolling content view.
	 * 
	 * @param master The master view.
	 */
	public TKScrollContentView(TKScrollContentView master) {
		super();
		setLayout(this);
		setOpaque(true);
		setMinimumSize(new Dimension(10, 10));
		master.addSlave(this);
	}

	private void addSlave(TKScrollContentView slave) {
		mSlaves.add(slave);
	}

	/**
	 * @param local Pass in <code>true</code> to get the coordinates based on the local view.
	 * @return The scrolling bounds.
	 */
	public Rectangle getScrollingBounds(boolean local) {
		Rectangle bounds = local ? getLocalBounds() : getBounds();

		if (!local) {
			Container parent = getParent();

			while (parent != null && !(parent instanceof TKScrollBarOwner)) {
				bounds.x += parent.getX();
				bounds.y += parent.getY();
				parent = parent.getParent();
			}
		}
		return bounds;
	}

	/**
	 * Sets the content contained by this view.
	 * 
	 * @param content The content to set.
	 */
	void setContent(TKPanel content) {
		if (mContent != null) {
			remove(mContent);
		}
		mContent = content;
		if (content != null) {
			add(content);
		}
	}

	/** @return The content contained by this view. */
	public TKPanel getContent() {
		return mContent;
	}

	public void addLayoutComponent(String name, Component component) {
		// Nothing to do...
	}

	public void layoutContainer(Container container) {
		TKPanel content = getContent();
		Dimension size = content.getPreferredSize();

		if (content instanceof TKScrollable) {
			TKScrollable scrollable = (TKScrollable) content;

			if (scrollable.shouldTrackViewportHeight()) {
				size.height = getHeight();
			}
			if (scrollable.shouldTrackViewportWidth()) {
				size.width = getWidth();
			}
		}
		content.setSize(size);
	}

	public Dimension minimumLayoutSize(Container container) {
		return getMinimumSize();
	}

	public Dimension preferredLayoutSize(Container container) {
		TKPanel content = getContent();

		return content != null ? content.getPreferredSize() : new Dimension(MAX_SIZE, MAX_SIZE);
	}

	public void removeLayoutComponent(Component component) {
		// Nothing to do...
	}

	@Override public void setBounds(int x, int y, int width, int height) {
		super.setBounds(x, y, width, height);
		if (width != 0 && height != 0) {
			for (TKScrollContentView slave : mSlaves) {
				slave.repaint();
			}
			repaint();
		}
	}
}
