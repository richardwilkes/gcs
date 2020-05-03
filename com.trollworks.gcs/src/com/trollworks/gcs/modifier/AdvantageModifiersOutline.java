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

public class AdvantageModifiersOutline extends ListOutline {
    /**
     * Create a new Advantage Modifiers outline.
     *
     * @param dataFile The owning data file.
     */
    public AdvantageModifiersOutline(DataFile dataFile) {
        this(dataFile, ((ListFile) dataFile).getModel());
    }

    /**
     * Create a new Advantage Modifiers outline.
     *
     * @param dataFile The owning data file.
     * @param model    The {@link OutlineModel} to use.
     */
    public AdvantageModifiersOutline(DataFile dataFile, OutlineModel model) {
        super(dataFile, model, AdvantageModifier.ID_LIST_CHANGED);
        AdvantageModifierColumnID.addColumns(this, !(dataFile instanceof AdvantageModifierList));
    }

    @Override
    protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, Row[] rows) {
        return !getModel().isLocked() && rows.length > 0 && rows[0] instanceof AdvantageModifier;
    }

    @Override
    public void convertDragRowsToSelf(List<Row> list) {
        OutlineModel model = getModel();
        for (Row row : model.getDragRows()) {
            AdvantageModifier modifier = new AdvantageModifier(mDataFile, (AdvantageModifier) row, true);
            model.collectRowsAndSetOwner(list, modifier, false);
        }
    }
}
