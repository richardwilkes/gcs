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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.ui.widget.outline.Outline;

import java.util.List;

/** Editor for {@link AdvantageModifierList}s. */
public class AdvantageModifierListEditor extends ModifierListEditor {
    /**
     * @param advantage The {@link Advantage} to edit.
     * @return An instance of {@link AdvantageModifierListEditor}.
     */
    public static AdvantageModifierListEditor createEditor(Advantage advantage) {
        return new AdvantageModifierListEditor(advantage.getDataFile(), advantage.getParent() != null ? ((Advantage) advantage.getParent()).getAllModifiers() : null, advantage.getModifiers());
    }

    /**
     * Creates a new {@link AdvantageModifierListEditor} editor.
     *
     * @param owner             The owning row.
     * @param readOnlyModifiers The list of {@link AdvantageModifier}s from parents, which are not
     *                          to be modified.
     * @param modifiers         The list of {@link AdvantageModifier}s to modify.
     */
    public AdvantageModifierListEditor(DataFile owner, List<AdvantageModifier> readOnlyModifiers, List<AdvantageModifier> modifiers) {
        super(owner, readOnlyModifiers, modifiers);
    }

    @Override
    protected void addColumns(Outline outline) {
        AdvantageModifierColumnID.addColumns(outline, true);
    }

    @Override
    protected Modifier createModifier(DataFile owner) {
        return new AdvantageModifier(owner, false);
    }
}
