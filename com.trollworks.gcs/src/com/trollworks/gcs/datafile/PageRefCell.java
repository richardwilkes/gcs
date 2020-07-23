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

package com.trollworks.gcs.datafile;

import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListTextCell;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.utility.I18n;

import javax.swing.SwingConstants;

public class PageRefCell extends ListTextCell {
    public static final String getStdToolTip(String type) {
        return String.format(I18n.Text("A reference to the book and page this %s appears on (e.g. B22 would refer to \"Basic Set\", page 22)"), type);
    }

    public static final String getStdCellToolTip(String text) {
        return (text.split("[, ]", 2).length == 1) ? null : text;
    }

    public PageRefCell() {
        super(SwingConstants.RIGHT, false);
    }

    @Override
    protected String getPresentationText(Outline outline, Row row, Column column) {
        String   text  = getData(row, column, false);
        String[] parts = text.split("[, ]", 2);
        if (parts.length == 1) {
            return parts[0];
        }
        return parts[0] + "+";
    }
}
