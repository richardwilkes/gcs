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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;

/** The possible skill attributes. */
public enum SkillAttribute {
	/** The strength attribute. */
	ST {
		@Override
		public String toString() {
			return ST_TITLE;
		}

		@Override
		public int getBaseSkillLevel(GURPSCharacter character) {
			return character != null ? character.getStrength() : Integer.MIN_VALUE;
		}
	},
	/** The dexterity attribute. */
	DX {
		@Override
		public String toString() {
			return DX_TITLE;
		}

		@Override
		public int getBaseSkillLevel(GURPSCharacter character) {
			return character != null ? character.getDexterity() : Integer.MIN_VALUE;
		}
	},
	/** The health attribute. */
	HT {
		@Override
		public String toString() {
			return HT_TITLE;
		}

		@Override
		public int getBaseSkillLevel(GURPSCharacter character) {
			return character != null ? character.getHealth() : Integer.MIN_VALUE;
		}
	},
	/** The intelligence attribute. */
	IQ {
		@Override
		public String toString() {
			return IQ_TITLE;
		}

		@Override
		public int getBaseSkillLevel(GURPSCharacter character) {
			return character != null ? character.getIntelligence() : Integer.MIN_VALUE;
		}
	},
	/** The will attribute. */
	Will {
		@Override
		public String toString() {
			return WILL_TITLE;
		}

		@Override
		public int getBaseSkillLevel(GURPSCharacter character) {
			return character != null ? character.getWill() : Integer.MIN_VALUE;
		}
	},
	/** The perception attribute. */
	Per {
		@Override
		public String toString() {
			return PER_TITLE;
		}

		@Override
		public int getBaseSkillLevel(GURPSCharacter character) {
			return character != null ? character.getPerception() : Integer.MIN_VALUE;
		}
	};

	@Localize("ST")
	@Localize(locale = "de", value = "ST")
	@Localize(locale = "ru", value = "СЛ")
	static String	ST_TITLE;
	@Localize("DX")
	@Localize(locale = "de", value = "GE")
	@Localize(locale = "ru", value = "ЛВ")
	static String	DX_TITLE;
	@Localize("IQ")
	@Localize(locale = "de", value = "IQ")
	@Localize(locale = "ru", value = "ИН")
	static String	IQ_TITLE;
	@Localize("HT")
	@Localize(locale = "de", value = "KO")
	@Localize(locale = "ru", value = "ЗД")
	static String	HT_TITLE;
	@Localize("Will")
	@Localize(locale = "de", value = "Wille")
	@Localize(locale = "ru", value = "Воля")
	static String	WILL_TITLE;
	@Localize("Per")
	@Localize(locale = "de", value = "WN")
	@Localize(locale = "ru", value = "Восп")
	static String	PER_TITLE;

	static {
		Localization.initialize();
	}

	/**
	 * @param character The character to work with.
	 * @return The base skill level for this attribute.
	 */
	public abstract int getBaseSkillLevel(GURPSCharacter character);
}
