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

package com.trollworks.gcs.library;

import com.trollworks.gcs.ui.RetinaIcon;

import java.awt.Component;
import javax.swing.DefaultListCellRenderer;
import javax.swing.JList;

/** An item renderer for {@link LibraryExplorerRow}s. */
public class LibraryExplorerRowRenderer extends DefaultListCellRenderer {
    @Override
    public Component getListCellRendererComponent(JList<?> list, Object value, int index, boolean isSelected, boolean cellHasFocus) {
        String     title;
        RetinaIcon icon;
        if (value instanceof LibraryExplorerSearchResult) {
            LibraryExplorerSearchResult row = (LibraryExplorerSearchResult) value;
            title = row.getTitle();
            icon = row.getRow().getIcon();
        } else {
            title = value.toString();
            icon = null;
        }
        Component comp = super.getListCellRendererComponent(list, title, index, isSelected, cellHasFocus);
        if (icon != null) {
            setIcon(icon);
        }
        return comp;
    }
}
