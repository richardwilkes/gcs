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
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.awt.Color;
import java.awt.Font;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import javax.swing.SwingConstants;

/** Displays text in a {@link TreeColumn}. */
public class TextTreeColumn extends TreeColumn {
    public static final int           ICON_GAP = 2;
    public static final int           HMARGIN  = 2;
    public static final int           VMARGIN  = 1;
    private             FieldAccessor mFieldAccessor;
    private             IconAccessor  mIconAccessor;
    private             int           mAlignment;
    private             WrappingMode  mWrappingMode;

    public enum WrappingMode {
        NORMAL, WRAPPED, SINGLE_LINE
    }

    /**
     * Creates a new left-aligned {@link TextTreeColumn} with no wrapping.
     *
     * @param name          The name of the {@link TreeColumn}.
     * @param fieldAccessor The {@link FieldAccessor} to use.
     */
    public TextTreeColumn(String name, FieldAccessor fieldAccessor) {
        this(name, fieldAccessor, SwingConstants.LEFT);
    }

    /**
     * Creates a new {@link TextTreeColumn} with no wrapping.
     *
     * @param name          The name of the {@link TreeColumn}.
     * @param fieldAccessor The {@link FieldAccessor} to use.
     * @param alignment     The horizontal text alignment.
     */
    public TextTreeColumn(String name, FieldAccessor fieldAccessor, int alignment) {
        this(name, fieldAccessor, null, alignment, WrappingMode.NORMAL);
    }

    /**
     * Creates a new {@link TextTreeColumn}.
     *
     * @param name          The name of the {@link TreeColumn}.
     * @param fieldAccessor The {@link FieldAccessor} to use.
     * @param alignment     The horizontal text alignment.
     * @param wrappingMode  The text wrapping mode.
     */
    public TextTreeColumn(String name, FieldAccessor fieldAccessor, int alignment, WrappingMode wrappingMode) {
        this(name, fieldAccessor, null, alignment, wrappingMode);
    }

    /**
     * Creates a new left-aligned {@link TextTreeColumn} with no wrapping.
     *
     * @param name          The name of the {@link TreeColumn}.
     * @param fieldAccessor The {@link FieldAccessor} to use.
     * @param iconAccessor  The {@link IconAccessor} to use.
     */
    public TextTreeColumn(String name, FieldAccessor fieldAccessor, IconAccessor iconAccessor) {
        this(name, fieldAccessor, iconAccessor, SwingConstants.LEFT);
    }

    /**
     * Creates a new {@link TextTreeColumn} with no wrapping.
     *
     * @param name          The name of the {@link TreeColumn}.
     * @param fieldAccessor The {@link FieldAccessor} to use.
     * @param iconAccessor  The {@link IconAccessor} to use.
     * @param alignment     The horizontal text alignment.
     */
    public TextTreeColumn(String name, FieldAccessor fieldAccessor, IconAccessor iconAccessor, int alignment) {
        this(name, fieldAccessor, iconAccessor, alignment, WrappingMode.NORMAL);
    }

    /**
     * Creates a new {@link TextTreeColumn}.
     *
     * @param name          The name of the {@link TreeColumn}.
     * @param fieldAccessor The {@link FieldAccessor} to use.
     * @param iconAccessor  The {@link IconAccessor} to use.
     * @param alignment     The horizontal text alignment.
     * @param wrappingMode  The text wrapping mode.
     */
    public TextTreeColumn(String name, FieldAccessor fieldAccessor, IconAccessor iconAccessor, int alignment, WrappingMode wrappingMode) {
        super(name);
        mFieldAccessor = fieldAccessor;
        mIconAccessor = iconAccessor;
        mAlignment = alignment;
        mWrappingMode = wrappingMode;
    }

    @Override
    public int calculatePreferredHeight(TreeRow row, int width) {
        Font       font = getFont(row);
        RetinaIcon icon = getIcon(row);
        int        height;
        if (mWrappingMode == WrappingMode.SINGLE_LINE) {
            height = TextDrawing.getFontHeight(font);
            if (icon != null) {
                int iconHeight = icon.getIconHeight();
                if (iconHeight > height) {
                    height = iconHeight;
                }
            }
        } else {
            width -= HMARGIN + HMARGIN;
            if (icon != null) {
                width -= icon.getIconWidth() + ICON_GAP;
            }
            height = calculatePreferredHeight(font, getPresentationText(row, font, width, true), icon);
        }
        return VMARGIN + height + VMARGIN;
    }

    private static int calculatePreferredHeight(Font font, String text, RetinaIcon icon) {
        int height = TextDrawing.getPreferredHeight(font, text);
        if (height == 0) {
            height = TextDrawing.getFontHeight(font);
        }
        if (icon != null) {
            int iconHeight = icon.getIconHeight();
            if (iconHeight > height) {
                height = iconHeight;
            }
        }
        return height;
    }

    @Override
    public int calculatePreferredWidth(TreeRow row) {
        int        width = TextDrawing.getPreferredSize(getFont(row), getText(row)).width;
        RetinaIcon icon  = getIcon(row);
        if (icon != null) {
            width += icon.getIconWidth() + ICON_GAP;
        }
        return HMARGIN + width + HMARGIN;
    }

    @Override
    public void draw(Graphics2D gc, TreePanel panel, TreeRow row, int position, int top, int left, int width, boolean selected, boolean active) {
        left += HMARGIN;
        width -= HMARGIN + HMARGIN;
        RetinaIcon icon = getIcon(row);
        if (icon != null) {
            icon.paintIcon(panel, gc, left, top + VMARGIN);
            int iconSize = icon.getIconWidth() + ICON_GAP;
            left += iconSize;
            width -= iconSize;
        }
        Font font = getFont(row);
        gc.setFont(font);
        String text        = getPresentationText(row, font, width, false);
        int    totalHeight = calculatePreferredHeight(font, text, icon);
        gc.setColor(getColor(panel, row, position, selected, active));
        TextDrawing.draw(gc, new Rectangle(left, top + VMARGIN, width, totalHeight), text, mAlignment, SwingConstants.TOP);
    }

    /**
     * @param row           The {@link TreeRow} to extract information from.
     * @param font          The {@link Font} to use.
     * @param width         The adjusted width of the column. This may be less than {@link
     *                      #getWidth()} due to display of disclosure controls.
     * @param forHeightOnly Will be {@code true} when only the number of lines matters.
     * @return The text to display, wrapped if necessary.
     */
    protected String getPresentationText(TreeRow row, Font font, int width, boolean forHeightOnly) {
        String text = getText(row);
        if (mWrappingMode == WrappingMode.WRAPPED) {
            return TextDrawing.wrapToPixelWidth(font, text, width);
        }
        if (mWrappingMode == WrappingMode.SINGLE_LINE) {
            int cut = text.indexOf('\n');
            if (cut != -1) {
                text = text.substring(0, cut);
            }
        }
        return forHeightOnly ? text : TextDrawing.truncateIfNecessary(font, text, width, SwingConstants.CENTER);
    }

    /**
     * @param row The {@link TreeRow} to extract information from.
     * @return The text to display.
     */
    protected String getText(TreeRow row) {
        String result = mFieldAccessor.getField(row);
        return result != null ? result : "";
    }

    /**
     * @param row The {@link TreeRow} to extract information from.
     * @return The text to display.
     */
    protected RetinaIcon getIcon(TreeRow row) {
        return mIconAccessor != null ? mIconAccessor.getIcon(row) : null;
    }

    /**
     * @param row The {@link TreeRow} to extract information from.
     * @return The {@link Font} to use.
     */
    @SuppressWarnings("static-method")
    public Font getFont(TreeRow row) {
        return Fonts.getDefaultFont();
    }

    /**
     * @param panel    The owning {@link TreePanel}.
     * @param row      The {@link TreeRow} to extract information from.
     * @param position The {@link TreeRow}'s position in the linear view.
     * @param selected Whether or not the {@link TreeRow} is currently selected.
     * @param active   Whether or not the active state should be displayed.
     * @return The foreground color.
     */
    @SuppressWarnings("static-method")
    public Color getColor(TreePanel panel, TreeRow row, int position, boolean selected, boolean active) {
        return panel.getDefaultRowForeground(position, selected, active);
    }

    @Override
    public int compare(TreeRow r1, TreeRow r2) {
        return NumericComparator.caselessCompareStrings(getText(r1), getText(r2));
    }
}
