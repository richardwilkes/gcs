/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.library;

import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.TextCell;

import javax.swing.Icon;

public class LibraryExplorerCell extends TextCell {
    @Override
    public ThemeFont getThemeFont(Row row, Column column) {
        return Fonts.LABEL_PRIMARY;
    }

    @Override
    public Icon getIcon(Row row, Column column) {
        return row.getIcon(column);
    }
}
