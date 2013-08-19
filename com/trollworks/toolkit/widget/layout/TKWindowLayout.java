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

package com.trollworks.toolkit.widget.layout;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;

import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKToolBar;
import com.trollworks.toolkit.widget.menu.TKMenuBar;
import com.trollworks.toolkit.window.TKBaseWindow;

/** The standard window layout manager. */
public class TKWindowLayout implements LayoutManager2 {
	private static TKWindowLayout	INSTANCE	= null;

	private TKWindowLayout() {
		// Prevents any creation except through getInstance()
	}

	/** @return The one instance of the window layout manager that can exist. */
	public static final TKWindowLayout getInstance() {
		if (INSTANCE == null) {
			INSTANCE = new TKWindowLayout();
		}
		return INSTANCE;
	}

	public void addLayoutComponent(Component comp, Object constraints) {
		// Nothing to do...
	}

	public Dimension maximumLayoutSize(Container parent) {
		Dimension size = new Dimension(Short.MAX_VALUE, Short.MAX_VALUE);

		if (parent instanceof TKBaseWindow) {
			TKBaseWindow base = (TKBaseWindow) parent;
			TKMenuBar menuBar = base.getTKMenuBar();
			TKToolBar toolBar = base.getTKToolBar();
			TKPanel content = base.getContent();
			Dimension tmp = content.getMaximumSize();
			Insets insets = parent.getInsets();

			size.width = tmp.width;
			size.height = tmp.height;

			if (menuBar != null) {
				tmp = menuBar.getMaximumSize();
				size.height += tmp.height;
				if (tmp.width > size.width) {
					size.width = tmp.width;
				}
			}
			if (toolBar != null) {
				tmp = toolBar.getMaximumSize();
				size.height += tmp.height;
				if (tmp.width > size.width) {
					size.width = tmp.width;
				}
			}

			size.width += insets.left + insets.right;
			size.height += insets.top + insets.bottom;
		}
		return size;
	}

	public float getLayoutAlignmentX(Container target) {
		return Component.CENTER_ALIGNMENT;
	}

	public float getLayoutAlignmentY(Container target) {
		return Component.CENTER_ALIGNMENT;
	}

	public void invalidateLayout(Container target) {
		// Nothing to do...
	}

	public void addLayoutComponent(String name, Component comp) {
		// Nothing to do...
	}

	public void removeLayoutComponent(Component comp) {
		// Nothing to do...
	}

	public Dimension preferredLayoutSize(Container parent) {
		Dimension size = new Dimension(Short.MAX_VALUE, Short.MAX_VALUE);

		if (parent instanceof TKBaseWindow) {
			TKBaseWindow base = (TKBaseWindow) parent;
			TKMenuBar menuBar = base.getTKMenuBar();
			TKToolBar toolBar = base.getTKToolBar();
			TKPanel content = base.getContent();
			Dimension tmp = content.getPreferredSize();
			Insets insets = parent.getInsets();

			size.width = tmp.width;
			size.height = tmp.height;

			if (menuBar != null) {
				tmp = menuBar.getPreferredSize();
				size.height += tmp.height;
				if (tmp.width > size.width) {
					size.width = tmp.width;
				}
			}
			if (toolBar != null) {
				tmp = toolBar.getPreferredSize();
				size.height += tmp.height;
				if (tmp.width > size.width) {
					size.width = tmp.width;
				}
			}

			size.width += insets.left + insets.right;
			size.height += insets.top + insets.bottom;
		}
		return size;
	}

	public Dimension minimumLayoutSize(Container parent) {
		Dimension size = new Dimension();

		if (parent instanceof TKBaseWindow) {
			TKBaseWindow base = (TKBaseWindow) parent;
			TKMenuBar menuBar = base.getTKMenuBar();
			TKToolBar toolBar = base.getTKToolBar();
			TKPanel content = base.getContent();
			Dimension tmp = content.getMinimumSize();
			Insets insets = parent.getInsets();

			size.width = tmp.width;
			size.height = tmp.height;

			if (menuBar != null) {
				tmp = menuBar.getMinimumSize();
				size.height += tmp.height;
				if (tmp.width > size.width) {
					size.width = tmp.width;
				}
			}
			if (toolBar != null) {
				tmp = toolBar.getMinimumSize();
				size.height += tmp.height;
				if (tmp.width > size.width) {
					size.width = tmp.width;
				}
			}

			size.width += insets.left + insets.right;
			size.height += insets.top + insets.bottom;
		}
		return size;
	}

	public void layoutContainer(Container parent) {
		if (parent instanceof TKBaseWindow) {
			TKBaseWindow base = (TKBaseWindow) parent;
			TKMenuBar menuBar = base.getTKMenuBar();
			TKToolBar toolBar = base.getTKToolBar();
			TKPanel content = base.getContent();
			Insets insets = parent.getInsets();
			Dimension size = parent.getSize();
			int width = size.width - (insets.left + insets.right);
			int y = insets.top;
			int height;

			if (menuBar != null) {
				height = menuBar.getPreferredSize().height;
				menuBar.setBounds(insets.left, y, width, height);
				y += height;
			}

			if (toolBar != null) {
				height = toolBar.getPreferredSize().height;
				toolBar.setBounds(insets.left, y, width, height);
				y += height;
			}

			height = size.height - (y + insets.bottom);
			if (height < 0) {
				height = 0;
			}
			content.setBounds(insets.left, y, width, height);
		}
	}
}
