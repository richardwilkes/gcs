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

package com.trollworks.toolkit.io;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link TKPreferences}. */
	public static String	UNINITIALIZED;

	/** Used by {@link TKFileFilter}. */
	public static String	ALL_FILES;
	/** Used by {@link TKFileFilter}. */
	public static String	DIRECTORIES_ONLY;
	/** Used by {@link TKFileFilter}. */
	public static String	AND;

	/** Used by {@link TKSafeFileUpdater}. */
	public static String	NO_TRANSACTION_IN_PROGRESS;
	/** Used by {@link TKSafeFileUpdater}. */
	public static String	FILE_SWAP_FAILED;
	/** Used by {@link TKSafeFileUpdater}. */
	public static String	MAY_NOT_BE_NULL;
	/** Used by {@link TKSafeFileUpdater}. */
	public static String	MAY_NOT_BE_DIRECTORY;

	/** Used by {@link TKUpdateChecker}. */
	public static String	CHECKING;
	/** Used by {@link TKUpdateChecker}. */
	public static String	UP_TO_DATE;
	/** Used by {@link TKUpdateChecker}. */
	public static String	OUT_OF_DATE;
	/** Used by {@link TKUpdateChecker}. */
	public static String	UPDATE_TITLE;
	/** Used by {@link TKUpdateChecker}. */
	public static String	IGNORE_TITLE;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
