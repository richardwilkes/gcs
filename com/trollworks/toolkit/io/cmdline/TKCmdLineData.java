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

package com.trollworks.toolkit.io.cmdline;

/** Provides storage of processed command-line arguments. */
class TKCmdLineData {
	private TKCmdLineOption	mOption;
	private String			mArgument;

	/**
	 * Creates a new {@link TKCmdLineData}.
	 * 
	 * @param option The option.
	 */
	TKCmdLineData(TKCmdLineOption option) {
		mOption = option;
	}

	/**
	 * Creates a new {@link TKCmdLineData}.
	 * 
	 * @param argument The original command-line argument.
	 */
	TKCmdLineData(String argument) {
		mArgument = argument;
	}

	/**
	 * Creates a new {@link TKCmdLineData}.
	 * 
	 * @param option The option.
	 * @param argument The option's argument.
	 */
	TKCmdLineData(TKCmdLineOption option, String argument) {
		mOption = option;
		mArgument = argument;
	}

	/** @return Whether this is an option. */
	boolean isOption() {
		return mOption != null;
	}

	/** @return The option, or <code>null</code> if this is not an option. */
	TKCmdLineOption getOption() {
		return mOption;
	}

	/** @return The option's argument, or the original command line argument. */
	String getArgument() {
		return mArgument;
	}
}
