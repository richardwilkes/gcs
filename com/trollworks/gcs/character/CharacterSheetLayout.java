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
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

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
