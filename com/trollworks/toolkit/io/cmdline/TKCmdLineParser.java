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

package com.trollworks.toolkit.io.cmdline;

import com.trollworks.toolkit.utility.TKPlatform;

import java.util.ArrayList;

/** Provides parsing of a string to generate an argument list. */
public class TKCmdLineParser {
	/**
	 * @param line An unparsed command-line.
	 * @return The command-line, separated into separate arguments.
	 */
	public static ArrayList<String> parseIntoList(String line) {
		ArrayList<String> args = new ArrayList<String>();
		StringBuilder buffer = new StringBuilder();
		int size = line.length();
		boolean inEscape = false;
		boolean inDoubleQuote = false;
		boolean inSingleQuote = false;
		boolean canEscape = !TKPlatform.isWindows();

		for (int i = 0; i < size; i++) {
			char ch = line.charAt(i);

			if (inEscape) {
				inEscape = false;
			} else if (canEscape && ch == '\\') {
				inEscape = true;
			} else if (inDoubleQuote) {
				if (ch == '"') {
					inDoubleQuote = false;
				} else {
					buffer.append(ch);
				}
			} else if (inSingleQuote) {
				if (ch == '\'') {
					inSingleQuote = false;
				} else {
					buffer.append(ch);
				}
			} else if (ch == '"') {
				inDoubleQuote = true;
			} else if (ch == '\'') {
				inSingleQuote = true;
			} else if (ch == ' ' || ch == '\t') {
				if (buffer.length() > 0) {
					args.add(buffer.toString());
					buffer.setLength(0);
				}
			} else {
				buffer.append(ch);
			}
		}

		if (inEscape) {
			buffer.append('\\');
		} else if (inDoubleQuote) {
			buffer.insert(0, '"');
		} else if (inSingleQuote) {
			buffer.insert(0, '\'');
		}
		if (buffer.length() > 0) {
			args.add(buffer.toString());
		}

		return args;
	}

	/**
	 * @param line An unparsed command-line.
	 * @return The command-line, separated into separate arguments.
	 */
	public static String[] parseIntoArray(String line) {
		return parseIntoList(line).toArray(new String[0]);
	}
}
