/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.EditorField;

import java.awt.Container;
import java.util.List;

/** An editor for ranged weapon statistics. */
public class RangedWeaponEditor extends WeaponEditor {
	private static String	MSG_RANGED_WEAPON;
	private static String	MSG_ACCURACY;
	private static String	MSG_RANGE;
	private static String	MSG_RATE_OF_FIRE;
	private static String	MSG_SHOTS;
	private static String	MSG_BULK;
	private static String	MSG_RECOIL;
	private EditorField		mAccuracy;
	private EditorField		mRange;
	private EditorField		mRateOfFire;
	private EditorField		mShots;
	private EditorField		mBulk;
	private EditorField		mRecoil;

	static {
		LocalizedMessages.initialize(RangedWeaponEditor.class);
	}

	/**
	 * Creates a new ranged weapon editor for the specified row.
	 * 
	 * @param row The row to edit ranged weapon statistics for.
	 * @return The editor, or <code>null</code> if the row is not appropriate.
	 */
	static public RangedWeaponEditor createEditor(ListRow row) {
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
	 * @param owner The owning row.
	 * @param weapons The weapons to modify.
	 */
	public RangedWeaponEditor(ListRow owner, List<WeaponStats> weapons) {
		super(owner, weapons, RangedWeaponStats.class);
	}

	@Override
	protected void createFields(Container parent) {
		mAccuracy = createTextField(parent, MSG_ACCURACY, EMPTY);
		mRange = createTextField(parent, MSG_RANGE, EMPTY);
		mRateOfFire = createTextField(parent, MSG_RATE_OF_FIRE, EMPTY);
		mShots = createTextField(parent, MSG_SHOTS, EMPTY);
		mRecoil = createTextField(parent, MSG_RECOIL, EMPTY);
		mBulk = createTextField(parent, MSG_BULK, EMPTY);
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
		mAccuracy.setValue(EMPTY);
		mRange.setValue(EMPTY);
		mRateOfFire.setValue(EMPTY);
		mShots.setValue(EMPTY);
		mBulk.setValue(EMPTY);
		mRecoil.setValue(EMPTY);
		super.blankFields();
	}

	@Override
	public String toString() {
		return MSG_RANGED_WEAPON;
	}
}
