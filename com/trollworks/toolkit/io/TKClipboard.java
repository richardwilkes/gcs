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

import com.trollworks.toolkit.text.TKTextUtility;

import java.awt.Toolkit;
import java.awt.datatransfer.DataFlavor;
import java.awt.datatransfer.StringSelection;
import java.awt.datatransfer.Transferable;

/** Provides standardized access to the system clipboard. */
public final class TKClipboard {
	/**
	 * Returns any text that may be present on the system clipboard. The result will have its line
	 * endings standardized to '\n'.
	 * 
	 * @return The text, or <code>null</code> if no text was present.
	 */
	public static final String getString() {
		return getString(true);
	}

	/**
	 * Returns any text that may be present on the system clipboard.
	 * 
	 * @param standardize Pass in <code>true</code> to have the text's line endings standardized
	 *            to '\n'.
	 * @return The text, or <code>null</code> if no text was present.
	 */
	public static final String getString(boolean standardize) {
		try {
			Transferable contents = Toolkit.getDefaultToolkit().getSystemClipboard().getContents(null);

			if (contents != null && contents.isDataFlavorSupported(DataFlavor.stringFlavor)) {
				String data = (String) contents.getTransferData(DataFlavor.stringFlavor);

				if (standardize && data != null) {
					data = TKTextUtility.standardizeLineEndings(data);
				}
				return data;
			}
		} catch (Exception exception) {
			// Ignore, as some conditions (at least under Linux) can cause an IOException
		}
		return null;
	}

	/** @return <code>true</code> if the system clipboard contains text. */
	public static final boolean hasText() {
		return getString(false) != null;
	}

	/**
	 * Puts text onto the system clipboard.
	 * 
	 * @param text The text to place on the system clipboard.
	 */
	public static final void putString(String text) {
		try {
			StringSelection clip = new StringSelection(text);

			Toolkit.getDefaultToolkit().getSystemClipboard().setContents(clip, null);
		} catch (Exception exception) {
			// Ignore, as some conditions (at least under Linux) can cause an IOException
		}
	}
}
