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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.utility.io.LocalizedMessages;

import java.util.HashSet;

/** The types of possible skill defaults. */
public enum SkillDefaultType {
	/** The type for ST-based defaults. */
	ST {

		@Override public int getSkillLevelFast(GURPSCharacter character, SkillDefault skillDefault, HashSet<String> excludes) {
			return finalLevel(skillDefault, SkillAttribute.ST.getBaseSkillLevel(character));
		}
	},
	/** The type for DX-based defaults. */
	DX {
		@Override public int getSkillLevelFast(GURPSCharacter character, SkillDefault skillDefault, HashSet<String> excludes) {
			return finalLevel(skillDefault, SkillAttribute.DX.getBaseSkillLevel(character));
		}
	},
	/** The type for IQ-based defaults. */
	IQ {
		@Override public int getSkillLevelFast(GURPSCharacter character, SkillDefault skillDefault, HashSet<String> excludes) {
			return finalLevel(skillDefault, SkillAttribute.IQ.getBaseSkillLevel(character));
		}
	},
	/** The type for HT-based defaults. */
	HT {
		@Override public int getSkillLevelFast(GURPSCharacter character, SkillDefault skillDefault, HashSet<String> excludes) {
			return finalLevel(skillDefault, SkillAttribute.HT.getBaseSkillLevel(character));
		}
	},
	/** The type for Will-based defaults. */
	Will {
		@Override public int getSkillLevelFast(GURPSCharacter character, SkillDefault skillDefault, HashSet<String> excludes) {
			return finalLevel(skillDefault, SkillAttribute.Will.getBaseSkillLevel(character));
		}
	},
	/** The type for Perception-based defaults. */
	Per {
		@Override public String toString() {
			return MSG_PERCEPTION;
		}

		@Override public int getSkillLevelFast(GURPSCharacter character, SkillDefault skillDefault, HashSet<String> excludes) {
			return finalLevel(skillDefault, SkillAttribute.Per.getBaseSkillLevel(character));
		}
	},
	/** The type for Skill-based defaults. */
	Skill {
		@Override public String toString() {
			return MSG_SKILL_NAMED;
		}

		@Override public int getSkillLevelFast(GURPSCharacter character, SkillDefault skillDefault, HashSet<String> excludes) {
			int best = Integer.MIN_VALUE;
			for (Skill skill : character.getSkillNamed(skillDefault.getName(), skillDefault.getSpecialization(), true, excludes)) {
				int level = skill.getLevel();
				if (level > best) {
					best = level;
				}
			}
			return finalLevel(skillDefault, best);
		}

		@Override public int getSkillLevel(GURPSCharacter character, SkillDefault skillDefault, HashSet<String> excludes) {
			int best = Integer.MIN_VALUE;
			for (Skill skill : character.getSkillNamed(skillDefault.getName(), skillDefault.getSpecialization(), true, excludes)) {
				int level = skill.getLevel(excludes);
				if (level > best) {
					best = level;
				}
			}
			return finalLevel(skillDefault, best);
		}

		@Override public boolean isSkillBased() {
			return true;
		}
	},
	/** The type for Parry-based defaults. */
	Parry {
		@Override public String toString() {
			return MSG_PARRY_SKILL_NAMED;
		}

		@Override public int getSkillLevelFast(GURPSCharacter character, SkillDefault skillDefault, HashSet<String> excludes) {
			int best = Integer.MIN_VALUE;
			for (Skill skill : character.getSkillNamed(skillDefault.getName(), skillDefault.getSpecialization(), true, excludes)) {
				int level = skill.getLevel();
				if (level > best) {
					best = level;
				}
			}
			return finalLevel(skillDefault, best / 2 + 3 + character.getParryBonus());
		}

		@Override public int getSkillLevel(GURPSCharacter character, SkillDefault skillDefault, HashSet<String> excludes) {
			int best = Integer.MIN_VALUE;
			for (Skill skill : character.getSkillNamed(skillDefault.getName(), skillDefault.getSpecialization(), true, excludes)) {
				int level = skill.getLevel(excludes);
				if (level > best) {
					best = level;
				}
			}
			return finalLevel(skillDefault, best / 2 + 3 + character.getParryBonus());
		}

		@Override public boolean isSkillBased() {
			return true;
		}
	},
	/** The type for Block-based defaults. */
	Block {
		@Override public String toString() {
			return MSG_BLOCK_SKILL_NAMED;
		}

		@Override public int getSkillLevelFast(GURPSCharacter character, SkillDefault skillDefault, HashSet<String> excludes) {
			int best = Integer.MIN_VALUE;
			for (Skill skill : character.getSkillNamed(skillDefault.getName(), skillDefault.getSpecialization(), true, excludes)) {
				int level = skill.getLevel();
				if (level > best) {
					best = level;
				}
			}
			return finalLevel(skillDefault, best / 2 + 3 + character.getBlockBonus());
		}

		@Override public int getSkillLevel(GURPSCharacter character, SkillDefault skillDefault, HashSet<String> excludes) {
			int best = Integer.MIN_VALUE;
			for (Skill skill : character.getSkillNamed(skillDefault.getName(), skillDefault.getSpecialization(), true, excludes)) {
				int level = skill.getLevel(excludes);
				if (level > best) {
					best = level;
				}
			}
			return finalLevel(skillDefault, best / 2 + 3 + character.getBlockBonus());
		}

		@Override public boolean isSkillBased() {
			return true;
		}
	};

	static String	MSG_PERCEPTION;
	static String	MSG_SKILL_NAMED;
	static String	MSG_PARRY_SKILL_NAMED;
	static String	MSG_BLOCK_SKILL_NAMED;

	static {
		LocalizedMessages.initialize(SkillDefaultType.class);
	}

	/**
	 * @param name The name of a {@link SkillDefaultType}, as returned from {@link #name()} or
	 *            {@link #toString()}.
	 * @return The matching {@link SkillDefaultType}, or {@link #Skill} if a match cannot be found.
	 */
	public static final SkillDefaultType getByName(String name) {
		for (SkillDefaultType type : values()) {
			if (type.name().equalsIgnoreCase(name) || type.toString().equalsIgnoreCase(name)) {
				return type;
			}
		}
		return Skill;
	}

	/** @return Whether the {@link SkillDefaultType} is based on another skill or not. */
	public boolean isSkillBased() {
		return false;
	}

	/**
	 * @param character The character to work with.
	 * @param skillDefault The default being calculated.
	 * @param excludes Exclude these {@link Skill}s from consideration.
	 * @return The base skill level for this {@link SkillDefaultType}.
	 */
	public abstract int getSkillLevelFast(GURPSCharacter character, SkillDefault skillDefault, HashSet<String> excludes);

	/**
	 * @param character The character to work with.
	 * @param skillDefault The default being calculated.
	 * @param excludes Exclude these {@link Skill}s from consideration.
	 * @return The base skill level for this {@link SkillDefaultType}.
	 */
	public int getSkillLevel(GURPSCharacter character, SkillDefault skillDefault, HashSet<String> excludes) {
		return getSkillLevelFast(character, skillDefault, excludes);
	}

	/**
	 * @param skillDefault The {@link SkillDefault}.
	 * @param level The level without the default modifier.
	 * @return The final level.
	 */
	protected int finalLevel(SkillDefault skillDefault, int level) {
		if (level != Integer.MIN_VALUE) {
			level += skillDefault.getModifier();
		}
		return level;
	}
}
