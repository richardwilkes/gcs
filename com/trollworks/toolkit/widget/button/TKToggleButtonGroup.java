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
import java.util.Collections;
import java.util.List;

/**
 * A standard toggle button group, which allows only one of the toggle buttons within a group to be
 * selected at a time.
 */
public class TKToggleButtonGroup implements ActionListener {
	private ArrayList<TKToggleButton>	mButtons;
	private int							mSelectedIndex;

	/** Creates a new, empty group. */
	public TKToggleButtonGroup() {
		mButtons = new ArrayList<TKToggleButton>();
	}

	public void actionPerformed(ActionEvent event) {
		setSelection((TKToggleButton) event.getSource());
	}

	/**
	 * Add a toggle button to the group.
	 * 
	 * @param button The button add.
	 */
	public void add(TKToggleButton button) {
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
	 * Remove a toggle button from the group.
	 * 
	 * @param button The button to remove.
	 */
	public void remove(TKToggleButton button) {
		TKToggleButton selection = getSelection();
		mButtons.remove(button);
		button.removeActionListener(this);
		if (button == selection) {
			mSelectedIndex = -1;
			if (mButtons.size() > 0) {
				mButtons.get(0).setSelected(true);
			}
		}
	}

	/** Removes all toggle buttons from the group. */
	public void clear() {
		for (TKToggleButton button : mButtons) {
			button.removeActionListener(this);
		}
		mButtons.clear();
	}

	/** @return The buttons that comprise this group. */
	public TKToggleButton[] getButtons() {
		return mButtons.toArray(new TKToggleButton[0]);
	}

	/** @return The buttons that comprise this group. */
	public List<TKToggleButton> getButtonList() {
		return Collections.unmodifiableList(mButtons);
	}

	/** @return The currently selected toggle button. */
	public TKToggleButton getSelection() {
		return mButtons.get(mSelectedIndex);
	}

	/**
	 * Sets the currently selected toggle button.
	 * 
	 * @param button The button to set.
	 * @return <code>true</code> if the selection was changed.
	 */
	public boolean setSelection(TKToggleButton button) {
		TKToggleButton selection = null;

		if (mSelectedIndex >= 0) {
			selection = getSelection();
		}

		if (button != null && button.isSelected() && selection != button && mButtons.contains(button)) {
			if (selection != null) {
				selection.setSelected(false);
			}
			mSelectedIndex = mButtons.indexOf(button);

			return true;
		}

		return false;
	}

	/** @return The index of the selected toggle button. */
	public int getSelectedIndex() {
		return mSelectedIndex;
	}
}
