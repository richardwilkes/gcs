/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantageOutline;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.equipment.EquipmentState;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.MultipleRowUndo;
import com.trollworks.gcs.widgets.outline.RowUndo;
import com.trollworks.toolkit.collections.FilteredIterator;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.widget.outline.OutlineProxy;
import com.trollworks.toolkit.utility.I18n;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.util.ArrayList;

/** Provides the "Rotate State" command. */
public class RotateStateCommand extends Command {
    /** The action command this command will issue. */
    public static final String             CMD_ROTATE_STATE = "RotateState";
    /** The singleton {@link RotateStateCommand}. */
    public static final RotateStateCommand INSTANCE         = new RotateStateCommand();

    private RotateStateCommand() {
        super(I18n.Text("Rotate State"), CMD_ROTATE_STATE, KeyEvent.VK_QUOTE);
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (focus instanceof EquipmentOutline || focus instanceof AdvantageOutline) {
            ListOutline outline = (ListOutline) focus;
            setEnabled(outline.getDataFile() instanceof GURPSCharacter && outline.getModel().hasSelection());
        } else {
            setEnabled(false);
        }
    }

    @SuppressWarnings("unused")
    @Override
    public void actionPerformed(ActionEvent event) {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        ArrayList<RowUndo> undos = new ArrayList<RowUndo>();
        if (focus instanceof EquipmentOutline) {
            EquipmentOutline outline = (EquipmentOutline) focus;
            for (Equipment equipment : new FilteredIterator<Equipment>(outline.getModel().getSelectionAsList(), Equipment.class)) {
                RowUndo          undo   = new RowUndo(equipment);
                EquipmentState[] values = EquipmentState.values();
                int              index  = equipment.getState().ordinal() - 1;
                if (index < 0) {
                    index = values.length - 1;
                }
                equipment.setState(values[index]);
                if (undo.finish()) {
                    undos.add(undo);
                }
            }
        } else if (focus instanceof AdvantageOutline) {
            AdvantageOutline outline = (AdvantageOutline) focus;
            for (Advantage adq : new FilteredIterator<Advantage>(outline.getModel().getSelectionAsList(), Advantage.class)) {
                RowUndo undo = new RowUndo(adq);
                adq.setEnabled(!adq.isSelfEnabled());
                if (undo.finish()) {
                    undos.add(undo);
                }
            }
        }
        if (!undos.isEmpty()) {
            ((ListOutline) focus).repaintSelection();
            new MultipleRowUndo(undos);
        }
    }
}
