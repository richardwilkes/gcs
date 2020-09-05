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

package com.trollworks.gcs.page;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.TitledBorder;

import java.awt.Color;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Insets;
import java.awt.LayoutManager;
import java.awt.Rectangle;
import java.util.HashMap;
import java.util.Map;
import java.util.Map.Entry;
import javax.swing.JPanel;
import javax.swing.SwingConstants;
import javax.swing.UIManager;
import javax.swing.border.CompoundBorder;

/** A standard panel with a drop shadow. */
public class DropPanel extends JPanel {
    private Map<Component, Color> mHorizontalBackgrounds = new HashMap<>();
    private Map<Component, Color> mVerticalBackgrounds   = new HashMap<>();
    private TitledBorder          mTitledBorder;
    private boolean               mOnlyReportPreferredSize;

    /**
     * Creates a standard panel with a drop shadow.
     *
     * @param layout The layout to use.
     */
    public DropPanel(LayoutManager layout) {
        this(layout, false);
    }

    /**
     * Creates a standard panel with a drop shadow.
     *
     * @param layout                  The layout to use.
     * @param onlyReportPreferredSize Whether or not minimum and maximum size is reported as
     *                                preferred size or not.
     */
    public DropPanel(LayoutManager layout, boolean onlyReportPreferredSize) {
        this(layout, null, null, onlyReportPreferredSize);
    }

    /**
     * Creates a standard panel with a drop shadow.
     *
     * @param layout The layout to use.
     * @param title  The title to use.
     */
    public DropPanel(LayoutManager layout, String title) {
        this(layout, title, UIManager.getFont(Fonts.KEY_LABEL_PRIMARY), false);
    }

    /**
     * Creates a standard panel with a drop shadow.
     *
     * @param layout                  The layout to use.
     * @param title                   The title to use.
     * @param onlyReportPreferredSize Whether or not minimum and maximum size is reported as
     *                                preferred size or not.
     */
    public DropPanel(LayoutManager layout, String title, boolean onlyReportPreferredSize) {
        this(layout, title, UIManager.getFont(Fonts.KEY_LABEL_PRIMARY), onlyReportPreferredSize);
    }

    /**
     * Creates a standard panel with a drop shadow.
     *
     * @param layout                  The layout to use.
     * @param title                   The title to use.
     * @param font                    The font to use for the title.
     * @param onlyReportPreferredSize Whether or not minimum and maximum size is reported as
     *                                preferred size or not.
     */
    public DropPanel(LayoutManager layout, String title, Font font, boolean onlyReportPreferredSize) {
        super(layout);
        setOpaque(true);
        setBackground(ThemeColor.PAGE);
        mTitledBorder = new TitledBorder(font, title);
        setBorder(new CompoundBorder(mTitledBorder, new EmptyBorder(0, 2, 1, 2)));
        setAlignmentY(TOP_ALIGNMENT);
        mOnlyReportPreferredSize = onlyReportPreferredSize;
    }

    @Override
    public Dimension getMinimumSize() {
        Dimension size     = mOnlyReportPreferredSize ? getPreferredSize() : super.getMinimumSize();
        int       minWidth = mTitledBorder.getMinimumWidth(this);
        if (size.width < minWidth) {
            size.width = minWidth;
        }
        return size;
    }

    @Override
    public Dimension getPreferredSize() {
        Dimension size     = super.getPreferredSize();
        int       minWidth = mTitledBorder.getMinimumWidth(this);
        if (size.width < minWidth) {
            size.width = minWidth;
        }
        return size;
    }

    @Override
    public Dimension getMaximumSize() {
        Dimension size     = mOnlyReportPreferredSize ? getPreferredSize() : super.getMaximumSize();
        int       minWidth = mTitledBorder.getMinimumWidth(this);
        if (size.width < minWidth) {
            size.width = minWidth;
        }
        return size;
    }

    /**
     * Marks an area with a specific background color. The panel specified will be used to calculate
     * the area's top and bottom, and the background color will span the width of the drop panel.
     *
     * @param panel      The panel to attach the color to.
     * @param background The color to attach.
     */
    public void addHorizontalBackground(Component panel, Color background) {
        mHorizontalBackgrounds.put(panel, background);
    }

    /**
     * Removes a horizontal background added with {@link #addHorizontalBackground(Component,
     * Color)}.
     *
     * @param panel The panel to remove.
     */
    public void removeHorizontalBackground(Component panel) {
        mHorizontalBackgrounds.remove(panel);
    }

    /**
     * Marks an area with a specific background color. The panel specified will be used to calculate
     * the area's left and right, and the background color will span the height of the drop panel.
     *
     * @param panel      The panel to attach the color to.
     * @param background The color to attach.
     */
    public void addVerticalBackground(Component panel, Color background) {
        mVerticalBackgrounds.put(panel, background);
    }

    /**
     * Removes a vertical background added with {@link #addVerticalBackground(Component, Color)}.
     *
     * @param panel The panel to remove.
     */
    public void removeVerticalBackground(Component panel) {
        mVerticalBackgrounds.remove(panel);
    }

    @Override
    protected void paintComponent(Graphics gc) {
        super.paintComponent(GraphicsUtilities.prepare(gc));
        Insets    insets      = mTitledBorder.getBorderInsets(this);
        Rectangle localBounds = getBounds();
        localBounds.x = insets.left;
        localBounds.y = insets.top;
        localBounds.width -= insets.left + insets.right;
        localBounds.height -= insets.top + insets.bottom;
        paintHorizontalBackgrounds(gc, localBounds);
        paintVerticalBackgrounds(gc, localBounds);
    }

    private void paintHorizontalBackgrounds(Graphics gc, Rectangle localBounds) {
        for (Entry<Component, Color> entry : mHorizontalBackgrounds.entrySet()) {
            Component panel  = entry.getKey();
            Rectangle bounds = panel.getBounds();
            Container parent = panel.getParent();
            if (parent != null) {
                if (parent != this) {
                    UIUtilities.convertRectangle(bounds, parent, this);
                }
                gc.setColor(entry.getValue());
                gc.fillRect(localBounds.x, bounds.y, localBounds.width, bounds.height);
            }
        }
    }

    private void paintVerticalBackgrounds(Graphics gc, Rectangle localBounds) {
        for (Entry<Component, Color> entry : mVerticalBackgrounds.entrySet()) {
            Component panel  = entry.getKey();
            Rectangle bounds = panel.getBounds();
            Container parent = panel.getParent();
            if (parent != null) {
                if (parent != this) {
                    UIUtilities.convertRectangle(bounds, parent, this);
                }
                gc.setColor(entry.getValue());
                gc.fillRect(bounds.x, localBounds.y, bounds.width, localBounds.height);
            }
        }
    }

    /** @return The {@link TitledBorder}. */
    public TitledBorder getTitledBorder() {
        return mTitledBorder;
    }

    /**
     * @param parent    The parent to use.
     * @param sheet     The {@link CharacterSheet} to use.
     * @param key       The notification ID to use.
     * @param title     The title to use.
     * @param tooltip   The tooltip to use.
     */
    @SuppressWarnings("static-method")
    protected void createLabelAndField(Container parent, CharacterSheet sheet, String key, String title, String tooltip) {
        PageField field = new PageField(sheet, key, SwingConstants.LEFT, true, tooltip);
        parent.add(new PageLabel(title, field));
        parent.add(field);
    }

    /**
     * @param parent    The parent to use.
     * @param sheet     The {@link CharacterSheet} to use.
     * @param key       The notification ID to use.
     * @param title     The title to use.
     * @param tooltip   The tooltip to use.
     */
    @SuppressWarnings("static-method")
    protected void createLabelAndDisabledField(Container parent, CharacterSheet sheet, String key, String title, String tooltip) {
        PageField field = new PageField(sheet, key, SwingConstants.LEFT, false, tooltip);
        parent.add(new PageLabel(title, field));
        parent.add(field);
    }

    /**
     * @param parent  The parent to use.
     * @param title   The title to use.
     * @param tooltip The tooltip to use.
     * @return The newly created header.
     */
    @SuppressWarnings("static-method")
    protected PageHeader createHeader(Container parent, String title, String tooltip) {
        PageHeader header = new PageHeader(title, tooltip);
        parent.add(header);
        return header;
    }
}
