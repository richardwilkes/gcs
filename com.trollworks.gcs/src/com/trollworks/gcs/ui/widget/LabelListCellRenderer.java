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

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.ui.border.EmptyBorder;

import java.awt.Component;
import javax.swing.JList;
import javax.swing.ListCellRenderer;

public class LabelListCellRenderer<E> extends Label implements ListCellRenderer<E> {
    public LabelListCellRenderer() {
        super("");
        setOpaque(true);
    }

    @Override
    public Component getListCellRendererComponent(JList<? extends E> list, E value, int index, boolean isSelected, boolean cellHasFocus) {
        setBackground(isSelected ? list.getSelectionBackground() : list.getBackground());
        setForeground(isSelected ? list.getSelectionForeground() : list.getForeground());
        setText(value == null ? "" : value.toString());
        setEnabled(list.isEnabled());
        setFont(list.getFont());
        setBorder(new EmptyBorder(2, 4, 2, 4));
        return this;
    }
}
