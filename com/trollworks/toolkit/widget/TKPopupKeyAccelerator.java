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

import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;

/** A helper class that allows {@link TKPopupMenu}s to be changed via the keyboard. */
public class TKPopupKeyAccelerator implements KeyListener {
	private TKPopupMenu	mPopup;

	/**
	 * Creates a new {@link TKPopupKeyAccelerator}.
	 * 
	 * @param popup The {@link TKPopupMenu} to control.
	 */
	public TKPopupKeyAccelerator(TKPopupMenu popup) {
		mPopup = popup;
	}

	public void keyTyped(KeyEvent event) {
		char ch = Character.toUpperCase(event.getKeyChar());
		int count = mPopup.getMenuItemCount();

		for (int i = 0; i < count; i++) {
			String title = mPopup.getMenuItem(i).getTitle();

			if (title != null && title.length() > 0) {
				char menuCh = Character.toUpperCase(title.charAt(0));

				if (ch == menuCh) {
					mPopup.setSelectedItem(i);
					event.consume();
					break;
				}
			}
		}
	}

	public void keyPressed(KeyEvent event) {
		// Not used
	}

	public void keyReleased(KeyEvent event) {
		// Not used
	}
}
