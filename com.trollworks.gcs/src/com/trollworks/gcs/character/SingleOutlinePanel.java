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

import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineHeader;
import com.trollworks.gcs.ui.widget.outline.OutlineProxy;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;
import java.awt.Rectangle;

/** An outline panel. */
public class SingleOutlinePanel extends DropPanel implements LayoutManager2 {
    private OutlineHeader mHeader;
    private Outline       mOutline;

    /**
     * Creates a new outline panel.
     *
     * @param scale    The scale to use.
     * @param outline  The outline to display.
     * @param title    The localized title for the panel.
     * @param useProxy {@code true} if a proxy of the outline should be used.
     */
    public SingleOutlinePanel(Scale scale, Outline outline, String title, boolean useProxy) {
        super(null);
        mOutline = useProxy ? new OutlineProxy(outline) : outline;
        mHeader = mOutline.getHeaderPanel();
        CollectedOutlines.prepOutline(mOutline);
        add(mHeader);
        add(mOutline);
        setBorder(getTitledBorder());
        setLayout(this);
    }

    /**
     * Sets the embedded outline's display range.
     *
     * @param first The first row to display.
     * @param last  The last row to display.
     */
    public void setOutlineRowRange(int first, int last) {
        mOutline.setFirstRowToDisplay(first);
        mOutline.setLastRowToDisplay(last);
    }

    @Override
    public void layoutContainer(Container parent) {
        Insets    insets = getInsets();
        Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
        int       height = mHeader.getPreferredSize().height;
        mHeader.setLocation(bounds.x, bounds.y);
        bounds.y += height;
        bounds.height -= height;
        mOutline.setBounds(bounds.x, bounds.y, bounds.width, bounds.height);
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
