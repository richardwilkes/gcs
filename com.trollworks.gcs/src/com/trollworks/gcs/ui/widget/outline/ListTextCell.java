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

import com.trollworks.gcs.ui.Fonts;

import java.awt.Color;
import java.awt.Font;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import javax.swing.UIManager;

/** Represents cells in a {@link Outline}. */
public class ListTextCell extends TextCell {
    /**
     * Create a new text cell.
     *
     * @param alignment The horizontal text alignment to use.
     * @param wrapped   Pass in {@code true} to enable wrapping.
     */
    public ListTextCell(int alignment, boolean wrapped) {
        super(alignment, wrapped);
    }

    @Override
    public Font getFont(Row row, Column column) {
        return UIManager.getFont(Fonts.KEY_FIELD_PRIMARY);
    }

    @Override
    public Color getColor(boolean selected, boolean active, Row row, Column column) {
        if (row instanceof ListRow && !((ListRow) row).isSatisfied()) {
            return Color.red;
        }
        return super.getColor(selected, active, row, column);
    }

    @Override
    public String getToolTipText(Outline outline, MouseEvent event, Rectangle bounds, Row row, Column column) {
        if (!(row instanceof ListRow) || ((ListRow) row).isSatisfied()) {
            return super.getToolTipText(outline, event, bounds, row, column);
        }
        return ((ListRow) row).getReasonForUnsatisfied();
    }
}
