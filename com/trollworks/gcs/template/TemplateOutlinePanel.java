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

package com.trollworks.gcs.template;

import com.trollworks.gcs.app.GCSFonts;
import com.trollworks.gcs.character.DropPanel;
import com.trollworks.ttk.border.BoxedDropShadowBorder;
import com.trollworks.ttk.text.TextDrawing;
import com.trollworks.ttk.widgets.outline.Column;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.OutlineHeader;
import com.trollworks.ttk.widgets.outline.OutlineModel;

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
		Insets insets = getInsets();
		Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
		int width = bounds.width;
		int height = mHeader.getPreferredSize().height;
		OutlineModel outlineModel = mOutline.getModel();
		int count = outlineModel.getColumnCount();
		ArrayList<Column> changed = new ArrayList<Column>();
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

	public void addLayoutComponent(String name, Component comp) {
		// Nothing to do...
	}

	public Dimension minimumLayoutSize(Container parent) {
		Dimension size = mOutline.getMinimumSize();
		int minHeight = TextDrawing.getPreferredSize(UIManager.getFont(GCSFonts.KEY_FIELD), "Mg").height; //$NON-NLS-1$
		if (size.height < minHeight) {
			size.height = minHeight;
		}
		return getLayoutSizeForOne(size);
	}

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
