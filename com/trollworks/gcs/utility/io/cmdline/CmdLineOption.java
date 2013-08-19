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

package com.trollworks.gcs.utility.io.cmdline;

/** Describes a command-line option. */
public class CmdLineOption {
	private String[]	mNames;
	private String		mDescription;
	private String		mArgumentLabel;

	/**
	 * Creates a new {@link CmdLineOption}.
	 * 
	 * @param description The description of this option.
	 * @param argumentLabel The name of this option's argument, if it has one. Use <code>null</code>
	 *            if it doesn't.
	 * @param names One or more names that can be used to invoke this option.
	 */
	public CmdLineOption(String description, String argumentLabel, String... names) {
		mNames = new String[names.length];
		mDescription = description;
		mArgumentLabel = argumentLabel;
		for (int i = 0; i < names.length; i++) {
			mNames[i] = names[i].trim().toLowerCase();
		}
	}

	/** @return The description of this option. */
	public String getDescription() {
		return mDescription;
	}

	/** @return The names this option can be invoked with. */
	public String[] getNames() {
		return mNames;
	}

	/** @return Whether this option takes an argument. */
	public boolean takesArgument() {
		return mArgumentLabel != null;
	}

	/** @return The argument label, if any. */
	public String getArgumentLabel() {
		return mArgumentLabel;
	}
}
