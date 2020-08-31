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

import java.awt.Color;
import java.awt.Cursor;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import javax.swing.SwingConstants;
import javax.swing.UIManager;

/**
 * A {@link Cell} for displaying both a piece of primary information and a piece of secondary
 * information for a {@link ListRow}.
 */
public class MultiCell implements Cell {
    private static final int     H_MARGIN = 2;
    private              int     mMaxPreferredWidth;
    private              boolean mForEditor;

    /** Creates a new {@link MultiCell} with a maximum preferred width of 250. */
    public MultiCell() {
        this(false);
    }

    /**
     * Creates a new {@link MultiCell} with a maximum preferred width of 250.
     *
     * @param forEditor Whether this is for an editor dialog or for a character sheet.
     */
    public MultiCell(boolean forEditor) {
        this(250, forEditor);
    }

    /**
     * Creates a new {@link MultiCell}.
     *
     * @param maxPreferredWidth The maximum preferred width to use. Pass in -1 for no limit.
     * @param forEditor         Whether this is for an editor dialog or for a character sheet.
     */
    public MultiCell(int maxPreferredWidth, boolean forEditor) {
        mMaxPreferredWidth = maxPreferredWidth;
        mForEditor = forEditor;
    }

    /** @return The primary font. */
    public Font getPrimaryFont() {
        return UIManager.getFont(mForEditor ? "TextField.font" : Fonts.KEY_FIELD_PRIMARY);
    }

    /** @return The secondary font, for notes. */
    public Font getSecondaryFont() {
        if (mForEditor) {
            Font font = getPrimaryFont();
            return font.deriveFont(font.getSize() * 7.0f / 8.0f);
        }
        return UIManager.getFont(Fonts.KEY_FIELD_SECONDARY);
    }

    /**
     * @param row The row to use.
     * @return The primary text to display.
     */
    @SuppressWarnings("static-method")
    protected String getPrimaryText(ListRow row) {
        return row.toString();
    }

    /**
     * @param row The row to use.
     * @return The text to sort.
     */
    protected String getSortText(ListRow row) {
        String text      = getPrimaryText(row);
        String secondary = getSecondaryText(row);
        if (secondary != null && !secondary.isEmpty()) {
            text += '\n';
            text += secondary;
        }
        return text;
    }

    /**
     * @param row The row to use.
     * @return The secondary text to display.
     */
    @SuppressWarnings("static-method")
    protected String getSecondaryText(ListRow row) {
        return row.getSecondaryText();
    }

    @Override
    public void drawCell(Outline outline, Graphics gc, Rectangle bounds, Row row, Column column, boolean selected, boolean active) {
        Scale     scale       = Scale.get(outline);
        ListRow   theRow      = (ListRow) row;
        int       hMargin     = scale.scale(H_MARGIN);
        Rectangle insetBounds = new Rectangle(bounds.x + hMargin, bounds.y, bounds.width - hMargin * 2, bounds.height);
        String    notes       = getSecondaryText(theRow);
        Font      font        = scale.scale(getPrimaryFont());
        int       pos;
        gc.setColor(getColor(selected, active, row, column));
        gc.setFont(font);
        Color strikeThru = row instanceof Switchable && !((Switchable) row).isEnabled() ? Color.RED : null;
        pos = TextDrawing.draw(gc, insetBounds, getPrimaryText(theRow), SwingConstants.LEFT, SwingConstants.TOP, strikeThru, scale.scale(1));
        if (!notes.trim().isEmpty()) {
            insetBounds.height -= pos - insetBounds.y;
            insetBounds.y = pos;
            gc.setFont(scale.scale(getSecondaryFont()));
            TextDrawing.draw(gc, insetBounds, notes, SwingConstants.LEFT, SwingConstants.TOP);
        }
    }

    /**
     * @param selected Whether or not the selected version of the color is needed.
     * @param active   Whether or not the active version of the color is needed.
     * @param row      The row.
     * @param column   The column.
     * @return The foreground color.
     */
    @SuppressWarnings("static-method")
    public Color getColor(boolean selected, boolean active, Row row, Column column) {
        if (((ListRow) row).isSatisfied()) {
            return Colors.getListForeground(selected, active);
        }
        return Color.RED;
    }

    @Override
    public int getPreferredWidth(Outline outline, Row row, Column column) {
        Scale   scale  = Scale.get(outline);
        ListRow theRow = (ListRow) row;
        int     width  = TextDrawing.getWidth(scale.scale(getPrimaryFont()), getPrimaryText(theRow));
        String  notes  = getSecondaryText(theRow);
        if (!notes.trim().isEmpty()) {
            int notesWidth = TextDrawing.getWidth(scale.scale(getSecondaryFont()), notes);
            if (notesWidth > width) {
                width = notesWidth;
            }
        }
        if (mMaxPreferredWidth > 0) {
            int scaledMax = scale.scale(mMaxPreferredWidth);
            if (scaledMax < width) {
                width = scaledMax;
            }
        }
        return width + scale.scale(H_MARGIN) * 2;
    }

    @Override
    public int getPreferredHeight(Outline outline, Row row, Column column) {
        Scale   scale  = Scale.get(outline);
        ListRow theRow = (ListRow) row;
        Font    font   = scale.scale(getPrimaryFont());
        int     height = TextDrawing.getPreferredSize(font, wrap(scale, theRow, column, getPrimaryText(theRow), font)).height;
        String  notes  = getSecondaryText(theRow);
        if (!notes.trim().isEmpty()) {
            font = scale.scale(getSecondaryFont());
            height += TextDrawing.getPreferredSize(font, wrap(scale, theRow, column, notes, font)).height;
        }
        return height;
    }

    private String wrap(Scale scale, ListRow row, Column column, String text, Font font) {
        int width = column.getWidth();
        if (width == -1) {
            if (mMaxPreferredWidth < 1) {
                return text;
            }
            width = scale.scale(mMaxPreferredWidth);
        }
        OutlineModel owner  = row.getOwner();
        int          indent = owner != null ? scale.scale(owner.getIndentWidth(row, column)) : 0;
        return TextDrawing.wrapToPixelWidth(font, text, width - (indent + scale.scale(H_MARGIN) * 2));
    }

    @Override
    public int compare(Column column, Row one, Row two) {
        return NumericComparator.caselessCompareStrings(getSortText((ListRow) one), getSortText((ListRow) two));
    }

    @Override
    public Cursor getCursor(MouseEvent event, Rectangle bounds, Row row, Column column) {
        return Cursor.getDefaultCursor();
    }

    /** Allow a satisfied row to display a tooltip */
    @Override
    public String getToolTipText(Outline outline, MouseEvent event, Rectangle bounds, Row row, Column column) {
        ListRow theRow = (ListRow) row;
        if (!theRow.isSatisfied()) {
            return theRow.getReasonForUnsatisfied();
        }
        String text = row.getToolTip(column);
        if (text == null || text.isBlank()) {
            return null;
        }
        return text;
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
