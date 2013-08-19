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

import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKDebug;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileReader;
import java.io.PrintStream;
import java.util.ArrayList;
import java.util.StringTokenizer;

/** General text utilities. */
public class TKTextUtility {
	private static final String	SPACE		= " ";		//$NON-NLS-1$
	private static final String	NEWLINE		= "\n";	//$NON-NLS-1$
	private static final char	ELLIPSIS	= '\u2026';

	/**
	 * If the text doesn't fit in the specified character count, it will be shortened and an ellipse
	 * ("...") will be added.
	 * 
	 * @param text The text to work on.
	 * @param count The maximum character count.
	 * @param truncationPolicy One of {@link TKAlignment#LEFT}, {@link TKAlignment#CENTER}, or
	 *            {@link TKAlignment#RIGHT}.
	 * @return The adjusted text.
	 */
	public static final String truncateIfNecessary(String text, int count, int truncationPolicy) {
		int tCount = text.length();

		count = tCount - count;
		if (count > 0) {
			count++; // Count is now the amount to remove from the string
			if (truncationPolicy == TKAlignment.LEFT) {
				return ELLIPSIS + text.substring(count);
			}
			if (truncationPolicy == TKAlignment.CENTER) {
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
	 * Extracts lines of text from the specified text file.
	 * 
	 * @param file The text file to extract lines from.
	 * @param tabWidth The width (in spaces) of a tab character. Pass in <code>0</code> or less to
	 *            leave tab characters in place.
	 * @return The lines of text.
	 */
	public static final ArrayList<String> extractLines(File file, int tabWidth) {
		ArrayList<String> lines = new ArrayList<String>();

		try {
			BufferedReader reader = new BufferedReader(new FileReader(file));
			String line = reader.readLine();

			while (line != null) {
				lines.add(tabWidth < 1 ? line : expandTabs(line, 0, tabWidth));
				line = reader.readLine();
			}
			reader.close();
		} catch (Exception exception) {
			assert false : TKDebug.throwableToString(exception);
		}

		return lines;
	}

	/**
	 * Expands the tabs in a string into spaces.
	 * 
	 * @param text The text to expand.
	 * @param startingPosition The column position the first character of this string will be in.
	 * @param tabWidth The number of spaces a tab represents.
	 * @return The tab-expanded text.
	 */
	public static final String expandTabs(String text, int startingPosition, int tabWidth) {
		return expandTabs(text, startingPosition, tabWidth, null);
	}

	/**
	 * Expands the tabs in a string into spaces.
	 * 
	 * @param text The text to expand.
	 * @param startingPosition The column position the first character of this string will be in.
	 * @param tabWidth The number of spaces a tab represents.
	 * @param indexes A buffer that will be filled in with the new index positions of the old
	 *            characters.
	 * @return The tab-expanded text.
	 */
	public static final String expandTabs(String text, int startingPosition, int tabWidth, int[] indexes) {
		int length = text.length();
		int column = startingPosition;
		StringBuilder buffer = new StringBuilder(length * 2);

		for (int i = 0; i < length; i++) {
			char ch = text.charAt(i);

			if (indexes != null) {
				indexes[i] = column - startingPosition;
			}
			if (ch == '\t') {
				int spaces = tabWidth - column % tabWidth;

				while (--spaces >= 0) {
					buffer.append(' ');
					column++;
				}
			} else {
				buffer.append(ch);
				if (ch != '\n') {
					column++;
				} else {
					column = 0;
				}
			}
		}
		return buffer.toString();
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

	/**
	 * Prints formatted columnized string data to the specified stream.
	 * 
	 * @param stream The print stream.
	 * @param alignment An array with integer values specifying the alignment for each column. A
	 *            negative value represents left alignment, a positive value represents right
	 *            alignment, zero represents centered alignment, and -2 represents left alignment
	 *            with no preceeding gap.
	 * @param data The data to print.
	 * @param gapSize The number of spaces to place between columns.
	 * @param centerFirst Pass in <code>true</code> to center each column's first row. Typically
	 *            used for headers.
	 */
	public static void dumpFormattedColumns(PrintStream stream, int[] alignment, String[][] data, int gapSize, boolean centerFirst) {
		String gap = makeFiller(gapSize, ' ');
		int[] widths = new int[alignment.length];
		int x;
		int y;
		int width;

		for (x = 0; x < alignment.length; x++) {
			widths[x] = 0;
		}

		for (y = 0; y < data.length; y++) {
			for (x = 0; x < data[y].length; x++) {
				width = data[y][x].length();
				if (width > widths[x]) {
					widths[x] = width;
				}
			}
		}

		for (y = 0; y < data.length; y++) {
			for (x = 0; x < data[y].length; x++) {
				String one = data[y][x];
				int align = centerFirst && y == 0 ? 0 : alignment[x];

				if (align != -2 && x != 0) {
					stream.print(gap);
				}

				if (one.equals("-")) { //$NON-NLS-1$
					one = makeFiller(widths[x], '-');
				}
				width = one.length();

				if (align < 0) { // Left
					stream.print(one + makeFiller(widths[x] - width, ' '));
				} else if (align > 0) { // Right
					stream.print(makeFiller(widths[x] - width, ' ') + one);
				} else { // Centered
					int left = (widths[x] - width) / 2;
					int right = widths[x] - (width + left);

					stream.print(makeFiller(left, ' ') + one + makeFiller(right, ' '));
				}
			}
			stream.println();
		}
	}

	/**
	 * Prints columnized string data to the specified stream.
	 * 
	 * @param stream The print stream.
	 * @param data The data to print.
	 */
	public static void dumpColumns(PrintStream stream, String[][] data) {
		for (String[] element : data) {
			for (int x = 0; x < element.length; x++) {
				stream.print(element[x]);
				if (x != element.length - 1) {
					stream.print('\t');
				}
			}
			stream.println();
		}
	}
}
