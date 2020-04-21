/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.widget.outline.Outline;

import java.awt.Color;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;
import java.awt.Rectangle;
import javax.swing.JPanel;

/** A panel that holds a pair of side-by-side outlines. */
public class DoubleOutlinePanel extends JPanel implements LayoutManager2 {
    private SingleOutlinePanel mLeftPanel;
    private SingleOutlinePanel mRightPanel;

    /**
     * Creates a new double outline panel.
     *
     * @param leftOutline  The outline to display on the left.
     * @param leftTitle    The localized title for the left panel.
     * @param rightOutline The outline to display on the right.
     * @param rightTitle   The localized title for the right panel.
     * @param useProxy     {@code true} if a proxy of the outlines should be used.
     */
    public DoubleOutlinePanel(Scale scale, Outline leftOutline, String leftTitle, Outline rightOutline, String rightTitle, boolean useProxy) {
        setLayout(this);
        setBackground(Color.WHITE);
        mLeftPanel = new SingleOutlinePanel(scale, leftOutline, leftTitle, useProxy);
        mRightPanel = new SingleOutlinePanel(scale, rightOutline, rightTitle, useProxy);
        add(mLeftPanel);
        add(mRightPanel);
    }

    /**
     * Sets the embedded outline's display range.
     *
     * @param forRight {@code true} to set the right outline.
     * @param first    The first row to display.
     * @param last     The last row to display.
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
        return getLayoutSize(parent, mLeftPanel.getPreferredSize(), mRightPanel.getPreferredSize());
    }

    @Override
    public Dimension minimumLayoutSize(Container parent) {
        return getLayoutSize(parent, mLeftPanel.getMinimumSize(), mRightPanel.getMinimumSize());
    }

    @Override
    public Dimension maximumLayoutSize(Container parent) {
        return getLayoutSize(parent, mLeftPanel.getMaximumSize(), mRightPanel.getMaximumSize());
    }

    @Override
    public void layoutContainer(Container parent) {
        Insets    insets = getInsets();
        Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
        Scale     scale  = Scale.get(parent);
        int       gap    = scale.scale(2);
        int       width  = (bounds.width - gap) / 2;
        mLeftPanel.setBounds(bounds.x, bounds.y, width, bounds.height);
        mRightPanel.setBounds(bounds.x + bounds.width - width, bounds.y, width, bounds.height);
    }

    private Dimension getLayoutSize(Container parent, Dimension leftSize, Dimension rightSize) {
        Dimension size   = new Dimension(leftSize.width + rightSize.width, Math.max(leftSize.height, rightSize.height));
        Insets    insets = getInsets();
        size.width += insets.left + Scale.get(parent).scale(2) + insets.right;
        size.height += insets.top + insets.bottom;
        return size;
    }
}
