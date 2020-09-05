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

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;

/** Provides the "New Equipment" command. */
public class NewEquipmentCommand extends Command {
    /** The action command this command will issue. */
    public static final String              CMD_NEW_EQUIPMENT                 = "NewEquipment";
    /** The action command this command will issue. */
    public static final String              CMD_NEW_EQUIPMENT_CONTAINER       = "NewEquipmentContainer";
    /** The action command this command will issue. */
    public static final String              CMD_NEW_OTHER_EQUIPMENT           = "NewOtherEquipment";
    /** The action command this command will issue. */
    public static final String              CMD_NEW_OTHER_EQUIPMENT_CONTAINER = "NewOtherEquipmentContainer";
    /** The "New Carried Equipment" command. */
    public static final NewEquipmentCommand CARRIED_INSTANCE                  = new NewEquipmentCommand(true, false, I18n.Text("New Carried Equipment"), CMD_NEW_EQUIPMENT, COMMAND_MODIFIER);
    /** The "New Carried Equipment Container" command. */
    public static final NewEquipmentCommand CARRIED_CONTAINER_INSTANCE        = new NewEquipmentCommand(true, true, I18n.Text("New Carried Equipment Container"), CMD_NEW_EQUIPMENT_CONTAINER, SHIFTED_COMMAND_MODIFIER);
    /** The "New Other Equipment" command. */
    public static final NewEquipmentCommand NOT_CARRIED_INSTANCE              = new NewEquipmentCommand(false, false, I18n.Text("New Other Equipment"), CMD_NEW_OTHER_EQUIPMENT, COMMAND_MODIFIER | InputEvent.ALT_DOWN_MASK);
    /** The "New Other Equipment Container" command. */
    public static final NewEquipmentCommand NOT_CARRIED_CONTAINER_INSTANCE    = new NewEquipmentCommand(false, true, I18n.Text("New Other Equipment Container"), CMD_NEW_OTHER_EQUIPMENT_CONTAINER, SHIFTED_COMMAND_MODIFIER | InputEvent.ALT_DOWN_MASK);
    private             boolean             mCarried;
    private             boolean             mContainer;

    private NewEquipmentCommand(boolean carried, boolean container, String title, String cmd, int modifiers) {
        super(title, cmd, KeyEvent.VK_E, modifiers);
        mCarried = carried;
        mContainer = container;
    }

    @Override
    public void adjust() {
        if (mCarried) {
            EquipmentDockable equipment = getTarget(EquipmentDockable.class);
            if (equipment != null) {
                setEnabled(!equipment.getOutline().getModel().isLocked());
                return;
            }
        }
        if (getTarget(SheetDockable.class) != null) {
            setEnabled(true);
        } else {
            setEnabled(getTarget(TemplateDockable.class) != null);
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        if (mCarried) {
            EquipmentDockable equipment = getTarget(EquipmentDockable.class);
            if (equipment != null) {
                ListOutline outline = equipment.getOutline();
                if (!outline.getModel().isLocked()) {
                    perform(equipment.getDataFile(), outline);
                }
                return;
            }
        }
        SheetDockable sheet = getTarget(SheetDockable.class);
        if (sheet != null) {
            ListOutline outline;
            outline = mCarried ? sheet.getSheet().getEquipmentOutline() : sheet.getSheet().getOtherEquipmentOutline();
            perform(sheet.getDataFile(), outline);
            return;
        }
        TemplateDockable template = getTarget(TemplateDockable.class);
        if (template != null) {
            ListOutline outline;
            outline = mCarried ? template.getTemplate().getEquipmentOutline() : template.getTemplate().getOtherEquipmentOutline();
            perform(template.getDataFile(), outline);
        }
    }

    private void perform(DataFile dataFile, ListOutline outline) {
        Equipment equipment = new Equipment(dataFile, mContainer);
        outline.addRow(equipment, getTitle(), false);
        outline.getModel().select(equipment, false);
        outline.scrollSelectionIntoView();
        outline.openDetailEditor(true);
    }
}
