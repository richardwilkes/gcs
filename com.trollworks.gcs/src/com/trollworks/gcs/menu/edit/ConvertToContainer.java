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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.MultipleRowUndo;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.OutlineProxy;
import com.trollworks.gcs.ui.widget.outline.RowUndo;
import com.trollworks.gcs.utility.I18n;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.util.ArrayList;

/** Provides the "Convert to Container" command. */
public class ConvertToContainer extends Command {
    /** The action command this command will issue. */
    public static final String             CMD_CONVERT_TO_CONTAINER = "ConvertToContainer";
    /** The singleton {@link ConvertToContainer}. */
    public static final ConvertToContainer INSTANCE                 = new ConvertToContainer();

    public ConvertToContainer() {
        super(I18n.Text("Convert to Container"), CMD_CONVERT_TO_CONTAINER);
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (!(focus instanceof EquipmentOutline)) {
            setEnabled(false);
            return;
        }
        OutlineModel model = ((EquipmentOutline) focus).getModel();
        setEnabled(!model.isLocked() && model.getSelectionCount() == 1 && !model.getSelectionAsList().get(0).canHaveChildren());
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        ListOutline  outline = (ListOutline) focus;
        OutlineModel model   = outline.getModel();
        if (!model.isLocked() && model.getSelectionCount() == 1) {
            Equipment equipment = (Equipment) model.getSelectionAsList().get(0);
            if (!equipment.canHaveChildren()) {
                RowUndo undo = new RowUndo(equipment);
                equipment.setQuantity(1);
                equipment.setCanHaveChildren(true);
                undo.finish();
                ArrayList<RowUndo> undos = new ArrayList<>();
                undos.add(undo);
                outline.repaintSelection();
                new MultipleRowUndo(undos);
            }
        }
    }
}
