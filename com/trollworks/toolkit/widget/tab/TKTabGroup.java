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

package com.trollworks.toolkit.widget.tab;

import com.trollworks.toolkit.widget.button.TKToggleButton;
import com.trollworks.toolkit.widget.button.TKToggleButtonGroup;

/**
 * Manages a group of {@link TKTab}s which cannot be selected simulaneously and updates a
 * {@link TKTabbedPanel} when a new tab is selected.
 */
public class TKTabGroup extends TKToggleButtonGroup {
	private TKTabbedPanel	mTabbedPanel;

	/**
	 * Constructs a tab group associated with the given tabbed panel.
	 * 
	 * @param tabbedPanel The tabbed panel of this group.
	 */
	public TKTabGroup(TKTabbedPanel tabbedPanel) {
		super();
		mTabbedPanel = tabbedPanel;
	}

	/**
	 * @return The list of tabs as a tab array.
	 */
	public TKTab[] getTabs() {
		return getButtonList().toArray(new TKTab[0]);
	}

	@Override public boolean setSelection(TKToggleButton button) {
		if (super.setSelection(button)) {
			mTabbedPanel.resetSelectedPanel();
			mTabbedPanel.getTabPanel().revalidate();
			return true;
		}
		return false;
	}
}
