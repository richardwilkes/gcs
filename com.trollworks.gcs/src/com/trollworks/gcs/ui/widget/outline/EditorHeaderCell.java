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

package com.trollworks.gcs.ui.widget.outline;

import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.ThemeColor;

import java.awt.Color;
import java.awt.Font;

/** Used to draw headers in the lists. */
public class EditorHeaderCell extends HeaderCell {
    @Override
    public Font getFont(Row row, Column column) {
        return Fonts.getDefaultSystemFont();
    }

    @Override
    public Color getColor(Outline outline, Row row, Column column, boolean selected, boolean active) {
        return ThemeColor.ON_HEADER;
    }
}
