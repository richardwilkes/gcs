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

import java.awt.Cursor;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;

/** Represents cells in a {@link Outline}. */
public interface Cell {
    /**
     * Draws the cell.
     *
     * @param outline  The {@link Outline} being drawn.
     * @param gc       The graphics context to use.
     * @param bounds   The bounds of the cell.
     * @param row      The row to draw.
     * @param column   The column to draw.
     * @param selected Pass in {@code true} if the cell should be drawn in its selected state.
     * @param active   Pass in {@code true} if the cell should be drawn in its active state.
     */
    void drawCell(Outline outline, Graphics gc, Rectangle bounds, Row row, Column column, boolean selected, boolean active);

    /**
     * @param outline The {@link Outline} being used.
     * @param row     The row to get data from.
     * @param column  The column to get data from.
     * @return The preferred width of the cell for the specified data.
     */
    int getPreferredWidth(Outline outline, Row row, Column column);

    /**
     * @param outline The {@link Outline} being used.
     * @param row     The row to get data from.
     * @param column  The column to get data from.
     * @return The preferred height of the cell for the specified data.
     */
    int getPreferredHeight(Outline outline, Row row, Column column);

    /**
     * Compare the column in row one with row two.
     *
     * @param column The column to compare.
     * @param one    The first row.
     * @param two    The second row.
     * @return {@code < 0} if row one is less than row two, {@code 0} if they are equal, and {@code
     *         > 0} if row one is greater than row two.
     */
    int compare(Column column, Row one, Row two);

    /**
     * @param event  The {@link MouseEvent} that caused the tooltip to be shown.
     * @param bounds The bounds of the cell.
     * @param row    The row to get data from.
     * @param column The column to get data from.
     * @return The cursor appropriate for the location within the cell.
     */
    Cursor getCursor(MouseEvent event, Rectangle bounds, Row row, Column column);

    /**
     * @param event  The {@link MouseEvent} that caused the tooltip to be shown.
     * @param bounds The bounds of the cell.
     * @param row    The row to get data from.
     * @param column The column to get data from.
     * @return The tooltip string for this cell.
     */
    String getToolTipText(Outline outline, MouseEvent event, Rectangle bounds, Row row, Column column);

    /**
     * @return {@code true} if this cell wants to participate in dynamic row layout.
     */
    boolean participatesInDynamicRowLayout();

    /**
     * Called when a mouse click has occurred on the cell.
     *
     * @param event  The {@link MouseEvent} that caused the call.
     * @param bounds The bounds of the cell.
     * @param row    The row to get data from.
     * @param column The column to get data from.
     */
    void mouseClicked(MouseEvent event, Rectangle bounds, Row row, Column column);
}
