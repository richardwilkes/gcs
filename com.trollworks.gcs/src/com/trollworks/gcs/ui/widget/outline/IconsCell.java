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

package com.trollworks.gcs.ui.widget.outline;

import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.awt.Cursor;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.util.ArrayList;
import java.util.List;
import javax.swing.SwingConstants;

/** Represents icons in an {@link Outline}. */
public class IconsCell implements Cell {
    private int mHAlignment;
    private int mVAlignment;

    /** Create a new image cell renderer. */
    public IconsCell() {
        this(SwingConstants.CENTER, SwingConstants.CENTER);
    }

    /**
     * Create a new image cell renderer.
     *
     * @param hAlignment The image horizontal alignment to use.
     * @param vAlignment The image vertical alignment to use.
     */
    public IconsCell(int hAlignment, int vAlignment) {
        mHAlignment = hAlignment;
        mVAlignment = vAlignment;
    }

    @Override
    public int compare(Column column, Row one, Row two) {
        String oneText = one.getDataAsText(column);
        String twoText = two.getDataAsText(column);
        return NumericComparator.caselessCompareStrings(oneText != null ? oneText : "", twoText != null ? twoText : "");
    }

    /**
     * @param row      The row to use.
     * @param column   The column to use.
     * @param selected Whether the row is selected.
     * @param active   Whether the outline is active.
     * @return The icon, if any.
     */
    @SuppressWarnings("static-method")
    protected List<RetinaIcon> getIcons(Row row, Column column, boolean selected, boolean active) {
        List<RetinaIcon> list = new ArrayList<>();
        Object           data = row.getData(column);
        if (data instanceof RetinaIcon) {
            list.add((RetinaIcon) data);
        } else if (data instanceof List) {
            for (Object obj : (List<?>) data) {
                if (obj instanceof RetinaIcon) {
                    list.add((RetinaIcon) obj);
                }
            }
        }
        return list;
    }

    @Override
    public void drawCell(Outline outline, Graphics gc, Rectangle bounds, Row row, Column column, boolean selected, boolean active) {
        if (row != null) {
            List<RetinaIcon> images = getIcons(row, column, selected, active);
            if (!images.isEmpty()) {
                Scale scale = Scale.get(outline);
                int   x     = bounds.x;
                int   y     = bounds.y;
                if (mHAlignment != SwingConstants.LEFT) {
                    int hDelta = bounds.width;
                    for (RetinaIcon img : images) {
                        hDelta -= scale.scale(img.getIconWidth());
                    }
                    if (mHAlignment == SwingConstants.CENTER) {
                        hDelta /= 2;
                    }
                    x += hDelta;
                }
                if (mVAlignment != SwingConstants.TOP) {
                    int max = 0;
                    for (RetinaIcon img : images) {
                        int height = scale.scale(img.getIconHeight());
                        if (max < height) {
                            max = height;
                        }
                    }
                    int vDelta = bounds.height - max;
                    if (mVAlignment == SwingConstants.CENTER) {
                        vDelta /= 2;
                    }
                    y += vDelta;
                }
                for (RetinaIcon img : images) {
                    img.paintIcon(outline, gc, x, y);
                    x += scale.scale(img.getIconWidth());
                }
            }
        }
    }

    @Override
    public int getPreferredWidth(Outline outline, Row row, Column column) {
        Scale scale = Scale.get(outline);
        int   width = 0;
        for (RetinaIcon img : getIcons(row, column, false, true)) {
            width += scale.scale(img.getIconWidth());
        }
        return width;
    }

    @Override
    public int getPreferredHeight(Outline outline, Row row, Column column) {
        Scale scale  = Scale.get(outline);
        int   height = 0;
        for (RetinaIcon img : getIcons(row, column, false, true)) {
            height += scale.scale(img.getIconHeight());
        }
        return height;
    }

    @Override
    public Cursor getCursor(MouseEvent event, Rectangle bounds, Row row, Column column) {
        return Cursor.getDefaultCursor();
    }

    @Override
    public String getToolTipText(Outline outline, MouseEvent event, Rectangle bounds, Row row, Column column) {
        return null;
    }

    @Override
    public boolean participatesInDynamicRowLayout() {
        return false;
    }

    @Override
    public void mouseClicked(MouseEvent event, Rectangle bounds, Row row, Column column) {
        // Does nothing
    }
}
