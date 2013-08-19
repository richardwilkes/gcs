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
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.model.skill;

import com.trollworks.gcs.model.CMCharacter;

import java.util.HashSet;

/** The types of possible skill defaults. */
public enum CMSkillDefaultType {
	/** The type for ST-based defaults. */
	ST {

		@Override public int getSkillLevelFast(CMCharacter character, CMSkillDefault skillDefault, HashSet<CMSkill> excludes) {
			return finalLevel(skillDefault, CMSkillAttribute.ST.getBaseSkillLevel(character));
		}
	},
	/** The type for DX-based defaults. */
	DX {
		@Override public int getSkillLevelFast(CMCharacter character, CMSkillDefault skillDefault, HashSet<CMSkill> excludes) {
			return finalLevel(skillDefault, CMSkillAttribute.DX.getBaseSkillLevel(character));
		}
	},
	/** The type for IQ-based defaults. */
	IQ {
		@Override public int getSkillLevelFast(CMCharacter character, CMSkillDefault skillDefault, HashSet<CMSkill> excludes) {
			return finalLevel(skillDefault, CMSkillAttribute.IQ.getBaseSkillLevel(character));
		}
	},
	/** The type for HT-based defaults. */
	HT {
		@Override public int getSkillLevelFast(CMCharacter character, CMSkillDefault skillDefault, HashSet<CMSkill> excludes) {
			return finalLevel(skillDefault, CMSkillAttribute.HT.getBaseSkillLevel(character));
		}
	},
	/** The type for Will-based defaults. */
	Will {
		@Override public int getSkillLevelFast(CMCharacter character, CMSkillDefault skillDefault, HashSet<CMSkill> excludes) {
			return finalLevel(skillDefault, CMSkillAttribute.Will.getBaseSkillLevel(character));
		}
	},
	/** The type for Perception-based defaults. */
	Per {

		@Override public String toString() {
			return Msgs.PERCEPTION;
		}

		@Override public int getSkillLevelFast(CMCharacter character, CMSkillDefault skillDefault, HashSet<CMSkill> excludes) {
			return finalLevel(skillDefault, CMSkillAttribute.Per.getBaseSkillLevel(character));
		}
	},
	/** The type for Skill-based defaults. */
	Skill {

		@Override public String toString() {
			return Msgs.SKILL_NAMED;
		}

		@Override public int getSkillLevelFast(CMCharacter character, CMSkillDefault skillDefault, HashSet<CMSkill> excludes) {
			int best = Integer.MIN_VALUE;

			for (CMSkill skill : character.getSkillNamed(skillDefault.getName(), skillDefault.getSpecialization(), true, excludes)) {
				int level = skill.getLevel();

				if (level > best) {
					best = level;
				}
			}
			return finalLevel(skillDefault, best);
		}

		@Override public int getSkillLevel(CMCharacter character, CMSkillDefault skillDefault, HashSet<CMSkill> excludes) {
			int best = Integer.MIN_VALUE;

			for (CMSkill skill : character.getSkillNamed(skillDefault.getName(), skillDefault.getSpecialization(), true, excludes)) {
				int level = skill.getLevel(excludes);

				if (level > best) {
					best = level;
				}
			}
			return finalLevel(skillDefault, best);
		}
	};

	/**
	 * @param name The name of a {@link CMSkillDefaultType}, as returned from {@link #name()} or
	 *            {@link #toString()}.
	 * @return The matching {@link CMSkillDefaultType}, or {@link #Skill} if a match cannot be
	 *         found.
	 */
	public static final CMSkillDefaultType getByName(String name) {
		for (CMSkillDefaultType type : values()) {
			if (type.name().equalsIgnoreCase(name) || type.toString().equalsIgnoreCase(name)) {
				return type;
			}
		}
		return Skill;
	}

	/**
	 * @param character The character to work with.
	 * @param skillDefault The default being calculated.
	 * @param excludes Exclude these {@link CMSkill}s from consideration.
	 * @return The base skill level for this {@link CMSkillDefaultType}.
	 */
	public abstract int getSkillLevelFast(CMCharacter character, CMSkillDefault skillDefault, HashSet<CMSkill> excludes);

	/**
	 * @param character The character to work with.
	 * @param skillDefault The default being calculated.
	 * @param excludes Exclude these {@link CMSkill}s from consideration.
	 * @return The base skill level for this {@link CMSkillDefaultType}.
	 */
	public int getSkillLevel(CMCharacter character, CMSkillDefault skillDefault, HashSet<CMSkill> excludes) {
		return getSkillLevelFast(character, skillDefault, excludes);
	}

	/**
	 * @param skillDefault The {@link CMSkillDefault}.
	 * @param level The level without the default modifier.
	 * @return The final level.
	 */
	protected int finalLevel(CMSkillDefault skillDefault, int level) {
		if (level != Integer.MIN_VALUE) {
			level += skillDefault.getModifier();
		}
		return level;
	}
}
