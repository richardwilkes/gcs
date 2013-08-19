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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.utility.io.LocalizedMessages;

/** The attribute affected by a {@link AttributeBonus}. */
public enum BonusAttributeType {
	/** The ST attribute. */
	ST {
		@Override public String toString() {
			return MSG_ST;
		}
	},
	/** The DX attribute. */
	DX {
		@Override public String toString() {
			return MSG_DX;
		}
	},
	/** The IQ attribute. */
	IQ {
		@Override public String toString() {
			return MSG_IQ;
		}
	},
	/** The HT attribute. */
	HT {
		@Override public String toString() {
			return MSG_HT;
		}
	},
	/** The Will attribute. */
	WILL {
		@Override public String toString() {
			return MSG_WILL;
		}
	},
	/** The Fright Check attribute. */
	FRIGHT_CHECK {
		@Override public String toString() {
			return MSG_FRIGHT_CHECK;
		}
	},
	/** The Perception attribute. */
	PERCEPTION {
		@Override public String toString() {
			return MSG_PERCEPTION;
		}
	},
	/** The Vision attribute. */
	VISION {
		@Override public String toString() {
			return MSG_VISION;
		}
	},
	/** The Hearing attribute. */
	HEARING {
		@Override public String toString() {
			return MSG_HEARING;
		}
	},
	/** The TasteSmell attribute. */
	TASTE_SMELL {
		@Override public String toString() {
			return MSG_TASTE_SMELL;
		}
	},
	/** The Touch attribute. */
	TOUCH {
		@Override public String toString() {
			return MSG_TOUCH;
		}
	},
	/** The Dodge attribute. */
	DODGE {
		@Override public String toString() {
			return MSG_DODGE;
		}
	},
	/** The Dodge attribute. */
	PARRY {
		@Override public String toString() {
			return MSG_PARRY;
		}
	},
	/** The Dodge attribute. */
	BLOCK {
		@Override public String toString() {
			return MSG_BLOCK;
		}
	},
	/** The Speed attribute. */
	SPEED {
		@Override public String toString() {
			return MSG_SPEED;
		}

		@Override public boolean isIntegerOnly() {
			return false;
		}
	},
	/** The Move attribute. */
	MOVE {
		@Override public String toString() {
			return MSG_MOVE;
		}
	},
	/** The FP attribute. */
	FP {
		@Override public String toString() {
			return MSG_FP;
		}
	},
	/** The HP attribute. */
	HP {
		@Override public String toString() {
			return MSG_HP;
		}
	},
	/** The size modifier attribute. */
	SM {
		@Override public String toString() {
			return MSG_SM;
		}
	};

	static String	MSG_ST;
	static String	MSG_DX;
	static String	MSG_IQ;
	static String	MSG_HT;
	static String	MSG_WILL;
	static String	MSG_FRIGHT_CHECK;
	static String	MSG_PERCEPTION;
	static String	MSG_VISION;
	static String	MSG_HEARING;
	static String	MSG_TASTE_SMELL;
	static String	MSG_TOUCH;
	static String	MSG_DODGE;
	static String	MSG_PARRY;
	static String	MSG_BLOCK;
	static String	MSG_SPEED;
	static String	MSG_MOVE;
	static String	MSG_FP;
	static String	MSG_HP;
	static String	MSG_SM;

	private String	mTag;

	static {
		LocalizedMessages.initialize(BonusAttributeType.class);
	}

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
	public boolean isIntegerOnly() {
		return true;
	}

	/** @return The XML tag to use for this {@link BonusAttributeType}. */
	public String getXMLTag() {
		return mTag;
	}
}
