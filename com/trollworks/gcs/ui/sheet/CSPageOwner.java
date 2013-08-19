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

package com.trollworks.gcs.ui.sheet;

import com.trollworks.toolkit.print.TKPrintManager;

import java.awt.Graphics2D;
import java.awt.Insets;

/** Objects which control printable pages must implement this interface. */
public interface CSPageOwner {
	/** @return The page settings. */
	public TKPrintManager getPageSettings();

	/**
	 * Called so the page owner can draw headers, footers, etc.
	 * 
	 * @param page The page to work on.
	 * @param g2d The graphics object to work with.
	 */
	public void drawPageAdornments(CSPage page, Graphics2D g2d);

	/**
	 * @param page The page to work on.
	 * @return The insets required for the page adornments.
	 */
	public Insets getPageAdornmentsInsets(CSPage page);
}
