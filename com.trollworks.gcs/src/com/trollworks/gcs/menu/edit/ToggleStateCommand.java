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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantageOutline;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.MultipleRowUndo;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.OutlineProxy;
import com.trollworks.gcs.ui.widget.outline.RowUndo;
import com.trollworks.gcs.utility.FilteredIterator;
import com.trollworks.gcs.utility.I18n;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.util.ArrayList;

/** Provides the "Toggle State" command. */
public class ToggleStateCommand extends Command {
    /** The action command this command will issue. */
    public static final String             CMD_TOGGLE_STATE = "ToggleState";
    /** The singleton {@link ToggleStateCommand}. */
    public static final ToggleStateCommand INSTANCE         = new ToggleStateCommand();

    private ToggleStateCommand() {
        super(I18n.Text("Toggle State"), CMD_TOGGLE_STATE, KeyEvent.VK_QUOTE);
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (focus instanceof AdvantageOutline) {
            ListOutline outline = (ListOutline) focus;
            setEnabled(outline.getDataFile() instanceof GURPSCharacter && outline.getModel().hasSelection());
            return;
        }
        if (focus instanceof EquipmentOutline) {
            ListOutline  outline = (ListOutline) focus;
            OutlineModel model   = outline.getModel();
            if (model.hasSelection() && model.getProperty(EquipmentList.TAG_OTHER_ROOT) == null && outline.getDataFile() instanceof GURPSCharacter) {
                setEnabled(true);
                return;
            }
        }
        setEnabled(false);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        ArrayList<RowUndo> undos = new ArrayList<>();
        if (focus instanceof EquipmentOutline) {
            EquipmentOutline outline = (EquipmentOutline) focus;
            for (Equipment equipment : new FilteredIterator<>(outline.getModel().getSelectionAsList(), Equipment.class)) {
                RowUndo undo = new RowUndo(equipment);
                equipment.setEquipped(!equipment.isEquipped());
                if (undo.finish()) {
                    undos.add(undo);
                }
            }
        } else if (focus instanceof AdvantageOutline) {
            AdvantageOutline outline = (AdvantageOutline) focus;
            for (Advantage adq : new FilteredIterator<>(outline.getModel().getSelectionAsList(), Advantage.class)) {
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
