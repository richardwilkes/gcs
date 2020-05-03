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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.Row;

import java.awt.dnd.DropTargetDragEvent;
import java.util.List;

public class EquipmentModifiersOutline extends ListOutline {
    /**
     * Create a new Equipment Modifiers outline.
     *
     * @param dataFile The owning data file.
     */
    public EquipmentModifiersOutline(DataFile dataFile) {
        this(dataFile, ((ListFile) dataFile).getModel());
    }

    /**
     * Create a new Equipment Modifiers outline.
     *
     * @param dataFile The owning data file.
     * @param model    The {@link OutlineModel} to use.
     */
    public EquipmentModifiersOutline(DataFile dataFile, OutlineModel model) {
        super(dataFile, model, EquipmentModifier.ID_LIST_CHANGED);
        EquipmentModifierColumnID.addColumns(this, !(dataFile instanceof EquipmentModifierList));
    }

    @Override
    protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, Row[] rows) {
        return !getModel().isLocked() && rows.length > 0 && rows[0] instanceof EquipmentModifier;
    }

    @Override
    public void convertDragRowsToSelf(List<Row> list) {
        OutlineModel model = getModel();
        for (Row row : model.getDragRows()) {
            EquipmentModifier modifier = new EquipmentModifier(mDataFile, (EquipmentModifier) row, true);
            model.collectRowsAndSetOwner(list, modifier, false);
        }
    }
}
