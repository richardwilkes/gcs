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

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

import java.awt.Container;
import java.util.List;
import javax.swing.JPanel;

/** An editor for melee weapon statistics. */
public class MeleeWeaponListEditor extends WeaponListEditor {
    private EditorField mReach;
    private EditorField mParry;
    private EditorField mBlock;

    /**
     * Creates a new {@link MeleeWeaponStats} editor.
     *
     * @param owner   The owning row.
     * @param weapons The weapons to modify.
     */
    public MeleeWeaponListEditor(ListRow owner, List<WeaponStats> weapons) {
        super(owner, weapons, MeleeWeaponStats.class);
    }

    @Override
    protected void createFields(Container parent) {
        JPanel panel = new JPanel(new PrecisionLayout().setMargins(0).setColumns(5));
        mReach = addField(parent, panel, "C-99**", I18n.Text("Reach"));
        mParry = addField(panel, panel, "+99**", I18n.Text("Parry Modifier"));
        mBlock = addField(panel, panel, "+99**", I18n.Text("Block Modifier"));
        parent.add(panel, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
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
    public String toString() {
        return I18n.Text("Melee Weapon");
    }
}
