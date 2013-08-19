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

package com.trollworks.gcs.common;

import com.trollworks.gcs.utility.io.LocalizedMessages;

import java.io.IOException;

/** An exception for data files that are too new to be loaded. */
public class NewerVersionException extends IOException {
	private static String	MSG_VERSION_NEWER;

	static {
		LocalizedMessages.initialize(NewerVersionException.class);
	}

	/** Creates a new {@link NewerVersionException}. */
	public NewerVersionException() {
		super(MSG_VERSION_NEWER);
	}
}
