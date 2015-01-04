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

import com.trollworks.toolkit.ui.widget.outline.Outline;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;
import java.awt.Rectangle;

import javax.swing.JPanel;

/** A panel that holds a pair of side-by-side outlines. */
public class DoubleOutlinePanel extends JPanel implements LayoutManager2 {
	private SingleOutlinePanel	mLeftPanel;
	private SingleOutlinePanel	mRightPanel;

	/**
	 * Creates a new double outline panel.
	 * 
	 * @param leftOutline The outline to display on the left.
	 * @param leftTitle The localized title for the left panel.
	 * @param rightOutline The outline to display on the right.
	 * @param rightTitle The localized title for the right panel.
	 * @param useProxy <code>true</code> if a proxy of the outlines should be used.
	 */
	public DoubleOutlinePanel(Outline leftOutline, String leftTitle, Outline rightOutline, String rightTitle, boolean useProxy) {
		super();
		setLayout(this);
		mLeftPanel = new SingleOutlinePanel(leftOutline, leftTitle, useProxy);
		mRightPanel = new SingleOutlinePanel(rightOutline, rightTitle, useProxy);
		add(mLeftPanel);
		add(mRightPanel);
	}

	/**
	 * Sets the embedded outline's display range.
	 * 
	 * @param forRight <code>true</code> to set the right outline.
	 * @param first The first row to display.
	 * @param last The last row to display.
	 */
	public void setOutlineRowRange(boolean forRight, int first, int last) {
		(forRight ? mRightPanel : mLeftPanel).setOutlineRowRange(first, last);
	}

	@Override
	public void addLayoutComponent(Component comp, Object constraints) {
		// Not used.
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
	public Dimension preferredLayoutSize(Container parent) {
		return getLayoutSize(mLeftPanel.getPreferredSize(), mRightPanel.getPreferredSize());
	}

	@Override
	public Dimension minimumLayoutSize(Container parent) {
		return getLayoutSize(mLeftPanel.getMinimumSize(), mRightPanel.getMinimumSize());
	}

	@Override
	public Dimension maximumLayoutSize(Container target) {
		return getLayoutSize(mLeftPanel.getMaximumSize(), mRightPanel.getMaximumSize());
	}

	@Override
	public void layoutContainer(Container parent) {
		Insets insets = getInsets();
		Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
		int left = mLeftPanel.getPreferredWidth();
		int right = mRightPanel.getPreferredWidth();

		if (left + right != bounds.width - 2) {
			int half = (bounds.width - 2) / 2;

			if (left < half && right < half) {
				right = half;
				left = bounds.width - (half + 2);
			} else if (left < half) {
				right = bounds.width - (left + 2);
			} else if (right < half) {
				left = bounds.width - (right + 2);
			} else {
				right = half;
				left = bounds.width - (half + 2);
			}
		}
		mLeftPanel.setBounds(bounds.x, bounds.y, left, bounds.height);
		mRightPanel.setBounds(bounds.x + left + 2, bounds.y, right, bounds.height);
	}

	private Dimension getLayoutSize(Dimension leftSize, Dimension rightSize) {
		Dimension size = new Dimension(leftSize);
		Insets insets = getInsets();

		if (leftSize.height < rightSize.height) {
			size.height = rightSize.height;
		}
		size.width += rightSize.width + 2 + insets.left + insets.right;
		size.height += insets.top + insets.bottom;
		return size;
	}
}
