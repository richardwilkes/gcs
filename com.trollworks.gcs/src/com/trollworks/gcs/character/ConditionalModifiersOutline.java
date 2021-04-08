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

package com.trollworks.gcs.character;

import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.Row;

import java.awt.dnd.DropTargetDragEvent;

public class ConditionalModifiersOutline extends Outline {
    public ConditionalModifiersOutline() {
        super(false);
        ConditionalModifierColumn.addColumns(this, false);
        OutlineModel model           = getModel();
        Column       conditionColumn = model.getColumnWithID(ConditionalModifierColumn.CONDITION.ordinal());
        conditionColumn.setSortCriteria(0, true);
        model.setHierarchyColumn(conditionColumn);
        setEnabled(false);
    }

    @Override
    protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, Row[] rows) {
        return false;
    }
}
