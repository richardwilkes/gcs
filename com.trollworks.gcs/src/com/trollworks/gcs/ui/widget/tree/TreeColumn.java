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

package com.trollworks.gcs.ui.widget.tree;

import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.TextDrawing;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics2D;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.datatransfer.DataFlavor;
import java.awt.datatransfer.Transferable;
import java.awt.datatransfer.UnsupportedFlavorException;
import java.util.Comparator;
import javax.swing.SwingConstants;
import javax.swing.UIManager;

/** The abstract base class for columns within a {@link TreeRoot}. */
public abstract class TreeColumn implements Comparator<TreeRow>, Transferable {
    /** The data flavor for this class. */
    public static final  DataFlavor DATA_FLAVOR       = new DataFlavor(TreeColumn.class, "Tree Column");
    private static final int        SORTER_WIDTH      = 12;
    private static final int        HALF_SORTER_WIDTH = SORTER_WIDTH / 2;
    private              String     mName;
    private              int        mWidth;

    /**
     * Creates a new {@link TreeColumn}.
     *
     * @param name The name of the {@link TreeColumn}.
     */
    protected TreeColumn(String name) {
        mName = name;
    }

    /** @return The name of this {@link TreeColumn}. */
    public String getName() {
        return mName;
    }

    /**
     * @param row The {@link TreeRow} to be displayed.
     * @return The preferred width of this {@link TreeColumn}.
     */
    public abstract int calculatePreferredWidth(TreeRow row);

    /**
     * @param row   The {@link TreeRow} to be displayed.
     * @param width The adjusted width of the column. This may be less than {@link #getWidth()} due
     *              to display of disclosure controls.
     * @return The preferred height of this {@link TreeColumn}.
     */
    public abstract int calculatePreferredHeight(TreeRow row, int width);

    /**
     * @param panel The owning {@link TreePanel}.
     * @return The preferred height of this {@link TreeColumn}'s header.
     */
    public Dimension calculatePreferredHeaderSize(TreePanel panel) {
        Font font  = getHeaderFont();
        int  width = getWidth();
        if (panel.isUserSortable()) {
            width -= SORTER_WIDTH;
        }
        Dimension size = TextDrawing.getPreferredSize(font, TextDrawing.wrapToPixelWidth(font, getName(), width));
        if (panel.isUserSortable()) {
            size.width += SORTER_WIDTH;
        }
        return size;
    }

    /** @return The full width of this {@link TreeColumn}. */
    public int getWidth() {
        return mWidth;
    }

    /** @param width The new full width of this {@link TreeColumn}. */
    public void setWidth(int width) {
        mWidth = width;
    }

    /**
     * @param panel The owning {@link TreePanel}.
     * @return The minimum width the column should be set to.
     */
    public int getMinimumWidth(TreePanel panel) {
        int width = TextDrawing.getPreferredSize(getHeaderFont(), getName()).width;
        if (panel.isUserSortable()) {
            width += SORTER_WIDTH;
        }
        return width;
    }

    /**
     * Draws the header for this {@link TreeColumn}.
     *
     * @param gc     The {@link Graphics2D} context to use.
     * @param panel  The owning {@link TreePanel}.
     * @param bounds The bounds the {@link TreeColumn} header.
     * @param active Whether or not the active state should be displayed.
     */
    public void drawHeader(Graphics2D gc, TreePanel panel, Rectangle bounds, boolean active) {
        Font  savedFont  = gc.getFont();
        Color savedColor = gc.getColor();
        gc.setColor(UIManager.getColor("Panel.background"));
        gc.fill(bounds);
        gc.setColor(UIManager.getColor("Panel.foreground"));
        Font font = getHeaderFont();
        gc.setFont(font);
        int        sortSequence = -1;
        TreeSorter treeSorter   = panel.getTreeSorter();
        if (panel.isUserSortable()) {
            sortSequence = treeSorter.getSortSequence(this);
        }
        if (sortSequence != -1) {
            bounds.width -= SORTER_WIDTH;
        }
        String text = TextDrawing.wrapToPixelWidth(font, getName(), bounds.width);
        TextDrawing.draw(gc, bounds, text, SwingConstants.CENTER, SwingConstants.CENTER);
        if (sortSequence != -1) {
            bounds.width += SORTER_WIDTH;
            gc.setColor(getSorterColor());
            int x = bounds.x + bounds.width - HALF_SORTER_WIDTH;
            int y = bounds.y + bounds.height / 2 - 1;
            if (treeSorter.hasMultipleCriteria()) {
                int yb = y - 2;
                int yc = yb - 2;
                int yt = yb - 4;
                int xl = x - 1;
                int xc = xl + 1;
                int xr = xl + 2;
                switch (sortSequence + 1) {
                case 1 -> gc.drawLine(xc, yt, xc, yb);
                case 2 -> {
                    gc.drawLine(xl, yt, xr, yt);
                    gc.drawLine(xr, yt, xr, yc);
                    gc.drawLine(xl, yc, xr, yc);
                    gc.drawLine(xl, yc, xl, yb);
                    gc.drawLine(xl, yb, xr, yb);
                }
                case 3 -> {
                    gc.drawLine(xl, yt, xr, yt);
                    gc.drawLine(xr, yt, xr, yb);
                    gc.drawLine(xc, yc, xr, yc);
                    gc.drawLine(xl, yb, xr, yb);
                }
                case 4 -> {
                    gc.drawLine(xl, yt, xl, yc);
                    gc.drawLine(xl, yc, xr, yc);
                    gc.drawLine(xr, yt, xr, yb);
                }
                case 5 -> {
                    gc.drawLine(xl, yt, xr, yt);
                    gc.drawLine(xl, yt, xl, yc);
                    gc.drawLine(xl, yc, xr, yc);
                    gc.drawLine(xr, yc, xr, yb);
                    gc.drawLine(xl, yb, xr, yb);
                }
                case 6 -> {
                    gc.drawLine(xl, yt, xr, yt);
                    gc.drawLine(xl, yt, xl, yb);
                    gc.drawLine(xl, yc, xr, yc);
                    gc.drawLine(xr, yc, xr, yb);
                    gc.drawLine(xl, yb, xr, yb);
                }
                case 7 -> {
                    gc.drawLine(xl, yt, xr, yt);
                    gc.drawLine(xr, yt, xr, yb);
                }
                case 8 -> {
                    gc.drawLine(xl, yt, xr, yt);
                    gc.drawLine(xl, yt, xl, yb);
                    gc.drawLine(xl, yc, xr, yc);
                    gc.drawLine(xl, yb, xr, yb);
                    gc.drawLine(xr, yt, xr, yb);
                }
                case 9 -> {
                    gc.drawLine(xl, yt, xr, yt);
                    gc.drawLine(xl, yt, xl, yc);
                    gc.drawLine(xl, yc, xr, yc);
                    gc.drawLine(xr, yt, xr, yb);
                    gc.drawLine(xl, yb, xr, yb);
                }
                default -> {
                    gc.drawLine(xl, yt, xr, yt);
                    gc.drawLine(xr, yt, xr, yc);
                    gc.drawLine(xc, yc, xr, yc);
                    gc.drawLine(xc, yb, xc, yb);
                }
                }
            }
            boolean ascending = treeSorter.isSortAscending(this);
            for (int i = 0; i < HALF_SORTER_WIDTH - 1; i++) {
                int x1;
                int x2;
                if (ascending) {
                    x1 = x - i;
                    x2 = x + i;
                } else {
                    x1 = x + i - (HALF_SORTER_WIDTH - 2);
                    x2 = x + HALF_SORTER_WIDTH - 2 - i;
                }
                gc.drawLine(x1, y + i, x2, y + i);
            }
        }
        gc.setFont(savedFont);
        gc.setColor(savedColor);
    }

    /**
     * Draws the portion of the specified {@link TreeRow} this {@link TreeColumn} is responsible
     * for.
     *
     * @param gc       The {@link Graphics2D} context to use.
     * @param panel    The owning {@link TreePanel}.
     * @param row      The {@link TreeRow} to draw.
     * @param position The {@link TreeRow}'s position in the linear view.
     * @param top      The y-coordinate for the top of the {@link TreeRow}.
     * @param left     The x-coordinate for the left of the {@link TreeColumn}.
     * @param width    The adjusted width of the column. This may be less than {@link #getWidth()}
     *                 due to display of disclosure controls.
     * @param selected Whether or not the {@link TreeRow} is currently selected.
     * @param active   Whether or not the active state should be displayed.
     */
    public abstract void draw(Graphics2D gc, TreePanel panel, TreeRow row, int position, int top, int left, int width, boolean selected, boolean active);

    /**
     * Called when a mouse press occurs within the column. By default, does nothing.
     *
     * @param row   The {@link TreeRow} at the mouse press location.
     * @param where The mouse press location.
     * @return {@code true} if the mouse press has been handled.
     */
    @SuppressWarnings("static-method")
    public boolean mousePress(TreeRow row, Point where) {
        return false;
    }

    /** @return The {@link Font} to use for the header. */
    @SuppressWarnings("static-method")
    public Font getHeaderFont() {
        return Fonts.getDefaultFont();
    }

    /** @return The {@link Color} to use for the sorter controls. */
    @SuppressWarnings("static-method")
    public Color getSorterColor() {
        return Color.blue.darker();
    }

    @Override
    public DataFlavor[] getTransferDataFlavors() {
        return new DataFlavor[]{DATA_FLAVOR};
    }

    @Override
    public boolean isDataFlavorSupported(DataFlavor flavor) {
        return DATA_FLAVOR.equals(flavor);
    }

    @Override
    public Object getTransferData(DataFlavor flavor) throws UnsupportedFlavorException {
        if (DATA_FLAVOR.equals(flavor)) {
            return this;
        }
        throw new UnsupportedFlavorException(flavor);
    }
}
