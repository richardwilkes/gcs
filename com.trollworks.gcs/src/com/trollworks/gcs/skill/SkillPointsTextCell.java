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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListTextCell;
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
    public Color getColor(boolean selected, boolean active, Row row, Column column) {
        if (!selected && row.canHaveChildren()) {
            return Color.LIGHT_GRAY;
        }
        return super.getColor(selected, active, row, column);
    }
}
