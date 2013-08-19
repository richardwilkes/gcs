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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.feature.LeveledAmount;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillDefaultType;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.text.NumberUtils;
import com.trollworks.ttk.utility.Dice;
import com.trollworks.ttk.xml.XMLNodeType;
import com.trollworks.ttk.xml.XMLReader;
import com.trollworks.ttk.xml.XMLWriter;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashSet;
import java.util.List;

/** The stats for a weapon. */
public abstract class WeaponStats {
	private static final String		TAG_DAMAGE		= "damage";									//$NON-NLS-1$
	private static final String		TAG_STRENGTH	= "strength";									//$NON-NLS-1$
	private static final String		TAG_USAGE		= "usage";										//$NON-NLS-1$
	/** The prefix used in front of all IDs for weapons. */
	public static final String		PREFIX			= GURPSCharacter.CHARACTER_PREFIX + "weapon.";	//$NON-NLS-1$
	/** The field ID for damage changes. */
	public static final String		ID_DAMAGE		= PREFIX + TAG_DAMAGE;
	/** The field ID for strength changes. */
	public static final String		ID_STRENGTH		= PREFIX + TAG_STRENGTH;
	/** The field ID for usage changes. */
	public static final String		ID_USAGE		= PREFIX + TAG_USAGE;
	/** An empty string. */
	protected static final String	EMPTY			= "";											//$NON-NLS-1$
	private ListRow					mOwner;
	private String					mDamage;
	private String					mStrength;
	private String					mUsage;
	private ArrayList<SkillDefault>	mDefaults;

	/**
	 * Creates a new weapon.
	 * 
	 * @param owner The owning piece of equipment or advantage.
	 */
	protected WeaponStats(ListRow owner) {
		mOwner = owner;
		mDamage = EMPTY;
		mStrength = EMPTY;
		mUsage = EMPTY;
		mDefaults = new ArrayList<SkillDefault>();
		initialize();
	}

	/**
	 * Creates a clone of the specified weapon.
	 * 
	 * @param owner The owning piece of equipment or advantage.
	 * @param other The weapon to clone.
	 */
	protected WeaponStats(ListRow owner, WeaponStats other) {
		mOwner = owner;
		mDamage = other.mDamage;
		mStrength = other.mStrength;
		mUsage = other.mUsage;
		mDefaults = new ArrayList<SkillDefault>();
		for (SkillDefault skillDefault : other.mDefaults) {
			mDefaults.add(new SkillDefault(skillDefault));
		}
	}

	/**
	 * Creates a weapon.
	 * 
	 * @param owner The owning piece of equipment or advantage.
	 * @param reader The reader to load from.
	 * @throws IOException
	 */
	public WeaponStats(ListRow owner, XMLReader reader) throws IOException {
		this(owner);

		String marker = reader.getMarker();

		do {
			if (reader.next() == XMLNodeType.START_TAG) {
				loadSelf(reader);
			}
		} while (reader.withinMarker(marker));
	}

	/**
	 * Creates a clone of this weapon.
	 * 
	 * @param owner The owning piece of equipment or advantage.
	 * @return The cloned weapon.
	 */
	public abstract WeaponStats clone(ListRow owner);

	/** Called so that sub-classes can initialize themselves. */
	protected abstract void initialize();

	/**
	 * @param reader The reader to load from.
	 * @throws IOException
	 */
	protected void loadSelf(XMLReader reader) throws IOException {
		String name = reader.getName();

		if (TAG_DAMAGE.equals(name)) {
			mDamage = reader.readText();
		} else if (TAG_STRENGTH.equals(name)) {
			mStrength = reader.readText();
		} else if (TAG_USAGE.equals(name)) {
			mUsage = reader.readText();
		} else if (SkillDefault.TAG_ROOT.equals(name)) {
			mDefaults.add(new SkillDefault(reader));
		} else {
			reader.skipTag(name);
		}
	}

	/** @return The root XML tag to use when saving. */
	protected abstract String getRootTag();

	/**
	 * Saves the weapon.
	 * 
	 * @param out The XML writer to use.
	 */
	public void save(XMLWriter out) {
		out.startSimpleTagEOL(getRootTag());
		out.simpleTagNotEmpty(TAG_DAMAGE, mDamage);
		out.simpleTagNotEmpty(TAG_STRENGTH, mStrength);
		out.simpleTagNotEmpty(TAG_USAGE, mUsage);
		saveSelf(out);
		for (SkillDefault skillDefault : mDefaults) {
			skillDefault.save(out);
		}
		out.endTagEOL(getRootTag(), true);
	}

	/**
	 * Called so that sub-classes can save their own data.
	 * 
	 * @param out The XML writer to use.
	 */
	protected abstract void saveSelf(XMLWriter out);

	/** @return The defaults for this weapon. */
	public List<SkillDefault> getDefaults() {
		return Collections.unmodifiableList(mDefaults);
	}

	/**
	 * @param defaults The new defaults for this weapon.
	 * @return Whether there was a change or not.
	 */
	public boolean setDefaults(List<SkillDefault> defaults) {
		if (!mDefaults.equals(defaults)) {
			mDefaults = new ArrayList<SkillDefault>(defaults);
			return true;
		}
		return false;
	}

	/** @param id The ID to use for notification. */
	protected void notifySingle(String id) {
		if (mOwner != null) {
			mOwner.notifySingle(id);
		}
	}

	/** @return A description of the weapon. */
	public String getDescription() {
		if (mOwner instanceof Equipment) {
			return ((Equipment) mOwner).getDescription();
		}
		if (mOwner instanceof Advantage) {
			return ((Advantage) mOwner).getName();
		}
		if (mOwner instanceof Spell) {
			return ((Spell) mOwner).getName();
		}
		if (mOwner instanceof Skill) {
			return ((Skill) mOwner).getName();
		}
		return EMPTY;
	}

	@Override
	public String toString() {
		return getDescription();
	}

	/** @return The notes for this weapon. */
	public String getNotes() {
		return mOwner != null ? mOwner.getNotes() : EMPTY;
	}

	/** @return The damage. */
	public String getDamage() {
		return mDamage;
	}

	/** @return The damage, fully resolved for the user's sw or thr, if possible. */
	public String getResolvedDamage() {
		DataFile df = mOwner.getDataFile();
		String damage = mDamage;

		if (df instanceof GURPSCharacter) {
			GURPSCharacter character = (GURPSCharacter) df;
			HashSet<LeveledAmount> bonuses = new HashSet<LeveledAmount>();

			for (SkillDefault one : getDefaults()) {
				if (one.getType().isSkillBased()) {
					bonuses.addAll(character.getWeaponComparedBonusesFor(Skill.ID_NAME + "*", one.getName(), one.getSpecialization())); //$NON-NLS-1$
					bonuses.addAll(character.getWeaponComparedBonusesFor(Skill.ID_NAME + "/" + one.getName(), one.getName(), one.getSpecialization())); //$NON-NLS-1$
				}
			}
			damage = resolveDamage(damage, bonuses);
		}
		return damage.trim();
	}

	private String resolveDamage(String damage, HashSet<LeveledAmount> bonuses) {
		int maxST = getMinStrengthValue() * 3;
		GURPSCharacter character = (GURPSCharacter) mOwner.getDataFile();
		int st = character.getStrength();
		Dice dice;
		String savedDamage;

		if (maxST > 0 && maxST < st) {
			st = maxST;
		}

		dice = GURPSCharacter.getSwing(st + character.getStrikingStrengthBonus());
		do {
			savedDamage = damage;
			damage = resolveDamage(damage, "sw", dice, bonuses); //$NON-NLS-1$
		} while (!savedDamage.equals(damage));

		dice = GURPSCharacter.getThrust(st + character.getStrikingStrengthBonus());
		do {
			savedDamage = damage;
			damage = resolveDamage(damage, "thr", dice, bonuses); //$NON-NLS-1$
		} while (!savedDamage.equals(damage));
		return damage;
	}

	private String resolveDamage(String damage, String type, Dice dice, HashSet<LeveledAmount> bonuses) {
		int where = damage.indexOf(type);

		if (where != -1) {
			int last = where + type.length();
			int max = damage.length();
			StringBuffer buffer = new StringBuffer();
			int tmp;

			if (where > 0) {
				buffer.append(damage.substring(0, where));
			}

			tmp = skipSpaces(damage, last);
			if (tmp < max) {
				char ch = damage.charAt(tmp);

				if (ch == '+' || ch == '-') {
					int modifier = 0;

					tmp = skipSpaces(damage, tmp + 1);
					while (tmp < max) {
						char digit = damage.charAt(tmp);

						if (digit >= '0' && digit <= '9') {
							modifier *= 10;
							modifier += digit - '0';
							tmp++;
						} else {
							break;
						}
					}
					if (ch == '-') {
						modifier = -modifier;
					}
					last = tmp;
					dice = (Dice) dice.clone();
					dice.add(modifier);
				}
				if (last < max - 1 && damage.charAt(last) == ':') {
					tmp = last + 1;
					ch = damage.charAt(tmp++);
					if (ch == '+' || ch == '-') {
						int perDie = 0;

						while (tmp < max) {
							char digit = damage.charAt(tmp);

							if (digit >= '0' && digit <= '9') {
								perDie *= 10;
								perDie += digit - '0';
								tmp++;
							} else {
								break;
							}
						}
						last = tmp;
						if (perDie > 0) {
							if (ch == '-') {
								perDie = -perDie;
							}
							dice = (Dice) dice.clone();
							dice.add(perDie * dice.getDieCount());
						}
					}
				}
			}

			for (LeveledAmount bonus : bonuses) {
				int amt = bonus.getIntegerAmount();

				if (bonus.isPerLevel()) {
					dice.add(amt * dice.getDieCount());
				} else {
					dice.add(amt);
				}
			}

			buffer.append(dice.toString());
			if (last < max) {
				buffer.append(damage.substring(last));
			}
			return buffer.toString();
		}
		return damage;
	}

	/**
	 * @param buffer The string to find the next non-space character within.
	 * @param index The index to start looking.
	 * @return The index of the next non-space character.
	 */
	protected int skipSpaces(String buffer, int index) {
		int max = buffer.length();
		while (index < max && buffer.charAt(index) == ' ') {
			index++;
		}
		return index;
	}

	/**
	 * Sets the value of damage.
	 * 
	 * @param damage The value to set.
	 */
	public void setDamage(String damage) {
		damage = sanitize(damage);
		if (!mDamage.equals(damage)) {
			mDamage = damage;
			notifySingle(ID_DAMAGE);
		}
	}

	/** @return The skill level. */
	public int getSkillLevel() {
		DataFile df = mOwner.getDataFile();

		if (df instanceof GURPSCharacter) {
			return getSkillLevel((GURPSCharacter) df);
		}
		return 0;
	}

	private int getSkillLevel(GURPSCharacter character) {
		int best = Integer.MIN_VALUE;

		for (SkillDefault skillDefault : getDefaults()) {
			SkillDefaultType type = skillDefault.getType();
			int level = type.getSkillLevelFast(character, skillDefault, new HashSet<String>());

			if (level > best) {
				best = level;
			}
		}

		if (best != Integer.MIN_VALUE) {
			int minST = getMinStrengthValue() - character.getStrength();

			if (minST > 0) {
				best -= minST;
				if (best < 0) {
					best = 0;
				}
			}
		} else {
			best = 0;
		}
		return best;
	}

	/** @return The minimum ST to use this weapon, or -1 if there is none. */
	public int getMinStrengthValue() {
		StringBuilder builder = new StringBuilder();
		int count = mStrength.length();
		boolean started = false;

		for (int i = 0; i < count; i++) {
			char ch = mStrength.charAt(i);

			if (Character.isDigit(ch)) {
				builder.append(ch);
				started = true;
			} else if (started) {
				break;
			}
		}

		return started ? NumberUtils.getInteger(builder.toString(), -1) : -1;
	}

	/** @return The usage. */
	public String getUsage() {
		return mUsage;
	}

	/** @param usage The value to set. */
	public void setUsage(String usage) {
		usage = sanitize(usage);
		if (!mUsage.equals(usage)) {
			mUsage = usage;
			notifySingle(ID_USAGE);
		}
	}

	/** @return The strength. */
	public String getStrength() {
		return mStrength;
	}

	/**
	 * Sets the value of strength.
	 * 
	 * @param strength The value to set.
	 */
	public void setStrength(String strength) {
		strength = sanitize(strength);
		if (!mStrength.equals(strength)) {
			mStrength = strength;
			notifySingle(ID_STRENGTH);
		}
	}

	/** @return The owner. */
	public ListRow getOwner() {
		return mOwner;
	}

	/**
	 * Sets the value of owner.
	 * 
	 * @param owner The value to set.
	 */
	public void setOwner(ListRow owner) {
		mOwner = owner;
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof WeaponStats) {
			WeaponStats other = (WeaponStats) obj;

			return mDamage.equals(other.mDamage) && mStrength.equals(other.mStrength) && mUsage.equals(other.mUsage) && mDefaults.equals(other.mDefaults);
		}
		return false;
	}

	/**
	 * @param data The data to sanitize.
	 * @return The original data, or "", if the data was <code>null</code>.
	 */
	protected String sanitize(String data) {
		if (data == null) {
			return EMPTY;
		}
		return data;
	}
}
