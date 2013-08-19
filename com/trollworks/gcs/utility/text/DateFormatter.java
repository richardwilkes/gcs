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

package com.trollworks.gcs.utility.text;

import java.text.DateFormat;
import java.text.ParseException;
import java.util.Date;

import javax.swing.JFormattedTextField;

/** Provides date field conversion. */
public class DateFormatter extends JFormattedTextField.AbstractFormatter {
	private int	mType;

	/**
	 * Creates a new {@link DateFormatter}.
	 * 
	 * @param type The type of date format to use, one of {@link DateFormat#SHORT},
	 *            {@link DateFormat#MEDIUM}, {@link DateFormat#LONG}, or {@link DateFormat#FULL}.
	 */
	public DateFormatter(int type) {
		mType = type;
	}

	@Override public Object stringToValue(String text) throws ParseException {
		return new Long(NumberUtils.getDate(text));
	}

	@Override public String valueToString(Object value) throws ParseException {
		Date date = new Date(((Long) value).longValue());
		return DateFormat.getDateInstance(mType).format(date);
	}
}
