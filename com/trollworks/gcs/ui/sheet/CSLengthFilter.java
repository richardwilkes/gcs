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

import com.trollworks.toolkit.utility.TKKeystroke;
import com.trollworks.toolkit.widget.TKKeyEventFilter;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKTextField;

import java.awt.event.KeyEvent;

/** A filter that limits the length of a field to at most a specific number of characters. */
public class CSLengthFilter implements TKKeyEventFilter {
	private int	mMax;

	/**
	 * Creates a new filter that limits the length of a field to at most <code>max</code>
	 * characters.
	 * 
	 * @param max The maximum number of characters to allow.
	 */
	public CSLengthFilter(int max) {
		mMax = max;
	}

	public boolean filterKeyEvent(TKPanel owner, KeyEvent event, boolean isReal) {
		if (TKKeystroke.isCommandKeyDown(event)) {
			return true;
		}

		int id = event.getID();

		if (id != KeyEvent.KEY_TYPED) {
			return false;
		}

		char ch = event.getKeyChar();

		if (ch == '\n' || ch == '\r' || ch == '\t' || ch == '\b' || ch == KeyEvent.VK_DELETE) {
			return false;
		}

		if (owner instanceof TKTextField) {
			TKTextField field = (TKTextField) owner;
			StringBuffer buffer = new StringBuffer(field.getText());
			int start = field.getSelectionStart();
			int end = field.getSelectionEnd();

			if (start != end) {
				buffer.delete(start, end);
			}

			return buffer.length() >= mMax;
		}

		return true;
	}
}
