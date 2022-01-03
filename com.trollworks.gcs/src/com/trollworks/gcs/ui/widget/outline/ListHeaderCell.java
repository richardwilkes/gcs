/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.widget.outline;

import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Color;
import java.awt.Graphics2D;
import java.awt.Rectangle;

/** Used to draw headers in the lists. */
public class ListHeaderCell extends HeaderCell {
    private boolean mForSheet;

    /**
     * Create a new header cell.
     *
     * @param forSheet Whether the header will be displayed in the sheet or not.
     */
    public ListHeaderCell(boolean forSheet) {
        mForSheet = forSheet;
    }

    @Override
    public ThemeFont getThemeFont(Row row, Column column) {
        return Fonts.PAGE_LABEL_PRIMARY;
    }

    @Override
    public void drawCell(Outline outline, Graphics2D gc, Rectangle bounds, Row row, Column column, boolean selected, boolean active) {
        if (mForSheet) {
            drawCellSuper(outline, gc, bounds, row, column, selected, active);
        } else {
            super.drawCell(outline, gc, bounds, row, column, selected, active);
        }
    }

    @Override
    public int getPreferredWidth(Outline outline, Row row, Column column) {
        int width = super.getPreferredWidth(outline, row, column);
        if (mForSheet) {
            Scale scale  = Scale.get(outline);
            int   margin = scale.scale(2);
            width -= margin + scale.scale(SORTER_WIDTH) + margin;
        }
        return width;
    }

    @Override
    public Color getColor(Outline outline, Row row, Column column, boolean selected, boolean active) {
        if (mForSheet) {
            return Colors.ON_HEADER;
        }
        return super.getColor(outline, row, column, selected, active);
    }
}
