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

package com.trollworks.toolkit.widget;

import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.window.TKCorrectableManager;

/** A correctable popup menu. */
public class TKCorrectablePopupMenu extends TKPopupMenu implements TKCorrectable {
	private TKCorrectableLabel	mLabel;
	private boolean				mWasCorrected;
	private String				mSavedToolTip;

	/**
	 * Creates a popup menu.
	 * 
	 * @param label The {@link TKCorrectableLabel} to pair with.
	 * @param menu The menu to be displayed.
	 */
	public TKCorrectablePopupMenu(TKCorrectableLabel label, TKMenu menu) {
		super(menu);
		setLabel(label);
	}

	/** @param label The {@link TKCorrectableLabel} to pair with. */
	public void setLabel(TKCorrectableLabel label) {
		mLabel = label;
		if (mLabel != null) {
			mLabel.setLink(this);
		}
	}

	public boolean wasCorrected() {
		return mWasCorrected;
	}

	public void clearCorrectionState() {
		if (mWasCorrected) {
			mWasCorrected = false;
			setToolTipText(mSavedToolTip);
			mSavedToolTip = null;
			if (mLabel != null) {
				mLabel.repaint();
			}
		}
	}

	public void correct(Object correction, String reason) {
		preCorrect(reason);
		setSelectedUserObject(correction);
		postCorrect();
	}

	/**
	 * Corrects this item.
	 * 
	 * @param correction The corrected popup menu item index.
	 * @param reason The reason for the change.
	 */
	public void correct(int correction, String reason) {
		preCorrect(reason);
		setSelectedItem(correction);
		postCorrect();
	}

	/**
	 * Corrects this item.
	 * 
	 * @param correction The corrected popup menu item title.
	 * @param reason The reason for the change.
	 */
	public void correct(String correction, String reason) {
		preCorrect(reason);
		setSelectedItem(correction);
		postCorrect();
	}

	/**
	 * Corrects this item.
	 * 
	 * @param correction The corrected popup menu item.
	 * @param reason The reason for the change.
	 */
	public void correct(TKMenuItem correction, String reason) {
		preCorrect(reason);
		setSelectedItem(correction);
		postCorrect();
	}

	private void preCorrect(String reason) {
		if (mSavedToolTip == null) {
			mSavedToolTip = getToolTipText();
		}
		mWasCorrected = true;
		setToolTipText(reason);
	}

	private void postCorrect() {
		if (mLabel != null) {
			mLabel.repaint();
		}
		TKCorrectableManager.getInstance().register(this);
	}
}
