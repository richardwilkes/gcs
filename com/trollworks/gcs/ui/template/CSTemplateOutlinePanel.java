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

package com.trollworks.gcs.ui.template;

import com.trollworks.gcs.ui.common.CSDropPanel;
import com.trollworks.gcs.ui.common.CSFont;
import com.trollworks.toolkit.text.TKTextDrawing;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.border.TKBoxedDropShadowBorder;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKOutlineHeader;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;
import java.awt.Rectangle;
import java.util.ArrayList;

/** The template outline panel. */
public class CSTemplateOutlinePanel extends CSDropPanel implements LayoutManager2 {
	private TKOutlineHeader	mHeader;
	private TKOutline		mOutline;

	/**
	 * Creates a new template outline panel.
	 * 
	 * @param outline The outline to display.
	 * @param title The localized title for the panel.
	 */
	public CSTemplateOutlinePanel(TKOutline outline, String title) {
		super(null, title, false);
		mOutline = outline;
		mHeader = mOutline.getHeaderPanel();
		CSTemplate.prepOutline(mOutline);
		add(mHeader);
		add(mOutline);
		setBorder(new TKBoxedDropShadowBorder());
		setLayout(this);
	}

	public float getLayoutAlignmentX(Container target) {
		return CENTER_ALIGNMENT;
	}

	public float getLayoutAlignmentY(Container target) {
		return CENTER_ALIGNMENT;
	}

	public void invalidateLayout(Container target) {
		// Nothing to do...
	}

	public Dimension maximumLayoutSize(Container target) {
		return getLayoutSizeForOne(mOutline.getMaximumSize());
	}

	public void addLayoutComponent(Component comp, Object constraints) {
		// Nothing to do...
	}

	public void removeLayoutComponent(Component comp) {
		// Nothing to do...
	}

	public void layoutContainer(Container parent) {
		Rectangle bounds = getLocalInsetBounds();
		int width = bounds.width;
		int height = mHeader.getPreferredSize().height;
		TKOutlineModel outlineModel = mOutline.getModel();
		int count = outlineModel.getColumnCount();
		ArrayList<TKColumn> changed = new ArrayList<TKColumn>();
		TKColumn column;

		mHeader.setBounds(bounds.x, bounds.y, width, height);
		bounds.y += height;
		bounds.height -= height;
		mOutline.setBounds(bounds.x, bounds.y, width, bounds.height);

		for (int i = 0; i < count; i++) {
			column = outlineModel.getColumnAtIndex(i);
			if (column.getID() != 0) {
				int prefWidth = column.getPreferredWidth(mOutline);

				if (prefWidth != column.getWidth()) {
					column.setWidth(prefWidth);
					changed.add(column);
				}
				width -= prefWidth;
			}
		}

		if (mOutline.shouldDrawColumnDividers()) {
			width -= count - 1;
		}
		column = outlineModel.getColumnWithID(0);
		if (column.getWidth() != width) {
			column.setWidth(width);
			changed.add(column);
		}
		mOutline.updateRowHeightsIfNeeded(changed);
		mOutline.revalidateView();
	}

	public void addLayoutComponent(String name, Component comp) {
		// Nothing to do...
	}

	public Dimension minimumLayoutSize(Container parent) {
		Dimension size = mOutline.getMinimumSize();
		int minHeight = TKTextDrawing.getPreferredSize(TKFont.lookup(CSFont.KEY_FIELD), null, "Mg").height; //$NON-NLS-1$

		if (size.height < minHeight) {
			size.height = minHeight;
		}
		return getLayoutSizeForOne(size);
	}

	public Dimension preferredLayoutSize(Container parent) {
		return getLayoutSizeForOne(mOutline.getPreferredSize());
	}

	private Dimension getLayoutSizeForOne(Dimension one) {
		Insets insets = getInsets();

		return new Dimension(1 + insets.left + insets.right + one.width, insets.top + insets.bottom + one.height + mHeader.getPreferredSize().height);
	}
}
