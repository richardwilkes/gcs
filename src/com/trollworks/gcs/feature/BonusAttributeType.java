/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
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

	@Localize("ST")
	@Localize(locale = "de", value = "ST")
	@Localize(locale = "ru", value = "СЛ")
	@Localize(locale = "es", value = "FZ")
	static String	ST_TITLE;
	@Localize("DX")
	@Localize(locale = "de", value = "GE")
	@Localize(locale = "ru", value = "ЛВ")
	@Localize(locale = "es", value = "DS")
	static String	DX_TITLE;
	@Localize("IQ")
	@Localize(locale = "de", value = "IQ")
	@Localize(locale = "ru", value = "ИН")
	@Localize(locale = "es", value = "CI")
	static String	IQ_TITLE;
	@Localize("HT")
	@Localize(locale = "de", value = "KO")
	@Localize(locale = "ru", value = "ЗД")
	@Localize(locale = "es", value = "SL")
	static String	HT_TITLE;
	@Localize("will")
	@Localize(locale = "de", value = "Wille")
	@Localize(locale = "ru", value = "воле")
	@Localize(locale = "es", value = "Vol")
	static String	WILL_TITLE;
	@Localize("fright checks")
	@Localize(locale = "de", value = "Schreckproben")
	@Localize(locale = "ru", value = "проверкам страха")
	@Localize(locale = "es", value = "Chequeos de Miedo")
	static String	FRIGHT_CHECK_TITLE;
	@Localize("perception")
	@Localize(locale = "de", value = "Wahrnehmung")
	@Localize(locale = "ru", value = "восприятию")
	@Localize(locale = "es", value = "Perceción")
	static String	PERCEPTION_TITLE;
	@Localize("vision")
	@Localize(locale = "de", value = "Sehen")
	@Localize(locale = "ru", value = "зрению")
	@Localize(locale = "es", value = "Visión")
	static String	VISION_TITLE;
	@Localize("hearing")
	@Localize(locale = "de", value = "Hören")
	@Localize(locale = "ru", value = "слуху")
	@Localize(locale = "es", value = "Oido")
	static String	HEARING_TITLE;
	@Localize("taste & smell")
	@Localize(locale = "de", value = "Schmecken und Riechen")
	@Localize(locale = "ru", value = "вкусу и запаху")
	@Localize(locale = "es", value = "Gusto y olfato")
	static String	TASTE_SMELL_TITLE;
	@Localize("touch")
	@Localize(locale = "de", value = "Fühlen")
	@Localize(locale = "ru", value = "осязанию")
	@Localize(locale = "es", value = "Tacto")
	static String	TOUCH_TITLE;
	@Localize("dodge")
	@Localize(locale = "de", value = "Ausweichen")
	@Localize(locale = "ru", value = "уклонению")
	@Localize(locale = "es", value = "esquiva")
	static String	DODGE_TITLE;
	@Localize("parry")
	@Localize(locale = "de", value = "Parieren")
	@Localize(locale = "ru", value = "парированию")
	@Localize(locale = "es", value = "parada")
	static String	PARRY_TITLE;
	@Localize("block")
	@Localize(locale = "de", value = "Abblocken")
	@Localize(locale = "ru", value = "блоку")
	@Localize(locale = "es", value = "bloqueo")
	static String	BLOCK_TITLE;
	@Localize("basic speed")
	@Localize(locale = "de", value = "Grundgeschwindigkeit")
	@Localize(locale = "ru", value = "базовой скорости")
	@Localize(locale = "es", value = "velocidad básica")
	static String	SPEED_TITLE;
	@Localize("basic move")
	@Localize(locale = "de", value = "Grundbewegung")
	@Localize(locale = "ru", value = "базовому движению")
	@Localize(locale = "es", value = "movimiento básico")
	static String	MOVE_TITLE;
	@Localize("FP")
	@Localize(locale = "de", value = "EP")
	@Localize(locale = "ru", value = "ЕУ")
	@Localize(locale = "es", value = "PF")
	static String	FP_TITLE;
	@Localize("HP")
	@Localize(locale = "de", value = "TP")
	@Localize(locale = "ru", value = "ЕЖ")
	@Localize(locale = "es", value = "PV")
	static String	HP_TITLE;
	@Localize("size modifier")
	@Localize(locale = "de", value = "Größenmodifikator")
	@Localize(locale = "ru", value = "модификатору размера")
	@Localize(locale = "es", value = "Modificador por tamaño")
	static String	SM_TITLE;

	static {
		Localization.initialize();
	}

	private String mTag;

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
