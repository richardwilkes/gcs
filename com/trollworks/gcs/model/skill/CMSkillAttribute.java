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

/** The possible skill attributes. */
public enum CMSkillAttribute {
	/** The strength attribute. */
	ST {
		@Override public int getBaseSkillLevel(CMCharacter character) {
			return character != null ? character.getStrength() : Integer.MIN_VALUE;
		}
	},
	/** The dexterity attribute. */
	DX {
		@Override public int getBaseSkillLevel(CMCharacter character) {
			return character != null ? character.getDexterity() : Integer.MIN_VALUE;
		}
	},
	/** The health attribute. */
	HT {
		@Override public int getBaseSkillLevel(CMCharacter character) {
			return character != null ? character.getHealth() : Integer.MIN_VALUE;
		}
	},
	/** The intelligence attribute. */
	IQ {
		@Override public int getBaseSkillLevel(CMCharacter character) {
			return character != null ? character.getIntelligence() : Integer.MIN_VALUE;
		}
	},
	/** The will attribute. */
	Will {
		@Override public int getBaseSkillLevel(CMCharacter character) {
			return character != null ? character.getWill() : Integer.MIN_VALUE;
		}
	},
	/** The perception attribute. */
	Per {
		@Override public int getBaseSkillLevel(CMCharacter character) {
			return character != null ? character.getPerception() : Integer.MIN_VALUE;
		}
	};

	/**
	 * @param character The character to work with.
	 * @return The base skill level for this attribute.
	 */
	public abstract int getBaseSkillLevel(CMCharacter character);
}
