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

import com.trollworks.toolkit.utility.TKNumberUtils;

import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.Reader;
import java.nio.charset.Charset;
import java.nio.charset.CharsetDecoder;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.HashMap;

/** A very simple XML reader with very low memory overhead. */
public class TKXMLReader {
	/** Debug option: whether to output skipped tags to standard out. */
	public static boolean			SHOW_SKIPPED_TAGS	= TKNumberUtils.getBoolean(System.getProperty("SHOW_SKIPPED_TAGS", "false"));	//$NON-NLS-1$ //$NON-NLS-2$
	private static final String		UNEXPECTED_EOF		= "Unexpected EOF";															//$NON-NLS-1$
	private static final String		COLON				= ":";																			//$NON-NLS-1$
	private HashMap<String, String>	mEntityMap			= new HashMap<String, String>();
	private HashMap<String, String>	mAttributeMap		= new HashMap<String, String>();
	private ArrayList<String>		mStack				= new ArrayList<String>();
	private char[]					mBuffer				= new char[32768];
	private char[]					mTextBuffer			= new char[128];
	private TKXMLNodeType			mType				= TKXMLNodeType.START_DOCUMENT;
	private int						mLine				= 1;
	private int						mColumn				= 1;
	private Reader					mReader;
	private int						mPos;
	private int						mCount;
	private boolean					mEOF;
	private int						mPeek0;
	private int						mPeek1;
	private int						mTextPos;
	private String					mText;
	private boolean					mIsWhitespace;
	private String					mName;
	private boolean					mIsEmptyElementTag;

	/**
	 * Creates a new {@link TKXMLReader}.
	 * 
	 * @param stream The underlying {@link InputStream} to use.
	 * @throws IOException
	 */
	public TKXMLReader(InputStream stream) throws IOException {
		this(new InputStreamReader(stream));
	}

	/**
	 * Creates a new {@link TKXMLReader}.
	 * 
	 * @param stream The underlying {@link InputStream} to use.
	 * @param charset The {@link Charset} to use.
	 * @throws IOException
	 */
	public TKXMLReader(InputStream stream, Charset charset) throws IOException {
		this(new InputStreamReader(stream, charset));
	}

	/**
	 * Creates a new {@link TKXMLReader}.
	 * 
	 * @param stream The underlying {@link InputStream} to use.
	 * @param decoder The {@link CharsetDecoder} to use.
	 * @throws IOException
	 */
	public TKXMLReader(InputStream stream, CharsetDecoder decoder) throws IOException {
		this(new InputStreamReader(stream, decoder));
	}

	/**
	 * Creates a new {@link TKXMLReader}.
	 * 
	 * @param stream The underlying {@link InputStream} to use.
	 * @param charsetName The name of the {@link Charset} to use.
	 * @throws IOException
	 */
	public TKXMLReader(InputStream stream, String charsetName) throws IOException {
		this(new InputStreamReader(stream, charsetName));
	}

	/**
	 * Creates a new {@link TKXMLReader}.
	 * 
	 * @param reader The underlying {@link Reader} to use. It is assumed that the reader has been
	 *            created with a {@link Charset} appropriate for the data being read.
	 * @throws IOException
	 */
	public TKXMLReader(Reader reader) throws IOException {
		mReader = reader;
		mPeek0 = reader.read();
		mPeek1 = reader.read();
		mEOF = mPeek0 == -1;
		defineCharacterEntity("amp", "&"); //$NON-NLS-1$ //$NON-NLS-2$
		defineCharacterEntity("apos", "'"); //$NON-NLS-1$ //$NON-NLS-2$
		defineCharacterEntity("gt", ">"); //$NON-NLS-1$ //$NON-NLS-2$
		defineCharacterEntity("lt", "<"); //$NON-NLS-1$ //$NON-NLS-2$
		defineCharacterEntity("quot", "\""); //$NON-NLS-1$ //$NON-NLS-2$
	}

	/**
	 * Closes the underlying {@link Reader}.
	 * 
	 * @throws IOException
	 */
	public void close() throws IOException {
		mReader.close();
	}

	/** @return A marker for determining if you've come to the end of a specific tag. */
	public String getMarker() {
		switch (mType) {
			case START_TAG:
				return getDepth() - 1 + COLON + getName();
			case END_TAG:
				return getDepth() + COLON + getName();
			default:
				return getDepth() + COLON + mStack.get(mStack.size() - 1);
		}
	}

	/**
	 * Allows you to determine if you've reached the end of a tag you've previously marked. An
	 * example of use:
	 * 
	 * <pre>
	 *    String marker = reader.getMarker();
	 *   
	 *    do {
	 *        XMLNodeType type = reader.next();
	 *   
	 *        ... process the info between the start and close tag here ...
	 *    } while (reader.withinMarker(marker));
	 * </pre>
	 * 
	 * @param marker The marker, from a previous call to {@link #getMarker()}.
	 * @return Whether the current position is still within the marked range. If it is not, then
	 *         {@link #next()} will be called.
	 * @throws IOException
	 */
	public boolean withinMarker(String marker) throws IOException {
		if (mType == TKXMLNodeType.END_TAG) {
			if (marker.equals(getMarker())) {
				next();
				return false;
			}
		} else if (mType == TKXMLNodeType.END_DOCUMENT) {
			fail("expected: " + TKXMLNodeType.END_TAG.name() + "/" + marker.substring(marker.indexOf(':') + 1)); //$NON-NLS-1$ //$NON-NLS-2$
		}
		return true;
	}

	/**
	 * Requires the current type to be the specified type and optionally requires the name to match
	 * as well. If the current type is {@link TKXMLNodeType#TEXT} and {@link #isWhitespace()} is
	 * <code>true</code> and required type is not {@link TKXMLNodeType#TEXT}, then
	 * {@link #next()} is called prior to the check.
	 * 
	 * @param type The type to require.
	 * @param name The name to require. Pass in <code>null</code> to allow any name.
	 * @throws IOException
	 */
	public void require(TKXMLNodeType type, String name) throws IOException {
		if (mType == TKXMLNodeType.TEXT && type != TKXMLNodeType.TEXT && isWhitespace()) {
			next();
		}
		if (type != mType || name != null && !name.equals(getName())) {
			fail("expected: " + type.name() + "/" + name); //$NON-NLS-1$ //$NON-NLS-2$
		}
	}

	/**
	 * Skips over the specified tag, which must be the current tag.
	 * 
	 * @param name The name of the tag to skip.
	 * @throws IOException
	 */
	public void skipTag(String name) throws IOException {
		String marker;

		if (SHOW_SKIPPED_TAGS) {
			System.out.println("Skipping tag: " + name); //$NON-NLS-1$
		}
		require(TKXMLNodeType.START_TAG, name);
		marker = getMarker();
		do {
			next();
		} while (withinMarker(marker));
	}

	/**
	 * Reads the text of a simple tag (i.e. one without children). If the tag does have children,
	 * they will be skipped. The current position will be moved to right after the closing tag.
	 * 
	 * @return The text at the current position.
	 * @throws IOException
	 */
	public String readText() throws IOException {
		StringBuilder builder = new StringBuilder();
		String marker = getMarker();

		if (mType == TKXMLNodeType.START_TAG) {
			next();
		}
		do {
			if (mType == TKXMLNodeType.TEXT) {
				if (builder.length() > 0) {
					builder.append(' ');
				}
				builder.append(getText());
				next();
			} else if (mType == TKXMLNodeType.START_TAG) {
				skipTag(getName());
			}
		} while (withinMarker(marker));
		return builder.toString();
	}

	/**
	 * @return The boolean value of a call to {@link #readText()}.
	 * @throws IOException
	 */
	public boolean readBoolean() throws IOException {
		return TKNumberUtils.getBoolean(readText());
	}

	/**
	 * @param defValue The default value to return if the text cannot be converted to a short.
	 * @return The short value of a call to {@link #readText()}.
	 * @throws IOException
	 */
	public short readShort(short defValue) throws IOException {
		return TKNumberUtils.getNonLocalizedShort(readText(), defValue);
	}

	/**
	 * @param defValue The default value to return if the text cannot be converted to an integer.
	 * @return The integer value of a call to {@link #readText()}.
	 * @throws IOException
	 */
	public int readInteger(int defValue) throws IOException {
		return TKNumberUtils.getNonLocalizedInteger(readText(), defValue);
	}

	/**
	 * @param defValue The default value to return if the text cannot be converted to a long.
	 * @return The long value of a call to {@link #readText()}.
	 * @throws IOException
	 */
	public long readLong(long defValue) throws IOException {
		return TKNumberUtils.getNonLocalizedLong(readText(), defValue);
	}

	/**
	 * @param defValue The default value to return if the text cannot be converted to a float.
	 * @return The float value of a call to {@link #readText()}.
	 * @throws IOException
	 */
	public float readFloat(float defValue) throws IOException {
		return TKNumberUtils.getNonLocalizedFloat(readText(), defValue);
	}

	/**
	 * @param defValue The default value to return if the text cannot be converted to a double.
	 * @return The double value of a call to {@link #readText()}.
	 * @throws IOException
	 */
	public double readDouble(double defValue) throws IOException {
		return TKNumberUtils.getNonLocalizedDouble(readText(), defValue);
	}

	/** @return The {@link Calendar} representing the date and time read. */
	public Calendar readDateTime() {
		Calendar calendar = Calendar.getInstance();

		calendar.set(getAttributeAsInteger(TKXMLWriter.YEAR, 1970), getAttributeAsInteger(TKXMLWriter.MONTH, 1), getAttributeAsInteger(TKXMLWriter.DAY, 1), getAttributeAsInteger(TKXMLWriter.HOUR, 0), getAttributeAsInteger(TKXMLWriter.MINUTE, 0), getAttributeAsInteger(TKXMLWriter.SECOND, 0));
		return calendar;
	}

	private final int read() throws IOException {
		int result = mPeek0;

		mPeek0 = mPeek1;
		if (mPeek0 == -1) {
			mEOF = true;
			return result;
		} else if (result == '\n' || result == '\r') {
			mLine++;
			mColumn = 0;
			if (result == '\r' && mPeek0 == '\n') {
				mPeek0 = 0;
			}
		}
		mColumn++;

		if (mPos >= mCount) {
			mCount = mReader.read(mBuffer, 0, mBuffer.length);
			if (mCount <= 0) {
				mPeek1 = -1;
				return result;
			}
			mPos = 0;
		}

		mPeek1 = mBuffer[mPos++];
		return result;
	}

	private final void fail(String desc) throws IOException {
		throw new IOException(desc + " pos: " + getPositionDescription()); //$NON-NLS-1$
	}

	private final void push(int ch) {
		if (ch != 0) {
			if (mTextPos == mTextBuffer.length) {
				char[] bigger = new char[mTextPos * 4 / 3 + 4];

				System.arraycopy(mTextBuffer, 0, bigger, 0, mTextPos);
				mTextBuffer = bigger;
			}
			mTextBuffer[mTextPos++] = (char) ch;
		}
	}

	private final void read(char ch) throws IOException {
		if (read() != ch) {
			fail("expected: '" + ch + "'"); //$NON-NLS-1$ //$NON-NLS-2$
		}
	}

	private final void skip() throws IOException {
		while (!mEOF && mPeek0 <= ' ') {
			read();
		}
	}

	private final String pop(int pos) {
		String result = new String(mTextBuffer, pos, mTextPos - pos);

		mTextPos = pos;
		return result;
	}

	private final String readName() throws IOException {
		int pos = mTextPos;
		int ch = mPeek0;

		if ((ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') && ch != '_' && ch != ':') {
			fail("name expected"); //$NON-NLS-1$
		}

		do {
			push(read());
			ch = mPeek0;
		} while (ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '_' || ch == '-' || ch == ':' || ch == '.');

		return pop(pos);
	}

	private final void parseLegacy(boolean push) throws IOException {
		String req = ""; //$NON-NLS-1$
		int term;
		int ch;

		read(); // <
		ch = read();
		if (ch == '?') {
			term = '?';
		} else if (ch == '!') {
			if (mPeek0 == '-') {
				req = "--"; //$NON-NLS-1$
				term = '-';
			} else {
				req = "DOCTYPE"; //$NON-NLS-1$
				term = -1;
			}
		} else {
			if (ch != '[') {
				fail("can't reach: " + ch); //$NON-NLS-1$
			}
			req = "CDATA["; //$NON-NLS-1$
			term = ']';
		}

		for (int i = 0; i < req.length(); i++) {
			read(req.charAt(i));
		}

		if (term == -1) {
			parseDoctype();
		} else {
			while (true) {
				if (mEOF) {
					fail(UNEXPECTED_EOF);
				}

				ch = read();
				if (push) {
					push(ch);
				}

				if ((term == '?' || ch == term) && mPeek0 == term && mPeek1 == '>') {
					break;
				}
			}
			read();
			read();

			if (push && term != '?') {
				pop(mTextPos - 1);
			}
		}
	}

	private final void parseDoctype() throws IOException {
		int nesting = 1;

		while (true) {
			switch (read()) {
				case -1:
					fail(UNEXPECTED_EOF);
					break;
				case '<':
					nesting++;
					break;
				case '>':
					if (--nesting == 0) {
						return;
					}
					break;
			}
		}
	}

	private final void parseEndTag() throws IOException {
		int pos;

		read(); // '<'
		read(); // '/'
		mName = readName();
		pos = mStack.size() - 1;
		if (pos < 0) {
			fail("element stack empty"); //$NON-NLS-1$
		}
		if (mName.equals(mStack.get(pos))) {
			mStack.remove(pos);
		} else {
			fail("expected: " + mStack.get(pos)); //$NON-NLS-1$
		}
		skip();
		read('>');
	}

	private final TKXMLNodeType peekType() {
		switch (mPeek0) {
			case -1:
				return TKXMLNodeType.END_DOCUMENT;
			case '&':
				return TKXMLNodeType.ENTITY_REF;
			case '<':
				switch (mPeek1) {
					case '/':
						return TKXMLNodeType.END_TAG;
					case '[':
						return TKXMLNodeType.DATA;
					case '?':
					case '!':
						return TKXMLNodeType.OTHER;
					default:
						return TKXMLNodeType.START_TAG;
				}
			default:
				return TKXMLNodeType.TEXT;
		}
	}

	private final void parseStartTag() throws IOException {
		read(); // <
		mName = readName();
		mStack.add(mName);

		while (true) {
			String attrName;
			int ch;
			int pos;

			skip();
			ch = mPeek0;
			if (ch == '/') {
				mIsEmptyElementTag = true;
				read();
				skip();
				read('>');
				break;
			}
			if (ch == '>') {
				read();
				break;
			}
			if (ch == -1) {
				fail(UNEXPECTED_EOF);
			}

			attrName = readName();
			if (attrName.length() == 0) {
				fail("attribute name expected"); //$NON-NLS-1$
			}

			skip();
			read('=');
			skip();
			ch = read();
			if (ch != '\'' && ch != '"') {
				fail("<" + mName + ">: invalid delimiter: " + (char) ch); //$NON-NLS-1$ //$NON-NLS-2$
			}

			pos = mTextPos;
			pushText(ch);
			mAttributeMap.put(attrName, pop(pos));
			if (ch != ' ') {
				read(); // skip endquote
			}
		}
	}

	private final boolean pushEntity() throws IOException {
		boolean whitespace = true;
		int pos;
		String code;
		String result;

		read(); // &
		pos = mTextPos;
		while (!mEOF && mPeek0 != ';') {
			push(read());
		}

		code = pop(pos);
		read();
		if (code.length() > 0 && code.charAt(0) == '#') {
			int c = code.charAt(1) == 'x' ? Integer.parseInt(code.substring(2), 16) : Integer.parseInt(code.substring(1));

			push(c);
			return c <= ' ';
		}

		result = mEntityMap.get(code);
		if (result == null) {
			result = "&" + code + ";"; //$NON-NLS-1$ //$NON-NLS-2$
		}

		for (int i = 0; i < result.length(); i++) {
			char c = result.charAt(i);

			if (c > ' ') {
				whitespace = false;
			}
			push(c);
		}
		return whitespace;
	}

	private final boolean pushText(int delimiter) throws IOException {
		boolean whitespace = true;
		int next = mPeek0;

		while (!mEOF && next != delimiter) { // covers EOF, '<', '"'
			if (delimiter == ' ') {
				if (next <= ' ' || next == '>') {
					break;
				}
			}
			if (next == '&') {
				if (!pushEntity()) {
					whitespace = false;
				}
			} else {
				if (next > ' ') {
					whitespace = false;
				}
				push(read());
			}
			next = mPeek0;
		}
		return whitespace;
	}

	/**
	 * Define a character entity mapping.
	 * 
	 * @param entity The XML entity.
	 * @param value The value to substitute.
	 */
	public void defineCharacterEntity(String entity, String value) {
		mEntityMap.put(entity, value);
	}

	/** @return The current tag depth. */
	public int getDepth() {
		return mStack.size();
	}

	/** @return A description of the current parse position. */
	public String getPositionDescription() {
		StringBuilder buffer = new StringBuilder(mType.name());

		buffer.append(" @" + mLine + COLON + mColumn + ": "); //$NON-NLS-1$ //$NON-NLS-2$
		if (mType == TKXMLNodeType.START_TAG) {
			buffer.append('<');
			buffer.append(mName);
			buffer.append('>');
		} else if (mType == TKXMLNodeType.END_TAG) {
			buffer.append("</"); //$NON-NLS-1$
			buffer.append(mName);
			buffer.append('>');
		} else if (mIsWhitespace) {
			buffer.append("[whitespace]"); //$NON-NLS-1$
		} else {
			buffer.append(getText());
		}
		return buffer.toString();
	}

	/** @return The current line number. */
	public int getLineNumber() {
		return mLine;
	}

	/** @return The current column number. */
	public int getColumnNumber() {
		return mColumn;
	}

	/** @return Whether the current position is whitespace. */
	public boolean isWhitespace() {
		return mIsWhitespace;
	}

	/** @return The text at the current position. */
	public String getText() {
		if (mText == null) {
			mText = pop(0);
		}
		return mText;
	}

	/** @return The current name. */
	public String getName() {
		return mName;
	}

	/** @return Whether the current element tag is empty. */
	public boolean isEmptyElementTag() {
		return mIsEmptyElementTag;
	}

	/**
	 * @param name The name of the attribute to use.
	 * @return The value of the attribute.
	 */
	public String getAttribute(String name) {
		return mAttributeMap.get(name);
	}

	/**
	 * @param name The name of the attribute to check.
	 * @return Whether the attribute is present.
	 */
	public boolean hasAttribute(String name) {
		return mAttributeMap.get(name) != null;
	}

	/**
	 * @param name The name of the attribute to check.
	 * @return Whether the attribute is present and set to a 'true' value.
	 */
	public boolean isAttributeSet(String name) {
		return TKNumberUtils.getBoolean(mAttributeMap.get(name));
	}

	/**
	 * @param name The attribute name.
	 * @param defValue The default value to use if the attribute value can't be converted to a
	 *            short.
	 * @return The value of the tag.
	 */
	public short getAttributeAsShort(String name, short defValue) {
		return TKNumberUtils.getNonLocalizedShort(mAttributeMap.get(name), defValue);
	}

	/**
	 * @param name The attribute name.
	 * @param defValue The default value to use if the attribute value can't be converted to an
	 *            integer.
	 * @return The value of the tag.
	 */
	public int getAttributeAsInteger(String name, int defValue) {
		return TKNumberUtils.getNonLocalizedInteger(mAttributeMap.get(name), defValue);
	}

	/**
	 * @param name The attribute name.
	 * @param defValue The default value to use if the attribute value can't be converted to a long.
	 * @return The value of the tag.
	 */
	public long getAttributeAsLong(String name, long defValue) {
		return TKNumberUtils.getNonLocalizedLong(mAttributeMap.get(name), defValue);
	}

	/**
	 * @param name The attribute name.
	 * @param defValue The default value to use if the attribute value can't be converted to a
	 *            float.
	 * @return The value of the tag.
	 */
	public float getAttributeAsFloat(String name, float defValue) {
		return TKNumberUtils.getNonLocalizedFloat(mAttributeMap.get(name), defValue);
	}

	/**
	 * @param name The attribute name.
	 * @param defValue The default value to use if the attribute value can't be converted to a
	 *            double.
	 * @return The value of the tag.
	 */
	public double getAttributeAsDouble(String name, double defValue) {
		return TKNumberUtils.getNonLocalizedDouble(mAttributeMap.get(name), defValue);
	}

	/** @return The map of attributes. */
	public HashMap<String, String> getAttributes() {
		return mAttributeMap;
	}

	/** @return The type at the current position. */
	public TKXMLNodeType getType() {
		return mType;
	}

	/**
	 * Advances to the next position.
	 * 
	 * @return The type.
	 * @throws IOException
	 */
	public TKXMLNodeType next() throws IOException {
		if (mIsEmptyElementTag) {
			mType = TKXMLNodeType.END_TAG;
			mIsEmptyElementTag = false;
			mStack.remove(mStack.size() - 1);
		} else {
			int textOrdinal = TKXMLNodeType.TEXT.ordinal();

			mTextPos = 0;
			mIsWhitespace = true;
			do {
				mAttributeMap.clear();
				mName = null;
				mText = null;
				mType = peekType();
				switch (mType) {
					case ENTITY_REF:
						mIsWhitespace &= pushEntity();
						mType = TKXMLNodeType.TEXT;
						break;
					case START_TAG:
						parseStartTag();
						break;
					case END_TAG:
						parseEndTag();
						break;
					case END_DOCUMENT:
						break;
					case TEXT:
						mIsWhitespace &= pushText('<');
						break;
					case DATA:
						parseLegacy(true);
						mIsWhitespace = false;
						mType = TKXMLNodeType.TEXT;
						break;
					default:
						parseLegacy(false);
						break;
				}
			} while (mType.ordinal() > textOrdinal || mType == TKXMLNodeType.TEXT && peekType().ordinal() >= textOrdinal);

			mIsWhitespace &= mType == TKXMLNodeType.TEXT;
		}
		return mType;
	}
}
