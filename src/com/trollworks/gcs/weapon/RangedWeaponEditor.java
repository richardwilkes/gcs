/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.widget.EditorField;
import com.trollworks.toolkit.utility.Localization;

import java.awt.Container;
import java.util.List;

/** An editor for ranged weapon statistics. */
public class RangedWeaponEditor extends WeaponEditor {
	@Localize("Ranged Weapon")
	@Localize(locale = "de", value = "Fernkampfwaffe")
	@Localize(locale = "ru", value = "Дистанционное оружие")
	private static String	RANGED_WEAPON;
	@Localize("Accuracy")
	@Localize(locale = "de", value = "Genauigkeit")
	@Localize(locale = "ru", value = "Точность")
	private static String	ACCURACY;
	@Localize("Range")
	@Localize(locale = "de", value = "Reichweite")
	@Localize(locale = "ru", value = "Дальность")
	private static String	RANGE;
	@Localize("Rate of Fire")
	@Localize(locale = "de", value = "Schussrate")
	@Localize(locale = "ru", value = "Скорострельность")
	private static String	RATE_OF_FIRE;
	@Localize("Shots")
	@Localize(locale = "de", value = "Schüsse")
	@Localize(locale = "ru", value = "Боезапас")
	private static String	SHOTS;
	@Localize("Bulk")
	@Localize(locale = "de", value = "Handlichkeit")
	@Localize(locale = "ru", value = "Размер")
	private static String	BULK;
	@Localize("Recoil")
	@Localize(locale = "de", value = "Rückstoß")
	@Localize(locale = "ru", value = "Отдача")
	private static String	RECOIL;

	static {
		Localization.initialize();
	}

	private EditorField		mAccuracy;
	private EditorField		mRange;
	private EditorField		mRateOfFire;
	private EditorField		mShots;
	private EditorField		mBulk;
	private EditorField		mRecoil;

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
		mAccuracy = createTextField(parent, ACCURACY, EMPTY);
		mRange = createTextField(parent, RANGE, EMPTY);
		mRateOfFire = createTextField(parent, RATE_OF_FIRE, EMPTY);
		mShots = createTextField(parent, SHOTS, EMPTY);
		mRecoil = createTextField(parent, RECOIL, EMPTY);
		mBulk = createTextField(parent, BULK, EMPTY);
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
		return RANGED_WEAPON;
	}
}
