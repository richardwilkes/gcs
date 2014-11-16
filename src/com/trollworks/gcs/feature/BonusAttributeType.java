/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.feature;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;

/** The attribute affected by a {@link AttributeBonus}. */
public enum BonusAttributeType {
	/** The ST attribute. */
	ST {
		@Override
		public String toString() {
			return ST_TITLE;
		}
	},
	/** The DX attribute. */
	DX {
		@Override
		public String toString() {
			return DX_TITLE;
		}
	},
	/** The IQ attribute. */
	IQ {
		@Override
		public String toString() {
			return IQ_TITLE;
		}
	},
	/** The HT attribute. */
	HT {
		@Override
		public String toString() {
			return HT_TITLE;
		}
	},
	/** The Will attribute. */
	WILL {
		@Override
		public String toString() {
			return WILL_TITLE;
		}
	},
	/** The Fright Check attribute. */
	FRIGHT_CHECK {
		@Override
		public String toString() {
			return FRIGHT_CHECK_TITLE;
		}
	},
	/** The Perception attribute. */
	PERCEPTION {
		@Override
		public String toString() {
			return PERCEPTION_TITLE;
		}
	},
	/** The Vision attribute. */
	VISION {
		@Override
		public String toString() {
			return VISION_TITLE;
		}
	},
	/** The Hearing attribute. */
	HEARING {
		@Override
		public String toString() {
			return HEARING_TITLE;
		}
	},
	/** The TasteSmell attribute. */
	TASTE_SMELL {
		@Override
		public String toString() {
			return TASTE_SMELL_TITLE;
		}
	},
	/** The Touch attribute. */
	TOUCH {
		@Override
		public String toString() {
			return TOUCH_TITLE;
		}
	},
	/** The Dodge attribute. */
	DODGE {
		@Override
		public String toString() {
			return DODGE_TITLE;
		}
	},
	/** The Dodge attribute. */
	PARRY {
		@Override
		public String toString() {
			return PARRY_TITLE;
		}
	},
	/** The Dodge attribute. */
	BLOCK {
		@Override
		public String toString() {
			return BLOCK_TITLE;
		}
	},
	/** The Speed attribute. */
	SPEED {
		@Override
		public String toString() {
			return SPEED_TITLE;
		}

		@Override
		public boolean isIntegerOnly() {
			return false;
		}
	},
	/** The Move attribute. */
	MOVE {
		@Override
		public String toString() {
			return MOVE_TITLE;
		}
	},
	/** The FP attribute. */
	FP {
		@Override
		public String toString() {
			return FP_TITLE;
		}
	},
	/** The HP attribute. */
	HP {
		@Override
		public String toString() {
			return HP_TITLE;
		}
	},
	/** The size modifier attribute. */
	SM {
		@Override
		public String toString() {
			return SM_TITLE;
		}
	};

	@Localize("to ST")
	@Localize(locale = "ru", value = "СЛ")
	static String	ST_TITLE;
	@Localize("to DX")
	@Localize(locale = "ru", value = "ЛВ")
	static String	DX_TITLE;
	@Localize("to IQ")
	@Localize(locale = "ru", value = "ИН")
	static String	IQ_TITLE;
	@Localize("to HT")
	@Localize(locale = "ru", value = "ЗД")
	static String	HT_TITLE;
	@Localize("to will")
	@Localize(locale = "ru", value = "воле")
	static String	WILL_TITLE;
	@Localize("to fright checks")
	@Localize(locale = "ru", value = "проверкам страха")
	static String	FRIGHT_CHECK_TITLE;
	@Localize("to perception")
	@Localize(locale = "ru", value = "восприятию")
	static String	PERCEPTION_TITLE;
	@Localize("to vision")
	@Localize(locale = "ru", value = "зрению")
	static String	VISION_TITLE;
	@Localize("to hearing")
	@Localize(locale = "ru", value = "слуху")
	static String	HEARING_TITLE;
	@Localize("to taste & smell")
	@Localize(locale = "ru", value = "вкусу и запаху")
	static String	TASTE_SMELL_TITLE;
	@Localize("to touch")
	@Localize(locale = "ru", value = "осязанию")
	static String	TOUCH_TITLE;
	@Localize("to dodge")
	@Localize(locale = "ru", value = "уклонению")
	static String	DODGE_TITLE;
	@Localize("to parry")
	@Localize(locale = "ru", value = "парированию")
	static String	PARRY_TITLE;
	@Localize("to block")
	@Localize(locale = "ru", value = "блоку")
	static String	BLOCK_TITLE;
	@Localize("to basic speed")
	@Localize(locale = "ru", value = "базовой скорости")
	static String	SPEED_TITLE;
	@Localize("to basic move")
	@Localize(locale = "ru", value = "базовому движению")
	static String	MOVE_TITLE;
	@Localize("to FP")
	@Localize(locale = "ru", value = "ЕУ")
	static String	FP_TITLE;
	@Localize("to HP")
	@Localize(locale = "ru", value = "ЕЖ")
	static String	HP_TITLE;
	@Localize("to size modifier")
	@Localize(locale = "ru", value = "модификатору размера")
	static String	SM_TITLE;

	static {
		Localization.initialize();
	}

	private String	mTag;

	private BonusAttributeType() {
		mTag = name();
		if (mTag.length() > 2) {
			mTag = mTag.toLowerCase();
		}
	}

	/** @return The presentation name. */
	public String getPresentationName() {
		String name = name();
		if (name.length() > 2) {
			name = Character.toUpperCase(name.charAt(0)) + name.substring(1).toLowerCase();
		}
		return name;
	}

	/** @return <code>true</code> if only integer values are permitted. */
	@SuppressWarnings("static-method")
	public boolean isIntegerOnly() {
		return true;
	}

	/** @return The XML tag to use for this {@link BonusAttributeType}. */
	public String getXMLTag() {
		return mTag;
	}
}
