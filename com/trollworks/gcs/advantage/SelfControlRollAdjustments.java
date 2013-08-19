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
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.feature.Bonus;
import com.trollworks.gcs.feature.LeveledAmount;
import com.trollworks.gcs.feature.SkillBonus;
import com.trollworks.ttk.text.Numbers;
import com.trollworks.ttk.utility.LocalizedMessages;

import java.text.MessageFormat;
import java.util.ArrayList;

/** The possible adjustments for self-control rolls. */
public enum SelfControlRollAdjustments {
	/** None. */
	NONE {
		@Override
		public String toString() {
			return MSG_NONE;
		}

		@Override
		public String getDescription(SelfControlRoll cr) {
			return EMPTY;
		}

		@Override
		public int getAdjustment(SelfControlRoll cr) {
			return 0;
		}
	},
	/** General action penalty. */
	ACTION_PENALTY {
		@Override
		public String toString() {
			return MSG_ACTION_PENALTY;
		}

		@Override
		public String getDescription(SelfControlRoll cr) {
			if (cr == SelfControlRoll.NONE_REQUIRED) {
				return EMPTY;
			}
			return MessageFormat.format(MSG_ACTION_PENALTY_DESCRIPTION, Numbers.formatWithForcedSign(getAdjustment(cr)));
		}

		@Override
		public int getAdjustment(SelfControlRoll cr) {
			return cr.ordinal() - 4;
		}
	},
	/** Reaction penalty. */
	REACTION_PENALTY {
		@Override
		public String toString() {
			return MSG_REACTION_PENALTY;
		}

		@Override
		public String getDescription(SelfControlRoll cr) {
			if (cr == SelfControlRoll.NONE_REQUIRED) {
				return EMPTY;
			}
			return MessageFormat.format(MSG_REACTION_PENALTY_DESCRIPTION, Numbers.formatWithForcedSign(getAdjustment(cr)));
		}

		@Override
		public int getAdjustment(SelfControlRoll cr) {
			return cr.ordinal() - 4;
		}
	},
	/** Fright Check penalty. */
	FRIGHT_CHECK_PENALTY {
		@Override
		public String toString() {
			return MSG_FRIGHT_CHECK_PENALTY;
		}

		@Override
		public String getDescription(SelfControlRoll cr) {
			if (cr == SelfControlRoll.NONE_REQUIRED) {
				return EMPTY;
			}
			return MessageFormat.format(MSG_FRIGHT_CHECK_PENALTY_DESCRIPTION, Numbers.formatWithForcedSign(getAdjustment(cr)));
		}

		@Override
		public int getAdjustment(SelfControlRoll cr) {
			return cr.ordinal() - 4;
		}
	},
	/** Fright Check bonus. */
	FRIGHT_CHECK_BONUS {
		@Override
		public String toString() {
			return MSG_FRIGHT_CHECK_BONUS;
		}

		@Override
		public String getDescription(SelfControlRoll cr) {
			if (cr == SelfControlRoll.NONE_REQUIRED) {
				return EMPTY;
			}
			return MessageFormat.format(MSG_FRIGHT_CHECK_BONUS_DESCRIPTION, Numbers.formatWithForcedSign(getAdjustment(cr)));
		}

		@Override
		public int getAdjustment(SelfControlRoll cr) {
			return 4 - cr.ordinal();
		}
	},
	/** Minor cost of living increase. */
	MINOR_COST_OF_LIVING_INCREASE {
		@Override
		public String toString() {
			return MSG_MINOR_COST_OF_LIVING_INCREASE;
		}

		@Override
		public String getDescription(SelfControlRoll cr) {
			if (cr == SelfControlRoll.NONE_REQUIRED) {
				return EMPTY;
			}
			return MessageFormat.format(MSG_MINOR_COST_OF_LIVING_INCREASE_DESCRIPTION, Numbers.formatWithForcedSign(getAdjustment(cr)));
		}

		@Override
		public int getAdjustment(SelfControlRoll cr) {
			return 5 * (4 - cr.ordinal());
		}
	},
	/** Major cost of living increase plus merchant penalty. */
	MAJOR_COST_OF_LIVING_INCREASE {
		@Override
		public String toString() {
			return MSG_MAJOR_COST_OF_LIVING_INCREASE;
		}

		@Override
		public String getDescription(SelfControlRoll cr) {
			if (cr == SelfControlRoll.NONE_REQUIRED) {
				return EMPTY;
			}
			return MessageFormat.format(MSG_MAJOR_COST_OF_LIVING_INCREASE_DESCRIPTION, Numbers.formatWithForcedSign(getAdjustment(cr)));
		}

		@Override
		public int getAdjustment(SelfControlRoll cr) {
			switch (cr) {
				case CR6:
					return 80;
				case CR9:
					return 40;
				case CR12:
					return 20;
				case CR15:
					return 10;
				default:
					return 0;
			}
		}

		@Override
		public ArrayList<Bonus> getBonuses(SelfControlRoll cr) {
			ArrayList<Bonus> list = new ArrayList<>();
			SkillBonus bonus = new SkillBonus();
			StringCriteria criteria = bonus.getNameCriteria();
			criteria.setType(StringCompareType.IS);
			criteria.setQualifier("Merchant"); //$NON-NLS-1$
			criteria = bonus.getSpecializationCriteria();
			criteria.setType(StringCompareType.IS_ANYTHING);
			LeveledAmount amount = bonus.getAmount();
			amount.setIntegerOnly(true);
			amount.setPerLevel(false);
			amount.setAmount(cr.ordinal() - 4);
			list.add(bonus);
			return list;
		}
	};

	static String		MSG_NONE;
	static String		MSG_ACTION_PENALTY;
	static String		MSG_ACTION_PENALTY_DESCRIPTION;
	static String		MSG_REACTION_PENALTY;
	static String		MSG_REACTION_PENALTY_DESCRIPTION;
	static String		MSG_FRIGHT_CHECK_PENALTY;
	static String		MSG_FRIGHT_CHECK_PENALTY_DESCRIPTION;
	static String		MSG_FRIGHT_CHECK_BONUS;
	static String		MSG_FRIGHT_CHECK_BONUS_DESCRIPTION;
	static String		MSG_MINOR_COST_OF_LIVING_INCREASE;
	static String		MSG_MINOR_COST_OF_LIVING_INCREASE_DESCRIPTION;
	static String		MSG_MAJOR_COST_OF_LIVING_INCREASE;
	static String		MSG_MAJOR_COST_OF_LIVING_INCREASE_DESCRIPTION;
	static final String	EMPTY	= "";									//$NON-NLS-1$

	static {
		LocalizedMessages.initialize(SelfControlRollAdjustments.class);
	}

	/**
	 * @param cr The {@link SelfControlRoll} being adjusted.
	 * @return The short description.
	 */
	public abstract String getDescription(SelfControlRoll cr);

	/**
	 * @param cr The {@link SelfControlRoll} being adjusted.
	 * @return The adjustment value.
	 */
	public abstract int getAdjustment(SelfControlRoll cr);

	/**
	 * @param cr The {@link SelfControlRoll} being adjusted.
	 * @return The set of bonuses that this adjustment provides.
	 */
	@SuppressWarnings("static-method")
	public ArrayList<Bonus> getBonuses(SelfControlRoll cr) {
		return new ArrayList<>();
	}
}
