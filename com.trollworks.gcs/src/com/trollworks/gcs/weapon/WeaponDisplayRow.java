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

import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.Row;

/** A non-editable row for displaying weapon information. */
public class WeaponDisplayRow extends Row {
    private WeaponStats mWeapon;

    /**
     * Creates a new weapon display row.
     *
     * @param weapon The weapon to display.
     */
    public WeaponDisplayRow(WeaponStats weapon) {
        mWeapon = weapon;
    }

    /** @return The weapon. */
    public WeaponStats getWeapon() {
        return mWeapon;
    }

    @Override
    public Object getData(Column column) {
        return WeaponColumn.values()[column.getID()].getData(mWeapon);
    }

    @Override
    public String getDataAsText(Column column) {
        return WeaponColumn.values()[column.getID()].getDataAsText(mWeapon);
    }

    @Override
    public void setData(Column column, Object data) {
        // Not used
    }

    @Override
    public String getToolTip(Column column) {
        return WeaponColumn.values()[column.getID()].getToolTip(this);
    }

    public String getSkillLevelToolTip() {
        return mWeapon.getSkillLevelToolTip();
    }

    public String getParryToolTip() {
        return mWeapon.getParryToolTip();
    }

    public String getBlockToolTip() {
        return mWeapon.getBlockToolTip();
    }

    public String getDamageToolTip() {
        return mWeapon.getDamage().getDamageToolTip();
    }
}
