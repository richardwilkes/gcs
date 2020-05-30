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

import com.trollworks.gcs.utility.text.Numbers;

import java.io.IOException;
import java.io.OutputStream;
import java.io.OutputStreamWriter;
import java.io.PrintWriter;
import java.util.Calendar;
import java.util.Date;

/** A {@link PrintWriter} that has been extended to provide common XML writing helper methods. */
public class XMLWriter extends PrintWriter {
    private static final String END_TAG             = "/>";
    private static final String ENTITY_CODE_PREFIX  = "&#";
    private static final String AMPERSAND_ENTITY    = "&amp;";
    private static final String LESS_THAN_ENTITY    = "&lt;";
    private static final String GREATER_THAN_ENTITY = "&gt;";
    private static final String END_COMMENT         = " -->";
    /** The encoding used. */
    public static final  String ENCODING            = "US-ASCII";
    /** The 'year' attribute. */
    public static final  String YEAR                = "year";
    /** The 'month' attribute. */
    public static final  String MONTH               = "month";
    /** The 'day' attribute. */
    public static final  String DAY                 = "day";
    /** The 'hour' attribute. */
    public static final  String HOUR                = "hour";
    /** The 'minute' attribute. */
    public static final  String MINUTE              = "minute";
    /** The 'second' attribute. */
    public static final  String SECOND              = "second";
    private              int    mIndent;

    /**
     * Creates a new XML writer.
     *
     * @param stream The stream to write to.
     */
    public XMLWriter(OutputStream stream) throws IOException {
        super(new OutputStreamWriter(stream, ENCODING));
    }

    /** Writes a standard XML header. */
    public void writeHeader() {
        print("<?xml version=\"1.0\" encoding=\"");
        print(ENCODING);
        println("\" ?>");
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
        print("<!-- ");
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
                print("&quot;");
            } else if (ch == '\'') {
                print("&apos;");
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
     * @param name  The name of the attribute.
     * @param value The value of the attribute.
     */
    public void writeAttribute(String name, boolean value) {
        writeAttribute(name, Numbers.format(value));
    }

    /**
     * Writes an XML attribute.
     *
     * @param name  The name of the attribute.
     * @param value The value of the attribute.
     */
    public void writeAttribute(String name, int value) {
        writeAttribute(name, Integer.toString(value));
    }

    /**
     * Writes an XML attribute.
     *
     * @param name  The name of the attribute.
     * @param value The value of the attribute.
     */
    public void writeAttribute(String name, long value) {
        writeAttribute(name, Long.toString(value));
    }

    /**
     * Writes an XML attribute.
     *
     * @param name  The name of the attribute.
     * @param value The value of the attribute.
     */
    public void writeAttribute(String name, float value) {
        writeAttribute(name, Float.toString(value));
    }

    /**
     * Writes an XML attribute.
     *
     * @param name  The name of the attribute.
     * @param value The value of the attribute.
     */
    public void writeAttribute(String name, double value) {
        writeAttribute(name, Double.toString(value));
    }

    /**
     * Writes an XML attribute.
     *
     * @param name  The name of the attribute.
     * @param value The value of the attribute.
     */
    public void writeAttribute(String name, String value) {
        print(' ');
        print(name);
        print("=\"");
        writeEncodedAttribute(value);
        print('"');
    }

    /**
     * Writes an XML attribute if its value is not zero.
     *
     * @param name  The name of the attribute.
     * @param value The value of the attribute.
     */
    public void writeAttributeNotZero(String name, int value) {
        if (value != 0) {
            writeAttribute(name, Integer.toString(value));
        }
    }

    /**
     * Writes an XML attribute if its value is not zero.
     *
     * @param name  The name of the attribute.
     * @param value The value of the attribute.
     */
    public void writeAttributeNotZero(String name, long value) {
        if (value != 0) {
            writeAttribute(name, Long.toString(value));
        }
    }

    /**
     * Writes an XML attribute.
     *
     * @param name  The name of the attribute.
     * @param value The value of the attribute.
     */
    public void writeAttributeNotZero(String name, float value) {
        if (value != 0) {
            writeAttribute(name, Float.toString(value));
        }
    }

    /**
     * Writes an XML attribute.
     *
     * @param name  The name of the attribute.
     * @param value The value of the attribute.
     */
    public void writeAttributeNotZero(String name, double value) {
        if (value != 0) {
            writeAttribute(name, Double.toString(value));
        }
    }

    /**
     * Writes an XML attribute.
     *
     * @param name  The name of the attribute.
     * @param value The value of the attribute.
     */
    public void writeAttributeNotOne(String name, float value) {
        if (value != 1) {
            writeAttribute(name, Float.toString(value));
        }
    }

    /**
     * Writes an XML attribute.
     *
     * @param name  The name of the attribute.
     * @param value The value of the attribute.
     */
    public void writeAttributeNotOne(String name, double value) {
        if (value != 1) {
            writeAttribute(name, Double.toString(value));
        }
    }

    /**
     * Write out an XML tag with a single attribute and no children, with a trailing line feed.
     *
     * @param name           The name of the tag.
     * @param value          The data to place between the tags.
     * @param attribute      The name of the attribute.
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
     * @param name           The name of the tag.
     * @param value          The data to place between the tags.
     * @param attribute      The name of the attribute.
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
     * @param name           The name of the tag.
     * @param value          The data to place between the tags.
     * @param attribute      The name of the attribute.
     * @param attributeValue The value of the attribute.
     */
    public void simpleTagWithAttribute(String name, int value, String attribute, boolean attributeValue) {
        startTag(name);
        writeAttribute(attribute, attributeValue);
        finishTag();
        writeEncodedData(Integer.toString(value));
        endTagEOL(name, false);
    }

    /**
     * Write out an XML tag with a single attribute and no children, with a trailing line feed.
     *
     * @param name           The name of the tag.
     * @param value          The data to place between the tags.
     * @param attribute      The name of the attribute.
     * @param attributeValue The value of the attribute.
     */
    public void simpleTagWithAttribute(String name, long value, String attribute, boolean attributeValue) {
        startTag(name);
        writeAttribute(attribute, attributeValue);
        finishTag();
        writeEncodedData(Long.toString(value));
        endTagEOL(name, false);
    }

    /**
     * Write out an XML tag with a single attribute and no children, with a trailing line feed.
     *
     * @param name           The name of the tag.
     * @param value          The data to place between the tags.
     * @param attribute      The name of the attribute.
     * @param attributeValue The value of the attribute.
     */
    public void simpleTagWithAttribute(String name, double value, String attribute, boolean attributeValue) {
        startTag(name);
        writeAttribute(attribute, attributeValue);
        finishTag();
        writeEncodedData(toString(value));
        endTagEOL(name, false);
    }

    /**
     * Write out an XML tag with a single attribute and no children, with a trailing line feed.
     *
     * @param name           The name of the tag.
     * @param value          The data to place between the tags.
     * @param attribute      The name of the attribute.
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
     * @param name           The name of the tag.
     * @param value          The data to place between the tags.
     * @param attribute      The name of the attribute.
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
     * @param name           The name of the tag.
     * @param value          The data to place between the tags.
     * @param attribute      The name of the attribute.
     * @param attributeValue The value of the attribute.
     */
    public void simpleTagWithAttribute(String name, double value, String attribute, String attributeValue) {
        startTag(name);
        writeAttribute(attribute, attributeValue);
        finishTag();
        writeEncodedData(toString(value));
        endTagEOL(name, false);
    }

    private static String toString(double value) {
        String result = Double.toString(value);
        if (result.endsWith(".0")) {
            result = result.substring(0, result.length() - 2);
        }
        return result;
    }

    /**
     * Write out an XML tag with a single attribute and no children, with a trailing line feed.
     *
     * @param name           The name of the tag.
     * @param value          The data to place between the tags.
     * @param attribute      The name of the attribute.
     * @param attributeValue The value of the attribute.
     */
    public void simpleTagWithAttribute(String name, int value, String attribute, String attributeValue) {
        startTag(name);
        writeAttribute(attribute, attributeValue);
        finishTag();
        writeEncodedData(Integer.toString(value));
        endTagEOL(name, false);
    }

    /**
     * Write out an XML tag with a single attribute and no children, with a trailing line feed.
     *
     * @param name           The name of the tag.
     * @param value          The data to place between the tags.
     * @param attribute      The name of the attribute.
     * @param attributeValue The value of the attribute.
     */
    public void simpleTagWithAttribute(String name, long value, String attribute, String attributeValue) {
        startTag(name);
        writeAttribute(attribute, attributeValue);
        finishTag();
        writeEncodedData(Long.toString(value));
        endTagEOL(name, false);
    }

    /**
     * Write out a simple XML tag (i.e. no children or attributes).
     *
     * @param name  The name of the tag.
     * @param value The data to place between the tags.
     */
    public void simpleTag(String name, boolean value) {
        simpleTag(name, Boolean.toString(value));
    }

    /**
     * Write out a simple XML tag (i.e. no children or attributes).
     *
     * @param name  The name of the tag.
     * @param value The data to place between the tags.
     */
    public void simpleTag(String name, int value) {
        simpleTag(name, Integer.toString(value));
    }

    /**
     * Write out a simple XML tag (i.e. no children or attributes).
     *
     * @param name  The name of the tag.
     * @param value The data to place between the tags.
     */
    public void simpleTagNotZero(String name, int value) {
        if (value != 0) {
            simpleTag(name, Integer.toString(value));
        }
    }

    /**
     * Write out a simple XML tag (i.e. no children or attributes).
     *
     * @param name  The name of the tag.
     * @param value The data to place between the tags.
     */
    public void simpleTag(String name, long value) {
        simpleTag(name, Long.toString(value));
    }

    /**
     * Write out a simple XML tag (i.e. no children or attributes).
     *
     * @param name  The name of the tag.
     * @param value The data to place between the tags.
     */
    public void simpleTagNotZero(String name, long value) {
        if (value != 0) {
            simpleTag(name, Long.toString(value));
        }
    }

    /**
     * Write out a simple XML tag (i.e. no children or attributes).
     *
     * @param name  The name of the tag.
     * @param value The data to place between the tags.
     */
    public void simpleTag(String name, double value) {
        simpleTag(name, toString(value));
    }

    /**
     * Write out a simple XML tag (i.e. no children or attributes).
     *
     * @param name  The name of the tag.
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
     * @param name  The name of the tag.
     * @param value The data to place between the tags.
     */
    public void simpleTagNotEmpty(String name, String value) {
        if (value != null && !value.isEmpty()) {
            startSimpleTag(name);
            writeEncodedData(value);
            endTagEOL(name, false);
        }
    }

    /**
     * Write out a simple XML tag (i.e. no children or attributes).
     *
     * @param name  The name of the tag.
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
     * @param name   The name of the tag.
     * @param indent Whether to indent before writing the end tag.
     */
    public void endTagEOL(String name, boolean indent) {
        outdent();
        if (indent) {
            writeIndentation();
        }
        print("</");
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
     * @param name           The name of the tag.
     * @param dateInMillis   The date to output.
     * @param includeDate    Whether to output the date fields.
     * @param includeTime    Whether to output the time fields.
     * @param includeSeconds Whether to output the seconds field. Only relevant if {@code
     *                       includeTime} was {@code true}.
     */
    public void writeDateTimeTag(String name, long dateInMillis, boolean includeDate, boolean includeTime, boolean includeSeconds) {
        Calendar calendar = Calendar.getInstance();

        calendar.setTimeInMillis(dateInMillis);
        writeDateTimeTag(name, calendar, includeDate, includeTime, includeSeconds);
    }

    /**
     * Write out a date and time XML tag.
     *
     * @param name           The name of the tag.
     * @param date           The date to output.
     * @param includeDate    Whether to output the date fields.
     * @param includeTime    Whether to output the time fields.
     * @param includeSeconds Whether to output the seconds field. Only relevant if {@code
     *                       includeTime} was {@code true}.
     */
    public void writeDateTimeTag(String name, Date date, boolean includeDate, boolean includeTime, boolean includeSeconds) {
        Calendar calendar = Calendar.getInstance();

        calendar.setTime(date);
        writeDateTimeTag(name, calendar, includeDate, includeTime, includeSeconds);
    }

    /**
     * Write out a date and time XML tag.
     *
     * @param name           The name of the tag.
     * @param calendar       The calendar to output.
     * @param includeDate    Whether to output the date fields.
     * @param includeTime    Whether to output the time fields.
     * @param includeSeconds Whether to output the seconds field. Only relevant if {@code
     *                       includeTime} was {@code true}.
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
        int           length = data.length();

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
