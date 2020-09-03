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

package com.trollworks.gcs.ui.border;

import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Component;
import java.awt.Graphics;
import java.awt.Insets;
import javax.swing.border.Border;

/** A replacement for the Swing EmptyBorder class that understands scaling. */
public class EmptyBorder implements Border {
    private int[] mThickness = new int[Edge.values().length];

    /**
     * Creates a new border.
     *
     * @param thickness The thickness of each edge.
     */
    public EmptyBorder(int thickness) {
        for (Edge edge : Edge.values()) {
            setThickness(edge, thickness);
        }
    }

    /**
     * Creates a new border.
     *
     * @param top    The thickness to use for the top side.
     * @param left   The thickness to use for the left side.
     * @param bottom The thickness to use for the bottom side.
     * @param right  The thickness to use for the right side.
     */
    public EmptyBorder(int top, int left, int bottom, int right) {
        setThickness(Edge.TOP, top);
        setThickness(Edge.LEFT, left);
        setThickness(Edge.BOTTOM, bottom);
        setThickness(Edge.RIGHT, right);
    }

    /**
     * @param edge      The edge to set.
     * @param thickness The thickness to use for the specified edge.
     */
    public void setThickness(Edge edge, int thickness) {
        mThickness[edge.ordinal()] = thickness;
    }

    @Override
    public void paintBorder(Component component, Graphics g, int x, int y, int width, int height) {
        // Does nothing
    }

    @Override
    public Insets getBorderInsets(Component component) {
        return Scale.get(component).scale(new Insets(mThickness[Edge.TOP.ordinal()], mThickness[Edge.LEFT.ordinal()], mThickness[Edge.BOTTOM.ordinal()], mThickness[Edge.RIGHT.ordinal()]));
    }

    @Override
    public boolean isBorderOpaque() {
        return false;
    }
}
