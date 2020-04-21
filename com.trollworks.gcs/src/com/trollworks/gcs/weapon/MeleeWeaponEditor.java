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

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

import java.awt.Container;
import java.util.List;
import javax.swing.JPanel;

/** An editor for melee weapon statistics. */
public class MeleeWeaponEditor extends WeaponEditor {
    private EditorField mReach;
    private EditorField mParry;
    private EditorField mBlock;

    /**
     * Creates a new melee weapon editor for the specified row.
     *
     * @param row The row to edit melee weapon statistics for.
     * @return The editor, or {@code null} if the row is not appropriate.
     */
    public static MeleeWeaponEditor createEditor(ListRow row) {
        if (row instanceof Equipment) {
            return new MeleeWeaponEditor(row, ((Equipment) row).getWeapons());
        } else if (row instanceof Advantage) {
            return new MeleeWeaponEditor(row, ((Advantage) row).getWeapons());
        } else if (row instanceof Spell) {
            return new MeleeWeaponEditor(row, ((Spell) row).getWeapons());
        } else if (row instanceof Skill) {
            return new MeleeWeaponEditor(row, ((Skill) row).getWeapons());
        }
        return null;
    }

    /**
     * Creates a new {@link MeleeWeaponStats} editor.
     *
     * @param owner   The owning row.
     * @param weapons The weapons to modify.
     */
    public MeleeWeaponEditor(ListRow owner, List<WeaponStats> weapons) {
        super(owner, weapons, MeleeWeaponStats.class);
    }

    @Override
    protected void createFields(Container parent) {
        JPanel panel   = new JPanel(new ColumnLayout(5));
        String tooltip = I18n.Text("Reach");
        mReach = createTextField("C-99**", tooltip);
        parent.add(new LinkedLabel(tooltip, mReach));
        panel.add(mReach);
        tooltip = I18n.Text("Parry Modifier");
        mParry = createTextField("+99**", tooltip);
        panel.add(new LinkedLabel(tooltip, mParry));
        panel.add(mParry);
        tooltip = I18n.Text("Block Modifier");
        mBlock = createTextField("+99**", tooltip);
        panel.add(new LinkedLabel(tooltip, mBlock));
        panel.add(mBlock);
        parent.add(panel);
    }

    @Override
    protected void updateFromField(Object source) {
        if (mReach == source) {
            changeReach();
        } else if (mParry == source) {
            changeParry();
        } else if (mBlock == source) {
            changeBlock();
        }
    }

    private void changeReach() {
        ((MeleeWeaponStats) getWeapon()).setReach((String) mReach.getValue());
        adjustOutlineToContent();
    }

    private void changeParry() {
        ((MeleeWeaponStats) getWeapon()).setParry((String) mParry.getValue());
        adjustOutlineToContent();
    }

    private void changeBlock() {
        ((MeleeWeaponStats) getWeapon()).setBlock((String) mBlock.getValue());
        adjustOutlineToContent();
    }

    @Override
    protected WeaponStats createWeaponStats() {
        return new MeleeWeaponStats(getOwner());
    }

    @Override
    protected void updateFields() {
        MeleeWeaponStats weapon = (MeleeWeaponStats) getWeapon();
        mReach.setValue(weapon.getReach());
        mParry.setValue(weapon.getParry());
        mBlock.setValue(weapon.getBlock());
        super.updateFields();
    }

    @Override
    protected void enableFields(boolean enabled) {
        mReach.setEnabled(enabled);
        mParry.setEnabled(enabled);
        mBlock.setEnabled(enabled);
        super.enableFields(enabled);
    }

    @Override
    protected void blankFields() {
        mReach.setValue("");
        mParry.setValue("");
        mBlock.setValue("");
        super.blankFields();
    }

    @Override
    public String toString() {
        return I18n.Text("Melee Weapon");
    }
}
