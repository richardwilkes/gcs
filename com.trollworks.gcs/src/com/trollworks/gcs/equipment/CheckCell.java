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

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListTextCell;
import com.trollworks.gcs.ui.widget.outline.MultipleRowUndo;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowUndo;

import java.awt.Font;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.util.ArrayList;
import javax.swing.SwingConstants;

public class CheckCell extends ListTextCell {
    public CheckCell(int alignment, boolean wrapped) {
        super(alignment, wrapped);
    }

    @Override
    protected Font deriveFont(Row row, Column column, Font font) {
        return new Font(Fonts.FONT_AWESOME_SOLID, font.getStyle(), (int) Math.round(font.getSize() * 0.9));
    }

    @Override
    public int getVAlignment() {
        return SwingConstants.CENTER;
    }

    @Override
    public void mouseClicked(MouseEvent event, Rectangle bounds, Row row, Column column) {
        if (row instanceof Equipment equipment) {
            RowUndo undo = new RowUndo(equipment);
            equipment.setEquipped(!equipment.isEquipped());
            if (undo.finish()) {
                ArrayList<RowUndo> list = new ArrayList<>();
                list.add(undo);
                new MultipleRowUndo(list);
            }
        }
    }
}
