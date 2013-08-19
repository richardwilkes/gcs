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

package com.trollworks.gcs.ui.common;

import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKBoxedDropShadowBorder;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.LayoutManager;
import java.awt.Rectangle;
import java.util.HashMap;
import java.util.Map.Entry;

/** A standard panel with a drop shadow. */
public class CSDropPanel extends TKPanel {
	private boolean					mOnlyReportPreferredSize;
	private HashMap<TKPanel, Color>	mHorizontalBackgrounds;
	private HashMap<TKPanel, Color>	mVerticalBackgrounds;
	private boolean					mPaintVerticalFirst;
	private TKBoxedDropShadowBorder	mBoxedDropShadowBorder;

	/**
	 * Creates a standard panel with a drop shadow.
	 * 
	 * @param layout The layout to use.
	 */
	public CSDropPanel(LayoutManager layout) {
		this(layout, false);
	}

	/**
	 * Creates a standard panel with a drop shadow.
	 * 
	 * @param layout The layout to use.
	 * @param onlyReportPreferredSize Whether or not minimum and maximum size is reported as
	 *            preferred size or not.
	 */
	public CSDropPanel(LayoutManager layout, boolean onlyReportPreferredSize) {
		this(layout, null, null, onlyReportPreferredSize);
	}

	/**
	 * Creates a standard panel with a drop shadow.
	 * 
	 * @param layout The layout to use.
	 * @param title The title to use.
	 */
	public CSDropPanel(LayoutManager layout, String title) {
		this(layout, title, TKFont.lookup(CSFont.KEY_LABEL), false);
	}

	/**
	 * Creates a standard panel with a drop shadow.
	 * 
	 * @param layout The layout to use.
	 * @param title The title to use.
	 * @param onlyReportPreferredSize Whether or not minimum and maximum size is reported as
	 *            preferred size or not.
	 */
	public CSDropPanel(LayoutManager layout, String title, boolean onlyReportPreferredSize) {
		this(layout, title, TKFont.lookup(CSFont.KEY_LABEL), onlyReportPreferredSize);
	}

	/**
	 * Creates a standard panel with a drop shadow.
	 * 
	 * @param layout The layout to use.
	 * @param title The title to use.
	 * @param font The font to use for the title.
	 * @param onlyReportPreferredSize Whether or not minimum and maximum size is reported as
	 *            preferred size or not.
	 */
	public CSDropPanel(LayoutManager layout, String title, Font font, boolean onlyReportPreferredSize) {
		super(layout);
		mBoxedDropShadowBorder = new TKBoxedDropShadowBorder(font, title);
		setBorder(new TKCompoundBorder(mBoxedDropShadowBorder, new TKEmptyBorder(0, 2, 1, 2)));
		setAlignmentY(TOP_ALIGNMENT);
		mOnlyReportPreferredSize = onlyReportPreferredSize;
		mHorizontalBackgrounds = new HashMap<TKPanel, Color>();
		mVerticalBackgrounds = new HashMap<TKPanel, Color>();
	}

	@Override protected Dimension getMinimumSizeSelf() {
		return mOnlyReportPreferredSize ? getPreferredSize() : super.getMinimumSizeSelf();
	}

	@Override protected Dimension getMaximumSizeSelf() {
		return mOnlyReportPreferredSize ? getPreferredSize() : super.getMaximumSizeSelf();
	}

	/**
	 * Marks an area with a specific background color. The panel specified will be used to calculate
	 * the area's top and bottom, and the background color will span the width of the drop panel.
	 * 
	 * @param panel The panel to attach the color to.
	 * @param background The color to attach.
	 */
	public void addHorizontalBackground(TKPanel panel, Color background) {
		mHorizontalBackgrounds.put(panel, background);
	}

	/**
	 * Removes a horizontal background added with {@link #addHorizontalBackground(TKPanel,Color)}.
	 * 
	 * @param panel The panel to remove.
	 */
	public void removeHorizontalBackground(TKPanel panel) {
		mHorizontalBackgrounds.remove(panel);
	}

	/**
	 * Marks an area with a specific background color. The panel specified will be used to calculate
	 * the area's left and right, and the background color will span the height of the drop panel.
	 * 
	 * @param panel The panel to attach the color to.
	 * @param background The color to attach.
	 */
	public void addVerticalBackground(TKPanel panel, Color background) {
		mVerticalBackgrounds.put(panel, background);
	}

	/**
	 * Removes a vertical background added with {@link #addVerticalBackground(TKPanel,Color)}.
	 * 
	 * @param panel The panel to remove.
	 */
	public void removeVerticalBackground(TKPanel panel) {
		mVerticalBackgrounds.remove(panel);
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		Insets insets = mBoxedDropShadowBorder.getBorderInsets(this);
		Rectangle localBounds = getLocalBounds();

		localBounds.x += insets.left;
		localBounds.y += insets.top;
		localBounds.width -= insets.left + insets.right;
		localBounds.height -= insets.top + insets.bottom;
		if (mPaintVerticalFirst) {
			paintVerticalBackgrounds(g2d, localBounds);
			paintHorizontalBackgrounds(g2d, localBounds);
		} else {
			paintHorizontalBackgrounds(g2d, localBounds);
			paintVerticalBackgrounds(g2d, localBounds);
		}
	}

	private void paintHorizontalBackgrounds(Graphics2D g2d, Rectangle localBounds) {
		for (Entry<TKPanel, Color> entry : mHorizontalBackgrounds.entrySet()) {
			TKPanel panel = entry.getKey();
			Rectangle bounds = panel.getBounds();
			Container parent = panel.getParent();

			if (parent != null) {
				if (parent != this) {
					convertRectangle(bounds, parent, this);
				}
				bounds.x = localBounds.x;
				bounds.width = localBounds.width;
				g2d.setColor(entry.getValue());
				g2d.fill(bounds);
			}
		}
	}

	private void paintVerticalBackgrounds(Graphics2D g2d, Rectangle localBounds) {
		for (Entry<TKPanel, Color> entry : mVerticalBackgrounds.entrySet()) {
			TKPanel panel = entry.getKey();
			Rectangle bounds = panel.getBounds();
			Container parent = panel.getParent();

			if (parent != null) {
				if (parent != this) {
					convertRectangle(bounds, parent, this);
				}
				bounds.y = localBounds.y;
				bounds.height = localBounds.height;
				g2d.setColor(entry.getValue());
				g2d.fill(bounds);
			}
		}
	}

	/** @return Whether or not to paint the vertical backgrounds first. */
	public final boolean isPaintVerticalFirst() {
		return mPaintVerticalFirst;
	}

	/** @param first Whether or not to paint the vertical backgrounds first. */
	public final void setPaintVerticalFirst(boolean first) {
		mPaintVerticalFirst = first;
	}

	/** @return The {@link TKBoxedDropShadowBorder}. */
	public TKBoxedDropShadowBorder getBoxedDropShadowBorder() {
		return mBoxedDropShadowBorder;
	}
}
