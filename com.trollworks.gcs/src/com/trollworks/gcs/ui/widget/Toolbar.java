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

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.widget.dock.DockColors;
import com.trollworks.gcs.utility.Log;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager;
import java.awt.LayoutManager2;
import java.util.HashMap;
import java.util.Map;
import javax.swing.JPanel;
import javax.swing.border.CompoundBorder;

public class Toolbar extends JPanel implements LayoutManager2 {
    public static final  String                 LAYOUT_FILL         = "fill";
    public static final  String                 LAYOUT_EXTRA_BEFORE = "extra_before";
    private static final int                    GAP                 = 2;
    private              Map<Component, String> mConstraints        = new HashMap<>();

    public Toolbar() {
        super(true);
        super.setLayout(this);
        setOpaque(true);
        setBackground(DockColors.BACKGROUND);
        setBorder(new CompoundBorder(new LineBorder(DockColors.SHADOW, 0, 0, 1, 0), new EmptyBorder(0, 4, 0, 4)));
    }

    @Override
    public void setLayout(LayoutManager mgr) {
        // Don't allow overrides
    }

    @Override
    public void addLayoutComponent(Component comp, Object constraints) {
        if (LAYOUT_FILL.equals(constraints) || LAYOUT_EXTRA_BEFORE.equals(constraints)) {
            mConstraints.put(comp, (String) constraints);
        } else if (constraints != null) {
            Log.error("Invalid toolbar constraints");
        }
    }

    @Override
    public void addLayoutComponent(String name, Component comp) {
        addLayoutComponent(comp, name);
    }

    @Override
    public void removeLayoutComponent(Component comp) {
        mConstraints.remove(comp);
    }

    @Override
    public float getLayoutAlignmentX(Container target) {
        return Component.CENTER_ALIGNMENT;
    }

    @Override
    public float getLayoutAlignmentY(Container target) {
        return Component.CENTER_ALIGNMENT;
    }

    @Override
    public Dimension preferredLayoutSize(Container parent) {
        return getLayoutSize((i) -> getComponent(i).getPreferredSize());
    }

    @Override
    public Dimension minimumLayoutSize(Container parent) {
        return getLayoutSize((i) -> getComponent(i).getMinimumSize());
    }

    @Override
    public Dimension maximumLayoutSize(Container target) {
        Dimension size = preferredLayoutSize(target);
        size.width = Integer.MAX_VALUE / 8;
        return size;
    }

    @Override
    public void invalidateLayout(Container target) {
        // Unused
    }

    private interface LayoutSizeRetriever {
        Dimension getLayoutSize(int index);
    }

    private Dimension getLayoutSize(LayoutSizeRetriever retriever) {
        Insets insets = getInsets();
        int    count  = getComponentCount();
        int    width  = count > 0 ? (count - 1) * GAP : 0;
        int    height = 0;
        for (int i = 0; i < count; i++) {
            Dimension size = retriever.getLayoutSize(i);
            width += size.width;
            if (height < size.height) {
                height = size.height;
            }
        }
        return new Dimension(insets.left + width + insets.right, insets.top + height + insets.bottom);
    }

    @Override
    public void layoutContainer(Container parent) {
        int         extra       = getWidth() - preferredLayoutSize(parent).width;
        Insets      insets      = getInsets();
        int         count       = getComponentCount();
        Component[] comps       = getComponents();
        int[]       widths      = new int[count];
        int[]       minWidths   = new int[count];
        int[]       heights     = new int[count];
        int[]       extraBefore = new int[count];
        boolean[]   isFill      = new boolean[count];
        boolean[]   isExtra     = new boolean[count];
        for (int i = 0; i < count; i++) {
            Dimension size = comps[i].getPreferredSize();
            widths[i] = size.width;
            heights[i] = size.height;
            minWidths[i] = comps[i].getMinimumSize().width;
            String constraint = mConstraints.get(comps[i]);
            isFill[i] = LAYOUT_FILL.equals(constraint);
            isExtra[i] = LAYOUT_EXTRA_BEFORE.equals(constraint);
        }
        if (extra < 0) {
            int     remaining = -extra;
            boolean found     = true;
            while (found && remaining > 0) {
                int avail = 0;
                found = false;
                for (int i = 0; i < count; i++) {
                    if (widths[i] > minWidths[i]) {
                        avail++;
                    }
                }
                if (avail > 0) {
                    int perComp = Math.max(remaining / avail, 1);
                    for (int i = 0; i < count && remaining > 0; i++) {
                        if (widths[i] > minWidths[i]) {
                            found = true;
                            remaining -= perComp;
                            widths[i] -= perComp;
                            if (widths[i] <= minWidths[i]) {
                                remaining += minWidths[i] - widths[i];
                                widths[i] = minWidths[i];
                            }
                        }
                    }
                }
            }
        } else if (extra > 0) {
            int     remaining = extra;
            boolean found     = true;
            while (found && remaining > 0) {
                int avail = 0;
                found = false;
                for (int i = 0; i < count; i++) {
                    if (isFill[i]) {
                        avail++;
                    }
                }
                if (avail > 0) {
                    int perComp = Math.max(remaining / avail, 1);
                    for (int i = 0; i < count && remaining > 0; i++) {
                        if (isFill[i]) {
                            found = true;
                            remaining -= perComp;
                            widths[i] += perComp;
                        }
                    }
                } else {
                    for (int i = 0; i < count; i++) {
                        if (isExtra[i]) {
                            avail++;
                        }
                    }
                    if (avail > 0) {
                        int perComp = Math.max(remaining / avail, 1);
                        for (int i = 0; i < count && remaining > 0; i++) {
                            if (isExtra[i]) {
                                found = true;
                                remaining -= perComp;
                                extraBefore[i] += perComp;
                            }
                        }
                    }
                }
            }
        }
        int x      = insets.left;
        int height = getHeight();
        for (int i = 0; i < count; i++) {
            x += extraBefore[i];
            comps[i].setBounds(x, insets.top + (height - (insets.top + heights[i] + insets.bottom)) / 2, widths[i], heights[i]);
            x += widths[i] + GAP;
        }
    }
}
