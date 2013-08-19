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

package com.trollworks.gcs;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CSMain}. */
	public static String	PDF_OPTION;
	/** Used by {@link CSMain}. */
	public static String	HTML_OPTION;
	/** Used by {@link CSMain}. */
	public static String	HTML_TEMPLATE_OPTION;
	/** Used by {@link CSMain}. */
	public static String	HTML_TEMPLATE_ARG;
	/** Used by {@link CSMain}. */
	public static String	PNG_OPTION;
	/** Used by {@link CSMain}. */
	public static String	SIZE_OPTION;
	/** Used by {@link CSMain}. */
	public static String	MARGIN_OPTION;
	/** Used by {@link CSMain}. */
	public static String	APP_NAME;
	/** Used by {@link CSMain}. */
	public static String	APP_VERSION;
	/** Used by {@link CSMain}. */
	public static String	APP_COPYRIGHT_YEARS;
	/** Used by {@link CSMain}. */
	public static String	APP_COPYRIGHT_OWNER;

	/** Used by {@link CSBatchApplication}. */
	public static String	NO_FILES_TO_PROCESS;
	/** Used by {@link CSBatchApplication}. */
	public static String	LOADING;
	/** Used by {@link CSBatchApplication}. */
	public static String	CREATING_PDF;
	/** Used by {@link CSBatchApplication}. */
	public static String	CREATING_HTML;
	/** Used by {@link CSBatchApplication}. */
	public static String	CREATING_PNG;
	/** Used by {@link CSBatchApplication}. */
	public static String	PROCESSING_FAILED;
	/** Used by {@link CSBatchApplication}. */
	public static String	FINISHED;
	/** Used by {@link CSBatchApplication}. */
	public static String	CREATED;
	/** Used by {@link CSBatchApplication}. */
	public static String	INVALID_PAPER_SIZE;
	/** Used by {@link CSBatchApplication}. */
	public static String	INVALID_PAPER_MARGINS;
	/** Used by {@link CSBatchApplication}. */
	public static String	TEMPLATE_USED;

	/** Used by {@link CSApplication}. */
	public static String	SHEET_DESCRIPTION;
	/** Used by {@link CSApplication}. */
	public static String	TEMPLATE_DESCRIPTION;
	/** Used by {@link CSApplication}. */
	public static String	TRAITS_DESCRIPTION;
	/** Used by {@link CSApplication}. */
	public static String	EQUIPMENT_DESCRIPTION;
	/** Used by {@link CSApplication}. */
	public static String	SKILLS_DESCRIPTION;
	/** Used by {@link CSApplication}. */
	public static String	SPELLS_DESCRIPTION;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
