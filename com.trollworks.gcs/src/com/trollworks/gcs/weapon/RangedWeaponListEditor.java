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
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

import java.awt.Container;
import java.util.List;

/** An editor for ranged weapon statistics. */
public class RangedWeaponListEditor extends WeaponListEditor {
    private EditorField mAccuracy;
    private EditorField mRange;
    private EditorField mRateOfFire;
    private EditorField mShots;
    private EditorField mBulk;
    private EditorField mRecoil;

    /**
     * Creates a new {@link RangedWeaponStats} editor.
     *
     * @param owner   The owning row.
     * @param weapons The weapons to modify.
     */
    public RangedWeaponListEditor(ListRow owner, List<WeaponStats> weapons) {
        super(owner, weapons, RangedWeaponStats.class);
    }

    @Override
    protected void createFields(Container parent) {
        Panel panel = new Panel(new PrecisionLayout().setMargins(0).setColumns(5));
        mAccuracy = addField(parent, panel, "99+99*", I18n.text("Accuracy"));
        mRateOfFire = addField(panel, panel, "999*", I18n.text("Rate of Fire"));
        mRange = addField(panel, panel, null, I18n.text("Range"));
        parent.add(panel, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        panel = new Panel(new PrecisionLayout().setMargins(0).setColumns(5));
        mRecoil = addField(parent, panel, "9999", I18n.text("Recoil"));
        mShots = addField(panel, panel, null, I18n.text("Shots"));
        mBulk = addField(panel, panel, "9999", I18n.text("Bulk"));
        parent.add(panel, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
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
    public String toString() {
        return I18n.text("Ranged Weapon");
    }
}
