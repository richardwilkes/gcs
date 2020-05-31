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

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.widget.outline.Cell;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.awt.Cursor;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import javax.swing.SwingConstants;
import javax.swing.UIManager;

/**
 * A {@link Cell} for displaying both a piece of primary information and a piece of secondary
 * information for a {@link WeaponDisplayRow}.
 */
public class WeaponDescriptionCell implements Cell {
    private static final int H_MARGIN = 2;

    /**
     * @param row The row to use.
     * @return The primary text to display.
     */
    @SuppressWarnings("static-method")
    protected String getPrimaryText(WeaponDisplayRow row) {
        return row.getWeapon().toString();
    }

    /**
     * @param row The row to use.
     * @return The secondary text to display.
     */
    @SuppressWarnings("static-method")
    protected String getSecondaryText(WeaponDisplayRow row) {
        return row.getWeapon().getNotes();
    }

    @Override
    public void drawCell(Outline outline, Graphics gc, Rectangle bounds, Row row, Column column, boolean selected, boolean active) {
        Scale            scale       = Scale.get(outline);
        int              hMargin     = scale.scale(H_MARGIN);
        WeaponDisplayRow theRow      = (WeaponDisplayRow) row;
        Rectangle        insetBounds = new Rectangle(bounds.x + hMargin, bounds.y, bounds.width - hMargin * 2, bounds.height);
        String           notes       = getSecondaryText(theRow);
        Font             font        = scale.scale(UIManager.getFont(Fonts.KEY_FIELD_PRIMARY));
        gc.setColor(Colors.getListForeground(selected, active));
        gc.setFont(font);
        int pos = TextDrawing.draw(gc, insetBounds, getPrimaryText(theRow), SwingConstants.LEFT, SwingConstants.TOP);
        if (!notes.trim().isEmpty()) {
            insetBounds.height -= pos - insetBounds.y;
            insetBounds.y = pos;
            gc.setFont(scale.scale(UIManager.getFont(Fonts.KEY_FIELD_SECONDARY)));
            TextDrawing.draw(gc, insetBounds, notes, SwingConstants.LEFT, SwingConstants.TOP);
        }
    }

    @Override
    public int getPreferredWidth(Outline outline, Row row, Column column) {
        Scale            scale  = Scale.get(outline);
        WeaponDisplayRow theRow = (WeaponDisplayRow) row;
        int              width  = TextDrawing.getWidth(scale.scale(UIManager.getFont(Fonts.KEY_FIELD_PRIMARY)), getPrimaryText(theRow));
        String           notes  = getSecondaryText(theRow);
        if (!notes.trim().isEmpty()) {
            int notesWidth = TextDrawing.getWidth(scale.scale(UIManager.getFont(Fonts.KEY_FIELD_SECONDARY)), notes);
            if (notesWidth > width) {
                width = notesWidth;
            }
        }
        return width + scale.scale(H_MARGIN) * 2;
    }

    @Override
    public int getPreferredHeight(Outline outline, Row row, Column column) {
        Scale            scale  = Scale.get(outline);
        WeaponDisplayRow theRow = (WeaponDisplayRow) row;
        Font             font   = scale.scale(UIManager.getFont(Fonts.KEY_FIELD_PRIMARY));
        int              height = TextDrawing.getPreferredSize(font, wrap(theRow, column, getPrimaryText(theRow), font, scale)).height;
        String           notes  = getSecondaryText(theRow);
        if (!notes.trim().isEmpty()) {
            font = scale.scale(UIManager.getFont(Fonts.KEY_FIELD_SECONDARY));
            height += TextDrawing.getPreferredSize(font, wrap(theRow, column, notes, font, scale)).height;
        }
        return height;
    }

    private static String wrap(WeaponDisplayRow row, Column column, String text, Font font, Scale scale) {
        int width = column.getWidth();
        if (width == -1) {
            return text;
        }
        return TextDrawing.wrapToPixelWidth(font, text, width - (scale.scale(row.getOwner().getIndentWidth(row, column)) + scale.scale(H_MARGIN) * 2));
    }

    @Override
    public int compare(Column column, Row one, Row two) {
        WeaponDisplayRow r1     = (WeaponDisplayRow) one;
        WeaponDisplayRow r2     = (WeaponDisplayRow) two;
        int              result = NumericComparator.caselessCompareStrings(getPrimaryText(r1), getPrimaryText(r2));
        if (result == 0) {
            result = NumericComparator.caselessCompareStrings(getSecondaryText(r1), getSecondaryText(r2));
        }
        return result;
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
