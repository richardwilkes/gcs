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

package com.trollworks.gcs.template;

import com.trollworks.gcs.app.GCSFonts;
import com.trollworks.gcs.character.DropPanel;
import com.trollworks.toolkit.ui.TextDrawing;
import com.trollworks.toolkit.ui.border.BoxedDropShadowBorder;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineHeader;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;
import java.awt.Rectangle;
import java.util.ArrayList;

import javax.swing.UIManager;

/** The template outline panel. */
public class TemplateOutlinePanel extends DropPanel implements LayoutManager2 {
	private OutlineHeader	mHeader;
	private Outline			mOutline;

	/**
	 * Creates a new template outline panel.
	 * 
	 * @param outline The outline to display.
	 * @param title The localized title for the panel.
	 */
	public TemplateOutlinePanel(Outline outline, String title) {
		super(null, title, false);
		mOutline = outline;
		mHeader = mOutline.getHeaderPanel();
		TemplateSheet.prepOutline(mOutline);
		add(mHeader);
		add(mOutline);
		setBorder(new BoxedDropShadowBorder());
		setLayout(this);
	}

	@Override
	public float getLayoutAlignmentX(Container target) {
		return CENTER_ALIGNMENT;
	}

	@Override
	public float getLayoutAlignmentY(Container target) {
		return CENTER_ALIGNMENT;
	}

	@Override
	public void invalidateLayout(Container target) {
		// Nothing to do...
	}

	@Override
	public Dimension maximumLayoutSize(Container target) {
		return getLayoutSizeForOne(mOutline.getMaximumSize());
	}

	@Override
	public void addLayoutComponent(Component comp, Object constraints) {
		// Nothing to do...
	}

	@Override
	public void removeLayoutComponent(Component comp) {
		// Nothing to do...
	}

	@Override
	public void layoutContainer(Container parent) {
		Insets insets = getInsets();
		Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
		int width = bounds.width;
		int height = mHeader.getPreferredSize().height;
		OutlineModel outlineModel = mOutline.getModel();
		int count = outlineModel.getColumnCount();
		ArrayList<Column> changed = new ArrayList<>();
		Column column;

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
			width -= count;
		}
		column = outlineModel.getColumnWithID(0);
		if (column.getWidth() != width) {
			column.setWidth(width);
			changed.add(column);
		}
		mOutline.updateRowHeightsIfNeeded(changed);
		mOutline.revalidateView();
	}

	@Override
	public void addLayoutComponent(String name, Component comp) {
		// Nothing to do...
	}

	@Override
	public Dimension minimumLayoutSize(Container parent) {
		Dimension size = mOutline.getMinimumSize();
		int minHeight = TextDrawing.getPreferredSize(UIManager.getFont(GCSFonts.KEY_FIELD), "Mg").height; //$NON-NLS-1$
		if (size.height < minHeight) {
			size.height = minHeight;
		}
		return getLayoutSizeForOne(size);
	}

	@Override
	public Dimension preferredLayoutSize(Container parent) {
		Dimension size = getLayoutSizeForOne(mOutline.getPreferredSize());
		Dimension min = getMinimumSize();
		if (size.width < min.width) {
			size.width = min.width;
		}
		if (size.height < min.height) {
			size.height = min.height;
		}
		return size;
	}

	private Dimension getLayoutSizeForOne(Dimension one) {
		Insets insets = getInsets();
		return new Dimension(1 + insets.left + insets.right + one.width, insets.top + insets.bottom + one.height + mHeader.getPreferredSize().height);
	}
}
