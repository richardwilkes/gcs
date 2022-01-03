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

package com.trollworks.gcs.ui.widget.outline;

import com.trollworks.gcs.ui.widget.LabelListCellRenderer;

import java.awt.Component;
import java.util.regex.Pattern;
import javax.swing.JList;

/** An item renderer for rows. */
public class RowItemRenderer extends LabelListCellRenderer<Object> {
    private static final Pattern TABS_OR_NEWLINES = Pattern.compile("[\\t\\n]");

    @Override
    public Component getListCellRendererComponent(JList<?> list, Object value, int index, boolean isSelected, boolean cellHasFocus) {
        String[] parts = TABS_OR_NEWLINES.split(value.toString().trim(), 2);
        String   text  = parts[0];
        if (text.length() > 100) {
            text = text.substring(0, 100) + "…";
        } else if (parts.length > 1) {
            text += "…";
        }
        Component comp = super.getListCellRendererComponent(list, text, index, isSelected, cellHasFocus);
        if (value instanceof ListRow) {
            setIcon(((ListRow) value).getIcon());
        }
        return comp;
    }
}
