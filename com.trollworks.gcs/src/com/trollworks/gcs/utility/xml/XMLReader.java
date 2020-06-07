/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.utility.xml;

import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;

import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.Reader;
import java.nio.charset.Charset;
import java.nio.charset.CharsetDecoder;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.HashMap;

/** A very simple XML reader with very low memory overhead. */
public class XMLReader implements AutoCloseable {
    /** Debug option: whether to output skipped tags to standard out. */
    public static        boolean                 SHOW_SKIPPED_TAGS = Numbers.extractBoolean(System.getProperty("SHOW_SKIPPED_TAGS", "false"));
    private static final String                  UNEXPECTED_EOF    = "Unexpected EOF";
    private              HashMap<String, String> mEntityMap        = new HashMap<>();
    private              HashMap<String, String> mAttributeMap     = new HashMap<>();
    private              ArrayList<String>       mStack            = new ArrayList<>();
    private              char[]                  mBuffer           = new char[32768];
    private              char[]                  mTextBuffer       = new char[128];
    private              XMLNodeType             mType             = XMLNodeType.START_DOCUMENT;
    private              int                     mLine             = 1;
    private              int                     mColumn           = 1;
    private              Reader                  mReader;
    private              int                     mPos;
    private              int                     mCount;
    private              boolean                 mEOF;
    private              int                     mPeek0;
    private              int                     mPeek1;
    private              int                     mTextPos;
    private              String                  mText;
    private              boolean                 mIsWhitespace;
    private              String                  mName;
    private              boolean                 mIsEmptyElementTag;

    /**
     * Creates a new {@link XMLReader}.
     *
     * @param stream The underlying {@link InputStream} to use.
     */
    public XMLReader(InputStream stream) throws IOException {
        this(new InputStreamReader(stream, StandardCharsets.UTF_8));
    }

    /**
     * Creates a new {@link XMLReader}.
     *
     * @param stream  The underlying {@link InputStream} to use.
     * @param charset The {@link Charset} to use.
     */
    public XMLReader(InputStream stream, Charset charset) throws IOException {
        this(new InputStreamReader(stream, charset));
    }

    /**
     * Creates a new {@link XMLReader}.
     *
     * @param stream  The underlying {@link InputStream} to use.
     * @param decoder The {@link CharsetDecoder} to use.
     */
    public XMLReader(InputStream stream, CharsetDecoder decoder) throws IOException {
        this(new InputStreamReader(stream, decoder));
    }

    /**
     * Creates a new {@link XMLReader}.
     *
     * @param stream      The underlying {@link InputStream} to use.
     * @param charsetName The name of the {@link Charset} to use.
     */
    public XMLReader(InputStream stream, String charsetName) throws IOException {
        this(new InputStreamReader(stream, charsetName));
    }

    /**
     * Creates a new {@link XMLReader}.
     *
     * @param reader The underlying {@link Reader} to use. It is assumed that the reader has been
     *               created with a {@link Charset} appropriate for the data being read.
     */
    public XMLReader(Reader reader) throws IOException {
        mReader = reader;
        mPeek0 = reader.read();
        mPeek1 = reader.read();
        mEOF = mPeek0 == -1;
        defineCharacterEntity("amp", "&");
        defineCharacterEntity("apos", "'");
        defineCharacterEntity("gt", ">");
        defineCharacterEntity("lt", "<");
        defineCharacterEntity("quot", "\"");
    }

    /** Closes the underlying {@link Reader}. */
    @Override
    public void close() throws IOException {
        mReader.close();
    }

    /** @return A marker for determining if you've come to the end of a specific tag. */
    public String getMarker() {
        switch (mType) {
        case START_TAG:
            return getDepth() - 1 + ":" + getName();
        case END_TAG:
            return getDepth() + ":" + getName();
        default:
            return getDepth() + ":" + mStack.get(mStack.size() - 1);
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
     */
    public boolean withinMarker(String marker) throws IOException {
        if (mType == XMLNodeType.END_TAG) {
            if (marker.equals(getMarker())) {
                next();
                return false;
            }
        } else if (mType == XMLNodeType.END_DOCUMENT) {
            fail("expected: " + XMLNodeType.END_TAG.name() + "/" + marker.substring(marker.indexOf(':') + 1));
        }
        return true;
    }

    /**
     * Requires the current type to be the specified type and optionally requires the name to match
     * as well. If the current type is {@link XMLNodeType#TEXT} and {@link #isWhitespace()} is
     * {@code true} and required type is not {@link XMLNodeType#TEXT}, then {@link #next()} is
     * called prior to the check.
     *
     * @param type The type to require.
     * @param name The name to require. Pass in {@code null} to allow any name.
     */
    public void require(XMLNodeType type, String name) throws IOException {
        if (mType == XMLNodeType.TEXT && type != XMLNodeType.TEXT && isWhitespace()) {
            next();
        }
        if (type != mType || name != null && !name.equals(getName())) {
            fail("expected: " + type.name() + "/" + name);
        }
    }

    /**
     * Skips over the specified tag, which must be the current tag.
     *
     * @param name The name of the tag to skip.
     */
    public void skipTag(String name) throws IOException {
        String marker;
        if (SHOW_SKIPPED_TAGS) {
            Log.warn("Skipping tag: " + name);
        }
        require(XMLNodeType.START_TAG, name);
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
     */
    public String readText() throws IOException {
        StringBuilder builder = new StringBuilder();
        String        marker  = getMarker();

        if (mType == XMLNodeType.START_TAG) {
            next();
        }
        do {
            if (mType == XMLNodeType.TEXT) {
                if (builder.length() > 0) {
                    builder.append(' ');
                }
                builder.append(getText());
                next();
            } else if (mType == XMLNodeType.START_TAG) {
                skipTag(getName());
            }
        } while (withinMarker(marker));
        return builder.toString();
    }

    /** @return The boolean value of a call to {@link #readText()}. */
    public boolean readBoolean() throws IOException {
        return Numbers.extractBoolean(readText());
    }

    /**
     * @param defValue The default value to return if the text cannot be converted to an integer.
     * @return The integer value of a call to {@link #readText()}.
     */
    public int readInteger(int defValue) throws IOException {
        return Numbers.extractInteger(readText(), defValue, false);
    }

    /**
     * @param defValue The default value to return if the text cannot be converted to a long.
     * @return The long value of a call to {@link #readText()}.
     */
    public long readLong(long defValue) throws IOException {
        return Numbers.extractLong(readText(), defValue, false);
    }

    /**
     * @param defValue The default value to return if the text cannot be converted to a double.
     * @return The double value of a call to {@link #readText()}.
     */
    public double readDouble(double defValue) throws IOException {
        return Numbers.extractDouble(readText(), defValue, false);
    }

    /** @return The {@link Calendar} representing the date and time read. */
    public Calendar readDateTime() {
        Calendar calendar = Calendar.getInstance();
        calendar.set(getAttributeAsInteger(XMLWriter.YEAR, 1970), getAttributeAsInteger(XMLWriter.MONTH, 1) - 1, getAttributeAsInteger(XMLWriter.DAY, 1), getAttributeAsInteger(XMLWriter.HOUR, 0), getAttributeAsInteger(XMLWriter.MINUTE, 0), getAttributeAsInteger(XMLWriter.SECOND, 0));
        return calendar;
    }

    private int read() throws IOException {
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

    private void fail(String desc) throws IOException {
        throw new IOException(desc + " pos: " + getPositionDescription());
    }

    private void push(int ch) {
        if (ch != 0) {
            if (mTextPos == mTextBuffer.length) {
                char[] bigger = new char[mTextPos * 4 / 3 + 4];

                System.arraycopy(mTextBuffer, 0, bigger, 0, mTextPos);
                mTextBuffer = bigger;
            }
            mTextBuffer[mTextPos++] = (char) ch;
        }
    }

    private void read(char ch) throws IOException {
        if (read() != ch) {
            fail("expected: '" + ch + "'");
        }
    }

    private void skip() throws IOException {
        while (!mEOF && mPeek0 <= ' ') {
            read();
        }
    }

    private String pop(int pos) {
        String result = new String(mTextBuffer, pos, mTextPos - pos);
        mTextPos = pos;
        return result;
    }

    private String readName() throws IOException {
        int pos = mTextPos;
        int ch  = mPeek0;
        if ((ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') && ch != '_' && ch != ':') {
            fail("name expected");
        }
        do {
            push(read());
            ch = mPeek0;
        } while (ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '_' || ch == '-' || ch == ':' || ch == '.');

        return pop(pos);
    }

    private void parseLegacy(boolean push) throws IOException {
        String req = "";
        int    term;
        read(); // <
        int ch = read();
        if (ch == '?') {
            term = '?';
        } else if (ch == '!') {
            if (mPeek0 == '-') {
                req = "--";
                term = '-';
            } else {
                req = "DOCTYPE";
                term = -1;
            }
        } else {
            if (ch != '[') {
                fail("can't reach: " + ch);
            }
            req = "CDATA[";
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

    private void parseDoctype() throws IOException {
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
            default:
                break;
            }
        }
    }

    private void parseEndTag() throws IOException {
        int pos;

        read(); // '<'
        read(); // '/'
        mName = readName();
        pos = mStack.size() - 1;
        if (pos < 0) {
            fail("element stack empty");
        }
        if (mName.equals(mStack.get(pos))) {
            mStack.remove(pos);
        } else {
            fail("expected: " + mStack.get(pos));
        }
        skip();
        read('>');
    }

    private XMLNodeType peekType() {
        switch (mPeek0) {
        case -1:
            return XMLNodeType.END_DOCUMENT;
        case '&':
            return XMLNodeType.ENTITY_REF;
        case '<':
            switch (mPeek1) {
            case '/':
                return XMLNodeType.END_TAG;
            case '[':
                return XMLNodeType.DATA;
            case '?':
            case '!':
                return XMLNodeType.OTHER;
            default:
                return XMLNodeType.START_TAG;
            }
        default:
            return XMLNodeType.TEXT;
        }
    }

    private void parseStartTag() throws IOException {
        read(); // <
        mName = readName();
        mStack.add(mName);

        while (true) {
            String attrName;
            int    ch;
            int    pos;

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
            if (attrName.isEmpty()) {
                fail("attribute name expected");
            }

            skip();
            read('=');
            skip();
            ch = read();
            if (ch != '\'' && ch != '"') {
                fail("<" + mName + ">: invalid delimiter: " + (char) ch);
            }

            pos = mTextPos;
            pushText(ch);
            mAttributeMap.put(attrName, pop(pos));
            //noinspection ConstantConditions
            if (ch != ' ') {
                read(); // skip end quote
            }
        }
    }

    private boolean pushEntity() throws IOException {
        boolean whitespace = true;
        int     pos;
        String  code;
        String  result;

        read(); // &
        pos = mTextPos;
        while (!mEOF && mPeek0 != ';') {
            push(read());
        }

        code = pop(pos);
        read();
        if (!code.isEmpty() && code.charAt(0) == '#') {
            int c = code.charAt(1) == 'x' ? Integer.parseInt(code.substring(2), 16) : Integer.parseInt(code.substring(1));

            push(c);
            return c <= ' ';
        }

        result = mEntityMap.get(code);
        if (result == null) {
            result = "&" + code + ";";
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

    private boolean pushText(int delimiter) throws IOException {
        boolean whitespace = true;
        int     next       = mPeek0;

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
     * @param value  The value to substitute.
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
        buffer.append(" @").append(mLine).append(":").append(mColumn).append(": ");
        if (mType == XMLNodeType.START_TAG) {
            buffer.append('<').append(mName).append('>');
        } else if (mType == XMLNodeType.END_TAG) {
            buffer.append("</").append(mName).append('>');
        } else if (mIsWhitespace) {
            buffer.append("[whitespace]");
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
            mText = Text.standardizeLineEndings(pop(0));
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
     * @param name     The name of the attribute to use.
     * @param defValue The default value to use if the attribute value isn't present.
     * @return The value of the attribute.
     */
    public String getAttribute(String name, String defValue) {
        String value = mAttributeMap.get(name);
        return value != null ? value : defValue;
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
        return Numbers.extractBoolean(mAttributeMap.get(name));
    }

    /**
     * @param name     The attribute name.
     * @param defValue The default value to use if the attribute value can't be converted to an
     *                 integer.
     * @return The value of the tag.
     */
    public int getAttributeAsInteger(String name, int defValue) {
        return Numbers.extractInteger(mAttributeMap.get(name), defValue, false);
    }

    /**
     * @param name     The attribute name.
     * @param defValue The default value to use if the attribute value can't be converted to a
     *                 long.
     * @return The value of the tag.
     */
    public long getAttributeAsLong(String name, long defValue) {
        return Numbers.extractLong(mAttributeMap.get(name), defValue, false);
    }

    /**
     * @param name     The attribute name.
     * @param defValue The default value to use if the attribute value can't be converted to a
     *                 double.
     * @return The value of the tag.
     */
    public double getAttributeAsDouble(String name, double defValue) {
        return Numbers.extractDouble(mAttributeMap.get(name), defValue, false);
    }

    /** @return The map of attributes. */
    public HashMap<String, String> getAttributes() {
        return mAttributeMap;
    }

    /** @return The type at the current position. */
    public XMLNodeType getType() {
        return mType;
    }

    /**
     * Advances to the next position.
     *
     * @return The type.
     */
    public XMLNodeType next() throws IOException {
        if (mIsEmptyElementTag) {
            mType = XMLNodeType.END_TAG;
            mIsEmptyElementTag = false;
            mStack.remove(mStack.size() - 1);
        } else {
            int textOrdinal = XMLNodeType.TEXT.ordinal();

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
                    mType = XMLNodeType.TEXT;
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
                    mType = XMLNodeType.TEXT;
                    break;
                default:
                    parseLegacy(false);
                    break;
                }
            } while (mType.ordinal() > textOrdinal || mType == XMLNodeType.TEXT && peekType().ordinal() >= textOrdinal);

            mIsWhitespace &= mType == XMLNodeType.TEXT;
        }
        return mType;
    }
}
