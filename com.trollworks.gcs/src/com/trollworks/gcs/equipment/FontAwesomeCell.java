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

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListHeaderCell;
import com.trollworks.gcs.ui.widget.outline.Row;

import java.awt.Font;
import javax.swing.SwingConstants;

public class FontAwesomeCell extends ListHeaderCell {
    private String mFontName;

    public FontAwesomeCell(String fontName, boolean forSheet) {
        super(forSheet);
        mFontName = fontName;
    }

    @Override
    protected Font deriveFont(Row row, Column column, Font font) {
        if (row != null) {
            return font;
        }
        return new Font(mFontName, font.getStyle(), (int) Math.round(font.getSize() * 0.9));
    }

    @Override
    public int getVAlignment() {
        return SwingConstants.CENTER;
    }
}
