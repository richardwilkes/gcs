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

/** An editor for melee weapon statistics. */
public class MeleeWeaponEditor extends WeaponEditor {
	@Localize("Melee Weapon")
	@Localize(locale = "de", value = "Nahkampfwaffe")
	@Localize(locale = "ru", value = "Контактное оружие")
	private static String	MELEE_WEAPON;
	@Localize("Reach")
	@Localize(locale = "de", value = "Reichweite")
	@Localize(locale = "ru", value = "Досягаемость")
	private static String	REACH;
	@Localize("Parry Modifier")
	@Localize(locale = "de", value = "Paradewert")
	@Localize(locale = "ru", value = "Модификатор парирования")
	private static String	PARRY;
	@Localize("Block Modifier")
	@Localize(locale = "de", value = "Abblockwert")
	@Localize(locale = "ru", value = "Модификатор блока")
	private static String	BLOCK;

	static {
		Localization.initialize();
	}

	private EditorField		mReach;
	private EditorField		mParry;
	private EditorField		mBlock;

	/**
	 * Creates a new melee weapon editor for the specified row.
	 *
	 * @param row The row to edit melee weapon statistics for.
	 * @return The editor, or <code>null</code> if the row is not appropriate.
	 */
	static public MeleeWeaponEditor createEditor(ListRow row) {
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
	 * @param owner The owning row.
	 * @param weapons The weapons to modify.
	 */
	public MeleeWeaponEditor(ListRow owner, List<WeaponStats> weapons) {
		super(owner, weapons, MeleeWeaponStats.class);
	}

	@Override
	protected void createFields(Container parent) {
		mParry = createTextField(parent, PARRY, EMPTY);
		mReach = createTextField(parent, REACH, EMPTY);
		mBlock = createTextField(parent, BLOCK, EMPTY);
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
		mReach.setValue(EMPTY);
		mParry.setValue(EMPTY);
		mBlock.setValue(EMPTY);
		super.blankFields();
	}

	@Override
	public String toString() {
		return MELEE_WEAPON;
	}
}
