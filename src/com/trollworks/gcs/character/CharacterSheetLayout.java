/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;

/** A layout for the character sheet that dynamically does n-up presentation. */
class CharacterSheetLayout implements LayoutManager2 {
	private static final int	MARGIN	= 1;

	@Override
	public Dimension minimumLayoutSize(Container target) {
		return preferredLayoutSize(target);
	}

	@Override
	public Dimension preferredLayoutSize(Container target) {
		Insets insets = target.getInsets();
		Component[] children = target.getComponents();
		int across = 1;
		int width = 0;
		int height = 0;

		if (children.length > 0) {
			Dimension size = children[0].getPreferredSize();
			Container parent = target.getParent();
			if (parent != null) {
				Insets parentInsets = parent.getInsets();
				int avail = parent.getWidth() - (parentInsets.left + parentInsets.right);
				int pageWidth = size.width;
				avail -= insets.left + insets.right + pageWidth;
				pageWidth += MARGIN;
				while (true) {
					avail -= pageWidth;
					if (avail >= 0) {
						across++;
					} else {
						break;
					}
				}
			}
			width = (size.width + MARGIN) * across - MARGIN;
			int pagesDown = children.length / across;
			if (children.length % across != 0) {
				pagesDown++;
			}
			height = (size.height + MARGIN) * pagesDown - MARGIN;
		}
		return new Dimension(insets.left + insets.right + width, insets.top + insets.bottom + height);
	}

	@Override
	public Dimension maximumLayoutSize(Container target) {
		return preferredLayoutSize(target);
	}

	@Override
	public void layoutContainer(Container target) {
		Component[] children = target.getComponents();
		if (children.length > 0) {
			Dimension size = children[0].getPreferredSize();
			Dimension avail = target.getSize();
			Insets insets = target.getInsets();
			int x = insets.left;
			int y = insets.top;
			for (Component child : children) {
				child.setBounds(x, y, size.width, size.height);
				x += size.width + MARGIN;
				if (x + size.width + insets.right > avail.width) {
					x = insets.left;
					y += size.height + MARGIN;
				}
			}
		}
	}

	@Override
	public float getLayoutAlignmentX(Container target) {
		return Component.LEFT_ALIGNMENT;
	}

	@Override
	public float getLayoutAlignmentY(Container target) {
		return Component.TOP_ALIGNMENT;
	}

	@Override
	public void addLayoutComponent(Component comp, Object constraints) {
		// Not used.
	}

	@Override
	public void addLayoutComponent(String name, Component comp) {
		// Not used.
	}

	@Override
	public void removeLayoutComponent(Component comp) {
		// Not used.
	}

	@Override
	public void invalidateLayout(Container target) {
		// Not used.
	}
}
