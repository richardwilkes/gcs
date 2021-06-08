/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.utility.Log;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager;
import java.awt.LayoutManager2;
import java.util.HashMap;
import java.util.Map;
import javax.swing.border.CompoundBorder;

public class Toolbar extends Panel {
    public static final  String LAYOUT_FILL         = "fill";
    public static final  String LAYOUT_EXTRA_BEFORE = "extra_before";
    private static final int    GAP                 = 8;

    public Toolbar() {
        super(new ToolbarLayout());
        setOpaque(true);
        setBorder(new CompoundBorder(new LineBorder(ThemeColor.DIVIDER, 0, 0, 1, 0), new EmptyBorder(0, GAP, 0, GAP)));
    }

    @Override
    public void setLayout(LayoutManager mgr) {
        if (mgr instanceof ToolbarLayout) {
            super.setLayout(mgr);
        }
    }

    private static class ToolbarLayout implements LayoutManager2 {
        private Map<Component, String> mConstraints = new HashMap<>();

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
            return getLayoutSize(parent, Component::getPreferredSize);
        }

        @Override
        public Dimension minimumLayoutSize(Container parent) {
            return getLayoutSize(parent, Component::getMinimumSize);
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

        private static Dimension getLayoutSize(Container parent, LayoutSizeRetriever retriever) {
            Insets insets = parent.getInsets();
            int    count  = parent.getComponentCount();
            int    width  = count > 0 ? (count - 1) * GAP : 0;
            int    height = 0;
            for (int i = 0; i < count; i++) {
                Dimension size = retriever.getLayoutSize(parent.getComponent(i));
                width += size.width;
                if (height < size.height) {
                    height = size.height;
                }
            }
            return new Dimension(insets.left + width + insets.right, insets.top + height + insets.bottom);
        }

        @Override
        public void layoutContainer(Container parent) {
            int         extra       = parent.getWidth() - preferredLayoutSize(parent).width;
            Insets      insets      = parent.getInsets();
            int         count       = parent.getComponentCount();
            Component[] comps       = parent.getComponents();
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
            int height = parent.getHeight();
            for (int i = 0; i < count; i++) {
                x += extraBefore[i];
                comps[i].setBounds(x, (height - heights[i]) / 2, widths[i], heights[i]);
                x += widths[i] + GAP;
            }
        }
    }

    private interface LayoutSizeRetriever {
        Dimension getLayoutSize(Component comp);
    }
}
