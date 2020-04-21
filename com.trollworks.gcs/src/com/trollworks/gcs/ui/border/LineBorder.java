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

import java.awt.Color;
import java.awt.Component;
import java.awt.Graphics;
import java.awt.Insets;
import javax.swing.border.Border;

/** A border that allows varying colors and thicknesses for each side. */
public class LineBorder implements Border {
    private Color[] mColor     = new Color[Edge.values().length];
    private int[]   mThickness = new int[Edge.values().length];

    /** Creates a black, 1 pixel border on all sides. */
    public LineBorder() {
        this(Color.BLACK, 1);
    }

    /**
     * Creates a 1 pixel border with the specified color.
     *
     * @param color The color to use for all sides.
     */
    public LineBorder(Color color) {
        for (Edge edge : Edge.values()) {
            setColorAndThickness(edge, color, 1);
        }
    }

    /**
     * Creates a border with the specified color and thickness on all sides.
     *
     * @param color     The color to use for all sides.
     * @param thickness The thickness to use for all sides.
     */
    public LineBorder(Color color, int thickness) {
        for (Edge edge : Edge.values()) {
            setColorAndThickness(edge, color, thickness);
        }
    }

    /**
     * Creates a border with the specified color on all sides and the specified thicknesses on each
     * side.
     *
     * @param color  The color to use for all sides.
     * @param top    The thickness to use for the top side.
     * @param left   The thickness to use for the left side.
     * @param bottom The thickness to use for the bottom side.
     * @param right  The thickness to use for the right side.
     */
    public LineBorder(Color color, int top, int left, int bottom, int right) {
        setColorAndThickness(Edge.TOP, color, top);
        setColorAndThickness(Edge.LEFT, color, left);
        setColorAndThickness(Edge.BOTTOM, color, bottom);
        setColorAndThickness(Edge.RIGHT, color, right);
    }

    /**
     * Creates a border.
     *
     * @param topColor    The color to use for the top side.
     * @param top         The thickness to use for the top side.
     * @param leftColor   The color to use for the left side.
     * @param left        The thickness to use for the left side.
     * @param bottomColor The color to use for the bottom side.
     * @param bottom      The thickness to use for the bottom side.
     * @param rightColor  The color to use for the right side.
     * @param right       The thickness to use for the right side.
     */
    public LineBorder(Color topColor, int top, Color leftColor, int left, Color bottomColor, int bottom, Color rightColor, int right) {
        setColorAndThickness(Edge.TOP, topColor, top);
        setColorAndThickness(Edge.LEFT, leftColor, left);
        setColorAndThickness(Edge.BOTTOM, bottomColor, bottom);
        setColorAndThickness(Edge.RIGHT, rightColor, right);
    }

    /**
     * @param edge The edge to query.
     * @return The color being used for the specified edge.
     */
    public Color getColor(Edge edge) {
        return mColor[edge.ordinal()];
    }

    /**
     * @param edge  The edge to set.
     * @param color The color to use for the specified edge.
     */
    public void setColor(Edge edge, Color color) {
        mColor[edge.ordinal()] = color;
    }

    /**
     * @param edge The edge to query.
     * @return The thickness being used for the specified edge.
     */
    public int getThickness(Edge edge) {
        return mThickness[edge.ordinal()];
    }

    /**
     * @param edge      The edge to set.
     * @param thickness The thickness to use for the specified edge.
     */
    public void setThickness(Edge edge, int thickness) {
        mThickness[edge.ordinal()] = thickness;
    }

    /**
     * @param edge      The edge to set.
     * @param color     The color to use for the specified edge.
     * @param thickness The thickness to use for the specified edge.
     */
    public void setColorAndThickness(Edge edge, Color color, int thickness) {
        int i = edge.ordinal();
        mColor[i] = color;
        mThickness[i] = thickness;
    }

    @Override
    public void paintBorder(Component component, Graphics gc, int x, int y, int width, int height) {
        Scale scale      = Scale.get(component);
        Color savedColor = gc.getColor();
        int   i          = Edge.LEFT.ordinal();
        int   thickness  = scale.scale(mThickness[i]);
        if (thickness > 0) {
            gc.setColor(mColor[i]);
            gc.fillRect(x, y, thickness, height);
        }
        i = Edge.RIGHT.ordinal();
        thickness = scale.scale(mThickness[i]);
        if (thickness > 0) {
            gc.setColor(mColor[i]);
            gc.fillRect(x + width - thickness, y, thickness, height);
        }
        i = Edge.TOP.ordinal();
        thickness = scale.scale(mThickness[i]);
        if (thickness > 0) {
            gc.setColor(mColor[i]);
            gc.fillRect(x, y, width, thickness);
        }
        i = Edge.BOTTOM.ordinal();
        thickness = scale.scale(mThickness[i]);
        if (thickness > 0) {
            gc.setColor(mColor[i]);
            gc.fillRect(x, y + height - thickness, width, thickness);
        }
        gc.setColor(savedColor);
    }

    @Override
    public Insets getBorderInsets(Component component) {
        return Scale.get(component).scale(new Insets(mThickness[Edge.TOP.ordinal()], mThickness[Edge.LEFT.ordinal()], mThickness[Edge.BOTTOM.ordinal()], mThickness[Edge.RIGHT.ordinal()]));
    }

    @Override
    public boolean isBorderOpaque() {
        return true;
    }
}
