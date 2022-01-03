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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.character.CollectedModels;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.OutlineProxy;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.undo.MultipleUndo;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.util.ArrayList;
import java.util.List;
import javax.swing.undo.StateEdit;

/** Provides the "Toggle State" command. */
public class MoveEquipmentCommand extends Command {
    public static final String  CMD_TO_CARRIED_EQUIPMENT = "ToCarriedEquipment";
    public static final String  CMD_TO_OTHER_EQUIPMENT   = "ToOtherEquipment";
    private             boolean mToCarried;

    public MoveEquipmentCommand(boolean toCarried) {
        super(getTitle(toCarried), toCarried ? CMD_TO_CARRIED_EQUIPMENT : CMD_TO_OTHER_EQUIPMENT);
        mToCarried = toCarried;
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy proxy) {
            focus = proxy.getRealOutline();
        }
        if (focus instanceof EquipmentOutline outline) {
            DataFile     dataFile = outline.getDataFile();
            OutlineModel model    = outline.getModel();
            boolean      isOther  = model.getProperty(EquipmentList.KEY_OTHER_ROOT) != null;
            setEnabled((dataFile instanceof CollectedModels) && isOther == mToCarried && model.hasSelection());
        } else {
            setEnabled(false);
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy proxy) {
            focus = proxy.getRealOutline();
        }
        if (focus instanceof EquipmentOutline f) {
            moveSelection(f.getDataFile(), mToCarried);
        }
    }

    private static String getTitle(boolean toCarried) {
        return toCarried ? I18n.text("Move to Carried Equipment") : I18n.text("Move to Other Equipment");
    }

    public static void moveSelection(DataFile dataFile, boolean toCarried) {
        OutlineModel from = null;
        OutlineModel to   = null;
        if (dataFile instanceof CollectedModels cmodels) {
            from = toCarried ? cmodels.getOtherEquipmentModel() : cmodels.getEquipmentModel();
            to = toCarried ? cmodels.getEquipmentModel() : cmodels.getOtherEquipmentModel();
        }
        if (from != null && from.hasSelection()) {
            MultipleUndo undo = new MultipleUndo(getTitle(toCarried));
            ((ListOutline) from.getProperty(ListOutline.OWNING_LIST)).postUndo(undo);
            List<Row>     rows    = new ArrayList<>();
            List<ListRow> topRows = new ArrayList<>();
            StateEdit     edit1   = new StateEdit(from, getTitle(toCarried));
            StateEdit     edit2   = new StateEdit(to, getTitle(toCarried));
            to.setDragRows(from.getSelectionAsList(true).toArray(new Row[0]));
            ListOutline toOutline = (ListOutline) to.getProperty(ListOutline.OWNING_LIST);
            toOutline.convertDragRowsToSelf(rows);
            to.setDragRows(null);
            for (Row row : rows) {
                if (row.getDepth() == 0 && row instanceof ListRow lr) {
                    topRows.add(lr);
                }
            }
            toOutline.addRow(topRows.toArray(new ListRow[0]), getTitle(toCarried), true);
            edit1.end();
            edit2.end();
            undo.addEdit(edit1);
            undo.addEdit(edit2);
            undo.end();
            from.select(topRows, false);
            toOutline.scrollSelectionIntoView();
        }
    }
}
