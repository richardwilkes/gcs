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

package com.trollworks.toolkit.text;

import com.trollworks.toolkit.utility.TKKeystroke;
import com.trollworks.toolkit.widget.TKKeyEventFilter;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKTextField;

import java.awt.event.KeyEvent;
import java.text.DecimalFormat;
import java.text.DecimalFormatSymbols;
import java.text.NumberFormat;

/** A standard numeric key entry filter. */
public class TKNumberFilter implements TKKeyEventFilter {
	private static final char	GROUP_CHAR;
	private static final char	DECIMAL_CHAR;
	private boolean				mIsHeightFilter;
	private boolean				mAllowDecimal;
	private boolean				mAllowSign;
	private boolean				mAllowGroup;
	private int					mMaxDigits;

	static {
		DecimalFormatSymbols symbols = ((DecimalFormat) NumberFormat.getNumberInstance()).getDecimalFormatSymbols();

		GROUP_CHAR = symbols.getGroupingSeparator();
		DECIMAL_CHAR = symbols.getDecimalSeparator();
	}

	/** Creates a new height key entry filter. */
	public TKNumberFilter() {
		this(false, false, false, Integer.MAX_VALUE);
		mIsHeightFilter = true;
	}

	/**
	 * Creates a new numeric key entry filter.
	 * 
	 * @param allowDecimal Pass in <code>true</code> to allow floating point.
	 * @param allowSign Pass in <code>true</code> to allow sign characters.
	 */
	public TKNumberFilter(boolean allowDecimal, boolean allowSign) {
		this(allowDecimal, allowSign, true, Integer.MAX_VALUE);
	}

	/**
	 * Creates a new numeric key entry filter.
	 * 
	 * @param allowDecimal Pass in <code>true</code> to allow floating point.
	 * @param allowSign Pass in <code>true</code> to allow sign characters.
	 * @param allowGroup Pass in <code>true</code> to allow group characters.
	 */
	public TKNumberFilter(boolean allowDecimal, boolean allowSign, boolean allowGroup) {
		this(allowDecimal, allowSign, allowGroup, Integer.MAX_VALUE);
	}

	/**
	 * Creates a new numeric key entry filter.
	 * 
	 * @param allowDecimal Pass in <code>true</code> to allow floating point.
	 * @param allowSign Pass in <code>true</code> to allow sign characters.
	 * @param maxDigits The maximum number of digits (not necessarily characters) the field can
	 *            have.
	 */
	public TKNumberFilter(boolean allowDecimal, boolean allowSign, int maxDigits) {
		this(allowDecimal, allowSign, true, maxDigits);
	}

	/**
	 * Creates a new numeric key entry filter.
	 * 
	 * @param allowDecimal Pass in <code>true</code> to allow floating point.
	 * @param allowSign Pass in <code>true</code> to allow sign characters.
	 * @param allowGroup Pass in <code>true</code> to allow group characters.
	 * @param maxDigits The maximum number of digits (not necessarily characters) the field can
	 *            have.
	 */
	public TKNumberFilter(boolean allowDecimal, boolean allowSign, boolean allowGroup, int maxDigits) {
		mAllowDecimal = allowDecimal;
		mAllowSign = allowSign;
		mAllowGroup = allowGroup;
		mMaxDigits = maxDigits;
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

		if (mAllowGroup && ch == GROUP_CHAR || ch >= '0' && ch <= '9' || mAllowSign && (ch == '-' || ch == '+') || mAllowDecimal && ch == DECIMAL_CHAR || mIsHeightFilter && (ch == '\'' || ch == '"' || ch == ' ')) {
			if (owner instanceof TKTextField) {
				TKTextField field = (TKTextField) owner;
				StringBuilder buffer = new StringBuilder(field.getText());
				int start = field.getSelectionStart();
				int end = field.getSelectionEnd();

				if (start != end) {
					buffer.delete(start, end);
				}

				if (ch >= '0' && ch <= '9') {
					int length = buffer.length();
					int count = 0;

					for (int i = 0; i < length; i++) {
						char one = buffer.charAt(i);

						if (one >= '0' && ch <= '9') {
							count++;
						}
					}
					if (count >= mMaxDigits) {
						return true;
					}
				}

				if (ch == GROUP_CHAR || ch >= '0' && ch <= '9') {
					return mAllowSign && start == 0 && buffer.length() > 0 && (buffer.charAt(0) == '-' || buffer.charAt(0) == '+');
				} else if (ch == '-' || ch == '+') {
					return start != 0;
				} else if (ch == DECIMAL_CHAR) {
					return buffer.indexOf("" + DECIMAL_CHAR) != -1 || mAllowSign && start == 0 && buffer.length() > 0 && (buffer.charAt(0) == '-' || buffer.charAt(0) == '+'); //$NON-NLS-1$
				} else if (ch == '\'') {
					return buffer.indexOf("'") != -1; //$NON-NLS-1$
				} else if (ch == '"') {
					return buffer.indexOf("\"") != -1; //$NON-NLS-1$
				}
			}
			return false;
		}

		return true;
	}
}
