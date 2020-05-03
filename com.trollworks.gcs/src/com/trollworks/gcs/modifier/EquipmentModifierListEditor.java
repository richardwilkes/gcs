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
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.ui.widget.outline.Outline;

import java.util.ArrayList;
import java.util.List;

/** Editor for {@link EquipmentModifierList}s. */
public class EquipmentModifierListEditor extends ModifierListEditor {
    /**
     * @param equipment The {@link Equipment} to edit.
     * @return An instance of {@link EquipmentModifierListEditor}.
     */
    public static EquipmentModifierListEditor createEditor(Equipment equipment) {
        return new EquipmentModifierListEditor(equipment.getDataFile(), equipment.getModifiers());
    }

    /**
     * Creates a new {@link EquipmentModifierListEditor} editor.
     *
     * @param owner     The owning row.
     * @param modifiers The list of {@link EquipmentModifier}s to modify.
     */
    public EquipmentModifierListEditor(DataFile owner, List<EquipmentModifier> modifiers) {
        super(owner, new ArrayList<EquipmentModifier>(), modifiers);
    }

    @Override
    protected void addColumns(Outline outline) {
        EquipmentModifierColumnID.addColumns(outline, true);
    }

    @Override
    protected Modifier createModifier(DataFile owner) {
        return new EquipmentModifier(owner, false);
    }
}
