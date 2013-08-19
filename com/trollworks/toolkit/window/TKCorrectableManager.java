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

package com.trollworks.toolkit.window;

import com.trollworks.toolkit.widget.TKCorrectable;

import java.awt.KeyboardFocusManager;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.util.HashSet;

/** Manages all {@link TKCorrectable}s. */
public class TKCorrectableManager implements PropertyChangeListener {
	private static TKCorrectableManager	INSTANCE;
	private HashSet<TKCorrectable>		mCorrectables;
	private boolean						mIgnoreNext;

	/** @return The one and only instance of this class that can exist. */
	public static final TKCorrectableManager getInstance() {
		if (INSTANCE == null) {
			INSTANCE = new TKCorrectableManager();
		}
		return INSTANCE;
	}

	private TKCorrectableManager() {
		mCorrectables = new HashSet<TKCorrectable>();
		KeyboardFocusManager.getCurrentKeyboardFocusManager().addPropertyChangeListener("permanentFocusOwner", this); //$NON-NLS-1$
	}

	public void propertyChange(PropertyChangeEvent event) {
		if (mIgnoreNext) {
			mIgnoreNext = false;
		} else {
			clearAllCorrectables();
		}
	}

	/**
	 * Registers a {@link TKCorrectable} so that this manager can clear its corrected state on focus
	 * change within the window.
	 * 
	 * @param correctable The {@link TKCorrectable} to register.
	 */
	public void register(TKCorrectable correctable) {
		mCorrectables.add(correctable);
		mIgnoreNext = true;
	}

	/** Clears all {@link TKCorrectable}s that have registered. */
	public void clearAllCorrectables() {
		for (TKCorrectable field : mCorrectables) {
			field.clearCorrectionState();
		}
		mCorrectables.clear();
	}
}
