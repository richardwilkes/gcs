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

package com.trollworks.gcs.ui.editor.prereq;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CSAdvantagePrereq} and {@link CSSkillPrereq}. */
	public static String	WHOSE_NAME;
	/** Used by {@link CSAdvantagePrereq} and {@link CSSkillPrereq}. */
	public static String	WHOSE_LEVEL;

	/** Used by {@link CSAdvantagePrereq}. */
	public static String	WHOSE_NOTES;

	/** Used by {@link CSAttributePrereq}. */
	public static String	COMBINED_WITH;

	/** Used by {@link CSAttributePrereq} and {@link CSContainedWeightPrereq}. */
	public static String	WHICH;

	/** Used by {@link CSBasePrereq}. */
	public static String	AND;
	/** Used by {@link CSBasePrereq}. */
	public static String	OR;
	/** Used by {@link CSBasePrereq}. */
	public static String	REMOVE_PREREQ_TOOLTIP;
	/** Used by {@link CSBasePrereq}. */
	public static String	REMOVE_PREREQ_LIST_TOOLTIP;
	/** Used by {@link CSBasePrereq}. */
	public static String	HAS;
	/** Used by {@link CSBasePrereq}. */
	public static String	DOES_NOT_HAVE;
	/** Used by {@link CSBasePrereq}. */
	public static String	ADVANTAGE;
	/** Used by {@link CSBasePrereq}. */
	public static String	ATTRIBUTE;
	/** Used by {@link CSBasePrereq}. */
	public static String	CONTAINED_WEIGHT;
	/** Used by {@link CSBasePrereq}. */
	public static String	SKILL;
	/** Used by {@link CSBasePrereq}. */
	public static String	SPELL;
	/** Used by {@link CSBasePrereq}. */
	public static String	EXACTLY;

	/** Used by {@link CSListPrereq}. */
	public static String	REQUIRES_ALL;
	/** Used by {@link CSListPrereq}. */
	public static String	REQUIRES_ANY;
	/** Used by {@link CSListPrereq}. */
	public static String	ADD_PREREQ_TOOLTIP;
	/** Used by {@link CSListPrereq}. */
	public static String	ADD_PREREQ_LIST_TOOLTIP;

	/** Used by {@link CSPrereqs}. */
	public static String	PREREQUISITES;

	/** Used by {@link CSSkillPrereq}. */
	public static String	WHOSE_SPECIALIZATION;

	/** Used by {@link CSSpellPrereq}. */
	public static String	WHOSE_SPELL_NAME;
	/** Used by {@link CSSpellPrereq}. */
	public static String	ANY;
	/** Used by {@link CSSpellPrereq}. */
	public static String	COLLEGE;
	/** Used by {@link CSSpellPrereq}. */
	public static String	COLLEGE_COUNT;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
