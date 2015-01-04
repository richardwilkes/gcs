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

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.feature.Bonus;
import com.trollworks.gcs.feature.LeveledAmount;
import com.trollworks.gcs.feature.SkillBonus;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.Numbers;

import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/** The possible adjustments for self-control rolls. */
public enum SelfControlRollAdjustments {
	/** None. */
	NONE {
		@Override
		public String toString() {
			return NONE_TITLE;
		}

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
		public String toString() {
			return ACTION_PENALTY_TITLE;
		}

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
		public String toString() {
			return REACTION_PENALTY_TITLE;
		}

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
		public String toString() {
			return FRIGHT_CHECK_PENALTY_TITLE;
		}

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
		public String toString() {
			return FRIGHT_CHECK_BONUS_TITLE;
		}

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
		public String toString() {
			return MINOR_COST_OF_LIVING_INCREASE_TITLE;
		}

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
		public String toString() {
			return MAJOR_COST_OF_LIVING_INCREASE_TITLE;
		}

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

	@Localize("None")
	@Localize(locale = "de", value = "Keine Besonderheit")
	@Localize(locale = "ru", value = "Нет")
	static String	NONE_TITLE;
	@Localize("Includes an Action Penalty for Failure")
	@Localize(locale = "de", value = "Beinhaltet einen Aktions-Malus wenn nicht geschafft")
	@Localize(locale = "ru", value = "Включает штрафы действий при провале")
	static String	ACTION_PENALTY_TITLE;
	@Localize("{0} Action Penalty")
	@Localize(locale = "de", value = "{0} Aktions-Malus")
	@Localize(locale = "ru", value = "{0} штраф действия")
	static String	ACTION_PENALTY_DESCRIPTION;
	@Localize("Includes a Reaction Penalty for Failure")
	@Localize(locale = "de", value = "Beinhaltet einen Reaktions-Malus wenn nicht geschafft")
	@Localize(locale = "ru", value = "Включает штрафы реакции при провале")
	static String	REACTION_PENALTY_TITLE;
	@Localize("{0} Reaction Penalty")
	@Localize(locale = "de", value = "{0} Reaktions-Malus")
	@Localize(locale = "ru", value = "{0} штраф реакции")
	static String	REACTION_PENALTY_DESCRIPTION;
	@Localize("Includes Fright Check Penalty")
	@Localize(locale = "de", value = "Beinhaltet Schreckproben-Malus")
	@Localize(locale = "ru", value = "Включает штраф броска страха")
	static String	FRIGHT_CHECK_PENALTY_TITLE;
	@Localize("{0} Fright Check Penalty")
	@Localize(locale = "de", value = "{0} Schreckproben-Malus")
	@Localize(locale = "ru", value = "{0} штраф проверки страха")
	static String	FRIGHT_CHECK_PENALTY_DESCRIPTION;
	@Localize("Includes Fright Check Bonus")
	@Localize(locale = "de", value = "Beinhaltet Schreckproben-Bonus")
	@Localize(locale = "ru", value = "Включает премию броска страха")
	static String	FRIGHT_CHECK_BONUS_TITLE;
	@Localize("{0} Fright Check Bonus")
	@Localize(locale = "de", value = "{0} Schreckproben-Bonus")
	@Localize(locale = "ru", value = "{0} премия проверки страха")
	static String	FRIGHT_CHECK_BONUS_DESCRIPTION;
	@Localize("Includes a Minor Cost of Living Increase")
	@Localize(locale = "de", value = "Beinhaltet eine geringe Erhöhung der Lebenshaltungskosten")
	@Localize(locale = "ru", value = "Включает малое возрастание расходов на жизнь")
	static String	MINOR_COST_OF_LIVING_INCREASE_TITLE;
	@Localize("{0}% Cost of Living Increase")
	@Localize(locale = "de", value = "{0}% höhere Lebenshaltungskosten")
	@Localize(locale = "ru", value = "{0}% возрастания расходов на жизнь")
	static String	MINOR_COST_OF_LIVING_INCREASE_DESCRIPTION;
	@Localize("Includes a Major Cost of Living Increase and Merchant Skill Penalty")
	@Localize(locale = "de", value = "Beinhaltet eine deutliche Erhöhung der Lebenshaltungskosten und einen Malus der Händler-Fertigkeit")
	@Localize(locale = "ru", value = "Включает большое возрастание расходов на жизнь и штраф на торговое дело")
	static String	MAJOR_COST_OF_LIVING_INCREASE_TITLE;
	@Localize("{0}% Cost of Living Increase")
	@Localize(locale = "de", value = "{0}% höhere Lebenshaltungskosten")
	@Localize(locale = "ru", value = "{0}% возрастания расходов на жизнь")
	static String	MAJOR_COST_OF_LIVING_INCREASE_DESCRIPTION;

	static {
		Localization.initialize();
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
