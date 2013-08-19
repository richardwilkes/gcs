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

package com.trollworks.gcs.utility.text;

import java.util.ArrayList;
import java.util.StringTokenizer;

import javax.swing.SwingConstants;

/** General text utilities. */
public class TextUtility {
	private static final String	SPACE		= " ";		//$NON-NLS-1$
	private static final String	NEWLINE		= "\n";	//$NON-NLS-1$
	private static final char	ELLIPSIS	= '\u2026';

	/**
	 * If the text doesn't fit in the specified character count, it will be shortened and an ellipse
	 * ("...") will be added.
	 * 
	 * @param text The text to work on.
	 * @param count The maximum character count.
	 * @param truncationPolicy One of {@link SwingConstants#LEFT}, {@link SwingConstants#CENTER},
	 *            or {@link SwingConstants#RIGHT}.
	 * @return The adjusted text.
	 */
	public static final String truncateIfNecessary(String text, int count, int truncationPolicy) {
		int tCount = text.length();

		count = tCount - count;
		if (count > 0) {
			count++; // Count is now the amount to remove from the string
			if (truncationPolicy == SwingConstants.LEFT) {
				return ELLIPSIS + text.substring(count);
			}
			if (truncationPolicy == SwingConstants.CENTER) {
				int remaining = tCount - count;
				int left = remaining / 2;
				int right = remaining - left;
				StringBuilder buffer = new StringBuilder(remaining + 1);

				if (left > 0) {
					buffer.append(text.substring(0, left));
				}
				buffer.append(ELLIPSIS);
				if (right > 0) {
					buffer.append(text.substring(tCount - right));
				}
				return buffer.toString();
			}
			return text.substring(0, tCount - count) + ELLIPSIS;
		}
		return text;
	}

	/**
	 * Convert text from other line ending formats into our internal format.
	 * 
	 * @param data The text to convert.
	 * @return The converted text.
	 */
	public static final String standardizeLineEndings(String data) {
		return standardizeLineEndings(data, NEWLINE);
	}

	/**
	 * Convert text from other line ending formats into a specific format.
	 * 
	 * @param data The text to convert.
	 * @param lineEnding The desired line ending.
	 * @return The converted text.
	 */
	public static final String standardizeLineEndings(String data, String lineEnding) {
		int length = data.length();
		StringBuilder buffer = new StringBuilder(length);
		char ignoreCh = 0;

		for (int i = 0; i < length; i++) {
			char ch = data.charAt(i);

			if (ch == ignoreCh) {
				ignoreCh = 0;
			} else if (ch == '\r') {
				ignoreCh = '\n';
				buffer.append(lineEnding);
			} else if (ch == '\n') {
				ignoreCh = '\r';
				buffer.append(lineEnding);
			} else {
				ignoreCh = 0;
				buffer.append(ch);
			}
		}

		return buffer.toString();
	}

	/**
	 * Extracts lines of text from the specified data.
	 * 
	 * @param data The text to extract lines from.
	 * @param tabWidth The width (in spaces) of a tab character. Pass in <code>0</code> or less to
	 *            leave tab characters in place.
	 * @return The lines of text.
	 */
	public static final ArrayList<String> extractLines(String data, int tabWidth) {
		int length = data.length();
		StringBuilder buffer = new StringBuilder(length);
		char ignoreCh = 0;
		ArrayList<String> lines = new ArrayList<String>();
		int column = 0;

		for (int i = 0; i < length; i++) {
			char ch = data.charAt(i);

			if (ch == ignoreCh) {
				ignoreCh = 0;
			} else if (ch == '\r') {
				ignoreCh = '\n';
				column = 0;
				lines.add(buffer.toString());
				buffer.setLength(0);
			} else if (ch == '\n') {
				ignoreCh = '\r';
				column = 0;
				lines.add(buffer.toString());
				buffer.setLength(0);
			} else if (ch == '\t' && tabWidth > 0) {
				int spaces = tabWidth - column % tabWidth;

				ignoreCh = 0;
				while (--spaces >= 0) {
					buffer.append(' ');
					column++;
				}
			} else {
				ignoreCh = 0;
				column++;
				buffer.append(ch);
			}
		}
		if (buffer.length() > 0) {
			lines.add(buffer.toString());
		}

		return lines;
	}

	/**
	 * @param amt The size of the string to create.
	 * @param filler The character to fill it with.
	 * @return A string filled with a specific character.
	 */
	public static String makeFiller(int amt, char filler) {
		StringBuilder buffer = new StringBuilder(amt);

		for (int i = 0; i < amt; i++) {
			buffer.append(filler);
		}
		return buffer.toString();
	}

	/**
	 * Creates a "note" whose second and subsequent lines are indented by the amount of the marker,
	 * which is prepended to the first line.
	 * 
	 * @param marker The prefix to use on the first line.
	 * @param note The text of the note.
	 * @return The formatted note.
	 */
	public static String makeNote(String marker, String note) {
		StringBuilder buffer = new StringBuilder(note.length() * 2);
		String indent = makeFiller(marker.length() + 1, ' ');
		StringTokenizer tokenizer = new StringTokenizer(note, NEWLINE);

		if (tokenizer.hasMoreTokens()) {
			buffer.append(marker);
			buffer.append(SPACE);
			buffer.append(tokenizer.nextToken());
			buffer.append(NEWLINE);

			while (tokenizer.hasMoreTokens()) {
				buffer.append(indent);
				buffer.append(tokenizer.nextToken());
				buffer.append(NEWLINE);
			}
		}

		return buffer.toString();
	}

	/**
	 * @param text The text to wrap.
	 * @param charCount The maximum character width to allow.
	 * @return A new, wrapped version of the text.
	 */
	public static String wrapToCharacterCount(String text, int charCount) {
		StringBuilder buffer = new StringBuilder(text.length() * 2);
		StringBuilder lineBuffer = new StringBuilder(charCount + 1);
		StringTokenizer tokenizer = new StringTokenizer(text + NEWLINE, NEWLINE, true);

		while (tokenizer.hasMoreTokens()) {
			String token = tokenizer.nextToken();

			if (token.equals(NEWLINE)) {
				buffer.append(token);
			} else {
				StringTokenizer tokenizer2 = new StringTokenizer(token, " \t", true); //$NON-NLS-1$
				int length = 0;

				lineBuffer.setLength(0);
				while (tokenizer2.hasMoreTokens()) {
					String token2 = tokenizer2.nextToken();
					int tokenLength = token2.length();

					if (length == 0 && token2.equals(SPACE)) {
						continue;
					}
					if (length == 0 || length + tokenLength <= charCount) {
						lineBuffer.append(token2);
						length += tokenLength;
					} else {
						buffer.append(lineBuffer);
						buffer.append(NEWLINE);
						lineBuffer.setLength(0);
						if (!token2.equals(SPACE)) {
							lineBuffer.append(token2);
							length = tokenLength;
						} else {
							length = 0;
						}
					}
				}
				if (length > 0) {
					buffer.append(lineBuffer);
				}
			}
		}
		buffer.setLength(buffer.length() - 1);
		return buffer.toString();
	}
}
