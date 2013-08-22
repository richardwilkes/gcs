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

package com.trollworks.gcs.advantage;

import static com.trollworks.gcs.advantage.SelfControlRollAdjustments_LS.*;

import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.feature.Bonus;
import com.trollworks.gcs.feature.LeveledAmount;
import com.trollworks.gcs.feature.SkillBonus;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.text.Numbers;

import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

@Localized({
				@LS(key = "NONE", msg = "None"),
				@LS(key = "ACTION_PENALTY", msg = "Includes an Action Penalty for Failure"),
				@LS(key = "ACTION_PENALTY_DESCRIPTION", msg = "{0} Action Penalty"),
				@LS(key = "REACTION_PENALTY", msg = "Includes a Reaction Penalty for Failure"),
				@LS(key = "REACTION_PENALTY_DESCRIPTION", msg = "{0} Reaction Penalty"),
				@LS(key = "FRIGHT_CHECK_PENALTY", msg = "Includes Fright Check Penalty"),
				@LS(key = "FRIGHT_CHECK_PENALTY_DESCRIPTION", msg = "{0} Fright Check Penalty"),
				@LS(key = "FRIGHT_CHECK_BONUS", msg = "Includes Fright Check Bonus"),
				@LS(key = "FRIGHT_CHECK_BONUS_DESCRIPTION", msg = "{0} Fright Check Bonus"),
				@LS(key = "MINOR_COST_OF_LIVING_INCREASE", msg = "Includes a Minor Cost of Living Increase"),
				@LS(key = "MINOR_COST_OF_LIVING_INCREASE_DESCRIPTION", msg = "{0}% Cost of Living Increase"),
				@LS(key = "MAJOR_COST_OF_LIVING_INCREASE", msg = "Includes a Major Cost of Living Increase and Merchant Skill Penalty"),
				@LS(key = "MAJOR_COST_OF_LIVING_INCREASE_DESCRIPTION", msg = "{0}% Cost of Living Increase"),
})
/** The possible adjustments for self-control rolls. */
public enum SelfControlRollAdjustments {
	/** None. */
	NONE {
		@Override
		public String getDescription(SelfControlRoll cr) {
			return ""; //$NON-NLS-1$
		}

		@Override
		public int getAdjustment(SelfControlRoll cr) {
			return 0;
		}
	},
	/** General action penalty. */
	ACTION_PENALTY {
		@Override
		public String getDescription(SelfControlRoll cr) {
			if (cr == SelfControlRoll.NONE_REQUIRED) {
				return ""; //$NON-NLS-1$
			}
			return MessageFormat.format(ACTION_PENALTY_DESCRIPTION, Numbers.formatWithForcedSign(getAdjustment(cr)));
		}

		@Override
		public int getAdjustment(SelfControlRoll cr) {
			return cr.ordinal() - 4;
		}
	},
	/** Reaction penalty. */
	REACTION_PENALTY {
		@Override
		public String getDescription(SelfControlRoll cr) {
			if (cr == SelfControlRoll.NONE_REQUIRED) {
				return ""; //$NON-NLS-1$
			}
			return MessageFormat.format(REACTION_PENALTY_DESCRIPTION, Numbers.formatWithForcedSign(getAdjustment(cr)));
		}

		@Override
		public int getAdjustment(SelfControlRoll cr) {
			return cr.ordinal() - 4;
		}
	},
	/** Fright Check penalty. */
	FRIGHT_CHECK_PENALTY {
		@Override
		public String getDescription(SelfControlRoll cr) {
			if (cr == SelfControlRoll.NONE_REQUIRED) {
				return ""; //$NON-NLS-1$
			}
			return MessageFormat.format(FRIGHT_CHECK_PENALTY_DESCRIPTION, Numbers.formatWithForcedSign(getAdjustment(cr)));
		}

		@Override
		public int getAdjustment(SelfControlRoll cr) {
			return cr.ordinal() - 4;
		}
	},
	/** Fright Check bonus. */
	FRIGHT_CHECK_BONUS {
		@Override
		public String getDescription(SelfControlRoll cr) {
			if (cr == SelfControlRoll.NONE_REQUIRED) {
				return ""; //$NON-NLS-1$
			}
			return MessageFormat.format(FRIGHT_CHECK_BONUS_DESCRIPTION, Numbers.formatWithForcedSign(getAdjustment(cr)));
		}

		@Override
		public int getAdjustment(SelfControlRoll cr) {
			return 4 - cr.ordinal();
		}
	},
	/** Minor cost of living increase. */
	MINOR_COST_OF_LIVING_INCREASE {
		@Override
		public String getDescription(SelfControlRoll cr) {
			if (cr == SelfControlRoll.NONE_REQUIRED) {
				return ""; //$NON-NLS-1$
			}
			return MessageFormat.format(MINOR_COST_OF_LIVING_INCREASE_DESCRIPTION, Numbers.formatWithForcedSign(getAdjustment(cr)));
		}

		@Override
		public int getAdjustment(SelfControlRoll cr) {
			return 5 * (4 - cr.ordinal());
		}
	},
	/** Major cost of living increase plus merchant penalty. */
	MAJOR_COST_OF_LIVING_INCREASE {
		@Override
		public String getDescription(SelfControlRoll cr) {
			if (cr == SelfControlRoll.NONE_REQUIRED) {
				return ""; //$NON-NLS-1$
			}
			return MessageFormat.format(MAJOR_COST_OF_LIVING_INCREASE_DESCRIPTION, Numbers.formatWithForcedSign(getAdjustment(cr)));
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
		public List<Bonus> getBonuses(SelfControlRoll cr) {
			List<Bonus> list = new ArrayList<>();
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

	@Override
	public String toString() {
		return SelfControlRollAdjustments_LS.toString(this);
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
	public List<Bonus> getBonuses(SelfControlRoll cr) {
		return Collections.emptyList();
	}
}
