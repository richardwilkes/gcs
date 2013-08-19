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

import com.trollworks.toolkit.utility.TKColor;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.LayoutManager2;
import java.awt.Rectangle;

/** Provides a framework for a widget embedded in the border. */
public class TKWidgetBorderPanel extends TKPanel implements LayoutManager2 {
	private static final int	GAP	= 3;
	private TKPanel				mBorderPanel;
	private TKPanel				mContentPanel;

	/** Creates a new widget border panel. */
	public TKWidgetBorderPanel() {
		super();
		setLayout(this);
	}

	/**
	 * Creates a new widget border panel.
	 * 
	 * @param borderPanel The widget panel.
	 * @param contentPanel The content panel.
	 */
	public TKWidgetBorderPanel(TKPanel borderPanel, TKPanel contentPanel) {
		this();
		setBorderPanel(borderPanel);
		setContentPanel(contentPanel);
	}

	/** @return The border panel. */
	public TKPanel getBorderPanel() {
		return mBorderPanel;
	}

	/** @param borderPanel The panel to set as the new border panel. */
	public void setBorderPanel(TKPanel borderPanel) {
		if (borderPanel != mBorderPanel) {
			if (mBorderPanel != null) {
				remove(mBorderPanel);
			}
			mBorderPanel = borderPanel;
			if (mBorderPanel != null) {
				add(mBorderPanel);
			}
		}
	}

	/** @return The content panel. */
	public TKPanel getContentPanel() {
		return mContentPanel;
	}

	/** @param contentPanel The panel to set as the new content panel. */
	public void setContentPanel(TKPanel contentPanel) {
		if (mContentPanel != contentPanel) {
			if (mContentPanel != null) {
				remove(mContentPanel);
			}
			mContentPanel = contentPanel;
			if (mContentPanel != null) {
				add(mContentPanel);
			}
		}
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		Rectangle bounds = getLocalInsetBounds();
		int left = bounds.x;
		int top = bounds.y + mBorderPanel.getHeight() / 2;
		int width = bounds.width - 1;
		int height = bounds.height - 1;

		g2d.setColor(TKColor.darker(getBackground(), 45));
		g2d.drawLine(left + 1, top, left + 1 + GAP, top);
		g2d.drawLine(left + 1 + GAP * 3 + mBorderPanel.getWidth(), top, width - 1, top);
		g2d.drawLine(left, top + 1, left, bounds.y + height - 1);
		g2d.drawLine(left + width, top + 1, left + width, bounds.y + height - 1);
		g2d.drawLine(left + 1, bounds.y + height, left + width - 1, bounds.y + height);
	}

	public void layoutContainer(Container parent) {
		Dimension size = mBorderPanel.getPreferredSize();
		Insets insets = getInsets();

		mBorderPanel.setBounds(insets.left + 1 + GAP * 2, insets.top, size.width, size.height);
		mContentPanel.setBounds(insets.left + 2, insets.top + size.height + 1, parent.getWidth() - (4 + insets.left + insets.right), parent.getHeight() - (size.height + 3 + insets.top + insets.bottom));
	}

	public Dimension minimumLayoutSize(Container parent) {
		return layoutSize(mContentPanel.getMinimumSize());
	}

	public Dimension maximumLayoutSize(Container target) {
		return layoutSize(mContentPanel.getMaximumSize());
	}

	public Dimension preferredLayoutSize(Container parent) {
		return layoutSize(mContentPanel.getPreferredSize());
	}

	private Dimension layoutSize(Dimension size) {
		Dimension borderSize = mBorderPanel.getPreferredSize();
		Insets insets = getInsets();
		int tmp = borderSize.width + GAP * 3;

		if (tmp > size.width) {
			size.width = tmp;
		}
		size.width += 4 + insets.left + insets.right;
		size.height += borderSize.height + 3 + insets.top + insets.bottom;
		return sanitizeSize(size);
	}

	public float getLayoutAlignmentX(Container target) {
		return Component.CENTER_ALIGNMENT;
	}

	public float getLayoutAlignmentY(Container target) {
		return Component.CENTER_ALIGNMENT;
	}

	public void addLayoutComponent(String name, Component comp) {
		// Nothing to do...
	}

	public void invalidateLayout(Container target) {
		// Nothing to do...
	}

	public void addLayoutComponent(Component comp, Object constraints) {
		// Nothing to do...
	}

	public void removeLayoutComponent(Component comp) {
		// Nothing to do...
	}
}
