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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.weapon;

import static com.trollworks.gcs.weapon.MeleeWeaponEditor_LS.*;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.widgets.EditorField;

import java.awt.Container;
import java.util.List;

@Localized({
				@LS(key = "MELEE_WEAPON", msg = "Melee Weapon"),
				@LS(key = "REACH", msg = "Reach"),
				@LS(key = "PARRY", msg = "Parry Modifier"),
				@LS(key = "BLOCK", msg = "Block Modifier"),
})
/** An editor for melee weapon statistics. */
public class MeleeWeaponEditor extends WeaponEditor {
	private EditorField	mReach;
	private EditorField	mParry;
	private EditorField	mBlock;

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
