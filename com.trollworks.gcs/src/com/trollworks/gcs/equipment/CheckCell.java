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

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListTextCell;
import com.trollworks.gcs.ui.widget.outline.MultipleRowUndo;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowUndo;

import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.util.ArrayList;

public class CheckCell extends ListTextCell {
    public CheckCell(int alignment, boolean wrapped) {
        super(alignment, wrapped);

    }

    @SuppressWarnings("unused")
    @Override
    public void mouseClicked(MouseEvent event, Rectangle bounds, Row row, Column column) {
        if (row instanceof Equipment) {
            Equipment equipment = (Equipment) row;
            RowUndo   undo      = new RowUndo(equipment);
            equipment.setEquipped(!equipment.isEquipped());
            if (undo.finish()) {
                ArrayList<RowUndo> list = new ArrayList<>();
                list.add(undo);
                new MultipleRowUndo(list);
            }
        }
    }
}
