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

/** An editor for ranged weapon statistics. */
public class RangedWeaponEditor extends WeaponEditor {
    private EditorField mAccuracy;
    private EditorField mRange;
    private EditorField mRateOfFire;
    private EditorField mShots;
    private EditorField mBulk;
    private EditorField mRecoil;

    /**
     * Creates a new ranged weapon editor for the specified row.
     *
     * @param row The row to edit ranged weapon statistics for.
     * @return The editor, or {@code null} if the row is not appropriate.
     */
    public static RangedWeaponEditor createEditor(ListRow row) {
        if (row instanceof Equipment) {
            return new RangedWeaponEditor(row, ((Equipment) row).getWeapons());
        } else if (row instanceof Advantage) {
            return new RangedWeaponEditor(row, ((Advantage) row).getWeapons());
        } else if (row instanceof Spell) {
            return new RangedWeaponEditor(row, ((Spell) row).getWeapons());
        } else if (row instanceof Skill) {
            return new RangedWeaponEditor(row, ((Skill) row).getWeapons());
        }
        return null;
    }

    /**
     * Creates a new {@link RangedWeaponStats} editor.
     *
     * @param owner   The owning row.
     * @param weapons The weapons to modify.
     */
    public RangedWeaponEditor(ListRow owner, List<WeaponStats> weapons) {
        super(owner, weapons, RangedWeaponStats.class);
    }

    @Override
    protected void createFields(Container parent) {
        JPanel panel   = new JPanel(new ColumnLayout(5));
        String tooltip = I18n.Text("Accuracy");
        mAccuracy = createTextField("99+99*", tooltip);
        parent.add(new LinkedLabel(tooltip, mAccuracy));
        panel.add(mAccuracy);
        tooltip = I18n.Text("Rate of Fire");
        mRateOfFire = createTextField("999*", tooltip);
        panel.add(new LinkedLabel(tooltip, mRateOfFire));
        panel.add(mRateOfFire);
        tooltip = I18n.Text("Range");
        mRange = createTextField(null, tooltip);
        panel.add(new LinkedLabel(tooltip, mRange));
        panel.add(mRange);
        parent.add(panel);

        panel = new JPanel(new ColumnLayout(5));
        tooltip = I18n.Text("Recoil");
        mRecoil = createTextField("9999", tooltip);
        parent.add(new LinkedLabel(tooltip, mRecoil));
        panel.add(mRecoil);
        tooltip = I18n.Text("Shots");
        mShots = createTextField(null, tooltip);
        panel.add(new LinkedLabel(tooltip, mShots));
        panel.add(mShots);
        tooltip = I18n.Text("Bulk");
        mBulk = createTextField("9999", tooltip);
        panel.add(new LinkedLabel(tooltip, mBulk));
        panel.add(mBulk);
        parent.add(panel);
    }

    @Override
    protected void updateFromField(Object source) {
        if (mAccuracy == source) {
            changeAccuracy();
        } else if (mRange == source) {
            changeRange();
        } else if (mRateOfFire == source) {
            changeRateOfFire();
        } else if (mShots == source) {
            changeShots();
        } else if (mBulk == source) {
            changeBulk();
        } else if (mRecoil == source) {
            changeRecoil();
        }
    }

    private void changeAccuracy() {
        ((RangedWeaponStats) getWeapon()).setAccuracy((String) mAccuracy.getValue());
        adjustOutlineToContent();
    }

    private void changeRange() {
        ((RangedWeaponStats) getWeapon()).setRange((String) mRange.getValue());
        adjustOutlineToContent();
    }

    private void changeRateOfFire() {
        ((RangedWeaponStats) getWeapon()).setRateOfFire((String) mRateOfFire.getValue());
        adjustOutlineToContent();
    }

    private void changeShots() {
        ((RangedWeaponStats) getWeapon()).setShots((String) mShots.getValue());
        adjustOutlineToContent();
    }

    private void changeBulk() {
        ((RangedWeaponStats) getWeapon()).setBulk((String) mBulk.getValue());
        adjustOutlineToContent();
    }

    private void changeRecoil() {
        ((RangedWeaponStats) getWeapon()).setRecoil((String) mRecoil.getValue());
        adjustOutlineToContent();
    }

    @Override
    protected WeaponStats createWeaponStats() {
        return new RangedWeaponStats(getOwner());
    }

    @Override
    protected void updateFields() {
        RangedWeaponStats weapon = (RangedWeaponStats) getWeapon();
        mAccuracy.setValue(weapon.getAccuracy());
        mRange.setValue(weapon.getRange());
        mRateOfFire.setValue(weapon.getRateOfFire());
        mShots.setValue(weapon.getShots());
        mBulk.setValue(weapon.getBulk());
        mRecoil.setValue(weapon.getRecoil());
        super.updateFields();
    }

    @Override
    protected void enableFields(boolean enabled) {
        mAccuracy.setEnabled(enabled);
        mRange.setEnabled(enabled);
        mRateOfFire.setEnabled(enabled);
        mShots.setEnabled(enabled);
        mBulk.setEnabled(enabled);
        mRecoil.setEnabled(enabled);
        super.enableFields(enabled);
    }

    @Override
    protected void blankFields() {
        mAccuracy.setValue("");
        mRange.setValue("");
        mRateOfFire.setValue("");
        mShots.setValue("");
        mBulk.setValue("");
        mRecoil.setValue("");
        super.blankFields();
    }

    @Override
    public String toString() {
        return I18n.Text("Ranged Weapon");
    }
}
