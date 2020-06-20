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

import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.awt.Cursor;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import javax.swing.SwingConstants;
import javax.swing.UIManager;

public class WrappedCell implements Cell {
    private static final int H_MARGIN = 2;

    @Override
    public void drawCell(Outline outline, Graphics gc, Rectangle bounds, Row row, Column column, boolean selected, boolean active) {
        Scale scale   = Scale.get(outline);
        int   hMargin = scale.scale(H_MARGIN);
        gc.setColor(Colors.getListForeground(selected, active));
        gc.setFont(scale.scale(UIManager.getFont(Fonts.KEY_FIELD_PRIMARY)));
        TextDrawing.draw(gc, new Rectangle(bounds.x + hMargin, bounds.y, bounds.width - hMargin * 2, bounds.height), row.getDataAsText(column), SwingConstants.LEFT, SwingConstants.TOP);
    }

    @Override
    public int getPreferredWidth(Outline outline, Row row, Column column) {
        Scale scale = Scale.get(outline);
        int   width = TextDrawing.getWidth(scale.scale(UIManager.getFont(Fonts.KEY_FIELD_PRIMARY)), row.getDataAsText(column));
        return width + scale.scale(H_MARGIN) * 2;
    }

    @Override
    public int getPreferredHeight(Outline outline, Row row, Column column) {
        Scale scale = Scale.get(outline);
        Font  font  = scale.scale(UIManager.getFont(Fonts.KEY_FIELD_PRIMARY));
        return TextDrawing.getPreferredSize(font, wrap(row, column, row.getDataAsText(column), font, scale)).height;
    }

    private static String wrap(Row row, Column column, String text, Font font, Scale scale) {
        int width = column.getWidth();
        if (width == -1) {
            return text;
        }
        return TextDrawing.wrapToPixelWidth(font, text, width - (scale.scale(row.getOwner().getIndentWidth(row, column)) + scale.scale(H_MARGIN) * 2));
    }

    @Override
    public int compare(Column column, Row one, Row two) {
        return NumericComparator.caselessCompareStrings(one.getDataAsText(column), two.getDataAsText(column));
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
        return true;
    }

    @Override
    public void mouseClicked(MouseEvent event, Rectangle bounds, Row row, Column column) {
        // Does nothing
    }
}
