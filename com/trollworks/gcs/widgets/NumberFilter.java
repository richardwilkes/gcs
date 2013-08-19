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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets;

import java.awt.Toolkit;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.text.DecimalFormat;
import java.text.DecimalFormatSymbols;
import java.text.NumberFormat;

import javax.swing.JTextField;

/** A standard numeric key entry filter. */
public class NumberFilter implements KeyListener {
	private static final char	GROUP_CHAR;
	private static final char	DECIMAL_CHAR;
	private JTextField			mField;
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

	/**
	 * Creates a new height key entry filter.
	 * 
	 * @param field The {@link JTextField} to filter.
	 */
	public NumberFilter(JTextField field) {
		this(field, false, false, false, Integer.MAX_VALUE);
		mIsHeightFilter = true;
	}

	/**
	 * Creates a new numeric key entry filter.
	 * 
	 * @param field The {@link JTextField} to filter.
	 * @param allowDecimal Pass in <code>true</code> to allow floating point.
	 * @param allowSign Pass in <code>true</code> to allow sign characters.
	 */
	public NumberFilter(JTextField field, boolean allowDecimal, boolean allowSign) {
		this(field, allowDecimal, allowSign, true, Integer.MAX_VALUE);
	}

	/**
	 * Creates a new numeric key entry filter.
	 * 
	 * @param field The {@link JTextField} to filter.
	 * @param allowDecimal Pass in <code>true</code> to allow floating point.
	 * @param allowSign Pass in <code>true</code> to allow sign characters.
	 * @param allowGroup Pass in <code>true</code> to allow group characters.
	 */
	public NumberFilter(JTextField field, boolean allowDecimal, boolean allowSign, boolean allowGroup) {
		this(field, allowDecimal, allowSign, allowGroup, Integer.MAX_VALUE);
	}

	/**
	 * Creates a new numeric key entry filter.
	 * 
	 * @param field The {@link JTextField} to filter.
	 * @param allowDecimal Pass in <code>true</code> to allow floating point.
	 * @param allowSign Pass in <code>true</code> to allow sign characters.
	 * @param maxDigits The maximum number of digits (not necessarily characters) the field can
	 *            have.
	 */
	public NumberFilter(JTextField field, boolean allowDecimal, boolean allowSign, int maxDigits) {
		this(field, allowDecimal, allowSign, true, maxDigits);
	}

	/**
	 * Creates a new numeric key entry filter.
	 * 
	 * @param field The {@link JTextField} to filter.
	 * @param allowDecimal Pass in <code>true</code> to allow floating point.
	 * @param allowSign Pass in <code>true</code> to allow sign characters.
	 * @param allowGroup Pass in <code>true</code> to allow group characters.
	 * @param maxDigits The maximum number of digits (not necessarily characters) the field can
	 *            have.
	 */
	public NumberFilter(JTextField field, boolean allowDecimal, boolean allowSign, boolean allowGroup, int maxDigits) {
		mField = field;
		mAllowDecimal = allowDecimal;
		mAllowSign = allowSign;
		mAllowGroup = allowGroup;
		mMaxDigits = maxDigits;
		for (KeyListener listener : mField.getKeyListeners()) {
			if (listener instanceof NumberFilter) {
				mField.removeKeyListener(listener);
			}
		}
		mField.addKeyListener(this);
	}

	public void keyPressed(KeyEvent event) {
		// Not used.
	}

	public void keyReleased(KeyEvent event) {
		// Not used.
	}

	public void keyTyped(KeyEvent event) {
		char ch = event.getKeyChar();
		if (ch != '\n' && ch != '\r' && ch != '\t' && ch != '\b' && ch != KeyEvent.VK_DELETE) {
			if (mAllowGroup && ch == GROUP_CHAR || ch >= '0' && ch <= '9' || mAllowSign && (ch == '-' || ch == '+') || mAllowDecimal && ch == DECIMAL_CHAR || mIsHeightFilter && (ch == '\'' || ch == '"' || ch == ' ')) {
				StringBuilder buffer = new StringBuilder(mField.getText());
				int start = mField.getSelectionStart();
				int end = mField.getSelectionEnd();

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
						filter(event);
						return;
					}
				}
				if (ch == GROUP_CHAR || ch >= '0' && ch <= '9') {
					if (mAllowSign && start == 0 && buffer.length() > 0 && (buffer.charAt(0) == '-' || buffer.charAt(0) == '+')) {
						filter(event);
					}
				} else if (ch == '-' || ch == '+') {
					if (start != 0) {
						filter(event);
					}
				} else if (ch == DECIMAL_CHAR) {
					if (buffer.indexOf("" + DECIMAL_CHAR) != -1 || mAllowSign && start == 0 && buffer.length() > 0 && (buffer.charAt(0) == '-' || buffer.charAt(0) == '+')) { //$NON-NLS-1$
						filter(event);
					}
				} else if (ch == '\'') {
					if (buffer.indexOf("'") != -1) { //$NON-NLS-1$
						filter(event);
					}
				} else if (ch == '"') {
					if (buffer.indexOf("\"") != -1) { //$NON-NLS-1$
						filter(event);
					}
				}
			} else {
				filter(event);
			}
		}
	}

	private void filter(KeyEvent event) {
		System.out.println(event);
		Toolkit.getDefaultToolkit().beep();
		event.consume();
	}
}
