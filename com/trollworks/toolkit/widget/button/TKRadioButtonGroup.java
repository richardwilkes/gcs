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

package com.trollworks.toolkit.widget.button;

import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;

/**
 * A standard radio button group, which allows only one of the radio buttons within a group to be
 * selected at a time.
 */
public class TKRadioButtonGroup implements ActionListener {
	private ArrayList<TKRadioButton>	mButtons;
	private TKRadioButton				mSelection;

	/** Creates a new, empty group. */
	public TKRadioButtonGroup() {
		mButtons = new ArrayList<TKRadioButton>();
	}

	public void actionPerformed(ActionEvent event) {
		setSelection((TKRadioButton) event.getSource());
	}

	/**
	 * Add a radio button to the group.
	 * 
	 * @param button The button to add.
	 */
	public void add(TKRadioButton button) {
		if (!mButtons.contains(button)) {
			if (mButtons.size() == 0) {
				button.setSelected(true);
			}
			mButtons.add(button);
			button.addActionListener(this);
			if (button.isSelected()) {
				setSelection(button);
			}
		}
	}

	/**
	 * Remove a radio button from the group.
	 * 
	 * @param button The button to remove.
	 */
	public void remove(TKRadioButton button) {
		mButtons.remove(button);
		button.removeActionListener(this);
		if (button == mSelection) {
			mSelection = null;
			if (mButtons.size() > 0) {
				mButtons.get(0).setSelected(true);
			}
		}
	}

	/** Removes all radio buttons from the group. */
	public void clear() {
		for (TKRadioButton button : mButtons) {
			button.removeActionListener(this);
		}
		mButtons.clear();
	}

	/** @return The buttons that comprise this group. */
	public TKRadioButton[] getButtons() {
		return mButtons.toArray(new TKRadioButton[0]);
	}

	/** @return The currently selected radio button. */
	public TKRadioButton getSelection() {
		return mSelection;
	}

	/** @param button The button to make the current selection. */
	public void setSelection(TKRadioButton button) {
		if (button != null && button.isSelected() && mSelection != button && mButtons.contains(button)) {
			if (mSelection != null) {
				mSelection.setSelected(false);
			}
			mSelection = button;
		}
	}
}
