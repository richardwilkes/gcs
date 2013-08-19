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

package com.trollworks.toolkit.io.xml;

import java.io.IOException;
import java.io.OutputStream;
import java.io.OutputStreamWriter;
import java.io.PrintWriter;
import java.util.Calendar;
import java.util.Date;

/** A {@link PrintWriter} that has been extended to provide common XML writing helper methods. */
public class TKXMLWriter extends PrintWriter {
	private static final String	END_TAG				= "/>";		//$NON-NLS-1$
	private static final String	ENTITY_CODE_PREFIX	= "&#";		//$NON-NLS-1$
	private static final String	AMPERSAND_ENTITY	= "&amp;";		//$NON-NLS-1$
	private static final String	LESS_THAN_ENTITY	= "&lt;";		//$NON-NLS-1$
	private static final String	GREATER_THAN_ENTITY	= "&gt;";		//$NON-NLS-1$
	private static final String	END_COMMENT			= " -->";		//$NON-NLS-1$
	/** The encoding used. */
	public static final String	ENCODING			= "US-ASCII";	//$NON-NLS-1$
	/** The 'year' attribute. */
	public static final String	YEAR				= "year";		//$NON-NLS-1$
	/** The 'month' attribute. */
	public static final String	MONTH				= "month";		//$NON-NLS-1$
	/** The 'day' attribute. */
	public static final String	DAY					= "day";		//$NON-NLS-1$
	/** The 'hour' attribute. */
	public static final String	HOUR				= "hour";		//$NON-NLS-1$
	/** The 'minute' attribute. */
	public static final String	MINUTE				= "minute";	//$NON-NLS-1$
	/** The 'second' attribute. */
	public static final String	SECOND				= "second";	//$NON-NLS-1$
	private int					mIndent;

	/**
	 * Creates a new XML writer.
	 * 
	 * @param stream The stream to write to.
	 * @throws IOException
	 */
	public TKXMLWriter(OutputStream stream) throws IOException {
		super(new OutputStreamWriter(stream, ENCODING));
	}

	/** Writes a standard XML header. */
	public void writeHeader() {
		print("<?xml version=\"1.0\" encoding=\""); //$NON-NLS-1$
		print(ENCODING);
		println("\" ?>"); //$NON-NLS-1$
	}

	/**
	 * Write out a simple XML comment with a trailing line feed.
	 * 
	 * @param comment The comment.
	 */
	public void writeComment(String comment) {
		writeIndentation();
		startComment();
		writeEncodedData(comment);
		finishCommentEOL();
	}

	/** Starts an XML comment. */
	public void startComment() {
		print("<!-- "); //$NON-NLS-1$
	}

	/** Finishes an XML comment. */
	public void finishComment() {
		print(END_COMMENT);
	}

	/** Finishes an XML comment and writes out a line feed. */
	public void finishCommentEOL() {
		println(END_COMMENT);
	}

	/**
	 * Writes out a piece of data in a way that encodes characters that have special meaning to XML.
	 * Does not encode for an attribute.
	 * 
	 * @param data The data to transform.
	 */
	public void writeEncodedData(String data) {
		if (data != null) {
			int length = data.length();

			for (int i = 0; i < length; i++) {
				char ch = data.charAt(i);

				if (ch == '<') {
					print(LESS_THAN_ENTITY);
				} else if (ch == '>') {
					print(GREATER_THAN_ENTITY);
				} else if (ch == '&') {
					print(AMPERSAND_ENTITY);
				} else if (ch == '\r' || ch == '\n') {
					println();
				} else if (ch == '\t') {
					print('\t');
				} else if (ch >= ' ' && ch <= '~') {
					print(ch);
				} else {
					print(ENTITY_CODE_PREFIX);
					print((int) ch);
					print(';');
				}
			}
		}
	}

	/**
	 * Writes out an attribute's value in a way that encodes characters that have special meaning to
	 * XML.
	 * 
	 * @param attribute The attribute value to transform.
	 */
	public void writeEncodedAttribute(String attribute) {
		int length = attribute.length();

		for (int i = 0; i < length; i++) {
			char ch = attribute.charAt(i);

			if (ch == '<') {
				print(LESS_THAN_ENTITY);
			} else if (ch == '>') {
				print(GREATER_THAN_ENTITY);
			} else if (ch == '&') {
				print(AMPERSAND_ENTITY);
			} else if (ch == '"') {
				print("&quot;"); //$NON-NLS-1$
			} else if (ch == '\'') {
				print("&apos;"); //$NON-NLS-1$
			} else if (ch >= ' ' && ch <= '~') {
				print(ch);
			} else {
				print(ENTITY_CODE_PREFIX);
				print((int) ch);
				print(';');
			}
		}
	}

	/**
	 * Writes an XML attribute.
	 * 
	 * @param name The name of the attribute.
	 * @param value The value of the attribute.
	 */
	public void writeAttribute(String name, boolean value) {
		writeAttribute(name, value ? "yes" : "no"); //$NON-NLS-1$ //$NON-NLS-2$
	}

	/**
	 * Writes an XML attribute.
	 * 
	 * @param name The name of the attribute.
	 * @param value The value of the attribute.
	 */
	public void writeAttribute(String name, int value) {
		writeAttribute(name, Integer.toString(value));
	}

	/**
	 * Writes an XML attribute.
	 * 
	 * @param name The name of the attribute.
	 * @param value The value of the attribute.
	 */
	public void writeAttribute(String name, long value) {
		writeAttribute(name, Long.toString(value));
	}

	/**
	 * Writes an XML attribute.
	 * 
	 * @param name The name of the attribute.
	 * @param value The value of the attribute.
	 */
	public void writeAttribute(String name, String value) {
		print(' ');
		print(name);
		print("=\""); //$NON-NLS-1$
		writeEncodedAttribute(value);
		print('"');
	}

	/**
	 * Write out an XML tag with a single attribute and no children, with a trailing line feed.
	 * 
	 * @param name The name of the tag.
	 * @param value The data to place between the tags.
	 * @param attribute The name of the attribute.
	 * @param attributeValue The value of the attribute.
	 */
	public void simpleTagWithAttribute(String name, String value, String attribute, String attributeValue) {
		startTag(name);
		writeAttribute(attribute, attributeValue);
		finishTag();
		writeEncodedData(value);
		endTagEOL(name, false);
	}

	/**
	 * Write out an XML tag with a single attribute and no children, with a trailing line feed.
	 * 
	 * @param name The name of the tag.
	 * @param value The data to place between the tags.
	 * @param attribute The name of the attribute.
	 * @param attributeValue The value of the attribute.
	 */
	public void simpleTagWithAttribute(String name, String value, String attribute, boolean attributeValue) {
		startTag(name);
		writeAttribute(attribute, attributeValue);
		finishTag();
		writeEncodedData(value);
		endTagEOL(name, false);
	}

	/**
	 * Write out an XML tag with a single attribute and no children, with a trailing line feed.
	 * 
	 * @param name The name of the tag.
	 * @param value The data to place between the tags.
	 * @param attribute The name of the attribute.
	 * @param attributeValue The value of the attribute.
	 */
	public void simpleTagWithAttribute(String name, String value, String attribute, int attributeValue) {
		startTag(name);
		writeAttribute(attribute, attributeValue);
		finishTag();
		writeEncodedData(value);
		endTagEOL(name, false);
	}

	/**
	 * Write out an XML tag with a single attribute and no children, with a trailing line feed.
	 * 
	 * @param name The name of the tag.
	 * @param value The data to place between the tags.
	 * @param attribute The name of the attribute.
	 * @param attributeValue The value of the attribute.
	 */
	public void simpleTagWithAttribute(String name, String value, String attribute, long attributeValue) {
		startTag(name);
		writeAttribute(attribute, attributeValue);
		finishTag();
		writeEncodedData(value);
		endTagEOL(name, false);
	}

	/**
	 * Write out an XML tag with a single attribute and no children, with a trailing line feed.
	 * 
	 * @param name The name of the tag.
	 * @param value The data to place between the tags.
	 * @param attribute The name of the attribute.
	 * @param attributeValue The value of the attribute.
	 */
	public void simpleTagWithAttribute(String name, double value, String attribute, String attributeValue) {
		startTag(name);
		writeAttribute(attribute, attributeValue);
		finishTag();
		writeEncodedData(Double.toString(value));
		endTagEOL(name, false);
	}

	/**
	 * Write out a simple XML tag (i.e. no children or attributes).
	 * 
	 * @param name The name of the tag.
	 * @param value The data to place between the tags.
	 */
	public void simpleTag(String name, boolean value) {
		simpleTag(name, Boolean.toString(value));
	}

	/**
	 * Write out a simple XML tag (i.e. no children or attributes).
	 * 
	 * @param name The name of the tag.
	 * @param value The data to place between the tags.
	 */
	public void simpleTag(String name, int value) {
		simpleTag(name, Integer.toString(value));
	}

	/**
	 * Write out a simple XML tag (i.e. no children or attributes).
	 * 
	 * @param name The name of the tag.
	 * @param value The data to place between the tags.
	 */
	public void simpleTag(String name, long value) {
		simpleTag(name, Long.toString(value));
	}

	/**
	 * Write out a simple XML tag (i.e. no children or attributes).
	 * 
	 * @param name The name of the tag.
	 * @param value The data to place between the tags.
	 */
	public void simpleTag(String name, double value) {
		simpleTag(name, Double.toString(value));
	}

	/**
	 * Write out a simple XML tag (i.e. no children or attributes).
	 * 
	 * @param name The name of the tag.
	 * @param value The data to place between the tags.
	 */
	public void simpleTag(String name, String value) {
		startSimpleTag(name);
		writeEncodedData(value);
		endTagEOL(name, false);
	}

	/**
	 * Write out a simple XML tag (i.e. no children or attributes).
	 * 
	 * @param name The name of the tag.
	 * @param value The data to place between the tags.
	 */
	public void simpleTag(String name, Object value) {
		if (value != null) {
			simpleTag(name, value.toString());
		}
	}

	/**
	 * Write out a simple XML start tag (i.e. no attributes).
	 * 
	 * @param name The name of the tag.
	 */
	public void startSimpleTag(String name) {
		startTag(name);
		finishTag();
	}

	/**
	 * Write out a simple XML start tag (i.e. no attributes) with a trailing line feed.
	 * 
	 * @param name The name of the tag.
	 */
	public void startSimpleTagEOL(String name) {
		startTag(name);
		finishTagEOL();
	}

	/**
	 * Write out an unterminated XML start tag.
	 * 
	 * @param name The name of the tag.
	 */
	public void startTag(String name) {
		writeIndentation();
		print('<');
		print(name);
		indent();
	}

	/**
	 * Write out a simple XML end tag.
	 * 
	 * @param name The name of the tag.
	 * @param indent Whether to indent before writing the end tag.
	 */
	public void endTagEOL(String name, boolean indent) {
		outdent();
		if (indent) {
			writeIndentation();
		}
		print("</"); //$NON-NLS-1$
		print(name);
		finishTagEOL();
	}

	/** Finish writing out an XML tag with a trailing line feed. */
	public void finishTagEOL() {
		println('>');
	}

	/** Finish writing out an XML tag. */
	public void finishTag() {
		print('>');
	}

	/** Finish writing out an empty XML tag with a trailing line feed. */
	public void finishEmptyTagEOL() {
		println(END_TAG);
		outdent();
	}

	/** Finish writing out an empty XML tag. */
	public void finishEmptyTag() {
		print(END_TAG);
		outdent();
	}

	/** Increments the current indentation level. */
	public void indent() {
		mIndent++;
	}

	/** Decrements the current indentation level. */
	public void outdent() {
		mIndent--;
	}

	/** Writes the current indentation. */
	public void writeIndentation() {
		for (int i = 0; i < mIndent; i++) {
			print('\t');
		}
	}

	/**
	 * Write out a date and time XML tag.
	 * 
	 * @param name The name of the tag.
	 * @param dateInMillis The date to output.
	 * @param includeDate Whether to output the date fields.
	 * @param includeTime Whether to output the time fields.
	 * @param includeSeconds Whether to output the seconds field. Only relevant if
	 *            <code>includeTime</code> was <code>true</code>.
	 */
	public void writeDateTimeTag(String name, long dateInMillis, boolean includeDate, boolean includeTime, boolean includeSeconds) {
		Calendar calendar = Calendar.getInstance();

		calendar.setTimeInMillis(dateInMillis);
		writeDateTimeTag(name, calendar, includeDate, includeTime, includeSeconds);
	}

	/**
	 * Write out a date and time XML tag.
	 * 
	 * @param name The name of the tag.
	 * @param date The date to output.
	 * @param includeDate Whether to output the date fields.
	 * @param includeTime Whether to output the time fields.
	 * @param includeSeconds Whether to output the seconds field. Only relevant if
	 *            <code>includeTime</code> was <code>true</code>.
	 */
	public void writeDateTimeTag(String name, Date date, boolean includeDate, boolean includeTime, boolean includeSeconds) {
		Calendar calendar = Calendar.getInstance();

		calendar.setTime(date);
		writeDateTimeTag(name, calendar, includeDate, includeTime, includeSeconds);
	}

	/**
	 * Write out a date and time XML tag.
	 * 
	 * @param name The name of the tag.
	 * @param calendar The calendar to output.
	 * @param includeDate Whether to output the date fields.
	 * @param includeTime Whether to output the time fields.
	 * @param includeSeconds Whether to output the seconds field. Only relevant if
	 *            <code>includeTime</code> was <code>true</code>.
	 */
	public void writeDateTimeTag(String name, Calendar calendar, boolean includeDate, boolean includeTime, boolean includeSeconds) {
		startTag(name);
		if (includeDate) {
			writeAttribute(YEAR, calendar.get(Calendar.YEAR));
			writeAttribute(MONTH, calendar.get(Calendar.MONTH) + 1);
			writeAttribute(DAY, calendar.get(Calendar.DAY_OF_MONTH));
		}
		if (includeTime) {
			writeAttribute(HOUR, calendar.get(Calendar.HOUR_OF_DAY));
			writeAttribute(MINUTE, calendar.get(Calendar.MINUTE));
			if (includeSeconds) {
				writeAttribute(SECOND, calendar.get(Calendar.SECOND));
			}
		}
		finishEmptyTagEOL();
	}

	/**
	 * Creates a new string in a way that encodes characters that have special meaning to XML. Does
	 * not encode for an attribute.
	 * 
	 * @param data The data to transform.
	 * @return An XML-encoded string.
	 */
	public static String encodeData(String data) {
		StringBuilder buffer = new StringBuilder();
		int length = data.length();

		for (int i = 0; i < length; i++) {
			char ch = data.charAt(i);

			if (ch == '<') {
				buffer.append(LESS_THAN_ENTITY);
			} else if (ch == '>') {
				buffer.append(GREATER_THAN_ENTITY);
			} else if (ch == '&') {
				buffer.append(AMPERSAND_ENTITY);
			} else if (ch >= ' ' && ch <= '~') {
				buffer.append(ch);
			} else {
				buffer.append(ENTITY_CODE_PREFIX);
				buffer.append((int) ch);
				buffer.append(';');
			}
		}
		return buffer.toString();
	}
}
