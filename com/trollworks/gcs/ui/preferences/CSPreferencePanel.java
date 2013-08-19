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

package com.trollworks.gcs.ui.preferences;

import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;

/** The abstract base class for all preference panels. */
public abstract class CSPreferencePanel extends TKPanel {
	private String	mTitle;

	/**
	 * Creates a new preference panel.
	 * 
	 * @param title The title for this panel.
	 */
	public CSPreferencePanel(String title) {
		super(new TKColumnLayout());
		mTitle = title;
	}

	/** Resets this panel back to its defaults. */
	public abstract void reset();

	/** @return Whether the panel is currently set to defaults or not. */
	public abstract boolean isSetToDefaults();

	/** Call to adjust the reset button for any changes that have been made. */
	protected void adjustResetButton() {
		CSPreferencesWindow window = (CSPreferencesWindow) getBaseWindow();

		if (window != null) {
			window.adjustResetButton();
		}
	}

	@Override public String toString() {
		return mTitle;
	}
}
