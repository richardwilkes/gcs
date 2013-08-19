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
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.character;

import com.trollworks.gcs.widgets.outline.ColumnUtils;
import com.trollworks.ttk.border.BoxedDropShadowBorder;
import com.trollworks.ttk.widgets.outline.Column;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.OutlineHeader;
import com.trollworks.ttk.widgets.outline.OutlineModel;
import com.trollworks.ttk.widgets.outline.OutlineProxy;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;
import java.awt.Rectangle;

/** An outline panel. */
public class SingleOutlinePanel extends DropPanel implements LayoutManager2 {
	private OutlineHeader	mHeader;
	private Outline			mOutline;

	/**
	 * Creates a new outline panel.
	 * 
	 * @param outline The outline to display.
	 * @param title The localized title for the panel.
	 * @param useProxy <code>true</code> if a proxy of the outline should be used.
	 */
	public SingleOutlinePanel(Outline outline, String title, boolean useProxy) {
		super(null, title, false);
		mOutline = useProxy ? new OutlineProxy(outline) : outline;
		mHeader = mOutline.getHeaderPanel();
		CharacterSheet.prepOutline(mOutline);
		add(mHeader);
		add(mOutline);
		setBorder(new BoxedDropShadowBorder());
		setLayout(this);
	}

	/**
	 * Sets the embedded outline's display range.
	 * 
	 * @param first The first row to display.
	 * @param last The last row to display.
	 */
	public void setOutlineRowRange(int first, int last) {
		mOutline.setFirstRowToDisplay(first);
		mOutline.setLastRowToDisplay(last);
	}

	/** @return The preferred width. */
	public int getPreferredWidth() {
		Insets insets = getInsets();
		int width = insets.left + insets.right;
		OutlineModel outlineModel = mOutline.getModel();
		int count = outlineModel.getColumnCount();
		if (mOutline.shouldDrawColumnDividers()) {
			width += count - 1;
		}
		for (int i = 0; i < count; i++) {
			Column column = outlineModel.getColumnAtIndex(i);
			width += column.getPreferredWidth(mOutline);
		}
		return width;
	}

	@Override
	public void layoutContainer(Container parent) {
		Insets insets = getInsets();
		Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
		int height = mHeader.getPreferredSize().height;
		mHeader.setBounds(bounds.x, bounds.y, bounds.width, height);
		bounds.y += height;
		bounds.height -= height;
		mOutline.setBounds(bounds.x, bounds.y, bounds.width, bounds.height);
		ColumnUtils.pack(mOutline, bounds.width);
	}

	@Override
	public Dimension minimumLayoutSize(Container parent) {
		return getLayoutSize(mOutline.getMinimumSize());
	}

	@Override
	public Dimension preferredLayoutSize(Container parent) {
		return getLayoutSize(mOutline.getPreferredSize());
	}

	@Override
	public Dimension maximumLayoutSize(Container target) {
		return getLayoutSize(mOutline.getMaximumSize());
	}

	private Dimension getLayoutSize(Dimension size) {
		Insets insets = getInsets();
		return new Dimension(insets.left + insets.right + size.width, insets.top + insets.bottom + size.height + mHeader.getPreferredSize().height);
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
	public void addLayoutComponent(String name, Component comp) {
		// Nothing to do...
	}

	@Override
	public void addLayoutComponent(Component comp, Object constraints) {
		// Nothing to do...
	}

	@Override
	public void removeLayoutComponent(Component comp) {
		// Nothing to do...
	}
}
