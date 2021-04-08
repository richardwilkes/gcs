/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.skill;

import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListTextCell;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.Row;

import java.awt.Color;
import javax.swing.SwingConstants;

/** Provides a grayed-out point value for skill containers. */
public class SkillPointsTextCell extends ListTextCell {
    /** Creates a new {@link SkillPointsTextCell}. */
    public SkillPointsTextCell() {
        super(SwingConstants.RIGHT, false);
    }

    @Override
    public Color getColor(Outline outline, Row row, Column column, boolean selected, boolean active) {
        Color color = super.getColor(outline, row, column, selected, active);
        if (!selected && row.canHaveChildren()) {
            color = new Color(color.getRed(), color.getGreen(), color.getBlue(), color.getAlpha() / 2);
        }
        return color;
    }
}
